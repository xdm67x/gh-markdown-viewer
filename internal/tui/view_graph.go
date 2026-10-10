package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/xdm67x/gh-markdown-viewer/internal/graph"
)

type graphView struct {
	g       *graph.Graph
	cursor  int
	yOffset int
}

func newGraphView(g *graph.Graph) graphView {
	return graphView{
		g:       g,
		cursor:  0,
		yOffset: 0,
	}
}

func (v *graphView) setGraph(g *graph.Graph) {
	v.g = g
	v.cursor = 0
	v.yOffset = 0
}

func (v *graphView) currentNode() *graph.Node {
	if v.g == nil {
		return nil
	}
	return v.g.NodeAt(v.cursor)
}

func (v *graphView) moveUp(viewportHeight int) {
	if v.cursor > 0 {
		v.cursor--
		v.adjustScroll(viewportHeight)
	}
}

func (v *graphView) moveDown(viewportHeight int) {
	if v.g != nil && v.cursor < len(v.g.FlattenedOrder)-1 {
		v.cursor++
		v.adjustScroll(viewportHeight)
	}
}

func (v *graphView) top(viewportHeight int) {
	v.cursor = 0
	v.adjustScroll(viewportHeight)
}

func (v *graphView) bottom(viewportHeight int) {
	if v.g != nil && len(v.g.FlattenedOrder) > 0 {
		v.cursor = len(v.g.FlattenedOrder) - 1
		v.adjustScroll(viewportHeight)
	}
}

func (v *graphView) adjustScroll(viewportHeight int) {
	if v.cursor < v.yOffset {
		v.yOffset = v.cursor
	} else if v.cursor >= v.yOffset+viewportHeight {
		v.yOffset = v.cursor - viewportHeight + 1
	}
}

func (v graphView) view(width, height int) string {
	if v.g == nil || len(v.g.FlattenedOrder) == 0 {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, "No sections found in this document.")
	}

	// If screen is wide, use dual panel (Tree left, Inspector right)
	if width >= 80 {
		inspectorWidth := 38
		if width > 110 {
			inspectorWidth = 46
		}
		treeWidth := width - inspectorWidth - 3

		treeContent := v.renderTree(treeWidth, height)
		inspectorContent := v.renderInspector(inspectorWidth, height)

		return lipgloss.JoinHorizontal(lipgloss.Top, treeContent, " │ ", inspectorContent)
	}

	return v.renderTree(width, height)
}

func (v graphView) renderTree(width, height int) string {
	var sb strings.Builder

	end := v.yOffset + height
	if end > len(v.g.FlattenedOrder) {
		end = len(v.g.FlattenedOrder)
	}

	cursorStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))

	for i := v.yOffset; i < end; i++ {
		node := v.g.NodeAt(i)
		if node == nil {
			continue
		}
		isCurrent := i == v.cursor

		// Tree branch indentation
		depth := node.Level - 1
		if depth < 0 {
			depth = 0
		}
		indent := strings.Repeat("  ", depth)
		branch := "├─ "
		if node.ParentID == -1 && depth == 0 {
			branch = "┌─ "
		} else if isLastChild(v.g, node) {
			branch = "└─ "
		}

		// Level tag: [#], [##], etc.
		levelTag := fmt.Sprintf("[%s]", strings.Repeat("#", node.Level))
		levelStr := graphLevelStyle.Render(levelTag)

		// Change status badge
		var statusBadge string
		switch {
		case node.Additions > 0 && node.Deletions > 0:
			statusBadge = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220")).Render("[●]")
		case node.Additions > 0:
			statusBadge = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("78")).Render("[+]")
		case node.Deletions > 0:
			statusBadge = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("167")).Render("[-]")
		default:
			statusBadge = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("[ ]")
		}

		// Title
		titleStr := graphTitleStyle.Render(node.Title)

		// Stats tags
		var tags []string
		if node.Additions > 0 || node.Deletions > 0 {
			statsText := fmt.Sprintf("+%d -%d", node.Additions, node.Deletions)
			tags = append(tags, lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(statsText))
		}
		if node.TotalTasks > 0 {
			taskText := fmt.Sprintf("[%d/%d ✓]", node.CompletedTasks, node.TotalTasks)
			tags = append(tags, graphTaskStyle.Render(taskText))
		}
		if node.CommentCount > 0 {
			commentText := fmt.Sprintf("💬 %d", node.CommentCount)
			tags = append(tags, commentBadgeStyle.Render(commentText))
		}
		if len(node.LinksTo) > 0 {
			linkText := fmt.Sprintf("↗ %d links", len(node.LinksTo))
			tags = append(tags, graphLinkStyle.Render(linkText))
		}

		tagStr := ""
		if len(tags) > 0 {
			tagStr = " " + strings.Join(tags, " ")
		}

		pointer := "  "
		if isCurrent {
			pointer = cursorStyle.Render("▶ ")
		}

		row := fmt.Sprintf("%s%s%s%s %s %s%s",
			pointer,
			graphBranchStyle.Render(indent),
			graphBranchStyle.Render(branch),
			statusBadge,
			levelStr,
			titleStr,
			tagStr,
		)

		if isCurrent {
			row = cursorLineStyle.Width(width).Render(row)
		}

		sb.WriteString(row + "\n")
	}

	linesRendered := end - v.yOffset
	for linesRendered < height {
		sb.WriteString("\n")
		linesRendered++
	}

	return sb.String()
}

