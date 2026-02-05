package streaming

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/ai/orchestration"
)

func TestHandler(t *testing.T) {
	t.Run("NewHandler", func(t *testing.T) {
		engine := orchestration.NewEngine()
		streamEngine := NewStreamingEngine(engine, nil)
		handler := NewHandler(streamEngine)

		if handler == nil {
			t.Fatal("expected non-nil handler")
		}
	})

	t.Run("HandleStreamNoWorkflowID", func(t *testing.T) {
		engine := orchestration.NewEngine()
		streamEngine := NewStreamingEngine(engine, nil)
		handler := NewHandler(streamEngine)

		req := httptest.NewRequest(http.MethodGet, "/stream", nil)
		w := httptest.NewRecorder()

		handler.HandleStream(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("HandleStreamWithWorkflowID", func(t *testing.T) {
		engine := orchestration.NewEngine()
		streamEngine := NewStreamingEngine(engine, nil)
		handler := NewHandler(streamEngine)

		req := httptest.NewRequest(http.MethodGet, "/stream?workflow_id=test-123", nil)
		w := httptest.NewRecorder()

		handler.HandleStream(w, req)

		// Should setup SSE and start streaming
		if w.Code != 0 && w.Code != http.StatusOK {
			t.Errorf("unexpected status code: %d", w.Code)
		}

		// Check SSE headers
		contentType := w.Header().Get("Content-Type")
		if !strings.Contains(contentType, "text/event-stream") {
			t.Errorf("expected text/event-stream content type, got %s", contentType)
		}
	})

	t.Run("HandleExecuteInvalidMethod", func(t *testing.T) {
		engine := orchestration.NewEngine()
		streamEngine := NewStreamingEngine(engine, nil)
		handler := NewHandler(streamEngine)

		req := httptest.NewRequest(http.MethodGet, "/execute", nil)
		w := httptest.NewRecorder()

		handler.HandleExecute(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", w.Code)
		}
	})

	t.Run("HandleExecuteInvalidJSON", func(t *testing.T) {
		engine := orchestration.NewEngine()
		streamEngine := NewStreamingEngine(engine, nil)
		handler := NewHandler(streamEngine)

		req := httptest.NewRequest(http.MethodPost, "/execute", strings.NewReader("invalid json"))
		w := httptest.NewRecorder()

		handler.HandleExecute(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("HandleExecuteNoWorkflowID", func(t *testing.T) {
		engine := orchestration.NewEngine()
		streamEngine := NewStreamingEngine(engine, nil)
		handler := NewHandler(streamEngine)

		body := strings.NewReader(`{"name": "test-workflow"}`)
		req := httptest.NewRequest(http.MethodPost, "/execute", body)
		w := httptest.NewRecorder()

		handler.HandleExecute(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("HandleExecuteValid", func(t *testing.T) {
		engine := orchestration.NewEngine()
		streamEngine := NewStreamingEngine(engine, nil)
		handler := NewHandler(streamEngine)

		body := strings.NewReader(`{"workflow_id": "test-123", "name": "test-workflow"}`)
		req := httptest.NewRequest(http.MethodPost, "/execute", body)
		w := httptest.NewRecorder()

		handler.HandleExecute(w, req)

		// Should setup SSE and start execution
		output := w.Body.String()
		if output == "" {
			t.Error("expected output, got empty string")
		}
	})
}

func TestStreamWorkflow(t *testing.T) {
	t.Run("BasicExecution", func(t *testing.T) {
		engine := orchestration.NewEngine()
		streamEngine := NewStreamingEngine(engine, nil)

		workflow := &orchestration.Workflow{
			Name:  "test-workflow",
			Steps: []orchestration.Step{},
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		err := StreamWorkflow(w, req, workflow, "workflow-1", streamEngine)
		if err != nil {
			t.Fatalf("StreamWorkflow failed: %v", err)
		}

		output := w.Body.String()
		if output == "" {
			t.Error("expected output")
		}
	})

	t.Run("WithAgents", func(t *testing.T) {
		engine := orchestration.NewEngine()
		streamEngine := NewStreamingEngine(engine, nil)

		mockProvider := &mockProvider{
			output: "test output",
		}

		mockAgent := &orchestration.Agent{
			Name:     "test-agent",
			Provider: mockProvider,
		}

		workflow := &orchestration.Workflow{
			Name:  "test-workflow",
			State: make(map[string]any),
			Steps: []orchestration.Step{
				&orchestration.SequentialStep{
					StepName:  "test-step",
					Agent:     mockAgent,
					InputFn:   func(state map[string]any) string { return "test input" },
					OutputKey: "result",
				},
			},
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		err := StreamWorkflow(w, req, workflow, "workflow-1", streamEngine)
		if err != nil {
			t.Fatalf("StreamWorkflow failed: %v", err)
		}

		output := w.Body.String()
		if !contains(output, "test-") {
			t.Error("expected test- in output")
		}
	})
}

func TestHandleWithCallback(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		engine := orchestration.NewEngine()
		streamEngine := NewStreamingEngine(engine, nil)

		callback := func(ctx context.Context) (*orchestration.Workflow, error) {
			return &orchestration.Workflow{
				Name:  "callback-workflow",
				Steps: []orchestration.Step{},
			}, nil
		}

		handler := HandleWithCallback(streamEngine, callback)

		req := httptest.NewRequest(http.MethodGet, "/?workflow_id=test-123", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		output := w.Body.String()
		if output == "" {
			t.Error("expected output")
		}
	})

	t.Run("CallbackError", func(t *testing.T) {
		engine := orchestration.NewEngine()
		streamEngine := NewStreamingEngine(engine, nil)

		callback := func(ctx context.Context) (*orchestration.Workflow, error) {
			return nil, fmt.Errorf("callback error")
		}

		handler := HandleWithCallback(streamEngine, callback)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})

	t.Run("AutoGenerateWorkflowID", func(t *testing.T) {
		engine := orchestration.NewEngine()
		streamEngine := NewStreamingEngine(engine, nil)

		callback := func(ctx context.Context) (*orchestration.Workflow, error) {
			return &orchestration.Workflow{
				Name:  "test-workflow",
				Steps: []orchestration.Step{},
			}, nil
		}

		handler := HandleWithCallback(streamEngine, callback)

		// No workflow_id in query
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		// Should succeed with auto-generated ID
		output := w.Body.String()
		if output == "" {
			t.Error("expected output with auto-generated workflow ID")
		}
	})
}
