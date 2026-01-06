package validator

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func TestMinDetailed(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		min      int64
		expected bool
	}{
		// Integer types
		{"int below", 5, 10, false},
		{"int equal", 10, 10, true},
		{"int above", 15, 10, true},
		{"int8 below", int8(5), 10, false},
		{"int8 equal", int8(10), 10, true},
		{"int16 below", int16(5), 10, false},
		{"int32 below", int32(5), 10, false},
		{"int64 below", int64(5), 10, false},

		// Unsigned types
		{"uint below", uint(5), 10, false},
		{"uint equal", uint(10), 10, true},
		{"uint above", uint(15), 10, true},
		{"uint8 below", uint8(5), 10, false},
		{"uint16 below", uint16(5), 10, false},
		{"uint32 below", uint32(5), 10, false},
		{"uint64 below", uint64(5), 10, false},

		// Float types
		{"float32 below", float32(5.5), 10, false},
		{"float32 equal", float32(10.0), 10, true},
		{"float64 below", float64(5.5), 10, false},
		{"float64 equal", float64(10.0), 10, true},

		// String types
		{"string numeric below", "5", 10, false},
		{"string numeric equal", "10", 10, true},
		{"string numeric above", "15", 10, true},
		{"string float below", "5.5", 10, false},
		{"string float equal", "10.0", 10, true},
		{"string length below", "short", 10, false},
		{"string length equal", "abcdefghij", 10, true},
		{"string length above", "very long string", 10, true},
		{"string with spaces", "  5  ", 10, false},

		// Edge cases
		{"nil value", nil, 10, true},
		{"zero min", 5, 0, true},
		{"negative min", -5, -10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := Min(tt.min)
			err := rule.Validate(tt.value)
			result := err == nil
			if result != tt.expected {
				t.Errorf("Min(%d).Validate(%v) = %v, expected %v", tt.min, tt.value, result, tt.expected)
				if err != nil {
					t.Logf("Error: %v", err)
				}
			}
		})
	}
}

func TestMaxDetailed(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		max      int64
		expected bool
	}{
		// Integer types
		{"int below", 5, 10, true},
		{"int equal", 10, 10, true},
		{"int above", 15, 10, false},
		{"int8 above", int8(15), 10, false},
		{"int16 above", int16(15), 10, false},
		{"int32 above", int32(15), 10, false},
		{"int64 above", int64(15), 10, false},

		// Unsigned types
		{"uint below", uint(5), 10, true},
		{"uint equal", uint(10), 10, true},
		{"uint above", uint(15), 10, false},
		{"uint8 above", uint8(15), 10, false},
		{"uint16 above", uint16(15), 10, false},
		{"uint32 above", uint32(15), 10, false},
		{"uint64 above", uint64(15), 10, false},

		// Float types
		{"float32 below", float32(5.5), 10, true},
		{"float32 equal", float32(10.0), 10, true},
		{"float64 above", float64(15.5), 10, false},

		// String types
		{"string numeric below", "5", 10, true},
		{"string numeric equal", "10", 10, true},
		{"string numeric above", "15", 10, false},
		{"string float below", "5.5", 10, true},
		{"string float above", "15.5", 10, false},
		{"string length below", "short", 10, true},
		{"string length equal", "abcdefghij", 10, true},
		{"string length above", "very long string", 10, false},
		{"string length equal exact", "abcdefghij", 10, true},
		{"string length above exact", "abcdefghijk", 10, false},

		// Edge cases
		{"nil value", nil, 10, true},
		{"negative max", -5, -10, false},
		{"zero max", 5, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := Max(tt.max)
			err := rule.Validate(tt.value)
			result := err == nil
			if result != tt.expected {
				t.Errorf("Max(%d).Validate(%v) = %v, expected %v", tt.max, tt.value, result, tt.expected)
				if err != nil {
					t.Logf("Error: %v", err)
				}
			}
		})
	}
}

func TestAlphaDetailed(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"valid lowercase", "abc", true},
		{"valid uppercase", "ABC", true},
		{"valid mixed", "AbCd", true},
		{"empty string", "", true},
		{"nil value", nil, true},
		{"with space", "abc def", false},
		{"with number", "abc123", false},
		{"with special", "abc!", false},
		{"with underscore", "abc_def", false},
		{"with dash", "abc-def", false},
		{"with dot", "abc.def", false},
		{"with unicode", "café", false},
		{"single char", "a", true},
		{"long string", "abcdefghijklmnopqrstuvwxy", true},
		{"non-string", 123, false},
		{"string pointer", strPtr("abc"), false}, // Should be string type
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := Alpha()
			err := rule.Validate(tt.value)
			result := err == nil
			if result != tt.expected {
				t.Errorf("Alpha().Validate(%v) = %v, expected %v", tt.value, result, tt.expected)
				if err != nil {
					t.Logf("Error: %v", err)
				}
			}
		})
	}
}

