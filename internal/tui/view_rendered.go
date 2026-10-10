package tui

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type renderedView struct {
	rawContent    string
	renderedLines []string
	cursor        int
	yOffset       int
	width         int
	err           error
}

func newRenderedView(content string, width int) renderedView {
	v := renderedView{
		rawContent: content,
		width:      width,
	}
	v.render(width)
	return v
}

func (v *renderedView) setContent(content string, width int) {
	v.rawContent = content
	v.cursor = 0
	v.yOffset = 0
	v.render(width)
}

func (v *renderedView) render(width int) {
	if strings.TrimSpace(v.rawContent) == "" {
		v.renderedLines = []string{"(Empty document)"}
		return
	}

	wrapWidth := width - 4
	if wrapWidth < 40 {
		wrapWidth = 40
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(wrapWidth),
	)
	if err != nil {
		v.err = err
		v.renderedLines = strings.Split(v.rawContent, "\n")
		return
	}

	out, err := renderer.Render(v.rawContent)
	if err != nil {
		v.err = err
		v.renderedLines = strings.Split(v.rawContent, "\n")
		return
	}

	// Clean trailing empty lines
	trimmed := strings.TrimRight(out, "\n")
	v.renderedLines = strings.Split(trimmed, "\n")
}

func (v *renderedView) moveDown(viewportHeight int) {
	if v.cursor < len(v.renderedLines)-1 {
		v.cursor++
		v.adjustScroll(viewportHeight)
	}
}

func (v *renderedView) moveUp(viewportHeight int) {
	if v.cursor > 0 {
		v.cursor--
		v.adjustScroll(viewportHeight)
	}
}

func (v *renderedView) pageDown(viewportHeight int) {
	delta := viewportHeight / 2
	if delta < 1 {
		delta = 1
	}
	v.cursor += delta
	if v.cursor >= len(v.renderedLines) {
		v.cursor = len(v.renderedLines) - 1
	}
	v.adjustScroll(viewportHeight)
}

func (v *renderedView) pageUp(viewportHeight int) {
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

func (v *renderedView) top(viewportHeight int) {
	v.cursor = 0
	v.adjustScroll(viewportHeight)
}

func (v *renderedView) bottom(viewportHeight int) {
	if len(v.renderedLines) > 0 {
		v.cursor = len(v.renderedLines) - 1
		v.adjustScroll(viewportHeight)
	}
}

func (v *renderedView) adjustScroll(viewportHeight int) {
	if v.cursor < v.yOffset {
		v.yOffset = v.cursor
	} else if v.cursor >= v.yOffset+viewportHeight {
		v.yOffset = v.cursor - viewportHeight + 1
	}
}

func (v renderedView) view(width, height int) string {
	if len(v.renderedLines) == 0 {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, "No content to render.")
	}

	var sb strings.Builder

	end := v.yOffset + height
	if end > len(v.renderedLines) {
		end = len(v.renderedLines)
	}

	cursorTag := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")).Render("▶ ")

	for i := v.yOffset; i < end; i++ {
		line := v.renderedLines[i]
		isCurrent := i == v.cursor

		prefix := "  "
		if isCurrent {
			prefix = cursorTag
		}

		fullLine := prefix + line
		if isCurrent {
			fullLine = cursorLineStyle.Width(width).Render(fullLine)
		}
		sb.WriteString(fullLine + "\n")
	}

	linesRendered := end - v.yOffset
	for linesRendered < height {
		sb.WriteString("\n")
		linesRendered++
	}

	return sb.String()
}
