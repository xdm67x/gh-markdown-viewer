package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xdm67x/gh-markdown-viewer/internal/diff"
	"github.com/xdm67x/gh-markdown-viewer/internal/ghpr"
	"github.com/xdm67x/gh-markdown-viewer/internal/graph"
)

type activeMode int

const (
	modeUnified activeMode = iota
	modeSplit
	modeRendered
	modeGraph
)

type activeModal int

const (
	modalNone activeModal = iota
	modalHelp
	modalFiles
	modalIntent
	modalComment
	modalThreads
)

// Async messages
type fileLoadedMsg struct {
	file        ghpr.File
	fileDiff    *diff.FileDiff
	headContent string
	baseContent string
	err         error
}

type threadsLoadedMsg struct {
	threads []ghpr.ReviewThread
	err     error
}

type commentSubmittedMsg struct {
	err error
}

type threadResolvedMsg struct {
	nodeID   string
	resolved bool
	err      error
}

type replySubmittedMsg struct {
	err error
}

// Model represents the root Bubble Tea application state.
type Model struct {
	info           ghpr.PRInfo
	files          []ghpr.File
	currentFileIdx int
	activeMode     activeMode
	activeModal    activeModal

	// File caches
	fileDiffs   map[string]*diff.FileDiff
	headContent map[string]string
	baseContent map[string]string

	// Threads
	threads      []ghpr.ReviewThread
	threadMap    map[string]map[int][]ghpr.ReviewThread // file -> newLine -> threads
	delThreadMap map[string]map[int][]ghpr.ReviewThread // file -> oldLine -> threads
	threadCount  map[string]int                         // file -> count

	// Views
	unified      unifiedView
	split        splitView
	rendered     renderedView
	graphV       graphView
	lastDiffMode activeMode

	// Graph cache
	graphMap map[string]*graph.Graph

	// Modals
	help    helpModal
	filesM  filesModal
	intentM intentModal
	commM   commentModal
	thM     threadsModal

	// UI Dimensions
	width  int
	height int

	// Status Feedback
	toastMsg     string
	toastIsError bool

	loading    bool
	loadingMsg string
	quitting   bool
}

// New creates and initializes the root Model.
func New(info ghpr.PRInfo, files []ghpr.File) Model {
	m := Model{
		info:           info,
		files:          files,
		currentFileIdx: 0,
		activeMode:     modeUnified,
		lastDiffMode:   modeUnified,
		activeModal:    modalNone,
		fileDiffs:      make(map[string]*diff.FileDiff),
		headContent:    make(map[string]string),
		baseContent:    make(map[string]string),
		threadMap:      make(map[string]map[int][]ghpr.ReviewThread),
		delThreadMap:   make(map[string]map[int][]ghpr.ReviewThread),
		threadCount:    make(map[string]int),
		graphMap:       make(map[string]*graph.Graph),
		help:           newHelpModal(),
		filesM:         newFilesModal(files, make(map[string]int)),
		loading:        true,
		loadingMsg:     "Loading PR markdown diffs…",
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadFileCmd(m.info, m.files[m.currentFileIdx]),
		fetchThreadsCmd(m.info),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.commM.mode != 0 || m.commM.path != "" {
			m.commM.setSize(msg.Width)
		}
		return m, nil

	case fileLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.toastMsg = fmt.Sprintf("Error loading file: %s", msg.err)
			m.toastIsError = true
			return m, nil
		}

		m.fileDiffs[msg.file.Filename] = msg.fileDiff
		m.headContent[msg.file.Filename] = msg.headContent
		m.baseContent[msg.file.Filename] = msg.baseContent

		tRight := m.threadMap[msg.file.Filename]
		if tRight == nil {
			tRight = make(map[int][]ghpr.ReviewThread)
		}
		tLeft := m.delThreadMap[msg.file.Filename]
		if tLeft == nil {
			tLeft = make(map[int][]ghpr.ReviewThread)
		}

		m.unified = newUnifiedView(msg.fileDiff.UnifiedLines, tRight, tLeft)
		m.split = newSplitView(msg.fileDiff.SplitRows, tRight, tLeft)
		m.rendered = newRenderedView(msg.headContent, m.width)
		m.intentM.setIntent(msg.fileDiff.Intent)

		g := graph.BuildSectionGraph(msg.headContent, msg.fileDiff.Hunks, tRight)
		m.graphMap[msg.file.Filename] = g
		m.graphV = newGraphView(g)

		return m, nil

	case threadsLoadedMsg:
		if msg.err == nil {
			m.threads = msg.threads
			m.rebuildThreadIndex()
			currentFile := m.files[m.currentFileIdx].Filename
			m.unified.updateThreads(m.threadMap[currentFile], m.delThreadMap[currentFile])
			m.split.updateThreads(m.threadMap[currentFile], m.delThreadMap[currentFile])
			m.filesM.updateFiles(m.files, m.threadCount)

			head := m.headContent[currentFile]
			if fd, ok := m.fileDiffs[currentFile]; ok && head != "" {
				g := graph.BuildSectionGraph(head, fd.Hunks, m.threadMap[currentFile])
				m.graphMap[currentFile] = g
				m.graphV.setGraph(g)
			}
		}
		return m, nil

	case commentSubmittedMsg:
		if msg.err != nil {
			m.toastMsg = fmt.Sprintf("Error posting comment: %s", msg.err)
			m.toastIsError = true
		} else {
			m.toastMsg = "✓ Review comment posted"
			m.toastIsError = false
		}
		m.activeModal = modalNone
		return m, fetchThreadsCmd(m.info)

	case replySubmittedMsg:
		if msg.err != nil {
			m.toastMsg = fmt.Sprintf("Error posting reply: %s", msg.err)
			m.toastIsError = true
		} else {
			m.toastMsg = "✓ Reply posted"
			m.toastIsError = false
		}
		m.activeModal = modalNone
		return m, fetchThreadsCmd(m.info)

	case threadResolvedMsg:
		if msg.err != nil {
			m.toastMsg = fmt.Sprintf("Error toggling thread: %s", msg.err)
			m.toastIsError = true
		} else {
			if msg.resolved {
				m.toastMsg = "✓ Thread resolved"
			} else {
				m.toastMsg = "✓ Thread marked unresolved"
			}
			m.toastIsError = false
		}
		return m, fetchThreadsCmd(m.info)

	case tea.KeyMsg:
		// Global quit
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

		// Modal handling
		if m.activeModal != modalNone {
			return m.handleModalKey(msg)
		}

		// Normal navigation
		return m.handleNormalKey(msg)
	}

	return m, nil
}

