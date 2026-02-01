package pubsub

import (
	"context"
	"sync"
	"sync/atomic"
)

// workerPool manages a pool of workers for async publish operations.
// It limits the number of concurrent goroutines to prevent resource exhaustion.
type workerPool struct {
	sem     chan struct{} // semaphore for limiting concurrency
	wg      sync.WaitGroup
	closed  atomic.Bool
	maxSize int
}

// newWorkerPool creates a new worker pool with the given maximum size.
// If maxSize <= 0, the pool is effectively unlimited (uses a very large buffer).
func newWorkerPool(maxSize int) *workerPool {
	if maxSize <= 0 {
		maxSize = 1 << 20 // effectively unlimited
	}

	return &workerPool{
		sem:     make(chan struct{}, maxSize),
		maxSize: maxSize,
	}
}

// Submit submits a task to the worker pool.
// Returns false if the pool is closed.
// If the pool is at capacity, this will block until a worker is available.
func (wp *workerPool) Submit(task func()) bool {
	if wp.closed.Load() {
		return false
	}

	wp.wg.Add(1)

	// Acquire semaphore
	select {
	case wp.sem <- struct{}{}:
	default:
		// Pool is at capacity, wait
		wp.sem <- struct{}{}
	}

	go func() {
		defer func() {
			<-wp.sem // Release semaphore
			wp.wg.Done()
		}()
		task()
	}()

	return true
}

// TrySubmit attempts to submit a task without blocking.
// Returns false if the pool is closed or at capacity.
func (wp *workerPool) TrySubmit(task func()) bool {
	if wp.closed.Load() {
		return false
	}

	select {
	case wp.sem <- struct{}{}:
		wp.wg.Add(1)
		go func() {
			defer func() {
				<-wp.sem
				wp.wg.Done()
			}()
			task()
		}()
		return true
	default:
		return false
	}
}

// SubmitWithContext submits a task with context cancellation support.
// Returns false if the pool is closed or context is cancelled before acquiring a slot.
func (wp *workerPool) SubmitWithContext(ctx context.Context, task func()) bool {
	if wp.closed.Load() {
		return false
	}

	wp.wg.Add(1)

	select {
	case wp.sem <- struct{}{}:
		go func() {
			defer func() {
				<-wp.sem
				wp.wg.Done()
			}()
			task()
		}()
		return true
	case <-ctx.Done():
		wp.wg.Done()
		return false
	}
}

// Wait waits for all submitted tasks to complete.
func (wp *workerPool) Wait() {
	wp.wg.Wait()
}

// Close marks the pool as closed and waits for all tasks to complete.
func (wp *workerPool) Close() {
	if wp.closed.Swap(true) {
		return
	}
	wp.wg.Wait()
}

// Size returns the current number of active workers.
func (wp *workerPool) Size() int {
	return len(wp.sem)
}

// MaxSize returns the maximum pool size.
func (wp *workerPool) MaxSize() int {
	return wp.maxSize
}

// Available returns the number of available worker slots.
func (wp *workerPool) Available() int {
	return wp.maxSize - len(wp.sem)
}
