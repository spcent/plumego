package rest

import "testing"

func TestApplyResourceSpecPreservesControllerDefaults(t *testing.T) {
	controller := &BaseContextResourceController{
		ResourceName: "users",
	}

	ApplyResourceSpec(controller, ResourceSpec{})

	if controller.ResourceName != "users" {
		t.Fatalf("expected resource name to preserve controller default, got %q", controller.ResourceName)
	}
	if controller.Spec.Name != "users" {
		t.Fatalf("expected spec name to preserve controller default, got %q", controller.Spec.Name)
	}
	if controller.Spec.Prefix != "/users" {
		t.Fatalf("expected normalized prefix to follow preserved resource name, got %q", controller.Spec.Prefix)
	}
	if controller.ParamExtractor == nil {
		t.Fatalf("expected ApplyResourceSpec to initialize param extractor")
	}
	if controller.QueryBuilder == nil {
		t.Fatalf("expected ApplyResourceSpec to initialize query builder")
	}
	if controller.Hooks == nil {
		t.Fatalf("expected ApplyResourceSpec to initialize hooks")
	}
	if controller.Transformer == nil {
		t.Fatalf("expected ApplyResourceSpec to initialize transformer")
	}
}
