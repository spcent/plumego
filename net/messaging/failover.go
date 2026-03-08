package messaging

import (
	"context"
	"fmt"
)

// failoverSend implements the core failover loop for any provider type.
// It tries each sender in order; the first success wins. If all fail,
// it returns ErrProviderFailure wrapping the last error.
func failoverSend[M any, R any](ctx context.Context, senders []func(context.Context, M) (*R, error), msg M) (*R, error) {
	var lastErr error
	for _, send := range senders {
		result, err := send(ctx, msg)
		if err == nil {
			return result, nil
		}
		lastErr = err
		// Context cancelled means caller gave up; stop trying.
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}
	if lastErr != nil {
		return nil, fmt.Errorf("%w: all providers failed: %v", ErrProviderFailure, lastErr)
	}
	return nil, fmt.Errorf("%w: no providers configured", ErrProviderFailure)
}

// FailoverSMSProvider tries providers in order until one succeeds.
// This provides automatic failover when the primary gateway is down.
type FailoverSMSProvider struct {
	providers []SMSProvider
}

// NewFailoverSMSProvider creates a failover chain.
// Providers are tried in the order given; the first success wins.
func NewFailoverSMSProvider(providers ...SMSProvider) *FailoverSMSProvider {
	return &FailoverSMSProvider{providers: providers}
}

func (p *FailoverSMSProvider) Name() string {
	if len(p.providers) > 0 {
		return "failover(" + p.providers[0].Name() + ")"
	}
	return "failover(empty)"
}

func (p *FailoverSMSProvider) Send(ctx context.Context, msg SMSMessage) (*SMSResult, error) {
	senders := make([]func(context.Context, SMSMessage) (*SMSResult, error), len(p.providers))
	for i, prov := range p.providers {
		senders[i] = prov.Send
	}
	return failoverSend(ctx, senders, msg)
}

// FailoverEmailProvider tries providers in order until one succeeds.
type FailoverEmailProvider struct {
	providers []EmailProvider
}

// NewFailoverEmailProvider creates a failover chain.
func NewFailoverEmailProvider(providers ...EmailProvider) *FailoverEmailProvider {
	return &FailoverEmailProvider{providers: providers}
}

func (p *FailoverEmailProvider) Name() string {
	if len(p.providers) > 0 {
		return "failover(" + p.providers[0].Name() + ")"
	}
	return "failover(empty)"
}

func (p *FailoverEmailProvider) Send(ctx context.Context, msg EmailMessage) (*EmailResult, error) {
	senders := make([]func(context.Context, EmailMessage) (*EmailResult, error), len(p.providers))
	for i, prov := range p.providers {
		senders[i] = prov.Send
	}
	return failoverSend(ctx, senders, msg)
}
