package config

import (
	"fmt"

	plumelog "github.com/spcent/plumego/log"

	"guardus/internal/alerting"
	"guardus/internal/alerting/provider"
	"guardus/internal/domain/alert"
	"guardus/internal/domain/endpoint"
	"guardus/internal/domain/maintenance"
	"guardus/internal/storage"
)

// ValidateAlertingConfig validates each enabled alerting provider, prunes
// invalid ones, and merges per-provider default alerts into endpoint alerts.
//
// Note: must run before ValidateEndpointsConfig so that
// endpoint.ValidateAndSetDefaults() sees fully-merged Alert entries. logger
// must not be nil.
func ValidateAlertingConfig(alertingCfg *alerting.Config, endpoints []*endpoint.Endpoint, externalEndpoints []*endpoint.ExternalEndpoint, logger plumelog.StructuredLogger) {
	if alertingCfg == nil {
		logger.Info("ValidateAlertingConfig: alerting is not configured", plumelog.Fields{"op": "config.ValidateAlertingConfig"})
		return
	}
	var validProviders, invalidProviders []alert.Type
	for _, alertType := range alerting.SupportedAlertTypes() {
		ap := alertingCfg.GetAlertingProviderByAlertType(alertType)
		if ap == nil {
			invalidProviders = append(invalidProviders, alertType)
			continue
		}
		if err := ap.Validate(); err != nil {
			logger.Warn("ValidateAlertingConfig: ignoring provider", plumelog.Fields{"op": "config.ValidateAlertingConfig", "provider": string(alertType), "err": err.Error()})
			invalidProviders = append(invalidProviders, alertType)
			alertingCfg.SetAlertingProviderToNil(ap)
			continue
		}
		mergeDefaultAlerts(ap, alertType, endpoints, externalEndpoints, logger)
		validProviders = append(validProviders, alertType)
	}
	logger.Info("ValidateAlertingConfig: provider summary", plumelog.Fields{
		"op":                  "config.ValidateAlertingConfig",
		"configuredProviders": fmt.Sprint(validProviders),
		"ignoredProviders":    fmt.Sprint(invalidProviders),
	})
}

func mergeDefaultAlerts(ap provider.AlertProvider, alertType alert.Type, endpoints []*endpoint.Endpoint, externalEndpoints []*endpoint.ExternalEndpoint, logger plumelog.StructuredLogger) {
	defaultAlert := ap.GetDefaultAlert()
	if defaultAlert == nil {
		return
	}
	for _, ep := range endpoints {
		for _, ea := range ep.Alerts {
			if alertType != ea.Type {
				continue
			}
			provider.MergeProviderDefaultAlertIntoEndpointAlert(defaultAlert, ea)
			if len(ea.ProviderOverride) > 0 {
				if err := ap.ValidateOverrides(ep.Group, ea); err != nil {
					logger.Warn("ValidateAlertingConfig: invalid endpoint overrides", plumelog.Fields{"op": "config.ValidateAlertingConfig", "endpoint": ep.Key(), "provider": string(alertType), "err": err.Error()})
				}
			}
		}
	}
	for _, ee := range externalEndpoints {
		for _, ea := range ee.Alerts {
			if alertType != ea.Type {
				continue
			}
			provider.MergeProviderDefaultAlertIntoEndpointAlert(defaultAlert, ea)
			if len(ea.ProviderOverride) > 0 {
				if err := ap.ValidateOverrides(ee.Group, ea); err != nil {
					logger.Warn("ValidateAlertingConfig: invalid external-endpoint overrides", plumelog.Fields{"op": "config.ValidateAlertingConfig", "external_endpoint": ee.Key(), "provider": string(alertType), "err": err.Error()})
				}
			}
		}
	}
}

// ValidateSecurityConfig returns ErrInvalidSecurityConfig when the basic
// configuration is incomplete.
func ValidateSecurityConfig(cfg *Config) error {
	if cfg.Security == nil {
		return nil
	}
	if !cfg.Security.ValidateAndSetDefaults() {
		return ErrInvalidSecurityConfig
	}
	return nil
}

