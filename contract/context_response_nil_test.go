package contract

import (
	"errors"
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
