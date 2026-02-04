package tenantcomponent

import (
	"context"
	"reflect"

	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/router"
	"github.com/spcent/plumego/tenant"
)

// TenantConfigComponent exposes a tenant config manager as a core component.
type TenantConfigComponent struct {
	Name    string
	Manager tenant.ConfigManager
}

// RegisterRoutes implements Component.
func (c *TenantConfigComponent) RegisterRoutes(_ *router.Router) {}

// RegisterMiddleware implements Component.
func (c *TenantConfigComponent) RegisterMiddleware(_ *middleware.Registry) {}

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

// Dependencies implements Component.
func (c *TenantConfigComponent) Dependencies() []reflect.Type { return nil }
