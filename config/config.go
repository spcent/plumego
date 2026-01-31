// Package config provides a flexible configuration management system with
// support for multiple sources (environment variables, files), type-safe
// accessors, validation, and hot-reloading.
//
// Basic usage:
//
//	cfg := config.New()
//	cfg.AddSource(config.NewEnvSource("APP_"))
//	cfg.AddSource(config.NewFileSource("config.json", config.FormatJSON, true))
//
//	if err := cfg.Load(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
//	port := cfg.GetInt("server_port", 8080)
//	debug := cfg.GetBool("debug_mode", false)
//
// Global configuration:
//
//	config.InitDefault() // Auto-loads from env and common config files
//	port := config.GetInt("server_port", 8080)
//
// Type-safe access with validation:
//
//	port, err := cfg.Int("port", 8080, &config.Range{Min: 1, Max: 65535})
//	url, err := cfg.String("api_url", "", &config.Required{}, &config.URL{})
package config

import (
	log "github.com/spcent/plumego/log"
)

// ConfigManager is an alias for Manager for backward compatibility.
type ConfigManager = Manager

// New creates a new configuration manager instance.
func New() *Manager {
	return NewManager(log.NewGLogger())
}

// NewConfigManager creates a new Manager instance (backward compatibility alias).
func NewConfigManager(logger log.StructuredLogger) *Manager {
	return NewManager(logger)
}
