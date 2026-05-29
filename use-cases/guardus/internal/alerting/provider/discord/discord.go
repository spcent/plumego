// Package discord implements an AlertProvider that posts incoming-webhook
// messages to Discord.
package discord

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
	WebhookURL     string `json:"webhook-url"`
	Title          string `json:"title,omitempty"`
	MessageContent string `json:"message-content,omitempty"`
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
	if len(override.MessageContent) > 0 {
		cfg.MessageContent = override.MessageContent
	}
}

type AlertProvider struct {
	DefaultConfig Config       `json:"default-config,omitempty"`
	DefaultAlert  *alert.Alert `json:"default-alert,omitempty"`
	Overrides     []Override   `json:"overrides,omitempty"`
}

type Override struct {
	Group  string `json:"group"`
	Config        // embedded; JSON flattens anonymous fields automatically
}

func (p *AlertProvider) GetAlertType() alert.Type { return alert.TypeDiscord }

func (p *AlertProvider) Validate() error {
	registered := make(map[string]bool)
	for _, o := range p.Overrides {
		if registered[o.Group] || o.Group == "" || len(o.WebhookURL) == 0 {
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
		return fmt.Errorf("discord alert returned status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

type body struct {
	Content string  `json:"content"`
	Embeds  []embed `json:"embeds"`
}

type embed struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Color       int     `json:"color"`
	Fields      []field `json:"fields,omitempty"`
}

type field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

func buildRequestBody(cfg *Config, ep *endpoint.Endpoint, a *alert.Alert, r *endpoint.Result, resolved bool) []byte {
	var message string
	var colorCode int
	if resolved {
		message = fmt.Sprintf("An alert for **%s** has been resolved after passing successfully %d time(s) in a row", ep.DisplayName(), a.SuccessThreshold)
		colorCode = 3066993
	} else {
		message = fmt.Sprintf("An alert for **%s** has been triggered due to having failed %d time(s) in a row", ep.DisplayName(), a.FailureThreshold)
		colorCode = 15158332
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
	title := ":helmet_with_white_cross: Gatus"
	if cfg.Title != "" {
		title = cfg.Title
	}
	b := body{
		Content: cfg.MessageContent,
		Embeds: []embed{{
			Title:       title,
			Description: message + description,
			Color:       colorCode,
		}},
	}
	if formatted != "" {
		b.Embeds[0].Fields = append(b.Embeds[0].Fields, field{Name: "Condition results", Value: formatted})
	}
	out, _ := json.Marshal(b)
	return out
}

func (p *AlertProvider) ValidateOverrides(group string, a *alert.Alert) error {
	_, err := p.GetConfig(group, a)
	return err
}

var _ provider.AlertProvider = (*AlertProvider)(nil)
