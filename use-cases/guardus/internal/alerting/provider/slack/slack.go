// Package slack implements an AlertProvider that posts incoming-webhook
// messages to Slack.
package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"guardus/internal/alerting/provider"
	"guardus/internal/client"
	"guardus/internal/domain/alert"
	"guardus/internal/domain/endpoint"
)

var (
	ErrWebhookURLNotSet       = errors.New("webhook-url not set")
	ErrDuplicateGroupOverride = errors.New("duplicate group override")
)

type Config struct {
	WebhookURL string `json:"webhook-url"`
	Title      string `json:"title,omitempty"`
}

func (cfg *Config) Validate() error {
	if len(cfg.WebhookURL) == 0 {
		return ErrWebhookURLNotSet
	}
	return nil
}

func (cfg *Config) Merge(override *Config) {
	if len(override.WebhookURL) > 0 {
		cfg.WebhookURL = override.WebhookURL
	}
	if len(override.Title) > 0 {
		cfg.Title = override.Title
	}
}

// AlertProvider is the slack-specific alerting provider configuration.
type AlertProvider struct {
	DefaultConfig Config       `json:"default-config,omitempty"`
	DefaultAlert  *alert.Alert `json:"default-alert,omitempty"`
	Overrides     []Override   `json:"overrides,omitempty"`
}

type Override struct {
	Group  string `json:"group"`
	Config        // embedded; JSON flattens anonymous fields automatically
}

func (p *AlertProvider) GetAlertType() alert.Type { return alert.TypeSlack }

func (p *AlertProvider) Validate() error {
	registered := make(map[string]bool)
	for _, o := range p.Overrides {
		if registered[o.Group] || o.Group == "" {
			return ErrDuplicateGroupOverride
		}
		registered[o.Group] = true
	}
	return p.DefaultConfig.Validate()
}

func (p *AlertProvider) GetDefaultAlert() *alert.Alert { return p.DefaultAlert }

func (p *AlertProvider) GetConfig(group string, a *alert.Alert) (*Config, error) {
	cfg := p.DefaultConfig
	for _, o := range p.Overrides {
		if group == o.Group {
			cfg.Merge(&o.Config)
			break
		}
	}
	if len(a.ProviderOverride) != 0 {
		overrideCfg := Config{}
		if err := json.Unmarshal(a.ProviderOverrideAsBytes(), &overrideCfg); err != nil {
			return nil, err
		}
		cfg.Merge(&overrideCfg)
	}
	return &cfg, cfg.Validate()
}

func (p *AlertProvider) Send(ep *endpoint.Endpoint, a *alert.Alert, r *endpoint.Result, resolved bool) error {
	cfg, err := p.GetConfig(ep.Group, a)
	if err != nil {
		return err
	}
	body := buildRequestBody(cfg, ep, a, r, resolved)
	req, err := http.NewRequest(http.MethodPost, cfg.WebhookURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.GetHTTPClient(nil).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 399 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack alert returned status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

type body struct {
	Text        string       `json:"text"`
	Attachments []attachment `json:"attachments"`
}

type attachment struct {
	Title  string  `json:"title"`
	Text   string  `json:"text"`
	Short  bool    `json:"short"`
	Color  string  `json:"color"`
	Fields []field `json:"fields,omitempty"`
}

type field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

func buildRequestBody(cfg *Config, ep *endpoint.Endpoint, a *alert.Alert, r *endpoint.Result, resolved bool) []byte {
	var message, color string
	if resolved {
		message = fmt.Sprintf("An alert for *%s* has been resolved after passing successfully %d time(s) in a row", ep.DisplayName(), a.SuccessThreshold)
		color = "#36A64F"
	} else {
		message = fmt.Sprintf("An alert for *%s* has been triggered due to having failed %d time(s) in a row", ep.DisplayName(), a.FailureThreshold)
		color = "#DD0000"
	}
	var formatted string
	for _, cr := range r.ConditionResults {
		prefix := ":x:"
		if cr.Success {
			prefix = ":white_check_mark:"
		}
		formatted += fmt.Sprintf("%s - `%s`\n", prefix, cr.Condition)
	}
	var description string
	if d := a.GetDescription(); len(d) > 0 {
		description = ":\n> " + d
	}
	b := body{
		Attachments: []attachment{{
			Title: cfg.Title,
			Text:  message + description,
			Color: color,
		}},
	}
	if b.Attachments[0].Title == "" {
		b.Attachments[0].Title = ":helmet_with_white_cross: Gatus"
	}
	if formatted != "" {
		b.Attachments[0].Fields = append(b.Attachments[0].Fields, field{Title: "Condition results", Value: formatted})
	}
	out, _ := json.Marshal(b)
	return out
}

func (p *AlertProvider) ValidateOverrides(group string, a *alert.Alert) error {
	_, err := p.GetConfig(group, a)
	return err
}

// Compile-time check that AlertProvider satisfies provider.AlertProvider.
var _ provider.AlertProvider = (*AlertProvider)(nil)
