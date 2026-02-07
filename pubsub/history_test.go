package pubsub

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// --- Unit tests for messageHistory (circular buffer) ---

func TestMessageHistory_AddAndGetAll(t *testing.T) {
	h := newMessageHistory(5)

	for i := 1; i <= 3; i++ {
		h.Add(Message{ID: fmt.Sprintf("m%d", i)})
	}

	msgs := h.GetAll()
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}
	for i, msg := range msgs {
		expected := fmt.Sprintf("m%d", i+1)
		if msg.ID != expected {
			t.Fatalf("msgs[%d].ID = %s, want %s", i, msg.ID, expected)
		}
	}
}

func TestMessageHistory_CircularEviction(t *testing.T) {
	h := newMessageHistory(3)

	for i := 1; i <= 5; i++ {
		h.Add(Message{ID: fmt.Sprintf("m%d", i)})
	}

	if h.Len() != 3 {
		t.Fatalf("expected 3 messages after eviction, got %d", h.Len())
	}

	msgs := h.GetAll()
	// Should retain m3, m4, m5 (oldest two evicted)
	expected := []string{"m3", "m4", "m5"}
	for i, msg := range msgs {
		if msg.ID != expected[i] {
			t.Fatalf("msgs[%d].ID = %s, want %s", i, msg.ID, expected[i])
		}
	}
}

func TestMessageHistory_GetLast(t *testing.T) {
	h := newMessageHistory(10)

	for i := 1; i <= 5; i++ {
		h.Add(Message{ID: fmt.Sprintf("m%d", i)})
	}

	msgs := h.GetLast(3)
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}
	expected := []string{"m3", "m4", "m5"}
	for i, msg := range msgs {
		if msg.ID != expected[i] {
			t.Fatalf("msgs[%d].ID = %s, want %s", i, msg.ID, expected[i])
		}
	}

	// Request more than available
	msgs = h.GetLast(20)
	if len(msgs) != 5 {
		t.Fatalf("expected 5 messages, got %d", len(msgs))
	}

	// Zero and negative
	if msgs := h.GetLast(0); msgs != nil {
		t.Fatalf("expected nil for n=0, got %v", msgs)
	}
	if msgs := h.GetLast(-1); msgs != nil {
		t.Fatalf("expected nil for n=-1, got %v", msgs)
	}
}

func TestMessageHistory_GetSince(t *testing.T) {
	h := newMessageHistory(10)

	for i := 1; i <= 5; i++ {
		h.Add(Message{ID: fmt.Sprintf("m%d", i)})
	}

	// Get messages since sequence 3 (should return m4, m5)
	msgs := h.GetSince(3)
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages since seq 3, got %d", len(msgs))
	}
	if msgs[0].ID != "m4" || msgs[1].ID != "m5" {
		t.Fatalf("unexpected messages: %v", msgs)
	}

	// Get since 0 should return all
	msgs = h.GetSince(0)
	if len(msgs) != 5 {
		t.Fatalf("expected 5 messages since seq 0, got %d", len(msgs))
	}

	// Get since current sequence returns nothing
	msgs = h.GetSince(5)
	if len(msgs) != 0 {
		t.Fatalf("expected 0 messages since current seq, got %d", len(msgs))
	}
}

func TestMessageHistory_GetWithTTL(t *testing.T) {
	h := newMessageHistory(10)

	// Add a message
	h.Add(Message{ID: "m1"})

	// All within TTL
	msgs := h.GetWithTTL(1 * time.Second)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message within TTL, got %d", len(msgs))
	}

	// Zero TTL should return nothing (cutoff is now)
	msgs = h.GetWithTTL(0)
	if len(msgs) != 0 {
		t.Fatalf("expected 0 messages with zero TTL, got %d", len(msgs))
	}
}

func TestMessageHistory_Sequence(t *testing.T) {
	h := newMessageHistory(5)

	if h.CurrentSequence() != 0 {
		t.Fatalf("expected initial sequence 0, got %d", h.CurrentSequence())
	}

	seq1 := h.Add(Message{ID: "m1"})
	seq2 := h.Add(Message{ID: "m2"})

	if seq1 != 1 || seq2 != 2 {
		t.Fatalf("expected sequences 1,2; got %d,%d", seq1, seq2)
	}

	if h.CurrentSequence() != 2 {
		t.Fatalf("expected current sequence 2, got %d", h.CurrentSequence())
	}
}

func TestMessageHistory_Clear(t *testing.T) {
	h := newMessageHistory(5)

	h.Add(Message{ID: "m1"})
	h.Add(Message{ID: "m2"})
	h.Clear()

	if h.Len() != 0 {
		t.Fatalf("expected 0 messages after clear, got %d", h.Len())
	}
	if msgs := h.GetAll(); msgs != nil {
		t.Fatalf("expected nil after clear, got %v", msgs)
	}
}

