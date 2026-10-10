package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/xdm67x/gh-markdown-viewer/internal/ghpr"
)

type threadsModal struct {
	path    string
	line    int
	threads []ghpr.ReviewThread
	cursor  int // if multiple threads exist on the same line
	yOffset int
}

func newThreadsModal(path string, line int, threads []ghpr.ReviewThread) threadsModal {
	return threadsModal{
		path:    path,
		line:    line,
		threads: threads,
		cursor:  0,
		yOffset: 0,
	}
}

func (m *threadsModal) activeThread() *ghpr.ReviewThread {
	if len(m.threads) == 0 {
		return nil
	}
	if m.cursor >= len(m.threads) {
		m.cursor = len(m.threads) - 1
	}
	return &m.threads[m.cursor]
}

func (m *threadsModal) scrollUp() {
	if m.yOffset > 0 {
		m.yOffset--
	}
}

func (m *threadsModal) scrollDown() {
	m.yOffset++
}

func (m *threadsModal) nextThread() {
	if m.cursor < len(m.threads)-1 {
		m.cursor++
		m.yOffset = 0
	}
}

func (m *threadsModal) prevThread() {
	if m.cursor > 0 {
		m.cursor--
		m.yOffset = 0
	}
}

func (m threadsModal) view(width, height int) string {
	var sb strings.Builder

	title := fmt.Sprintf("Review Thread · %s: line %d", m.path, m.line)
	if len(m.threads) > 1 {
		title += fmt.Sprintf(" (%d of %d threads)", m.cursor+1, len(m.threads))
	}
	sb.WriteString(modalTitleStyle.Render(title) + "\n\n")

	if len(m.threads) == 0 {
		sb.WriteString("No comments on this line.\n\n")
		sb.WriteString(modalFooterStyle.Render("esc close · c post new comment"))
		boxWidth := 60
		content := modalBoxStyle.Width(boxWidth).Render(sb.String())
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
	}

	th := m.activeThread()
	statusStyle := lipgloss.NewStyle().Bold(true)
	if th.Resolved {
		sb.WriteString(statusStyle.Foreground(lipgloss.Color("10")).Render("✓ Thread is Resolved") + "\n\n")
	} else {
		sb.WriteString(statusStyle.Foreground(lipgloss.Color("11")).Render("● Thread is Open / Unresolved") + "\n\n")
	}

	authorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	commentBodyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("254")).PaddingLeft(2)

	for _, c := range th.Comments {
		sb.WriteString(authorStyle.Render("@"+c.Author) + " " + timeStyle.Render(c.CreatedAt) + "\n")
		sb.WriteString(commentBodyStyle.Render(c.Body) + "\n\n")
	}

	resolveAction := "e resolve"
	if th.Resolved {
		resolveAction = "e unresolve"
	}

	footer := fmt.Sprintf("r reply · %s", resolveAction)
	if len(m.threads) > 1 {
		footer += " · [/] next/prev thread"
	}
	footer += " · esc close"

	sb.WriteString(modalFooterStyle.Render(footer))

	lines := strings.Split(sb.String(), "\n")
	maxVisible := height - 8
	if maxVisible < 10 {
		maxVisible = 10
	}

	if m.yOffset >= len(lines) {
		m.yOffset = len(lines) - 1
	}
	if m.yOffset < 0 {
		m.yOffset = 0
	}

	end := m.yOffset + maxVisible
	if end > len(lines) {
		end = len(lines)
	}

	visibleText := strings.Join(lines[m.yOffset:end], "\n")

	boxWidth := 74
	if width < 80 {
		boxWidth = width - 6
	}
	content := modalBoxStyle.Width(boxWidth).Render(visibleText)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
