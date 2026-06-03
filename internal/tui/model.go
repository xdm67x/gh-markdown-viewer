// Package tui contains the Bubble Tea program for gh-markdown-viewer.
package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xdm67x/gh-markdown-viewer/internal/ghpr"
)

type mode int

const (
	modeList   mode = iota // file picker screen
	modeViewer             // markdown viewer screen
	modeInput              // comment text-input overlay
)

// commentPostedMsg is sent after a review comment API call finishes.
type commentPostedMsg struct{ err error }

// Model is the root Bubble Tea model.
type Model struct {
	info ghpr.PRInfo

	// Screens
	list   listModel
	viewer viewerModel
	input  inputModel

	mode mode

	// Status bar message (transient feedback shown for one render cycle).
	statusMsg string
	width     int
	height    int

	quitting bool
}

// New creates the root model. files is the list of changed .md files.
func New(info ghpr.PRInfo, files []ghpr.File) Model {
	return Model{
		info: info,
		list: newListModel(files),
		mode: modeList,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.setSize(msg.Width, msg.Height)
		if m.viewer.ready {
			m.viewer = m.viewer.setSize(msg.Width, msg.Height)
		}
		return m, nil

	case tea.KeyMsg:
		// Global quit binds
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		if m.mode == modeList && msg.String() == "q" {
			m.quitting = true
			return m, tea.Quit
		}
		if m.mode == modeViewer && msg.String() == "q" {
			m.quitting = true
			return m, tea.Quit
		}

	case fileSelectedMsg:
		// Transition from file picker to viewer.
		v, cmd := newViewerModel(m.info, msg.file, m.width, m.height)
		m.viewer = v
		m.mode = modeViewer
		m.statusMsg = ""
		return m, cmd

	case openInputMsg:
		m.input = newInputModel(msg.srcLine)
		m.mode = modeInput
		return m, m.input.ti.Focus()

	case cancelInputMsg:
		m.mode = modeViewer
		return m, nil

	case submitCommentMsg:
		m.mode = modeViewer
		cmd := postCommentCmd(m.info, m.viewer.file, msg.body, msg.srcLine)
		return m, cmd

	case commentPostedMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("error: %s", msg.err)
		} else {
			m.statusMsg = "comment posted"
		}
		return m, nil

	case backToListMsg:
		m.mode = modeList
		m.statusMsg = ""
		return m, nil
	}

	// Delegate to active screen.
	var cmd tea.Cmd
	switch m.mode {
	case modeList:
		var updated listModel
		updated, cmd = m.list.update(msg)
		m.list = updated
	case modeViewer:
		var updated viewerModel
		updated, cmd, m.statusMsg = m.viewer.update(msg, m.statusMsg)
		m.viewer = updated
	case modeInput:
		var updated inputModel
		updated, cmd = m.input.update(msg)
		m.input = updated
	}
	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}
	switch m.mode {
	case modeList:
		return m.list.view()
	case modeViewer:
		return m.viewer.view(m.statusMsg, m.width)
	case modeInput:
		return m.viewer.view("", m.width) + "\n" + m.input.view(m.width)
	}
	return ""
}

// postCommentCmd returns a Cmd that posts a review comment asynchronously.
func postCommentCmd(info ghpr.PRInfo, file ghpr.File, body string, srcLine int) tea.Cmd {
	return func() tea.Msg {
		err := ghpr.PostLineComment(info, file.Filename, body, srcLine)
		return commentPostedMsg{err: err}
	}
}
