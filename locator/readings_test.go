package locator

import "testing"

func TestFilterLabelCaseAndConfidence(t *testing.T) {
	in := []Detection{
		{Label: "Person", Confidence: 0.9},
		{Label: "person", Confidence: 0.2},
		{Label: "dog", Confidence: 0.99},
	}
	got := Filter(in, "person", 0.5)
	if len(got) != 1 || got[0].Confidence != 0.9 {
		t.Fatalf("want only the 0.9 Person, got %+v", got)
	}
	if all := Filter(in, "", 0); len(all) != 3 {
		t.Fatalf("empty label should keep every label, got %d", len(all))
	}
}

func TestReadingsOmitsPositionAboveHorizon(t *testing.T) {
	c := cam()
	c.TiltDeg = 10
	tracks := []Track{
		{ID: 1, Det: Detection{Box: Box{300, 400, 340, 470}, Label: "person", Confidence: 0.8}, FramesSeen: 3},
		{ID: 2, Det: Detection{Box: Box{300, 0, 340, 0}, Label: "person", Confidence: 0.7}, FramesSeen: 1},
	}
	r := Readings(c, tracks)
	people := r["people"].([]interface{})
	if r["count"] != 2 || len(people) != 2 {
		t.Fatalf("want 2 people, got %+v", r)
	}
	p1 := people[0].(map[string]interface{})
	if _, ok := p1["forward_m"]; !ok || p1["id"] != 1 || p1["frames_seen"] != 3 {
		t.Fatalf("track 1 should have a position: %+v", p1)
	}
	p2 := people[1].(map[string]interface{})
	if _, ok := p2["forward_m"]; ok || p2["on_ground"] != false {
		t.Fatalf("track 2 is above the horizon and must have no position: %+v", p2)
	}
}

func TestReadingsEmpty(t *testing.T) {
	r := Readings(cam(), nil)
	if r["count"] != 0 || len(r["people"].([]interface{})) != 0 {
		t.Fatalf("empty readings wrong: %+v", r)
	}
}
