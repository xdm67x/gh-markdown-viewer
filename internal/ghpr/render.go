package ghpr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
)

// commentableLinesOutsideFences returns the subset of keys in commentable that
// fall outside fenced code blocks (``` or ~~~), in source order.
func commentableLinesOutsideFences(text string, commentable map[int]bool) []int {
	var out []int
	inFence := false
	for i, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			continue
		}
		if !inFence && commentable[i+1] {
			out = append(out, i+1)
		}
	}
	return out
}

// injectLineMarkers appends <a id="md-src-N"> at the END of each injectable
// line. End-injection preserves block-level syntax: "# Heading<a...>" keeps
// the # at position 0; prepending would break heading/list/blockquote parsing.
//
// Skipped even when commentable:
//   - blank lines (injection would collapse paragraph breaks)
//   - table rows (lines starting with |, pipe structure is strict)
//
// Returns the annotated text and the line numbers that were actually injected
// (a subset of lineNums after skipping the above).
func injectLineMarkers(text string, lineNums []int) (annotated string, injected []int) {
	if len(lineNums) == 0 {
		return text, nil
	}
	set := make(map[int]bool, len(lineNums))
	for _, n := range lineNums {
		set[n] = true
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if !set[i+1] {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "|") {
			continue
		}
		lines[i] = line + `<a id="md-src-` + strconv.Itoa(i+1) + `"></a>`
		injected = append(injected, i+1)
	}
	return strings.Join(lines, "\n"), injected
}

// RenderMarkdownAnnotated renders GFM markdown with line-marker anchors injected
// at each commentable line (outside fenced code blocks). Returns the HTML and
// the sorted slice of injected line numbers so the browser knows which anchors
// to wire up for inline commenting.
func RenderMarkdownAnnotated(info PRInfo, text string, commentable map[int]bool) (html string, injected []int, err error) {
	candidates := commentableLinesOutsideFences(text, commentable)
	annotated, injected := injectLineMarkers(text, candidates)
	html, err = renderRaw(info, annotated)
	if err != nil {
		return "", nil, err
	}
	return html, injected, nil
}

func renderRaw(info PRInfo, text string) (string, error) {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return "", err
	}

	payload, err := json.Marshal(map[string]string{
		"text":    text,
		"mode":    "gfm",
		"context": info.Ref.Owner + "/" + info.Ref.Repo,
	})
	if err != nil {
		return "", err
	}

	resp, err := client.Request("POST", "markdown", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("rendering markdown: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading markdown response: %w", err)
	}
	return string(raw), nil
}
