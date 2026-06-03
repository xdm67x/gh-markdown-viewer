package ghpr

import (
	"regexp"
	"strconv"
	"strings"
)

// reHunkHeader matches "@@ -a,b +c,d @@" and captures the post-change start
// line (group 1) and optional count (group 2).
var reHunkHeader = regexp.MustCompile(`@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@`)

// CommentableLines returns the set of post-change (RIGHT-side) line numbers
// that GitHub will accept a review comment on. This includes added lines (+)
// and context lines ( ) inside each diff hunk.
//
// Deleted lines (-) are on the LEFT side; we do not include them because we
// always render the post-change content of the file.
func CommentableLines(patch string) map[int]bool {
	result := make(map[int]bool)
	if patch == "" {
		return result
	}

	var lineNo int // current post-change line number being tracked
	for _, line := range strings.Split(patch, "\n") {
		if m := reHunkHeader.FindStringSubmatch(line); m != nil {
			start, _ := strconv.Atoi(m[1])
			lineNo = start
			continue
		}
		switch {
		case strings.HasPrefix(line, "+"):
			result[lineNo] = true
			lineNo++
		case strings.HasPrefix(line, " "):
			result[lineNo] = true
			lineNo++
		case strings.HasPrefix(line, "-"):
			// deleted line: does not advance post-change counter
		}
	}
	return result
}
