package utils

import (
	"html"
	"html/template"
	"strings"
)

// HTMLEscape escapes special HTML characters to prevent XSS attacks.
//
// This function should be used when embedding user-provided content in HTML responses.
// It escapes the following characters:
//   - < becomes &lt;
//   - > becomes &gt;
//   - & becomes &amp;
//   - " becomes &#34;
//   - ' becomes &#39;
//
// Example:
//
//	userInput := "<script>alert('xss')</script>"
//	safe := utils.HTMLEscape(userInput)
//	fmt.Fprintf(w, "<div>%s</div>", safe)
func HTMLEscape(s string) string {
	return html.EscapeString(s)
}

// HTMLUnescapeString is the inverse of HTMLEscape.
// It converts HTML entities back to their original characters.
func HTMLUnescapeString(s string) string {
	return html.UnescapeString(s)
}

// JSEscape escapes a string for safe embedding in JavaScript contexts.
//
// This function escapes characters that could break out of JavaScript string literals
// or cause script injection when embedding data in JavaScript code.
//
// Example:
//
//	userInput := `'; alert('xss'); //`
//	safe := utils.JSEscape(userInput)
//	fmt.Fprintf(w, "<script>var data = '%s';</script>", safe)
func JSEscape(s string) string {
	return template.JSEscapeString(s)
}

// URLQueryEscape escapes a string for safe use in URL query parameters.
//
// Example:
//
//	userInput := "hello world&param=value"
//	safe := utils.URLQueryEscape(userInput)
//	fmt.Fprintf(w, "<a href='/search?q=%s'>Link</a>", safe)
func URLQueryEscape(s string) string {
	return template.URLQueryEscaper(s)
}

// SanitizeHTML removes potentially dangerous HTML tags and attributes.
//
// WARNING: This is a basic sanitizer and should NOT be relied upon as the sole
// XSS protection mechanism. Use HTMLEscape for user content when possible.
//
// This function removes:
//   - Script tags
//   - Event handlers (onclick, onerror, etc.)
//   - javascript: protocol
//   - data: protocol
//
// For production use, consider using a robust library like bluemonday.
func SanitizeHTML(s string) string {
	// Remove script tags
	s = removeTag(s, "script")
	s = removeTag(s, "iframe")
	s = removeTag(s, "object")
	s = removeTag(s, "embed")

	// Remove dangerous attributes
	dangerousAttrs := []string{
		"onclick", "onerror", "onload", "onmouseover", "onfocus",
		"onblur", "onchange", "onsubmit", "onkeydown", "onkeyup",
		"onmouseenter", "onmouseleave",
	}

	for _, attr := range dangerousAttrs {
		s = removeAttribute(s, attr)
	}

	// Remove javascript: and data: protocols
	s = strings.ReplaceAll(s, "javascript:", "")
	s = strings.ReplaceAll(s, "data:", "")

	return s
}

func removeTag(s, tag string) string {
	// Simple tag removal - not foolproof
	lowerS := strings.ToLower(s)
	startTag := "<" + tag
	endTag := "</" + tag + ">"

	for {
		start := strings.Index(lowerS, startTag)
		if start == -1 {
			break
		}

		// Find the closing >
		closingBracket := strings.Index(lowerS[start:], ">")
		if closingBracket == -1 {
			break
		}

		// Find the closing tag
		end := strings.Index(lowerS[start:], endTag)
		if end == -1 {
			// Self-closing or no closing tag
			s = s[:start] + s[start+closingBracket+1:]
			lowerS = strings.ToLower(s)
		} else {
			// Remove from start to end of closing tag
			endPos := start + end + len(endTag)
			s = s[:start] + s[endPos:]
			lowerS = strings.ToLower(s)
		}
	}

	return s
}

func removeAttribute(s, attr string) string {
	lowerS := strings.ToLower(s)
	attrPattern := " " + attr + "="

	for {
		idx := strings.Index(lowerS, attrPattern)
		if idx == -1 {
			break
		}

		// Find the quote type
		quoteStart := idx + len(attrPattern)
		if quoteStart >= len(s) {
			break
		}

		quote := s[quoteStart]
		if quote != '"' && quote != '\'' {
			// No quote, find next space
			spaceIdx := strings.Index(s[quoteStart:], " ")
			if spaceIdx == -1 {
				s = s[:idx]
				lowerS = strings.ToLower(s)
			} else {
				s = s[:idx] + s[quoteStart+spaceIdx:]
				lowerS = strings.ToLower(s)
			}
			continue
		}

		// Find closing quote
		quoteEnd := strings.Index(s[quoteStart+1:], string(quote))
		if quoteEnd == -1 {
			s = s[:idx]
			lowerS = strings.ToLower(s)
		} else {
			endPos := quoteStart + 1 + quoteEnd + 1
			s = s[:idx] + s[endPos:]
			lowerS = strings.ToLower(s)
		}
	}

	return s
}
