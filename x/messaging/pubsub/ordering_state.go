package pubsub

// createQueue creates a new ordered queue.
func (ops *OrderedPubSub) createQueue(batchSize int) *orderedQueue {
	return &orderedQueue{
		ch:        make(chan orderedMessage, ops.config.QueueSize),
		batch:     make([]orderedMessage, 0, batchSize),
		batchSize: batchSize,
	}
}

// getOrCreateTopicQueue gets or creates a queue for a topic.
func (ops *OrderedPubSub) getOrCreateTopicQueue(topic string) *orderedQueue {
	ops.topicQueuesMu.RLock()
	queue, exists := ops.topicQueues[topic]
	ops.topicQueuesMu.RUnlock()

	if exists {
		return queue
	}

	ops.topicQueuesMu.Lock()
	defer ops.topicQueuesMu.Unlock()

	// Double-check
	if queue, exists = ops.topicQueues[topic]; exists {
		return queue
	}

	// Create new queue
	queue = ops.createQueue(ops.config.MaxBatchSize)
	ops.topicQueues[topic] = queue

	// Start worker
	ops.startQueueWorker(queue, topic)

	return queue
}

// getOrCreateKeyQueue gets or creates a queue for a key.
func (ops *OrderedPubSub) getOrCreateKeyQueue(key string) *orderedQueue {
	ops.keyQueuesMu.RLock()
	queue, exists := ops.keyQueues[key]
	ops.keyQueuesMu.RUnlock()

	if exists {
		return queue
	}

	ops.keyQueuesMu.Lock()
	defer ops.keyQueuesMu.Unlock()

	// Double-check
	if queue, exists = ops.keyQueues[key]; exists {
		return queue
	}

	// Create new queue
	queue = ops.createQueue(ops.config.MaxBatchSize)
	ops.keyQueues[key] = queue

	// Start worker
	ops.startQueueWorker(queue, key)

	return queue
}

// verifySequence verifies message sequence using the queue identifier as the
// sequence tracking key. This ensures correct behavior across all ordering levels:
//   - OrderPerTopic: identifier is the topic, so sequences are tracked per-topic
//   - OrderPerKey: identifier is the partition key, so sequences are tracked per-key
//   - OrderGlobal: identifier is globalQueueIdentifier, so all messages share one sequence space
func (ops *OrderedPubSub) verifySequence(om orderedMessage, identifier string) error {
	ops.sequencesMu.Lock()
	tracker, exists := ops.sequences[identifier]
	if !exists {
		tracker = &sequenceTracker{
			missing: make(map[uint64]bool),
		}
		ops.sequences[identifier] = tracker
	}
	ops.sequencesMu.Unlock()

	tracker.mu.Lock()
	defer tracker.mu.Unlock()

	expected := tracker.lastSequence + 1
	if om.sequence < expected {
		return ErrOutOfOrderMessage
	}

	if om.sequence > expected {
		// Gap detected - record missing sequences up to the cap.
		// If the cap is reached we stop tracking further gaps to prevent
		// unbounded memory growth from permanently lost messages.
		for seq := expected; seq < om.sequence; seq++ {
			if len(tracker.missing) >= maxMissingSequences {
				break
			}
			tracker.missing[seq] = true
		}
	}

	// Check if this fills a gap
	delete(tracker.missing, om.sequence)
	tracker.lastSequence = om.sequence

	return nil
}

// OrderingStats returns ordering statistics including per-queue metrics.
func (ops *OrderedPubSub) OrderingStats() OrderingStats {
	queueStats := make(map[string]QueueStat)

	ops.topicQueuesMu.RLock()
	topicQueues := len(ops.topicQueues)
	for id, q := range ops.topicQueues {
		queueStats[id] = QueueStat{
			Processed:      q.processed.Load(),
			BatchFlushes:   q.batchFlushes.Load(),
			SequenceErrors: q.sequenceErrors.Load(),
			Pending:        len(q.ch),
		}
	}
	ops.topicQueuesMu.RUnlock()

	ops.keyQueuesMu.RLock()
	keyQueues := len(ops.keyQueues)
	for id, q := range ops.keyQueues {
		queueStats[id] = QueueStat{
			Processed:      q.processed.Load(),
			BatchFlushes:   q.batchFlushes.Load(),
			SequenceErrors: q.sequenceErrors.Load(),
			Pending:        len(q.ch),
		}
	}
	ops.keyQueuesMu.RUnlock()

	if ops.globalQueue != nil {
		queueStats[globalQueueIdentifier] = QueueStat{
			Processed:      ops.globalQueue.processed.Load(),
			BatchFlushes:   ops.globalQueue.batchFlushes.Load(),
			SequenceErrors: ops.globalQueue.sequenceErrors.Load(),
			Pending:        len(ops.globalQueue.ch),
		}
	}

	return OrderingStats{
		OrderedPublishes: ops.orderedPublishes.Load(),
		QueuedMessages:   ops.queuedMessages.Load(),
		SequenceErrors:   ops.sequenceErrors.Load(),
		TopicQueues:      topicQueues,
		KeyQueues:        keyQueues,
		QueueStats:       queueStats,
	}
}

// MissingSequences returns missing sequence numbers.
func (os *OrderedSubscription) MissingSequences() []uint64 {
	os.tracker.mu.Lock()
	defer os.tracker.mu.Unlock()

	result := make([]uint64, 0, len(os.tracker.missing))
	for seq := range os.tracker.missing {
		result = append(result, seq)
	}

	return result
}