func (v graphView) renderInspector(width, height int) string {
	curr := v.currentNode()
	if curr == nil {
		return ""
	}

	var sb strings.Builder

	header := modalTitleStyle.Render("Section Inspector")
	sb.WriteString(header + "\n\n")

	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render(curr.RawHeading) + "\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(
		fmt.Sprintf("Line: %d · Level: %d · Anchor: #%s", curr.LineNo, curr.Level, curr.Anchor),
	) + "\n\n")

	// Diff stats
	addStr := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("78")).Render(fmt.Sprintf("+%d", curr.Additions))
	delStr := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("167")).Render(fmt.Sprintf("-%d", curr.Deletions))
	sb.WriteString(fmt.Sprintf("Changes: %s additions · %s deletions\n", addStr, delStr))

	// Review Comments
	if curr.CommentCount > 0 {
		sb.WriteString(commentBadgeStyle.Render(fmt.Sprintf("💬 %d review comments in this section", curr.CommentCount)) + "\n")
	} else {
		sb.WriteString("💬 No comments in this section\n")
	}
	sb.WriteString("\n")

	// Tasks
	if curr.TotalTasks > 0 {
		taskHeader := fmt.Sprintf("Tasks (%d/%d completed):", curr.CompletedTasks, curr.TotalTasks)
		sb.WriteString(graphTaskStyle.Render(taskHeader) + "\n")
		for _, t := range curr.Tasks {
			sb.WriteString(fmt.Sprintf("  • %s\n", t))
		}
		sb.WriteString("\n")
	}

	// Graph connections
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250")).Render("Hierarchy & Graph Links:") + "\n")
	if curr.ParentID != -1 && curr.ParentID < len(v.g.Nodes) {
		parent := &v.g.Nodes[curr.ParentID]
		sb.WriteString(fmt.Sprintf("  Parent: %s (line %d)\n", parent.Title, parent.LineNo))
	} else {
		sb.WriteString("  Parent: (Document Root)\n")
	}

	if len(curr.ChildrenIDs) > 0 {
		sb.WriteString(fmt.Sprintf("  Sub-sections: %d children\n", len(curr.ChildrenIDs)))
		for _, cid := range curr.ChildrenIDs {
			if cid < len(v.g.Nodes) {
				sb.WriteString(fmt.Sprintf("    ↳ %s\n", v.g.Nodes[cid].Title))
			}
		}
	}

	if len(curr.LinksTo) > 0 {
		sb.WriteString(graphLinkStyle.Render(fmt.Sprintf("  Cross-references (%d):", len(curr.LinksTo))) + "\n")
		for _, anchor := range curr.LinksTo {
			targetName := anchor
			if tid, ok := v.g.NodeByAnchor[anchor]; ok && tid < len(v.g.Nodes) {
				targetName = v.g.Nodes[tid].Title
			}
			sb.WriteString(fmt.Sprintf("    ↗ #%s (%s)\n", anchor, targetName))
		}
	}
	sb.WriteString("\n")

	sb.WriteString(modalFooterStyle.Render("↵ Jump to section in Diff view\n1/2/3 switch view · j/k navigate"))

	return graphInspectorPanelStyle.Width(width - 4).Render(sb.String())
}

func isLastChild(g *graph.Graph, n *graph.Node) bool {
	if n.ParentID == -1 {
		if len(g.RootIDs) == 0 {
			return true
		}
		return g.RootIDs[len(g.RootIDs)-1] == n.ID
	}
	if n.ParentID < len(g.Nodes) {
		parent := &g.Nodes[n.ParentID]
		if len(parent.ChildrenIDs) == 0 {
			return true
		}
		return parent.ChildrenIDs[len(parent.ChildrenIDs)-1] == n.ID
	}
	return true
}
