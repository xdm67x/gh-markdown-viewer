package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/xdm67x/gh-markdown-viewer/internal/ghpr"
)

// Run initializes and executes the Bubble Tea program for gh-markdown-viewer.
func Run(info ghpr.PRInfo, files []ghpr.File) error {
	p := tea.NewProgram(
		New(info, files),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}