func TestRegexDetailed(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		value    any
		expected bool
	}{
		{"empty pattern", "", "anything", true},
		{"simple pattern", "^[a-z]+$", "abc", true},
		{"simple pattern fail", "^[a-z]+$", "ABC", false},
		{"email pattern", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, "test@example.com", true},
		{"email pattern fail", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, "invalid", false},
		{"numeric pattern", `^\d+$`, "123", true},
		{"numeric pattern fail", `^\d+$`, "abc", false},
		{"nil value", "^[a-z]+$", nil, true},
		{"empty string", "^[a-z]+$", "", true},
		{"non-string", "^[a-z]+$", 123, false},
		{"complex pattern", `^[A-Z][a-z]*$`, "Hello", true},
		{"complex pattern fail", `^[A-Z][a-z]*$`, "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := Regex(tt.pattern)
			err := rule.Validate(tt.value)
			result := err == nil
			if result != tt.expected {
				t.Errorf("Regex(%q).Validate(%v) = %v, expected %v", tt.pattern, tt.value, result, tt.expected)
				if err != nil {
					t.Logf("Error: %v", err)
				}
			}
		})
	}
}

func TestRuleRegistryConcurrency(t *testing.T) {
	registry := NewRuleRegistry()

	// Test concurrent access
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Register rules concurrently
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			rule := Min(int64(idx))
			registry.Register("min"+string(rune('0'+idx)), rule)
		}(i)
	}

	// Wait for all registrations to complete
	wg.Wait()

	// Check rules concurrently
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, exists := registry.Get("min" + string(rune('0'+idx)))
			if !exists {
				errors <- fmt.Errorf("Expected rule %d to exist", idx)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}
}

func TestValidationErrorJSON(t *testing.T) {
	err := ValidationError{
		Field:   "Email",
		Code:    "invalid",
		Message: "Invalid email format",
		Value:   "test@",
	}

	// Test Error() method
	expected := "Email: Invalid email format (invalid)"
	if err.Error() != expected {
		t.Errorf("Error() = %q, expected %q", err.Error(), expected)
	}

	// Test FieldErrors
	fe := FieldErrors{
		errors: []ValidationError{err},
	}

	feError := fe.Error()
	if feError != expected {
		t.Errorf("FieldErrors.Error() = %q, expected %q", feError, expected)
	}

	errors := fe.Errors()
	if len(errors) != 1 {
		t.Errorf("FieldErrors.Errors() returned %d errors, expected 1", len(errors))
	}
}

func TestValidatorParseRule(t *testing.T) {
	v := NewValidator(nil)

	tests := []struct {
		spec     string
		hasRule  bool
		ruleName string
	}{
		{"min=10", true, "min"},
		{"max=100", true, "max"},
		{"minLength=5", true, "minLength"},
		{"maxLength=20", true, "maxLength"},
		{"regex=^[a-z]+$", true, "regex"},
		{"email", true, "email"},
		{"required", true, "required"},
		{"", false, ""},
		{"invalid", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			rule := v.parseRule(tt.spec)
			if tt.hasRule && rule == nil {
				t.Errorf("parseRule(%q) returned nil, expected a rule", tt.spec)
			}
			if !tt.hasRule && rule != nil {
				t.Errorf("parseRule(%q) returned a rule, expected nil", tt.spec)
			}
		})
	}
}

func TestValidatorParseValidationRules(t *testing.T) {
	type TestStruct struct {
		Field1 string `validate:"required,minLength=5"`
		Field2 int    `validate:"min=10,max=100"`
		Field3 string `validate:"email"`
		Field4 string `validate:""` // Empty tag
		Field5 string // No tag
	}

	v := NewValidator(nil)
	typ := reflect.TypeOf(TestStruct{})
	rules := v.parseValidationRules(typ)

	if len(rules) != 5 {
		t.Errorf("Expected 5 rule slices, got %d", len(rules))
	}

	// Field1 should have 2 rules
	if len(rules[0]) != 2 {
		t.Errorf("Field1 should have 2 rules, got %d", len(rules[0]))
	}

	// Field2 should have 2 rules
	if len(rules[1]) != 2 {
		t.Errorf("Field2 should have 2 rules, got %d", len(rules[1]))
	}

	// Field3 should have 1 rule
	if len(rules[2]) != 1 {
		t.Errorf("Field3 should have 1 rule, got %d", len(rules[2]))
	}

	// Field4 and Field5 should have 0 rules
	if len(rules[3]) != 0 {
		t.Errorf("Field4 should have 0 rules, got %d", len(rules[3]))
	}
	if len(rules[4]) != 0 {
		t.Errorf("Field5 should have 0 rules, got %d", len(rules[4]))
	}
}

func TestValidatorGetCacheKey(t *testing.T) {
	v := NewValidator(nil)

	type TestStruct1 struct {
		Field string
	}
	type TestStruct2 struct {
		Field string
	}

	typ1 := reflect.TypeOf(TestStruct1{})
	typ2 := reflect.TypeOf(TestStruct2{})
	typ3 := reflect.TypeOf(TestStruct1{}) // Same as typ1

	key1 := v.getCacheKey(typ1)
	key2 := v.getCacheKey(typ2)
	key3 := v.getCacheKey(typ3)

	if key1 == key2 {
		t.Errorf("Different types should have different cache keys")
	}

	if key1 != key3 {
		t.Errorf("Same type should have same cache key")
	}
}

// Helper function
func strPtr(s string) *string {
	return &s
}
