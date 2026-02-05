package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spcent/plumego/ai/orchestration"
	"github.com/spcent/plumego/ai/sse"
)

// WorkflowRequest represents a request to execute a workflow.
type WorkflowRequest struct {
	WorkflowID string                     `json:"workflow_id"`
	Name       string                     `json:"name"`
	Steps      []map[string]interface{}   `json:"steps,omitempty"`
	Metadata   map[string]string          `json:"metadata,omitempty"`
}

// Handler provides HTTP handlers for streaming workflows.
type Handler struct {
	engine *StreamingEngine
}

// NewHandler creates a new streaming workflow handler.
func NewHandler(engine *StreamingEngine) *Handler {
	return &Handler{
		engine: engine,
	}
}

// HandleStream handles SSE streaming requests for workflow execution.
func (h *Handler) HandleStream(w http.ResponseWriter, r *http.Request) {
	// Parse workflow ID from query or path
	workflowID := r.URL.Query().Get("workflow_id")
	if workflowID == "" {
		workflowID = r.URL.Query().Get("id")
	}
	if workflowID == "" {
		http.Error(w, "workflow_id required", http.StatusBadRequest)
		return
	}

	// Create SSE stream
	stream, err := sse.NewStream(r.Context(), w)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create SSE stream: %v", err), http.StatusInternalServerError)
		return
	}

	// Create a simple workflow (in production, load from DB or request body)
	workflow := &orchestration.Workflow{
		Name:  "streaming-workflow",
		Steps: []orchestration.Step{},
	}

	// Execute workflow with streaming
	_, err = h.engine.ExecuteStreaming(r.Context(), workflow, workflowID, stream)
	if err != nil {
		// Error already sent via stream
		return
	}

	// Send final completion event
	jsonData, _ := json.Marshal(map[string]string{
		"event":   "complete",
		"message": "Workflow execution completed",
	})
	stream.SendJSON("complete", string(jsonData))
}

// HandleExecute handles HTTP POST requests to execute workflows with streaming.
func (h *Handler) HandleExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req WorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if req.WorkflowID == "" {
		http.Error(w, "workflow_id required", http.StatusBadRequest)
		return
	}

	// Create SSE stream
	stream, err := sse.NewStream(r.Context(), w)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create SSE stream: %v", err), http.StatusInternalServerError)
		return
	}

	// Create workflow (simplified - in production, parse steps from request)
	workflow := &orchestration.Workflow{
		Name:  req.Name,
		Steps: []orchestration.Step{},
	}

	// Execute workflow with streaming
	results, err := h.engine.ExecuteStreaming(r.Context(), workflow, req.WorkflowID, stream)
	if err != nil {
		return
	}

	// Send final result
	jsonData, _ := json.Marshal(map[string]interface{}{
		"event":         "result",
		"success":       err == nil,
		"results_count": len(results),
	})
	stream.SendJSON("result", string(jsonData))
}

// StreamWorkflow is a convenience function to stream a workflow execution.
func StreamWorkflow(
	w http.ResponseWriter,
	r *http.Request,
	workflow *orchestration.Workflow,
	workflowID string,
	engine *StreamingEngine,
) error {
	// Create SSE stream
	stream, err := sse.NewStream(r.Context(), w)
	if err != nil {
		return fmt.Errorf("failed to create SSE stream: %w", err)
	}

	// Execute workflow with streaming
	_, err = engine.ExecuteStreaming(r.Context(), workflow, workflowID, stream)
	return err
}

// HandleWorkflowWithCallback handles workflow execution with custom callback.
type WorkflowCallback func(ctx context.Context) (*orchestration.Workflow, error)

// HandleWithCallback creates an HTTP handler that executes a workflow from a callback.
func HandleWithCallback(
	engine *StreamingEngine,
	callback WorkflowCallback,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get workflow ID
		workflowID := r.URL.Query().Get("workflow_id")
		if workflowID == "" {
			workflowID = fmt.Sprintf("workflow-%d", time.Now().Unix())
		}

		// Create workflow from callback
		workflow, err := callback(r.Context())
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create workflow: %v", err), http.StatusInternalServerError)
			return
		}

		// Stream workflow execution
		if err := StreamWorkflow(w, r, workflow, workflowID, engine); err != nil {
			// Error already handled by StreamWorkflow
			return
		}
	}
}
