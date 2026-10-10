package diff

// buildSplitRows converts hunks into paired rows for side-by-side display.
func buildSplitRows(hunks []DiffHunk) []SplitRow {
	var rows []SplitRow

	for _, hunk := range hunks {
		// Header row
		headerText := hunk.Header
		if hunk.Section != "" {
			headerText += " " + hunk.Section
		}
		rows = append(rows, SplitRow{
			IsHeader:   true,
			HeaderLine: headerText,
		})

		lines := hunk.Lines
		i := 0
		for i < len(lines) {
			if lines[i].Type == LineContext {
				lineCopy := lines[i]
				rows = append(rows, SplitRow{
					Left:  &lineCopy,
					Right: &lineCopy,
				})
				i++
				continue
			}

			// Gather consecutive deletions
			var deletions []DiffLine
			for i < len(lines) && lines[i].Type == LineDeleted {
				deletions = append(deletions, lines[i])
				i++
			}

			// Gather consecutive additions
			var additions []DiffLine
			for i < len(lines) && lines[i].Type == LineAdded {
				additions = append(additions, lines[i])
				i++
			}

			maxLen := len(deletions)
			if len(additions) > maxLen {
				maxLen = len(additions)
			}

			for p := 0; p < maxLen; p++ {
				var left, right *DiffLine
				if p < len(deletions) {
					delCopy := deletions[p]
					left = &delCopy
				}
				if p < len(additions) {
					addCopy := additions[p]
					right = &addCopy
				}
				rows = append(rows, SplitRow{
					Left:  left,
					Right: right,
				})
			}
		}
	}

	return rows
}
