package srcmap

import (
	"strings"

	"github.com/charmbracelet/glamour"
)

// Span maps a range of source lines to a range of rendered (viewport) lines.
type Span struct {
	SrcStart  int // 1-based, inclusive
	SrcEnd    int // 1-based, inclusive
	RendStart int // 0-based, inclusive
	RendEnd   int // 0-based, inclusive
}

// Map is the compiled source↔rendered line mapping for one markdown document.
type Map struct {
	spans    []Span
	rendered string // full concatenated rendered output
}

// Build renders src block-by-block using Glamour and constructs the span table.
// width is the terminal width used for word-wrap.
func Build(src []byte, width int) (*Map, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil, err
	}

	blocks := ParseBlocks(src)
	var sb strings.Builder
	var spans []Span
	rendLine := 0

	for _, b := range blocks {
		out, err := renderer.Render(string(b.Source))
		if err != nil {
			// On render failure, emit the raw block so the viewer still works.
			out = string(b.Source) + "\n"
		}
		lineCount := strings.Count(out, "\n")
		spans = append(spans, Span{
			SrcStart:  b.SrcStart,
			SrcEnd:    b.SrcEnd,
			RendStart: rendLine,
			RendEnd:   rendLine + lineCount - 1,
		})
		sb.WriteString(out)
		rendLine += lineCount
	}

	return &Map{spans: spans, rendered: sb.String()}, nil
}

// Rendered returns the full rendered string to be loaded into the viewport.
func (m *Map) Rendered() string { return m.rendered }

// RenderedToSource maps a 0-based rendered line to a 1-based source line.
// Returns 0 if the line is outside all spans (e.g. blank padding lines).
func (m *Map) RenderedToSource(rendLine int) int {
	span, ok := m.findByRend(rendLine)
	if !ok {
		return 0
	}
	srcRange := span.SrcEnd - span.SrcStart
	rendRange := span.RendEnd - span.RendStart
	if rendRange <= 0 {
		return span.SrcStart
	}
	offset := rendLine - span.RendStart
	src := span.SrcStart + offset*srcRange/rendRange
	if src > span.SrcEnd {
		src = span.SrcEnd
	}
	return src
}

// SourceToRendered maps a 1-based source line to a 0-based rendered line
// (the first rendered line of the containing span).
func (m *Map) SourceToRendered(srcLine int) int {
	for _, s := range m.spans {
		if srcLine >= s.SrcStart && srcLine <= s.SrcEnd {
			return s.RendStart
		}
	}
	return 0
}

// TotalRenderedLines returns the total number of rendered lines.
func (m *Map) TotalRenderedLines() int {
	if len(m.spans) == 0 {
		return 0
	}
	return m.spans[len(m.spans)-1].RendEnd + 1
}

func (m *Map) findByRend(rendLine int) (Span, bool) {
	lo, hi := 0, len(m.spans)-1
	for lo <= hi {
		mid := (lo + hi) / 2
		s := m.spans[mid]
		switch {
		case rendLine < s.RendStart:
			hi = mid - 1
		case rendLine > s.RendEnd:
			lo = mid + 1
		default:
			return s, true
		}
	}
	return Span{}, false
}