func (m Model) handleModalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch m.activeModal {
	case modalHelp:
		if key == "esc" || key == "?" || key == "q" {
			m.activeModal = modalNone
		}
		return m, nil

	case modalFiles:
		switch key {
		case "esc", "f", "b", "q":
			m.activeModal = modalNone
			return m, nil
		case "k", "up":
			m.filesM.moveUp()
			return m, nil
		case "j", "down":
			m.filesM.moveDown()
			return m, nil
		case "enter":
			selected := m.filesM.selectedIndex()
			m.activeModal = modalNone
			if selected != m.currentFileIdx {
				m.currentFileIdx = selected
				return m.switchToFile(m.currentFileIdx)
			}
			return m, nil
		}

	case modalIntent:
		switch key {
		case "esc", "i", "q":
			m.activeModal = modalNone
			return m, nil
		case "k", "up":
			m.intentM.scrollUp()
			return m, nil
		case "j", "down":
			m.intentM.scrollDown()
			return m, nil
		}

	case modalComment:
		switch key {
		case "esc":
			m.activeModal = modalNone
			return m, nil
		case "ctrl+s":
			body := m.commM.value()
			if body == "" {
				m.toastMsg = "Cannot submit empty comment"
				m.toastIsError = true
				return m, nil
			}
			if m.commM.mode == commentModeReply {
				return m, submitReplyCmd(m.info, m.commM.replyToID, body)
			}
			return m, submitCommentCmd(m.info, m.commM.path, body, m.commM.line, m.commM.side)
		}
		var cmd tea.Cmd
		m.commM, cmd = m.commM.update(msg)
		return m, cmd

	case modalThreads:
		switch key {
		case "esc", "q":
			m.activeModal = modalNone
			return m, nil
		case "k", "up":
			m.thM.scrollUp()
			return m, nil
		case "j", "down":
			m.thM.scrollDown()
			return m, nil
		case "]", "n":
			m.thM.nextThread()
			return m, nil
		case "[", "p":
			m.thM.prevThread()
			return m, nil
		case "r":
			th := m.thM.activeThread()
			if th != nil && len(th.Comments) > 0 {
				lastCommentID := th.Comments[len(th.Comments)-1].ID
				m.commM = newReplyModal(m.thM.path, m.thM.line, lastCommentID)
				m.commM.setSize(m.width)
				m.activeModal = modalComment
			}
			return m, nil
		case "e":
			th := m.thM.activeThread()
			if th != nil {
				return m, toggleResolveCmd(th.NodeID, !th.Resolved)
			}
			return m, nil
		}
	}

	return m, nil
}

