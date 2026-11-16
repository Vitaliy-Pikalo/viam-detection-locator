package locator

import "testing"

func det(x0, y0, x1, y1 float64) Detection {
	return Detection{Box: Box{x0, y0, x1, y1}, Label: "person", Confidence: 0.9}
}

func TestIoU(t *testing.T) {
	a := Box{0, 0, 10, 10}
	if got := a.IoU(a); got != 1 {
		t.Fatalf("self IoU = %v, want 1", got)
	}
	if got := a.IoU(Box{20, 20, 30, 30}); got != 0 {
		t.Fatalf("disjoint IoU = %v, want 0", got)
	}
	if got := a.IoU(Box{5, 0, 15, 10}); !near(got, 1.0/3) {
		t.Fatalf("half-overlap IoU = %v, want 1/3", got)
	}
}

func TestTrackKeepsID(t *testing.T) {
	tr := NewTracker(0.3, 5)
	tr.Update([]Detection{det(0, 0, 10, 20)})
	got := tr.Update([]Detection{det(1, 0, 11, 20)})
	if len(got) != 1 || got[0].ID != 1 || got[0].FramesSeen != 2 {
		t.Fatalf("want one track id 1 seen twice, got %+v", got)
	}
}

func TestNewObjectGetsNewID(t *testing.T) {
	tr := NewTracker(0.3, 5)
	tr.Update([]Detection{det(0, 0, 10, 20)})
	got := tr.Update([]Detection{det(0, 0, 10, 20), det(100, 0, 110, 20)})
	if len(got) != 2 || got[0].ID != 1 || got[1].ID != 2 {
		t.Fatalf("want ids 1 and 2, got %+v", got)
	}
}

func TestTrackExpires(t *testing.T) {
	tr := NewTracker(0.3, 5)
	tr.Update([]Detection{det(0, 0, 10, 20)})
	for i := 0; i < 5; i++ {
		if got := tr.Update(nil); len(got) != 1 || got[0].Missed != i+1 {
			t.Fatalf("frame %d: track should survive with missed=%d, got %+v", i, i+1, got)
		}
	}
	if got := tr.Update(nil); len(got) != 0 {
		t.Fatalf("track should be dropped after 6 empty frames, got %+v", got)
	}
}

func TestGreedyHighestIoUWins(t *testing.T) {
	tr := NewTracker(0.3, 5)
	tr.Update([]Detection{det(0, 0, 10, 20), det(8, 0, 18, 20)}) // ids 1, 2
	// shifted copies: the 2nd detection overlaps track 2 best, the 1st overlaps track 1 best
	got := tr.Update([]Detection{det(9, 0, 19, 20), det(1, 0, 11, 20)})
	byID := map[int]Box{}
	for _, tk := range got {
		byID[tk.ID] = tk.Det.Box
	}
	if byID[1] != (Box{1, 0, 11, 20}) || byID[2] != (Box{9, 0, 19, 20}) || len(got) != 2 {
		t.Fatalf("ids swapped or lost: %+v", got)
	}
}