func TestMessageHistory_NilOnZeroCapacity(t *testing.T) {
	h := newMessageHistory(0)
	if h != nil {
		t.Fatalf("expected nil for zero capacity")
	}

	h = newMessageHistory(-1)
	if h != nil {
		t.Fatalf("expected nil for negative capacity")
	}
}

func TestMessageHistory_EmptyGetters(t *testing.T) {
	h := newMessageHistory(5)

	if msgs := h.GetAll(); msgs != nil {
		t.Fatalf("expected nil from empty GetAll")
	}
	if msgs := h.GetSince(0); msgs != nil {
		t.Fatalf("expected nil from empty GetSince")
	}
	if msgs := h.GetLast(1); msgs != nil {
		t.Fatalf("expected nil from empty GetLast")
	}
	if msgs := h.GetWithTTL(time.Second); msgs != nil {
		t.Fatalf("expected nil from empty GetWithTTL")
	}
}

// --- Unit tests for topicHistory ---

func TestTopicHistory_GetOrCreate(t *testing.T) {
	th := newTopicHistory(DefaultHistoryConfig())

	h1 := th.GetOrCreate("topic.a", 0)
	if h1 == nil {
		t.Fatal("expected non-nil history")
	}
	if h1.Cap() != 100 { // DefaultRetention
		t.Fatalf("expected capacity 100, got %d", h1.Cap())
	}

	// Same topic returns same instance
	h2 := th.GetOrCreate("topic.a", 0)
	if h1 != h2 {
		t.Fatal("expected same history instance for same topic")
	}
}

func TestTopicHistory_CustomRetention(t *testing.T) {
	th := newTopicHistory(HistoryConfig{
		DefaultRetention: 50,
		MaxRetention:     200,
	})

	h := th.GetOrCreate("t1", 150)
	if h.Cap() != 150 {
		t.Fatalf("expected capacity 150, got %d", h.Cap())
	}

	// Exceeds max -> clamped
	h2 := th.GetOrCreate("t2", 500)
	if h2.Cap() != 200 {
		t.Fatalf("expected capacity clamped to 200, got %d", h2.Cap())
	}
}

func TestTopicHistory_Delete(t *testing.T) {
	th := newTopicHistory(DefaultHistoryConfig())
	th.GetOrCreate("topic.a", 0)

	th.Delete("topic.a")

	if h := th.Get("topic.a"); h != nil {
		t.Fatal("expected nil after delete")
	}
}

func TestTopicHistory_Topics(t *testing.T) {
	th := newTopicHistory(DefaultHistoryConfig())
	th.GetOrCreate("a", 0)
	th.GetOrCreate("b", 0)
	th.GetOrCreate("c", 0)

	topics := th.Topics()
	if len(topics) != 3 {
		t.Fatalf("expected 3 topics, got %d", len(topics))
	}
}

func TestTopicHistory_Stats(t *testing.T) {
	th := newTopicHistory(DefaultHistoryConfig())
	h := th.GetOrCreate("topic.a", 10)
	h.Add(Message{ID: "m1"})
	h.Add(Message{ID: "m2"})

	stats := th.Stats()
	s, ok := stats["topic.a"]
	if !ok {
		t.Fatal("expected stats for topic.a")
	}
	if s.Count != 2 || s.Capacity != 10 || s.Sequence != 2 {
		t.Fatalf("unexpected stats: %+v", s)
	}
}

// --- Integration tests: InProcPubSub with history ---

func TestPubSub_HistoryDisabledByDefault(t *testing.T) {
	ps := New()
	defer ps.Close()

	_, err := ps.GetTopicHistory("any")
	if err != ErrHistoryDisabled {
		t.Fatalf("expected ErrHistoryDisabled, got %v", err)
	}

	_, err = ps.GetTopicHistorySince("any", 0)
	if err != ErrHistoryDisabled {
		t.Fatalf("expected ErrHistoryDisabled, got %v", err)
	}

	_, err = ps.GetRecentMessages("any", 5)
	if err != ErrHistoryDisabled {
		t.Fatalf("expected ErrHistoryDisabled, got %v", err)
	}

	_, err = ps.GetTopicHistoryByTTL("any", time.Second)
	if err != ErrHistoryDisabled {
		t.Fatalf("expected ErrHistoryDisabled, got %v", err)
	}

	err = ps.ClearTopicHistory("any")
	if err != ErrHistoryDisabled {
		t.Fatalf("expected ErrHistoryDisabled, got %v", err)
	}

	_, err = ps.TopicHistoryStats()
	if err != ErrHistoryDisabled {
		t.Fatalf("expected ErrHistoryDisabled, got %v", err)
	}

	_, err = ps.TopicHistorySequence("any")
	if err != ErrHistoryDisabled {
		t.Fatalf("expected ErrHistoryDisabled, got %v", err)
	}
}

