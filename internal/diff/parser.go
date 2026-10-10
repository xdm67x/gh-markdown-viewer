package diff

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
)

var hunkHeaderRegex = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@(?:\s*(.*))?`)

// ParsePatch parses a unified diff patch string into a slice of DiffHunk.
func ParsePatch(patch string) []DiffHunk {
	if strings.TrimSpace(patch) == "" {
		return nil
	}

	var hunks []DiffHunk
	lines := strings.Split(patch, "\n")

	var currentHunk *DiffHunk
	var oldLineNo, newLineNo int
	currentSection := ""

	for _, rawLine := range lines {
		if rawLine == "" && currentHunk == nil {
			continue
		}

		if m := hunkHeaderRegex.FindStringSubmatch(rawLine); m != nil {
			if currentHunk != nil {
				hunks = append(hunks, *currentHunk)
			}

			oldStart, _ := strconv.Atoi(m[1])
			oldCount := 1
			if m[2] != "" {
				oldCount, _ = strconv.Atoi(m[2])
			}

			newStart, _ := strconv.Atoi(m[3])
			newCount := 1
			if m[4] != "" {
				newCount, _ = strconv.Atoi(m[4])
			}

			sec := strings.TrimSpace(m[5])
			if sec != "" {
				currentSection = sec
			}

			hunkHeaderOnly := fmt.Sprintf("@@ -%s +%s @@",
				formatHunkRange(oldStart, oldCount),
				formatHunkRange(newStart, newCount),
			)

			currentHunk = &DiffHunk{
				OldStart: oldStart,
				OldCount: oldCount,
				NewStart: newStart,
				NewCount: newCount,
				Header:   hunkHeaderOnly,
				Section:  currentSection,
			}

			oldLineNo = oldStart
			newLineNo = newStart
			continue
		}

		if currentHunk == nil {
			continue
		}

		if strings.HasPrefix(rawLine, "\\ No newline") {
			continue
		}

		switch {
		case strings.HasPrefix(rawLine, "+"):
			content := rawLine[1:]
			currentHunk.Lines = append(currentHunk.Lines, DiffLine{
				Type:      LineAdded,
				OldLineNo: 0,
				NewLineNo: newLineNo,
				Content:   content,
				Prefix:    "+",
				Section:   currentSection,
			})
			newLineNo++

		case strings.HasPrefix(rawLine, "-"):
			content := rawLine[1:]
			currentHunk.Lines = append(currentHunk.Lines, DiffLine{
				Type:      LineDeleted,
				OldLineNo: oldLineNo,
				NewLineNo: 0,
				Content:   content,
				Prefix:    "-",
				Section:   currentSection,
			})
			oldLineNo++

		case strings.HasPrefix(rawLine, " "):
			content := rawLine[1:]
			currentHunk.Lines = append(currentHunk.Lines, DiffLine{
				Type:      LineContext,
				OldLineNo: oldLineNo,
				NewLineNo: newLineNo,
				Content:   content,
				Prefix:    " ",
				Section:   currentSection,
			})
			oldLineNo++
			newLineNo++

		case rawLine == "":
			// Empty context line in some unified diffs
			currentHunk.Lines = append(currentHunk.Lines, DiffLine{
				Type:      LineContext,
				OldLineNo: oldLineNo,
				NewLineNo: newLineNo,
				Content:   "",
				Prefix:    " ",
				Section:   currentSection,
			})
			oldLineNo++
			newLineNo++
		}
	}

	if currentHunk != nil {
		hunks = append(hunks, *currentHunk)
	}

	return hunks
}

func formatHunkRange(start, count int) string {
	if count == 1 {
		return strconv.Itoa(start)
	}
	return fmt.Sprintf("%d,%d", start, count)
}

// GeneratePatch creates a unified diff patch between base and head content.
func GeneratePatch(baseText, headText string) string {
	if baseText == "" && headText == "" {
		return ""
	}

	if baseText == "" {
		lines := strings.Split(headText, "\n")
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("@@ -0,0 +1,%d @@\n", len(lines)))
		for _, l := range lines {
			sb.WriteString("+" + l + "\n")
		}
		return sb.String()
	}

	if headText == "" {
		lines := strings.Split(baseText, "\n")
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("@@ -1,%d +0,0 @@\n", len(lines)))
		for _, l := range lines {
			sb.WriteString("-" + l + "\n")
		}
		return sb.String()
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(baseText, headText, false)
	diffs = dmp.DiffCleanupSemantic(diffs)
	patches := dmp.PatchMake(baseText, diffs)
	return dmp.PatchToText(patches)
}

// NewFileDiff constructs a complete FileDiff model with intra-line diffs,
// unified lines, side-by-side rows, and intent summary.
func NewFileDiff(path, patch, baseContent, headContent string) *FileDiff {
	if patch == "" && (baseContent != "" || headContent != "") {
		patch = GeneratePatch(baseContent, headContent)
	}

	hunks := ParsePatch(patch)
	hunks = applyWordDiff(hunks)

	// Build unified lines
	var unifiedLines []DiffLine
	for _, hunk := range hunks {
		// Insert hunk header line as a navigable/renderable line
		hdrText := hunk.Header
		if hunk.Section != "" {
			hdrText += " " + hunk.Section
		}
		unifiedLines = append(unifiedLines, DiffLine{
			Type:    LineHunkHeader,
			Content: hdrText,
			Prefix:  "@@",
			Section: hunk.Section,
		})
		unifiedLines = append(unifiedLines, hunk.Lines...)
	}

	splitRows := buildSplitRows(hunks)
	intent := extractIntent(hunks)

	return &FileDiff{
		Path:         path,
		Hunks:        hunks,
		UnifiedLines: unifiedLines,
		SplitRows:    splitRows,
		Intent:       intent,
		HeadContent:  headContent,
		BaseContent:  baseContent,
	}
}
