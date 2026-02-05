package file

import (
	"crypto/rand"
	"encoding/hex"
	"path/filepath"
	"strings"
)

// generateID generates a random file ID (32 bytes, hex-encoded).
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// isPathSafe checks if a path is safe to use (prevents path traversal attacks).
func isPathSafe(path string) bool {
	// Reject paths containing ".."
	if strings.Contains(path, "..") {
		return false
	}

	// Reject absolute paths
	if filepath.IsAbs(path) {
		return false
	}

	// Ensure cleaned path matches original
	cleaned := filepath.Clean(path)
	if cleaned != path {
		return false
	}

	return true
}

// mimeToExt converts a MIME type to a file extension.
func mimeToExt(mimeType string) string {
	mimeMap := map[string]string{
		"image/jpeg":       ".jpg",
		"image/png":        ".png",
		"image/gif":        ".gif",
		"image/webp":       ".webp",
		"image/svg+xml":    ".svg",
		"application/pdf":  ".pdf",
		"text/plain":       ".txt",
		"text/html":        ".html",
		"text/css":         ".css",
		"text/javascript":  ".js",
		"application/json": ".json",
		"application/xml":  ".xml",
		"application/zip":  ".zip",
	}

	if ext, ok := mimeMap[mimeType]; ok {
		return ext
	}

	return ""
}

// extToMime converts a file extension to a MIME type.
func extToMime(ext string) string {
	ext = strings.ToLower(ext)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	extMap := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".pdf":  "application/pdf",
		".txt":  "text/plain",
		".html": "text/html",
		".css":  "text/css",
		".js":   "text/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".zip":  "application/zip",
	}

	if mime, ok := extMap[ext]; ok {
		return mime
	}

	return "application/octet-stream"
}
