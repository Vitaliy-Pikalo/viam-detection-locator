package locator

import (
	"math"
	"testing"
)

func cam() Camera {
	return Camera{HeightM: 2, TiltDeg: 45, HFOVDeg: 90, VFOVDeg: 60, WidthPx: 640, HeightPx: 480}
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func TestGroundPointCentre(t *testing.T) {
	f, r, ok := cam().GroundPoint(320, 240)
	if !ok || !near(f, 2.0) || !near(r, 0) {
		t.Fatalf("got forward=%v right=%v ok=%v, want 2, 0, true", f, r, ok)
	}
}

func TestGroundPointAboveHorizon(t *testing.T) {
	c := cam()
	c.TiltDeg = 10 // top of a 60° vertical FOV looks 20° above horizontal
	if _, _, ok := c.GroundPoint(320, 0); ok {
		t.Fatal("pixel above the horizon should not hit the ground")
	}
}

func TestGroundPointRightHalf(t *testing.T) {
	_, r, ok := cam().GroundPoint(600, 240)
	if !ok || r <= 0 {
		t.Fatalf("right half should give right > 0, got %v ok=%v", r, ok)
	}
	_, l, _ := cam().GroundPoint(40, 240)
	if !near(l, -r) {
		t.Fatalf("mirror pixels should give mirror offsets: %v vs %v", l, r)
	}
}

func TestGroundPointLowerIsCloser(t *testing.T) {
	near1, _, _ := cam().GroundPoint(320, 470)
	far1, _, _ := cam().GroundPoint(320, 250)
	if !(near1 < far1) {
		t.Fatalf("lower pixel should be closer: %v vs %v", near1, far1)
	}
}

func TestValidate(t *testing.T) {
	if err := cam().Validate(); err != nil {
		t.Fatalf("valid camera rejected: %v", err)
	}
	bad := []Camera{
		{HeightM: 0, TiltDeg: 45, HFOVDeg: 90, VFOVDeg: 60, WidthPx: 640, HeightPx: 480},
		{HeightM: 2, TiltDeg: 45, HFOVDeg: 0, VFOVDeg: 60, WidthPx: 640, HeightPx: 480},
		{HeightM: 2, TiltDeg: 45, HFOVDeg: 90, VFOVDeg: 180, WidthPx: 640, HeightPx: 480},
		{HeightM: 2, TiltDeg: 45, HFOVDeg: 90, VFOVDeg: 60, WidthPx: 0, HeightPx: 480},
		{HeightM: 2, TiltDeg: 95, HFOVDeg: 90, VFOVDeg: 60, WidthPx: 640, HeightPx: 480},
	}
	for i, c := range bad {
		if c.Validate() == nil {
			t.Errorf("case %d: expected error", i)
		}
	}
}