func (m Model) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	vpHeight := m.viewportHeight()

	switch key {
	case "q":
		m.quitting = true
		return m, tea.Quit

	case "?":
		m.activeModal = modalHelp
		return m, nil

	case "1":
		m.activeMode = modeUnified
		m.lastDiffMode = modeUnified
		m.toastMsg = ""
		return m, nil

	case "2":
		m.activeMode = modeSplit
		m.lastDiffMode = modeSplit
		m.toastMsg = ""
		return m, nil

	case "3":
		m.activeMode = modeRendered
		m.lastDiffMode = modeRendered
		m.toastMsg = ""
		return m, nil

	case "4", "s":
		if m.activeMode != modeGraph {
			m.lastDiffMode = m.activeMode
		}
		m.activeMode = modeGraph
		m.toastMsg = ""
		return m, nil

	case "tab":
		m.activeMode = (m.activeMode + 1) % 4
		if m.activeMode != modeGraph {
			m.lastDiffMode = m.activeMode
		}
		m.toastMsg = ""
		return m, nil

	case "i":
		m.activeModal = modalIntent
		return m, nil

	case "f", "b":
		m.filesM.updateFiles(m.files, m.threadCount)
		m.activeModal = modalFiles
		return m, nil

	case "j", "down":
		m.scrollDown(vpHeight)
		return m, nil

	case "k", "up":
		m.scrollUp(vpHeight)
		return m, nil

	case "d", "ctrl+d":
		m.pageDown(vpHeight)
		return m, nil

	case "u", "ctrl+u":
		m.pageUp(vpHeight)
		return m, nil

	case "g":
		m.top(vpHeight)
		return m, nil

	case "G":
		m.bottom(vpHeight)
		return m, nil

	case "n", "]":
		m.nextHunk(vpHeight)
		return m, nil

	case "p", "[":
		m.prevHunk(vpHeight)
		return m, nil

	case "c":
		return m.initiateComment()

	case "t":
		return m.openThreads()

	case "enter":
		if m.activeMode == modeGraph {
			node := m.graphV.currentNode()
			if node != nil {
				m.jumpToSection(node, vpHeight)
				targetMode := m.lastDiffMode
				if targetMode == modeGraph {
					targetMode = modeUnified
				}
				m.activeMode = targetMode
				m.toastMsg = fmt.Sprintf("✓ Jumped to %s", node.RawHeading)
				m.toastIsError = false
			}
			return m, nil
		}
		return m.openThreads()
	}

	return m, nil
}

func (m *Model) scrollDown(vpHeight int) {
	switch m.activeMode {
	case modeUnified:
		m.unified.moveDown(vpHeight)
	case modeSplit:
		m.split.moveDown(vpHeight)
	case modeRendered:
		m.rendered.moveDown(vpHeight)
	case modeGraph:
		m.graphV.moveDown(vpHeight)
	}
}

func (m *Model) scrollUp(vpHeight int) {
	switch m.activeMode {
	case modeUnified:
		m.unified.moveUp(vpHeight)
	case modeSplit:
		m.split.moveUp(vpHeight)
	case modeRendered:
		m.rendered.moveUp(vpHeight)
	case modeGraph:
		m.graphV.moveUp(vpHeight)
	}
}

func (m *Model) pageDown(vpHeight int) {
	switch m.activeMode {
	case modeUnified:
		m.unified.pageDown(vpHeight)
	case modeSplit:
		m.split.pageDown(vpHeight)
	case modeRendered:
		m.rendered.pageDown(vpHeight)
	case modeGraph:
		delta := vpHeight / 2
		if delta < 1 {
			delta = 1
		}
		for i := 0; i < delta; i++ {
			m.graphV.moveDown(vpHeight)
		}
	}
}

func (m *Model) pageUp(vpHeight int) {
	switch m.activeMode {
	case modeUnified:
		m.unified.pageUp(vpHeight)
	case modeSplit:
		m.split.pageUp(vpHeight)
	case modeRendered:
		m.rendered.pageUp(vpHeight)
	case modeGraph:
		delta := vpHeight / 2
		if delta < 1 {
			delta = 1
		}
		for i := 0; i < delta; i++ {
			m.graphV.moveUp(vpHeight)
		}
	}
}

