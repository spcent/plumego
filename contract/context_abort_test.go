package contract

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAbortAndIsAborted(t *testing.T) {
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), nil)
	defer ctx.Close()

	if ctx.IsAborted() {
		t.Fatal("new context should not be aborted")
	}

	ctx.Abort()

	if !ctx.IsAborted() {
		t.Fatal("context should be aborted after Abort()")
	}
}

func TestAbortIdempotent(t *testing.T) {
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), nil)
	defer ctx.Close()

	// Calling Abort multiple times must not panic.
	ctx.Abort()
	ctx.Abort()

	if !ctx.IsAborted() {
		t.Fatal("context should remain aborted")
	}
}

func TestAbortWithStatus(t *testing.T) {
	rr := httptest.NewRecorder()
	ctx := NewCtx(rr, httptest.NewRequest(http.MethodGet, "/", nil), nil)
	defer ctx.Close()

	ctx.AbortWithStatus(http.StatusForbidden)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
	if !ctx.IsAborted() {
		t.Fatal("context should be aborted after AbortWithStatus")
	}
}

func TestAbortCancelsContext(t *testing.T) {
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), nil)
	// Don't defer Close here because Abort already cancels.

	reqCtx := ctx.R.Context()
	ctx.Abort()

	select {
	case <-reqCtx.Done():
		// expected
	default:
		t.Fatal("Abort should cancel the request context")
	}
}

func TestErrorAppend(t *testing.T) {
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), nil)
	defer ctx.Close()

	if len(ctx.Errors) != 0 {
		t.Fatal("new context should have no errors")
	}

	err1 := errors.New("first")
	err2 := errors.New("second")

	ret := ctx.Error(err1)
	if ret != err1 {
		t.Fatal("Error() should return the same error")
	}
	ctx.Error(err2)

	if len(ctx.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(ctx.Errors))
	}
	if ctx.Errors[0] != err1 || ctx.Errors[1] != err2 {
		t.Fatal("errors should be in append order")
	}
}

func TestErrorNilIgnored(t *testing.T) {
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), nil)
	defer ctx.Close()

	ctx.Error(nil)

	if len(ctx.Errors) != 0 {
		t.Fatal("nil errors should not be appended")
	}
}

func TestAbortStopsMiddlewareChain(t *testing.T) {
	// Simulate an abort-aware middleware pattern.
	var order []string

	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), nil)
	defer ctx.Close()

	// Simulate middleware 1: aborts
	order = append(order, "mw1-before")
	ctx.Abort()
	order = append(order, "mw1-after-abort")

	// Simulate middleware 2: checks abort
	if !ctx.IsAborted() {
		order = append(order, "mw2")
	}

	// Simulate handler: checks abort
	if !ctx.IsAborted() {
		order = append(order, "handler")
	}

	expected := []string{"mw1-before", "mw1-after-abort"}
	if len(order) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, order)
	}
	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("expected %v, got %v", expected, order)
		}
	}
}
