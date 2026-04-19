package pubsub

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

const (
	// ReplyToHeader is the metadata key for the reply topic.
	ReplyToHeader = "X-Reply-To"
	// CorrelationIDHeader is the metadata key for the correlation ID.
	CorrelationIDHeader = "X-Correlation-ID"
	// DefaultRequestTimeout is the default timeout for Request operations.
	DefaultRequestTimeout = 30 * time.Second
)

// requestManager manages the request-reply pattern using a shared reply subscription.
type requestManager struct {
	mu         sync.RWMutex
	pending    map[string]chan Message
	ps         *InProcBroker
	replyTopic string
	replySub   Subscription
	done       chan struct{}
}

// newRequestManager creates a request manager for the given broker.
func newRequestManager(ps *InProcBroker) (*requestManager, error) {
	rm := &requestManager{
		pending:    make(map[string]chan Message),
		ps:         ps,
		replyTopic: generateReplyTopic(),
		done:       make(chan struct{}),
	}

	sub, err := ps.Subscribe(context.Background(), rm.replyTopic, SubOptions{
		BufferSize: 256,
		Policy:     DropNewest,
	})
	if err != nil {
		return nil, err
	}
	rm.replySub = sub

	go rm.handleReplies()
	return rm, nil
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

// Request sends a request and waits for a response using the shared reply subscription.
func (rm *requestManager) Request(ctx context.Context, topic string, msg Message) (Message, error) {
	if rm.ps.closed.Load() {
		return Message{}, newErr(ErrCodeClosed, "request", topic, "broker is closed", nil)
	}

	correlationID := generateCorrelationID()
	if msg.Meta == nil {
		msg.Meta = make(map[string]string)
	}
	msg.Meta[ReplyToHeader] = rm.replyTopic
	msg.Meta[CorrelationIDHeader] = correlationID

	respCh := make(chan Message, 1)

	rm.mu.Lock()
	rm.pending[correlationID] = respCh
	rm.mu.Unlock()

	defer func() {
		rm.mu.Lock()
		delete(rm.pending, correlationID)
		rm.mu.Unlock()
	}()

	if err := rm.ps.Publish(topic, msg); err != nil {
		return Message{}, err
	}

	select {
	case resp := <-respCh:
		return resp, nil
	case <-rm.done:
		return Message{}, newErr(ErrCodeClosed, "request", topic, "broker is closed", nil)
	case <-ctx.Done():
		return Message{}, ctx.Err()
	}
}

// RequestWithTimeout sends a request with an explicit timeout.
func (rm *requestManager) RequestWithTimeout(topic string, msg Message, timeout time.Duration) (Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return rm.Request(ctx, topic, msg)
}

// handleReplies dispatches incoming reply messages to waiting Request calls.
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
			}
		}
	}
}

// Close shuts down the request manager.
func (rm *requestManager) Close() {
	if rm.replySub != nil {
		rm.replySub.Cancel()
	}
	close(rm.done)
	rm.mu.Lock()
	rm.pending = make(map[string]chan Message)
	rm.mu.Unlock()
}

// Reply sends a reply to a request message.
// The request message must contain ReplyToHeader and CorrelationIDHeader metadata.
func Reply(broker Broker, reqMsg Message, respMsg Message) error {
	replyTo := reqMsg.Meta[ReplyToHeader]
	if replyTo == "" {
		return newErr(ErrCodeNotFound, "reply", reqMsg.Topic, "no reply-to header in request", nil)
	}
	correlationID := reqMsg.Meta[CorrelationIDHeader]
	if correlationID == "" {
		return newErr(ErrCodeNotFound, "reply", reqMsg.Topic, "no correlation ID in request", nil)
	}
	if respMsg.Meta == nil {
		respMsg.Meta = make(map[string]string)
	}
	respMsg.Meta[CorrelationIDHeader] = correlationID
	return broker.Publish(replyTo, respMsg)
}

// IsRequest reports whether the message is a request expecting a reply.
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
