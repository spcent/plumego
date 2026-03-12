package pubsub

import (
	"context"
	"path"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spcent/plumego/metrics"
)

// subKind distinguishes the type of subscription.
type subKind int

const (
	subKindTopic   subKind = iota // exact topic subscription
	subKindPattern                // glob/filepath pattern subscription
	subKindMQTT                   // MQTT-style pattern subscription
)

// subscriber represents a subscription to a topic.
type subscriber struct {
	id     uint64
	topic  string
	ch     chan Message
	opts   SubOptions
	closed atomic.Bool
	once   sync.Once
	mu     sync.Mutex
	kind   subKind
	done   chan struct{}

	// Ring buffer for O(1) DropOldest (nil when not using ring buffer)
	ringBuf  *ringBuffer
	pumpDone chan struct{}

	// Statistics
	received atomic.Uint64
	dropped  atomic.Uint64

	// back-reference to parent
	ps *InProcBroker

	// context cancellation
	cancel context.CancelFunc
}

type patternSnapshot struct {
	pattern string
	isGlob  bool // true when pattern contains glob metacharacters (*?[]\\)
	subs    []*subscriber
}

// C returns the receive-only message channel.
func (s *subscriber) C() <-chan Message { return s.ch }

// ID returns the unique subscription ID.
func (s *subscriber) ID() uint64 { return s.id }

// Topic returns the topic or pattern this subscription is for.
func (s *subscriber) Topic() string { return s.topic }

// Done returns a channel that is closed when the subscription is cancelled.
func (s *subscriber) Done() <-chan struct{} { return s.done }

// Stats returns subscription statistics.
func (s *subscriber) Stats() SubscriptionStats {
	if s.ringBuf != nil {
		return SubscriptionStats{
			Received: s.received.Load(),
			Dropped:  s.dropped.Load(),
			QueueLen: s.ringBuf.Len(),
			QueueCap: s.ringBuf.Cap(),
		}
	}
	return SubscriptionStats{
		Received: s.received.Load(),
		Dropped:  s.dropped.Load(),
		QueueLen: len(s.ch),
		QueueCap: cap(s.ch),
	}
}

// Cancel unsubscribes and closes the channel.
// sync.Once guarantees this runs at most once; no secondary closed check needed.
func (s *subscriber) Cancel() {
	s.once.Do(func() {
		s.mu.Lock()
		s.closed.Store(true)
		if s.ringBuf != nil {
			s.ringBuf.Close()
		} else {
			close(s.ch)
		}
		close(s.done)
		if s.cancel != nil {
			s.cancel()
		}
		s.mu.Unlock()

		// Wait for pump goroutine to finish if using ring buffer
		if s.pumpDone != nil {
			<-s.pumpDone
		}

		s.ps.notifyUnsubscribe(s.topic, s.id)
		s.ps.removeSubscriber(s.topic, s.id, s.kind)
	})
}

// pumpRingBuffer moves messages from the ring buffer to the output channel.
// Termination is driven exclusively by s.done and ringBuf.Notify() closure.
func (s *subscriber) pumpRingBuffer() {
	defer close(s.ch)
	defer close(s.pumpDone)

	for {
		for {
			msg, ok := s.ringBuf.Pop()
			if !ok {
				break
			}
			select {
			case s.ch <- msg:
			case <-s.done:
				return
			}
		}

		select {
		case _, ok := <-s.ringBuf.Notify():
			if !ok {
				return
			}
		case <-s.done:
			return
		}
	}
}

// InProcBroker is the in-memory Broker implementation.
//
// It implements Broker, PatternSubscriber, MQTTSubscriber, BatchPublisher,
// and Drainable. Callers should declare variables using the narrowest
// interface they depend on.
//
// Example:
//
//	b := pubsub.New(
//		pubsub.WithShardCount(32),
//		pubsub.WithWorkerPoolSize(512),
//	)
//	defer b.Close()
//
//	sub, err := b.Subscribe("user.created", pubsub.DefaultSubOptions())
//	if err != nil {
//		// handle
//	}
//	defer sub.Cancel()
//
//	err = b.Publish("user.created", pubsub.Message{Data: "hello"})
type InProcBroker struct {
	shards *shardedMap
	config Config

	closed atomic.Bool
	nextID atomic.Uint64

	metrics    metricsPubSub
	workerPool *workerPool

	// draining state
	draining atomic.Bool
	drainWg  sync.WaitGroup
}

