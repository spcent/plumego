// Example: non-canonical
//
// This demo wires the stable-tier x/ai subpackages with offline behavior:
// provider, session, streaming, and tool.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware/recovery"
	"github.com/spcent/plumego/middleware/requestid"
	"github.com/spcent/plumego/x/ai/provider"
	"github.com/spcent/plumego/x/ai/session"
	"github.com/spcent/plumego/x/ai/streaming"
	"github.com/spcent/plumego/x/ai/tool"
)

type aiApp struct {
	provider *provider.MockProvider
	sessions *session.Manager
	streams  *streaming.StreamManager
	tools    *tool.Registry
}

type chatRequest struct {
	Message string `json:"message"`
}

type chatResponse struct {
	SessionID   string         `json:"session_id"`
	Reply       string         `json:"reply"`
	ToolOutput  map[string]any `json:"tool_output"`
	StreamCount int            `json:"stream_count"`
}

func main() {
	cfg := core.DefaultConfig()
	cfg.Addr = envString("APP_ADDR", ":8086")

	app := core.New(cfg, core.AppDependencies{Logger: plumelog.NewLogger()})
	if err := app.Use(requestid.Middleware(), recovery.Recovery(app.Logger())); err != nil {
		log.Fatalf("register middleware: %v", err)
	}

	ai := newAIApp()
	if err := app.Post("/api/chat", http.HandlerFunc(ai.chat)); err != nil {
		log.Fatalf("register chat route: %v", err)
	}
	if err := app.Get("/api/ai/status", http.HandlerFunc(ai.status)); err != nil {
		log.Fatalf("register ai status route: %v", err)
	}

	if err := serve(app, cfg); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func newAIApp() *aiApp {
	registry := tool.NewRegistry(tool.WithPolicy(tool.NewAllowListPolicy([]string{"echo"})))
	_ = registry.Register(tool.NewEchoTool())
	_ = registry.Register(tool.NewCalculatorTool())

	return &aiApp{
		provider: provider.NewMockProvider("offline-mock"),
		sessions: session.NewManager(
			session.NewMemoryStorage(),
		),
		streams: streaming.NewStreamManager(),
		tools:   registry,
	}
}

func (a *aiApp) chat(w http.ResponseWriter, r *http.Request) {
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Status(http.StatusBadRequest).
			Category(contract.CategoryValidation).
			Code(contract.CodeInvalidJSON).
			Message("invalid request body").
			Build())
		return
	}
	if req.Message == "" {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeRequired).
			Detail("field", "message").
			Message("message is required").
			Build())
		return
	}

	ctx := r.Context()
	s, err := a.sessions.Create(ctx, session.CreateOptions{
		TenantID: "demo-tenant",
		UserID:   "demo-user",
		Model:    "mock-model",
	})
	if err != nil {
		writeInternal(w, r)
		return
	}
	_ = a.sessions.AppendMessage(ctx, s.ID, provider.NewTextMessage(provider.RoleUser, req.Message))

	a.provider.QueueResponse(&provider.CompletionResponse{
		Model: "mock-model",
		Content: []provider.ContentBlock{
			{Type: provider.ContentTypeText, Text: "offline reply: " + req.Message},
		},
	})
	resp, err := a.provider.Complete(ctx, &provider.CompletionRequest{
		Model:    "mock-model",
		Messages: []provider.Message{provider.NewTextMessage(provider.RoleUser, req.Message)},
	})
	if err != nil {
		writeInternal(w, r)
		return
	}

	toolResult, err := a.tools.Execute(ctx, "echo", map[string]any{"message": req.Message})
	if err != nil {
		writeInternal(w, r)
		return
	}
	toolOutput, _ := toolResult.Output.(map[string]any)

	_ = contract.WriteResponse(w, r, http.StatusOK, chatResponse{
		SessionID:   s.ID,
		Reply:       resp.GetText(),
		ToolOutput:  toolOutput,
		StreamCount: a.streams.Count(),
	}, nil)
}

func (a *aiApp) status(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"provider":     a.provider.Name(),
		"sessions":     "memory",
		"tools":        len(a.tools.ToProviderTools(r.Context())),
		"stream_count": a.streams.Count(),
		"live_network": false,
	}, nil)
}

func writeInternal(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeInternal).
		Message("internal error").
		Build())
}

func serve(app *core.App, cfg core.AppConfig) error {
	if err := app.Prepare(); err != nil {
		return fmt.Errorf("prepare server: %w", err)
	}
	srv, err := app.Server()
	if err != nil {
		return fmt.Errorf("get server: %w", err)
	}
	defer app.Shutdown(context.Background())

	log.Printf("Starting with-ai demo on %s", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
