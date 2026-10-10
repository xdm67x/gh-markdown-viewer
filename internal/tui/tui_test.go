package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xdm67x/gh-markdown-viewer/internal/diff"
	"github.com/xdm67x/gh-markdown-viewer/internal/ghpr"
)

func TestModelInitialization(t *testing.T) {
	info := ghpr.PRInfo{
		Ref: ghpr.Ref{
			Owner:  "test-owner",
			Repo:   "test-repo",
			Number: 99,
		},
		Title: "Update AI prompt and harness plan",
	}

	files := []ghpr.File{
		{
			Filename:  "plan.md",
			Status:    "modified",
			Additions: 10,
			Deletions: 2,
			Patch:     "@@ -1,3 +1,4 @@ ## Setup\n Line 1\n+New line\n Line 2\n",
		},
		{
			Filename:  "prompts.md",
			Status:    "added",
			Additions: 25,
			Deletions: 0,
			Patch:     "@@ -0,0 +1,5 @@ ## Prompts\n+Prompt 1\n",
		},
	}

	m := New(info, files)
	if m.currentFileIdx != 0 {
		t.Errorf("expected currentFileIdx 0, got %d", m.currentFileIdx)
	}
	if m.activeMode != modeUnified {
		t.Errorf("expected activeMode modeUnified, got %v", m.activeMode)
	}

	// Trigger WindowSizeMsg
	mUpdated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m = mUpdated.(Model)

	if m.width != 100 || m.height != 40 {
		t.Errorf("expected dimensions 100x40, got %dx%d", m.width, m.height)
	}

	// Simulate fileLoadedMsg
	fd := diff.NewFileDiff("plan.md", files[0].Patch, "", "## Setup\nLine 1\nNew line\nLine 2\n")
	mUpdated, _ = m.Update(fileLoadedMsg{
		file:        files[0],
		fileDiff:    fd,
		headContent: "## Setup\nLine 1\nNew line\nLine 2\n",
	})
	m = mUpdated.(Model)

	// Test switching modes via keys
	mUpdated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m = mUpdated.(Model)
	if m.activeMode != modeSplit {
		t.Errorf("expected activeMode modeSplit, got %v", m.activeMode)
	}

	mUpdated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = mUpdated.(Model)
	if m.activeMode != modeRendered {
		t.Errorf("expected activeMode modeRendered, got %v", m.activeMode)
	}

	// Tab from 3 (Rendered) cycles to 3 (Graph)
	mUpdated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = mUpdated.(Model)
	if m.activeMode != modeGraph {
		t.Errorf("expected activeMode modeGraph after tab, got %v", m.activeMode)
	}

	// Tab from Graph cycles back to 0 (Unified)
	mUpdated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = mUpdated.(Model)
	if m.activeMode != modeUnified {
		t.Errorf("expected activeMode modeUnified after tab, got %v", m.activeMode)
	}

	// Test opening Help modal
	mUpdated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = mUpdated.(Model)
	if m.activeModal != modalHelp {
		t.Errorf("expected modalHelp, got %v", m.activeModal)
	}

	// Close modal with esc
	mUpdated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = mUpdated.(Model)
	if m.activeModal != modalNone {
		t.Errorf("expected modalNone, got %v", m.activeModal)
	}

	// Test opening Files modal
	mUpdated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m = mUpdated.(Model)
	if m.activeModal != modalFiles {
		t.Errorf("expected modalFiles, got %v", m.activeModal)
	}

	// Close modal with esc
	mUpdated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = mUpdated.(Model)
	if m.activeModal != modalNone {
		t.Errorf("expected modalNone, got %v", m.activeModal)
	}

	// Test opening Intent modal
	mUpdated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	m = mUpdated.(Model)
	if m.activeModal != modalIntent {
		t.Errorf("expected modalIntent, got %v", m.activeModal)
	}

	// Close modal with esc
	mUpdated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = mUpdated.(Model)

	// Test switching to Section Graph view with key '4'
	mUpdated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	m = mUpdated.(Model)
	if m.activeMode != modeGraph {
		t.Errorf("expected activeMode modeGraph with '4', got %v", m.activeMode)
	}

	// Test navigating graph and jumping with Enter back to diff view
	mUpdated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mUpdated.(Model)
	if m.activeMode != modeUnified {
		t.Errorf("expected activeMode modeUnified after Enter jump, got %v", m.activeMode)
	}

	// Test switching to Section Graph view with key 's'
	mUpdated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = mUpdated.(Model)
	if m.activeMode != modeGraph {
		t.Errorf("expected activeMode modeGraph with 's', got %v", m.activeMode)
	}
}
