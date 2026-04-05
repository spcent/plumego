// Package ai provides AI gateway capabilities for Plumego extensions.
//
// This package includes:
//   - SSE (Server-Sent Events): Stream real-time responses to clients
//   - Provider: Unified interface for LLM providers (Claude, OpenAI, etc.)
//   - Session: Conversation management with context window control
//   - Tokenizer: Token counting and quota management
//   - Tool: Function calling framework for agent actions
//
// Example usage:
//
//	import (
//		"net/http"
//
//		"github.com/spcent/plumego/core"
//		"github.com/spcent/plumego/x/ai/provider"
//		"github.com/spcent/plumego/x/ai/session"
//	)
//
//	func main() {
//		model := provider.NewClaudeProvider(apiKey)
//		sessions := session.NewManager(session.NewMemoryStorage())
//
//		cfg := core.DefaultConfig()
//		cfg.Addr = ":8080"
//		app := core.New(cfg)
//		app.Post("/chat", func(w http.ResponseWriter, r *http.Request) {
//			_ = model
//			_ = sessions
//			w.WriteHeader(http.StatusNotImplemented)
//		})
//	}
package ai
