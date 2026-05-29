package search

import (
	"regexp"
	"strings"
)

var (
	reHTMLTag    = regexp.MustCompile(`<[^>]+>`)
	reCodeFence  = regexp.MustCompile("(?m)^```[^\\n]*\\n([\\s\\S]*?)\\n```")
	reInlineCode = regexp.MustCompile("`([^`]+)`")
	reImage      = regexp.MustCompile(`!\[([^\]]*)\]\([^)]*\)`)
	reLink       = regexp.MustCompile(`\[([^\]]+)\]\([^)]*\)`)
	reHeading    = regexp.MustCompile(`(?m)^#{1,6}\s+`)
	// Strip bold (*** / ** / *) and italic (_ / __) markers individually.
	reBoldTriple  = regexp.MustCompile(`\*{3}([^*]+)\*{3}`)
	reBoldDouble  = regexp.MustCompile(`\*{2}([^*]+)\*{2}`)
	reBoldSingle  = regexp.MustCompile(`\*([^*]+)\*`)
	reUnderDouble = regexp.MustCompile(`__([^_]+)__`)
	reUnderSingle = regexp.MustCompile(`_([^_]+)_`)
	reBlockquote  = regexp.MustCompile(`(?m)^>\s*`)
	reHRule       = regexp.MustCompile(`(?m)^[-*_]{3,}\s*$`)
	reTableRow    = regexp.MustCompile(`\|`)
	reMultiBlank  = regexp.MustCompile(`\n{3,}`)
)

// Clean strips Markdown syntax from content and returns plain text suitable
// for FTS indexing. It preserves the text content of headings, links, images,
// emphasis, code blocks, and blockquotes while removing all markers.
func Clean(content string) string {
	// Fenced code blocks: keep inner content.
	s := reCodeFence.ReplaceAllString(content, "\n$1\n")

	// Inline code: keep content.
	s = reInlineCode.ReplaceAllString(s, "$1")

	// Images: keep alt text.
	s = reImage.ReplaceAllString(s, "$1")

	// Links: keep link text.
	s = reLink.ReplaceAllString(s, "$1")

	// Headings: strip # markers (keep text).
	s = reHeading.ReplaceAllString(s, "")

	// Bold / italic: strip markers, keep inner text (longest match first).
	s = reBoldTriple.ReplaceAllString(s, "$1")
	s = reBoldDouble.ReplaceAllString(s, "$1")
	s = reBoldSingle.ReplaceAllString(s, "$1")
	s = reUnderDouble.ReplaceAllString(s, "$1")
	s = reUnderSingle.ReplaceAllString(s, "$1")

	// Blockquote markers.
	s = reBlockquote.ReplaceAllString(s, "")

	// Horizontal rules.
	s = reHRule.ReplaceAllString(s, "")

	// Table pipe characters.
	s = reTableRow.ReplaceAllString(s, " ")

	// HTML tags.
	s = reHTMLTag.ReplaceAllString(s, "")

	// Collapse excessive blank lines.
	s = reMultiBlank.ReplaceAllString(s, "\n\n")

	return strings.TrimSpace(s)
}
