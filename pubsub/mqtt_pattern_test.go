package pubsub

import (
	"context"
	"sort"
	"sync"
	"testing"
	"time"
)

// --- Unit tests for MQTTMatch ---

func TestMQTTMatch(t *testing.T) {
	tests := []struct {
		pattern string
		topic   string
		want    bool
	}{
		// Exact match
		{"a/b/c", "a/b/c", true},
		{"a/b/c", "a/b/d", false},
		{"a", "a", true},
		{"a", "b", false},

		// Single-level wildcard (+)
		{"a/+/c", "a/b/c", true},
		{"a/+/c", "a/x/c", true},
		{"a/+/c", "a/b/d", false},
		{"a/+/c", "a/b/c/d", false},
		{"+/b/c", "a/b/c", true},
		{"+/+/+", "a/b/c", true},
		{"+", "a", true},
		{"+", "a/b", false},

		// Multi-level wildcard (#)
		{"a/#", "a", true},
		{"a/#", "a/b", true},
		{"a/#", "a/b/c", true},
		{"a/#", "a/b/c/d", true},
		{"a/#", "b", false},
		{"#", "a", true},
		{"#", "a/b/c", true},

		// Combined
		{"a/+/c/#", "a/b/c", true},
		{"a/+/c/#", "a/b/c/d", true},
		{"a/+/c/#", "a/b/c/d/e", true},
		{"+/#", "a", true},
		{"+/#", "a/b", true},

		// Empty
		{"", "", true},
		{"", "a", false},
		{"a", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_vs_"+tt.topic, func(t *testing.T) {
			got := MQTTMatch(tt.pattern, tt.topic)
			if got != tt.want {
				t.Errorf("MQTTMatch(%q, %q) = %v, want %v", tt.pattern, tt.topic, got, tt.want)
			}
		})
	}
}

func TestMQTTMatcher(t *testing.T) {
	m := &MQTTMatcher{}
	if !m.Match("devices/+/temp", "devices/sensor1/temp") {
		t.Error("MQTTMatcher.Match should match single-level wildcard")
	}
	if m.Match("devices/+/temp", "devices/sensor1/humidity") {
		t.Error("MQTTMatcher.Match should not match different last level")
	}
}

func TestValidateMQTTPattern(t *testing.T) {
	tests := []struct {
		pattern string
		wantErr bool
	}{
		{"a/b/c", false},
		{"a/+/c", false},
		{"a/#", false},
		{"+/+/+", false},
		{"#", false},
		{"+", false},

		// Invalid
		{"", true},
		{"a/#/b", true},   // # must be last
		{"a/b+c", true},   // wildcard mixed with chars
		{"a/b#c", true},   // wildcard mixed with chars
		{"a/+b", true},    // + mixed with chars
		{"a/#/c/d", true}, // # not last
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			err := ValidateMQTTPattern(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMQTTPattern(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
			}
		})
	}
}

func TestContainsMQTTWildcard(t *testing.T) {
	if !ContainsMQTTWildcard("a/+/c") {
		t.Error("should contain wildcard")
	}
	if !ContainsMQTTWildcard("a/#") {
		t.Error("should contain wildcard")
	}
	if ContainsMQTTWildcard("a/b/c") {
		t.Error("should not contain wildcard")
	}
}

// --- Unit tests for mqttPatternSubscribers ---

func TestMQTTPatternSubscribers_AddAndMatch(t *testing.T) {
	mps := newMQTTPatternSubscribers()

	sub1 := &subscriber{id: 1, topic: "devices/+/temp"}
	sub2 := &subscriber{id: 2, topic: "devices/#"}
	sub3 := &subscriber{id: 3, topic: "devices/sensor1/temp"}

	mps.Add("devices/+/temp", 1, sub1)
	mps.Add("devices/#", 2, sub2)
	mps.Add("devices/sensor1/temp", 3, sub3)

	// Match against a specific topic
	matched := mps.Match("devices/sensor1/temp")
	if len(matched) != 3 {
		t.Errorf("expected 3 matches, got %d", len(matched))
	}

	// Match against a topic that only matches #
	matched = mps.Match("devices/sensor1/humidity")
	if len(matched) != 1 {
		t.Errorf("expected 1 match (only #), got %d", len(matched))
	}
	if matched[0].id != 2 {
		t.Errorf("expected subscriber 2, got %d", matched[0].id)
	}

	// Match against a topic that matches nothing
	matched = mps.Match("other/topic")
	if len(matched) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matched))
	}
}

