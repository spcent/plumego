package pubsub

import (
	"context"
	"sync"
)

// workerPool manages a pool of workers for async publish operations.
// It limits the number of concurrent goroutines to prevent resource exhaustion.
type workerPool struct {
	sem     chan struct{} // semaphore for limiting concurrency
	wg      sync.WaitGroup
	mu      sync.Mutex // serialises closed check with wg.Add to prevent TOCTOU
	closed  bool
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
	// Hold the mutex across the closed check AND wg.Add so that Close() cannot
	// call wg.Wait() between the check and the Add, which would panic.
	wp.mu.Lock()
	if wp.closed {
		wp.mu.Unlock()
		return false
	}
	wp.wg.Add(1)
	wp.mu.Unlock()

	// Block until a semaphore slot is available.
	wp.sem <- struct{}{}

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
	select {
	case wp.sem <- struct{}{}:
	default:
		return false
	}

	wp.mu.Lock()
	if wp.closed {
		wp.mu.Unlock()
		<-wp.sem // release the slot we just acquired
		return false
	}
	wp.wg.Add(1)
	wp.mu.Unlock()

	go func() {
		defer func() {
			<-wp.sem
			wp.wg.Done()
		}()
		task()
	}()
	return true
}

// SubmitWithContext submits a task with context cancellation support.
// Returns false if the pool is closed or context is cancelled before acquiring a slot.
func (wp *workerPool) SubmitWithContext(ctx context.Context, task func()) bool {
	select {
	case wp.sem <- struct{}{}:
	case <-ctx.Done():
		return false
	}

	wp.mu.Lock()
	if wp.closed {
		wp.mu.Unlock()
		<-wp.sem
		return false
	}
	wp.wg.Add(1)
	wp.mu.Unlock()

	go func() {
		defer func() {
			<-wp.sem
			wp.wg.Done()
		}()
		task()
	}()
	return true
}

// Wait waits for all submitted tasks to complete.
func (wp *workerPool) Wait() {
	wp.wg.Wait()
}

// Close marks the pool as closed and waits for all tasks to complete.
func (wp *workerPool) Close() {
	wp.mu.Lock()
	if wp.closed {
		wp.mu.Unlock()
		return
	}
	wp.closed = true
	wp.mu.Unlock()
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
