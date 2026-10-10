package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/xdm67x/gh-markdown-viewer/internal/ghpr"
)

type filesModal struct {
	files       []ghpr.File
	cursor      int
	threadCount map[string]int // path -> number of comments
}

func newFilesModal(files []ghpr.File, threadCount map[string]int) filesModal {
	return filesModal{
		files:       files,
		cursor:      0,
		threadCount: threadCount,
	}
}

func (m *filesModal) updateFiles(files []ghpr.File, threadCount map[string]int) {
	m.files = files
	m.threadCount = threadCount
	if m.cursor >= len(files) && len(files) > 0 {
		m.cursor = len(files) - 1
	}
}

func (m *filesModal) moveUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

func (m *filesModal) moveDown() {
	if m.cursor < len(m.files)-1 {
		m.cursor++
	}
}

func (m filesModal) selectedIndex() int {
	return m.cursor
}

func (m filesModal) view(width, height int) string {
	var sb strings.Builder

	title := fmt.Sprintf("Changed Markdown Files (%d files)", len(m.files))
	sb.WriteString(modalTitleStyle.Render(title) + "\n\n")

	cursorPointer := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	selectedItem := lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("238"))
	statStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	commentCountStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))

	for i, f := range m.files {
		pointer := "  "
		if i == m.cursor {
			pointer = cursorPointer.Render("▶ ")
		}

		var badge string
		switch f.Status {
		case "added":
			badge = badgeAddStyle.Render()
		case "modified", "changed":
			badge = badgeModStyle.Render()
		case "renamed":
			badge = badgeRenStyle.Render()
		case "removed":
			badge = badgeDelStyle.Render()
		default:
			badge = f.Status
		}

		stats := statStyle.Render(fmt.Sprintf("+%d -%d", f.Additions, f.Deletions))
		comments := ""
		if count := m.threadCount[f.Filename]; count > 0 {
			comments = commentCountStyle.Render(fmt.Sprintf(" 💬 %d", count))
		}

		line := fmt.Sprintf("%s%s %s  %s%s", pointer, badge, f.Filename, stats, comments)
		if i == m.cursor {
			line = selectedItem.Render(line)
		}
		sb.WriteString(line + "\n")
	}

	sb.WriteString("\n" + modalFooterStyle.Render("↑/k up · ↓/j down · enter select · esc close"))

	boxWidth := 70
	if width < 76 {
		boxWidth = width - 6
	}
	content := modalBoxStyle.Width(boxWidth).Render(sb.String())

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
