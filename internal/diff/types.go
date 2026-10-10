package diff

// LineType indicates the kind of diff line.
type LineType int

const (
	LineContext LineType = iota
	LineAdded
	LineDeleted
	LineHunkHeader
)

// DiffSegment is an intra-line slice of text with diff highlighting.
type DiffSegment struct {
	Text      string
	Highlight bool // true if this word/token was modified
}

// DiffLine represents a single line in a diff hunk.
type DiffLine struct {
	Type      LineType
	OldLineNo int // 1-based line number in base (0 if added)
	NewLineNo int // 1-based line number in head (0 if deleted)
	Content   string
	Prefix    string // " ", "+", "-", "@@"
	Segments  []DiffSegment
	Section   string // Markdown section heading this line falls under
}

// DiffHunk groups a contiguous block of diff changes.
type DiffHunk struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Header   string // e.g. "@@ -10,5 +12,6 @@"
	Section  string // e.g. "## Architecture"
	Lines    []DiffLine
}

// SplitRow represents a synchronized row in side-by-side view.
type SplitRow struct {
	IsHeader   bool
	HeaderLine string
	Left       *DiffLine // Base/Before line (nil if line was added on right)
	Right      *DiffLine // Head/After line (nil if line was deleted on left)
}

// SectionChange captures changes under a specific markdown heading.
type SectionChange struct {
	Heading   string
	Level     int // 1 for #, 2 for ##, etc.
	Additions int
	Deletions int
}

// ChecklistChange captures changes to task list items (- [ ] / - [x]).
type ChecklistChange struct {
	Task   string
	Status string // "completed", "unchecked", "added", "removed"
}

// IntentSummary summarizes the high-level semantic intent of changes.
type IntentSummary struct {
	Sections       []SectionChange
	Checklists     []ChecklistChange
	TotalAdditions int
	TotalDeletions int
	HunkCount      int
}

// FileDiff holds all processed diff structures for a markdown file.
type FileDiff struct {
	Path         string
	Hunks        []DiffHunk
	UnifiedLines []DiffLine
	SplitRows    []SplitRow
	Intent       IntentSummary
	HeadContent  string
	BaseContent  string
}
