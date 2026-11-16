# detection-locator Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Viam sensor module that tracks detected people and reports ground offsets in metres.

**Architecture:** `locator/` is a stdlib-only core (geometry + tracker). The repo root package adapts it to the RDK sensor API; `cmd/module` is the entrypoint.

**Tech Stack:** Go 1.23, go.viam.com/rdk.

**Spec:** `docs/superpowers/specs/2026-11-16-detection-locator-design.md`

## Global Constraints
- Module id `vitaliy-pikalo:detection-locator`, model `vitaliy-pikalo:detection-locator:locator`, API `rdk:component:sensor`.
- IoU threshold 0.3; max missed frames 5; default label `person`; default min confidence 0.5.
- No paid Viam services.

## Review Focus
- Box bottom above the horizon → position omitted, no NaN/Inf in readings.
- Zero/negative config values (height, FOV, image size) → Validate error.
- Two people crossing → ids must not swap when IoU is ambiguous (greedy highest-IoU first).
- Label case ("Person" vs "person") → match case-insensitively.
- Empty detection list → existing tracks age, nothing crashes.

---

### Task 1: Ground geometry
**Files:** Create `locator/geometry.go`, Test `locator/geometry_test.go`
**Produces:** `type Camera struct{HeightM, TiltDeg, HFOVDeg, VFOVDeg float64; WidthPx, HeightPx int}`; `func (c Camera) Validate() error`; `func (c Camera) GroundPoint(u, v float64) (forwardM, rightM float64, ok bool)`
- [ ] Tests: centre pixel at tilt 45°, height 2 → forward 2.0, right 0; pixel above horizon → ok=false; right half of image → right > 0; Validate rejects zero height/FOV/size.
- [ ] Implement: ray in camera frame (xc = xn·tan(hfov/2), yc = yn·tan(vfov/2), 1), rotate by tilt; t = h/(yc·cosθ + sinθ); forward = t(cosθ − yc·sinθ), right = t·xc.
- [ ] `go test ./locator` passes; commit.

### Task 2: IoU tracker
**Files:** Create `locator/tracker.go`, Test `locator/tracker_test.go`
**Produces:** `type Box struct{X0,Y0,X1,Y1 float64}`; `type Detection struct{Box Box; Label string; Confidence float64}`; `type Track struct{ID int; Det Detection; FramesSeen, Missed int}`; `func NewTracker(iou float64, maxMissed int) *Tracker`; `func (t *Tracker) Update(dets []Detection) []Track`
- [ ] Tests: same box across frames keeps id 1; new far box gets id 2; track dropped after 6 empty frames; greedy highest IoU wins.
- [ ] Implement; tests pass; commit.

### Task 3: Viam sensor wrapper + packaging
**Files:** `module.go`, `cmd/module/main.go`, `meta.json`, `Makefile`, `go.mod`, `README.md`, `.github/workflows/deploy.yml`
- [ ] Config + Validate (required deps: camera, vision); `Readings` filters label/confidence, calls `Tracker.Update`, `Camera.GroundPoint` on box bottom-centre.
- [ ] `make module.tar.gz` builds (verified by Viam cloud build); commit.