func TestMQTTPatternSubscribers_Remove(t *testing.T) {
	mps := newMQTTPatternSubscribers()

	sub1 := &subscriber{id: 1}
	sub2 := &subscriber{id: 2}

	mps.Add("a/+/c", 1, sub1)
	mps.Add("a/+/c", 2, sub2)

	if mps.Count("a/+/c") != 2 {
		t.Errorf("expected 2, got %d", mps.Count("a/+/c"))
	}

	mps.Remove("a/+/c", 1)
	if mps.Count("a/+/c") != 1 {
		t.Errorf("expected 1, got %d", mps.Count("a/+/c"))
	}

	mps.Remove("a/+/c", 2)
	if mps.Count("a/+/c") != 0 {
		t.Errorf("expected 0, got %d", mps.Count("a/+/c"))
	}

	// Pattern should be cleaned up
	if len(mps.patterns) != 0 {
		t.Error("expected patterns map to be empty after removing all subscribers")
	}
}

func TestMQTTPatternSubscribers_RemoveNonExistent(t *testing.T) {
	mps := newMQTTPatternSubscribers()
	if mps.Remove("nonexistent", 1) {
		t.Error("removing from non-existent pattern should return false")
	}
}

func TestMQTTPatternSubscribers_List(t *testing.T) {
	mps := newMQTTPatternSubscribers()

	mps.Add("a/+/c", 1, &subscriber{id: 1})
	mps.Add("b/#", 2, &subscriber{id: 2})
	mps.Add("c/d", 3, &subscriber{id: 3})

	patterns := mps.List()
	sort.Strings(patterns)

	if len(patterns) != 3 {
		t.Fatalf("expected 3 patterns, got %d", len(patterns))
	}
	expected := []string{"a/+/c", "b/#", "c/d"}
	for i, p := range patterns {
		if p != expected[i] {
			t.Errorf("pattern[%d] = %q, want %q", i, p, expected[i])
		}
	}
}

func TestMQTTPatternSubscribers_All(t *testing.T) {
	mps := newMQTTPatternSubscribers()

	mps.Add("a/+", 1, &subscriber{id: 1})
	mps.Add("b/#", 2, &subscriber{id: 2})
	mps.Add("a/+", 3, &subscriber{id: 3})

	all := mps.All()
	if len(all) != 3 {
		t.Errorf("expected 3 subscribers, got %d", len(all))
	}
}

func TestMQTTPatternSubscribers_Clear(t *testing.T) {
	mps := newMQTTPatternSubscribers()

	mps.Add("a/+", 1, &subscriber{id: 1})
	mps.Add("b/#", 2, &subscriber{id: 2})

	mps.Clear()

	if len(mps.patterns) != 0 {
		t.Error("expected empty patterns after Clear")
	}
	if len(mps.All()) != 0 {
		t.Error("expected no subscribers after Clear")
	}
}

// --- Integration tests: SubscribeMQTT with InProcPubSub ---

