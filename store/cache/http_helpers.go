package cache

import "net/http"

func ensureNoSniff(header http.Header) {
	if header == nil {
		return
	}
	if header.Get("X-Content-Type-Options") == "" {
		header.Set("X-Content-Type-Options", "nosniff")
	}
}

func safeWrite(w http.ResponseWriter, body []byte) (int, error) {
	if w == nil {
		return 0, nil
	}
	ensureNoSniff(w.Header())
	return w.Write(body)
}
