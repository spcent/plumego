package search

import (
	"strings"
	"testing"
)

func TestClean_empty(t *testing.T) {
	if got := Clean(""); got != "" {
		t.Errorf("empty: want '', got %q", got)
	}
}

func TestClean_headings(t *testing.T) {
	input := "# Title\n## Subtitle\n\nBody text.\n"
	got := Clean(input)
	if strings.Contains(got, "#") {
		t.Errorf("heading markers should be removed: %q", got)
	}
	if !strings.Contains(got, "Title") {
		t.Errorf("heading text should be preserved: %q", got)
	}
	if !strings.Contains(got, "Subtitle") {
		t.Errorf("subheading text should be preserved: %q", got)
	}
}

func TestClean_codeBlock(t *testing.T) {
	input := "Before.\n\n```go\nfmt.Println()\n```\n\nAfter.\n"
	got := Clean(input)
	if !strings.Contains(got, "fmt.Println()") {
		t.Errorf("code block content should be preserved: %q", got)
	}
	if strings.Contains(got, "```") {
		t.Errorf("code fence markers should be removed: %q", got)
	}
}

func TestClean_inlineCode(t *testing.T) {
	input := "Use `fmt.Println` in Go.\n"
	got := Clean(input)
	if strings.Contains(got, "`") {
		t.Errorf("backticks should be removed: %q", got)
	}
	if !strings.Contains(got, "fmt.Println") {
		t.Errorf("inline code content should be preserved: %q", got)
	}
}

func TestClean_link(t *testing.T) {
	input := "See [the docs](https://example.com) for details.\n"
	got := Clean(input)
	if strings.Contains(got, "https://example.com") {
		t.Errorf("link URL should be removed: %q", got)
	}
	if !strings.Contains(got, "the docs") {
		t.Errorf("link text should be preserved: %q", got)
	}
}

func TestClean_image(t *testing.T) {
	input := "See ![a cat](cat.png) here.\n"
	got := Clean(input)
	if strings.Contains(got, "cat.png") {
		t.Errorf("image URL should be removed: %q", got)
	}
	if !strings.Contains(got, "a cat") {
		t.Errorf("image alt text should be preserved: %q", got)
	}
}

func TestClean_blockquote(t *testing.T) {
	input := "> This is a quote.\n"
	got := Clean(input)
	if strings.HasPrefix(strings.TrimSpace(got), ">") {
		t.Errorf("blockquote marker should be removed: %q", got)
	}
	if !strings.Contains(got, "This is a quote.") {
		t.Errorf("blockquote content should be preserved: %q", got)
	}
}

func TestClean_htmlTags(t *testing.T) {
	input := "Hello <strong>world</strong> and <br/> more.\n"
	got := Clean(input)
	if strings.Contains(got, "<strong>") || strings.Contains(got, "</strong>") || strings.Contains(got, "<br/>") {
		t.Errorf("HTML tags should be removed: %q", got)
	}
	if !strings.Contains(got, "world") {
		t.Errorf("HTML content should be preserved: %q", got)
	}
}

func TestClean_tables(t *testing.T) {
	input := "| Col A | Col B |\n|---|---|\n| val1 | val2 |\n"
	got := Clean(input)
	if !strings.Contains(got, "Col A") || !strings.Contains(got, "val1") {
		t.Errorf("table text should be preserved: %q", got)
	}
}

func TestClean_horizontalRule(t *testing.T) {
	input := "Above\n\n---\n\nBelow\n"
	got := Clean(input)
	if strings.Contains(got, "---") {
		t.Errorf("horizontal rule should be removed: %q", got)
	}
	if !strings.Contains(got, "Above") || !strings.Contains(got, "Below") {
		t.Errorf("surrounding text should be preserved: %q", got)
	}
}

func TestSanitizeFTSQuery_basic(t *testing.T) {
	got := sanitizeFTSQuery("hello world")
	if got != `"hello" "world"` {
		t.Errorf("want '\"hello\" \"world\"', got %q", got)
	}
}

func TestSanitizeFTSQuery_empty(t *testing.T) {
	if got := sanitizeFTSQuery("   "); got != "" {
		t.Errorf("empty input: want '', got %q", got)
	}
}

func TestSanitizeFTSQuery_escapeQuotes(t *testing.T) {
	got := sanitizeFTSQuery(`say "hello"`)
	if !strings.Contains(got, `""hello""`) {
		t.Errorf("inner quotes should be escaped: %q", got)
	}
}

func TestSanitizeFTSQuery_single(t *testing.T) {
	got := sanitizeFTSQuery("golang")
	if got != `"golang"` {
		t.Errorf("single token: want '\"golang\"', got %q", got)
	}
}
