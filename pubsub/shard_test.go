package pubsub

import (
	"testing"
)

func TestGetShardIndex_Deterministic(t *testing.T) {
	sm := newShardedMap(16)

	// Same key must always return the same shard index.
	idx1 := sm.getShardIndex("user.created")
	idx2 := sm.getShardIndex("user.created")
	if idx1 != idx2 {
		t.Fatalf("getShardIndex not deterministic: %d != %d", idx1, idx2)
	}
}

func TestGetShardIndex_WithinBounds(t *testing.T) {
	for _, shardCount := range []int{1, 2, 4, 8, 16, 32, 64} {
		sm := newShardedMap(shardCount)
		topics := []string{
			"user.created", "user.updated", "order.placed",
			"payment.completed", "notification.sent", "system.health",
			"", "a", "very.long.topic.name.with.many.segments",
		}
		for _, topic := range topics {
			idx := sm.getShardIndex(topic)
			if idx < 0 || idx >= sm.shardCount {
				t.Fatalf("shardCount=%d topic=%q: index %d out of bounds [0, %d)",
					shardCount, topic, idx, sm.shardCount)
			}
		}
	}
}

func TestGetShardIndex_ConsistentWithGetShard(t *testing.T) {
	sm := newShardedMap(16)

	topics := []string{"a", "b", "user.created", "order.*", "system.health.check"}
	for _, topic := range topics {
		idx := sm.getShardIndex(topic)
		shard := sm.getShard(topic)
		if sm.shards[idx] != shard {
			t.Fatalf("topic=%q: getShardIndex(%d) does not match getShard", topic, idx)
		}
	}
}

func TestGetShardIndex_Distribution(t *testing.T) {
	sm := newShardedMap(16)
	counts := make(map[int]int)

	// Hash 100 different topics and verify they spread across shards.
	for i := 0; i < 100; i++ {
		topic := "topic." + string(rune('a'+i%26)) + "." + string(rune('0'+i%10))
		idx := sm.getShardIndex(topic)
		counts[idx]++
	}

	// With 100 topics across 16 shards, at least 4 shards should be used.
	if len(counts) < 4 {
		t.Fatalf("poor distribution: only %d shards used out of 16", len(counts))
	}
}

func TestShardStats_Empty(t *testing.T) {
	sm := newShardedMap(8)
	stats := sm.shardStats()

	if len(stats) != 8 {
		t.Fatalf("expected 8 shard stats, got %d", len(stats))
	}

	for i, s := range stats {
		if s.Index != i {
			t.Fatalf("shard %d: expected index %d, got %d", i, i, s.Index)
		}
		if s.TopicCount != 0 || s.PatternCount != 0 || s.SubscriberCount != 0 {
			t.Fatalf("shard %d: expected all zeros, got topics=%d patterns=%d subs=%d",
				i, s.TopicCount, s.PatternCount, s.SubscriberCount)
		}
	}
}

func TestShardStats_WithSubscribers(t *testing.T) {
	ps := New(WithShardCount(16))
	defer ps.Close()

	topics := []string{"user.created", "order.placed", "payment.done"}
	for _, topic := range topics {
		_, err := ps.Subscribe(topic, SubOptions{BufferSize: 4, Policy: DropOldest})
		if err != nil {
			t.Fatalf("subscribe %s: %v", topic, err)
		}
	}

	stats := ps.ShardStats()
	if len(stats) != 16 {
		t.Fatalf("expected 16 shard stats, got %d", len(stats))
	}

	totalTopics := 0
	totalSubs := 0
	for _, s := range stats {
		totalTopics += s.TopicCount
		totalSubs += s.SubscriberCount
	}

	if totalTopics != 3 {
		t.Fatalf("expected 3 total topics, got %d", totalTopics)
	}
	if totalSubs != 3 {
		t.Fatalf("expected 3 total subscribers, got %d", totalSubs)
	}
}

func TestShardStats_WithPatterns(t *testing.T) {
	ps := New(WithShardCount(8))
	defer ps.Close()

	_, err := ps.Subscribe("exact.topic", SubOptions{BufferSize: 4, Policy: DropOldest})
	if err != nil {
		t.Fatalf("subscribe topic: %v", err)
	}
	_, err = ps.SubscribePattern("user.*", SubOptions{BufferSize: 4, Policy: DropOldest})
	if err != nil {
		t.Fatalf("subscribe pattern: %v", err)
	}

	stats := ps.ShardStats()

	totalTopics := 0
	totalPatterns := 0
	totalSubs := 0
	for _, s := range stats {
		totalTopics += s.TopicCount
		totalPatterns += s.PatternCount
		totalSubs += s.SubscriberCount
	}

	if totalTopics != 1 {
		t.Fatalf("expected 1 topic, got %d", totalTopics)
	}
	if totalPatterns != 1 {
		t.Fatalf("expected 1 pattern, got %d", totalPatterns)
	}
	if totalSubs != 2 {
		t.Fatalf("expected 2 subscribers, got %d", totalSubs)
	}
}

func TestTopicShard_MatchesInternalIndex(t *testing.T) {
	ps := New(WithShardCount(16))
	defer ps.Close()

	topics := []string{"user.created", "order.placed", "system.health"}
	for _, topic := range topics {
		idx := ps.TopicShard(topic)
		if idx < 0 || idx >= 16 {
			t.Fatalf("topic=%q: shard index %d out of bounds", topic, idx)
		}
		// Verify consistency: same topic always maps to same shard.
		if ps.TopicShard(topic) != idx {
			t.Fatalf("topic=%q: non-deterministic shard index", topic)
		}
	}
}

