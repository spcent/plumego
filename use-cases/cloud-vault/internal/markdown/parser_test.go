package markdown

import (
	"encoding/json"
	"testing"
)

func TestParse_empty(t *testing.T) {
	m := Parse("")
	if m.Title != "" || m.WordCount != 0 || m.LineCount != 0 {
		t.Errorf("empty input: got %+v", m)
	}
}

func TestParse_titleFromH1(t *testing.T) {
	m := Parse("# Hello World\n\nSome text.\n")
	if m.Title != "Hello World" {
		t.Errorf("title: want 'Hello World', got %q", m.Title)
	}
	if m.HeadingText != "Hello World" {
		t.Errorf("heading_text: want 'Hello World', got %q", m.HeadingText)
	}
	if len(m.Headings) != 1 || m.Headings[0].Level != 1 {
		t.Errorf("headings: %+v", m.Headings)
	}
}

func TestParse_titleFallsBackToFirstHeading(t *testing.T) {
	m := Parse("## Section\n\ntext\n")
	if m.Title != "Section" {
		t.Errorf("title: want 'Section', got %q", m.Title)
	}
}

func TestParse_multipleHeadings(t *testing.T) {
	input := "## Intro\n\n### Deep\n\n# Main\n"
	m := Parse(input)
	if m.Title != "Main" {
		t.Errorf("title: want 'Main', got %q", m.Title)
	}
	if len(m.Headings) != 3 {
		t.Errorf("expected 3 headings, got %d", len(m.Headings))
	}
	// HeadingText is the first heading encountered, not necessarily H1.
	if m.HeadingText != "Intro" {
		t.Errorf("heading_text: want 'Intro', got %q", m.HeadingText)
	}
}

func TestParse_codeBlocks(t *testing.T) {
	input := "# T\n\n```go\nfmt.Println()\n```\n\n```python\nprint()\n```\n\n```go\nduplicate\n```\n"
	m := Parse(input)
	if m.CodeBlockCount != 3 {
		t.Errorf("code blocks: want 3, got %d", m.CodeBlockCount)
	}
	// go should appear only once (deduplicated).
	if len(m.CodeLanguages) != 2 {
		t.Errorf("code languages: want 2, got %v", m.CodeLanguages)
	}
}

func TestParse_linksAndImages(t *testing.T) {
	input := "See [link](http://example.com) and ![img](photo.png) and [another](url).\n"
	m := Parse(input)
	if m.LinkCount != 2 {
		t.Errorf("links: want 2, got %d", m.LinkCount)
	}
	if m.ImageCount != 1 {
		t.Errorf("images: want 1, got %d", m.ImageCount)
	}
}

func TestParse_summaryTruncation(t *testing.T) {
	long := ""
	for i := 0; i < 30; i++ {
		long += "word "
	}
	m := Parse(long)
	if len(m.Summary) > 300 {
		t.Errorf("summary exceeds 300 chars: len=%d", len(m.Summary))
	}
}

func TestParse_codeBlockContentIgnored(t *testing.T) {
	// Links inside code blocks must not be counted.
	input := "```\n[link](url)\n![img](img.png)\n```\n"
	m := Parse(input)
	if m.LinkCount != 0 || m.ImageCount != 0 {
		t.Errorf("links/images inside code block should not be counted: links=%d images=%d", m.LinkCount, m.ImageCount)
	}
}

func TestParse_wordAndLineCount(t *testing.T) {
	input := "hello world\nfoo bar baz\n"
	m := Parse(input)
	if m.WordCount != 5 {
		t.Errorf("word count: want 5, got %d", m.WordCount)
	}
	if m.LineCount != 2 {
		t.Errorf("line count: want 2, got %d", m.LineCount)
	}
}

func TestHeadingsJSON(t *testing.T) {
	m := Meta{Headings: []Heading{{Level: 1, Text: "Hi"}, {Level: 2, Text: "Sub"}}}
	raw := m.HeadingsJSON()
	var out []Heading
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(out) != 2 || out[0].Text != "Hi" {
		t.Errorf("unexpected headings: %+v", out)
	}
}

func TestHeadingsJSON_empty(t *testing.T) {
	m := Meta{}
	if m.HeadingsJSON() != "" {
		t.Errorf("empty headings should return ''")
	}
}

func TestCodeLanguagesJSON_empty(t *testing.T) {
	m := Meta{}
	if m.CodeLanguagesJSON() != "" {
		t.Errorf("empty langs should return ''")
	}
}

func TestExtractTitle(t *testing.T) {
	if got := ExtractTitle("# Vault\n"); got != "Vault" {
		t.Errorf("want 'Vault', got %q", got)
	}
	if got := ExtractTitle("no heading"); got != "" {
		t.Errorf("want '', got %q", got)
	}
}
