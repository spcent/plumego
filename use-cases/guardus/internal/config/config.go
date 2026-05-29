// Package config loads and validates the guardus configuration.
//
// Deployment configuration (web, storage, security, UI, alerting providers,
// connectivity, maintenance) is sourced from process environment variables —
// see env.go and env.example. Endpoints live in the storage layer (sqlite or
// memory) and are loaded through storage.Store.ListEndpointConfigs after env
// load; the optional GUARDUS_BOOTSTRAP_FILE path can seed them on first
// start.
package config

import (
	"errors"
	"slices"
	"sort"
	"strings"

	"github.com/spcent/plumego/core"
	plumelog "github.com/spcent/plumego/log"

	"guardus/internal/alerting"
	"guardus/internal/domain/connectivity"
	"guardus/internal/domain/endpoint"
	"guardus/internal/domain/maintenance"
	"guardus/internal/storage"
)

const (
	// DefaultConcurrency is the default number of endpoints monitored concurrently.
	DefaultConcurrency = 3
)

var (
	// ErrNoEndpointInConfig is returned when there is nothing to monitor.
	ErrNoEndpointInConfig = errors.New("configuration should contain at least one endpoint")
	// ErrInvalidSecurityConfig is returned when the security block is invalid.
	ErrInvalidSecurityConfig = errors.New("invalid security configuration")
)

// Config is the root guardus configuration.
type Config struct {
	// Metrics enables /metrics exposition.
	Metrics bool `json:"metrics,omitempty"`

	// Concurrency caps in-flight endpoint probes. Defaults to DefaultConcurrency.
	// Set to 0 for unlimited.
	Concurrency int `json:"concurrency,omitempty"`

	Security          *SecurityConfig              `json:"security,omitempty"`
	Alerting          *alerting.Config             `json:"alerting,omitempty"`
	Endpoints         []*endpoint.Endpoint         `json:"endpoints,omitempty"`
	ExternalEndpoints []*endpoint.ExternalEndpoint `json:"external-endpoints,omitempty"`
	Storage           *storage.Config              `json:"storage,omitempty"`
	Web               *WebConfig                   `json:"web,omitempty"`
	UI                *UIConfig                    `json:"ui,omitempty"`
	Maintenance       *maintenance.Config          `json:"maintenance,omitempty"`
	Connectivity      *connectivity.Config         `json:"connectivity,omitempty"`

	// Core holds plumego AppConfig built from Web/Address/Port/TLS.
	Core core.AppConfig `json:"-"`

	// Version is set at build time and threaded through to handlers.
	Version string `json:"-"`
}

// Defaults returns the in-memory default configuration.
//
// Used as a sane baseline; consumers must add at least one endpoint via
// storage.Store.UpsertEndpointConfig (or the bootstrap file) before serving
// useful traffic.
func Defaults() Config {
	c := Config{
		Web:          defaultWebConfig(),
		UI:           defaultUIConfig(),
		Storage:      &storage.Config{Type: storage.TypeMemory},
		Maintenance:  maintenance.GetDefaultConfig(),
		Connectivity: nil,
		Concurrency:  DefaultConcurrency,
		Version:      "dev",
	}
	c.applyToCore()
	return c
}

// Load returns the env-driven configuration. logger is forwarded to validation
// so providers and endpoint counts are surfaced through the standard logging
// path; it must not be nil.
func Load(logger plumelog.StructuredLogger) (Config, error) {
	cfg, err := LoadFromEnv(logger)
	if err != nil {
		return Config{}, err
	}
	return *cfg, nil
}

// GetEndpointByKey returns the Endpoint with the given key (lowercased), or nil.
func (c *Config) GetEndpointByKey(key string) *endpoint.Endpoint {
	target := strings.ToLower(key)
	for _, ep := range c.Endpoints {
		if ep.Key() == target {
			return ep
		}
	}
	return nil
}

// GetExternalEndpointByKey returns the ExternalEndpoint with the given key, or nil.
func (c *Config) GetExternalEndpointByKey(key string) *endpoint.ExternalEndpoint {
	target := strings.ToLower(key)
	for _, ee := range c.ExternalEndpoints {
		if ee.Key() == target {
			return ee
		}
	}
	return nil
}

// GetUniqueExtraMetricLabels returns a sorted, deduplicated slice of labels
// declared on enabled endpoints.
func (c *Config) GetUniqueExtraMetricLabels() []string {
	labels := make([]string, 0)
	for _, ep := range c.Endpoints {
		if !ep.IsEnabled() {
			continue
		}
		for label := range ep.ExtraLabels {
			if !slices.Contains(labels, label) {
				labels = append(labels, label)
			}
		}
	}
	if len(labels) > 1 {
		sort.Strings(labels)
	}
	return labels
}

// applyToCore translates Web/TLS into a plumego core.AppConfig.
func (c *Config) applyToCore() {
	if c.Web == nil {
		c.Web = defaultWebConfig()
		_ = c.Web.ValidateAndSetDefaults()
	}
	core := core.DefaultConfig()
	core.Addr = c.Web.SocketAddress()
	if c.Web.HasTLS() {
		core.TLS.Enabled = true
		core.TLS.CertFile = c.Web.TLS.CertificateFile
		core.TLS.KeyFile = c.Web.TLS.PrivateKeyFile
	}
	c.Core = core
}
