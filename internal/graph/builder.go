package graph

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/xdm67x/gh-markdown-viewer/internal/diff"
	"github.com/xdm67x/gh-markdown-viewer/internal/ghpr"
)

var (
	headingRegex = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	chkBoxRegex  = regexp.MustCompile(`^[\s\-\*\+]*\[([ xX])\]\s*(.*)$`)
	linkRefRegex = regexp.MustCompile(`\[.*?\]\(#([a-zA-Z0-9\-_]+)\)`)
)

// Slugify generates a URL anchor slug from a heading title.
func Slugify(title string) string {
	title = strings.ToLower(strings.TrimSpace(title))
	var sb strings.Builder
	for _, r := range title {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' {
			sb.WriteRune('-')
		}
	}
	res := sb.String()
	for strings.Contains(res, "--") {
		res = strings.ReplaceAll(res, "--", "-")
	}
	return strings.Trim(res, "-")
}

// BuildSectionGraph constructs a navigable section graph from Markdown content and diff data.
func BuildSectionGraph(
	content string,
	hunks []diff.DiffHunk,
	threadMap map[int][]ghpr.ReviewThread,
) *Graph {
	lines := strings.Split(content, "\n")
	var rawNodes []Node
	var sectionEndLines []int

	// 1. Scan for headings
	for lineIdx, rawLine := range lines {
		lineNo := lineIdx + 1
		m := headingRegex.FindStringSubmatch(rawLine)
		if m != nil {
			lvl := len(m[1])
			title := strings.TrimSpace(m[2])
			rawHeading := strings.TrimSpace(rawLine)
			anchor := Slugify(title)

			node := Node{
				ID:          len(rawNodes),
				Title:       title,
				RawHeading:  rawHeading,
				Anchor:      anchor,
				Level:       lvl,
				LineNo:      lineNo,
				ParentID:    -1,
				ChildrenIDs: nil,
			}
			rawNodes = append(rawNodes, node)
		}
	}

	// Fallback for files without any headings: create a single root node
	if len(rawNodes) == 0 {
		rawNodes = append(rawNodes, Node{
			ID:         0,
			Title:      "Document",
			RawHeading: "Document",
			Anchor:     "document",
			Level:      1,
			LineNo:     1,
			ParentID:   -1,
		})
		sectionEndLines = append(sectionEndLines, len(lines)+1)
	} else {
		for i := 0; i < len(rawNodes); i++ {
			if i+1 < len(rawNodes) {
				sectionEndLines = append(sectionEndLines, rawNodes[i+1].LineNo)
			} else {
				sectionEndLines = append(sectionEndLines, len(lines)+1)
			}
		}
	}

	// 2. Scan section bodies for tasks and internal anchor links
	for i := range rawNodes {
		startLine := rawNodes[i].LineNo
		endLine := sectionEndLines[i]

		for l := startLine - 1; l < endLine-1 && l < len(lines); l++ {
			lineText := lines[l]

			// Tasks / Checklists
			if mChk := chkBoxRegex.FindStringSubmatch(lineText); mChk != nil {
				rawNodes[i].TotalTasks++
				if strings.ToLower(mChk[1]) == "x" {
					rawNodes[i].CompletedTasks++
				}
				rawNodes[i].Tasks = append(rawNodes[i].Tasks, strings.TrimSpace(mChk[2]))
			}

			// Internal cross-links [text](#anchor)
			for _, mLink := range linkRefRegex.FindAllStringSubmatch(lineText, -1) {
				if len(mLink) >= 2 {
					targetAnchor := mLink[1]
					rawNodes[i].LinksTo = append(rawNodes[i].LinksTo, targetAnchor)
				}
			}
		}
	}

	// 3. Build tree hierarchy
	var rootIDs []int
	var stack []*Node

	for i := range rawNodes {
		node := &rawNodes[i]

		// Pop stack until parent has smaller level
		for len(stack) > 0 && stack[len(stack)-1].Level >= node.Level {
			stack = stack[:len(stack)-1]
		}

		if len(stack) == 0 {
			node.ParentID = -1
			rootIDs = append(rootIDs, node.ID)
		} else {
			parent := stack[len(stack)-1]
			node.ParentID = parent.ID
			parent.ChildrenIDs = append(parent.ChildrenIDs, node.ID)
		}

		stack = append(stack, node)
	}

	// 4. Correlate with diff hunks & review comments
	for i := range rawNodes {
		startLine := rawNodes[i].LineNo
		endLine := sectionEndLines[i]

		// Diff additions & deletions in this line span
		for _, hunk := range hunks {
			for _, dl := range hunk.Lines {
				if dl.Type == diff.LineAdded && dl.NewLineNo >= startLine && dl.NewLineNo < endLine {
					rawNodes[i].Additions++
				}
				// If deleted line matches section heading or hunk section
				if dl.Type == diff.LineDeleted {
					if hunk.Section == rawNodes[i].RawHeading || dl.Section == rawNodes[i].RawHeading {
						rawNodes[i].Deletions++
					}
				}
			}
		}

		// Comments in this line span
		if threadMap != nil {
			for l := startLine; l < endLine; l++ {
				rawNodes[i].CommentCount += len(threadMap[l])
			}
		}
	}

	// 5. Build pre-order flattened traversal order
	var flattenedOrder []int
	var traverse func(nodeID int)
	traverse = func(nodeID int) {
		flattenedOrder = append(flattenedOrder, nodeID)
		for _, childID := range rawNodes[nodeID].ChildrenIDs {
			traverse(childID)
		}
	}

	for _, rootID := range rootIDs {
		traverse(rootID)
	}

	// 6. Build index map and edges
	nodeByAnchor := make(map[string]int, len(rawNodes))
	for _, n := range rawNodes {
		nodeByAnchor[n.Anchor] = n.ID
	}

	var edges []Edge
	for _, n := range rawNodes {
		for _, childID := range n.ChildrenIDs {
			edges = append(edges, Edge{
				FromID: n.ID,
				ToID:   childID,
				Type:   "hierarchy",
			})
		}
		for _, targetAnchor := range n.LinksTo {
			if targetID, ok := nodeByAnchor[targetAnchor]; ok {
				edges = append(edges, Edge{
					FromID: n.ID,
					ToID:   targetID,
					Type:   "crosslink",
				})
			}
		}
	}

	return &Graph{
		Nodes:          rawNodes,
		RootIDs:        rootIDs,
		Edges:          edges,
		NodeByAnchor:   nodeByAnchor,
		FlattenedOrder: flattenedOrder,
	}
}
