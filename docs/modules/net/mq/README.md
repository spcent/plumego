# In-Memory Message Queue

> **Package**: `github.com/spcent/plumego/x/mq` | **Feature**: Message broker

`x/mq` is a lower-level queue primitive package. For new feature discovery, start in `x/messaging` and drop to `x/mq` only when you need direct queue semantics.

## Overview

The `mq` package provides an in-process message queue with a pub/sub broker. It supports priority queues, message acknowledgment, dead-letter queuing, persistence, deduplication, and TTL-based expiry.

## Quick Start

```go
import "github.com/spcent/plumego/x/mq"

// Create broker
broker := mq.NewBroker(
    mq.WithWorkers(4),
    mq.WithQueueSize(10000),
)
broker.Start()
defer broker.Stop()

// Subscribe (consumer)
broker.Subscribe("orders", func(msg *mq.Message) error {
    var order Order
    json.Unmarshal(msg.Body, &order)
    return processOrder(order)
})

// Publish (producer)
orderJSON, _ := json.Marshal(order)
broker.Publish("orders", orderJSON)
```

## Broker Configuration

```go
broker := mq.NewBroker(
    // Worker pool
    mq.WithWorkers(8),
    mq.WithQueueSize(10000),

    // Retry policy
    mq.WithMaxRetries(3),
    mq.WithRetryDelay(time.Second),

    // Dead-letter queue
    mq.WithDeadLetterQueue("orders.dlq"),

    // Persistence (survive restarts)
    mq.WithPersistence(persistenceStore),

    // Message TTL
    mq.WithDefaultTTL(24 * time.Hour),
)
```

## Publishing

```go
// Simple publish
broker.Publish("topic", messageBytes)

// Publish with options
broker.PublishWithOptions("orders", messageBytes,
    mq.WithMessageID("order-123"),          // Idempotency key
    mq.WithTTL(1*time.Hour),               // Message expiry
    mq.WithPriority(mq.PriorityHigh),      // Priority queue
    mq.WithDelay(5*time.Second),           // Delayed delivery
    mq.WithHeaders(map[string]string{
        "source":     "payment-service",
        "event_type": "payment.completed",
    }),
)
```

## Subscribing

```go
// Single subscriber
broker.Subscribe("orders", handleOrder)

// Multiple subscribers (fan-out)
broker.Subscribe("order.created", updateInventory)
broker.Subscribe("order.created", sendConfirmationEmail)
broker.Subscribe("order.created", recordAnalytics)

// Unsubscribe
unsub := broker.Subscribe("orders", handleOrder)
unsub() // Remove subscription
```

### Handler Function

```go
func handleOrder(msg *mq.Message) error {
    // Message fields
    msg.ID          // Unique message ID
    msg.Topic       // Topic name
    msg.Body        // Message payload (bytes)
    msg.Headers     // Metadata headers
    msg.PublishedAt // Publication timestamp
    msg.Attempt     // Delivery attempt count (1 for first try)

    var order Order
    if err := json.Unmarshal(msg.Body, &order); err != nil {
        return err // Causes retry
    }

    if err := db.SaveOrder(context.Background(), order); err != nil {
        return err // Retry on database error
    }

    return nil // Acknowledge success
}
```

## Priority Queue

Messages can be delivered in priority order:

```go
const (
    mq.PriorityLow    = 0
    mq.PriorityNormal = 5  // Default
    mq.PriorityHigh   = 10
)

// High-priority messages processed first
broker.PublishWithOptions("notifications", urgentMsg,
    mq.WithPriority(mq.PriorityHigh),
)

broker.PublishWithOptions("notifications", normalMsg,
    mq.WithPriority(mq.PriorityNormal),
)
```

## Dead-Letter Queue

Messages that fail all retry attempts go to the DLQ:

```go
broker := mq.NewBroker(
    mq.WithDeadLetterQueue("orders.dlq"),
    mq.WithMaxRetries(3),
)

// Subscribe to DLQ for manual inspection
broker.Subscribe("orders.dlq", func(msg *mq.Message) error {
    log.Errorf("Dead-lettered message: id=%s attempts=%d error=%s",
        msg.ID, msg.Attempt, msg.Headers["last_error"])

    // Alert engineering team
    alerting.Send("message_dead_lettered", msg)

    return nil
})
```

## Deduplication

Prevent processing duplicate messages:

```go
broker := mq.NewBroker(
    mq.WithDeduplication(dedupStore, 1*time.Hour),
)

// Messages with same ID within 1 hour are deduplicated
broker.PublishWithOptions("payments", payload,
    mq.WithMessageID("payment-abc-123"),
)
```

## Message Acknowledgment

```go
// Manual acknowledgment (more control)
sub := broker.SubscribeWithAck("orders", func(msg *mq.Message, ack mq.Ackfunc) {
    if err := processOrder(msg); err != nil {
        ack(err) // Negative ack - will retry
        return
    }
    ack(nil) // Positive ack - mark delivered
})
```

## Persistence

Messages can be persisted to survive process restarts:

```go
// KV-backed persistence
kvStore, _ := kv.Open("./data/mq.kv")
persistence := mq.NewKVPersistence(kvStore)

broker := mq.NewBroker(
    mq.WithPersistence(persistence),
)

// SQL-backed persistence
sqlPersistence := mq.NewSQLPersistence(database,
    mq.WithTable("mq_messages"),
)
```

## Topics Pattern

```go
// Event-driven architecture with topics
const (
    TopicOrderCreated  = "order.created"
    TopicOrderShipped  = "order.shipped"
    TopicOrderDelivered = "order.delivered"
    TopicOrderCanceled = "order.canceled"
    TopicPaymentFailed = "payment.failed"
)

// Publish domain events
func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*Order, error) {
    order, err := s.db.Create(ctx, req)
    if err != nil {
        return nil, err
    }

    eventJSON, _ := json.Marshal(OrderCreatedEvent{
        OrderID:  order.ID,
        UserID:   order.UserID,
        TenantID: order.TenantID,
        Total:    order.Total,
    })
    broker.Publish(TopicOrderCreated, eventJSON)

    return order, nil
}

// Multiple independent consumers
broker.Subscribe(TopicOrderCreated, inventoryService.ReserveItems)
broker.Subscribe(TopicOrderCreated, emailService.SendConfirmation)
broker.Subscribe(TopicOrderCreated, analyticsService.TrackOrder)
broker.Subscribe(TopicOrderCreated, warehouseService.QueueFulfillment)
```

## Metrics

```go
// Emitted metrics
mq_messages_published_total{topic="orders"}
mq_messages_consumed_total{topic="orders", status="success|failure"}
mq_messages_dlq_total{topic="orders"}
mq_queue_depth{topic="orders"}
mq_processing_duration_seconds{topic="orders"}
```

## Comparison with PubSub

| Feature | `x/mq` | `pubsub` |
|---------|----------|----------|
| Message persistence | ✅ | ❌ |
| Priority queues | ✅ | ❌ |
| Dead-letter queue | ✅ | ❌ |
| Message acknowledgment | ✅ | ❌ |
| Deduplication | ✅ | ❌ |
| Fan-out | ✅ | ✅ |
| Simplicity | Medium | High |

Use `x/mq` for reliable task processing. Use `pubsub` for simple event distribution.

## Related Documentation

- [Net Overview](../README.md) — Network module
- [Archived Webhook Out Migration](/Users/bingrong.yan/projects/go/plumego/docs/legacy/modules/net/webhookout/README.md) - Historical webhook delivery notes
- [Scheduler](../../scheduler/README.md) — Task scheduling
