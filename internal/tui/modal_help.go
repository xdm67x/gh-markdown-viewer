package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type helpModal struct {
	width  int
	height int
}

func newHelpModal() helpModal {
	return helpModal{}
}

func (h helpModal) view(width, height int) string {
	var sb strings.Builder

	sb.WriteString(modalTitleStyle.Render("Keyboard Shortcuts & Help") + "\n\n")

	categoryStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")).Width(14)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	renderRow := func(key, desc string) {
		sb.WriteString("  " + keyStyle.Render(key) + descStyle.Render(desc) + "\n")
	}

	sb.WriteString(categoryStyle.Render("Navigation") + "\n")
	renderRow("j / ↓", "Move cursor down one line")
	renderRow("k / ↑", "Move cursor up one line")
	renderRow("d / ctrl+d", "Scroll half page down")
	renderRow("u / ctrl+u", "Scroll half page up")
	renderRow("g / G", "Jump to top / bottom")
	renderRow("n / ]", "Jump to next diff hunk")
	renderRow("p / [", "Jump to previous diff hunk")
	sb.WriteString("\n")

	sb.WriteString(categoryStyle.Render("Views") + "\n")
	renderRow("1", "Styled Unified Diff (with intra-line word diffs)")
	renderRow("2", "Side-by-Side Split Diff (Base vs Head)")
	renderRow("3", "Full Rendered Document (Terminal Markdown via Glamour)")
	renderRow("4 / s", "Section Graph & Hierarchy Navigator (interactive tree)")
	renderRow("Tab", "Cycle through view modes")
	renderRow("Enter", "Jump from Graph node to section in Diff view")
	renderRow("i", "AI Intent & Change Summary (sections & checklists)")
	renderRow("f / b", "Switch file (File picker)")
	sb.WriteString("\n")

	sb.WriteString(categoryStyle.Render("Reviews & Comments") + "\n")
	renderRow("c", "Post review comment on current line (if in diff)")
	renderRow("t / Enter", "View thread / comments on current line")
	renderRow("r", "Reply to thread (inside thread viewer)")
	renderRow("e", "Toggle resolve / unresolve thread")
	sb.WriteString("\n")

	sb.WriteString(categoryStyle.Render("General") + "\n")
	renderRow("?", "Toggle this help overlay")
	renderRow("q / ctrl+c", "Quit")
	sb.WriteString("\n")

	sb.WriteString(modalFooterStyle.Render("Press esc or ? to return to diff view"))

	boxWidth := 64
	if width < 70 {
		boxWidth = width - 4
	}
	content := modalBoxStyle.Width(boxWidth).Render(sb.String())

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
