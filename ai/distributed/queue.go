package distributed

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spcent/plumego/ai/orchestration"
	"github.com/spcent/plumego/net/mq"
	"github.com/spcent/plumego/pubsub"
)

// TaskStatus represents the status of a workflow task.
type TaskStatus struct {
	TaskID      string    `json:"task_id"`
	Status      string    `json:"status"` // pending/running/completed/failed
	WorkerID    string    `json:"worker_id,omitempty"`
	StartedAt   time.Time `json:"started_at,omitempty"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	Retries     int       `json:"retries"`
	LastError   string    `json:"last_error,omitempty"`
}

// Task status constants
const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"
)

// WorkflowTask represents a distributed workflow step to be executed.
type WorkflowTask struct {
	ID            string                 `json:"id"`
	ExecutionID   string                 `json:"execution_id"`
	WorkflowID    string                 `json:"workflow_id"`
	StepIndex     int                    `json:"step_index"`
	StepName      string                 `json:"step_name"`
	WorkflowState map[string]any         `json:"workflow_state"`
	Priority      int                    `json:"priority"`
	CreatedAt     time.Time              `json:"created_at"`
	Timeout       time.Duration          `json:"timeout"`
	Metadata      map[string]string      `json:"metadata,omitempty"`
}

// TaskResult represents the result of task execution.
type TaskResult struct {
	TaskID       string                       `json:"task_id"`
	AgentResult  *orchestration.AgentResult   `json:"agent_result"`
	UpdatedState map[string]any               `json:"updated_state"`
	Error        string                       `json:"error,omitempty"`
	CompletedAt  time.Time                    `json:"completed_at"`
}

// TaskHandler is a function that processes a workflow task.
type TaskHandler func(ctx context.Context, task *WorkflowTask) (*TaskResult, error)

// TaskQueue defines the interface for distributed task queue operations.
type TaskQueue interface {
	// Enqueue adds a task to the queue.
	Enqueue(ctx context.Context, task *WorkflowTask) error

	// Subscribe allows a worker to receive and process tasks.
	Subscribe(ctx context.Context, workerID string, handler TaskHandler) error

	// UpdateTaskStatus updates the status of a task.
	UpdateTaskStatus(ctx context.Context, taskID string, status *TaskStatus) error

	// GetTaskStatus retrieves the current status of a task.
	GetTaskStatus(ctx context.Context, taskID string) (*TaskStatus, error)

	// PublishResult publishes a task result.
	PublishResult(ctx context.Context, result *TaskResult) error

	// SubscribeResults subscribes to task results.
	SubscribeResults(ctx context.Context, executionID string) (<-chan *TaskResult, error)

	// Close closes the task queue.
	Close() error
}

// MQTaskQueue implements TaskQueue using the message queue.
type MQTaskQueue struct {
	broker            mq.Broker
	taskTopic         string
	resultTopicPrefix string
	statusTopicPrefix string
	persistence       WorkflowPersistence
}

// MQTaskQueueConfig configures the MQ-based task queue.
type MQTaskQueueConfig struct {
	TaskTopic         string
	ResultTopicPrefix string
	StatusTopicPrefix string
}

// DefaultMQTaskQueueConfig returns default configuration.
func DefaultMQTaskQueueConfig() MQTaskQueueConfig {
	return MQTaskQueueConfig{
		TaskTopic:         "distributed:tasks",
		ResultTopicPrefix: "distributed:results:",
		StatusTopicPrefix: "distributed:status:",
	}
}

// NewMQTaskQueue creates a new MQ-based task queue.
func NewMQTaskQueue(broker mq.Broker, persistence WorkflowPersistence, config MQTaskQueueConfig) *MQTaskQueue {
	if config.TaskTopic == "" {
		config = DefaultMQTaskQueueConfig()
	}

	return &MQTaskQueue{
		broker:            broker,
		taskTopic:         config.TaskTopic,
		resultTopicPrefix: config.ResultTopicPrefix,
		statusTopicPrefix: config.StatusTopicPrefix,
		persistence:       persistence,
	}
}

// Enqueue adds a task to the queue.
func (q *MQTaskQueue) Enqueue(ctx context.Context, task *WorkflowTask) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}
	if task.ID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	// Set creation time if not set
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}

	// Serialize task
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// Create message with priority
	msg := pubsub.Message{
		ID:   task.ID,
		Data: data,
		Meta: map[string]string{"priority": fmt.Sprintf("%d", task.Priority)},
	}

	// Publish to task topic
	if err := q.broker.Publish(ctx, q.taskTopic, msg); err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	// Update task status to pending
	status := &TaskStatus{
		TaskID:    task.ID,
		Status:    TaskStatusPending,
		StartedAt: task.CreatedAt,
	}

	return q.UpdateTaskStatus(ctx, task.ID, status)
}

// Subscribe allows a worker to receive and process tasks.
func (q *MQTaskQueue) Subscribe(ctx context.Context, workerID string, handler TaskHandler) error {
	if workerID == "" {
		return fmt.Errorf("worker ID cannot be empty")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	// Subscribe to task topic
	sub, err := q.broker.Subscribe(ctx, q.taskTopic, mq.SubOptions{
		BufferSize: 100,
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to tasks: %w", err)
	}

	// Process tasks in background
	go q.processTasks(ctx, workerID, sub, handler)

	return nil
}

// processTasks processes incoming tasks.
func (q *MQTaskQueue) processTasks(ctx context.Context, workerID string, sub mq.Subscription, handler TaskHandler) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-sub.C():
			if !ok {
				return
			}

			// Deserialize task
			var task WorkflowTask
			data, ok := msg.Data.([]byte)
			if !ok {
				continue
			}
			if err := json.Unmarshal(data, &task); err != nil {
				continue
			}

			// Update status to running
			status := &TaskStatus{
				TaskID:    task.ID,
				Status:    TaskStatusRunning,
				WorkerID:  workerID,
				StartedAt: time.Now(),
			}
			q.UpdateTaskStatus(ctx, task.ID, status)

			// Execute task
			result, err := handler(ctx, &task)

			// Update status based on result
			if err != nil {
				status.Status = TaskStatusFailed
				status.LastError = err.Error()
				status.CompletedAt = time.Now()
			} else {
				status.Status = TaskStatusCompleted
				status.CompletedAt = time.Now()

				// Publish result
				if result != nil {
					result.TaskID = task.ID
					result.CompletedAt = time.Now()
					q.PublishResult(ctx, result)
				}
			}

			q.UpdateTaskStatus(ctx, task.ID, status)
		}
	}
}

// UpdateTaskStatus updates the status of a task.
func (q *MQTaskQueue) UpdateTaskStatus(ctx context.Context, taskID string, status *TaskStatus) error {
	if taskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	if status == nil {
		return fmt.Errorf("status cannot be nil")
	}

	// Serialize status
	data, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	// Publish status update
	topic := q.statusTopicPrefix + taskID
	msg := pubsub.Message{
		ID:   taskID + "-status",
		Data: data,
	}

	return q.broker.Publish(ctx, topic, msg)
}

// GetTaskStatus retrieves the current status of a task.
func (q *MQTaskQueue) GetTaskStatus(ctx context.Context, taskID string) (*TaskStatus, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task ID cannot be empty")
	}

	// Subscribe to status topic
	topic := q.statusTopicPrefix + taskID
	sub, err := q.broker.Subscribe(ctx, topic, mq.SubOptions{
		BufferSize: 1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to status: %w", err)
	}
	defer sub.Cancel()

	// Wait for status update with timeout
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg, ok := <-sub.C():
		if !ok {
			return nil, fmt.Errorf("status subscription closed")
		}

		var status TaskStatus
		data, ok := msg.Data.([]byte)
		if !ok {
			return nil, fmt.Errorf("invalid message data type")
		}
		if err := json.Unmarshal(data, &status); err != nil {
			return nil, fmt.Errorf("failed to unmarshal status: %w", err)
		}

		return &status, nil
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout waiting for task status")
	}
}

// PublishResult publishes a task result.
func (q *MQTaskQueue) PublishResult(ctx context.Context, result *TaskResult) error {
	if result == nil {
		return fmt.Errorf("result cannot be nil")
	}
	if result.TaskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	// Serialize result
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	// Publish to result topic
	// Extract execution ID from task metadata or use task ID
	topic := q.resultTopicPrefix + result.TaskID
	msg := pubsub.Message{
		ID:   result.TaskID + "-result",
		Data: data,
	}

	return q.broker.Publish(ctx, topic, msg)
}

// SubscribeResults subscribes to task results.
func (q *MQTaskQueue) SubscribeResults(ctx context.Context, executionID string) (<-chan *TaskResult, error) {
	if executionID == "" {
		return nil, fmt.Errorf("execution ID cannot be empty")
	}

	// Subscribe to result topic pattern
	topic := q.resultTopicPrefix + executionID + ":*"
	sub, err := q.broker.Subscribe(ctx, topic, mq.SubOptions{
		BufferSize: 100,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to results: %w", err)
	}

	// Create result channel
	results := make(chan *TaskResult, 100)

	// Process results in background
	go func() {
		defer close(results)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-sub.C():
				if !ok {
					return
				}

				var result TaskResult
				data, ok := msg.Data.([]byte)
				if !ok {
					continue
				}
				if err := json.Unmarshal(data, &result); err != nil {
					continue
				}

				select {
				case results <- &result:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return results, nil
}

// Close closes the task queue.
func (q *MQTaskQueue) Close() error {
	if q.broker != nil {
		return q.broker.Close()
	}
	return nil
}
