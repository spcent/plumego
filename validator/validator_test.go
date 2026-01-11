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
	Username string `validate:"required,minLength=3,maxLength=20"`
}

type advancedUser struct {
	Email    string `validate:"required,email"`
	Phone    string `validate:"phone"`
	Website  string `validate:"url"`
	Age      int    `validate:"min=18,max=100"`
	Username string `validate:"required,alphaNum"`
	Code     string `validate:"regex=^[A-Z]{3}$"`
	Score    string `validate:"numeric"`
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
		{name: "short username", payload: user{Email: "demo@example.com", Password: "superpass", Username: "ab"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Validate(tt.payload); err == nil {
				t.Fatalf("expected validation error")
			}
		})
	}
}

func TestAdvancedValidationRules(t *testing.T) {
	tests := []struct {
		name    string
		payload advancedUser
		wantErr bool
	}{
		{
			name: "valid user",
			payload: advancedUser{
				Email:    "user@example.com",
				Phone:    "+1234567890",
				Website:  "https://example.com",
				Age:      25,
				Username: "johndoe123",
				Code:     "ABC",
				Score:    "95.5",
			},
			wantErr: false,
		},
		{
			name: "invalid phone",
			payload: advancedUser{
				Email:    "user@example.com",
				Phone:    "invalid-phone",
				Website:  "https://example.com",
				Age:      25,
				Username: "johndoe123",
				Code:     "ABC",
				Score:    "95.5",
			},
			wantErr: true,
		},
		{
			name: "invalid url",
			payload: advancedUser{
				Email:    "user@example.com",
				Phone:    "+1234567890",
				Website:  "not-a-url",
				Age:      25,
				Username: "johndoe123",
				Code:     "ABC",
				Score:    "95.5",
			},
			wantErr: true,
		},
		{
			name: "age out of range",
			payload: advancedUser{
				Email:    "user@example.com",
				Phone:    "+1234567890",
				Website:  "https://example.com",
				Age:      15,
				Username: "johndoe123",
				Code:     "ABC",
				Score:    "95.5",
			},
			wantErr: true,
		},
		{
			name: "invalid username",
			payload: advancedUser{
				Email:    "user@example.com",
				Phone:    "+1234567890",
				Website:  "https://example.com",
				Age:      25,
				Username: "john-doe", // contains dash, not alphanumeric
				Code:     "ABC",
				Score:    "95.5",
			},
			wantErr: true,
		},
		{
			name: "invalid code",
			payload: advancedUser{
				Email:    "user@example.com",
				Phone:    "+1234567890",
				Website:  "https://example.com",
				Age:      25,
				Username: "johndoe123",
				Code:     "ABCD", // too long
				Score:    "95.5",
			},
			wantErr: true,
		},
		{
			name: "non-numeric score",
			payload: advancedUser{
				Email:    "user@example.com",
				Phone:    "+1234567890",
				Website:  "https://example.com",
				Age:      25,
				Username: "johndoe123",
				Code:     "ABC",
				Score:    "abc",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.payload)
			if tt.wantErr && err == nil {
				t.Fatalf("expected validation error, got none")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no validation error, got: %v", err)
			}
		})
	}
}

func TestValidationErrorStructure(t *testing.T) {
	u := user{Email: "invalid-email", Password: "short", Username: "a"}
	err := Validate(u)

	if err == nil {
		t.Fatalf("expected validation error")
	}

	// Check if it's a FieldErrors type
	fieldErrors, ok := err.(FieldErrors)
	if !ok {
		t.Fatalf("expected FieldErrors type, got %T", err)
	}

	errors := fieldErrors.Errors()
	if len(errors) == 0 {
		t.Fatalf("expected field errors")
	}

	// Check error structure
	for _, e := range errors {
		if e.Field == "" {
			t.Errorf("expected non-empty field name")
		}
		if e.Code == "" {
			t.Errorf("expected non-empty error code")
		}
		if e.Message == "" {
			t.Errorf("expected non-empty error message")
		}
	}
}

func TestCustomRuleRegistration(t *testing.T) {
	// Create custom rule
	customRule := RuleFunc(func(value any) *ValidationError {
		if str, ok := value.(string); ok {
			if str == "forbidden" {
				return &ValidationError{Code: "forbidden", Message: "this value is forbidden"}
			}
		}
		return nil
	})

	// Create registry and register custom rule
	registry := NewRuleRegistry()
	registry.Register("forbidden", customRule)

	validator := NewValidator(registry)

	type testStruct struct {
		Name string `validate:"forbidden"`
	}

	// Test valid value
	ts := testStruct{Name: "allowed"}
	if err := validator.Validate(ts); err != nil {
		t.Fatalf("expected validation success, got: %v", err)
	}

	// Test forbidden value
	ts.Name = "forbidden"
	if err := validator.Validate(ts); err == nil {
		t.Fatalf("expected validation error for forbidden value")
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

func TestPointerHandling(t *testing.T) {
	type pointerStruct struct {
		Email    *string `validate:"required,email"`
		Password *string `validate:"required,min=8"`
	}

	// Test with nil pointers
	ps := pointerStruct{}
	err := Validate(ps)
	if err == nil {
		t.Fatalf("expected validation error for nil pointers")
	}

	// Test with valid values
	email := "test@example.com"
	password := "validpass"
	ps = pointerStruct{Email: &email, Password: &password}
	if err := Validate(ps); err != nil {
		t.Fatalf("expected validation success, got: %v", err)
	}

	// Test with invalid email
	ps.Email = &email
	ps.Password = &password
	invalidEmail := "invalid-email"
	ps.Email = &invalidEmail
	err = Validate(ps)
	if err == nil {
		t.Fatalf("expected validation error for invalid email")
	}
}

func TestEmptyAndWhitespaceHandling(t *testing.T) {
	type whitespaceStruct struct {
		Email    string `validate:"required,email"`
		Username string `validate:"required,minLength=3"`
	}

	// Test with whitespace-only values
	ws := whitespaceStruct{
		Email:    "   ",
		Username: "   ",
	}
	err := Validate(ws)
	if err == nil {
		t.Fatalf("expected validation error for whitespace values")
	}
}

func TestNumericStringValidation(t *testing.T) {
	type numericStruct struct {
		Age    int    `validate:"min=18,max=100"`
		Score  string `validate:"numeric"`
		Height int64  `validate:"min=100,max=300"`
	}

	ns := numericStruct{
		Age:    25,
		Score:  "95.5",
		Height: 175,
	}

	if err := Validate(ns); err != nil {
		t.Fatalf("expected validation success, got: %v", err)
	}

	// Test invalid numeric string
	ns.Score = "not-a-number"
	err := Validate(ns)
	if err == nil {
		t.Fatalf("expected validation error for non-numeric string")
	}
}
