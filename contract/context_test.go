package contract

import (
	"reflect"
	"testing"
)

func TestRequestContextFromContext(t *testing.T) {
	result := RequestContextFromContext(t.Context())
	if len(result.Params) != 0 {
		t.Fatal("expected empty RequestContext")
	}

	rc := RequestContext{
		Params:       map[string]string{"id": "123"},
		RoutePattern: "/users/:id",
		RouteName:    "user_show",
	}
	ctx := WithRequestContext(t.Context(), rc)
	result = RequestContextFromContext(ctx)
	if result.Params == nil || result.Params["id"] != "123" {
		t.Fatal("expected RequestContext with params")
	}
	if result.RoutePattern != "/users/:id" || result.RouteName != "user_show" {
		t.Fatal("expected route fields from RequestContext")
	}
}

func TestWithRequestContextNilContext(t *testing.T) {
	rc := RequestContext{
		Params:       map[string]string{"id": "123"},
		RoutePattern: "/users/:id",
		RouteName:    "user_show",
	}

	ctx := WithRequestContext(nil, rc)
	result := RequestContextFromContext(ctx)
	if result.Params == nil || result.Params["id"] != "123" {
		t.Fatal("expected RequestContext to survive nil parent context")
	}
	if result.RoutePattern != "/users/:id" || result.RouteName != "user_show" {
		t.Fatal("expected route fields from nil-parent RequestContext")
	}
}

func TestRequestContextParamsAreCopied(t *testing.T) {
	params := map[string]string{"id": "42"}
	ctx := WithRequestContext(t.Context(), RequestContext{
		Params:       params,
		RoutePattern: "/users/:id",
		RouteName:    "users.show",
	})

	params["id"] = "mutated"

	rc := RequestContextFromContext(ctx)
	if rc.Params["id"] != "42" {
		t.Fatalf("expected stored params to be isolated, got %q", rc.Params["id"])
	}

	rc.Params["id"] = "returned-mutated"
	again := RequestContextFromContext(ctx)
	if again.Params["id"] != "42" {
		t.Fatalf("expected returned params mutation to be isolated, got %q", again.Params["id"])
	}
	if again.RoutePattern != "/users/:id" || again.RouteName != "users.show" {
		t.Fatalf("expected route metadata to be preserved, got %+v", again)
	}
}

func TestRequestContextStableCarrierFields(t *testing.T) {
	rt := reflect.TypeOf(RequestContext{})
	want := []string{"Params", "RoutePattern", "RouteName"}
	if rt.NumField() != len(want) {
		t.Fatalf("RequestContext field count = %d, want %d", rt.NumField(), len(want))
	}
	for i, name := range want {
		if got := rt.Field(i).Name; got != name {
			t.Fatalf("RequestContext field %d = %s, want %s", i, got, name)
		}
	}
}
