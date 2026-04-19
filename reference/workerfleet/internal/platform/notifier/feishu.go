package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spcent/plumego/reference/workerfleet/internal/domain"
)

type FeishuConfig struct {
	WebhookURL string
	HTTPClient *http.Client
}

type FeishuNotifier struct {
	webhookURL string
	httpClient *http.Client
}

type feishuTextMessage struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
}

func NewFeishuNotifier(cfg FeishuConfig) *FeishuNotifier {
	client := cfg.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	return &FeishuNotifier{
		webhookURL: strings.TrimSpace(cfg.WebhookURL),
		httpClient: client,
	}
}

func (n *FeishuNotifier) Notify(ctx context.Context, alert domain.AlertRecord) error {
	if n.webhookURL == "" {
		return fmt.Errorf("feishu webhook URL is required")
	}

	message := feishuTextMessage{MsgType: "text"}
	message.Content.Text = renderAlertText(alert)

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(message); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.webhookURL, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("feishu notify failed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}

	return nil
}
