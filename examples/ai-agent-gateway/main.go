package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/spcent/plumego/ai/provider"
	"github.com/spcent/plumego/ai/session"
	"github.com/spcent/plumego/ai/sse"
	"github.com/spcent/plumego/ai/tokenizer"
	"github.com/spcent/plumego/ai/tool"
	"github.com/spcent/plumego/core"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("CLAUDE_API_KEY")
	if apiKey == "" {
		log.Println("Warning: CLAUDE_API_KEY not set, using mock provider")
	}

	// Create providers
	providerMgr := provider.NewManager()
	if apiKey != "" {
		claudeProvider := provider.NewClaudeProvider(apiKey)
		providerMgr.Register(claudeProvider)
	}

	// Create session manager
	sessionMgr := session.NewManager(
		session.NewMemoryStorage(),
		session.WithTokenizer(tokenizer.NewClaudeTokenizer("claude-3")),
	)

	// Create tool registry
	toolRegistry := tool.NewRegistry()
	toolRegistry.Register(tool.NewEchoTool())
	toolRegistry.Register(tool.NewCalculatorTool())
	toolRegistry.Register(tool.NewTimestampTool())

	// Create application
	app := core.New(
		core.WithAddr(":8080"),
		core.WithDebug(),
		core.WithLogging(),
		core.WithRecovery(),
	)

	// Routes
	app.Get("/", indexHandler)
	app.Post("/api/sessions", createSessionHandler(sessionMgr))
	app.Post("/api/sessions/:id/messages", sendMessageHandler(sessionMgr, providerMgr, toolRegistry))
	app.Get("/api/sessions/:id/stream", streamHandler(sessionMgr, providerMgr, toolRegistry))
	app.Get("/api/tools", listToolsHandler(toolRegistry))

	log.Println("AI Agent Gateway starting on http://localhost:8080")
	log.Println("Endpoints:")
	log.Println("  POST /api/sessions - Create a new session")
	log.Println("  POST /api/sessions/:id/messages - Send a message")
	log.Println("  GET  /api/sessions/:id/stream - Stream responses (SSE)")
	log.Println("  GET  /api/tools - List available tools")

	if err := app.Boot(); err != nil {
		log.Fatal(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>AI Agent Gateway</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        .endpoint { background: #f5f5f5; padding: 10px; margin: 10px 0; border-left: 4px solid #4CAF50; }
        code { background: #eee; padding: 2px 5px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>🤖 AI Agent Gateway (Phase 1)</h1>
    <p>A lightweight AI agent gateway built with Plumego</p>

    <h2>Features</h2>
    <ul>
        <li>✅ SSE (Server-Sent Events) for streaming</li>
        <li>✅ LLM Provider abstraction (Claude, OpenAI)</li>
        <li>✅ Session management</li>
        <li>✅ Token counting</li>
        <li>✅ Tool calling framework</li>
    </ul>

    <h2>API Endpoints</h2>

    <div class="endpoint">
        <strong>POST /api/sessions</strong>
        <p>Create a new conversation session</p>
        <code>curl -X POST http://localhost:8080/api/sessions -H "Content-Type: application/json" -d '{"model": "claude-3-opus"}'</code>
    </div>

    <div class="endpoint">
        <strong>POST /api/sessions/:id/messages</strong>
        <p>Send a message to the agent</p>
        <code>curl -X POST http://localhost:8080/api/sessions/{id}/messages -H "Content-Type: application/json" -d '{"message": "Calculate 5 + 3"}'</code>
    </div>

    <div class="endpoint">
        <strong>GET /api/sessions/:id/stream</strong>
        <p>Stream responses via SSE</p>
        <code>curl -N http://localhost:8080/api/sessions/{id}/stream</code>
    </div>

    <div class="endpoint">
        <strong>GET /api/tools</strong>
        <p>List available tools</p>
        <code>curl http://localhost:8080/api/tools</code>
    </div>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func createSessionHandler(sessionMgr *session.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Model string `json:"model"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Model == "" {
			req.Model = "claude-3-opus-20240229"
		}

		sess, err := sessionMgr.Create(r.Context(), session.CreateOptions{
			Model: req.Model,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"session_id": sess.ID,
			"model":      sess.Model,
			"created_at": sess.CreatedAt,
		})
	}
}

func sendMessageHandler(sessionMgr *session.Manager, providerMgr *provider.Manager, toolRegistry *tool.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Query().Get(":id")
		if sessionID == "" {
			http.Error(w, "session_id required", http.StatusBadRequest)
			return
		}

		var req struct {
			Message string `json:"message"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Append user message
		userMsg := provider.NewTextMessage(provider.RoleUser, req.Message)
		if err := sessionMgr.AppendMessage(r.Context(), sessionID, userMsg); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		// Mock response (in real implementation, call provider)
		responseText := fmt.Sprintf("Received: %s (Mock response - set CLAUDE_API_KEY to use real LLM)", req.Message)

		// Append assistant message
		assistantMsg := provider.NewTextMessage(provider.RoleAssistant, responseText)
		if err := sessionMgr.AppendMessage(r.Context(), sessionID, assistantMsg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"response": responseText,
		})
	}
}

func streamHandler(sessionMgr *session.Manager, providerMgr *provider.Manager, toolRegistry *tool.Registry) http.HandlerFunc {
	return sse.Handle(func(s *sse.Stream) error {
		// Get session ID from query
		// Note: In a real implementation, you'd parse the URL properly
		// This is simplified for demonstration

		// Send welcome message
		if err := s.SendJSON("message", `{"type": "welcome", "text": "Connected to AI Agent Gateway"}`); err != nil {
			return err
		}

		// Simulate streaming response
		chunks := []string{
			"Hello! ",
			"I'm your ",
			"AI assistant. ",
			"How can ",
			"I help ",
			"you today?",
		}

		for i, chunk := range chunks {
			if err := s.SendJSON("chunk", fmt.Sprintf(`{"index": %d, "text": "%s"}`, i, chunk)); err != nil {
				return err
			}
		}

		// Send completion
		return s.SendJSON("done", `{"type": "complete", "message": "Stream finished"}`)
	})
}

func listToolsHandler(toolRegistry *tool.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tools := toolRegistry.ToProviderTools(context.Background())

		response := make([]map[string]any, len(tools))
		for i, t := range tools {
			response[i] = map[string]any{
				"name":        t.Function.Name,
				"description": t.Function.Description,
				"parameters":  t.Function.Parameters,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"tools": response,
			"count": len(tools),
		})
	}
}
