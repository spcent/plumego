package messaging

import (
	"context"
	"fmt"
)

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
	var lastErr error
	for _, provider := range p.providers {
		result, err := provider.Send(ctx, msg)
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
	var lastErr error
	for _, provider := range p.providers {
		result, err := provider.Send(ctx, msg)
		if err == nil {
			return result, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}
	if lastErr != nil {
		return nil, fmt.Errorf("%w: all providers failed: %v", ErrProviderFailure, lastErr)
	}
	return nil, fmt.Errorf("%w: no providers configured", ErrProviderFailure)
}
