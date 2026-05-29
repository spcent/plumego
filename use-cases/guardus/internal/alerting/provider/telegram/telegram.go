// Package telegram implements an AlertProvider that posts alerts via the
// Telegram Bot API.
package telegram

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

const APIURL = "https://api.telegram.org"

var (
	ErrTokenNotSet            = errors.New("token not set")
	ErrIDNotSet               = errors.New("id not set")
	ErrDuplicateGroupOverride = errors.New("duplicate group override")
)

type Config struct {
	Token        string         `json:"token"`
	ID           string         `json:"id"`
	TopicID      string         `json:"topic-id,omitempty"`
	APIURL       string         `json:"api-url"`
	ClientConfig *client.Config `json:"client,omitempty"`
}

func (cfg *Config) Validate() error {
	if len(cfg.APIURL) == 0 {
		cfg.APIURL = APIURL
	}
	if len(cfg.Token) == 0 {
		return ErrTokenNotSet
	}
	if len(cfg.ID) == 0 {
		return ErrIDNotSet
	}
	return nil
}

func (cfg *Config) Merge(override *Config) {
	if override.ClientConfig != nil {
		cfg.ClientConfig = override.ClientConfig
	}
	if len(override.Token) > 0 {
		cfg.Token = override.Token
	}
	if len(override.ID) > 0 {
		cfg.ID = override.ID
	}
	if len(override.TopicID) > 0 {
		cfg.TopicID = override.TopicID
	}
	if len(override.APIURL) > 0 {
		cfg.APIURL = override.APIURL
	}
}

type AlertProvider struct {
	DefaultConfig Config       `json:"default-config,omitempty"`
	DefaultAlert  *alert.Alert `json:"default-alert,omitempty"`
	Overrides     []*Override  `json:"overrides,omitempty"`
}

type Override struct {
	Group  string `json:"group"`
	Config        // embedded; JSON flattens anonymous fields automatically
}

func (p *AlertProvider) GetAlertType() alert.Type { return alert.TypeTelegram }

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
	url := fmt.Sprintf("%s/bot%s/sendMessage", cfg.APIURL, cfg.Token)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.GetHTTPClient(cfg.ClientConfig).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 399 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram alert returned status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

type body struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
	TopicID   string `json:"message_thread_id,omitempty"`
}

func buildRequestBody(cfg *Config, ep *endpoint.Endpoint, a *alert.Alert, r *endpoint.Result, resolved bool) []byte {
	var message string
	if resolved {
		message = fmt.Sprintf("An alert for *%s* has been resolved:\n—\n    _healthcheck passing successfully %d time(s) in a row_\n—  ", ep.DisplayName(), a.SuccessThreshold)
	} else {
		message = fmt.Sprintf("An alert for *%s* has been triggered:\n—\n    _healthcheck failed %d time(s) in a row_\n—  ", ep.DisplayName(), a.FailureThreshold)
	}
	var formatted string
	if len(r.ConditionResults) > 0 {
		formatted = "\n*Condition results*\n"
		for _, cr := range r.ConditionResults {
			prefix := "❌"
			if cr.Success {
				prefix = "✅"
			}
			formatted += fmt.Sprintf("%s - `%s`\n", prefix, cr.Condition)
		}
	}
	var text string
	if d := a.GetDescription(); len(d) > 0 {
		text = fmt.Sprintf("⛑ *Gatus* \n%s \n*Description* \n%s  \n%s", message, d, formatted)
	} else {
		text = fmt.Sprintf("⛑ *Gatus* \n%s%s", message, formatted)
	}
	out, _ := json.Marshal(body{
		ChatID:    cfg.ID,
		Text:      text,
		ParseMode: "MARKDOWN",
		TopicID:   cfg.TopicID,
	})
	return out
}

func (p *AlertProvider) ValidateOverrides(group string, a *alert.Alert) error {
	_, err := p.GetConfig(group, a)
	return err
}

var _ provider.AlertProvider = (*AlertProvider)(nil)
