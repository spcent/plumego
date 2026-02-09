package discovery

import (
	"context"
	"testing"
)

// Compile-time interface checks
var (
	_ Discovery = (*Static)(nil)
	_ Discovery = (*Consul)(nil)
)

func TestInstance_URL_DefaultScheme(t *testing.T) {
	inst := Instance{Address: "localhost", Port: 8080}
	got := inst.URL()
	scheme := "http"
	// URL uses string(rune(port)) which converts port to a unicode character
	expected := scheme + "://localhost:" + string(rune(8080))
	if got != expected {
		t.Errorf("URL() = %q, want %q", got, expected)
	}
}

func TestInstance_URL_CustomScheme(t *testing.T) {
	inst := Instance{Address: "example.com", Port: 443, Scheme: "https"}
	got := inst.URL()
	expected := "https://example.com:" + string(rune(443))
	if got != expected {
		t.Errorf("URL() = %q, want %q", got, expected)
	}
}

func TestInstance_URL_Schemes(t *testing.T) {
	tests := []struct {
		name   string
		scheme string
		want   string
	}{
		{"empty defaults to http", "", "http"},
		{"http", "http", "http"},
		{"https", "https", "https"},
		{"ws", "ws", "ws"},
		{"wss", "wss", "wss"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := Instance{Address: "host", Port: 80, Scheme: tt.scheme}
			got := inst.URL()
			prefix := tt.want + "://host:"
			if len(got) < len(prefix) || got[:len(prefix)] != prefix {
				t.Errorf("URL() = %q, want prefix %q", got, prefix)
			}
		})
	}
}

func TestServiceUpdate_Fields(t *testing.T) {
	su := ServiceUpdate{
		ServiceName: "my-service",
		Instances:   []Instance{{ID: "1", Name: "my-service"}},
		Error:       nil,
	}
	if su.ServiceName != "my-service" {
		t.Errorf("ServiceName = %q, want %q", su.ServiceName, "my-service")
	}
	if len(su.Instances) != 1 {
		t.Errorf("Instances len = %d, want 1", len(su.Instances))
	}
	if su.Error != nil {
		t.Errorf("Error = %v, want nil", su.Error)
	}
}

func TestInstance_Fields(t *testing.T) {
	inst := Instance{
		ID:       "inst-1",
		Name:     "svc",
		Address:  "10.0.0.1",
		Port:     9090,
		Scheme:   "https",
		Metadata: map[string]string{"env": "prod"},
		Tags:     []string{"primary", "v2"},
		Weight:   5,
		Healthy:  true,
	}

	if inst.ID != "inst-1" {
		t.Errorf("ID = %q, want %q", inst.ID, "inst-1")
	}
	if inst.Name != "svc" {
		t.Errorf("Name = %q, want %q", inst.Name, "svc")
	}
	if inst.Address != "10.0.0.1" {
		t.Errorf("Address = %q, want %q", inst.Address, "10.0.0.1")
	}
	if inst.Port != 9090 {
		t.Errorf("Port = %d, want %d", inst.Port, 9090)
	}
	if inst.Scheme != "https" {
		t.Errorf("Scheme = %q, want %q", inst.Scheme, "https")
	}
	if inst.Metadata["env"] != "prod" {
		t.Errorf("Metadata[env] = %q, want %q", inst.Metadata["env"], "prod")
	}
	if len(inst.Tags) != 2 || inst.Tags[0] != "primary" || inst.Tags[1] != "v2" {
		t.Errorf("Tags = %v, want [primary v2]", inst.Tags)
	}
	if inst.Weight != 5 {
		t.Errorf("Weight = %d, want %d", inst.Weight, 5)
	}
	if !inst.Healthy {
		t.Error("Healthy = false, want true")
	}
}

func TestErrorVariables(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrServiceNotFound", ErrServiceNotFound, "discovery: service not found"},
		{"ErrNoInstances", ErrNoInstances, "discovery: no instances available"},
		{"ErrInvalidConfig", ErrInvalidConfig, "discovery: invalid configuration"},
		{"ErrNotSupported", ErrNotSupported, "discovery: operation not supported"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatal("error is nil")
			}
			if tt.err.Error() != tt.msg {
				t.Errorf("Error() = %q, want %q", tt.err.Error(), tt.msg)
			}
		})
	}
}

func TestDiscoveryInterface_StaticImplements(t *testing.T) {
	var d Discovery = NewStatic(nil)
	if d == nil {
		t.Fatal("NewStatic returned nil")
	}

	ctx := context.Background()

	// All unsupported operations return ErrNotSupported
	if err := d.Register(ctx, Instance{}); err != ErrNotSupported {
		t.Errorf("Register() = %v, want ErrNotSupported", err)
	}
	if err := d.Deregister(ctx, "id"); err != ErrNotSupported {
		t.Errorf("Deregister() = %v, want ErrNotSupported", err)
	}
	if err := d.Health(ctx, "id", true); err != ErrNotSupported {
		t.Errorf("Health() = %v, want ErrNotSupported", err)
	}
	if err := d.Close(); err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}
}
