package utils

import (
	"strings"
	"testing"
)

func TestHTMLEscape(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "XSS script tag",
			input: "<script>alert('xss')</script>",
			want:  "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:  "HTML entities",
			input: `<div class="test">Hello & "World"</div>`,
			want:  "&lt;div class=&#34;test&#34;&gt;Hello &amp; &#34;World&#34;&lt;/div&gt;",
		},
		{
			name:  "Single quotes",
			input: "it's a test",
			want:  "it&#39;s a test",
		},
		{
			name:  "Ampersands",
			input: "Tom & Jerry",
			want:  "Tom &amp; Jerry",
		},
		{
			name:  "Empty string",
			input: "",
			want:  "",
		},
		{
			name:  "No special chars",
			input: "Hello World",
			want:  "Hello World",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := HTMLEscape(tc.input)
			if got != tc.want {
				t.Errorf("HTMLEscape(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestHTMLUnescapeString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Escaped script tag",
			input: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
			want:  "<script>alert('xss')</script>",
		},
		{
			name:  "HTML entities",
			input: "&lt;div class=&#34;test&#34;&gt;Hello &amp; &#34;World&#34;&lt;/div&gt;",
			want:  `<div class="test">Hello & "World"</div>`,
		},
		{
			name:  "Single quotes",
			input: "it&#39;s a test",
			want:  "it's a test",
		},
		{
			name:  "Empty string",
			input: "",
			want:  "",
		},
		{
			name:  "No entities",
			input: "Hello World",
			want:  "Hello World",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := HTMLUnescapeString(tc.input)
			if got != tc.want {
				t.Errorf("HTMLUnescapeString(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestJSEscape(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Script injection attempt",
			input: `'; alert('xss'); //`,
			want:  `\'; alert(\'xss\'); //`,
		},
		{
			name:  "Newlines",
			input: "line1\nline2",
			want:  `line1\u000Aline2`,
		},
		{
			name:  "Unicode",
			input: "Hello 世界",
			want:  `Hello 世界`,
		},
		{
			name:  "Empty string",
			input: "",
			want:  "",
		},
		{
			name:  "Safe string",
			input: "Hello World",
			want:  "Hello World",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := JSEscape(tc.input)
			if got != tc.want {
				t.Errorf("JSEscape(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestURLQueryEscape(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Spaces and special chars",
			input: "hello world&param=value",
			want:  "hello+world%26param%3Dvalue",
		},
		{
			name:  "URL with path",
			input: "/path?query=test",
			want:  "%2Fpath%3Fquery%3Dtest",
		},
		{
			name:  "Empty string",
			input: "",
			want:  "",
		},
		{
			name:  "Safe string",
			input: "HelloWorld123",
			want:  "HelloWorld123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := URLQueryEscape(tc.input)
			if got != tc.want {
				t.Errorf("URLQueryEscape(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSanitizeHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Script tag removal",
			input: "<div>Safe</div><script>alert('xss')</script><p>Text</p>",
			want:  "<div>Safe</div><p>Text</p>",
		},
		{
			name:  "Script tag with attributes",
			input: `<script type="text/javascript">alert('xss')</script>`,
			want:  "",
		},
		{
			name:  "Iframe removal",
			input: `<iframe src="evil.com"></iframe>`,
			want:  "",
		},
		{
			name:  "Object and embed removal",
			input: `<object data="evil.swf"></object><embed src="evil.swf">`,
			want:  "",
		},
		{
			name:  "Onclick handler removal",
			input: `<div onclick="alert('xss')">Click me</div>`,
			want:  "<div>Click me</div>",
		},
		{
			name:  "Onerror handler removal",
			input: `<img src="x" onerror="alert('xss')">`,
			want:  `<img src="x">`,
		},
		{
			name:  "Multiple event handlers",
			input: `<div onload="evil()" onmouseover="bad()">Text</div>`,
			want:  "<div>Text</div>",
		},
		{
			name:  "JavaScript protocol removal",
			input: `<a href="javascript:alert('xss')">Link</a>`,
			want:  `<a href="alert('xss')">Link</a>`,
		},
		{
			name:  "Data protocol removal",
			input: `<a href="data:text/html,<script>alert('xss')</script>">Link</a>`,
			want:  `<a href="text/html,">Link</a>`,
		},
		{
			name:  "Safe HTML unchanged",
			input: `<div class="test"><p>Hello <strong>World</strong></p></div>`,
			want:  `<div class="test"><p>Hello <strong>World</strong></p></div>`,
		},
		{
			name:  "Empty string",
			input: "",
			want:  "",
		},
		{
			name:  "Case insensitive tag removal",
			input: "<SCRIPT>alert('xss')</SCRIPT>",
			want:  "",
		},
		{
			name:  "Case insensitive attribute removal",
			input: `<div ONCLICK="alert('xss')">Click</div>`,
			want:  "<div>Click</div>",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := SanitizeHTML(tc.input)
			if got != tc.want {
				t.Errorf("SanitizeHTML(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSanitizeHTML_NestedScripts(t *testing.T) {
	input := "<div><script>outer<script>inner</script>outer</script></div>"
	result := SanitizeHTML(input)

	// Should remove all script tags
	if strings.Contains(strings.ToLower(result), "<script") {
		t.Errorf("SanitizeHTML did not remove all script tags: %q", result)
	}
}

func TestSanitizeHTML_Warning(t *testing.T) {
	// This test documents that SanitizeHTML is a basic sanitizer
	// and should not be relied upon as the sole XSS protection.
	// For production use, consider using a robust library like bluemonday.

	input := `<div>Test</div>`
	result := SanitizeHTML(input)

	// The result should be reasonably safe, but this is not a guarantee
	if result != input {
		t.Logf("Basic sanitization applied: %q -> %q", input, result)
	}
}
