package pool

import (
	"encoding/json"
	"testing"
)

func TestJSONBufferPool(t *testing.T) {
	// Test GetBuffer
	buf := GetBuffer()
	if buf == nil {
		t.Error("GetBuffer returned nil")
	}

	// Test buffer is pre-allocated
	if buf.Cap() < 1024 {
		t.Errorf("Expected buffer capacity >= 1024, got %d", buf.Cap())
	}

	// Test PutBuffer
	buf.WriteString("test data")
	PutBuffer(buf)

	// Get another buffer and verify it was reset
	buf2 := GetBuffer()
	if buf2.Len() != 0 {
		t.Errorf("Expected buffer to be reset, got length %d", buf2.Len())
	}

	// Test PutBuffer with nil
	PutBuffer(nil) // Should not panic
}

func TestMapPool(t *testing.T) {
	// Test GetMap
	m := GetMap()
	if m == nil {
		t.Error("GetMap returned nil")
	}

	// Test PutMap
	m["key"] = "value"
	PutMap(m)

	// Get another map and verify it was cleared
	m2 := GetMap()
	if len(m2) != 0 {
		t.Errorf("Expected map to be cleared, got length %d", len(m2))
	}

	// Test PutMap with nil
	PutMap(nil) // Should not panic
}

func TestByteSlicePool(t *testing.T) {
	// Test GetBytes
	b := GetBytes()
	if b == nil {
		t.Error("GetBytes returned nil")
	}

	// Test slice is pre-allocated
	if cap(b) < 512 {
		t.Errorf("Expected slice capacity >= 512, got %d", cap(b))
	}

	// Test PutBytes
	b = append(b, 1, 2, 3, 4, 5)
	PutBytes(b)

	// Get another slice and verify it was reset
	b2 := GetBytes()
	if len(b2) != 0 {
		t.Errorf("Expected slice to be reset, got length %d", len(b2))
	}

	// Test PutBytes with nil
	PutBytes(nil) // Should not panic
}

func TestJSONMarshal(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	data := TestStruct{Name: "test", Value: 42}
	result, err := JSONMarshal(data)

	if err != nil {
		t.Errorf("JSONMarshal failed: %v", err)
	}

	expected := `{"name":"test","value":42}` + "\n"
	if string(result) != expected {
		t.Errorf("Expected %s, got %s", expected, string(result))
	}

	// Verify buffer was returned to pool (should not leak)
}

func TestJSONUnmarshal(t *testing.T) {
	data := []byte(`{"name":"test","value":42}`)

	// Test with map
	var m map[string]any
	err := JSONUnmarshal(data, &m)
	if err != nil {
		t.Errorf("JSONUnmarshal failed: %v", err)
	}
	if m["name"] != "test" || m["value"] != float64(42) {
		t.Errorf("Unexpected map values: %v", m)
	}

	// Test with struct
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	var s TestStruct
	err = JSONUnmarshal(data, &s)
	if err != nil {
		t.Errorf("JSONUnmarshal failed: %v", err)
	}
	if s.Name != "test" || s.Value != 42 {
		t.Errorf("Unexpected struct values: %+v", s)
	}
}

func TestExtractField(t *testing.T) {
	data := []byte(`{"name":"test","value":42,"nested":{"key":"value"}}`)

	// Extract string field
	name, err := ExtractField(data, "name")
	if err != nil {
		t.Errorf("ExtractField failed: %v", err)
	}
	if name != "test" {
		t.Errorf("Expected 'test', got %v", name)
	}

	// Extract number field
	value, err := ExtractField(data, "value")
	if err != nil {
		t.Errorf("ExtractField failed: %v", err)
	}
	if value != float64(42) {
		t.Errorf("Expected 42, got %v", value)
	}

	// Extract nested field
	nested, err := ExtractField(data, "nested")
	if err != nil {
		t.Errorf("ExtractField failed: %v", err)
	}
	nestedMap, ok := nested.(map[string]any)
	if !ok || nestedMap["key"] != "value" {
		t.Errorf("Expected nested map with key='value', got %v", nested)
	}

	// Extract non-existent field
	missing, err := ExtractField(data, "missing")
	if err != nil {
		t.Errorf("ExtractField failed: %v", err)
	}
	if missing != nil {
		t.Errorf("Expected nil for missing field, got %v", missing)
	}

	// Test with invalid JSON
	_, err = ExtractField([]byte("invalid json"), "key")
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestJSONUnmarshalWithInvalidData(t *testing.T) {
	// Test with invalid JSON
	var m map[string]any
	err := JSONUnmarshal([]byte("invalid"), &m)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	// Test with nil target
	err = JSONUnmarshal([]byte(`{"key":"value"}`), nil)
	if err == nil {
		t.Error("Expected error for nil target")
	}
}

func TestJSONMarshalWithComplexData(t *testing.T) {
	// Test with nested structures
	data := map[string]any{
		"string": "value",
		"number": 123,
		"bool":   true,
		"array":  []int{1, 2, 3},
		"nested": map[string]any{"key": "value"},
	}

	result, err := JSONMarshal(data)
	if err != nil {
		t.Errorf("JSONMarshal failed: %v", err)
	}

	// Verify it's valid JSON
	var decoded map[string]any
	if err := json.Unmarshal(result, &decoded); err != nil {
		t.Errorf("Result is not valid JSON: %v", err)
	}
}

func TestPoolConcurrency(t *testing.T) {
	// Test concurrent access to pools
	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func() {
			buf := GetBuffer()
			buf.WriteString("concurrent")
			PutBuffer(buf)

			m := GetMap()
			m["key"] = "value"
			PutMap(m)

			b := GetBytes()
			b = append(b, 1, 2, 3)
			PutBytes(b)

			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}
