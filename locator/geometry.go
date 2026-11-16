// Package locator turns 2D detections into tracked ground positions.
// It uses only the standard library so it can be tested without the Viam SDK.
package locator

import (
	"errors"
	"math"
)

// Camera describes a fixed pinhole camera looking down at flat ground.
type Camera struct {
	HeightM  float64 // lens height above the ground, metres
	TiltDeg  float64 // optical axis angle below horizontal, degrees (0 = level, 90 = straight down)
	HFOVDeg  float64 // horizontal field of view, degrees
	VFOVDeg  float64 // vertical field of view, degrees
	WidthPx  int     // image width, pixels
	HeightPx int     // image height, pixels
}

// Validate reports the first config value that makes the geometry meaningless.
func (c Camera) Validate() error {
	switch {
	case !(c.HeightM > 0):
		return errors.New("camera height must be > 0 m")
	case c.TiltDeg < 0 || c.TiltDeg > 90:
		return errors.New("tilt must be between 0 and 90 degrees")
	case !(c.HFOVDeg > 0 && c.HFOVDeg < 180):
		return errors.New("horizontal FOV must be between 0 and 180 degrees")
	case !(c.VFOVDeg > 0 && c.VFOVDeg < 180):
		return errors.New("vertical FOV must be between 0 and 180 degrees")
	case c.WidthPx <= 0 || c.HeightPx <= 0:
		return errors.New("image width and height must be > 0 px")
	}
	return nil
}

func rad(deg float64) float64 { return deg * math.Pi / 180 }

// GroundPoint projects pixel (u, v) onto flat ground. It returns the offset of
// that ground point from the point directly below the camera: forwardM along the
// camera's heading and rightM to its right. ok is false when the pixel's ray
// never reaches the ground (at or above the horizon).
func (c Camera) GroundPoint(u, v float64) (forwardM, rightM float64, ok bool) {
	xn := (u - float64(c.WidthPx)/2) / (float64(c.WidthPx) / 2)
	yn := (v - float64(c.HeightPx)/2) / (float64(c.HeightPx) / 2) // + is down
	xc := xn * math.Tan(rad(c.HFOVDeg)/2)
	yc := yn * math.Tan(rad(c.VFOVDeg)/2)
	th := rad(c.TiltDeg)

	down := yc*math.Cos(th) + math.Sin(th) // downward component of the ray
	if down <= 1e-9 {
		return 0, 0, false
	}
	t := c.HeightM / down
	return t * (math.Cos(th) - yc*math.Sin(th)), t * xc, true
}
