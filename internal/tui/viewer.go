package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xdm67x/gh-markdown-viewer/internal/ghpr"
	"github.com/xdm67x/gh-markdown-viewer/internal/srcmap"
)

// openInputMsg asks the root model to open the comment input overlay.
type openInputMsg struct{ srcLine int }

var (
	cursorStyle = lipgloss.NewStyle().Reverse(true)
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Padding(0, 1)
	commentableStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	notInDiffStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

type viewerModel struct {
	file            ghpr.File
	sm              *srcmap.Map
	commentable     map[int]bool
	vp              viewport.Model
	ready           bool
	cursorLine      int // 0-based rendered line
	totalRendLines  int
	renderedLines   []string // split of sm.Rendered() — rebuilt on cursor move
}

// newViewerModel loads file content, builds the srcmap, and initialises the viewport.
func newViewerModel(info ghpr.PRInfo, file ghpr.File, w, h int) (viewerModel, tea.Cmd) {
	return viewerModel{}, func() tea.Msg {
		content, err := ghpr.FileContent(info, file.Filename)
		if err != nil {
			content = fmt.Sprintf("*error loading file: %s*", err)
		}
		sm, err := srcmap.Build([]byte(content), w)
		if err != nil {
			sm, _ = srcmap.Build([]byte("*render error*"), w)
		}
		return fileLoadedMsg{
			file:        file,
			sm:          sm,
			commentable: ghpr.CommentableLines(file.Patch),
			w:           w,
			h:           h,
		}
	}
}

// fileLoadedMsg is received after async file fetch + srcmap build.
type fileLoadedMsg struct {
	file        ghpr.File
	sm          *srcmap.Map
	commentable map[int]bool
	w, h        int
}

func (m viewerModel) setSize(w, h int) viewerModel {
	m.vp.Width = w
	m.vp.Height = h - 2 // reserve 2 lines for status bar
	m.refreshContent()
	return m
}

func (m *viewerModel) refreshContent() {
	if !m.ready {
		return
	}
	lines := m.renderedLines
	if m.cursorLine >= 0 && m.cursorLine < len(lines) {
		lines = make([]string, len(m.renderedLines))
		copy(lines, m.renderedLines)
		lines[m.cursorLine] = cursorStyle.Render(m.renderedLines[m.cursorLine])
	}
	m.vp.SetContent(strings.Join(lines, "\n"))
	// Auto-scroll to keep cursor visible.
	if m.cursorLine < m.vp.YOffset {
		m.vp.SetYOffset(m.cursorLine)
	} else if m.cursorLine >= m.vp.YOffset+m.vp.Height {
		m.vp.SetYOffset(m.cursorLine - m.vp.Height + 1)
	}
}

func (m viewerModel) update(msg tea.Msg, statusMsg string) (viewerModel, tea.Cmd, string) {
	switch msg := msg.(type) {

	case fileLoadedMsg:
		vp := viewport.New(msg.w, msg.h-2)
		m.file = msg.file
		m.sm = msg.sm
		m.commentable = msg.commentable
		m.vp = vp
		m.ready = true
		m.totalRendLines = msg.sm.TotalRenderedLines()
		m.renderedLines = strings.Split(msg.sm.Rendered(), "\n")
		m.cursorLine = 0
		m.refreshContent()
		return m, nil, statusMsg

	case tea.KeyMsg:
		if !m.ready {
			return m, nil, statusMsg
		}
		switch msg.String() {
		case "up", "k":
			if m.cursorLine > 0 {
				m.cursorLine--
				m.refreshContent()
				statusMsg = ""
			}
			return m, nil, statusMsg
		case "down", "j":
			if m.cursorLine < m.totalRendLines-1 {
				m.cursorLine++
				m.refreshContent()
				statusMsg = ""
			}
			return m, nil, statusMsg
		case "g":
			m.cursorLine = 0
			m.refreshContent()
			return m, nil, statusMsg
		case "G":
			m.cursorLine = m.totalRendLines - 1
			m.refreshContent()
			return m, nil, statusMsg
		case "c":
			srcLine := m.sm.RenderedToSource(m.cursorLine)
			if srcLine == 0 || !m.commentable[srcLine] {
				return m, nil, notInDiffStyle.Render("✗ not in diff — cannot comment here")
			}
			return m, func() tea.Msg { return openInputMsg{srcLine: srcLine} }, statusMsg
		case "esc", "left", "backspace":
			return m, func() tea.Msg { return backToListMsg{} }, ""
		}
	}

	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd, statusMsg
}

func (m viewerModel) view(statusMsg string, width int) string {
	if !m.ready {
		return "Loading…"
	}

	srcLine := m.sm.RenderedToSource(m.cursorLine)
	var commentStatus string
	if srcLine > 0 && m.commentable[srcLine] {
		commentStatus = commentableStyle.Render(fmt.Sprintf("✓ line %d (in diff)", srcLine))
	} else if srcLine > 0 {
		commentStatus = notInDiffStyle.Render(fmt.Sprintf("✗ line %d (not in diff)", srcLine))
	}

	header := statusStyle.Render(fmt.Sprintf("%s · %s", m.file.Filename, commentStatus))
	footer := buildFooter(statusMsg, m.cursorLine+1, m.totalRendLines, width)

	return header + "\n" + m.vp.View() + "\n" + footer
}

func buildFooter(msg string, cursor, total, width int) string {
	hint := "j/k move  c comment  esc back  q quit"
	if msg != "" {
		hint = msg
	}
	progress := fmt.Sprintf("line %d/%d", cursor, total)
	padding := width - len([]rune(hint)) - len([]rune(progress)) - 2
	if padding < 1 {
		padding = 1
	}
	return statusStyle.Render(hint + strings.Repeat(" ", padding) + progress)
}
