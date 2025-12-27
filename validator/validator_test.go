package validator

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type user struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
	Username string `validate:"required,min=3,max=20"`
}

func TestValidateSuccess(t *testing.T) {
	u := user{Email: "demo@example.com", Password: "superpass", Username: "demo"}
	if err := Validate(u); err != nil {
		t.Fatalf("expected validation success, got %v", err)
	}
}

func TestValidateFailures(t *testing.T) {
	tests := []struct {
		name    string
		payload user
	}{
		{name: "missing email", payload: user{Password: "superpass", Username: "demo"}},
		{name: "bad email", payload: user{Email: "invalid", Password: "superpass", Username: "demo"}},
		{name: "short password", payload: user{Email: "demo@example.com", Password: "short", Username: "demo"}},
		{name: "long username", payload: user{Email: "demo@example.com", Password: "superpass", Username: "this-username-is-way-too-long"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Validate(tt.payload); err == nil {
				t.Fatalf("expected validation error")
			}
		})
	}
}

func TestBindJSONAndValidate(t *testing.T) {
	body := bytes.NewBufferString(`{"email":"demo@example.com","password":"superpass","username":"demo"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)

	var u user
	if err := BindJSON(req, &u); err != nil {
		t.Fatalf("bind+validate failed: %v", err)
	}
}

func TestBindJSONInvalid(t *testing.T) {
	body := bytes.NewBufferString(`{"email":"demo@example.com","password":"short","username":"demo"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)

	var u user
	err := BindJSON(req, &u)
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !strings.Contains(err.Error(), "Password") {
		t.Fatalf("unexpected error: %v", err)
	}
}
