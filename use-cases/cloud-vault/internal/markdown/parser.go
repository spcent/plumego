package markdown

import (
	"bufio"
	"encoding/json"
	"strings"
	"unicode"
)

// Heading represents a single ATX heading extracted from Markdown.
type Heading struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
}

// Meta holds extracted metadata from a Markdown document.
type Meta struct {
	Title          string
	WordCount      int
	LineCount      int
	Summary        string    // first ~300 chars of non-heading plain text
	HeadingText    string    // first heading text (for quick display)
	Headings       []Heading // all headings in document order
	CodeLanguages  []string  // deduplicated list of fenced code block languages
	CodeBlockCount int
	LinkCount      int
	ImageCount     int
}

// HeadingsJSON returns the Headings slice serialised as a JSON string.
// Returns "" when there are no headings.
func (m *Meta) HeadingsJSON() string {
	if len(m.Headings) == 0 {
		return ""
	}
	b, _ := json.Marshal(m.Headings)
	return string(b)
}

// CodeLanguagesJSON returns CodeLanguages serialised as a JSON string.
// Returns "" when there are none.
func (m *Meta) CodeLanguagesJSON() string {
	if len(m.CodeLanguages) == 0 {
		return ""
	}
	b, _ := json.Marshal(m.CodeLanguages)
	return string(b)
}

// Parse extracts metadata from raw Markdown text.
func Parse(content string) Meta {
	var (
		headings      []Heading
		codeLanguages []string
		seenLangs     = make(map[string]bool)
		codeBlocks    int
		links         int
		images        int
		summary       strings.Builder
		inCodeBlock   bool
	)

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()

		// Fenced code block toggle.
		if strings.HasPrefix(line, "```") {
			if !inCodeBlock {
				inCodeBlock = true
				codeBlocks++
				lang := strings.TrimSpace(line[3:])
				// Strip optional trailing ``` from single-line blocks.
				if idx := strings.Index(lang, "```"); idx >= 0 {
					lang = lang[:idx]
				}
				lang = strings.TrimSpace(lang)
				if lang != "" && !seenLangs[lang] {
					codeLanguages = append(codeLanguages, lang)
					seenLangs[lang] = true
				}
			} else {
				inCodeBlock = false
			}
			continue
		}
		if inCodeBlock {
			continue
		}

		// ATX headings (#, ##, …, ######).
		if strings.HasPrefix(line, "#") {
			level := 0
			for _, ch := range line {
				if ch == '#' {
					level++
				} else {
					break
				}
			}
			if level >= 1 && level <= 6 && len(line) > level && line[level] == ' ' {
				text := strings.TrimSpace(line[level+1:])
				if text != "" {
					headings = append(headings, Heading{Level: level, Text: text})
				}
			}
			continue
		}

		// Count inline links [text](url) and images ![alt](url).
		images += countImagesInLine(line)
		links += countLinksInLine(line)

		// Build summary from plain-text lines (skip blank lines).
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && summary.Len() < 300 {
			if summary.Len() > 0 {
				summary.WriteByte(' ')
			}
			remaining := 300 - summary.Len()
			if len(trimmed) > remaining {
				summary.WriteString(trimmed[:remaining])
			} else {
				summary.WriteString(trimmed)
			}
		}
	}

	// Derive title: prefer level-1 heading, fall back to first heading.
	title := ""
	headingText := ""
	if len(headings) > 0 {
		headingText = headings[0].Text
		for _, h := range headings {
			if h.Level == 1 {
				title = h.Text
				break
			}
		}
		if title == "" {
			title = headingText
		}
	}

	return Meta{
		Title:          title,
		WordCount:      countWords(content),
		LineCount:      countLines(content),
		Summary:        summary.String(),
		HeadingText:    headingText,
		Headings:       headings,
		CodeLanguages:  codeLanguages,
		CodeBlockCount: codeBlocks,
		LinkCount:      links,
		ImageCount:     images,
	}
}

// ExtractTitle returns the first ATX heading text, or empty string.
func ExtractTitle(content string) string {
	return Parse(content).Title
}

func countLinksInLine(line string) int {
	count := 0
	i := 0
	for i < len(line) {
		// Image takes priority: skip ![ sequences.
		if i+1 < len(line) && line[i] == '!' && line[i+1] == '[' {
			// Skip both '!' and '[' so the '[' is not counted as a link.
			i += 2
			continue
		}
		if line[i] == '[' {
			if j := strings.Index(line[i:], "]("); j >= 0 {
				count++
				i += j + 2
				continue
			}
		}
		i++
	}
	return count
}

func countImagesInLine(line string) int {
	count := 0
	i := 0
	for i+1 < len(line) {
		if line[i] == '!' && line[i+1] == '[' {
			if j := strings.Index(line[i:], "]("); j >= 0 {
				count++
				i += j + 2
				continue
			}
		}
		i++
	}
	return count
}

func countWords(content string) int {
	count := 0
	inWord := false
	for _, r := range content {
		if unicode.IsSpace(r) {
			inWord = false
		} else {
			if !inWord {
				count++
				inWord = true
			}
		}
	}
	return count
}

func countLines(content string) int {
	if content == "" {
		return 0
	}
	lines := strings.Count(content, "\n")
	if !strings.HasSuffix(content, "\n") {
		lines++
	}
	return lines
}
