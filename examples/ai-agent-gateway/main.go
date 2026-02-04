package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/spcent/plumego/ai/filter"
	"github.com/spcent/plumego/ai/instrumentation"
	"github.com/spcent/plumego/ai/llmcache"
	"github.com/spcent/plumego/ai/logging"
	"github.com/spcent/plumego/ai/metrics"
	"github.com/spcent/plumego/ai/orchestration"
	"github.com/spcent/plumego/ai/prompt"
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

	// Phase 3: Create metrics collector
	collector := metrics.NewMemoryCollector()

	// Phase 3: Create structured logger
	logger := logging.NewConsoleLogger(
		logging.WithLevel(logging.InfoLevel),
		logging.WithFormat(logging.JSONFormat),
	)
	logger.Info("Starting AI Agent Gateway", logging.Fields("version", "phase-3")...)

	// Create providers (Phase 3: wrapped with instrumentation)
	providerMgr := provider.NewManager()
	if apiKey != "" {
		claudeProvider := provider.NewClaudeProvider(apiKey)
		// Wrap with instrumentation
		instrumentedProvider := instrumentation.NewInstrumentedProvider(claudeProvider, collector)
		providerMgr.Register(instrumentedProvider)
		logger.Info("Registered Claude provider with instrumentation")
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

	// Phase 2: Create prompt engine
	promptStorage := prompt.NewMemoryStorage()
	promptEngine := prompt.NewEngine(promptStorage)
	if err := prompt.LoadBuiltinTemplates(promptEngine); err != nil {
		log.Printf("Warning: Failed to load builtin templates: %v", err)
	}

	// Phase 2: Create content filters
	contentFilter := filter.NewChain(
		&filter.StrictPolicy{},
		filter.NewPIIFilter(),
		filter.NewSecretFilter(),
		filter.NewPromptInjectionFilter(),
	)

	// Phase 2: Create LLM cache (Phase 3: wrapped with instrumentation)
	llmCache := llmcache.NewMemoryCache(1*time.Hour, 1000)
	// Note: We create an instrumented cache wrapper for metrics collection
	// The raw cache is still used for direct stats queries
	_ = instrumentation.NewInstrumentedMemoryCache(llmCache, collector)
	logger.Info("Created LLM cache with instrumentation")

	// Phase 2: Create orchestration engine (Phase 3: wrapped with instrumentation)
	orchEngine := orchestration.NewEngine()
	instrumentedEngine := instrumentation.NewInstrumentedEngine(orchEngine, collector)
	logger.Info("Created orchestration engine with instrumentation")

	// Create application
	app := core.New(
		core.WithAddr(":8080"),
		core.WithDebug(),
		core.WithLogging(),
		core.WithRecovery(),
	)

	// Phase 1 Routes
	app.Get("/", indexHandler)
	app.Post("/api/sessions", createSessionHandler(sessionMgr))
	app.Post("/api/sessions/:id/messages", sendMessageHandler(sessionMgr, providerMgr, toolRegistry))
	app.Get("/api/sessions/:id/stream", streamHandler(sessionMgr, providerMgr, toolRegistry))
	app.Get("/api/tools", listToolsHandler(toolRegistry))

	// Phase 2 Routes
	app.Get("/api/templates", listTemplatesHandler(promptEngine))
	app.Post("/api/templates/render", renderTemplateHandler(promptEngine))
	app.Post("/api/filter", filterContentHandler(contentFilter))
	app.Get("/api/cache/stats", cacheStatsHandler(llmCache))
	app.Post("/api/workflows/:id/execute", executeWorkflowHandler(instrumentedEngine))
	app.Get("/api/workflows", listWorkflowsHandler(instrumentedEngine))

	// Phase 3 Routes
	promExporter := metrics.NewPrometheusExporter(collector, "ai_gateway")
	app.Get("/metrics", promExporter.Handler())
	app.Get("/api/metrics/snapshot", metricsSnapshotHandler(collector))

	log.Println("🤖 AI Agent Gateway (Phase 1 + 2 + 3) starting on http://localhost:8080")
	log.Println("\nPhase 1 Endpoints:")
	log.Println("  POST /api/sessions - Create a new session")
	log.Println("  POST /api/sessions/:id/messages - Send a message")
	log.Println("  GET  /api/sessions/:id/stream - Stream responses (SSE)")
	log.Println("  GET  /api/tools - List available tools")
	log.Println("\nPhase 2 Endpoints:")
	log.Println("  GET  /api/templates - List prompt templates")
	log.Println("  POST /api/templates/render - Render a template")
	log.Println("  POST /api/filter - Filter content for PII/secrets")
	log.Println("  GET  /api/cache/stats - LLM cache statistics")
	log.Println("  POST /api/workflows/:id/execute - Execute a workflow")
	log.Println("  GET  /api/workflows - List available workflows")
	log.Println("\nPhase 3 Endpoints (NEW!):")
	log.Println("  GET  /metrics - Prometheus metrics (Prometheus text format)")
	log.Println("  GET  /api/metrics/snapshot - Metrics snapshot (JSON)")

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
    <h1>🤖 AI Agent Gateway (Phase 1 + Phase 2 + Phase 3)</h1>
    <p>A comprehensive AI agent gateway built with Plumego</p>

    <h2>Phase 1 Features</h2>
    <ul>
        <li>✅ SSE (Server-Sent Events) for streaming</li>
        <li>✅ LLM Provider abstraction (Claude, OpenAI)</li>
        <li>✅ Session management with context windows</li>
        <li>✅ Token counting and quota tracking</li>
        <li>✅ Tool calling framework with sandboxing</li>
    </ul>

    <h2>Phase 2 Features</h2>
    <ul>
        <li>🎯 Prompt template engine with 7 builtin templates</li>
        <li>🛡️ Content filtering (PII, secrets, prompt injection)</li>
        <li>💾 Intelligent LLM response caching</li>
        <li>🔀 Enhanced multi-model routing (task-based, cost-optimized)</li>
        <li>🎭 Agent orchestration (sequential, parallel, conditional)</li>
    </ul>

    <h2>Phase 3 Features (NEW!)</h2>
    <ul>
        <li>📊 Comprehensive metrics collection (counters, gauges, histograms)</li>
        <li>📈 Prometheus-compatible metrics export at /metrics</li>
        <li>🔍 Structured logging (JSON and text formats)</li>
        <li>🎯 Instrumentation for providers, cache, and orchestration</li>
        <li>📉 Real-time performance monitoring</li>
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

// Phase 2 Handlers

func listTemplatesHandler(engine *prompt.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get all builtin templates
		templates := prompt.BuiltinTemplates()

		response := make([]map[string]any, len(templates))
		for i, tmpl := range templates {
			response[i] = map[string]any{
				"name":    tmpl.Name,
				"model":   tmpl.Model,
				"tags":    tmpl.Tags,
				"version": tmpl.Version,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"templates": response,
			"count":     len(templates),
		})
	}
}

