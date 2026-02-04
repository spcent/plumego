// Package ai provides AI agent gateway capabilities for plumego.
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
//		"github.com/spcent/plumego/ai/provider"
//		"github.com/spcent/plumego/ai/session"
//		"github.com/spcent/plumego/core"
//	)
//
//	func main() {
//		// Create LLM provider
//		claude := provider.NewClaudeProvider(apiKey)
//
//		// Create session manager
//		sessionMgr := session.NewManager(session.WithStorage(
//			session.NewMemoryStorage(),
//		))
//
//		// Setup application
//		app := core.New(
//			core.WithAddr(":8080"),
//			core.WithAIProvider(claude),
//			core.WithSessionManager(sessionMgr),
//		)
//
//		app.Boot()
//	}
package ai
