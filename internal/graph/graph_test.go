package graph

import (
	"testing"

	"github.com/xdm67x/gh-markdown-viewer/internal/diff"
	"github.com/xdm67x/gh-markdown-viewer/internal/ghpr"
)

const sampleMarkdown = `# System Plan

Intro description referencing [Architecture](#architecture).

## Architecture

Detailed architecture design.
- [x] Define data structures
- [ ] Implement runner

### Runner Subsystem

Runner details.

## Deployment

Deployment steps.
`

func TestBuildSectionGraph(t *testing.T) {
	hunks := []diff.DiffHunk{
		{
			Section: "## Architecture",
			Lines: []diff.DiffLine{
				{Type: diff.LineAdded, NewLineNo: 8},
				{Type: diff.LineAdded, NewLineNo: 9},
			},
		},
	}

	threadMap := map[int][]ghpr.ReviewThread{
		8: {{Line: 8, Comments: []ghpr.ReviewComment{{ID: 1}}}},
	}

	g := BuildSectionGraph(sampleMarkdown, hunks, threadMap)

	if len(g.Nodes) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(g.Nodes))
	}

	// Root should be # System Plan
	root := g.Nodes[0]
	if root.Title != "System Plan" || root.Level != 1 {
		t.Errorf("unexpected root node: %+v", root)
	}

	// Root should have 2 children: Architecture and Deployment
	if len(root.ChildrenIDs) != 2 {
		t.Fatalf("expected root to have 2 children, got %d", len(root.ChildrenIDs))
	}

	// Architecture should have 1 child: Runner Subsystem
	arch := g.Nodes[root.ChildrenIDs[0]]
	if arch.Title != "Architecture" || arch.Level != 2 {
		t.Errorf("unexpected arch node: %+v", arch)
	}
	if len(arch.ChildrenIDs) != 1 {
		t.Fatalf("expected arch to have 1 child, got %d", len(arch.ChildrenIDs))
	}

	// Verify tasks in Architecture
	if arch.TotalTasks != 2 || arch.CompletedTasks != 1 {
		t.Errorf("expected 2 tasks (1 completed) in Architecture, got total=%d, completed=%d",
			arch.TotalTasks, arch.CompletedTasks)
	}

	// Verify cross-links: Root links to #architecture
	if len(root.LinksTo) != 1 || root.LinksTo[0] != "architecture" {
		t.Errorf("expected root to link to 'architecture', got %+v", root.LinksTo)
	}

	// Verify edges include crosslink
	hasCrosslink := false
	for _, e := range g.Edges {
		if e.Type == "crosslink" && e.FromID == root.ID && e.ToID == arch.ID {
			hasCrosslink = true
			break
		}
	}
	if !hasCrosslink {
		t.Errorf("expected crosslink edge from System Plan to Architecture")
	}

	// Verify additions and comment count on Architecture
	if arch.Additions != 2 {
		t.Errorf("expected 2 additions on Architecture, got %d", arch.Additions)
	}
	if arch.CommentCount != 1 {
		t.Errorf("expected 1 comment on Architecture, got %d", arch.CommentCount)
	}

	// Verify FlattenedOrder has all 4 nodes
	if len(g.FlattenedOrder) != 4 {
		t.Errorf("expected FlattenedOrder to have 4 nodes, got %d", len(g.FlattenedOrder))
	}
}

func TestSlugify(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"Architecture Overview", "architecture-overview"},
		{"1. Step-by-Step Plan", "1-step-by-step-plan"},
		{"FAQ / Questions?", "faq-questions"},
	}

	for _, c := range cases {
		got := Slugify(c.input)
		if got != c.want {
			t.Errorf("Slugify(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}
