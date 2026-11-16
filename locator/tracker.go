package locator

import "sort"

// Box is an axis-aligned bounding box in pixels.
type Box struct{ X0, Y0, X1, Y1 float64 }

func (b Box) area() float64 {
	w, h := b.X1-b.X0, b.Y1-b.Y0
	if w <= 0 || h <= 0 {
		return 0
	}
	return w * h
}

// IoU is the intersection-over-union of two boxes, in [0, 1].
func (b Box) IoU(o Box) float64 {
	inter := Box{max(b.X0, o.X0), max(b.Y0, o.Y0), min(b.X1, o.X1), min(b.Y1, o.Y1)}.area()
	union := b.area() + o.area() - inter
	if union <= 0 {
		return 0
	}
	return inter / union
}

// Detection is one labelled box from a vision service.
type Detection struct {
	Box        Box
	Label      string
	Confidence float64
}

// Track is a detection followed across frames.
type Track struct {
	ID         int
	Det        Detection // most recent matched detection
	FramesSeen int
	Missed     int // consecutive frames with no match
}

// Tracker matches detections to tracks with greedy highest-IoU-first assignment.
type Tracker struct {
	iou       float64
	maxMissed int
	nextID    int
	tracks    []*Track
}

// NewTracker returns a tracker that matches at IoU >= iou and drops a track
// after more than maxMissed consecutive unmatched frames.
func NewTracker(iou float64, maxMissed int) *Tracker {
	return &Tracker{iou: iou, maxMissed: maxMissed, nextID: 1}
}

// Update feeds one frame of detections and returns the live tracks, ordered by ID.
func (t *Tracker) Update(dets []Detection) []Track {
	type pair struct {
		ti, di int
		iou    float64
	}
	var pairs []pair
	for ti, tk := range t.tracks {
		for di, d := range dets {
			if v := tk.Det.Box.IoU(d.Box); v >= t.iou {
				pairs = append(pairs, pair{ti, di, v})
			}
		}
	}
	sort.SliceStable(pairs, func(a, b int) bool { return pairs[a].iou > pairs[b].iou })

	trackUsed := make([]bool, len(t.tracks))
	detUsed := make([]bool, len(dets))
	for _, p := range pairs {
		if trackUsed[p.ti] || detUsed[p.di] {
			continue
		}
		trackUsed[p.ti], detUsed[p.di] = true, true
		tk := t.tracks[p.ti]
		tk.Det, tk.FramesSeen, tk.Missed = dets[p.di], tk.FramesSeen+1, 0
	}

	live := t.tracks[:0]
	for i, tk := range t.tracks {
		if !trackUsed[i] {
			tk.Missed++
		}
		if tk.Missed <= t.maxMissed {
			live = append(live, tk)
		}
	}
	t.tracks = live
	for di, d := range dets {
		if !detUsed[di] {
			t.tracks = append(t.tracks, &Track{ID: t.nextID, Det: d, FramesSeen: 1})
			t.nextID++
		}
	}

	out := make([]Track, len(t.tracks))
	for i, tk := range t.tracks {
		out[i] = *tk
	}
	sort.Slice(out, func(a, b int) bool { return out[a].ID < out[b].ID })
	return out
}
