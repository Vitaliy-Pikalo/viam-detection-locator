package locator

import "strings"

// Filter keeps detections whose label matches (case-insensitive; empty label
// keeps all) and whose confidence is at least minConf.
func Filter(dets []Detection, label string, minConf float64) []Detection {
	var out []Detection
	for _, d := range dets {
		if (label == "" || strings.EqualFold(d.Label, label)) && d.Confidence >= minConf {
			out = append(out, d)
		}
	}
	return out
}

// Readings builds a sensor reading from live tracks. Each person's ground
// position comes from the bottom-centre of their box. Values use only types
// that convert cleanly to protobuf Structs (float64, int, bool, string, slices, maps).
func Readings(c Camera, tracks []Track) map[string]interface{} {
	people := make([]interface{}, 0, len(tracks))
	for _, tk := range tracks {
		b := tk.Det.Box
		p := map[string]interface{}{
			"id":          tk.ID,
			"label":       tk.Det.Label,
			"confidence":  tk.Det.Confidence,
			"frames_seen": tk.FramesSeen,
			"missed":      tk.Missed,
		}
		if f, r, ok := c.GroundPoint((b.X0+b.X1)/2, b.Y1); ok {
			p["forward_m"], p["right_m"], p["on_ground"] = f, r, true
		} else {
			p["on_ground"] = false
		}
		people = append(people, p)
	}
	return map[string]interface{}{"count": len(tracks), "people": people}
}
