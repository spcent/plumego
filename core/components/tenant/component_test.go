package tenantcomponent

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/spcent/plumego/health"
	"github.com/spcent/plumego/tenant"
)

// stubConfigManager implements tenant.ConfigManager for testing.
type stubConfigManager struct{}

func (s *stubConfigManager) GetTenantConfig(_ context.Context, tenantID string) (tenant.Config, error) {
	if tenantID == "" {
		return tenant.Config{}, fmt.Errorf("empty tenant ID")
	}
	return tenant.Config{TenantID: tenantID}, nil
}

func TestHealthWithManager(t *testing.T) {
	c := &TenantConfigComponent{
		Name:    "my-tenant",
		Manager: &stubConfigManager{},
	}
	name, status := c.Health()
	if name != "my-tenant" {
		t.Fatalf("expected name %q, got %q", "my-tenant", name)
	}
	if status.Status != health.StatusHealthy {
		t.Fatalf("expected healthy, got %s", status.Status)
	}
}

func TestHealthWithoutManager(t *testing.T) {
	c := &TenantConfigComponent{Name: "test"}
	name, status := c.Health()
	if name != "test" {
		t.Fatalf("expected name %q, got %q", "test", name)
	}
	if status.Status != health.StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %s", status.Status)
	}
	if status.Message == "" {
		t.Fatal("expected non-empty message for nil manager")
	}
}

func TestHealthDefaultName(t *testing.T) {
	c := &TenantConfigComponent{Manager: &stubConfigManager{}}
	name, _ := c.Health()
	if name != "tenant-config" {
		t.Fatalf("expected default name %q, got %q", "tenant-config", name)
	}
}

func TestHealthEmptyNameNoManager(t *testing.T) {
	c := &TenantConfigComponent{}
	name, status := c.Health()
	if name != "tenant-config" {
		t.Fatalf("expected default name, got %q", name)
	}
	if status.Status != health.StatusUnhealthy {
		t.Fatalf("expected unhealthy, got %s", status.Status)
	}
}

func TestStartReturnsNil(t *testing.T) {
	c := &TenantConfigComponent{Manager: &stubConfigManager{}}
	if err := c.Start(context.Background()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestStopReturnsNil(t *testing.T) {
	c := &TenantConfigComponent{}
	if err := c.Stop(context.Background()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestRegisterRoutesNoop(t *testing.T) {
	c := &TenantConfigComponent{}
	// Should not panic with nil router
	c.RegisterRoutes(nil)
}

func TestRegisterMiddlewareNoop(t *testing.T) {
	c := &TenantConfigComponent{}
	c.RegisterMiddleware(nil)
}

func TestDependenciesReturnsNil(t *testing.T) {
	c := &TenantConfigComponent{}
	deps := c.Dependencies()
	if deps != nil {
		t.Fatalf("expected nil dependencies, got %v", deps)
	}
}

func TestComponentInterface(t *testing.T) {
	// Verify TenantConfigComponent satisfies the expected interface shape
	c := &TenantConfigComponent{}
	_ = c.Dependencies()
	_, _ = c.Health()
	_ = c.Start(context.Background())
	_ = c.Stop(context.Background())

	// Verify type implements expected methods via reflection
	typ := reflect.TypeOf(c)
	methods := []string{"RegisterRoutes", "RegisterMiddleware", "Start", "Stop", "Health", "Dependencies"}
	for _, m := range methods {
		if _, ok := typ.MethodByName(m); !ok {
			t.Fatalf("TenantConfigComponent missing method %s", m)
		}
	}
}
