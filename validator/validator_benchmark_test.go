package validator

import (
	"testing"
)

// Benchmark structures
type benchmarkUser struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
	Username string `validate:"required,min=3,max=20"`
}

type benchmarkAdvancedUser struct {
	Email    string `validate:"required,email"`
	Phone    string `validate:"phone"`
	Website  string `validate:"url"`
	Age      int    `validate:"min=18,max=100"`
	Username string `validate:"required,alphaNum"`
	Code     string `validate:"regex=^[A-Z]{3}$"`
	Score    string `validate:"numeric"`
}

// BenchmarkValidateSimple benchmarks simple validation
func BenchmarkValidateSimple(b *testing.B) {
	u := benchmarkUser{
		Email:    "user@example.com",
		Password: "superpass123",
		Username: "johndoe",
	}

	validator := NewValidator(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := validator.Validate(u); err != nil {
			b.Fatalf("validation failed: %v", err)
		}
	}
}

// BenchmarkValidateAdvanced benchmarks advanced validation with multiple rules
func BenchmarkValidateAdvanced(b *testing.B) {
	u := benchmarkAdvancedUser{
		Email:    "user@example.com",
		Phone:    "+1234567890",
		Website:  "https://example.com",
		Age:      25,
		Username: "johndoe123",
		Code:     "ABC",
		Score:    "95.5",
	}

	validator := NewValidator(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := validator.Validate(u); err != nil {
			b.Fatalf("validation failed: %v", err)
		}
	}
}

// BenchmarkValidateWithErrors benchmarks validation that fails
func BenchmarkValidateWithErrors(b *testing.B) {
	u := benchmarkUser{
		Email:    "invalid-email",
		Password: "short",
		Username: "ab", // too short
	}

	validator := NewValidator(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := validator.Validate(u); err == nil {
			b.Fatalf("expected validation error")
		}
	}
}

// BenchmarkCacheEffectiveness benchmarks the caching mechanism
func BenchmarkCacheEffectiveness(b *testing.B) {
	validator := NewValidator(nil)

	// First validation will populate cache
	u := benchmarkUser{
		Email:    "user@example.com",
		Password: "superpass123",
		Username: "johndoe",
	}

	validator.Validate(u)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := validator.Validate(u); err != nil {
			b.Fatalf("validation failed: %v", err)
		}
	}
}

// BenchmarkCustomRegistry benchmarks custom registry usage
func BenchmarkCustomRegistry(b *testing.B) {
	registry := NewRuleRegistry()

	// Add some custom rules
	customRule := RuleFunc(func(value any) *ValidationError {
		if str, ok := value.(string); ok {
			if len(str) < 5 {
				return &ValidationError{Code: "custom", Message: "too short"}
			}
		}
		return nil
	})

	registry.Register("customLength", customRule)
	validator := NewValidator(registry)

	type customStruct struct {
		Name string `validate:"customLength"`
	}

	cs := customStruct{Name: "validname"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := validator.Validate(cs); err != nil {
			b.Fatalf("validation failed: %v", err)
		}
	}
}

// BenchmarkReflectionVsNoReflection benchmarks the efficiency of our reflection usage
func BenchmarkMultipleStructTypes(b *testing.B) {
	validator := NewValidator(nil)

	type type1 struct {
		Field1 string `validate:"required"`
		Field2 int    `validate:"min=1"`
	}

	type type2 struct {
		Field3 string `validate:"email"`
		Field4 bool   `validate:"required"`
	}

	// Create instances
	t1 := type1{Field1: "test", Field2: 10}
	t2 := type2{Field3: "test@example.com", Field4: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Alternate between different struct types
		if i%2 == 0 {
			if err := validator.Validate(t1); err != nil {
				b.Fatalf("validation failed: %v", err)
			}
		} else {
			if err := validator.Validate(t2); err != nil {
				b.Fatalf("validation failed: %v", err)
			}
		}
	}
}

// BenchmarkValidationWithPointers benchmarks pointer field validation
func BenchmarkValidationWithPointers(b *testing.B) {
	type pointerStruct struct {
		Email    *string `validate:"required,email"`
		Password *string `validate:"required,min=8"`
		Age      *int    `validate:"min=18,max=100"`
	}

	email := "user@example.com"
	password := "superpass123"
	age := 25

	ps := pointerStruct{
		Email:    &email,
		Password: &password,
		Age:      &age,
	}

	validator := NewValidator(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := validator.Validate(ps); err != nil {
			b.Fatalf("validation failed: %v", err)
		}
	}
}

// BenchmarkLargeStruct benchmarks validation of a large struct
func BenchmarkLargeStruct(b *testing.B) {
	type largeStruct struct {
		Field1  string `validate:"required,email"`
		Field2  string `validate:"required,min=10"`
		Field3  string `validate:"required,max=50"`
		Field4  int    `validate:"min=1,max=100"`
		Field5  bool   `validate:"required"`
		Field6  string `validate:"alphaNum"`
		Field7  string `validate:"url"`
		Field8  string `validate:"phone"`
		Field9  string `validate:"regex=^[A-Z]{3}$"`
		Field10 string `validate:"numeric"`
	}

	ls := largeStruct{
		Field1:  "test@example.com",
		Field2:  "this is a test field",
		Field3:  "medium length",
		Field4:  50,
		Field5:  true,
		Field6:  "test123",
		Field7:  "https://example.com",
		Field8:  "+1234567890",
		Field9:  "ABC",
		Field10: "123.45",
	}

	validator := NewValidator(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := validator.Validate(ls); err != nil {
			b.Fatalf("validation failed: %v", err)
		}
	}
}
