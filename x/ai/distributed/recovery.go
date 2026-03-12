package distributed

import (
	"context"
	"fmt"
	"time"
)

// RecoveryManager handles fault tolerance and task recovery.
type RecoveryManager struct {
	persistence WorkflowPersistence
	taskQueue   TaskQueue
	engine      *DistributedEngine
}

// NewRecoveryManager creates a new recovery manager.
func NewRecoveryManager(
	persistence WorkflowPersistence,
	taskQueue TaskQueue,
	engine *DistributedEngine,
) *RecoveryManager {
	return &RecoveryManager{
		persistence: persistence,
		taskQueue:   taskQueue,
		engine:      engine,
	}
}

// RecoverStaleTasks scans for failed/stale tasks and retries them.
func (rm *RecoveryManager) RecoverStaleTasks(ctx context.Context, staleThreshold time.Duration) (int, error) {
	// This would require scanning the KV store for tasks that haven't been updated
	// within the threshold. For now, return 0 as this requires key scanning support.
	return 0, nil
}

// RecoverFailedExecution attempts to recover a failed execution.
func (rm *RecoveryManager) RecoverFailedExecution(ctx context.Context, executionID string) error {
	// Load snapshot
	snapshot, err := rm.persistence.LoadSnapshot(ctx, executionID)
	if err != nil {
		return fmt.Errorf("failed to load snapshot: %w", err)
	}

	if snapshot.Status != StatusFailed {
		return fmt.Errorf("execution is not in failed state: %s", snapshot.Status)
	}

	// Resume from last checkpoint
	return rm.engine.Resume(ctx, executionID)
}

// MonitorTimeouts monitors for task timeouts and handles them.
func (rm *RecoveryManager) MonitorTimeouts(ctx context.Context, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Check for timed-out executions
			// This requires scanning snapshots
			// For production, implement proper timeout tracking
		}
	}
}

// RetryWithPolicy retries an operation with exponential backoff.
func RetryWithPolicy(
	ctx context.Context,
	policy *RetryPolicy,
	operation func() error,
) error {
	var lastErr error
	delay := policy.InitialDelay

	for i := 0; i <= policy.MaxRetries; i++ {
		lastErr = operation()
		if lastErr == nil {
			return nil
		}

		if i < policy.MaxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Calculate next delay with exponential backoff
				delay = time.Duration(float64(delay) * policy.Multiplier)
				if delay > policy.MaxDelay {
					delay = policy.MaxDelay
				}
			}
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", policy.MaxRetries, lastErr)
}
