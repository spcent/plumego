// Package alerting holds the alerting configuration: typed fields per
// provider and an explicit type→provider switch (gatus uses reflection; we
// keep the lookup boring on purpose).
package alerting

import (
	"guardus/internal/alerting/provider"
	"guardus/internal/alerting/provider/custom"
	"guardus/internal/alerting/provider/discord"
	"guardus/internal/alerting/provider/email"
	"guardus/internal/alerting/provider/slack"
	"guardus/internal/alerting/provider/telegram"
	"guardus/internal/domain/alert"
)

// Config is the alerting configuration. Each non-nil field corresponds to a
// configured provider.
type Config struct {
	Custom   *custom.AlertProvider   `json:"custom,omitempty"`
	Discord  *discord.AlertProvider  `json:"discord,omitempty"`
	Email    *email.AlertProvider    `json:"email,omitempty"`
	Slack    *slack.AlertProvider    `json:"slack,omitempty"`
	Telegram *telegram.AlertProvider `json:"telegram,omitempty"`
}

// SupportedAlertTypes returns the alert types validated and dispatched by
// guardus v1.
func SupportedAlertTypes() []alert.Type {
	return []alert.Type{
		alert.TypeCustom,
		alert.TypeDiscord,
		alert.TypeEmail,
		alert.TypeSlack,
		alert.TypeTelegram,
	}
}

// GetAlertingProviderByAlertType returns the provider configured for the given
// alert.Type, or nil if no provider is configured (or unsupported in v1).
func (c *Config) GetAlertingProviderByAlertType(t alert.Type) provider.AlertProvider {
	if c == nil {
		return nil
	}
	switch t {
	case alert.TypeCustom:
		if c.Custom != nil {
			return c.Custom
		}
	case alert.TypeDiscord:
		if c.Discord != nil {
			return c.Discord
		}
	case alert.TypeEmail:
		if c.Email != nil {
			return c.Email
		}
	case alert.TypeSlack:
		if c.Slack != nil {
			return c.Slack
		}
	case alert.TypeTelegram:
		if c.Telegram != nil {
			return c.Telegram
		}
	}
	return nil
}

// SetAlertingProviderToNil clears the field whose provider matches p so that
// validation only happens once per alert dispatch.
func (c *Config) SetAlertingProviderToNil(p provider.AlertProvider) {
	if c == nil || p == nil {
		return
	}
	switch p.GetAlertType() {
	case alert.TypeCustom:
		c.Custom = nil
	case alert.TypeDiscord:
		c.Discord = nil
	case alert.TypeEmail:
		c.Email = nil
	case alert.TypeSlack:
		c.Slack = nil
	case alert.TypeTelegram:
		c.Telegram = nil
	}
}

// ValidateOverrider is implemented by providers that support per-endpoint or
// per-group overrides. Optional — most providers implement it.
type ValidateOverrider interface {
	ValidateOverrides(group string, a *alert.Alert) error
}
