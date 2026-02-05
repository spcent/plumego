package streaming

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spcent/plumego/ai/orchestration"
	"github.com/spcent/plumego/ai/sse"
)

func TestStreamManager(t *testing.T) {
	t.Run("RegisterAndGet", func(t *testing.T) {
		sm := NewStreamManager()

		w := httptest.NewRecorder()
		stream := sse.NewStream(w)

		sm.Register("workflow-1", stream)

		retrieved, ok := sm.Get("workflow-1")
		if !ok {
			t.Fatal("expected to find registered stream")
		}

		if retrieved != stream {
			t.Error("retrieved stream does not match registered stream")
		}
	})

	t.Run("Unregister", func(t *testing.T) {
		sm := NewStreamManager()

		w := httptest.NewRecorder()
		stream := sse.NewStream(w)

		sm.Register("workflow-1", stream)
		sm.Unregister("workflow-1")

		_, ok := sm.Get("workflow-1")
		if ok {
			t.Error("stream should have been unregistered")
		}
	})

	t.Run("Count", func(t *testing.T) {
		sm := NewStreamManager()

		if sm.Count() != 0 {
			t.Error("expected 0 streams initially")
		}

		w1 := httptest.NewRecorder()
		w2 := httptest.NewRecorder()
		sm.Register("workflow-1", sse.NewStream(w1))
		sm.Register("workflow-2", sse.NewStream(w2))

		if sm.Count() != 2 {
			t.Errorf("expected 2 streams, got %d", sm.Count())
		}
	})

	t.Run("SendUpdate", func(t *testing.T) {
		sm := NewStreamManager()

		w := httptest.NewRecorder()
		stream := sse.NewStream(w)
		stream.Setup()

		sm.Register("workflow-1", stream)

		update := &ProgressUpdate{
			WorkflowID: "workflow-1",
			StepName:   "test-step",
			Status:     StatusStarted,
			Progress:   0.5,
			Timestamp:  time.Now(),
		}

		err := sm.SendUpdate("workflow-1", update)
		if err != nil {
			t.Fatalf("SendUpdate failed: %v", err)
		}

		// Check that data was written
		output := w.Body.String()
		if output == "" {
			t.Error("expected output, got empty string")
		}
	})

	t.Run("SendUpdateNotFound", func(t *testing.T) {
		sm := NewStreamManager()

		update := &ProgressUpdate{
			WorkflowID: "nonexistent",
			StepName:   "test",
			Status:     StatusStarted,
			Timestamp:  time.Now(),
		}

		err := sm.SendUpdate("nonexistent", update)
		if err == nil {
			t.Error("expected error for nonexistent stream")
		}
	})

	t.Run("Close", func(t *testing.T) {
		sm := NewStreamManager()

		w := httptest.NewRecorder()
		stream := sse.NewStream(w)

		sm.Register("workflow-1", stream)

		err := sm.Close("workflow-1")
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}

		_, ok := sm.Get("workflow-1")
		if ok {
			t.Error("stream should have been closed and unregistered")
		}
	})
}

func TestProgressUpdate(t *testing.T) {
	t.Run("MarshalJSON", func(t *testing.T) {
		update := &ProgressUpdate{
			WorkflowID: "workflow-1",
			StepName:   "test-step",
			Status:     StatusCompleted,
			Progress:   1.0,
			Timestamp:  time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		}

		data, err := json.Marshal(update)
		if err != nil {
			t.Fatalf("Marshal failed: %v", err)
		}

		// Check that timestamp is formatted correctly
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		timestamp, ok := result["timestamp"].(string)
		if !ok {
			t.Fatal("timestamp should be a string")
		}

		if timestamp == "" {
			t.Error("timestamp should not be empty")
		}
	})
}

func TestStreamConfig(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultStreamConfig()

		if !config.EnableProgress {
			t.Error("EnableProgress should be true by default")
		}

		if !config.EnableStepLogs {
			t.Error("EnableStepLogs should be true by default")
		}

		if config.KeepAlive != 15*time.Second {
			t.Errorf("expected KeepAlive 15s, got %v", config.KeepAlive)
		}

		if config.BufferSize != 100 {
			t.Errorf("expected BufferSize 100, got %d", config.BufferSize)
		}
	})
}

