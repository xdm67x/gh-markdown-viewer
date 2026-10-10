package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/xdm67x/gh-markdown-viewer/internal/diff"
	"github.com/xdm67x/gh-markdown-viewer/internal/ghpr"
)

type splitView struct {
	rows         []diff.SplitRow
	cursor       int
	yOffset      int
	threadMap    map[int][]ghpr.ReviewThread
	delThreadMap map[int][]ghpr.ReviewThread
}

func newSplitView(rows []diff.SplitRow, threadMap map[int][]ghpr.ReviewThread, delThreadMap map[int][]ghpr.ReviewThread) splitView {
	return splitView{
		rows:         rows,
		cursor:       0,
		yOffset:      0,
		threadMap:    threadMap,
		delThreadMap: delThreadMap,
	}
}

func (v *splitView) updateThreads(threadMap map[int][]ghpr.ReviewThread, delThreadMap map[int][]ghpr.ReviewThread) {
	v.threadMap = threadMap
	v.delThreadMap = delThreadMap
}

func (v *splitView) currentRow() *diff.SplitRow {
	if len(v.rows) == 0 || v.cursor < 0 || v.cursor >= len(v.rows) {
		return nil
	}
	return &v.rows[v.cursor]
}

func (v *splitView) moveDown(viewportHeight int) {
	if v.cursor < len(v.rows)-1 {
		v.cursor++
		v.adjustScroll(viewportHeight)
	}
}

func (v *splitView) moveUp(viewportHeight int) {
	if v.cursor > 0 {
		v.cursor--
		v.adjustScroll(viewportHeight)
	}
}

func (v *splitView) pageDown(viewportHeight int) {
	delta := viewportHeight / 2
	if delta < 1 {
		delta = 1
	}
	v.cursor += delta
	if v.cursor >= len(v.rows) {
		v.cursor = len(v.rows) - 1
	}
	v.adjustScroll(viewportHeight)
}

func (v *splitView) pageUp(viewportHeight int) {
	delta := viewportHeight / 2
	if delta < 1 {
		delta = 1
	}
	v.cursor -= delta
	if v.cursor < 0 {
		v.cursor = 0
	}
	v.adjustScroll(viewportHeight)
}

func (v *splitView) top(viewportHeight int) {
	v.cursor = 0
	v.adjustScroll(viewportHeight)
}

func (v *splitView) bottom(viewportHeight int) {
	if len(v.rows) > 0 {
		v.cursor = len(v.rows) - 1
		v.adjustScroll(viewportHeight)
	}
}

func (v *splitView) nextHunk(viewportHeight int) {
	for i := v.cursor + 1; i < len(v.rows); i++ {
		if v.rows[i].IsHeader {
			v.cursor = i
			v.adjustScroll(viewportHeight)
			return
		}
	}
}

func (v *splitView) prevHunk(viewportHeight int) {
	for i := v.cursor - 1; i >= 0; i-- {
		if v.rows[i].IsHeader {
			v.cursor = i
			v.adjustScroll(viewportHeight)
			return
		}
	}
}

func (v *splitView) adjustScroll(viewportHeight int) {
	if v.cursor < v.yOffset {
		v.yOffset = v.cursor
	} else if v.cursor >= v.yOffset+viewportHeight {
		v.yOffset = v.cursor - viewportHeight + 1
	}
}

func (v splitView) renderHalf(dl *diff.DiffLine, targetWidth int, isRight bool) string {
	if dl == nil {
		// Empty side (line added on other side, or line deleted on other side)
		return strings.Repeat(" ", targetWidth)
	}

	numStr := "    "
	if isRight && dl.NewLineNo > 0 {
		numStr = lineNumStyle.Render(fmt.Sprintf("%4d", dl.NewLineNo))
	} else if !isRight && dl.OldLineNo > 0 {
		numStr = lineNumStyle.Render(fmt.Sprintf("%4d", dl.OldLineNo))
	}

	prefix := " "
	if dl.Type == diff.LineAdded {
		prefix = "+"
	} else if dl.Type == diff.LineDeleted {
		prefix = "-"
	}

	var contentSB strings.Builder
	switch dl.Type {
	case diff.LineAdded:
		if len(dl.Segments) > 0 {
			for _, seg := range dl.Segments {
				if seg.Highlight {
					contentSB.WriteString(addedWordStyle.Render(seg.Text))
				} else {
					contentSB.WriteString(addedLineStyle.Render(seg.Text))
				}
			}
		} else {
			contentSB.WriteString(addedLineStyle.Render(dl.Content))
		}
	case diff.LineDeleted:
		if len(dl.Segments) > 0 {
			for _, seg := range dl.Segments {
				if seg.Highlight {
					contentSB.WriteString(deletedWordStyle.Render(seg.Text))
				} else {
					contentSB.WriteString(deletedLineStyle.Render(seg.Text))
				}
			}
		} else {
			contentSB.WriteString(deletedLineStyle.Render(dl.Content))
		}
	default:
		contentSB.WriteString(contextLineStyle.Render(dl.Content))
	}

	raw := fmt.Sprintf("%s %s %s", numStr, prefix, contentSB.String())
	return lipgloss.NewStyle().MaxWidth(targetWidth).Render(raw)
}

func (v splitView) view(width, height int) string {
	if len(v.rows) == 0 {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, "No diff changes in this file.")
	}

	dividerWidth := 3 // " │ "
	halfWidth := (width - dividerWidth) / 2
	if halfWidth < 20 {
		halfWidth = 20
	}

	var sb strings.Builder

	end := v.yOffset + height
	if end > len(v.rows) {
		end = len(v.rows)
	}

	for i := v.yOffset; i < end; i++ {
		row := v.rows[i]
		isCurrent := i == v.cursor

		if row.IsHeader {
			hdr := hunkHeaderStyle.Width(width).Render(row.HeaderLine)
			if isCurrent {
				hdr = cursorLineStyle.Width(width).Render(hdr)
			}
			sb.WriteString(hdr + "\n")
			continue
		}

		leftRendered := v.renderHalf(row.Left, halfWidth, false)
		rightRendered := v.renderHalf(row.Right, halfWidth, true)

		// Pad left side if needed
		leftLen := lipgloss.Width(leftRendered)
		if leftLen < halfWidth {
			leftRendered += strings.Repeat(" ", halfWidth-leftLen)
		}

		divider := " │ "
		cursorMark := " "
		if isCurrent {
			cursorMark = "▶"
		}

		commentMark := " "
		if row.Right != nil && len(v.threadMap[row.Right.NewLineNo]) > 0 {
			commentMark = "💬"
		} else if row.Left != nil && len(v.delThreadMap[row.Left.OldLineNo]) > 0 {
			commentMark = "💬"
		}

		combined := fmt.Sprintf("%s%s%s%s%s",
			cursorMark,
			leftRendered,
			splitDividerStyle.Render(divider),
			commentMark,
			rightRendered,
		)

		if isCurrent {
			combined = cursorLineStyle.Width(width).Render(combined)
		}

		sb.WriteString(combined + "\n")
	}

	linesRendered := end - v.yOffset
	for linesRendered < height {
		sb.WriteString("\n")
		linesRendered++
	}

	return sb.String()
}
