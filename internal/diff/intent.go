package diff

import (
	"regexp"
	"strings"
)

var (
	headingRegex   = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	chkBoxRegex    = regexp.MustCompile(`^[\s\-\*\+]*\[([ xX])\]\s*(.*)$`)
	chkBoxItemTest = regexp.MustCompile(`^[\s\-\*\+]*\[[ xX]\]`)
)

// extractIntent analyzes changed sections and task checklists.
func extractIntent(hunks []DiffHunk) IntentSummary {
	var summary IntentSummary
	summary.HunkCount = len(hunks)

	sectionMap := make(map[string]*SectionChange)
	var sectionOrder []string

	currentSection := "(Top Level)"
	currentLevel := 1

	var delChecklists []string
	var delChecked []bool
	var addChecklists []string
	var addChecked []bool

	for _, hunk := range hunks {
		if hunk.Section != "" {
			currentSection = hunk.Section
			m := headingRegex.FindStringSubmatch(currentSection)
			if m != nil {
				currentLevel = len(m[1])
			}
		}

		for _, line := range hunk.Lines {
			switch line.Type {
			case LineHunkHeader:
				if line.Section != "" {
					currentSection = line.Section
					m := headingRegex.FindStringSubmatch(currentSection)
					if m != nil {
						currentLevel = len(m[1])
					}
				}

			case LineContext:
				m := headingRegex.FindStringSubmatch(line.Content)
				if m != nil {
					currentSection = line.Content
					currentLevel = len(m[1])
				}

			case LineDeleted:
				summary.TotalDeletions++
				mHead := headingRegex.FindStringSubmatch(line.Content)
				sec := currentSection
				lvl := currentLevel
				if mHead != nil {
					sec = line.Content
					lvl = len(mHead[1])
				}

				if _, exists := sectionMap[sec]; !exists {
					sectionMap[sec] = &SectionChange{
						Heading: sec,
						Level:   lvl,
					}
					sectionOrder = append(sectionOrder, sec)
				}
				sectionMap[sec].Deletions++

				if chkBoxItemTest.MatchString(line.Content) {
					sub := chkBoxRegex.FindStringSubmatch(line.Content)
					if len(sub) >= 3 {
						delChecklists = append(delChecklists, strings.TrimSpace(sub[2]))
						delChecked = append(delChecked, strings.ToLower(sub[1]) == "x")
					}
				}

			case LineAdded:
				summary.TotalAdditions++
				mHead := headingRegex.FindStringSubmatch(line.Content)
				sec := currentSection
				lvl := currentLevel
				if mHead != nil {
					sec = line.Content
					lvl = len(mHead[1])
					currentSection = line.Content
					currentLevel = lvl
				}

				if _, exists := sectionMap[sec]; !exists {
					sectionMap[sec] = &SectionChange{
						Heading: sec,
						Level:   lvl,
					}
					sectionOrder = append(sectionOrder, sec)
				}
				sectionMap[sec].Additions++

				if chkBoxItemTest.MatchString(line.Content) {
					sub := chkBoxRegex.FindStringSubmatch(line.Content)
					if len(sub) >= 3 {
						addChecklists = append(addChecklists, strings.TrimSpace(sub[2]))
						addChecked = append(addChecked, strings.ToLower(sub[1]) == "x")
					}
				}
			}
		}
	}

	for _, sec := range sectionOrder {
		summary.Sections = append(summary.Sections, *sectionMap[sec])
	}

	// Match checklist transitions
	matchedDel := make(map[int]bool)
	matchedAdd := make(map[int]bool)

	for aIdx, addTask := range addChecklists {
		for dIdx, delTask := range delChecklists {
			if matchedDel[dIdx] {
				continue
			}
			if addTask == delTask {
				// Same task, toggle state
				matchedDel[dIdx] = true
				matchedAdd[aIdx] = true
				if addChecked[aIdx] && !delChecked[dIdx] {
					summary.Checklists = append(summary.Checklists, ChecklistChange{
						Task:   addTask,
						Status: "completed",
					})
				} else if !addChecked[aIdx] && delChecked[dIdx] {
					summary.Checklists = append(summary.Checklists, ChecklistChange{
						Task:   addTask,
						Status: "unchecked",
					})
				}
				break
			}
		}
	}

	for aIdx, addTask := range addChecklists {
		if !matchedAdd[aIdx] {
			status := "added"
			if addChecked[aIdx] {
				status = "completed"
			}
			summary.Checklists = append(summary.Checklists, ChecklistChange{
				Task:   addTask,
				Status: status,
			})
		}
	}

	for dIdx, delTask := range delChecklists {
		if !matchedDel[dIdx] {
			summary.Checklists = append(summary.Checklists, ChecklistChange{
				Task:   delTask,
				Status: "removed",
			})
		}
	}

	return summary
}
