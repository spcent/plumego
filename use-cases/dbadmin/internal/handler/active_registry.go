package handler

import "sync"

// activeRegistry is a generic, goroutine-safe map for tracking active items by string ID.
// T is the concrete info type stored per item (e.g. QueryInfo, OperationInfo).
type activeRegistry[T any] struct {
	mu    sync.RWMutex
	items map[string]*T
}

func newActiveRegistry[T any]() activeRegistry[T] {
	return activeRegistry[T]{items: make(map[string]*T)}
}

func (r *activeRegistry[T]) put(id string, v *T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[id] = v
}

func (r *activeRegistry[T]) remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, id)
}

func (r *activeRegistry[T]) get(id string) (*T, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.items[id]
	return v, ok
}

func (r *activeRegistry[T]) list() []*T {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*T, 0, len(r.items))
	for _, v := range r.items {
		result = append(result, v)
	}
	return result
}

func (r *activeRegistry[T]) count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.items)
}
