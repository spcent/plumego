// Package custom implements a generic HTTP-based AlertProvider that can be
// configured to call any webhook (acts as v1's "webhook" provider).
package custom

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"guardus/internal/alerting/provider"
	"guardus/internal/client"
	"guardus/internal/domain/alert"
	"guardus/internal/domain/endpoint"
)

var ErrURLNotSet = errors.New("url not set")

type Config struct {
	URL          string                       `json:"url"`
	Method       string                       `json:"method,omitempty"`
	Body         string                       `json:"body,omitempty"`
	Headers      map[string]string            `json:"headers,omitempty"`
	Placeholders map[string]map[string]string `json:"placeholders,omitempty"`
	ClientConfig *client.Config               `json:"client,omitempty"`
}

func (cfg *Config) Validate() error {
	if len(cfg.URL) == 0 {
		return ErrURLNotSet
	}
	return nil
}

func (cfg *Config) Merge(override *Config) {
	if override.ClientConfig != nil {
		cfg.ClientConfig = override.ClientConfig
	}
	if len(override.URL) > 0 {
		cfg.URL = override.URL
	}
	if len(override.Method) > 0 {
		cfg.Method = override.Method
	}
	if len(override.Body) > 0 {
		cfg.Body = override.Body
	}
	if len(override.Headers) > 0 {
		cfg.Headers = override.Headers
	}
	if len(override.Placeholders) > 0 {
		cfg.Placeholders = override.Placeholders
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

func (p *AlertProvider) GetAlertType() alert.Type { return alert.TypeCustom }

func (p *AlertProvider) Validate() error { return p.DefaultConfig.Validate() }

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
	req := buildHTTPRequest(cfg, ep, a, r, resolved)
	resp, err := client.GetHTTPClient(cfg.ClientConfig).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 399 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("custom alert returned status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func buildHTTPRequest(cfg *Config, ep *endpoint.Endpoint, a *alert.Alert, r *endpoint.Result, resolved bool) *http.Request {
	body, url, method := cfg.Body, cfg.URL, cfg.Method
	repl := func(s string) string {
		s = strings.ReplaceAll(s, "[ALERT_DESCRIPTION]", a.GetDescription())
		s = strings.ReplaceAll(s, "[ENDPOINT_NAME]", ep.Name)
		s = strings.ReplaceAll(s, "[ENDPOINT_GROUP]", ep.Group)
		s = strings.ReplaceAll(s, "[ENDPOINT_URL]", ep.URL)
		errs := strings.ReplaceAll(strings.Join(r.Errors, ","), "\"", "\\\"")
		s = strings.ReplaceAll(s, "[RESULT_ERRORS]", errs)
		s = strings.ReplaceAll(s, "[ALERT_TRIGGERED_OR_RESOLVED]", alertStatePlaceholder(cfg, resolved))
		return s
	}
	body = repl(body)
	url = repl(url)
	if len(r.ConditionResults) > 0 && strings.Contains(body, "[RESULT_CONDITIONS]") {
		var formatted string
		for i, cr := range r.ConditionResults {
			prefix := "❌"
			if cr.Success {
				prefix = "✅"
			}
			formatted += fmt.Sprintf("%s - `%s`", prefix, cr.Condition)
			if i < len(r.ConditionResults)-1 {
				formatted += ", "
			}
		}
		body = strings.ReplaceAll(body, "[RESULT_CONDITIONS]", formatted)
		url = strings.ReplaceAll(url, "[RESULT_CONDITIONS]", formatted)
	}
	if method == "" {
		method = http.MethodGet
	}
	req, _ := http.NewRequest(method, url, bytes.NewBufferString(body))
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}
	return req
}

func alertStatePlaceholder(cfg *Config, resolved bool) string {
	status := "TRIGGERED"
	if resolved {
		status = "RESOLVED"
	}
	if m, ok := cfg.Placeholders["ALERT_TRIGGERED_OR_RESOLVED"]; ok {
		if v, ok := m[status]; ok {
			return v
		}
	}
	return status
}

func (p *AlertProvider) ValidateOverrides(group string, a *alert.Alert) error {
	_, err := p.GetConfig(group, a)
	return err
}

var _ provider.AlertProvider = (*AlertProvider)(nil)
