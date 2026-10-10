package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/xdm67x/gh-markdown-viewer/internal/diff"
	"github.com/xdm67x/gh-markdown-viewer/internal/ghpr"
)

var (
	mdHeadingPattern = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	chkBoxPattern    = regexp.MustCompile(`^([\s\-\*\+]*\[)([ xX])(\]\s*)(.*)$`)
)

type unifiedView struct {
	lines        []diff.DiffLine
	cursor       int
	yOffset      int
	threadMap    map[int][]ghpr.ReviewThread // newLineNo -> threads
	delThreadMap map[int][]ghpr.ReviewThread // oldLineNo -> threads
}

func newUnifiedView(lines []diff.DiffLine, threadMap map[int][]ghpr.ReviewThread, delThreadMap map[int][]ghpr.ReviewThread) unifiedView {
	return unifiedView{
		lines:        lines,
		cursor:       0,
		yOffset:      0,
		threadMap:    threadMap,
		delThreadMap: delThreadMap,
	}
}

func (v *unifiedView) updateThreads(threadMap map[int][]ghpr.ReviewThread, delThreadMap map[int][]ghpr.ReviewThread) {
	v.threadMap = threadMap
	v.delThreadMap = delThreadMap
}

func (v *unifiedView) currentLine() *diff.DiffLine {
	if len(v.lines) == 0 || v.cursor < 0 || v.cursor >= len(v.lines) {
		return nil
	}
	return &v.lines[v.cursor]
}

func (v *unifiedView) moveDown(viewportHeight int) {
	if v.cursor < len(v.lines)-1 {
		v.cursor++
		v.adjustScroll(viewportHeight)
	}
}

func (v *unifiedView) moveUp(viewportHeight int) {
	if v.cursor > 0 {
		v.cursor--
		v.adjustScroll(viewportHeight)
	}
}

func (v *unifiedView) pageDown(viewportHeight int) {
	delta := viewportHeight / 2
	if delta < 1 {
		delta = 1
	}
	v.cursor += delta
	if v.cursor >= len(v.lines) {
		v.cursor = len(v.lines) - 1
	}
	v.adjustScroll(viewportHeight)
}