func TestPubSub_HistoryRecordsOnPublish(t *testing.T) {
	ps := New(WithHistory())
	defer ps.Close()

	for i := 1; i <= 5; i++ {
		if err := ps.Publish("events", Message{ID: fmt.Sprintf("m%d", i)}); err != nil {
			t.Fatalf("publish: %v", err)
		}
	}

	msgs, err := ps.GetTopicHistory("events")
	if err != nil {
		t.Fatalf("GetTopicHistory: %v", err)
	}
	if len(msgs) != 5 {
		t.Fatalf("expected 5 messages in history, got %d", len(msgs))
	}
	for i, msg := range msgs {
		expected := fmt.Sprintf("m%d", i+1)
		if msg.ID != expected {
			t.Fatalf("msgs[%d].ID = %s, want %s", i, msg.ID, expected)
		}
	}
}

func TestPubSub_HistoryRecordsOnPublishBatch(t *testing.T) {
	ps := New(WithHistory())
	defer ps.Close()

	msgs := []Message{
		{ID: "b1"},
		{ID: "b2"},
		{ID: "b3"},
	}
	if err := ps.PublishBatch("batch.topic", msgs); err != nil {
		t.Fatalf("PublishBatch: %v", err)
	}

	history, err := ps.GetTopicHistory("batch.topic")
	if err != nil {
		t.Fatalf("GetTopicHistory: %v", err)
	}
	if len(history) != 3 {
		t.Fatalf("expected 3 messages in history, got %d", len(history))
	}
	expected := []string{"b1", "b2", "b3"}
	for i, msg := range history {
		if msg.ID != expected[i] {
			t.Fatalf("history[%d].ID = %s, want %s", i, msg.ID, expected[i])
		}
	}
}

