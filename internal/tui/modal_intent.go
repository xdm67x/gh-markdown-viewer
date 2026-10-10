package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/xdm67x/gh-markdown-viewer/internal/diff"
)

type intentModal struct {
	intent  diff.IntentSummary
	yOffset int
	height  int
}

func newIntentModal(intent diff.IntentSummary) intentModal {
	return intentModal{
		intent: intent,
	}
}

func (m *intentModal) setIntent(intent diff.IntentSummary) {
	m.intent = intent
	m.yOffset = 0
}

func (m *intentModal) scrollUp() {
	if m.yOffset > 0 {
		m.yOffset--
	}
}

func (m *intentModal) scrollDown() {
	m.yOffset++
}

func (m intentModal) view(width, height int) string {
	var sb strings.Builder

	title := "AI Intent & Changes Summary"
	sb.WriteString(modalTitleStyle.Render(title) + "\n\n")

	statsStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255"))
	addStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("78"))
	delStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("167"))
	sectionHeadStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	subheadStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))

	overview := fmt.Sprintf("%s  %s  %s  %s",
		statsStyle.Render(fmt.Sprintf("Hunks: %d", m.intent.HunkCount)),
		addStyle.Render(fmt.Sprintf("+%d additions", m.intent.TotalAdditions)),
		delStyle.Render(fmt.Sprintf("-%d deletions", m.intent.TotalDeletions)),
		statsStyle.Render(fmt.Sprintf("Net: %+d", m.intent.TotalAdditions-m.intent.TotalDeletions)),
	)
	sb.WriteString(overview + "\n\n")

	// Sections modified
	sb.WriteString(sectionHeadStyle.Render("Modified Sections & Structure:") + "\n")
	if len(m.intent.Sections) == 0 {
		sb.WriteString("  (No markdown headings directly modified)\n")
	} else {
		for _, sec := range m.intent.Sections {
			indent := strings.Repeat("  ", sec.Level)
			stats := fmt.Sprintf(" %s %s",
				addStyle.Render(fmt.Sprintf("+%d", sec.Additions)),
				delStyle.Render(fmt.Sprintf("-%d", sec.Deletions)),
			)
			sb.WriteString(fmt.Sprintf("%s• %s%s\n", indent, sec.Heading, stats))
		}
	}
	sb.WriteString("\n")

	// Task lists & Checklists
	sb.WriteString(sectionHeadStyle.Render("Task & Checklist Status:") + "\n")
	if len(m.intent.Checklists) == 0 {
		sb.WriteString("  (No task list checkboxes modified)\n")
	} else {
		for _, chk := range m.intent.Checklists {
			switch chk.Status {
			case "completed":
				sb.WriteString(fmt.Sprintf("  %s %s\n",
					addStyle.Render("[✓ Completed]"),
					chk.Task,
				))
			case "unchecked":
				sb.WriteString(fmt.Sprintf("  %s %s\n",
					delStyle.Render("[✗ Reopened]"),
					chk.Task,
				))
			case "added":
				sb.WriteString(fmt.Sprintf("  %s %s\n",
					subheadStyle.Render("[+ New Task]"),
					chk.Task,
				))
			case "removed":
				sb.WriteString(fmt.Sprintf("  %s %s\n",
					delStyle.Render("[- Removed Task]"),
					chk.Task,
				))
			}
		}
	}
	sb.WriteString("\n")

	sb.WriteString(modalFooterStyle.Render("j/↓ scroll down · k/↑ scroll up · esc close"))

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
