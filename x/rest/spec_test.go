package rest

import (
	"context"
	"net/http/httptest"
	"testing"
)

type stubTransformer struct{}

func (stubTransformer) Transform(_ context.Context, resource any) (any, error) {
	return resource, nil
}

func (stubTransformer) TransformCollection(_ context.Context, resources any) (any, error) {
	return resources, nil
}

func TestApplyResourceSpecOrchestratesControllerDefaults(t *testing.T) {
	controller := NewBaseContextResourceController("users")
	spec := ResourceSpec{
		Name:   "accounts",
		Prefix: "/api/accounts",
		Options: &ResourceOptions{
			DefaultPageSize:   25,
			MaxPageSize:       50,
			AllowedSorts:      []string{"created_at"},
			AllowedFilters:    []string{"tenant_id"},
			EnableHooks:       false,
			EnableTransformer: false,
		},
		Transformer: stubTransformer{},
	}

	ApplyResourceSpec(controller, spec)

	if controller.ResourceName != "accounts" {
		t.Fatalf("expected resource name to come from spec, got %q", controller.ResourceName)
	}
	if controller.Spec.Name != "accounts" {
		t.Fatalf("expected spec to be stored on controller")
	}
	if controller.QueryBuilder == nil {
		t.Fatalf("expected query builder to be initialized")
	}
	req := httptest.NewRequest("GET", "/api/accounts?page_size=200&sort=ignored&tenant_id=t1&skip=x", nil)
	params := controller.QueryBuilder.Parse(req)
	if params.PageSize != 50 {
		t.Fatalf("expected query builder to use spec max page size, got %d", params.PageSize)
	}
	if len(params.Sort) != 0 {
		t.Fatalf("expected query builder to reject non-allowed sorts, got %#v", params.Sort)
	}
	if len(params.Filters) != 1 || params.Filters["tenant_id"] != "t1" {
		t.Fatalf("expected query builder to keep only allowed filters, got %#v", params.Filters)
	}
	if _, ok := controller.Hooks.(*NoOpResourceHooks); !ok {
		t.Fatalf("expected hooks to fall back to NoOpResourceHooks when disabled")
	}
	if _, ok := controller.Transformer.(*IdentityTransformer); !ok {
		t.Fatalf("expected transformer to fall back to IdentityTransformer when disabled")
	}
}

func TestParseQueryParamsUsesSpecDrivenNormalization(t *testing.T) {
	controller := NewBaseContextResourceController("users").ApplySpec(ResourceSpec{
		Name:   "users",
		Prefix: "/api/users",
		Options: &ResourceOptions{
			DefaultPageSize: 20,
			MaxPageSize:     50,
			AllowedSorts:    []string{"created_at"},
			AllowedFilters:  []string{"tenant_id"},
			EnableHooks:     true,
		},
	})

	req := httptest.NewRequest("GET", "/api/users?page=0&page_size=200&sort=name,-created_at&tenant_id=t1&ignored=x", nil)
	params := controller.ParseQueryParams(req)

	if params.Page != 1 {
		t.Fatalf("expected page to default to 1, got %d", params.Page)
	}
	if params.PageSize != 50 {
		t.Fatalf("expected page size to be capped to 50, got %d", params.PageSize)
	}
	if params.Limit != 50 {
		t.Fatalf("expected limit to track normalized page size, got %d", params.Limit)
	}
	if params.Offset != 0 {
		t.Fatalf("expected offset to be normalized from page and page size, got %d", params.Offset)
	}
	if len(params.Sort) != 1 || params.Sort[0].Field != "created_at" || !params.Sort[0].Desc {
		t.Fatalf("expected sort allowlist to keep only created_at desc, got %#v", params.Sort)
	}
	if len(params.Filters) != 1 || params.Filters["tenant_id"] != "t1" {
		t.Fatalf("expected filter allowlist to keep only tenant_id, got %#v", params.Filters)
	}
}