func (m *Model) top(vpHeight int) {
	switch m.activeMode {
	case modeUnified:
		m.unified.top(vpHeight)
	case modeSplit:
		m.split.top(vpHeight)
	case modeRendered:
		m.rendered.top(vpHeight)
	case modeGraph:
		m.graphV.top(vpHeight)
	}
}

func (m *Model) bottom(vpHeight int) {
	switch m.activeMode {
	case modeUnified:
		m.unified.bottom(vpHeight)
	case modeSplit:
		m.split.bottom(vpHeight)
	case modeRendered:
		m.rendered.bottom(vpHeight)
	case modeGraph:
		m.graphV.bottom(vpHeight)
	}
}

func (m *Model) nextHunk(vpHeight int) {
	switch m.activeMode {
	case modeUnified:
		m.unified.nextHunk(vpHeight)
	case modeSplit:
		m.split.nextHunk(vpHeight)
	}
}

func (m *Model) prevHunk(vpHeight int) {
	switch m.activeMode {
	case modeUnified:
		m.unified.prevHunk(vpHeight)
	case modeSplit:
		m.split.prevHunk(vpHeight)
	}
}

func (m *Model) jumpToSection(node *graph.Node, vpHeight int) {
	if node == nil {
		return
	}

	// 1. Unified view jump
	for i, dl := range m.unified.lines {
		if (dl.NewLineNo > 0 && dl.NewLineNo >= node.LineNo) ||
			strings.Contains(dl.Content, node.Title) ||
			strings.TrimSpace(dl.Content) == strings.TrimSpace(node.RawHeading) {
			m.unified.cursor = i
			m.unified.adjustScroll(vpHeight)
			break
		}
	}

	// 2. Split view jump
	for i, row := range m.split.rows {
		if (row.Right != nil && row.Right.NewLineNo >= node.LineNo) ||
			strings.Contains(row.HeaderLine, node.Title) ||
			(row.Right != nil && strings.Contains(row.Right.Content, node.Title)) {
			m.split.cursor = i
			m.split.adjustScroll(vpHeight)
			break
		}
	}

	// 3. Rendered view jump
	for i, line := range m.rendered.renderedLines {
		if strings.Contains(line, node.Title) {
			m.rendered.cursor = i
			m.rendered.adjustScroll(vpHeight)
			break
		}
	}
}

func (m *Model) switchToFile(idx int) (tea.Model, tea.Cmd) {
	f := m.files[idx]
	if fd, exists := m.fileDiffs[f.Filename]; exists {
		head := m.headContent[f.Filename]
		tRight := m.threadMap[f.Filename]
		if tRight == nil {
			tRight = make(map[int][]ghpr.ReviewThread)
		}
		tLeft := m.delThreadMap[f.Filename]
		if tLeft == nil {
			tLeft = make(map[int][]ghpr.ReviewThread)
		}
		m.unified = newUnifiedView(fd.UnifiedLines, tRight, tLeft)
		m.split = newSplitView(fd.SplitRows, tRight, tLeft)
		m.rendered = newRenderedView(head, m.width)
		m.intentM.setIntent(fd.Intent)

		if g, ok := m.graphMap[f.Filename]; ok {
			m.graphV = newGraphView(g)
		} else if head != "" {
			g := graph.BuildSectionGraph(head, fd.Hunks, tRight)
			m.graphMap[f.Filename] = g
			m.graphV = newGraphView(g)
		}

		return m, nil
	}

	m.loading = true
	m.loadingMsg = fmt.Sprintf("Loading %s…", f.Filename)
	return m, loadFileCmd(m.info, f)
}

