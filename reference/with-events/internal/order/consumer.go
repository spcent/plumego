package order

import (
	"context"
	"sync/atomic"

	"github.com/spcent/plumego/x/messaging/pubsub"
)

type Subscriber interface {
	Subscribe(ctx context.Context, topic string, opts pubsub.SubOptions) (pubsub.Subscription, error)
}

type OrderConsumer struct {
	bus       Subscriber
	processed IdempotencyStore
	count     atomic.Int64
}

func NewConsumer(bus Subscriber, processed IdempotencyStore) *OrderConsumer {
	return &OrderConsumer{bus: bus, processed: processed}
}

func (c *OrderConsumer) Start(ctx context.Context) error {
	sub, err := c.bus.Subscribe(ctx, CreatedTopic, pubsub.SubOptions{BufferSize: 16})
	if err != nil {
		return err
	}
	go func() {
		defer sub.Cancel()
		for {
			select {
			case msg, ok := <-sub.C():
				if !ok {
					return
				}
				_ = c.ProcessMessage(ctx, msg)
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (c *OrderConsumer) ProcessMessage(_ context.Context, msg pubsub.Message) bool {
	if c == nil {
		return false
	}
	if c.processed != nil && c.processed.Seen(msg.ID) {
		return false
	}
	if c.processed != nil {
		c.processed.Mark(msg.ID)
	}
	c.count.Add(1)
	return true
}

func (c *OrderConsumer) Processed() int64 {
	if c == nil {
		return 0
	}
	return c.count.Load()
}