// ValidateEndpointsConfig validates each endpoint and external endpoint and
// rejects duplicate keys. logger must not be nil.
func ValidateEndpointsConfig(cfg *Config, logger plumelog.StructuredLogger) error {
	seen := make(map[string]bool)
	for _, ep := range cfg.Endpoints {
		if seen[ep.Key()] {
			return fmt.Errorf("invalid endpoint %s: name and group combination must be unique", ep.Key())
		}
		seen[ep.Key()] = true
		if err := ep.ValidateAndSetDefaults(); err != nil {
			return fmt.Errorf("invalid endpoint %s: %w", ep.Key(), err)
		}
	}
	for _, ee := range cfg.ExternalEndpoints {
		if seen[ee.Key()] {
			return fmt.Errorf("invalid external endpoint %s: name and group combination must be unique", ee.Key())
		}
		seen[ee.Key()] = true
		if err := ee.ValidateAndSetDefaults(); err != nil {
			return fmt.Errorf("invalid external endpoint %s: %w", ee.Key(), err)
		}
	}
	logger.Info("ValidateEndpointsConfig: validation summary", plumelog.Fields{
		"op":                 "config.ValidateEndpointsConfig",
		"endpoints":          len(cfg.Endpoints),
		"external_endpoints": len(cfg.ExternalEndpoints),
	})
	return nil
}

// ValidateWebConfig fills in web defaults.
func ValidateWebConfig(cfg *Config) error {
	if cfg.Web == nil {
		cfg.Web = defaultWebConfig()
	}
	return cfg.Web.ValidateAndSetDefaults()
}

// ValidateUIConfig fills in UI defaults.
func ValidateUIConfig(cfg *Config) error {
	if cfg.UI == nil {
		cfg.UI = defaultUIConfig()
		return nil
	}
	return cfg.UI.ValidateAndSetDefaults()
}

// ValidateMaintenanceConfig fills in maintenance defaults.
func ValidateMaintenanceConfig(cfg *Config) error {
	if cfg.Maintenance == nil {
		cfg.Maintenance = maintenance.GetDefaultConfig()
		return nil
	}
	return cfg.Maintenance.ValidateAndSetDefaults()
}

// ValidateStorageConfig fills in storage defaults.
func ValidateStorageConfig(cfg *Config) error {
	if cfg.Storage == nil {
		cfg.Storage = &storage.Config{Type: storage.TypeMemory}
	}
	return cfg.Storage.ValidateAndSetDefaults()
}

// ValidateConnectivityConfig validates the optional connectivity checker.
func ValidateConnectivityConfig(cfg *Config) error {
	if cfg.Connectivity == nil {
		return nil
	}
	return cfg.Connectivity.ValidateAndSetDefaults()
}

// ValidateUniqueKeys ensures endpoint keys do not collide with external endpoint keys.
func ValidateUniqueKeys(cfg *Config) error {
	seen := make(map[string]string)
	for _, ep := range cfg.Endpoints {
		k := ep.Key()
		if existing, ok := seen[k]; ok {
			return fmt.Errorf("duplicate key '%s': endpoint conflicts with %s", k, existing)
		}
		seen[k] = fmt.Sprintf("endpoint '%s'", k)
	}
	for _, ee := range cfg.ExternalEndpoints {
		k := ee.Key()
		if existing, ok := seen[k]; ok {
			return fmt.Errorf("duplicate key '%s': external endpoint conflicts with %s", k, existing)
		}
		seen[k] = fmt.Sprintf("external endpoint '%s'", k)
	}
	return nil
}

// ValidateAndSetConcurrencyDefaults clamps Concurrency to a sane default.
func ValidateAndSetConcurrencyDefaults(cfg *Config) {
	if cfg.Concurrency < 0 {
		cfg.Concurrency = 0
		return
	}
	if cfg.Concurrency == 0 {
		cfg.Concurrency = DefaultConcurrency
	}
}
