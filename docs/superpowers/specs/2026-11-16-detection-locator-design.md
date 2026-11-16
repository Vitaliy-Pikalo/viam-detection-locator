# detection-locator — design

## Intent
A Viam registry module that turns person detections from any vision service into
tracked people with ground positions in metres. It continues the CMU Project
Ignite search-and-rescue drone work (thermal person detection → coordinates for
the flight system) and shows Go work on Viam's own platform.

## Goals
- Sensor component `vitaliy-pikalo:detection-locator:locator`.
- Config: camera name, vision service name, camera height (m), tilt below
  horizontal (deg), horizontal and vertical field of view (deg), image width and
  height (px), label filter (default `person`), minimum confidence (default 0.5).
- Each `Readings` call: query detections, keep matching labels above the
  confidence floor, match them to existing tracks by IoU, return one entry per
  live track: `id`, `label`, `confidence`, `forward_m`, `right_m`, `frames_seen`.
- Pure-Go core (`locator/`) with stdlib-only unit tests.

## Non-goals
- Training or shipping a detection model.
- Drop logic, alerts, SMS (not Vitaliy's work on the drone).
- GPS / world coordinates (offsets are relative to the camera's ground point).

## Anti-goals
- Claiming accuracy the flat-ground pinhole model can't give.
- Any paid Viam usage.

## Approaches
1. Thermal detector in Go — rejected: duplicates `viam-labs:near-ir-person-detector`, needs weights.
2. Python wrapper — rejected: leaves the Go gap on the resume.
3. **Go wrapper with tracking + ground geometry — picked.**

## Decisions
- Geometry: flat ground, pinhole camera, tilt about the camera x-axis only.
  Bottom-centre of each box is the ground contact point. Rays at or above the
  horizon return "no ground hit" and the detection still tracks with no position.
- Tracking: greedy IoU matching, threshold 0.3; a track is dropped after 5
  consecutive missed frames; ids are increasing integers starting at 1.

## Left to the builder
- RDK API details (validated by Viam cloud build; the sandbox that wrote this
  could not reach the Go module proxy).
