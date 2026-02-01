package core

import (
	"context"

	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/tenant"
)

// TenantConfigComponent exposes a tenant config manager as a core component.
type TenantConfigComponent struct {
	BaseComponent
	Name    string
	Manager tenant.ConfigManager
}

// Start implements Component.
func (c *TenantConfigComponent) Start(ctx context.Context) error {
	return nil
}

// Stop implements Component.
func (c *TenantConfigComponent) Stop(ctx context.Context) error {
	return nil
}

// Health reports the status of the config manager.
func (c *TenantConfigComponent) Health() (name string, status health.HealthStatus) {
	componentName := c.Name
	if componentName == "" {
		componentName = "tenant-config"
	}
	if c.Manager == nil {
		return componentName, health.HealthStatus{Status: health.StatusUnhealthy, Message: "tenant config manager is nil"}
	}
	return componentName, health.HealthStatus{Status: health.StatusHealthy}
}