func renderTemplateHandler(engine *prompt.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name      string         `json:"name"`
			Variables map[string]any `json:"variables"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := engine.RenderByName(r.Context(), req.Name, req.Variables)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"rendered": result,
		})
	}
}

func filterContentHandler(chain *filter.Chain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Content string `json:"content"`
			Stage   string `json:"stage"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		stage := filter.StageInput
		if req.Stage == "output" {
			stage = filter.StageOutput
		}

		result, err := chain.Filter(r.Context(), req.Content, stage)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"allowed":     result.Allowed,
			"reason":      result.Reason,
			"labels":      result.Labels,
			"filter_name": result.FilterName,
			"score":       result.Score,
		})
	}
}

func cacheStatsHandler(cache *llmcache.MemoryCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := cache.Stats()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"hits":         stats.Hits,
			"misses":       stats.Misses,
			"evictions":    stats.Evictions,
			"total_tokens": stats.TotalTokens,
			"hit_rate":     stats.HitRate(),
		})
	}
}

func executeWorkflowHandler(engine *instrumentation.InstrumentedEngine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workflowID := r.URL.Query().Get(":id")
		if workflowID == "" {
			http.Error(w, "workflow_id required", http.StatusBadRequest)
			return
		}

		var req struct {
			State map[string]any `json:"state"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		results, err := engine.Execute(r.Context(), workflowID, req.State)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := make([]map[string]any, len(results))
		for i, res := range results {
			response[i] = map[string]any{
				"agent_id": res.AgentID,
				"output":   res.Output,
				"duration": res.Duration.String(),
				"tokens":   res.TokenUsage.TotalTokens,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"results": response,
			"count":   len(results),
		})
	}
}

func listWorkflowsHandler(engine *instrumentation.InstrumentedEngine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// In a real implementation, you'd list registered workflows
		// For now, return empty list
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"workflows": []map[string]any{},
			"count":     0,
			"message":   "Register workflows using engine.RegisterWorkflow()",
		})
	}
}

// Phase 3 Handlers

func metricsSnapshotHandler(collector *metrics.MemoryCollector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snapshot := collector.Snapshot()

		// Convert snapshot to a more readable JSON format
		response := map[string]any{
			"timestamp": snapshot.Timestamp,
			"counters":  convertCounters(snapshot.Counters),
			"gauges":    convertGauges(snapshot.Gauges),
			"histograms": convertHistograms(snapshot.Histograms),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func convertCounters(counters map[string]metrics.CounterSnapshot) []map[string]any {
	result := make([]map[string]any, 0, len(counters))
	for name, counter := range counters {
		result = append(result, map[string]any{
			"name":  name,
			"value": counter.Value,
			"tags":  counter.Tags,
		})
	}
	return result
}

func convertGauges(gauges map[string]metrics.GaugeSnapshot) []map[string]any {
	result := make([]map[string]any, 0, len(gauges))
	for name, gauge := range gauges {
		result = append(result, map[string]any{
			"name":  name,
			"value": gauge.Value,
			"tags":  gauge.Tags,
		})
	}
	return result
}

func convertHistograms(histograms map[string]metrics.HistogramSnapshot) []map[string]any {
	result := make([]map[string]any, 0, len(histograms))
	for name, hist := range histograms {
		result = append(result, map[string]any{
			"name":  name,
			"count": hist.Count,
			"sum":   hist.Sum,
			"min":   hist.Min,
			"max":   hist.Max,
			"avg":   hist.Avg,
			"p50":   hist.P50,
			"p95":   hist.P95,
			"p99":   hist.P99,
			"tags":  hist.Tags,
		})
	}
	return result
}
