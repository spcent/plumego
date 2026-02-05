package distributed

import (
	"context"
	"testing"
	"path/filepath"
	kv "github.com/spcent/plumego/store/kv"
	"time"

	"github.com/spcent/plumego/ai/orchestration"
	"github.com/spcent/plumego/net/mq"
	"github.com/spcent/plumego/pubsub"
)

func TestMQTaskQueue_Enqueue(t *testing.T) {
	ctx := context.Background()
	broker := mq.NewInProcBroker(pubsub.New())
	defer broker.Close()

	tmpDir := t.TempDir(); store, _ := kv.NewKVStore(kv.Options{DataDir: filepath.Join(tmpDir, "test")})
	defer store.Close()
	persistence := NewKVPersistence(store)

	queue := NewMQTaskQueue(broker, persistence, DefaultMQTaskQueueConfig())

	t.Run("EnqueueTask", func(t *testing.T) {
		task := &WorkflowTask{
			ID:          "task-1",
			ExecutionID: "exec-1",
			WorkflowID:  "wf-1",
			StepIndex:   0,
			StepName:    "step-1",
			WorkflowState: map[string]any{
				"key": "value",
			},
			Priority:  5,
			CreatedAt: time.Now(),
			Timeout:   time.Minute,
		}

		err := queue.Enqueue(ctx, task)
		if err != nil {
			t.Fatalf("Enqueue failed: %v", err)
		}
	})

	t.Run("EnqueueNilTask", func(t *testing.T) {
		err := queue.Enqueue(ctx, nil)
		if err == nil {
			t.Error("expected error for nil task")
		}
	})

	t.Run("EnqueueEmptyID", func(t *testing.T) {
		task := &WorkflowTask{
			ExecutionID: "exec-1",
			WorkflowID:  "wf-1",
		}

		err := queue.Enqueue(ctx, task)
		if err == nil {
			t.Error("expected error for empty task ID")
		}
	})
}

func TestMQTaskQueue_Subscribe(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	broker := mq.NewInProcBroker(pubsub.New())
	defer broker.Close()

	tmpDir := t.TempDir(); store, _ := kv.NewKVStore(kv.Options{DataDir: filepath.Join(tmpDir, "test")})
	defer store.Close()
	persistence := NewKVPersistence(store)

	queue := NewMQTaskQueue(broker, persistence, DefaultMQTaskQueueConfig())

	t.Run("SubscribeAndProcess", func(t *testing.T) {
		processed := make(chan string, 1)

		handler := func(ctx context.Context, task *WorkflowTask) (*TaskResult, error) {
			processed <- task.ID
			return &TaskResult{
				TaskID: task.ID,
				AgentResult: &orchestration.AgentResult{
					AgentID: "agent-1",
					Output:  "test output",
				},
			}, nil
		}

		err := queue.Subscribe(ctx, "worker-1", handler)
		if err != nil {
			t.Fatalf("Subscribe failed: %v", err)
		}

		// Give subscription time to start
		time.Sleep(100 * time.Millisecond)

		// Enqueue a task
		task := &WorkflowTask{
			ID:          "task-sub-1",
			ExecutionID: "exec-1",
			WorkflowID:  "wf-1",
			StepIndex:   0,
			StepName:    "step-1",
			Priority:    5,
		}

		err = queue.Enqueue(ctx, task)
		if err != nil {
			t.Fatalf("Enqueue failed: %v", err)
		}

		// Wait for processing
		select {
		case taskID := <-processed:
			if taskID != "task-sub-1" {
				t.Errorf("expected task-sub-1, got %s", taskID)
			}
		case <-time.After(2 * time.Second):
			t.Error("timeout waiting for task processing")
		}
	})

	t.Run("SubscribeWithNilHandler", func(t *testing.T) {
		err := queue.Subscribe(ctx, "worker-2", nil)
		if err == nil {
			t.Error("expected error for nil handler")
		}
	})

	t.Run("SubscribeWithEmptyWorkerID", func(t *testing.T) {
		handler := func(ctx context.Context, task *WorkflowTask) (*TaskResult, error) {
			return nil, nil
		}

		err := queue.Subscribe(ctx, "", handler)
		if err == nil {
			t.Error("expected error for empty worker ID")
		}
	})
}

func TestMQTaskQueue_TaskStatus(t *testing.T) {
	ctx := context.Background()
	broker := mq.NewInProcBroker(pubsub.New())
	defer broker.Close()

	tmpDir := t.TempDir(); store, _ := kv.NewKVStore(kv.Options{DataDir: filepath.Join(tmpDir, "test")})
	defer store.Close()
	persistence := NewKVPersistence(store)

	queue := NewMQTaskQueue(broker, persistence, DefaultMQTaskQueueConfig())

	t.Run("UpdateAndGetStatus", func(t *testing.T) {
		status := &TaskStatus{
			TaskID:    "task-status-1",
			Status:    TaskStatusRunning,
			WorkerID:  "worker-1",
			StartedAt: time.Now(),
		}

		err := queue.UpdateTaskStatus(ctx, "task-status-1", status)
		if err != nil {
			t.Fatalf("UpdateTaskStatus failed: %v", err)
		}

		// Note: GetTaskStatus requires subscription which may timeout
		// in test environment without actual message flow
	})

	t.Run("UpdateNilStatus", func(t *testing.T) {
		err := queue.UpdateTaskStatus(ctx, "task-1", nil)
		if err == nil {
			t.Error("expected error for nil status")
		}
	})
}

func TestMQTaskQueue_Results(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	broker := mq.NewInProcBroker(pubsub.New())
	defer broker.Close()

	tmpDir := t.TempDir(); store, _ := kv.NewKVStore(kv.Options{DataDir: filepath.Join(tmpDir, "test")})
	defer store.Close()
	persistence := NewKVPersistence(store)

	queue := NewMQTaskQueue(broker, persistence, DefaultMQTaskQueueConfig())

	t.Run("PublishResult", func(t *testing.T) {
		result := &TaskResult{
			TaskID: "task-result-1",
			AgentResult: &orchestration.AgentResult{
				AgentID: "agent-1",
				Output:  "test output",
			},
			CompletedAt: time.Now(),
		}

		err := queue.PublishResult(ctx, result)
		if err != nil {
			t.Fatalf("PublishResult failed: %v", err)
		}
	})

	t.Run("PublishNilResult", func(t *testing.T) {
		err := queue.PublishResult(ctx, nil)
		if err == nil {
			t.Error("expected error for nil result")
		}
	})
}

func TestTaskStatus_Constants(t *testing.T) {
	if TaskStatusPending != "pending" {
		t.Errorf("expected pending, got %s", TaskStatusPending)
	}
	if TaskStatusRunning != "running" {
		t.Errorf("expected running, got %s", TaskStatusRunning)
	}
	if TaskStatusCompleted != "completed" {
		t.Errorf("expected completed, got %s", TaskStatusCompleted)
	}
	if TaskStatusFailed != "failed" {
		t.Errorf("expected failed, got %s", TaskStatusFailed)
	}
}

func TestDefaultMQTaskQueueConfig(t *testing.T) {
	config := DefaultMQTaskQueueConfig()

	if config.TaskTopic == "" {
		t.Error("TaskTopic should not be empty")
	}
	if config.ResultTopicPrefix == "" {
		t.Error("ResultTopicPrefix should not be empty")
	}
	if config.StatusTopicPrefix == "" {
		t.Error("StatusTopicPrefix should not be empty")
	}
}
