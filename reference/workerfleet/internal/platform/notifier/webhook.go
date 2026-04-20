package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"workerfleet/internal/domain"
)

type WebhookConfig struct {
	URL        string
	Headers    map[string]string
	HTTPClient *http.Client
}

type WebhookNotifier struct {
	url        string
	headers    map[string]string
	httpClient *http.Client
}

type webhookEnvelope struct {
	Alert domain.AlertRecord `json:"alert"`
}

func NewWebhookNotifier(cfg WebhookConfig) *WebhookNotifier {
	client := cfg.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	headers := make(map[string]string, len(cfg.Headers))
	for key, value := range cfg.Headers {
		headers[key] = value
	}
	return &WebhookNotifier{
		url:        strings.TrimSpace(cfg.URL),
		headers:    headers,
		httpClient: client,
	}
}

func (n *WebhookNotifier) Notify(ctx context.Context, alert domain.AlertRecord) error {
	if n.url == "" {
		return fmt.Errorf("webhook URL is required")
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(webhookEnvelope{Alert: alert}); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.url, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range n.headers {
		req.Header.Set(key, value)
	}

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("webhook notify failed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}

	return nil
}
