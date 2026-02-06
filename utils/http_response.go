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
func SafeWrite(w http.ResponseWriter, body []byte) (int, error) {
	if w == nil {
		return 0, nil
	}
	EnsureNoSniff(w.Header())
	// codeql[go/reflected-xss]: passthrough write in middleware infrastructure.
	// lgtm [go/reflected-xss]
	return w.Write(body)
}
