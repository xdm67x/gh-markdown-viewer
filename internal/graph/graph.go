package graph

// Node represents a section / heading node in the Markdown graph.
type Node struct {
	ID             int
	Title          string
	RawHeading     string
	Anchor         string
	Level          int // 1 for #, 2 for ##, etc.
	LineNo         int // 1-based line in head content
	ParentID       int // -1 if root
	ChildrenIDs    []int
	LinksTo        []string // Anchor slugs referenced in section body
	Additions      int
	Deletions      int
	TotalTasks     int
	CompletedTasks int
	CommentCount   int
	Tasks          []string
}

// Edge represents a directed relation between sections (parent->child or link->target).
type Edge struct {
	FromID int
	ToID   int
	Type   string // "hierarchy" or "crosslink"
}

// Graph models the full hierarchical and cross-reference graph of a Markdown document.
type Graph struct {
	Nodes          []Node
	RootIDs        []int
	Edges          []Edge
	NodeByAnchor   map[string]int
	FlattenedOrder []int // Pre-order traversal IDs for linear list navigation
}

// NodeAt returns the node for the given flattened index.
func (g *Graph) NodeAt(flatIdx int) *Node {
	if flatIdx < 0 || flatIdx >= len(g.FlattenedOrder) {
		return nil
	}
	id := g.FlattenedOrder[flatIdx]
	return &g.Nodes[id]
}

// TotalNodes returns the count of sections in the graph.
func (g *Graph) TotalNodes() int {
	return len(g.Nodes)
}
