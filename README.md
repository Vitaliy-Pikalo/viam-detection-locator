# detection-locator

A [Viam](https://www.viam.com) sensor module that turns person detections from **any** vision service into **tracked people with ground positions in metres**.

Detectors answer "is there a person in this frame?" with a box in pixels. A robot needs "where is that person, and is it the same person as last frame?" This module answers both. It started from a search-and-rescue drone I led at CMU Project Ignite, where thermal detections had to become coordinates the flight system could act on.

| Model | API |
|---|---|
| `vitaliy-pikalo:detection-locator:locator` | `rdk:component:sensor` |

## How it works

1. Each `Readings` call asks the configured vision service for detections on the configured camera.
2. Detections are filtered by label (case-insensitive, default `person`) and confidence (default `0.5`).
3. A greedy IoU tracker (threshold 0.3) matches them to existing tracks. Unmatched detections start new tracks; a track is dropped after 5 consecutive missed frames.
4. The bottom-centre of each box is projected onto flat ground with a pinhole model, using camera height, tilt and field of view.

## Configuration

```json
{
  "camera": "cam",
  "vision": "people-detector",
  "camera_height_m": 1.5,
  "tilt_deg": 30,
  "horizontal_fov_deg": 70,
  "vertical_fov_deg": 43,
  "image_width_px": 640,
  "image_height_px": 480,
  "label": "person",
  "min_confidence": 0.5
}
```

| Attribute | Type | Required | Notes |
|---|---|---|---|
| `camera` | string | yes | camera the vision service reads from |
| `vision` | string | yes | any vision service that returns detections (for example `mlmodel` or `viam-labs:near-ir-person-detector`) |
| `camera_height_m` | float | yes | lens height above the ground |
| `tilt_deg` | float | yes | angle below horizontal, 0–90 |
| `horizontal_fov_deg`, `vertical_fov_deg` | float | yes | from the camera's datasheet |
| `image_width_px`, `image_height_px` | int | yes | resolution the detector sees |
| `label` | string | no | `""` keeps every label |
| `min_confidence` | float | no | 0–1 |

## Readings

```json
{
  "count": 1,
  "people": [
    {"id": 1, "label": "person", "confidence": 0.91, "frames_seen": 12, "missed": 0,
     "on_ground": true, "forward_m": 4.2, "right_m": -0.6}
  ]
}
```

`forward_m` and `right_m` are measured from the point on the ground directly below the camera. If a box's bottom edge is at or above the horizon, `on_ground` is `false` and no position is reported.

`DoCommand({"reset": true})` clears all tracks.

## Limits

- Assumes flat ground and no camera roll. Slopes, stairs and lens distortion add error.
- Uses the bottom of the box as the feet. Partly hidden people (cut off at the knees) will read as farther away.
- IoU tracking can swap ids when two people cross closely. It doesn't use appearance features.

## Development

```bash
make test               # unit tests for the geometry, tracker and readings (stdlib only)
make module.tar.gz      # build the module for upload
```

The core logic lives in `locator/` and has no Viam dependency, so it can be tested anywhere.
