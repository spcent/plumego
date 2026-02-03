package pubsub

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

const (
	// ReplyToHeader is the metadata key for reply topic
	ReplyToHeader = "X-Reply-To"
	// CorrelationIDHeader is the metadata key for correlation ID
	CorrelationIDHeader = "X-Correlation-ID"
	// DefaultRequestTimeout is the default timeout for Request operations
	DefaultRequestTimeout = 30 * time.Second
)

// requestManager manages request-reply pattern.
type requestManager struct {
	mu         sync.RWMutex
	pending    map[string]chan Message // correlationID -> response channel
	ps         *InProcPubSub
	replyTopic string
	replySub   Subscription
}

// newRequestManager creates a new request manager.
func newRequestManager(ps *InProcPubSub) *requestManager {
	rm := &requestManager{
		pending:    make(map[string]chan Message),
		ps:         ps,
		replyTopic: generateReplyTopic(),
	}

	// Subscribe to reply topic
	sub, err := ps.Subscribe(rm.replyTopic, SubOptions{
		BufferSize: 256,
		Policy:     DropNewest,
	})
	if err != nil {
		return nil
	}
	rm.replySub = sub

	// Start reply handler
	go rm.handleReplies()

	return rm
}

// generateReplyTopic generates a unique reply topic.
func generateReplyTopic() string {
	b := make([]byte, 16)
	rand.Read(b)
	return "_reply." + hex.EncodeToString(b)
}

// generateCorrelationID generates a unique correlation ID.
func generateCorrelationID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Request sends a message and waits for a response.
func (rm *requestManager) Request(ctx context.Context, topic string, msg Message) (Message, error) {
	if rm.ps.closed.Load() {
		return Message{}, ErrClosed
	}

	// Generate correlation ID
	correlationID := generateCorrelationID()

	// Ensure metadata exists
	if msg.Meta == nil {
		msg.Meta = make(map[string]string)
	}
	msg.Meta[ReplyToHeader] = rm.replyTopic
	msg.Meta[CorrelationIDHeader] = correlationID

	// Create response channel
	respCh := make(chan Message, 1)

	rm.mu.Lock()
	rm.pending[correlationID] = respCh
	rm.mu.Unlock()

	// Cleanup on exit
	defer func() {
		rm.mu.Lock()
		delete(rm.pending, correlationID)
		rm.mu.Unlock()
	}()

	// Publish request
	if err := rm.ps.Publish(topic, msg); err != nil {
		return Message{}, err
	}

	// Wait for response
	select {
	case resp := <-respCh:
		return resp, nil
	case <-ctx.Done():
		return Message{}, ctx.Err()
	}
}

// RequestWithTimeout sends a message and waits for a response with timeout.
func (rm *requestManager) RequestWithTimeout(topic string, msg Message, timeout time.Duration) (Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return rm.Request(ctx, topic, msg)
}

// handleReplies processes incoming replies.
func (rm *requestManager) handleReplies() {
	for msg := range rm.replySub.C() {
		correlationID := msg.Meta[CorrelationIDHeader]
		if correlationID == "" {
			continue
		}

		rm.mu.RLock()
		respCh, ok := rm.pending[correlationID]
		rm.mu.RUnlock()

		if ok {
			select {
			case respCh <- msg:
			default:
				// Response channel full or closed
			}
		}
	}
}

// Close closes the request manager.
func (rm *requestManager) Close() {
	if rm.replySub != nil {
		rm.replySub.Cancel()
	}

	rm.mu.Lock()
	for _, ch := range rm.pending {
		close(ch)
	}
	rm.pending = make(map[string]chan Message)
	rm.mu.Unlock()
}

// Reply sends a reply to a request message.
func (ps *InProcPubSub) Reply(reqMsg Message, respMsg Message) error {
	replyTo := reqMsg.Meta[ReplyToHeader]
	if replyTo == "" {
		return NewErrorWithTopic(ErrCodeNotFound, "reply", reqMsg.Topic, "no reply-to header in request")
	}

	correlationID := reqMsg.Meta[CorrelationIDHeader]
	if correlationID == "" {
		return NewErrorWithTopic(ErrCodeNotFound, "reply", reqMsg.Topic, "no correlation ID in request")
	}

	// Copy correlation ID to response
	if respMsg.Meta == nil {
		respMsg.Meta = make(map[string]string)
	}
	respMsg.Meta[CorrelationIDHeader] = correlationID

	return ps.Publish(replyTo, respMsg)
}

// Request sends a message and waits for a response.
// This creates a temporary subscription for the reply.
func (ps *InProcPubSub) Request(ctx context.Context, topic string, msg Message) (Message, error) {
	if ps.closed.Load() {
		return Message{}, ErrClosed
	}

	// Generate unique reply topic and correlation ID
	replyTopic := generateReplyTopic()
	correlationID := generateCorrelationID()

	// Set up message metadata
	if msg.Meta == nil {
		msg.Meta = make(map[string]string)
	}
	msg.Meta[ReplyToHeader] = replyTopic
	msg.Meta[CorrelationIDHeader] = correlationID

	// Subscribe to reply topic
	replySub, err := ps.Subscribe(replyTopic, SubOptions{
		BufferSize: 1,
		Policy:     DropNewest,
	})
	if err != nil {
		return Message{}, err
	}
	defer replySub.Cancel()

	// Publish the request
	if err := ps.Publish(topic, msg); err != nil {
		return Message{}, err
	}

	// Wait for response
	select {
	case resp := <-replySub.C():
		return resp, nil
	case <-ctx.Done():
		return Message{}, ctx.Err()
	}
}

// RequestWithTimeout sends a message and waits for a response with timeout.
func (ps *InProcPubSub) RequestWithTimeout(topic string, msg Message, timeout time.Duration) (Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return ps.Request(ctx, topic, msg)
}

// IsRequest checks if a message is a request that expects a reply.
func IsRequest(msg Message) bool {
	return msg.Meta != nil && msg.Meta[ReplyToHeader] != ""
}

// GetCorrelationID returns the correlation ID from a message.
func GetCorrelationID(msg Message) string {
	if msg.Meta == nil {
		return ""
	}
	return msg.Meta[CorrelationIDHeader]
}

// GetReplyTo returns the reply-to topic from a message.
func GetReplyTo(msg Message) string {
	if msg.Meta == nil {
		return ""
	}
	return msg.Meta[ReplyToHeader]
}