func TestStreamingEngine(t *testing.T) {
	t.Run("NewStreamingEngine", func(t *testing.T) {
		engine := orchestration.NewEngine()
		streamEngine := NewStreamingEngine(engine, nil)

		if streamEngine == nil {
			t.Fatal("expected non-nil streaming engine")
		}

		if streamEngine.StreamManager() == nil {
			t.Error("expected non-nil stream manager")
		}
	})

	t.Run("ExecuteStreamingEmptyWorkflow", func(t *testing.T) {
		engine := orchestration.NewEngine()
		streamEngine := NewStreamingEngine(engine, nil)

		workflow := &orchestration.Workflow{
			Name:  "test-workflow",
			Steps: []orchestration.Step{},
		}

		w := httptest.NewRecorder()
		stream := sse.NewStream(w)
		stream.Setup()

		ctx := context.Background()
		result, err := streamEngine.ExecuteStreaming(ctx, workflow, "workflow-1", stream)
		if err != nil {
			t.Fatalf("ExecuteStreaming failed: %v", err)
		}

		if !result.Success {
			t.Error("expected successful result for empty workflow")
		}

		// Check that events were sent
		output := w.Body.String()
		if output == "" {
			t.Error("expected output events")
		}
	})

	t.Run("ExecuteStreamingWithAgentStep", func(t *testing.T) {
		engine := orchestration.NewEngine()
		streamEngine := NewStreamingEngine(engine, nil)

		mockAgent := &orchestration.Agent{
			Name: "test-agent",
			System: "You are a test agent",
			Execute: func(ctx context.Context) (*orchestration.AgentResult, error) {
				return &orchestration.AgentResult{
					Success: true,
					Output:  "test output",
				}, nil
			},
		}

		workflow := &orchestration.Workflow{
			Name: "test-workflow",
			Steps: []orchestration.Step{
				&orchestration.AgentStep{Agent: mockAgent},
			},
		}

		w := httptest.NewRecorder()
		stream := sse.NewStream(w)
		stream.Setup()

		ctx := context.Background()
		result, err := streamEngine.ExecuteStreaming(ctx, workflow, "workflow-1", stream)
		if err != nil {
			t.Fatalf("ExecuteStreaming failed: %v", err)
		}

		if !result.Success {
			t.Error("expected successful result")
		}

		// Parse events from output
		output := w.Body.String()
		if output == "" {
			t.Fatal("expected output events")
		}

		// Should contain started and completed events
		if !contains(output, "started") {
			t.Error("expected started event")
		}
		if !contains(output, "completed") {
			t.Error("expected completed event")
		}
	})

	t.Run("ExecuteStreamingWithError", func(t *testing.T) {
		engine := orchestration.NewEngine()
		streamEngine := NewStreamingEngine(engine, nil)

		mockAgent := &orchestration.Agent{
			Name: "failing-agent",
			Execute: func(ctx context.Context) (*orchestration.AgentResult, error) {
				return nil, fmt.Errorf("agent error")
			},
		}

		workflow := &orchestration.Workflow{
			Name: "test-workflow",
			Steps: []orchestration.Step{
				&orchestration.AgentStep{Agent: mockAgent},
			},
		}

		w := httptest.NewRecorder()
		stream := sse.NewStream(w)
		stream.Setup()

		ctx := context.Background()
		_, err := streamEngine.ExecuteStreaming(ctx, workflow, "workflow-1", stream)
		if err == nil {
			t.Error("expected error from failing agent")
		}

		// Parse events from output
		output := w.Body.String()
		if !contains(output, "failed") {
			t.Error("expected failed event")
		}
	})
}

func TestGetStepName(t *testing.T) {
	tests := []struct {
		name     string
		step     orchestration.Step
		expected string
	}{
		{
			name: "AgentStep",
			step: &orchestration.AgentStep{
				Agent: &orchestration.Agent{Name: "test-agent"},
			},
			expected: "test-agent",
		},
		{
			name:     "SequentialStep",
			step:     &orchestration.SequentialStep{},
			expected: "sequential",
		},
		{
			name:     "ParallelStep",
			step:     &orchestration.ParallelStep{},
			expected: "parallel",
		},
		{
			name:     "ConditionalStep",
			step:     &orchestration.ConditionalStep{},
			expected: "conditional",
		},
		{
			name:     "RetryStep",
			step:     &orchestration.RetryStep{},
			expected: "retry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStepName(tt.step)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetStepType(t *testing.T) {
	tests := []struct {
		name     string
		step     orchestration.Step
		expected string
	}{
		{
			name:     "AgentStep",
			step:     &orchestration.AgentStep{},
			expected: "agent",
		},
		{
			name:     "SequentialStep",
			step:     &orchestration.SequentialStep{},
			expected: "sequential",
		},
		{
			name:     "ParallelStep",
			step:     &orchestration.ParallelStep{},
			expected: "parallel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStepType(tt.step)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
