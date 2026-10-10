package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type commentMode int

const (
	commentModeNew commentMode = iota
	commentModeReply
)

type commentModal struct {
	mode      commentMode
	ta        textarea.Model
	path      string
	line      int
	side      string
	replyToID int64
}

func newCommentModal(path string, line int, side string) commentModal {
	ta := textarea.New()
	ta.Placeholder = "Write review comment (Markdown supported)…"
	ta.SetWidth(64)
	ta.SetHeight(6)
	ta.ShowLineNumbers = false
	ta.Focus()

	return commentModal{
		mode: commentModeNew,
		ta:   ta,
		path: path,
		line: line,
		side: side,
	}
}

func newReplyModal(path string, line int, replyToID int64) commentModal {
	ta := textarea.New()
	ta.Placeholder = "Write reply to this comment…"
	ta.SetWidth(64)
	ta.SetHeight(5)
	ta.ShowLineNumbers = false
	ta.Focus()

	return commentModal{
		mode:      commentModeReply,
		ta:        ta,
		path:      path,
		line:      line,
		replyToID: replyToID,
	}
}

func (m *commentModal) setSize(width int) {
	taWidth := width - 12
	if taWidth > 74 {
		taWidth = 74
	}
	if taWidth < 40 {
		taWidth = 40
	}
	m.ta.SetWidth(taWidth)
}

func (m commentModal) update(msg tea.Msg) (commentModal, tea.Cmd) {
	var cmd tea.Cmd
	m.ta, cmd = m.ta.Update(msg)
	return m, cmd
}

func (m commentModal) value() string {
	return strings.TrimSpace(m.ta.Value())
}

func (m commentModal) view(width, height int) string {
	var sb strings.Builder

	var title string
	if m.mode == commentModeNew {
		title = fmt.Sprintf("Post Review Comment · %s:%d (%s)", m.path, m.line, m.side)
	} else {
		title = fmt.Sprintf("Reply to Thread · %s:%d", m.path, m.line)
	}

	sb.WriteString(modalTitleStyle.Render(title) + "\n\n")
	sb.WriteString(m.ta.View() + "\n\n")
	sb.WriteString(modalFooterStyle.Render("ctrl+s submit · esc cancel"))

	boxWidth := m.ta.Width() + 4
	content := modalBoxStyle.Width(boxWidth).Render(sb.String())

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
