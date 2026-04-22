package file

import (
	"strings"
	"testing"
)

// --- isPathSafe ---

func TestIsPathSafe_Safe(t *testing.T) {
	safePaths := []string{
		"uploads/image.jpg",
		"a/b/c.txt",
		"file.txt",
	}
	for _, p := range safePaths {
		if !isPathSafe(p) {
			t.Errorf("isPathSafe(%q) = false, want true", p)
		}
	}
}

func TestIsPathSafe_Traversal(t *testing.T) {
	unsafe := []string{
		"../etc/passwd",
		"a/../../b",
		"uploads/../secret",
	}
	for _, p := range unsafe {
		if isPathSafe(p) {
			t.Errorf("isPathSafe(%q) = true, want false (path traversal)", p)
		}
	}
}

func TestIsPathSafe_AbsolutePath(t *testing.T) {
	if isPathSafe("/etc/passwd") {
		t.Error("isPathSafe(/etc/passwd) = true, want false (absolute path)")
	}
}

// --- mimeToExt ---

func TestMimeToExt_KnownTypes(t *testing.T) {
	cases := []struct {
		mime string
		ext  string
	}{
		{"image/jpeg", ".jpg"},
		{"image/png", ".png"},
		{"image/gif", ".gif"},
		{"image/webp", ".webp"},
		{"image/svg+xml", ".svg"},
		{"application/pdf", ".pdf"},
		{"text/plain", ".txt"},
		{"text/html", ".html"},
		{"text/css", ".css"},
		{"text/javascript", ".js"},
		{"application/json", ".json"},
		{"application/xml", ".xml"},
		{"application/zip", ".zip"},
	}
	for _, tc := range cases {
		if got := mimeToExt(tc.mime); got != tc.ext {
			t.Errorf("mimeToExt(%q) = %q, want %q", tc.mime, got, tc.ext)
		}
	}
}

func TestMimeToExt_Unknown(t *testing.T) {
	if got := mimeToExt("application/octet-stream"); got != "" {
		t.Errorf("mimeToExt(unknown) = %q, want empty", got)
	}
}

// --- extToMime ---

func TestExtToMime_KnownExtensions(t *testing.T) {
	cases := []struct {
		ext  string
		mime string
	}{
		{".jpg", "image/jpeg"},
		{".jpeg", "image/jpeg"},
		{".png", "image/png"},
		{".gif", "image/gif"},
		{".webp", "image/webp"},
		{".svg", "image/svg+xml"},
		{".pdf", "application/pdf"},
		{".txt", "text/plain"},
		{".html", "text/html"},
		{".css", "text/css"},
		{".js", "text/javascript"},
		{".json", "application/json"},
		{".xml", "application/xml"},
		{".zip", "application/zip"},
	}
	for _, tc := range cases {
		if got := extToMime(tc.ext); got != tc.mime {
			t.Errorf("extToMime(%q) = %q, want %q", tc.ext, got, tc.mime)
		}
	}
}

func TestExtToMime_WithoutLeadingDot(t *testing.T) {
	if got := extToMime("jpg"); got != "image/jpeg" {
		t.Errorf("extToMime(jpg) = %q, want image/jpeg", got)
	}
}

func TestExtToMime_CaseInsensitive(t *testing.T) {
	if got := extToMime(".JPG"); got != "image/jpeg" {
		t.Errorf("extToMime(.JPG) = %q, want image/jpeg", got)
	}
}

func TestExtToMime_Unknown(t *testing.T) {
	if got := extToMime(".xyz"); got != "application/octet-stream" {
		t.Errorf("extToMime(.xyz) = %q, want application/octet-stream", got)
	}
}

// round-trip: mimeToExt → extToMime for image types
func TestMimeExtRoundTrip(t *testing.T) {
	mimes := []string{
		"image/jpeg", "image/png", "image/gif", "application/pdf",
		"text/plain", "application/json", "application/zip",
	}
	for _, mime := range mimes {
		ext := mimeToExt(mime)
		if ext == "" {
			t.Errorf("mimeToExt(%q) returned empty", mime)
			continue
		}
		if got := extToMime(ext); !strings.HasPrefix(got, strings.Split(mime, "/")[0]) {
			t.Errorf("extToMime(mimeToExt(%q)) = %q, expected mime family match", mime, got)
		}
	}
}
