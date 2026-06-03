// Package srcmap maps between rendered viewport lines and source markdown lines.
//
// Glamour reflows source markdown so a rendered line does not correspond
// directly to a source line. This package builds a span table by:
//  1. Parsing the source with goldmark to get top-level block AST nodes.
//  2. Rendering each block independently with Glamour, counting output lines.
//  3. Storing (srcStart, srcEnd, rendStart, rendEnd) spans.
//
// The mapping is approximate at sub-block granularity (e.g. inside a 10-line
// paragraph the interpolation is linear). This is acceptable: we only need to
// anchor comments to a source line within the same top-level block.
package srcmap

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Block holds the source line range of one top-level markdown block.
type Block struct {
	SrcStart int // 1-based, inclusive
	SrcEnd   int // 1-based, inclusive
	// Raw source text of this block (slice of the original []byte).
	Source []byte
}

// ParseBlocks parses src and returns the top-level block nodes with their
// source line ranges.
func ParseBlocks(src []byte) []Block {
	parser := goldmark.DefaultParser()
	reader := text.NewReader(src)
	doc := parser.Parse(reader)

	// Build a byte-offset → line-number table.
	lineStarts := buildLineStarts(src)

	var blocks []Block
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		segs := n.Lines()
		if segs == nil || segs.Len() == 0 {
			// Leaf blocks (e.g. code fences) store their content differently.
			// Fall back to the block's own HasChildren walk.
			segs = collectLines(n)
		}
		if segs == nil || segs.Len() == 0 {
			continue
		}
		first := segs.At(0)
		last := segs.At(segs.Len() - 1)

		srcStart := offsetToLine(first.Start, lineStarts)
		srcEnd := offsetToLine(last.Stop-1, lineStarts) // Stop is exclusive

		// Slice the raw source for this block's lines.
		blockSrc := src[first.Start:last.Stop]
		// Strip any trailing blank line so per-block renders are clean.
		blockSrc = bytes.TrimRight(blockSrc, "\n")

		blocks = append(blocks, Block{
			SrcStart: srcStart,
			SrcEnd:   srcEnd,
			Source:   blockSrc,
		})
	}
	return blocks
}

// buildLineStarts returns a slice where lineStarts[i] is the byte offset of
// the first character on line i+1 (1-based lines, 0-based slice).
func buildLineStarts(src []byte) []int {
	starts := []int{0}
	for i, b := range src {
		if b == '\n' && i+1 < len(src) {
			starts = append(starts, i+1)
		}
	}
	return starts
}

// offsetToLine converts a byte offset into a 1-based line number.
func offsetToLine(offset int, lineStarts []int) int {
	lo, hi := 0, len(lineStarts)-1
	for lo <= hi {
		mid := (lo + hi) / 2
		if lineStarts[mid] <= offset {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	return hi + 1 // hi is the last start ≤ offset, +1 for 1-based
}

// collectLines gathers lines from a node's children (for nodes without their
// own Lines, e.g. fenced code blocks in some goldmark versions).
// Inline nodes must be skipped — calling Lines() on them panics.
func collectLines(n ast.Node) *text.Segments {
	segs := &text.Segments{}
	_ = ast.Walk(n, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if node.Type() == ast.TypeInline {
			return ast.WalkSkipChildren, nil
		}
		if lines := node.Lines(); lines != nil {
			for i := 0; i < lines.Len(); i++ {
				segs.Append(lines.At(i))
			}
		}
		return ast.WalkContinue, nil
	})
	return segs
}