func TestPubSub_HistoryGetRecentMessages(t *testing.T) {
	ps := New(WithHistory())
	defer ps.Close()

	for i := 1; i <= 10; i++ {
		ps.Publish("topic", Message{ID: fmt.Sprintf("m%d", i)})
	}

	msgs, err := ps.GetRecentMessages("topic", 3)
	if err != nil {
		t.Fatalf("GetRecentMessages: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}
	expected := []string{"m8", "m9", "m10"}
	for i, msg := range msgs {
		if msg.ID != expected[i] {
			t.Fatalf("msgs[%d].ID = %s, want %s", i, msg.ID, expected[i])
		}
	}
}

func TestPubSub_HistoryGetSince(t *testing.T) {
	ps := New(WithHistory())
	defer ps.Close()

	for i := 1; i <= 5; i++ {
		ps.Publish("topic", Message{ID: fmt.Sprintf("m%d", i)})
	}

	seq, err := ps.TopicHistorySequence("topic")
	if err != nil {
		t.Fatalf("TopicHistorySequence: %v", err)
	}
	if seq != 5 {
		t.Fatalf("expected sequence 5, got %d", seq)
	}

	// Get messages since sequence 3
	msgs, err := ps.GetTopicHistorySince("topic", 3)
	if err != nil {
		t.Fatalf("GetTopicHistorySince: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].ID != "m4" || msgs[1].ID != "m5" {
		t.Fatalf("unexpected messages: %v, %v", msgs[0].ID, msgs[1].ID)
	}
}

func TestPubSub_HistoryGetByTTL(t *testing.T) {
	ps := New(WithHistory())
	defer ps.Close()

	ps.Publish("topic", Message{ID: "m1"})

	msgs, err := ps.GetTopicHistoryByTTL("topic", 1*time.Second)
	if err != nil {
		t.Fatalf("GetTopicHistoryByTTL: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message within TTL, got %d", len(msgs))
	}
}

func TestPubSub_HistoryClear(t *testing.T) {
	ps := New(WithHistory())
	defer ps.Close()

	ps.Publish("topic", Message{ID: "m1"})

	if err := ps.ClearTopicHistory("topic"); err != nil {
		t.Fatalf("ClearTopicHistory: %v", err)
	}

	msgs, err := ps.GetTopicHistory("topic")
	if err != nil {
		t.Fatalf("GetTopicHistory: %v", err)
	}
	if msgs != nil {
		t.Fatalf("expected nil after clear, got %v", msgs)
	}
}

func TestPubSub_HistoryStats(t *testing.T) {
	ps := New(WithHistory())
	defer ps.Close()

	ps.Publish("a", Message{ID: "m1"})
	ps.Publish("a", Message{ID: "m2"})
	ps.Publish("b", Message{ID: "m3"})

	stats, err := ps.TopicHistoryStats()
	if err != nil {
		t.Fatalf("TopicHistoryStats: %v", err)
	}

	if len(stats) != 2 {
		t.Fatalf("expected stats for 2 topics, got %d", len(stats))
	}
	if stats["a"].Count != 2 {
		t.Fatalf("expected 2 messages for topic a, got %d", stats["a"].Count)
	}
	if stats["b"].Count != 1 {
		t.Fatalf("expected 1 message for topic b, got %d", stats["b"].Count)
	}
}

func TestPubSub_HistoryNonexistentTopic(t *testing.T) {
	ps := New(WithHistory())
	defer ps.Close()

	msgs, err := ps.GetTopicHistory("nonexistent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if msgs != nil {
		t.Fatalf("expected nil for nonexistent topic, got %v", msgs)
	}

	seq, err := ps.TopicHistorySequence("nonexistent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if seq != 0 {
		t.Fatalf("expected sequence 0, got %d", seq)
	}
}

func TestPubSub_HistoryWithCustomConfig(t *testing.T) {
	ps := New(WithHistory(HistoryConfig{
		DefaultRetention: 3,
		MaxRetention:     5,
	}))
	defer ps.Close()

	for i := 1; i <= 10; i++ {
		ps.Publish("topic", Message{ID: fmt.Sprintf("m%d", i)})
	}

	msgs, err := ps.GetTopicHistory("topic")
	if err != nil {
		t.Fatalf("GetTopicHistory: %v", err)
	}
	// Only 3 retained (DefaultRetention = 3)
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages (retention limit), got %d", len(msgs))
	}
	expected := []string{"m8", "m9", "m10"}
	for i, msg := range msgs {
		if msg.ID != expected[i] {
			t.Fatalf("msgs[%d].ID = %s, want %s", i, msg.ID, expected[i])
		}
	}
}

func TestPubSub_HistoryMultipleTopics(t *testing.T) {
	ps := New(WithHistory())
	defer ps.Close()

	ps.Publish("users", Message{ID: "u1"})
	ps.Publish("users", Message{ID: "u2"})
	ps.Publish("orders", Message{ID: "o1"})

	users, _ := ps.GetTopicHistory("users")
	orders, _ := ps.GetTopicHistory("orders")

	if len(users) != 2 {
		t.Fatalf("expected 2 user messages, got %d", len(users))
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 order message, got %d", len(orders))
	}
}

func TestPubSub_HistoryDoesNotAffectDelivery(t *testing.T) {
	ps := New(WithHistory())
	defer ps.Close()

	sub, err := ps.Subscribe("topic", SubOptions{BufferSize: 8, Policy: DropOldest})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	defer sub.Cancel()

	ps.Publish("topic", Message{ID: "m1"})

	// Subscriber still receives the message
	select {
	case msg := <-sub.C():
		if msg.ID != "m1" {
			t.Fatalf("expected m1, got %s", msg.ID)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("subscriber timeout")
	}

	// History also has it
	msgs, _ := ps.GetTopicHistory("topic")
	if len(msgs) != 1 || msgs[0].ID != "m1" {
		t.Fatalf("expected m1 in history, got %v", msgs)
	}
}

func TestPubSub_HistoryConcurrentAccess(t *testing.T) {
	ps := New(WithHistory(HistoryConfig{
		DefaultRetention: 1000,
		MaxRetention:     1000,
	}))
	defer ps.Close()

	var wg sync.WaitGroup
	numWriters := 10
	msgsPerWriter := 100

	for w := 0; w < numWriters; w++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for i := 0; i < msgsPerWriter; i++ {
				ps.Publish("concurrent", Message{
					ID: fmt.Sprintf("w%d-m%d", writerID, i),
				})
			}
		}(w)
	}

	// Concurrent readers
	for r := 0; r < 5; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				ps.GetTopicHistory("concurrent")
				ps.GetRecentMessages("concurrent", 10)
				ps.TopicHistoryStats()
			}
		}()
	}

	wg.Wait()

	msgs, err := ps.GetTopicHistory("concurrent")
	if err != nil {
		t.Fatalf("GetTopicHistory: %v", err)
	}
	if len(msgs) != numWriters*msgsPerWriter {
		t.Fatalf("expected %d messages, got %d", numWriters*msgsPerWriter, len(msgs))
	}
}

func TestPubSub_HistoryInterfaceCompliance(t *testing.T) {
	ps := New(WithHistory())
	defer ps.Close()

	// Ensure InProcPubSub satisfies HistoryPubSub interface
	var _ HistoryPubSub = ps
}
