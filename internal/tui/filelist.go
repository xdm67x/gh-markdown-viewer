package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xdm67x/gh-markdown-viewer/internal/ghpr"
)

// fileSelectedMsg is emitted when the user picks a file.
type fileSelectedMsg struct{ file ghpr.File }

// backToListMsg signals a return to the file picker.
type backToListMsg struct{}

// fileItem implements list.Item.
type fileItem struct{ file ghpr.File }

func (f fileItem) Title() string { return f.file.Filename }
func (f fileItem) Description() string {
	icon := statusIcon(f.file.Status)
	return fmt.Sprintf("%s  +%d -%d", icon, f.file.Additions, f.file.Deletions)
}
func (f fileItem) FilterValue() string { return f.file.Filename }

func statusIcon(s string) string {
	switch s {
	case "added":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("A")
	case "modified", "changed":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("M")
	case "renamed":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Render("R")
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("?")
	}
}

type listModel struct {
	l list.Model
}

func newListModel(files []ghpr.File) listModel {
	items := make([]list.Item, len(files))
	for i, f := range files {
		items[i] = fileItem{file: f}
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Markdown files changed in this PR"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Padding(0, 1)
	return listModel{l: l}
}

func (m *listModel) setSize(w, h int) {
	m.l.SetSize(w, h)
}

func (m listModel) update(msg tea.Msg) (listModel, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok && km.String() == "enter" {
		if item, ok := m.l.SelectedItem().(fileItem); ok {
			return m, func() tea.Msg { return fileSelectedMsg{file: item.file} }
		}
	}
	var cmd tea.Cmd
	m.l, cmd = m.l.Update(msg)
	return m, cmd
}

func (m listModel) view() string {
	return m.l.View()
}