// Compile-time interface assertions.
var (
	_ Broker            = (*InProcBroker)(nil)
	_ PatternSubscriber = (*InProcBroker)(nil)
	_ MQTTSubscriber    = (*InProcBroker)(nil)
	_ BatchPublisher    = (*InProcBroker)(nil)
	_ Drainable         = (*InProcBroker)(nil)
)

// New creates a new InProcBroker with the given options.
//
// Capability features (history, scheduling, request-reply, etc.) are provided
// by wrapping the returned broker with sub-package constructors — not by
// passing Enable* flags here.
//
// Example:
//
//	b := pubsub.New(
//		pubsub.WithShardCount(32),
//		pubsub.WithDefaultBufferSize(64),
//		pubsub.WithWorkerPoolSize(512),
//	)
//	defer b.Close()
func New(opts ...Option) *InProcBroker {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(&config)
	}

	b := &InProcBroker{
		shards:     newShardedMap(config.ShardCount),
		config:     config,
		workerPool: newWorkerPool(config.WorkerPoolSize),
	}
	b.nextID.Store(0)

	return b
}

// Close shuts down the broker and cancels all active subscriptions.
func (b *InProcBroker) Close() error {
	if b.closed.Swap(true) {
		return nil
	}

	// Wait for pending async operations
	b.workerPool.Close()

	// Cancel all subscribers
	for _, s := range b.shards.collectAllSubscribers() {
		s.Cancel()
	}

	return nil
}

