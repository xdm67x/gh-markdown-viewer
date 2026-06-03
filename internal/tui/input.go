package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// submitCommentMsg is emitted when the user confirms a comment.
type submitCommentMsg struct {
	body    string
	srcLine int
}

// cancelInputMsg is emitted when the user cancels input.
type cancelInputMsg struct{}

var (
	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("4")).
			Padding(0, 1)
	inputHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
)

type inputModel struct {
	ti      textarea.Model
	srcLine int
}

func newInputModel(srcLine int) inputModel {
	ti := textarea.New()
	ti.Placeholder = fmt.Sprintf("Comment on line %d…", srcLine)
	ti.SetWidth(60)
	ti.SetHeight(5)
	ti.ShowLineNumbers = false
	ti.KeyMap.InsertNewline.SetEnabled(true)
	return inputModel{ti: ti, srcLine: srcLine}
}

func (m inputModel) update(msg tea.Msg) (inputModel, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "esc":
			return m, func() tea.Msg { return cancelInputMsg{} }
		case "ctrl+s":
			body := strings.TrimSpace(m.ti.Value())
			if body == "" {
				return m, func() tea.Msg { return cancelInputMsg{} }
			}
			return m, func() tea.Msg {
				return submitCommentMsg{body: body, srcLine: m.srcLine}
			}
		}
	}
	var cmd tea.Cmd
	m.ti, cmd = m.ti.Update(msg)
	return m, cmd
}

func (m inputModel) view(width int) string {
	hint := inputHintStyle.Render("ctrl+s submit  esc cancel")
	box := inputBoxStyle.Render(m.ti.View())
	// Center the box horizontally.
	boxWidth := 60 + 4 // textarea width + border + padding
	left := (width - boxWidth) / 2
	if left < 0 {
		left = 0
	}
	pad := strings.Repeat(" ", left)
	return pad + box + "\n" + pad + hint
}
