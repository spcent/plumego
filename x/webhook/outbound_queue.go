package webhook

import (
	"context"
	"errors"
	"time"
)

var ErrQueueFull = errors.New("webhook queue full")

type Task struct {
	DeliveryID string
}

type Queue struct {
	ch         chan Task
	dropPolicy DropPolicy
	blockWait  time.Duration
}

func NewQueue(size int, policy DropPolicy, blockWait time.Duration) *Queue {
	return &Queue{
		ch:         make(chan Task, size),
		dropPolicy: policy,
		blockWait:  blockWait,
	}
}

func (q *Queue) Enqueue(ctx context.Context, t Task) error {
	switch q.dropPolicy {
	case DropNewest:
		select {
		case q.ch <- t:
			return nil
		default:
			return ErrQueueFull
		}
	case FailFast:
		select {
		case q.ch <- t:
			return nil
		default:
			return ErrQueueFull
		}
	case BlockWithLimit:
		if q.blockWait <= 0 {
			select {
			case q.ch <- t:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			default:
				return ErrQueueFull
			}
		}
		timer := time.NewTimer(q.blockWait)
		defer timer.Stop()
		select {
		case q.ch <- t:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return ErrQueueFull
		}
	default:
		select {
		case q.ch <- t:
			return nil
		default:
			return ErrQueueFull
		}
	}
}

func (q *Queue) Chan() <-chan Task { return q.ch }
func (q *Queue) Close()            { close(q.ch) }