func (m Model) initiateComment() (tea.Model, tea.Cmd) {
	currentFile := m.files[m.currentFileIdx].Filename
	var line int
	side := "RIGHT"

	switch m.activeMode {
	case modeUnified:
		dl := m.unified.currentLine()
		if dl == nil || dl.Type == diff.LineHunkHeader {
			m.toastMsg = "Cannot comment on hunk header"
			m.toastIsError = true
			return m, nil
		}
		if dl.NewLineNo > 0 {
			line = dl.NewLineNo
			side = "RIGHT"
		} else if dl.OldLineNo > 0 {
			line = dl.OldLineNo
			side = "LEFT"
		}

	case modeSplit:
		row := m.split.currentRow()
		if row == nil || row.IsHeader {
			m.toastMsg = "Cannot comment on header"
			m.toastIsError = true
			return m, nil
		}
		if row.Right != nil && row.Right.NewLineNo > 0 {
			line = row.Right.NewLineNo
			side = "RIGHT"
		} else if row.Left != nil && row.Left.OldLineNo > 0 {
			line = row.Left.OldLineNo
			side = "LEFT"
		}

	case modeRendered, modeGraph:
		m.toastMsg = "Press Enter on a section to jump to Diff view, then press 'c' to comment"
		m.toastIsError = true
		return m, nil
	}

	if line <= 0 {
		m.toastMsg = "Selected line is not commentable"
		m.toastIsError = true
		return m, nil
	}

	m.commM = newCommentModal(currentFile, line, side)
	m.commM.setSize(m.width)
	m.activeModal = modalComment
	return m, nil
}

func (m Model) openThreads() (tea.Model, tea.Cmd) {
	currentFile := m.files[m.currentFileIdx].Filename
	var targetLine int
	var threads []ghpr.ReviewThread

	switch m.activeMode {
	case modeUnified:
		dl := m.unified.currentLine()
		if dl != nil {
			if dl.NewLineNo > 0 {
				targetLine = dl.NewLineNo
				threads = m.threadMap[currentFile][targetLine]
			} else if dl.OldLineNo > 0 {
				targetLine = dl.OldLineNo
				threads = m.delThreadMap[currentFile][targetLine]
			}
		}

	case modeSplit:
		row := m.split.currentRow()
		if row != nil {
			if row.Right != nil && row.Right.NewLineNo > 0 {
				targetLine = row.Right.NewLineNo
				threads = m.threadMap[currentFile][targetLine]
			} else if row.Left != nil && row.Left.OldLineNo > 0 {
				targetLine = row.Left.OldLineNo
				threads = m.delThreadMap[currentFile][targetLine]
			}
		}
	}

	if len(threads) == 0 {
		m.toastMsg = "No comments on current line (press 'c' to post one)"
		m.toastIsError = false
		return m, nil
	}

	m.thM = newThreadsModal(currentFile, targetLine, threads)
	m.activeModal = modalThreads
	return m, nil
}

func (m *Model) rebuildThreadIndex() {
	m.threadMap = make(map[string]map[int][]ghpr.ReviewThread)
	m.delThreadMap = make(map[string]map[int][]ghpr.ReviewThread)
	m.threadCount = make(map[string]int)

	for _, th := range m.threads {
		m.threadCount[th.Path] += len(th.Comments)

		if _, ok := m.threadMap[th.Path]; !ok {
			m.threadMap[th.Path] = make(map[int][]ghpr.ReviewThread)
		}
		if _, ok := m.delThreadMap[th.Path]; !ok {
			m.delThreadMap[th.Path] = make(map[int][]ghpr.ReviewThread)
		}

		m.threadMap[th.Path][th.Line] = append(m.threadMap[th.Path][th.Line], th)
	}
}

func (m Model) viewportHeight() int {
	h := m.height - 5
	if h < 5 {
		return 5
	}
	return h
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if m.loading {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.loadingMsg)
	}

	var sb strings.Builder

	// Top Header Bar
	sb.WriteString(m.renderHeader() + "\n")

	// Viewport Content
	vpHeight := m.viewportHeight()
	var mainContent string

	switch m.activeMode {
	case modeUnified:
		mainContent = m.unified.view(m.width, vpHeight)
	case modeSplit:
		mainContent = m.split.view(m.width, vpHeight)
	case modeRendered:
		mainContent = m.rendered.view(m.width, vpHeight)
	case modeGraph:
		mainContent = m.graphV.view(m.width, vpHeight)
	}

	sb.WriteString(mainContent + "\n")

	// Bottom Status Bar
	sb.WriteString(m.renderFooter())

	baseView := sb.String()

	// Overlay Modals
	switch m.activeModal {
	case modalHelp:
		return m.help.view(m.width, m.height)
	case modalFiles:
		return m.filesM.view(m.width, m.height)
	case modalIntent:
		return m.intentM.view(m.width, m.height)
	case modalComment:
		return m.commM.view(m.width, m.height)
	case modalThreads:
		return m.thM.view(m.width, m.height)
	}

	return baseView
}

