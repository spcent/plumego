package webhook

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

type delayedItem struct {
	at   time.Time
	task Task
	idx  int
}

type delayHeap []*delayedItem

func (h delayHeap) Len() int           { return len(h) }
func (h delayHeap) Less(i, j int) bool { return h[i].at.Before(h[j].at) }
func (h delayHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i]; h[i].idx = i; h[j].idx = j }
func (h *delayHeap) Push(x any)        { it := x.(*delayedItem); it.idx = len(*h); *h = append(*h, it) }
func (h *delayHeap) Pop() any          { old := *h; n := len(old); it := old[n-1]; *h = old[:n-1]; return it }
func (h delayHeap) Peek() *delayedItem {
	if len(h) == 0 {
		return nil
	}
	return h[0]
}

type DelayScheduler struct {
	mu    sync.Mutex
	h     delayHeap
	wake  chan struct{}
	stop  chan struct{}
	queue *Queue
}

func NewDelayScheduler(q *Queue) *DelayScheduler {
	ds := &DelayScheduler{
		h:     make(delayHeap, 0),
		wake:  make(chan struct{}, 1),
		stop:  make(chan struct{}),
		queue: q,
	}
	heap.Init(&ds.h)
	return ds
}

func (d *DelayScheduler) Schedule(at time.Time, task Task) {
	d.mu.Lock()
	heap.Push(&d.h, &delayedItem{at: at, task: task})
	d.mu.Unlock()
	select {
	case d.wake <- struct{}{}:
	default:
	}
}

func (d *DelayScheduler) Run(ctx context.Context) {
	for {
		nextWait := d.nextWaitDuration()
		var timer <-chan time.Time
		if nextWait > 0 {
			timer = time.After(nextWait)
		}

		select {
		case <-ctx.Done():
			return
		case <-d.stop:
			return
		case <-d.wake:
			continue
		case <-timer:
			d.flushDue(ctx)
		}
	}
}

func (d *DelayScheduler) Stop() {
	close(d.stop)
}

func (d *DelayScheduler) nextWaitDuration() time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()
	it := d.h.Peek()
	if it == nil {
		return time.Second // idle poll
	}
	now := time.Now().UTC()
	if !it.at.After(now) {
		return 0
	}
	return it.at.Sub(now)
}

func (d *DelayScheduler) flushDue(ctx context.Context) {
	for {
		d.mu.Lock()
		it := d.h.Peek()
		if it == nil {
			d.mu.Unlock()
			return
		}
		now := time.Now().UTC()
		if it.at.After(now) {
			d.mu.Unlock()
			return
		}
		heap.Pop(&d.h)
		d.mu.Unlock()

		_ = d.queue.Enqueue(ctx, it.task) // enqueue failure is accounted elsewhere (best-effort)
	}
}
