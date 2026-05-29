package handler

import "net/http"

// Health returns a minimal liveness handler.
//
// guardus is a single-process app; readiness is satisfied as soon as the
// HTTP server accepts connections, so liveness == readiness here.
func Health() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}
