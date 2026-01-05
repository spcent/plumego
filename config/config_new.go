package config

import (
	log "github.com/spcent/plumego/log"
)

// New creates a new ConfigManager instance (backward compatibility alias)
func New() *ConfigManager {
	return NewConfigManager(log.NewGLogger())
}
