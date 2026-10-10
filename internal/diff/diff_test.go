package diff

import (
	"testing"
)

const samplePatch = `@@ -1,6 +1,8 @@ ## Implementation Plan
 Context line 1
-Old intro sentence.
+New intro sentence.
-- [ ] Task 1: Initialize harness
+- [x] Task 1: Initialize harness
+- [ ] Task 2: Add validation
 Context line 2
@@ -20,4 +22,5 @@ ## Architecture
  Existing arch line
+New component: Agent Runner
  Ending arch line
`

func TestParsePatch(t *testing.T) {
	hunks := ParsePatch(samplePatch)
	if len(hunks) != 2 {
		t.Fatalf("expected 2 hunks, got %d", len(hunks))
	}

	hunk1 := hunks[0]
	if hunk1.Section != "## Implementation Plan" {
		t.Errorf("expected section '## Implementation Plan', got %q", hunk1.Section)
	}
	if hunk1.OldStart != 1 || hunk1.NewStart != 1 {
		t.Errorf("expected starts 1, 1; got %d, %d", hunk1.OldStart, hunk1.NewStart)
	}

	hunk2 := hunks[1]
	if hunk2.Section != "## Architecture" {
		t.Errorf("expected section '## Architecture', got %q", hunk2.Section)
	}
}

func TestWordDiff(t *testing.T) {
	hunks := ParsePatch(samplePatch)
	hunks = applyWordDiff(hunks)

	hunk1 := hunks[0]
	// Line 1 is context
	// Line 2 is deleted: "-Old intro sentence."
	// Line 3 is added: "+New intro sentence."
	delLine := hunk1.Lines[1]
	addLine := hunk1.Lines[2]

	if delLine.Type != LineDeleted || addLine.Type != LineAdded {
		t.Fatalf("unexpected line types: %v, %v", delLine.Type, addLine.Type)
	}

	// Verify segments: "Old" vs "New" should be highlighted
	hasDelHighlight := false
	for _, s := range delLine.Segments {
		if s.Highlight && s.Text == "Old" {
			hasDelHighlight = true
		}
	}
	if !hasDelHighlight {
		t.Errorf("expected deleted segment 'Old' to be highlighted, got %+v", delLine.Segments)
	}

	hasAddHighlight := false
	for _, s := range addLine.Segments {
		if s.Highlight && s.Text == "New" {
			hasAddHighlight = true
		}
	}
	if !hasAddHighlight {
		t.Errorf("expected added segment 'New' to be highlighted, got %+v", addLine.Segments)
	}
}

func TestBuildSplitRows(t *testing.T) {
	hunks := ParsePatch(samplePatch)
	hunks = applyWordDiff(hunks)
	rows := buildSplitRows(hunks)

	if len(rows) == 0 {
		t.Fatalf("expected split rows, got none")
	}

	// First row should be hunk header
	if !rows[0].IsHeader {
		t.Errorf("expected first row to be header")
	}
}

func TestExtractIntent(t *testing.T) {
	hunks := ParsePatch(samplePatch)
	summary := extractIntent(hunks)

	if summary.TotalAdditions == 0 || summary.TotalDeletions == 0 {
		t.Errorf("expected non-zero additions/deletions, got +%d -%d",
			summary.TotalAdditions, summary.TotalDeletions)
	}

	if len(summary.Sections) < 2 {
		t.Errorf("expected at least 2 changed sections, got %d", len(summary.Sections))
	}

	// Check checklist detection
	var foundCompleted bool
	for _, chk := range summary.Checklists {
		if chk.Status == "completed" && chk.Task == "Task 1: Initialize harness" {
			foundCompleted = true
			break
		}
	}
	if !foundCompleted {
		t.Errorf("expected checklist 'Task 1: Initialize harness' to be marked completed, got %+v",
			summary.Checklists)
	}
}

func TestNewFileDiff(t *testing.T) {
	fd := NewFileDiff("test.md", samplePatch, "", "")
	if fd.Path != "test.md" {
		t.Errorf("expected path 'test.md', got %q", fd.Path)
	}
	if len(fd.UnifiedLines) == 0 {
		t.Errorf("expected unified lines, got empty")
	}
	if len(fd.SplitRows) == 0 {
		t.Errorf("expected split rows, got empty")
	}
}

func TestGeneratePatch(t *testing.T) {
	base := "# Hello\nThis is base."
	head := "# Hello\nThis is head."
	patch := GeneratePatch(base, head)
	if patch == "" {
		t.Fatalf("expected non-empty patch generated")
	}

	fd := NewFileDiff("generated.md", "", base, head)
	if fd.Intent.TotalAdditions == 0 {
		t.Errorf("expected additions in generated diff")
	}
}

func TestChecklistUnchecked(t *testing.T) {
	patch := `@@ -1,2 +1,2 @@
-- [x] Task 1: Initialize harness
+- [ ] Task 1: Initialize harness
 `
	hunks := ParsePatch(patch)
	summary := extractIntent(hunks)
	if len(summary.Checklists) == 0 {
		t.Fatalf("expected checklist change")
	}
	if summary.Checklists[0].Status != "unchecked" {
		t.Errorf("expected status 'unchecked', got %q", summary.Checklists[0].Status)
	}
}