// Drain blocks until all in-flight messages are delivered or ctx is cancelled.
func (b *InProcBroker) Drain(ctx context.Context) error {
	if b.closed.Load() {
		return newErr(ErrCodeClosed, "drain", "", "broker is closed", nil)
	}

	b.draining.Store(true)
	defer b.draining.Store(false)

	done := make(chan struct{})
	go func() {
		b.drainWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Subscribe creates a new subscription to an exact topic.
func (b *InProcBroker) Subscribe(topic string, opts SubOptions) (Subscription, error) {
	return b.subscribeInternal(context.Background(), topic, opts, false)
}

// SubscribeWithContext creates a subscription cancelled when ctx is done.
func (b *InProcBroker) SubscribeWithContext(ctx context.Context, topic string, opts SubOptions) (Subscription, error) {
	return b.subscribeInternal(ctx, topic, opts, false)
}

// SubscribePattern creates a subscription matching a glob pattern.
func (b *InProcBroker) SubscribePattern(pattern string, opts SubOptions) (Subscription, error) {
	return b.subscribeInternal(context.Background(), pattern, opts, true)
}

// SubscribePatternWithContext creates a pattern subscription cancelled when ctx is done.
func (b *InProcBroker) SubscribePatternWithContext(ctx context.Context, pattern string, opts SubOptions) (Subscription, error) {
	return b.subscribeInternal(ctx, pattern, opts, true)
}

// SubscribeMQTT creates a subscription using MQTT-style topic patterns.
func (b *InProcBroker) SubscribeMQTT(pattern string, opts SubOptions) (Subscription, error) {
	return b.subscribeMQTTInternal(context.Background(), pattern, opts)
}

// SubscribeMQTTWithContext creates an MQTT subscription cancelled when ctx is done.
func (b *InProcBroker) SubscribeMQTTWithContext(ctx context.Context, pattern string, opts SubOptions) (Subscription, error) {
	return b.subscribeMQTTInternal(ctx, pattern, opts)
}

// initSubscriberChannel sets up the subscriber's message channel.
func initSubscriberChannel(sub *subscriber, opts SubOptions) {
	if opts.UseRingBuffer && opts.Policy == DropOldest {
		sub.ringBuf = newRingBuffer(opts.BufferSize)
		sub.ch = make(chan Message)
		sub.pumpDone = make(chan struct{})
		go sub.pumpRingBuffer()
	} else {
		sub.ch = make(chan Message, opts.BufferSize)
	}
}

// subscribeInternal handles both exact-topic and glob-pattern subscriptions.
func (b *InProcBroker) subscribeInternal(ctx context.Context, topic string, opts SubOptions, isPattern bool) (Subscription, error) {
	start := time.Now()
	var err error
	defer func() {
		b.recordMetrics(ctx, "subscribe", topic, time.Since(start), err)
	}()

	if b.closed.Load() {
		err = newErr(ErrCodeClosed, "subscribe", topic, "broker is closed", nil)
		return nil, err
	}

	topic = strings.TrimSpace(topic)
	if topic == "" {
		if isPattern {
			err = newErr(ErrCodeInvalidPattern, "subscribe", "", "pattern must not be empty", nil)
		} else {
			err = newErr(ErrCodeInvalidTopic, "subscribe", "", "topic must not be empty", nil)
		}
		return nil, err
	}

	if isPattern {
		if _, matchErr := path.Match(topic, "probe"); matchErr != nil {
			err = newErr(ErrCodeInvalidPattern, "subscribe", topic, "invalid glob pattern", matchErr)
			return nil, err
		}
	}

	opts = b.normalizeSubOptions(opts)
	if opts.BufferSize < 1 {
		err = newErr(ErrCodeInvalidOptions, "subscribe", topic, "buffer size must be at least 1", nil)
		return nil, err
	}

	id := b.nextID.Add(1)

	needsCtxWatch := ctx != nil && ctx != context.Background() && ctx != context.TODO()

	var cancel context.CancelFunc
	if needsCtxWatch {
		ctx, cancel = context.WithCancel(ctx)
	}

	kind := subKindTopic
	if isPattern {
		kind = subKindPattern
	}

	sub := &subscriber{
		id:     id,
		topic:  topic,
		opts:   opts,
		ps:     b,
		kind:   kind,
		done:   make(chan struct{}),
		cancel: cancel,
	}

	initSubscriberChannel(sub, opts)

	if isPattern {
		b.shards.addPattern(topic, id, sub)
	} else {
		b.shards.addTopic(topic, id, sub)
	}
	b.metrics.addSubs(topic, 1)

	b.notifySubscribe(topic, id)

	if needsCtxWatch {
		go func() {
			select {
			case <-ctx.Done():
				sub.Cancel()
			case <-sub.done:
			}
		}()
	}

	return sub, nil
}

// subscribeMQTTInternal handles MQTT-pattern subscriptions.
func (b *InProcBroker) subscribeMQTTInternal(ctx context.Context, pattern string, opts SubOptions) (Subscription, error) {
	start := time.Now()
	var err error
	defer func() {
		b.recordMetrics(ctx, "subscribe_mqtt", pattern, time.Since(start), err)
	}()

	if b.closed.Load() {
		err = newErr(ErrCodeClosed, "subscribe_mqtt", pattern, "broker is closed", nil)
		return nil, err
	}

	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		err = newErr(ErrCodeInvalidPattern, "subscribe_mqtt", "", "pattern must not be empty", nil)
		return nil, err
	}

	if validErr := ValidateMQTTPattern(pattern); validErr != nil {
		err = newErr(ErrCodeInvalidPattern, "subscribe_mqtt", pattern, validErr.Error(), validErr)
		return nil, err
	}

	opts = b.normalizeSubOptions(opts)
	if opts.BufferSize < 1 {
		err = newErr(ErrCodeInvalidOptions, "subscribe_mqtt", pattern, "buffer size must be at least 1", nil)
		return nil, err
	}

	id := b.nextID.Add(1)

	needsCtxWatch := ctx != nil && ctx != context.Background() && ctx != context.TODO()

	var cancel context.CancelFunc
	if needsCtxWatch {
		ctx, cancel = context.WithCancel(ctx)
	}

	sub := &subscriber{
		id:     id,
		topic:  pattern,
		opts:   opts,
		ps:     b,
		kind:   subKindMQTT,
		done:   make(chan struct{}),
		cancel: cancel,
	}

	initSubscriberChannel(sub, opts)

	b.shards.addMQTTPattern(pattern, id, sub)
	b.metrics.addSubs(pattern, 1)

	b.notifySubscribe(pattern, id)

	if needsCtxWatch {
		go func() {
			select {
			case <-ctx.Done():
				sub.Cancel()
			case <-sub.done:
			}
		}()
	}

	return sub, nil
}

// Publish sends a message to a topic.
func (b *InProcBroker) Publish(topic string, msg Message) error {
	return b.publishInternal(context.Background(), topic, msg, false)
}

// PublishWithContext publishes a message with context support.
func (b *InProcBroker) PublishWithContext(ctx context.Context, topic string, msg Message) error {
	return b.publishInternal(ctx, topic, msg, false)
}

// PublishAsync publishes without waiting for delivery (fire-and-forget).
func (b *InProcBroker) PublishAsync(topic string, msg Message) error {
	return b.publishInternal(context.Background(), topic, msg, true)
}

// PublishBatch publishes multiple messages to a topic atomically.
func (b *InProcBroker) PublishBatch(topic string, msgs []Message) error {
	if b.closed.Load() {
		return newErr(ErrCodeClosed, "publish_batch", topic, "broker is closed", nil)
	}

	topic = strings.TrimSpace(topic)
	if topic == "" {
		return newErr(ErrCodeInvalidTopic, "publish_batch", "", "topic must not be empty", nil)
	}

	for i := range msgs {
		msgs[i].Topic = topic
		if msgs[i].Time.IsZero() {
			msgs[i].Time = time.Now().UTC()
		}
	}

	subs := b.shards.getTopicSubscribersIfAny(topic)
	var patterns []patternSnapshot
	if b.shards.hasAnyPatterns() {
		patterns = b.shards.getAllPatterns()
	}
	var mqttSubs []*subscriber
	if b.shards.hasAnyMQTTPatterns() {
		mqttSubs = b.shards.getMQTTPatternMatches(topic)
	}

	b.drainWg.Add(1)
	defer b.drainWg.Done()

	bgCtx := context.Background()
	for _, msg := range msgs {
		clonedMsg := cloneMessage(msg)
		b.metrics.incPublish(topic)
		b.notifyPublish(topic, &clonedMsg)

		for _, s := range subs {
			b.deliver(bgCtx, s, clonedMsg)
		}
		if len(patterns) > 0 {
			b.deliverToPatterns(bgCtx, patterns, topic, clonedMsg)
		}
		for _, s := range mqttSubs {
			b.deliver(bgCtx, s, clonedMsg)
		}
	}

	return nil
}

// PublishMulti publishes messages to multiple topics.
func (b *InProcBroker) PublishMulti(msgs map[string][]Message) error {
	if b.closed.Load() {
		return newErr(ErrCodeClosed, "publish_multi", "", "broker is closed", nil)
	}

	for topic, topicMsgs := range msgs {
		if err := b.PublishBatch(topic, topicMsgs); err != nil {
			return err
		}
	}

	return nil
}

// publishInternal is the shared implementation for Publish and PublishAsync.
func (b *InProcBroker) publishInternal(ctx context.Context, topic string, msg Message, async bool) error {
	start := time.Now()
	var err error
	defer func() {
		op := "publish"
		if async {
			op = "publish_async"
		}
		b.recordMetrics(ctx, op, topic, time.Since(start), err)
	}()

	if ctx.Err() != nil {
		err = ctx.Err()
		return err
	}

	if b.closed.Load() {
		err = newErr(ErrCodeClosed, "publish", topic, "broker is closed", nil)
		return err
	}

	topic = strings.TrimSpace(topic)
	if topic == "" {
		err = newErr(ErrCodeInvalidTopic, "publish", "", "topic must not be empty", nil)
		return err
	}

	msg.Topic = topic
	if msg.Time.IsZero() {
		msg.Time = time.Now().UTC()
	}
	msg = cloneMessage(msg)

	b.metrics.incPublish(topic)
	b.notifyPublish(topic, &msg)

	if async {
		b.drainWg.Add(1)
		submitted := b.workerPool.SubmitWithContext(ctx, func() {
			defer b.drainWg.Done()
			b.doPublish(ctx, topic, msg)
		})
		if !submitted {
			b.drainWg.Done()
			if ctx.Err() != nil {
				err = ctx.Err()
				return err
			}
			b.doPublish(ctx, topic, msg)
		}
	} else {
		b.drainWg.Add(1)
		defer b.drainWg.Done()
		b.doPublish(ctx, topic, msg)
	}

	return nil
}

// doPublish performs the actual fan-out delivery.
func (b *InProcBroker) doPublish(ctx context.Context, topic string, msg Message) {
	if ctx.Err() != nil {
		return
	}

	hasPatternSubs := b.shards.hasAnyPatterns()
	hasMQTTSubs := b.shards.hasAnyMQTTPatterns()

	subs := b.shards.getTopicSubscribersIfAny(topic)
	if subs == nil && !hasPatternSubs && !hasMQTTSubs {
		return
	}

	for _, s := range subs {
		if ctx.Err() != nil {
			return
		}
		b.deliver(ctx, s, msg)
	}

	if hasPatternSubs {
		b.deliverToPatterns(ctx, b.shards.getAllPatterns(), topic, msg)
	}

	if hasMQTTSubs {
		for _, s := range b.shards.getMQTTPatternMatches(topic) {
			if ctx.Err() != nil {
				return
			}
			b.deliver(ctx, s, msg)
		}
	}
}

// deliverToPatterns delivers a message to all matching glob-pattern subscribers.
func (b *InProcBroker) deliverToPatterns(ctx context.Context, patterns []patternSnapshot, topic string, msg Message) {
	for _, entry := range patterns {
		if ctx.Err() != nil {
			return
		}

		var matched bool
		if entry.isGlob {
			var matchErr error
			matched, matchErr = path.Match(entry.pattern, topic)
			if matchErr != nil {
				matched = false
			}
		} else {
			matched = entry.pattern == topic
		}

		if matched {
			for _, s := range entry.subs {
				if ctx.Err() != nil {
					return
				}
				b.deliver(ctx, s, msg)
			}
		}
	}
}

// GetSubscriberCount returns the number of exact-topic subscribers.
func (b *InProcBroker) GetSubscriberCount(topic string) int {
	return b.shards.getTopicSubscriberCount(topic)
}

// GetPatternSubscriberCount returns the number of glob-pattern subscribers.
func (b *InProcBroker) GetPatternSubscriberCount(pattern string) int {
	return b.shards.getPatternSubscriberCount(pattern)
}

// GetMQTTSubscriberCount returns the number of MQTT-pattern subscribers.
func (b *InProcBroker) GetMQTTSubscriberCount(pattern string) int {
	return b.shards.getMQTTPatternSubscriberCount(pattern)
}

// TopicExists reports whether the topic has any active exact subscribers.
func (b *InProcBroker) TopicExists(topic string) bool {
	return b.shards.topicExists(topic)
}

// PatternExists reports whether the glob pattern has any active subscribers.
func (b *InProcBroker) PatternExists(pattern string) bool {
	return b.shards.patternExists(pattern)
}

// MQTTPatternExists reports whether the MQTT pattern has any active subscribers.
func (b *InProcBroker) MQTTPatternExists(pattern string) bool {
	return b.shards.mqttPatternExists(pattern)
}

// HasSubscribers reports whether publishing to topic would reach any subscriber
// (exact, glob-pattern, or MQTT-pattern).
func (b *InProcBroker) HasSubscribers(topic string) bool {
	if b.shards.topicExists(topic) {
		return true
	}
	if b.shards.hasAnyPatterns() {
		for _, entry := range b.shards.getAllPatterns() {
			if entry.isGlob {
				if matched, err := path.Match(entry.pattern, topic); err == nil && matched {
					return true
				}
			} else if entry.pattern == topic {
				return true
			}
		}
	}
	if b.shards.hasAnyMQTTPatterns() {
		if len(b.shards.getMQTTPatternMatches(topic)) > 0 {
			return true
		}
	}
	return false
}

// ListTopics returns all active exact-topic names, sorted.
func (b *InProcBroker) ListTopics() []string {
	topics := b.shards.listTopics()
	sort.Strings(topics)
	return topics
}

// ListPatterns returns all active glob-pattern subscriptions, sorted.
func (b *InProcBroker) ListPatterns() []string {
	patterns := b.shards.listPatterns()
	sort.Strings(patterns)
	return patterns
}

// ListMQTTPatterns returns all active MQTT-pattern subscriptions, sorted.
func (b *InProcBroker) ListMQTTPatterns() []string {
	patterns := b.shards.listMQTTPatterns()
	sort.Strings(patterns)
	return patterns
}

// removeSubscriber removes a subscription from the sharded map.
func (b *InProcBroker) removeSubscriber(topic string, id uint64, kind subKind) {
	switch kind {
	case subKindPattern:
		b.shards.removePattern(topic, id)
	case subKindMQTT:
		b.shards.removeMQTTPattern(topic, id)
	default:
		b.shards.removeTopic(topic, id)
	}
	b.metrics.addSubs(topic, -1)
}

// deliver handles message delivery to a single subscriber with the configured backpressure policy.
func (b *InProcBroker) deliver(ctx context.Context, s *subscriber, msg Message) {
	if b.config.EnablePanicRecovery {
		defer func() {
			if r := recover(); r != nil {
				if b.config.OnPanic != nil {
					b.config.OnPanic(s.topic, s.id, r)
				}
				s.Cancel()
			}
		}()
	}

	metricTopic := s.topic
	if s.kind != subKindTopic {
		metricTopic = msg.Topic
	}

	if s.opts.Filter != nil && !s.opts.Filter(msg) {
		return
	}

	if !s.opts.ZeroCopy {
		msg = cloneMessage(msg)
	}

	if s.closed.Load() {
		return
	}

	if s.opts.Policy == BlockWithTimeout {
		if b.deliverBlockWithTimeout(ctx, s, msg, metricTopic) {
			s.Cancel()
		}
		return
	}

	s.mu.Lock()
	if s.closed.Load() {
		s.mu.Unlock()
		return
	}

	var cancelSub bool

	switch s.opts.Policy {
	case DropOldest:
		cancelSub = b.deliverDropOldest(s, msg, metricTopic)
	case DropNewest:
		cancelSub = b.deliverDropNewest(s, msg, metricTopic)
	case CloseSubscriber:
		cancelSub = b.deliverCloseSubscriber(s, msg, metricTopic)
	default:
		cancelSub = b.deliverDropOldest(s, msg, metricTopic)
	}

	s.mu.Unlock()

	if cancelSub {
		s.Cancel()
	}
}

// deliverDropOldest drops the oldest message when the buffer is full.
func (b *InProcBroker) deliverDropOldest(s *subscriber, msg Message, metricTopic string) bool {
	if s.ringBuf != nil {
		dropped, wasDropped := s.ringBuf.Push(msg)
		if wasDropped {
			b.metrics.incDropped(metricTopic, DropOldest)
			s.dropped.Add(1)
			b.notifyDrop(metricTopic, s.id, &dropped, DropOldest)
		}
		b.metrics.incDelivered(metricTopic)
		s.received.Add(1)
		b.notifyDeliver(metricTopic, s.id, &msg)
		return false
	}

	for spinCount := 0; ; spinCount++ {
		if s.closed.Load() {
			return false
		}
		select {
		case s.ch <- msg:
			b.metrics.incDelivered(metricTopic)
			s.received.Add(1)
			b.notifyDeliver(metricTopic, s.id, &msg)
			return false
		default:
			select {
			case dropped := <-s.ch:
				b.metrics.incDropped(metricTopic, DropOldest)
				s.dropped.Add(1)
				b.notifyDrop(metricTopic, s.id, &dropped, DropOldest)
			default:
				runtime.Gosched()
			}
		}
		if spinCount > 0 && spinCount%64 == 0 {
			runtime.Gosched()
		}
	}
}

// deliverDropNewest discards the incoming message when the buffer is full.
func (b *InProcBroker) deliverDropNewest(s *subscriber, msg Message, metricTopic string) bool {
	select {
	case s.ch <- msg:
		b.metrics.incDelivered(metricTopic)
		s.received.Add(1)
		b.notifyDeliver(metricTopic, s.id, &msg)
		return false
	default:
		b.metrics.incDropped(metricTopic, DropNewest)
		s.dropped.Add(1)
		b.notifyDrop(metricTopic, s.id, &msg, DropNewest)
		return false
	}
}

// deliverBlockWithTimeout blocks until buffer space is available or the deadline elapses.
func (b *InProcBroker) deliverBlockWithTimeout(ctx context.Context, s *subscriber, msg Message, metricTopic string) bool {
	timeout := s.opts.BlockTimeout
	if timeout <= 0 {
		timeout = b.config.DefaultBlockTimeout
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case s.ch <- msg:
		b.metrics.incDelivered(metricTopic)
		s.received.Add(1)
		b.notifyDeliver(metricTopic, s.id, &msg)
		return false
	case <-timer.C:
		b.metrics.incDropped(metricTopic, BlockWithTimeout)
		s.dropped.Add(1)
		b.notifyDrop(metricTopic, s.id, &msg, BlockWithTimeout)
		return false
	case <-ctx.Done():
		b.metrics.incDropped(metricTopic, BlockWithTimeout)
		s.dropped.Add(1)
		b.notifyDrop(metricTopic, s.id, &msg, BlockWithTimeout)
		return false
	}
}

// deliverCloseSubscriber closes the subscription when the buffer is full.
func (b *InProcBroker) deliverCloseSubscriber(s *subscriber, msg Message, metricTopic string) bool {
	select {
	case s.ch <- msg:
		b.metrics.incDelivered(metricTopic)
		s.received.Add(1)
		b.notifyDeliver(metricTopic, s.id, &msg)
		return false
	default:
		b.metrics.incDropped(metricTopic, CloseSubscriber)
		s.dropped.Add(1)
		b.notifyDrop(metricTopic, s.id, &msg, CloseSubscriber)
		return true
	}
}

// normalizeSubOptions fills in defaults from the broker config.
func (b *InProcBroker) normalizeSubOptions(opts SubOptions) SubOptions {
	if opts.BufferSize <= 0 {
		opts.BufferSize = b.config.DefaultBufferSize
	}

	switch opts.Policy {
	case DropOldest, DropNewest, BlockWithTimeout, CloseSubscriber:
	default:
		opts.Policy = b.config.DefaultPolicy
	}

	if opts.Policy == BlockWithTimeout && opts.BlockTimeout <= 0 {
		opts.BlockTimeout = b.config.DefaultBlockTimeout
	}

	// Ring buffer is the default for DropOldest subscribers.
	if opts.Policy == DropOldest {
		opts.UseRingBuffer = true
	}

	return opts
}

// Snapshot returns a point-in-time view of pubsub metrics.
func (b *InProcBroker) Snapshot() MetricsSnapshot {
	return b.metrics.Snapshot()
}

// SetMetricsObserver replaces the external metrics sink at runtime.
func (b *InProcBroker) SetMetricsObserver(observer metrics.PubSubObserver) {
	b.config.MetricsObserver = observer
}

// Config returns a copy of the broker's configuration.
func (b *InProcBroker) Config() Config {
	return b.config
}

// TopicShard returns the shard index for the given topic/pattern.
func (b *InProcBroker) TopicShard(topic string) int {
	return b.shards.getShardIndex(topic)
}

// ShardStats returns per-shard statistics.
func (b *InProcBroker) ShardStats() []ShardStat {
	return b.shards.shardStats()
}

// TopicShardMapping returns topic→shard-index mapping for all active topics.
func (b *InProcBroker) TopicShardMapping() map[string]int {
	return b.shards.topicShardMapping()
}

// recordMetrics forwards an operation observation to the external collector, if any.
func (b *InProcBroker) recordMetrics(ctx context.Context, operation, topic string, duration time.Duration, err error) {
	if b.config.MetricsObserver == nil {
		return
	}
	b.config.MetricsObserver.ObservePubSub(ctx, operation, topic, duration, err)
}

// --- Observer notification helpers ---

func (b *InProcBroker) notifyPublish(topic string, msg *Message) {
	for _, o := range b.config.observers {
		o.OnPublish(topic, msg)
	}
}

func (b *InProcBroker) notifySubscribe(topic string, subID uint64) {
	for _, o := range b.config.observers {
		o.OnSubscribe(topic, subID)
	}
}

func (b *InProcBroker) notifyUnsubscribe(topic string, subID uint64) {
	for _, o := range b.config.observers {
		o.OnUnsubscribe(topic, subID)
	}
}

func (b *InProcBroker) notifyDeliver(topic string, subID uint64, msg *Message) {
	for _, o := range b.config.observers {
		o.OnDeliver(topic, subID, msg)
	}
}

func (b *InProcBroker) notifyDrop(topic string, subID uint64, msg *Message, policy BackpressurePolicy) {
	for _, o := range b.config.observers {
		o.OnDrop(topic, subID, msg, policy)
	}
}
