package core

import (
	"sync"
	"sync/atomic"
)

// ResettableOnce is a sync.Once that can be reset for testing purposes.
// It provides the same semantics as sync.Once but allows resetting the state
// to enable re-initialization in test scenarios.
type ResettableOnce struct {
	done uint32
	m    sync.Mutex
}

// Do executes the function f if and only if Do has not been called for this instance.
// In contrast to sync.Once, this implementation allows resetting the state.
func (o *ResettableOnce) Do(f func()) {
	if atomic.LoadUint32(&o.done) == 1 {
		return
	}
	// Slow path.
	o.doSlow(f)
}

func (o *ResettableOnce) doSlow(f func()) {
	o.m.Lock()
	defer o.m.Unlock()
	if o.done == 0 {
		defer atomic.StoreUint32(&o.done, 1)
		f()
	}
}

// Reset resets the Once state, allowing Do to be called again.
// This method is intended for testing purposes only.
func (o *ResettableOnce) Reset() {
	o.m.Lock()
	defer o.m.Unlock()
	atomic.StoreUint32(&o.done, 0)
}

// IsDone returns true if the Once has been executed.
func (o *ResettableOnce) IsDone() bool {
	return atomic.LoadUint32(&o.done) == 1
}
