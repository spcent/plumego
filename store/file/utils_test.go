package file

import (
	"strings"
	"testing"
)

func TestGenerateID(t *testing.T) {
	// Generate multiple IDs
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := generateID()

		// Check length (32 hex chars)
		if len(id) != 32 {
			t.Errorf("ID length = %d, want 32", len(id))
		}

		// Check uniqueness
		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		ids[id] = true

		// Check hex format
		for _, c := range id {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Errorf("ID contains non-hex character: %c", c)
			}
		}
	}
}

func TestIsPathSafe(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "safe relative path",
			path: "tenant/2026/02/05/file.txt",
			want: true,
		},
		{
			name: "safe single file",
			path: "file.txt",
			want: true,
		},
		{
			name: "path traversal with ..",
			path: "../../../etc/passwd",
			want: false,
		},
		{
			name: "path traversal in middle",
			path: "tenant/../../../etc/passwd",
			want: false,
		},
		{
			name: "absolute path",
			path: "/etc/passwd",
			want: false,
		},
		{
			name: "path with ./",
			path: "./tenant/file.txt",
			want: false, // filepath.Clean removes ./ but original != cleaned
		},
		{
			name: "empty path",
			path: "",
			want: false,
		},
		{
			name: "path with double slash",
			path: "tenant//file.txt",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPathSafe(tt.path)
			if got != tt.want {
				t.Errorf("isPathSafe(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestMimeToExt(t *testing.T) {
	tests := []struct {
		mimeType string
		want     string
	}{
		{"image/jpeg", ".jpg"},
		{"image/png", ".png"},
		{"image/gif", ".gif"},
		{"image/webp", ".webp"},
		{"image/svg+xml", ".svg"},
		{"application/pdf", ".pdf"},
		{"text/plain", ".txt"},
		{"text/html", ".html"},
		{"application/json", ".json"},
		{"application/xml", ".xml"},
		{"application/zip", ".zip"},
		{"unknown/type", ""}, // Unknown MIME type
		{"", ""},             // Empty MIME type
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			got := mimeToExt(tt.mimeType)
			if got != tt.want {
				t.Errorf("mimeToExt(%q) = %q, want %q", tt.mimeType, got, tt.want)
			}
		})
	}
}

func TestExtToMime(t *testing.T) {
	tests := []struct {
		ext  string
		want string
	}{
		{".jpg", "image/jpeg"},
		{"jpg", "image/jpeg"}, // Without leading dot
		{".jpeg", "image/jpeg"},
		{".JPG", "image/jpeg"}, // Case insensitive
		{".png", "image/png"},
		{".gif", "image/gif"},
		{".webp", "image/webp"},
		{".svg", "image/svg+xml"},
		{".pdf", "application/pdf"},
		{".txt", "text/plain"},
		{".html", "text/html"},
		{".json", "application/json"},
		{".xml", "application/xml"},
		{".zip", "application/zip"},
		{".unknown", "application/octet-stream"}, // Unknown extension
		{"", "application/octet-stream"},         // Empty extension
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			got := extToMime(tt.ext)
			if got != tt.want {
				t.Errorf("extToMime(%q) = %q, want %q", tt.ext, got, tt.want)
			}
		})
	}
}

// Benchmark ID generation
func BenchmarkGenerateID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateID()
	}
}

// Benchmark path safety checks
func BenchmarkIsPathSafe(b *testing.B) {
	testPaths := []string{
		"tenant/2026/02/05/file.txt",
		"../../../etc/passwd",
		"/etc/passwd",
		"safe/path/to/file.txt",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isPathSafe(testPaths[i%len(testPaths)])
	}
}

// Test edge cases
func TestIsPathSafeEdgeCases(t *testing.T) {
	// Very long path
	longPath := strings.Repeat("a/", 500) + "file.txt"
	if !isPathSafe(longPath) {
		t.Error("Long safe path should be allowed")
	}

	// Unicode characters
	unicodePath := "租户/文件.txt"
	if !isPathSafe(unicodePath) {
		t.Error("Unicode path should be allowed")
	}

	// Special characters (but safe)
	specialPath := "tenant_123/file-name.txt"
	if !isPathSafe(specialPath) {
		t.Error("Path with dash and underscore should be allowed")
	}
}