func (m Model) renderHeader() string {
	currFile := m.files[m.currentFileIdx]
	prTitle := m.info.Title
	if prTitle == "" {
		prTitle = fmt.Sprintf("PR #%d", m.info.Ref.Number)
	}
	topTitle := fmt.Sprintf(" %s/%s#%d: %s ", m.info.Ref.Owner, m.info.Ref.Repo, m.info.Ref.Number, prTitle)

	// Mode tabs
	tab1 := inactiveTabStyle.Render("[1] Unified Diff")
	tab2 := inactiveTabStyle.Render("[2] Split Diff")
	tab3 := inactiveTabStyle.Render("[3] Rendered")
	tab4 := inactiveTabStyle.Render("[4] Graph")

	switch m.activeMode {
	case modeUnified:
		tab1 = activeTabStyle.Render("[1] Unified Diff")
	case modeSplit:
		tab2 = activeTabStyle.Render("[2] Split Diff")
	case modeRendered:
		tab3 = activeTabStyle.Render("[3] Rendered")
	case modeGraph:
		tab4 = activeTabStyle.Render("[4] Graph")
	}

	intentTab := inactiveTabStyle.Render("[i] Intent")
	fileTab := inactiveTabStyle.Render(fmt.Sprintf("[f] Files (%d)", len(m.files)))
	helpTab := inactiveTabStyle.Render("[?] Help")

	tabBar := fmt.Sprintf("%s %s %s %s   %s %s %s", tab1, tab2, tab3, tab4, intentTab, fileTab, helpTab)

	// Second line: file info
	fileNum := fmt.Sprintf("(%d/%d)", m.currentFileIdx+1, len(m.files))
	statBadge := fmt.Sprintf("+%d -%d", currFile.Additions, currFile.Deletions)
	commentStr := ""
	if count := m.threadCount[currFile.Filename]; count > 0 {
		commentStr = fmt.Sprintf(" · 💬 %d comments", count)
	}
	fileInfo := fmt.Sprintf(" File: %s %s  %s%s",
		currFile.Filename, fileNum, statBadge, commentStr)

	return titleStyle.Render(topTitle) + "\n" + tabBar + "\n" + headerSubStyle.Render(fileInfo)
}

func (m Model) renderFooter() string {
	hint := "j/k scroll · 1/2/3/4 view · s graph · c comment · t threads · n/p hunk · f files · q quit"
	if m.activeMode == modeGraph {
		hint = "j/k navigate tree · enter jump to diff · 1/2/3 switch view · q quit"
	}

	if m.toastMsg != "" {
		if m.toastIsError {
			return statusErrorStyle.Width(m.width).Render("✕ " + m.toastMsg)
		}
		return statusToastStyle.Width(m.width).Render(m.toastMsg)
	}

	return statusBarStyle.Width(m.width).Render(hint)
}

// Async commands
func loadFileCmd(info ghpr.PRInfo, file ghpr.File) tea.Cmd {
	return func() tea.Msg {
		head, err := ghpr.FileContent(info, file.Filename)
		if err != nil {
			return fileLoadedMsg{file: file, err: err}
		}
		base, _ := ghpr.FileBaseContent(info, file.Filename)
		fd := diff.NewFileDiff(file.Filename, file.Patch, base, head)
		return fileLoadedMsg{
			file:        file,
			fileDiff:    fd,
			headContent: head,
			baseContent: base,
		}
	}
}

func fetchThreadsCmd(info ghpr.PRInfo) tea.Cmd {
	return func() tea.Msg {
		threads, err := ghpr.FetchThreads(info)
		return threadsLoadedMsg{threads: threads, err: err}
	}
}

func submitCommentCmd(info ghpr.PRInfo, path, body string, line int, side string) tea.Cmd {
	return func() tea.Msg {
		err := ghpr.PostLineComment(info, path, body, line, side)
		return commentSubmittedMsg{err: err}
	}
}

func submitReplyCmd(info ghpr.PRInfo, replyToID int64, body string) tea.Cmd {
	return func() tea.Msg {
		err := ghpr.ReplyToComment(info, replyToID, body)
		return replySubmittedMsg{err: err}
	}
}

func toggleResolveCmd(nodeID string, resolve bool) tea.Cmd {
	return func() tea.Msg {
		var err error
		if resolve {
			err = ghpr.ResolveThread(nodeID)
		} else {
			err = ghpr.UnresolveThread(nodeID)
		}
		return threadResolvedMsg{nodeID: nodeID, resolved: resolve, err: err}
	}
}
