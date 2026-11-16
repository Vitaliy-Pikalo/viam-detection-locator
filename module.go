// Package detectionlocator is a Viam sensor that tracks people seen by a vision
// service and reports where each one stands on the ground, in metres from the camera.
package detectionlocator

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/services/vision"

	"github.com/Vitaliy-Pikalo/viam-detection-locator/locator"
)

// Locator is the model triplet registered with viam-server.
var Locator = resource.NewModel("vitaliy-pikalo", "detection-locator", "locator")

func init() {
	resource.RegisterComponent(sensor.API, Locator,
		resource.Registration[sensor.Sensor, *Config]{Constructor: newLocator})
}

// Config is the JSON attributes block for the model.
type Config struct {
	Camera        string   `json:"camera"`
	Vision        string   `json:"vision"`
	HeightM       float64  `json:"camera_height_m"`
	TiltDeg       float64  `json:"tilt_deg"`
	HFOVDeg       float64  `json:"horizontal_fov_deg"`
	VFOVDeg       float64  `json:"vertical_fov_deg"`
	WidthPx       int      `json:"image_width_px"`
	HeightPx      int      `json:"image_height_px"`
	Label         *string  `json:"label,omitempty"`          // default "person"; "" keeps all labels
	MinConfidence *float64 `json:"min_confidence,omitempty"` // default 0.5
}

func (c *Config) camera() locator.Camera {
	return locator.Camera{HeightM: c.HeightM, TiltDeg: c.TiltDeg, HFOVDeg: c.HFOVDeg,
		VFOVDeg: c.VFOVDeg, WidthPx: c.WidthPx, HeightPx: c.HeightPx}
}

// Validate checks the config and declares the camera and vision service as required dependencies.
func (c *Config) Validate(path string) ([]string, []string, error) {
	if c.Camera == "" {
		return nil, nil, fmt.Errorf("%s: \"camera\" is required", path)
	}
	if c.Vision == "" {
		return nil, nil, fmt.Errorf("%s: \"vision\" is required", path)
	}
	if err := c.camera().Validate(); err != nil {
		return nil, nil, fmt.Errorf("%s: %w", path, err)
	}
	if c.MinConfidence != nil && (*c.MinConfidence < 0 || *c.MinConfidence > 1) {
		return nil, nil, fmt.Errorf("%s: min_confidence must be between 0 and 1", path)
	}
	return []string{c.Camera, c.Vision}, nil, nil
}

type detectionLocator struct {
	resource.Named
	resource.AlwaysRebuild
	resource.TriviallyCloseable

	logger  logging.Logger
	vis     vision.Service
	camName string
	cam     locator.Camera
	label   string
	minConf float64

	mu      sync.Mutex
	tracker *locator.Tracker
}

func newLocator(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}
	vis, err := resource.FromDependencies[vision.Service](deps, vision.Named(conf.Vision))
	if err != nil {
		return nil, fmt.Errorf("vision service %q: %w", conf.Vision, err)
	}
	label, minConf := "person", 0.5
	if conf.Label != nil {
		label = *conf.Label
	}
	if conf.MinConfidence != nil {
		minConf = *conf.MinConfidence
	}
	return &detectionLocator{
		Named:   rawConf.ResourceName().AsNamed(),
		logger:  logger,
		vis:     vis,
		camName: conf.Camera,
		cam:     conf.camera(),
		label:   label,
		minConf: minConf,
		tracker: locator.NewTracker(0.3, 5),
	}, nil
}

// Readings runs detection on the latest frame, updates tracks and returns one entry per tracked person.
func (d *detectionLocator) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	raw, err := d.vis.DetectionsFromCamera(ctx, d.camName, extra)
	if err != nil {
		return nil, err
	}
	dets := make([]locator.Detection, 0, len(raw))
	for _, r := range raw {
		bb := r.BoundingBox()
		if bb == nil {
			continue
		}
		dets = append(dets, locator.Detection{
			Box:        locator.Box{X0: float64(bb.Min.X), Y0: float64(bb.Min.Y), X1: float64(bb.Max.X), Y1: float64(bb.Max.Y)},
			Label:      r.Label(),
			Confidence: r.Score(),
		})
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	return locator.Readings(d.cam, d.tracker.Update(locator.Filter(dets, d.label, d.minConf))), nil
}

// DoCommand supports {"reset": true} to clear all tracks.
func (d *detectionLocator) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	if v, ok := cmd["reset"].(bool); ok && v {
		d.mu.Lock()
		d.tracker = locator.NewTracker(0.3, 5)
		d.mu.Unlock()
		return map[string]interface{}{"reset": true}, nil
	}
	return nil, errors.New(`unknown command; supported: {"reset": true}`)
}