func TestSubscribeMQTT_BasicPublishReceive(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, err := ps.SubscribeMQTT("devices/+/temperature", SubOptions{
		BufferSize: 8,
		Policy:     DropOldest,
	})
	if err != nil {
		t.Fatalf("SubscribeMQTT: %v", err)
	}
	defer sub.Cancel()

	// Publish to a matching topic
	err = ps.Publish("devices/sensor1/temperature", Message{Data: "22.5"})
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}

	select {
	case msg := <-sub.C():
		if msg.Topic != "devices/sensor1/temperature" {
			t.Errorf("expected topic devices/sensor1/temperature, got %s", msg.Topic)
		}
		if msg.Data != "22.5" {
			t.Errorf("expected data 22.5, got %v", msg.Data)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestSubscribeMQTT_MultiLevelWildcard(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, err := ps.SubscribeMQTT("devices/#", DefaultSubOptions())
	if err != nil {
		t.Fatalf("SubscribeMQTT: %v", err)
	}
	defer sub.Cancel()

	topics := []string{
		"devices/sensor1/temperature",
		"devices/sensor1/humidity",
		"devices/sensor2/temperature",
		"devices",
	}

	for _, topic := range topics {
		err = ps.Publish(topic, Message{Data: topic})
		if err != nil {
			t.Fatalf("Publish(%s): %v", topic, err)
		}
	}

	received := make(map[string]bool)
	for i := 0; i < len(topics); i++ {
		select {
		case msg := <-sub.C():
			received[msg.Topic] = true
		case <-time.After(time.Second):
			t.Fatalf("timeout waiting for message %d", i)
		}
	}

	for _, topic := range topics {
		if !received[topic] {
			t.Errorf("did not receive message for topic %s", topic)
		}
	}
}

func TestSubscribeMQTT_NoMatchNoDelivery(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, err := ps.SubscribeMQTT("devices/+/temperature", DefaultSubOptions())
	if err != nil {
		t.Fatalf("SubscribeMQTT: %v", err)
	}
	defer sub.Cancel()

	// Publish to a non-matching topic
	err = ps.Publish("devices/sensor1/humidity", Message{Data: "65"})
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}

	select {
	case msg := <-sub.C():
		t.Fatalf("unexpected message: %v", msg)
	case <-time.After(100 * time.Millisecond):
		// Expected: no message received
	}
}

func TestSubscribeMQTT_InvalidPattern(t *testing.T) {
	ps := New()
	defer ps.Close()

	// Empty pattern
	_, err := ps.SubscribeMQTT("", DefaultSubOptions())
	if err != ErrInvalidPattern {
		t.Errorf("expected ErrInvalidPattern for empty, got %v", err)
	}

	// # not at end
	_, err = ps.SubscribeMQTT("a/#/b", DefaultSubOptions())
	if err == nil {
		t.Error("expected error for invalid MQTT pattern a/#/b")
	}

	// Wildcard mixed with characters
	_, err = ps.SubscribeMQTT("a/b+c", DefaultSubOptions())
	if err == nil {
		t.Error("expected error for invalid MQTT pattern a/b+c")
	}
}

func TestSubscribeMQTT_ClosedPubSub(t *testing.T) {
	ps := New()
	ps.Close()

	_, err := ps.SubscribeMQTT("a/+/b", DefaultSubOptions())
	if err != ErrSubscribeToClosed {
		t.Errorf("expected ErrSubscribeToClosed, got %v", err)
	}
}

func TestSubscribeMQTT_Cancel(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, err := ps.SubscribeMQTT("devices/+/temp", SubOptions{
		BufferSize: 4,
		Policy:     DropOldest,
	})
	if err != nil {
		t.Fatalf("SubscribeMQTT: %v", err)
	}

	if !ps.MQTTPatternExists("devices/+/temp") {
		t.Error("MQTT pattern should exist before cancel")
	}

	sub.Cancel()

	// After cancel, channel should be closed
	<-sub.Done()

	if ps.MQTTPatternExists("devices/+/temp") {
		t.Error("MQTT pattern should not exist after cancel")
	}
}

func TestSubscribeMQTT_WithContext(t *testing.T) {
	ps := New()
	defer ps.Close()

	ctx, cancel := context.WithCancel(context.Background())
	sub, err := ps.SubscribeMQTTWithContext(ctx, "devices/#", DefaultSubOptions())
	if err != nil {
		t.Fatalf("SubscribeMQTTWithContext: %v", err)
	}

	cancel() // Cancel the context

	// Wait for subscription to be cancelled
	select {
	case <-sub.Done():
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for subscription cancellation")
	}
}

func TestSubscribeMQTT_CoexistsWithGlobPatterns(t *testing.T) {
	ps := New()
	defer ps.Close()

	// MQTT pattern subscriber
	mqttSub, err := ps.SubscribeMQTT("events/+/created", SubOptions{
		BufferSize: 8,
		Policy:     DropOldest,
	})
	if err != nil {
		t.Fatalf("SubscribeMQTT: %v", err)
	}
	defer mqttSub.Cancel()

	// Glob pattern subscriber (using filepath matching)
	globSub, err := ps.SubscribePattern("events.*", SubOptions{
		BufferSize: 8,
		Policy:     DropOldest,
	})
	if err != nil {
		t.Fatalf("SubscribePattern: %v", err)
	}
	defer globSub.Cancel()

	// Exact topic subscriber
	exactSub, err := ps.Subscribe("events/user/created", SubOptions{
		BufferSize: 8,
		Policy:     DropOldest,
	})
	if err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	defer exactSub.Cancel()

	// Publish to a topic that matches MQTT and exact
	err = ps.Publish("events/user/created", Message{Data: "test"})
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}

	// MQTT subscriber should receive it
	select {
	case msg := <-mqttSub.C():
		if msg.Data != "test" {
			t.Errorf("mqtt sub: expected data 'test', got %v", msg.Data)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for MQTT subscriber message")
	}

	// Exact subscriber should receive it
	select {
	case msg := <-exactSub.C():
		if msg.Data != "test" {
			t.Errorf("exact sub: expected data 'test', got %v", msg.Data)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for exact subscriber message")
	}

	// Glob subscriber should NOT receive it (different naming convention)
	select {
	case msg := <-globSub.C():
		t.Fatalf("glob sub should not receive message (different naming), got: %v", msg)
	case <-time.After(100 * time.Millisecond):
		// Expected
	}
}

func TestSubscribeMQTT_MultipleSubscribers(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub1, err := ps.SubscribeMQTT("devices/+/temperature", DefaultSubOptions())
	if err != nil {
		t.Fatalf("SubscribeMQTT 1: %v", err)
	}
	defer sub1.Cancel()

	sub2, err := ps.SubscribeMQTT("devices/+/temperature", DefaultSubOptions())
	if err != nil {
		t.Fatalf("SubscribeMQTT 2: %v", err)
	}
	defer sub2.Cancel()

	err = ps.Publish("devices/sensor1/temperature", Message{Data: "25"})
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}

	// Both subscribers should receive the message
	for i, sub := range []Subscription{sub1, sub2} {
		select {
		case msg := <-sub.C():
			if msg.Data != "25" {
				t.Errorf("sub%d: expected data '25', got %v", i+1, msg.Data)
			}
		case <-time.After(time.Second):
			t.Fatalf("timeout waiting for sub%d message", i+1)
		}
	}
}

func TestSubscribeMQTT_HashMatchesZeroLevels(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, err := ps.SubscribeMQTT("devices/#", DefaultSubOptions())
	if err != nil {
		t.Fatalf("SubscribeMQTT: %v", err)
	}
	defer sub.Cancel()

	// "devices/#" should match "devices" (zero levels after)
	err = ps.Publish("devices", Message{Data: "root"})
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}

	select {
	case msg := <-sub.C():
		if msg.Data != "root" {
			t.Errorf("expected data 'root', got %v", msg.Data)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout: devices/# should match devices")
	}
}

func TestSubscribeMQTT_PublishBatch(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, err := ps.SubscribeMQTT("orders/+/status", SubOptions{
		BufferSize: 16,
		Policy:     DropOldest,
	})
	if err != nil {
		t.Fatalf("SubscribeMQTT: %v", err)
	}
	defer sub.Cancel()

	msgs := []Message{
		{Data: "pending"},
		{Data: "shipped"},
		{Data: "delivered"},
	}

	err = ps.PublishBatch("orders/123/status", msgs)
	if err != nil {
		t.Fatalf("PublishBatch: %v", err)
	}

	for i := 0; i < 3; i++ {
		select {
		case msg := <-sub.C():
			if msg.Topic != "orders/123/status" {
				t.Errorf("msg %d: unexpected topic %s", i, msg.Topic)
			}
		case <-time.After(time.Second):
			t.Fatalf("timeout waiting for batch message %d", i)
		}
	}
}

func TestSubscribeMQTT_HasSubscribers(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, err := ps.SubscribeMQTT("home/+/temperature", DefaultSubOptions())
	if err != nil {
		t.Fatalf("SubscribeMQTT: %v", err)
	}

	if !ps.HasSubscribers("home/living/temperature") {
		t.Error("HasSubscribers should return true for matching MQTT pattern")
	}

	if ps.HasSubscribers("home/living/humidity") {
		t.Error("HasSubscribers should return false for non-matching topic")
	}

	sub.Cancel()

	// After cancel, pattern should be removed
	if ps.HasSubscribers("home/living/temperature") {
		t.Error("HasSubscribers should return false after cancel")
	}
}

func TestSubscribeMQTT_ListAndCount(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub1, err := ps.SubscribeMQTT("a/+/c", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub1.Cancel()

	sub2, err := ps.SubscribeMQTT("b/#", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub2.Cancel()

	patterns := ps.ListMQTTPatterns()
	sort.Strings(patterns)
	if len(patterns) != 2 {
		t.Fatalf("expected 2 MQTT patterns, got %d", len(patterns))
	}
	if patterns[0] != "a/+/c" || patterns[1] != "b/#" {
		t.Errorf("unexpected patterns: %v", patterns)
	}

	if ps.GetMQTTSubscriberCount("a/+/c") != 1 {
		t.Errorf("expected 1 subscriber for a/+/c, got %d", ps.GetMQTTSubscriberCount("a/+/c"))
	}

	if !ps.MQTTPatternExists("b/#") {
		t.Error("MQTTPatternExists should return true")
	}

	if ps.MQTTPatternExists("c/d") {
		t.Error("MQTTPatternExists should return false for non-existent pattern")
	}
}

func TestSubscribeMQTT_SubscriptionProperties(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, err := ps.SubscribeMQTT("a/+/c", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	if sub.Topic() != "a/+/c" {
		t.Errorf("expected topic a/+/c, got %s", sub.Topic())
	}

	if sub.ID() == 0 {
		t.Error("expected non-zero subscription ID")
	}

	stats := sub.Stats()
	if stats.Received != 0 {
		t.Errorf("expected 0 received, got %d", stats.Received)
	}
}

func TestSubscribeMQTT_ConcurrentPublish(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, err := ps.SubscribeMQTT("events/+/log", SubOptions{
		BufferSize: 256,
		Policy:     DropOldest,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	const numPublishers = 10
	const msgsPerPublisher = 20

	var wg sync.WaitGroup
	wg.Add(numPublishers)

	for i := 0; i < numPublishers; i++ {
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < msgsPerPublisher; j++ {
				ps.Publish("events/service/log", Message{Data: idx*msgsPerPublisher + j})
			}
		}(i)
	}

	wg.Wait()

	// Drain messages
	count := 0
	timeout := time.After(2 * time.Second)
	for {
		select {
		case <-sub.C():
			count++
			if count == numPublishers*msgsPerPublisher {
				return // All received
			}
		case <-timeout:
			// With DropOldest some may be dropped under contention, but we should
			// receive at least a significant portion
			if count == 0 {
				t.Fatal("received no messages from concurrent publishers")
			}
			return
		}
	}
}

func TestSubscribeMQTT_CloseCleanup(t *testing.T) {
	ps := New()

	sub1, err := ps.SubscribeMQTT("a/+/b", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}

	sub2, err := ps.SubscribeMQTT("c/#", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}

	ps.Close()

	// After Close, subscriptions should be cancelled
	select {
	case <-sub1.Done():
	case <-time.After(time.Second):
		t.Fatal("sub1 not cancelled after Close")
	}

	select {
	case <-sub2.Done():
	case <-time.After(time.Second):
		t.Fatal("sub2 not cancelled after Close")
	}
}

func TestSubscribeMQTT_Hooks(t *testing.T) {
	var subscribedTopic string
	var subscribedID uint64
	var unsubscribedTopic string
	var deliveredTopic string

	ps := New(WithHooks(Hooks{
		OnSubscribe: func(topic string, subID uint64) {
			subscribedTopic = topic
			subscribedID = subID
		},
		OnUnsubscribe: func(topic string, subID uint64) {
			unsubscribedTopic = topic
		},
		OnDeliver: func(topic string, subID uint64, msg *Message) {
			deliveredTopic = topic
		},
	}))
	defer ps.Close()

	sub, err := ps.SubscribeMQTT("devices/+/temp", DefaultSubOptions())
	if err != nil {
		t.Fatal(err)
	}

	if subscribedTopic != "devices/+/temp" {
		t.Errorf("OnSubscribe topic = %q, want devices/+/temp", subscribedTopic)
	}
	if subscribedID != sub.ID() {
		t.Errorf("OnSubscribe ID = %d, want %d", subscribedID, sub.ID())
	}

	// Publish and verify OnDeliver gets the actual topic, not the pattern
	ps.Publish("devices/sensor1/temp", Message{Data: "20"})

	select {
	case <-sub.C():
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}

	if deliveredTopic != "devices/sensor1/temp" {
		t.Errorf("OnDeliver topic = %q, want devices/sensor1/temp", deliveredTopic)
	}

	sub.Cancel()

	if unsubscribedTopic != "devices/+/temp" {
		t.Errorf("OnUnsubscribe topic = %q, want devices/+/temp", unsubscribedTopic)
	}
}

func TestSubscribeMQTT_Filter(t *testing.T) {
	ps := New()
	defer ps.Close()

	sub, err := ps.SubscribeMQTT("sensor/+/data", SubOptions{
		BufferSize: 8,
		Policy:     DropOldest,
		Filter: func(msg Message) bool {
			// Only accept messages with string data starting with "temp:"
			s, ok := msg.Data.(string)
			return ok && len(s) > 5 && s[:5] == "temp:"
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Cancel()

	ps.Publish("sensor/s1/data", Message{Data: "temp:22.5"})
	ps.Publish("sensor/s1/data", Message{Data: "humidity:65"})
	ps.Publish("sensor/s2/data", Message{Data: "temp:18.0"})

	received := 0
	timeout := time.After(500 * time.Millisecond)
loop:
	for {
		select {
		case <-sub.C():
			received++
		case <-timeout:
			break loop
		}
	}

	if received != 2 {
		t.Errorf("expected 2 filtered messages, got %d", received)
	}
}
