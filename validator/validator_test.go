package validator

import (
	"net/http/httptest"
	"strings"
	"testing"
)

// TestInt tests the Int validation rule
func TestInt(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid int", 42, false},
		{"valid int8", int8(42), false},
		{"valid int16", int16(42), false},
		{"valid int32", int32(42), false},
		{"valid int64", int64(42), false},
		{"valid uint", uint(42), false},
		{"valid uint8", uint8(42), false},
		{"valid uint16", uint16(42), false},
		{"valid uint32", uint32(42), false},
		{"valid uint64", uint64(42), false},
		{"valid string int", "42", false},
		{"invalid string", "abc", true},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := Int()
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("Int() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestFloat tests the Float validation rule
func TestFloat(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid float32", float32(3.14), false},
		{"valid float64", float64(3.14), false},
		{"valid string float", "3.14", false},
		{"invalid string", "abc", true},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := Float()
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("Float() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestBool tests the Bool validation rule
func TestBool(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid bool true", true, false},
		{"valid bool false", false, false},
		{"valid string true", "true", false},
		{"valid string false", "false", false},
		{"valid string 1", "1", false},
		{"valid string 0", "0", false},
		{"invalid string", "abc", true},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := Bool()
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("Bool() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestUUID tests the UUID validation rule
func TestUUID(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid UUID", "550e8400-e29b-41d4-a716-446655440000", false},
		{"valid UUID uppercase", "550E8400-E29B-41D4-A716-446655440000", false},
		{"invalid UUID", "not-a-uuid", true},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := UUID()
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("UUID() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestIPv4 tests the IPv4 validation rule
func TestIPv4(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid IPv4", "192.168.1.1", false},
		{"valid IPv4 localhost", "127.0.0.1", false},
		{"invalid IPv4", "999.999.999.999", true},
		{"IPv6 address", "2001:db8::1", true},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := IPv4()
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("IPv4() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestIPv6 tests the IPv6 validation rule
func TestIPv6(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid IPv6", "2001:db8::1", false},
		{"valid IPv6 full", "2001:0db8:0000:0000:0000:0000:0000:0001", false},
		{"invalid IPv6", "not-an-ipv6", true},
		{"IPv4 address", "192.168.1.1", true},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := IPv6()
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("IPv6() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestIP tests the IP validation rule
func TestIP(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid IPv4", "192.168.1.1", false},
		{"valid IPv6", "2001:db8::1", false},
		{"invalid IP", "not-an-ip", true},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := IP()
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("IP() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestMAC tests the MAC validation rule
func TestMAC(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid MAC with colon", "00:1A:2B:3C:4D:5E", false},
		{"valid MAC with dash", "00-1A-2B-3C-4D-5E", false},
		{"invalid MAC", "not-a-mac", true},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := MAC()
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("MAC() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestHex tests the Hex validation rule
func TestHex(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid hex lowercase", "abc123", false},
		{"valid hex uppercase", "ABC123", false},
		{"valid hex mixed", "AbC123", false},
		{"invalid hex", "xyz", true},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := Hex()
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("Hex() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestBase64 tests the Base64 validation rule
func TestBase64(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid base64", "SGVsbG8gV29ybGQ=", false},
		{"valid base64 no padding", "SGVsbG8gV29ybGQ", false},
		{"invalid base64", "not-base64!@#", true},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := Base64()
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("Base64() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestDate tests the Date validation rule
func TestDate(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid date", "2024-01-15", false},
		{"valid date leap year", "2024-02-29", false},
		{"invalid date format", "15-01-2024", true},
		{"invalid date", "2024-13-01", true},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := Date()
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("Date() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestTime tests the Time validation rule
func TestTime(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid time with seconds", "12:34:56", false},
		{"valid time without seconds", "12:34", false},
		{"invalid time", "25:61:61", true},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := Time()
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("Time() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestDateTime tests the DateTime validation rule
func TestDateTime(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid datetime", "2024-01-15 12:34:56", false},
		{"valid RFC3339", "2024-01-15T12:34:56Z", false},
		{"invalid datetime", "not-a-datetime", true},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := DateTime()
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("DateTime() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestArray tests the Array validation rule
func TestArray(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid slice", []int{1, 2, 3}, false},
		{"valid array", [3]int{1, 2, 3}, false},
		{"invalid array", "not-an-array", true},
		{"nil value", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := Array()
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("Array() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestObject tests the Object validation rule
func TestObject(t *testing.T) {
	type TestStruct struct {
		Name string
	}

	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid struct", TestStruct{Name: "test"}, false},
		{"valid map", map[string]string{"key": "value"}, false},
		{"invalid object", "not-an-object", true},
		{"nil value", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := Object()
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("Object() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestRange tests the Range validation rule
func TestRange(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid int in range", 5, false},
		{"valid int below range", 0, true},
		{"valid int above range", 11, true},
		{"valid string in range", "5", false},
		{"nil value", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := Range(1, 10)
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("Range() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestMinFloat tests the MinFloat validation rule
func TestMinFloat(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid float above min", 3.14, false},
		{"valid float below min", 1.0, true},
		{"valid string float", "3.14", false},
		{"nil value", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := MinFloat(2.0)
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("MinFloat() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestMaxFloat tests the MaxFloat validation rule
func TestMaxFloat(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid float below max", 3.14, false},
		{"valid float above max", 10.0, true},
		{"valid string float", "3.14", false},
		{"nil value", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := MaxFloat(5.0)
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("MaxFloat() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestRangeFloat tests the RangeFloat validation rule
func TestRangeFloat(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid float in range", 3.14, false},
		{"valid float below range", 1.0, true},
		{"valid float above range", 10.0, true},
		{"valid string float", "3.14", false},
		{"nil value", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := RangeFloat(2.0, 5.0)
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("RangeFloat() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestEmailList tests the EmailList validation rule
func TestEmailList(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid email list", "test@example.com,admin@example.com", false},
		{"valid single email", "test@example.com", false},
		{"invalid email in list", "test@example.com,invalid-email", true},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := EmailList()
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("EmailList() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestURLList tests the URLList validation rule
func TestURLList(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid URL list", "https://example.com,https://google.com", false},
		{"valid single URL", "https://example.com", false},
		{"invalid URL in list", "https://example.com,not-a-url", true},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := URLList()
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("URLList() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestOneOf tests the OneOf validation rule
func TestOneOf(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid value", "admin", false},
		{"invalid value", "guest", true},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := OneOf("admin", "user", "moderator")
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("OneOf() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestNotEmpty tests the NotEmpty validation rule
func TestNotEmpty(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid non-empty string", "hello", false},
		{"invalid empty string", "", true},
		{"invalid nil", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NotEmpty()
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("NotEmpty() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestNotZero tests the NotZero validation rule
func TestNotZero(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid non-zero int", 42, false},
		{"valid non-zero string", "hello", false},
		{"invalid zero int", 0, true},
		{"invalid empty string", "", true},
		{"invalid nil", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NotZero()
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("NotZero() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestAfterDate tests the AfterDate validation rule
func TestAfterDate(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid date after", "2024-01-15", false},
		{"invalid date before", "2023-12-31", true},
		{"invalid date same", "2024-01-01", true},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := AfterDate("2024-01-01")
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("AfterDate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestBeforeDate tests the BeforeDate validation rule
func TestBeforeDate(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid date before", "2023-12-31", false},
		{"invalid date after", "2024-01-15", true},
		{"invalid date same", "2024-01-01", true},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := BeforeDate("2024-01-01")
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("BeforeDate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestContains tests the Contains validation rule
func TestContains(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid contains", "hello world", false},
		{"invalid contains", "hello", true},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := Contains("world")
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("Contains() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestHasPrefix tests the HasPrefix validation rule
func TestHasPrefix(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid prefix", "hello world", false},
		{"invalid prefix", "world hello", true},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := HasPrefix("hello")
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("HasPrefix() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestHasSuffix tests the HasSuffix validation rule
func TestHasSuffix(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid suffix", "hello world", false},
		{"invalid suffix", "world hello", true},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := HasSuffix("world")
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("HasSuffix() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestMinItems tests the MinItems validation rule
func TestMinItems(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid slice with enough items", []int{1, 2, 3}, false},
		{"invalid slice with too few items", []int{1}, true},
		{"nil value", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := MinItems(2)
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("MinItems() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestMaxItems tests the MaxItems validation rule
func TestMaxItems(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid slice with few items", []int{1, 2}, false},
		{"invalid slice with too many items", []int{1, 2, 3, 4}, true},
		{"nil value", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := MaxItems(3)
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("MaxItems() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestUnique tests the Unique validation rule
func TestUnique(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid unique slice", []int{1, 2, 3}, false},
		{"invalid duplicate slice", []int{1, 2, 2}, true},
		{"nil value", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := Unique()
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("Unique() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestMinMapKeys tests the MinMapKeys validation rule
func TestMinMapKeys(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid map with enough keys", map[string]int{"a": 1, "b": 2}, false},
		{"invalid map with too few keys", map[string]int{"a": 1}, true},
		{"nil value", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := MinMapKeys(2)
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("MinMapKeys() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestMaxMapKeys tests the MaxMapKeys validation rule
func TestMaxMapKeys(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid map with few keys", map[string]int{"a": 1}, false},
		{"invalid map with too many keys", map[string]int{"a": 1, "b": 2, "c": 3}, true},
		{"nil value", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := MaxMapKeys(2)
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("MaxMapKeys() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestCustomRule tests the CustomRule validation rule
func TestCustomRule(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid custom rule", 5, false},
		{"invalid custom rule", 3, true},
		{"nil value", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := CustomRule("even", func(v interface{}) bool {
				if i, ok := v.(int); ok {
					return i >= 5
				}
				return false
			})
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("CustomRule() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestOptional tests the Optional validation rule
func TestOptional(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantError bool
	}{
		{"valid required value", "hello", false},
		{"nil value", nil, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := Optional(Required())
			err := rule.Validate(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("Optional() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestWithMessage tests the WithMessage validation rule
func TestWithMessage(t *testing.T) {
	rule := WithMessage(Min(5), "custom error message")
	err := rule.Validate(3)
	if err == nil {
		t.Error("WithMessage() expected error, got nil")
	}
	if err.Message != "custom error message" {
		t.Errorf("WithMessage() error message = %v, want 'custom error message'", err.Message)
	}
}

// TestWithCode tests the WithCode validation rule
func TestWithCode(t *testing.T) {
	rule := WithCode(Min(5), "custom_code")
	err := rule.Validate(3)
	if err == nil {
		t.Error("WithCode() expected error, got nil")
	}
	if err.Code != "custom_code" {
		t.Errorf("WithCode() error code = %v, want 'custom_code'", err.Code)
	}
}

// TestBindJSON tests the BindJSON method
func TestBindJSON(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"required"`
		Age  int    `validate:"min=18"`
	}

	tests := []struct {
		name      string
		body      string
		wantError bool
	}{
		{"valid JSON", `{"name":"John","age":25}`, false},
		{"invalid JSON", `{"name":"","age":15}`, true},
		{"missing required field", `{"age":25}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")

			var data TestStruct
			validator := NewValidator(nil)
			err := validator.BindJSON(req, &data)

			if (err != nil) != tt.wantError {
				t.Errorf("BindJSON() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestValidate tests the Validate function
func TestValidate(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"required"`
		Age  int    `validate:"min=18"`
	}

	tests := []struct {
		name      string
		data      TestStruct
		wantError bool
	}{
		{"valid data", TestStruct{Name: "John", Age: 25}, false},
		{"invalid data", TestStruct{Name: "", Age: 15}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.data)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestBindJSONConvenience tests the BindJSON convenience function
func TestBindJSONConvenience(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"required"`
		Age  int    `validate:"min=18"`
	}

	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"John","age":25}`))
	req.Header.Set("Content-Type", "application/json")

	var data TestStruct
	err := BindJSON(req, &data)

	if err != nil {
		t.Errorf("BindJSON() error = %v", err)
	}
}
