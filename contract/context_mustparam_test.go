package contract

import (
	"errors"
	"testing"
)

func TestMustParamWrapsErrMissingParam(t *testing.T) {
	ctx := &Ctx{Params: map[string]string{}}

	_, err := ctx.MustParam("id")
	if err == nil {
		t.Fatal("expected error for missing param")
	}
	if !errors.Is(err, ErrMissingParam) {
		t.Fatalf("expected errors.Is(err, ErrMissingParam) to be true, got %v", err)
	}
	if got := err.Error(); got != "missing parameter: id" {
		t.Fatalf("unexpected error message %q", got)
	}
}

func TestMustParamNilCtxWrapsErrMissingParam(t *testing.T) {
	var ctx *Ctx

	_, err := ctx.MustParam("id")
	if err == nil {
		t.Fatal("expected missing param error")
	}
	if !errors.Is(err, ErrMissingParam) {
		t.Fatalf("expected errors.Is(err, ErrMissingParam) to be true, got %v", err)
	}
	if got := err.Error(); got != "missing parameter: id" {
		t.Fatalf("unexpected error message %q", got)
	}
}
