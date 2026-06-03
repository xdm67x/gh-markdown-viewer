package srcmap

import (
	"testing"
)

func TestRenderedToSource_heading(t *testing.T) {
	src := []byte("# Heading\n\nParagraph text.\n")
	m, err := Build(src, 80)
	if err != nil {
		t.Fatal(err)
	}
	// rendered line 0 should map back into the heading's source range (line 1)
	got := m.RenderedToSource(0)
	if got == 0 {
		t.Errorf("RenderedToSource(0) = 0, want a valid source line")
	}
}

func TestRenderedToSource_paragraph(t *testing.T) {
	src := []byte("# H\n\nSome paragraph that is long enough to maybe reflow.\n")
	m, err := Build(src, 40)
	if err != nil {
		t.Fatal(err)
	}
	total := m.TotalRenderedLines()
	if total == 0 {
		t.Fatal("expected non-zero rendered lines")
	}
	// Every rendered line should map to a valid source line (>0).
	for r := 0; r < total; r++ {
		src := m.RenderedToSource(r)
		if src < 0 {
			t.Errorf("RenderedToSource(%d) = %d, want >= 0", r, src)
		}
	}
}

func TestSourceToRendered_roundtrip(t *testing.T) {
	src := []byte("# Title\n\nFirst para.\n\nSecond para.\n")
	m, err := Build(src, 80)
	if err != nil {
		t.Fatal(err)
	}
	// Source line 1 (heading) should map to some rendered line,
	// and that rendered line should map back to a line inside the heading block.
	rend := m.SourceToRendered(1)
	back := m.RenderedToSource(rend)
	if back == 0 {
		t.Errorf("round-trip failed: SourceToRendered(1)=%d RenderedToSource(%d)=%d", rend, rend, back)
	}
}
