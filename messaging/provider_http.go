package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	phttp "github.com/spcent/plumego/net/http"
)

// HTTPSMSProvider implements SMSProvider using an HTTP API gateway.
type HTTPSMSProvider struct {
	client  *phttp.Client
	baseURL string
	apiKey  string
}

// HTTPSMSProviderConfig configures the HTTP SMS provider.
type HTTPSMSProviderConfig struct {
	BaseURL    string
	APIKey     string
	Timeout    time.Duration
	MaxRetries int
}

// NewHTTPSMSProvider creates an SMS provider backed by an HTTP API.
func NewHTTPSMSProvider(cfg HTTPSMSProviderConfig) *HTTPSMSProvider {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 2
	}
	client := phttp.New(
		phttp.WithTimeout(timeout),
		phttp.WithRetryCount(maxRetries),
		phttp.WithDefaultSSRFProtection(),
		phttp.WithRetryPolicy(phttp.StatusCodeRetryPolicy{
			Codes: []int{500, 502, 503, 504, 429},
		}),
	)
	return &HTTPSMSProvider{
		client:  client,
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
	}
}

func (p *HTTPSMSProvider) Name() string { return "http-sms" }

func (p *HTTPSMSProvider) Send(ctx context.Context, msg SMSMessage) (*SMSResult, error) {
	body, err := p.client.PostJson(ctx, p.baseURL+"/send", map[string]string{
		"to":   msg.To,
		"body": msg.Body,
	}, phttp.WithHeader("Authorization", "Bearer "+p.apiKey))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderFailure, err)
	}
	var resp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("%w: failed to parse response: %v", ErrProviderFailure, err)
	}
	return &SMSResult{ProviderID: resp.ID}, nil
}

// HTTPEmailProvider implements EmailProvider using an HTTP API gateway.
type HTTPEmailProvider struct {
	client  *phttp.Client
	baseURL string
	apiKey  string
}

// HTTPEmailProviderConfig configures the HTTP email provider.
type HTTPEmailProviderConfig struct {
	BaseURL    string
	APIKey     string
	Timeout    time.Duration
	MaxRetries int
}

// NewHTTPEmailProvider creates an email provider backed by an HTTP API.
func NewHTTPEmailProvider(cfg HTTPEmailProviderConfig) *HTTPEmailProvider {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 2
	}
	client := phttp.New(
		phttp.WithTimeout(timeout),
		phttp.WithRetryCount(maxRetries),
		phttp.WithDefaultSSRFProtection(),
		phttp.WithRetryPolicy(phttp.StatusCodeRetryPolicy{
			Codes: []int{500, 502, 503, 504, 429},
		}),
	)
	return &HTTPEmailProvider{
		client:  client,
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
	}
}

func (p *HTTPEmailProvider) Name() string { return "http-email" }

func (p *HTTPEmailProvider) Send(ctx context.Context, msg EmailMessage) (*EmailResult, error) {
	contentType := "text/plain"
	if msg.HTML {
		contentType = "text/html"
	}
	body, err := p.client.PostJson(ctx, p.baseURL+"/send", map[string]string{
		"to":           msg.To,
		"subject":      msg.Subject,
		"body":         msg.Body,
		"content_type": contentType,
	}, phttp.WithHeader("Authorization", "Bearer "+p.apiKey))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderFailure, err)
	}
	var resp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("%w: failed to parse response: %v", ErrProviderFailure, err)
	}
	return &EmailResult{MessageID: resp.ID}, nil
}
