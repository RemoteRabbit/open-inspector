// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// commentIndex maps a block's start byte offset (block.DefRange.Start.Byte) to the cleaned
// text of the contiguous comment run directly above it. It is built once per native HCL file.
// JSON files have no comments and yield a nil index.
type commentIndex map[int]string

// buildCommentIndex lexes a native HCL file and records, for each run of comment tokens
// that butts up against the start of the next token, the cleaned comment text keyed by
// that next token's start byte. The block decoders look up block.DefRange.Start.Byte to
// retrieve a block's leading comment. Lexer diagnostics are ignored: the file already parsed
// cleanly for the amin decode pass, and comment association is best-effort.
func buildCommentIndex(filename string, source []byte) commentIndex {
	tokens, _ := hclsyntax.LexConfig(source, filename, hcl.InitialPos)
	index := make(commentIndex)

	var run []hclsyntax.Token
	flush := func(next hclsyntax.Token) {
		if len(run) == 0 {
			return
		}

		// Associate only if no blank line separates the last comment from the next token.
		// TokenComment for # and // lines includes its trailing newline, so an adjacent
		// token starts on the line right after; a blank line shows up as a >1 line gap.
		last := run[len(run)-1]
		if next.Range.Start.Line-last.Range.End.Line <= 1 {
			if text := cleanComments(run); text != "" {
				index[next.Range.Start.Byte] = text
			}
		}
		run = nil
	}

	for _, token := range tokens {
		switch token.Type {
		case hclsyntax.TokenComment:
			run = append(run, token)
		case hclsyntax.TokenNewline:
			// A standalone newline between comments keeps the run going; a blank line (newline
			// with no preceding comment-adjacency) is handled by the line-gap check in flush.
			continue
		default:
			flush(token)
		}
	}
	return index
}

// cleanComments strips comment markers (#, //, /* */, leading *) and joins the lines of a
// comment run into a single string.
func cleanComments(run []hclsyntax.Token) string {
	var lines []string
	for _, token := range run {
		text := string(token.Bytes)
		switch {
		case strings.HasPrefix(text, "#"):
			lines = append(lines, cleanLine(strings.TrimPrefix(text, "#")))
		case strings.HasPrefix(text, "//"):
			lines = append(lines, cleanLine(strings.TrimPrefix(text, "//")))
		case strings.HasPrefix(text, "/*"):
			body := strings.TrimSuffix(strings.TrimPrefix(text, "/*"), "*/")
			for _, raw := range strings.Split(body, "\n") {
				lines = append(lines, cleanLine(strings.TrimPrefix(strings.TrimSpace(raw), "*")))
			}
		}
	}
	// Drop leading and trailing blank lines, keep interior ones.
	for len(lines) > 0 && lines[0] == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

// cleanLine trims a single comment line: from the trailing newline the lexar attaches
// to # and // comments, then one optional leading space.
func cleanLine(line string) string {
	line = strings.TrimRight(line, "\r\n")
	return strings.TrimPrefix(line, " ")
}