func (v *unifiedView) pageUp(viewportHeight int) {
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

func (v *unifiedView) top(viewportHeight int) {
	v.cursor = 0
	v.adjustScroll(viewportHeight)
}

func (v *unifiedView) bottom(viewportHeight int) {
	if len(v.lines) > 0 {
		v.cursor = len(v.lines) - 1
		v.adjustScroll(viewportHeight)
	}
}

func (v *unifiedView) nextHunk(viewportHeight int) {
	for i := v.cursor + 1; i < len(v.lines); i++ {
		if v.lines[i].Type == diff.LineHunkHeader {
			v.cursor = i
			v.adjustScroll(viewportHeight)
			return
		}
	}
}

func (v *unifiedView) prevHunk(viewportHeight int) {
	for i := v.cursor - 1; i >= 0; i-- {
		if v.lines[i].Type == diff.LineHunkHeader {
			v.cursor = i
			v.adjustScroll(viewportHeight)
			return
		}
	}
}

func (v *unifiedView) adjustScroll(viewportHeight int) {
	if v.cursor < v.yOffset {
		v.yOffset = v.cursor
	} else if v.cursor >= v.yOffset+viewportHeight {
		v.yOffset = v.cursor - viewportHeight + 1
	}
}

func (v unifiedView) renderLineContent(dl diff.DiffLine, isCurrent bool) string {
	switch dl.Type {
	case diff.LineHunkHeader:
		return hunkHeaderStyle.Render(dl.Content)

	case diff.LineAdded:
		var sb strings.Builder
		sb.WriteString(addedPrefixStyle.Render("+ "))

		// Check for markdown heading
		if m := mdHeadingPattern.FindStringSubmatch(dl.Content); m != nil {
			lvl := len(m[1])
			headingColor := "82"
			if lvl == 1 {
				headingColor = "46"
			}
			hStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(headingColor))
			return sb.String() + hStyle.Render(dl.Content)
		}

		// Check for checklist
		if m := chkBoxPattern.FindStringSubmatch(dl.Content); m != nil {
			chkState := strings.ToLower(m[2]) == "x"
			chkIcon := "[ ]"
			if chkState {
				chkIcon = "[✓]"
			}
			chkStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("46"))
			sb.WriteString(chkStyle.Render(chkIcon) + " " + addedLineStyle.Render(m[4]))
			return sb.String()
		}

		if len(dl.Segments) > 0 {
			for _, seg := range dl.Segments {
				if seg.Highlight {
					sb.WriteString(addedWordStyle.Render(seg.Text))
				} else {
					sb.WriteString(addedLineStyle.Render(seg.Text))
				}
			}
		} else {
			sb.WriteString(addedLineStyle.Render(dl.Content))
		}
		return sb.String()

	case diff.LineDeleted:
		var sb strings.Builder
		sb.WriteString(deletedPrefixStyle.Render("- "))

		if m := mdHeadingPattern.FindStringSubmatch(dl.Content); m != nil {
			hStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("203")).Strikethrough(true)
			return sb.String() + hStyle.Render(dl.Content)
		}

		if len(dl.Segments) > 0 {
			for _, seg := range dl.Segments {
				if seg.Highlight {
					sb.WriteString(deletedWordStyle.Render(seg.Text))
				} else {
					sb.WriteString(deletedLineStyle.Render(seg.Text))
				}
			}
		} else {
			sb.WriteString(deletedLineStyle.Render(dl.Content))
		}
		return sb.String()

	case diff.LineContext:
		var sb strings.Builder
		sb.WriteString("  ")
		if m := mdHeadingPattern.FindStringSubmatch(dl.Content); m != nil {
			hStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
			sb.WriteString(hStyle.Render(dl.Content))
		} else if m := chkBoxPattern.FindStringSubmatch(dl.Content); m != nil {
			chkState := strings.ToLower(m[2]) == "x"
			chkIcon := "[ ]"
			if chkState {
				chkIcon = "[✓]"
			}
			chkStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250"))
			if chkState {
				chkStyle = chkStyle.Foreground(lipgloss.Color("78"))
			}
			sb.WriteString(chkStyle.Render(chkIcon) + " " + contextLineStyle.Render(m[4]))
		} else {
			sb.WriteString(contextLineStyle.Render(dl.Content))
		}
		return sb.String()
	}

	return dl.Content
}

func (v unifiedView) view(width, height int) string {
	if len(v.lines) == 0 {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, "No diff changes in this file.")
	}

	var sb strings.Builder

	end := v.yOffset + height
	if end > len(v.lines) {
		end = len(v.lines)
	}

	for i := v.yOffset; i < end; i++ {
		dl := v.lines[i]
		isCurrent := i == v.cursor

		// Cursor indicator
		cursorTag := "  "
		if isCurrent {
			cursorTag = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")).Render("▶ ")
		}

		// Comment indicator
		threadsCount := 0
		if dl.NewLineNo > 0 {
			threadsCount = len(v.threadMap[dl.NewLineNo])
		} else if dl.OldLineNo > 0 {
			threadsCount = len(v.delThreadMap[dl.OldLineNo])
		}

		commentTag := "   "
		if threadsCount > 0 {
			commentTag = commentBadgeStyle.Render("💬 ")
		}

		// Line numbers
		oldNumStr := "    "
		if dl.OldLineNo > 0 {
			oldNumStr = lineNumStyle.Render(fmt.Sprintf("%4d", dl.OldLineNo))
		}
		newNumStr := "    "
		if dl.NewLineNo > 0 {
			newNumStr = lineNumStyle.Render(fmt.Sprintf("%4d", dl.NewLineNo))
		}

		gutterSep := gutterSepStyle.Render("│ ")

		lineContent := v.renderLineContent(dl, isCurrent)

		fullLine := fmt.Sprintf("%s%s%s %s %s%s",
			cursorTag,
			commentTag,
			oldNumStr,
			newNumStr,
			gutterSep,
			lineContent,
		)

		if isCurrent {
			fullLine = cursorLineStyle.Width(width).Render(fullLine)
		}

		sb.WriteString(fullLine + "\n")
	}

	// Pad remaining height
	linesRendered := end - v.yOffset
	for linesRendered < height {
		sb.WriteString("\n")
		linesRendered++
	}

	return sb.String()
}
