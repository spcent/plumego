package docsite

import (
	"html"
	"html/template"
	"strings"
)

// markdownToHTML converts a subset of Markdown to HTML.
// Supported: headings, paragraphs, bold, inline code, code blocks, unordered
// lists, and pipe-delimited tables. This is intentionally minimal — it avoids
// an external dependency at the cost of full CommonMark coverage.
func markdownToHTML(md string) template.HTML {
	lines := strings.Split(md, "\n")
	var b strings.Builder
	inList := false
	inCode := false
	inTable := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		switch {
		case strings.HasPrefix(line, "```"):
			if inTable {
				b.WriteString("</tbody></table>")
				inTable = false
			}
			if inCode {
				b.WriteString("</code></pre>")
				inCode = false
			} else {
				if inList {
					b.WriteString("</ul>")
					inList = false
				}
				b.WriteString("<pre><code>")
				inCode = true
			}
			continue

		case inCode:
			b.WriteString(html.EscapeString(line))
			b.WriteString("\n")
			continue

		case inTable:
			if hasPipeRow(line) {
				b.WriteString("<tr>")
				for _, cell := range parseTableRow(line) {
					b.WriteString("<td>" + renderInline(cell) + "</td>")
				}
				b.WriteString("</tr>")
				continue
			}
			b.WriteString("</tbody></table>")
			inTable = false
		}

		if isTableStart(lines, i) {
			if inList {
				b.WriteString("</ul>")
				inList = false
			}
			b.WriteString("<table><thead><tr>")
			for _, cell := range parseTableRow(line) {
				b.WriteString("<th>" + renderInline(cell) + "</th>")
			}
			b.WriteString("</tr></thead><tbody>")
			inTable = true
			i++ // skip separator row
			continue
		}

		switch {
		case strings.HasPrefix(line, "- "):
			if !inList {
				b.WriteString("<ul>")
				inList = true
			}
			b.WriteString("<li>" + renderInline(strings.TrimSpace(line[2:])) + "</li>")
			continue
		default:
			if inList {
				b.WriteString("</ul>")
				inList = false
			}
		}

		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "#### "):
			b.WriteString("<h4>" + renderInline(strings.TrimSpace(trimmed[5:])) + "</h4>")
		case strings.HasPrefix(trimmed, "### "):
			b.WriteString("<h3>" + renderInline(strings.TrimSpace(trimmed[4:])) + "</h3>")
		case strings.HasPrefix(trimmed, "## "):
			b.WriteString("<h2>" + renderInline(strings.TrimSpace(trimmed[3:])) + "</h2>")
		case strings.HasPrefix(trimmed, "# "):
			b.WriteString("<h1>" + renderInline(strings.TrimSpace(trimmed[2:])) + "</h1>")
		case trimmed == "":
			b.WriteString("<p></p>")
		default:
			b.WriteString("<p>" + renderInline(line) + "</p>")
		}
	}

	if inList {
		b.WriteString("</ul>")
	}
	if inTable {
		b.WriteString("</tbody></table>")
	}
	if inCode {
		b.WriteString("</code></pre>")
	}
	return template.HTML(b.String())
}

func renderInline(text string) string {
	escaped := html.EscapeString(text)
	parts := strings.Split(escaped, "`")
	for i, part := range parts {
		if i%2 == 1 {
			parts[i] = "<code>" + part + "</code>"
		} else {
			parts[i] = renderBold(part)
		}
	}
	return strings.Join(parts, "")
}

func renderBold(text string) string {
	var b strings.Builder
	for {
		start := strings.Index(text, "**")
		if start == -1 {
			b.WriteString(text)
			break
		}
		end := strings.Index(text[start+2:], "**")
		if end == -1 {
			b.WriteString(text)
			break
		}
		b.WriteString(text[:start])
		b.WriteString("<strong>")
		b.WriteString(text[start+2 : start+2+end])
		b.WriteString("</strong>")
		text = text[start+2+end+2:]
	}
	return b.String()
}

func isTableStart(lines []string, index int) bool {
	if index+1 >= len(lines) {
		return false
	}
	return hasPipeRow(lines[index]) && isTableSeparator(lines[index+1])
}

func hasPipeRow(line string) bool {
	return strings.Contains(line, "|")
}

func isTableSeparator(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	parts := strings.Split(trimmed, "|")
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return false
		}
		for _, r := range part {
			if r != '-' && r != ':' {
				return false
			}
		}
	}
	return true
}

func parseTableRow(line string) []string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "|")
	trimmed = strings.TrimSuffix(trimmed, "|")
	parts := strings.Split(trimmed, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
