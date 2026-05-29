// Package email implements an SMTP-based AlertProvider using stdlib net/smtp.
//
// gatus uses gomail; we keep this dependency-free by sending plain RFC 5322
// messages with optional STARTTLS / direct TLS, controlled by Insecure.
package email

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"guardus/internal/alerting/provider"
	"guardus/internal/client"
	"guardus/internal/domain/alert"
	"guardus/internal/domain/endpoint"
)

var (
	ErrDuplicateGroupOverride = errors.New("duplicate group override")
	ErrMissingFromOrToFields  = errors.New("from and to fields are required")
	ErrInvalidPort            = errors.New("port must be between 1 and 65535 inclusively")
	ErrMissingHost            = errors.New("host is required")
)

type Config struct {
	From         string         `json:"from"`
	Username     string         `json:"username,omitempty"`
	Password     string         `json:"password,omitempty"`
	Host         string         `json:"host"`
	Port         int            `json:"port"`
	To           string         `json:"to"`
	ClientConfig *client.Config `json:"client,omitempty"`
}

func (cfg *Config) Validate() error {
	if len(cfg.From) == 0 || len(cfg.To) == 0 {
		return ErrMissingFromOrToFields
	}
	if cfg.Port < 1 || cfg.Port > math.MaxUint16 {
		return ErrInvalidPort
	}
	if len(cfg.Host) == 0 {
		return ErrMissingHost
	}
	return nil
}

func (cfg *Config) Merge(override *Config) {
	if override.ClientConfig != nil {
		cfg.ClientConfig = override.ClientConfig
	}
	if len(override.From) > 0 {
		cfg.From = override.From
	}
	if len(override.Username) > 0 {
		cfg.Username = override.Username
	}
	if len(override.Password) > 0 {
		cfg.Password = override.Password
	}
	if len(override.Host) > 0 {
		cfg.Host = override.Host
	}
	if override.Port > 0 {
		cfg.Port = override.Port
	}
	if len(override.To) > 0 {
		cfg.To = override.To
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

func (p *AlertProvider) GetAlertType() alert.Type { return alert.TypeEmail }

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
	subject, body := buildMessage(ep, a, r, resolved)
	username := cfg.Username
	if username == "" {
		username = cfg.From
	}
	to := splitAndTrim(cfg.To)
	msg := composeRFC822(cfg.From, to, subject, body)
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	insecure := cfg.ClientConfig != nil && cfg.ClientConfig.Insecure

	var auth smtp.Auth
	if cfg.Password != "" {
		auth = smtp.PlainAuth("", username, cfg.Password, cfg.Host)
	}

	timeout := 30 * time.Second
	if cfg.ClientConfig != nil && cfg.ClientConfig.Timeout > 0 {
		timeout = cfg.ClientConfig.Timeout
	}
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return err
	}
	c, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		conn.Close()
		return err
	}
	defer c.Close()
	if ok, _ := c.Extension("STARTTLS"); ok {
		tlsCfg := &tls.Config{ServerName: cfg.Host, InsecureSkipVerify: insecure}
		if err := c.StartTLS(tlsCfg); err != nil {
			return err
		}
	}
	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err := c.Auth(auth); err != nil {
				return err
			}
		}
	}
	if err := c.Mail(cfg.From); err != nil {
		return err
	}
	for _, addr := range to {
		if err := c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func composeRFC822(from string, to []string, subject, body string) []byte {
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + strings.Join(to, ", ") + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return []byte(b.String())
}

func buildMessage(ep *endpoint.Endpoint, a *alert.Alert, r *endpoint.Result, resolved bool) (string, string) {
	var subject, message string
	if resolved {
		subject = fmt.Sprintf("%s: Alert resolved", ep.DisplayName())
		message = fmt.Sprintf("An alert for %s has been resolved after passing successfully %d time(s) in a row", ep.DisplayName(), a.SuccessThreshold)
	} else {
		subject = fmt.Sprintf("%s: Alert triggered", ep.DisplayName())
		message = fmt.Sprintf("An alert for %s has been triggered due to having failed %d time(s) in a row", ep.DisplayName(), a.FailureThreshold)
	}
	var formatted string
	if len(r.ConditionResults) > 0 {
		formatted = "\n\nCondition results:\n"
		for _, cr := range r.ConditionResults {
			prefix := "❌"
			if cr.Success {
				prefix = "✅"
			}
			formatted += fmt.Sprintf("%s %s\n", prefix, cr.Condition)
		}
	}
	var description string
	if d := a.GetDescription(); len(d) > 0 {
		description = "\n\nAlert description: " + d
	}
	var labels string
	if len(ep.ExtraLabels) > 0 {
		labels = "\n\nExtra labels:\n"
		for k, v := range ep.ExtraLabels {
			labels += fmt.Sprintf("  %s: %s\n", k, v)
		}
	}
	return subject, message + description + labels + formatted
}

func (p *AlertProvider) ValidateOverrides(group string, a *alert.Alert) error {
	_, err := p.GetConfig(group, a)
	return err
}

var _ provider.AlertProvider = (*AlertProvider)(nil)