func TestTopicShardMapping_Empty(t *testing.T) {
	ps := New()
	defer ps.Close()

	mapping := ps.TopicShardMapping()
	if len(mapping) != 0 {
		t.Fatalf("expected empty mapping, got %d entries", len(mapping))
	}
}

func TestTopicShardMapping_WithSubscribers(t *testing.T) {
	ps := New(WithShardCount(16))
	defer ps.Close()

	topics := []string{"user.created", "order.placed", "payment.done"}
	for _, topic := range topics {
		_, err := ps.Subscribe(topic, SubOptions{BufferSize: 4, Policy: DropOldest})
		if err != nil {
			t.Fatalf("subscribe %s: %v", topic, err)
		}
	}

	_, err := ps.SubscribePattern("event.*", SubOptions{BufferSize: 4, Policy: DropOldest})
	if err != nil {
		t.Fatalf("subscribe pattern: %v", err)
	}

	mapping := ps.TopicShardMapping()

	// 3 topics + 1 pattern = 4 entries
	if len(mapping) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(mapping))
	}

	// Verify each mapping is consistent with TopicShard.
	for key, idx := range mapping {
		expected := ps.TopicShard(key)
		if idx != expected {
			t.Fatalf("key=%q: mapping says %d, TopicShard says %d", key, idx, expected)
		}
	}
}

func TestTopicShardMapping_AfterUnsubscribe(t *testing.T) {
	ps := New(WithShardCount(8))
	defer ps.Close()

	sub, err := ps.Subscribe("temp.topic", SubOptions{BufferSize: 4, Policy: DropOldest})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	mapping := ps.TopicShardMapping()
	if _, ok := mapping["temp.topic"]; !ok {
		t.Fatal("expected temp.topic in mapping")
	}

	sub.Cancel()

	mapping = ps.TopicShardMapping()
	if _, ok := mapping["temp.topic"]; ok {
		t.Fatal("expected temp.topic removed from mapping after cancel")
	}
}

func TestDiagnosticInfo_ContainsShardData(t *testing.T) {
	ps := New(WithShardCount(8))
	defer ps.Close()

	_, err := ps.Subscribe("diag.topic", SubOptions{BufferSize: 4, Policy: DropOldest})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	info := ps.DiagnosticInfo()

	// Verify shards field is []ShardStat
	shards, ok := info["shards"]
	if !ok {
		t.Fatal("DiagnosticInfo missing 'shards' key")
	}
	shardStats, ok := shards.([]ShardStat)
	if !ok {
		t.Fatalf("shards is %T, want []ShardStat", shards)
	}
	if len(shardStats) != 8 {
		t.Fatalf("expected 8 shard stats, got %d", len(shardStats))
	}

	// Verify topic_shard_mapping field
	mappingRaw, ok := info["topic_shard_mapping"]
	if !ok {
		t.Fatal("DiagnosticInfo missing 'topic_shard_mapping' key")
	}
	mapping, ok := mappingRaw.(map[string]int)
	if !ok {
		t.Fatalf("topic_shard_mapping is %T, want map[string]int", mappingRaw)
	}
	if _, ok := mapping["diag.topic"]; !ok {
		t.Fatal("expected diag.topic in topic_shard_mapping")
	}
}

func TestNextPowerOf2(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{0, 1},
		{1, 1},
		{2, 2},
		{3, 4},
		{4, 4},
		{5, 8},
		{7, 8},
		{8, 8},
		{9, 16},
		{15, 16},
		{16, 16},
		{17, 32},
	}

	for _, tt := range tests {
		got := nextPowerOf2(tt.input)
		if got != tt.expected {
			t.Errorf("nextPowerOf2(%d) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestShardedMap_NonPowerOf2Count(t *testing.T) {
	// Should round up to next power of 2.
	sm := newShardedMap(10)
	if sm.shardCount != 16 {
		t.Fatalf("expected 16 shards for input 10, got %d", sm.shardCount)
	}
	if sm.shardMask != 15 {
		t.Fatalf("expected mask 15, got %d", sm.shardMask)
	}
}

func TestShardedMap_ZeroCount(t *testing.T) {
	sm := newShardedMap(0)
	if sm.shardCount != 16 {
		t.Fatalf("expected default 16 shards for input 0, got %d", sm.shardCount)
	}
}

func TestShardStats_MultipleSubscribersSameTopic(t *testing.T) {
	ps := New(WithShardCount(4))
	defer ps.Close()

	for i := 0; i < 5; i++ {
		_, err := ps.Subscribe("shared.topic", SubOptions{BufferSize: 4, Policy: DropOldest})
		if err != nil {
			t.Fatalf("subscribe %d: %v", i, err)
		}
	}

	stats := ps.ShardStats()
	totalSubs := 0
	totalTopics := 0
	for _, s := range stats {
		totalSubs += s.SubscriberCount
		totalTopics += s.TopicCount
	}

	if totalTopics != 1 {
		t.Fatalf("expected 1 topic, got %d", totalTopics)
	}
	if totalSubs != 5 {
		t.Fatalf("expected 5 subscribers, got %d", totalSubs)
	}
}
