package contract

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCtxResponseMethodsNilSafe(t *testing.T) {
	var c *Ctx

	tests := []struct {
		name string
		call func() error
	}{
		{name: "Response", call: func() error { return c.Response(200, nil, nil) }},
		{name: "Text", call: func() error { return c.Text(200, "") }},
		{name: "Bytes", call: func() error { return c.Bytes(200, nil) }},
		{name: "Redirect", call: func() error { return c.Redirect(302, "/") }},
		{name: "UnsafeRedirect", call: func() error { return c.UnsafeRedirect(302, "/") }},
		{name: "File", call: func() error { return c.File("/tmp/x") }},
		{name: "Cookie", call: func() error { _, err := c.Cookie("x"); return err }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call()
			if !errors.Is(err, ErrContextNil) {
				t.Fatalf("expected ErrContextNil, got %v", err)
			}
		})
	}
}

func TestSetCookieNilCtxNoPanic(t *testing.T) {
	var c *Ctx
	c.SetCookie(&http.Cookie{Name: "x", Value: "y"}) // must not panic
}

func TestSetCookieDoesNotMutateCaller(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := NewCtx(rec, httptest.NewRequest("GET", "/", nil), nil)

	original := &http.Cookie{Name: "tok", Value: "abc", Secure: false, Path: ""}
	ctx.SetCookie(original)

	if original.Secure {
		t.Error("SetCookie must not set Secure on caller's cookie")
	}
	if original.Path != "" {
		t.Errorf("SetCookie must not set Path on caller's cookie, got %q", original.Path)
	}
}
