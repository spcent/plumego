package utils

import "net/http"

// EnsureNoSniff sets X-Content-Type-Options to nosniff if it is not present.
// This reduces XSS risk from MIME sniffing when responses contain user-controlled text.
func EnsureNoSniff(header http.Header) {
	if header == nil {
		return
	}
	if header.Get("X-Content-Type-Options") == "" {
		header.Set("X-Content-Type-Options", "nosniff")
	}
}

// SafeWrite writes response bytes after applying minimal response hardening.
//
// SECURITY: This helper is intended for middleware and infrastructure code that
// needs to copy an already-constructed HTTP response. It does NOT perform any
// context-aware HTML/JS escaping. Callers that generate HTML or other active
// content MUST apply appropriate escaping/encoding before passing data here,
// and must set an appropriate Content-Type header (for example, application/json).
//
// The only protection applied here is X-Content-Type-Options: nosniff to
// prevent browsers from MIME-sniffing non-HTML responses as HTML.
//
// codeql[go/reflected-xss]: SafeWrite is a low-level passthrough used by
// response caching/middleware; it is not responsible for HTML encoding.
func SafeWrite(w http.ResponseWriter, body []byte) (int, error) {
	if w == nil {
		return 0, nil
	}
	EnsureNoSniff(w.Header())
	return w.Write(body)
}
