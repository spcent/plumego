package mq

import (
	"context"
	"fmt"
	"testing"

	"github.com/spcent/plumego/pubsub"
)

func BenchmarkPublish(b *testing.B) {
	broker := NewInProcBroker(pubsub.New())
	defer broker.Close()

	ctx := context.Background()
	msg := Message{ID: "bench-1", Data: "benchmark data"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := broker.Publish(ctx, "bench-topic", msg); err != nil {
			b.Fatalf("publish error: %v", err)
		}
	}
}

func BenchmarkPublishBatch(b *testing.B) {
	broker := NewInProcBroker(pubsub.New())
	defer broker.Close()

	ctx := context.Background()
	msgs := make([]Message, 10)
	for i := 0; i < 10; i++ {
		msgs[i] = Message{ID: fmt.Sprintf("bench-%d", i), Data: "benchmark data"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := broker.PublishBatch(ctx, "bench-topic", msgs); err != nil {
			b.Fatalf("publish batch error: %v", err)
		}
	}
}

func BenchmarkPublishPriority(b *testing.B) {
	broker := NewInProcBroker(pubsub.New())
	defer broker.Close()

	ctx := context.Background()
	msg := PriorityMessage{
		Message:  Message{ID: "bench-1", Data: "benchmark data"},
		Priority: PriorityNormal,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := broker.PublishPriority(ctx, "bench-topic", msg); err != nil {
			b.Fatalf("publish priority error: %v", err)
		}
	}
}

func BenchmarkSubscribe(b *testing.B) {
	broker := NewInProcBroker(pubsub.New())
	defer broker.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sub, err := broker.Subscribe(ctx, fmt.Sprintf("bench-topic-%d", i), SubOptions{BufferSize: 10})
		if err != nil {
			b.Fatalf("subscribe error: %v", err)
		}
		sub.Cancel()
	}
}

func BenchmarkPublishSubscribe(b *testing.B) {
	broker := NewInProcBroker(pubsub.New())
	defer broker.Close()

	ctx := context.Background()
	topic := "bench-pubsub"

	// Subscribe first
	sub, err := broker.Subscribe(ctx, topic, SubOptions{BufferSize: b.N})
	if err != nil {
		b.Fatalf("subscribe error: %v", err)
	}
	defer sub.Cancel()

	msg := Message{ID: "bench-1", Data: "benchmark data"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := broker.Publish(ctx, topic, msg); err != nil {
			b.Fatalf("publish error: %v", err)
		}
		<-sub.C() // Receive the message
	}
}

func BenchmarkConfigValidation(b *testing.B) {
	cfg := DefaultConfig()
	cfg.EnableCluster = true
	cfg.ClusterNodeID = "node-1"
	cfg.EnablePersistence = true
	cfg.PersistencePath = "/tmp/mq-data"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := cfg.Validate(); err != nil {
			b.Fatalf("validate error: %v", err)
		}
	}
}

func BenchmarkValidateTopic(b *testing.B) {
	topic := "valid-topic-name"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := validateTopic(topic); err != nil {
			b.Fatalf("validate topic error: %v", err)
		}
	}
}

func BenchmarkValidateMessage(b *testing.B) {
	msg := Message{ID: "bench-1", Data: "benchmark data"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := validateMessage(msg); err != nil {
			b.Fatalf("validate message error: %v", err)
		}
	}
}

func BenchmarkHealthCheck(b *testing.B) {
	broker := NewInProcBroker(pubsub.New())
	defer broker.Close()

	// Subscribe to some topics to make health check more realistic
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		sub, _ := broker.Subscribe(ctx, fmt.Sprintf("health-topic-%d", i), SubOptions{BufferSize: 10})
		defer sub.Cancel()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = broker.HealthCheck()
	}
}

func BenchmarkMemoryUsage(b *testing.B) {
	broker := NewInProcBroker(pubsub.New())
	defer broker.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = broker.GetMemoryUsage()
	}
}

func BenchmarkCheckMemoryLimit(b *testing.B) {
	cfg := DefaultConfig()
	cfg.MaxMemoryUsage = 1 << 30 // 1GB limit
	broker := NewInProcBroker(pubsub.New(), WithConfig(cfg))
	defer broker.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := broker.checkMemoryLimit(); err != nil {
			b.Fatalf("check memory limit error: %v", err)
		}
	}
}

// Benchmark parallel publishing
func BenchmarkPublishParallel(b *testing.B) {
	broker := NewInProcBroker(pubsub.New())
	defer broker.Close()

	ctx := context.Background()
	msg := Message{ID: "bench-1", Data: "benchmark data"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := broker.Publish(ctx, "bench-topic", msg); err != nil {
				b.Errorf("publish error: %v", err)
			}
		}
	})
}

// Benchmark parallel subscribing
func BenchmarkSubscribeParallel(b *testing.B) {
	broker := NewInProcBroker(pubsub.New())
	defer broker.Close()

	ctx := context.Background()
	counter := 0

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			counter++
			sub, err := broker.Subscribe(ctx, fmt.Sprintf("bench-topic-%d", counter), SubOptions{BufferSize: 10})
			if err != nil {
				b.Errorf("subscribe error: %v", err)
			}
			sub.Cancel()
		}
	})
}
