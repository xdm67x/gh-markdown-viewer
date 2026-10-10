package diff

import (
	"github.com/sergi/go-diff/diffmatchpatch"
)

// applyWordDiff computes intra-line diff highlights for pairs of modified lines in hunks.
func applyWordDiff(hunks []DiffHunk) []DiffHunk {
	dmp := diffmatchpatch.New()

	for hIdx := range hunks {
		hunk := &hunks[hIdx]
		lines := hunk.Lines

		i := 0
		for i < len(lines) {
			// Find consecutive deleted lines
			delStart := -1
			if lines[i].Type == LineDeleted {
				delStart = i
				for i < len(lines) && lines[i].Type == LineDeleted {
					i++
				}
			}
			delEnd := i

			// Check if followed by consecutive added lines
			addStart := -1
			if i < len(lines) && lines[i].Type == LineAdded {
				addStart = i
				for i < len(lines) && lines[i].Type == LineAdded {
					i++
				}
			}
			addEnd := i

			if delStart != -1 && addStart != -1 {
				delCount := delEnd - delStart
				addCount := addEnd - addStart
				minCount := delCount
				if addCount < minCount {
					minCount = addCount
				}

				for p := 0; p < minCount; p++ {
					delLine := &lines[delStart+p]
					addLine := &lines[addStart+p]

					diffs := dmp.DiffMain(delLine.Content, addLine.Content, false)
					diffs = dmp.DiffCleanupSemantic(diffs)

					var delSegs []DiffSegment
					var addSegs []DiffSegment

					for _, d := range diffs {
						switch d.Type {
						case diffmatchpatch.DiffEqual:
							delSegs = append(delSegs, DiffSegment{Text: d.Text, Highlight: false})
							addSegs = append(addSegs, DiffSegment{Text: d.Text, Highlight: false})
						case diffmatchpatch.DiffDelete:
							delSegs = append(delSegs, DiffSegment{Text: d.Text, Highlight: true})
						case diffmatchpatch.DiffInsert:
							addSegs = append(addSegs, DiffSegment{Text: d.Text, Highlight: true})
						}
					}

					delLine.Segments = delSegs
					addLine.Segments = addSegs
				}

				// Unpaired deleted lines
				for p := minCount; p < delCount; p++ {
					lines[delStart+p].Segments = []DiffSegment{{
						Text:      lines[delStart+p].Content,
						Highlight: false,
					}}
				}
				// Unpaired added lines
				for p := minCount; p < addCount; p++ {
					lines[addStart+p].Segments = []DiffSegment{{
						Text:      lines[addStart+p].Content,
						Highlight: false,
					}}
				}
			} else {
				// Single block or context line
				if delStart != -1 {
					for p := delStart; p < delEnd; p++ {
						lines[p].Segments = []DiffSegment{{
							Text:      lines[p].Content,
							Highlight: false,
						}}
					}
				} else if addStart != -1 {
					for p := addStart; p < addEnd; p++ {
						lines[p].Segments = []DiffSegment{{
							Text:      lines[p].Content,
							Highlight: false,
						}}
					}
				} else {
					// Context line
					lines[i].Segments = []DiffSegment{{
						Text:      lines[i].Content,
						Highlight: false,
					}}
					i++
				}
			}
		}
	}

	return hunks
}
