package pipeline

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/spcent/plumego/examples/sms-gateway/internal/tasks"
)

var ErrProviderUnavailable = errors.New("provider unavailable")

// ProviderSender sends SMS via a provider.
type ProviderSender interface {
	Send(ctx context.Context, payload tasks.SendTaskPayload) error
}

// MockProvider simulates a provider with optional failures.
type MockProvider struct {
	Name       string
	FailFirstN int

	mu       sync.Mutex
	attempts map[string]int
}

func (p *MockProvider) Send(ctx context.Context, payload tasks.SendTaskPayload) error {
	if p == nil {
		return ErrProviderUnavailable
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.attempts == nil {
		p.attempts = make(map[string]int)
	}
	attempt := p.attempts[payload.MessageID] + 1
	p.attempts[payload.MessageID] = attempt
	if p.FailFirstN > 0 && attempt <= p.FailFirstN {
		return fmt.Errorf("provider %s simulated failure", p.Name)
	}
	return nil
}
