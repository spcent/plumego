package pool

import (
	"encoding/json"
	"testing"
)

// BenchmarkJSONMarshalWithoutPool tests JSON serialization without object pool
func BenchmarkJSONMarshalWithoutPool(b *testing.B) {
	data := map[string]any{
		"id":   12345,
		"name": "test",
		"data": map[string]any{"key": "value"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkJSONMarshalWithPool tests JSON serialization with object pool
func BenchmarkJSONMarshalWithPool(b *testing.B) {
	data := map[string]any{
		"id":   12345,
		"name": "test",
		"data": map[string]any{"key": "value"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := JSONMarshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkJSONUnmarshalWithoutPool tests JSON deserialization without object pool
func BenchmarkJSONUnmarshalWithoutPool(b *testing.B) {
	data := []byte(`{"id":12345,"name":"test","data":{"key":"value"}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result map[string]any
		if err := json.Unmarshal(data, &result); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkJSONUnmarshalWithPool tests JSON deserialization with object pool
func BenchmarkJSONUnmarshalWithPool(b *testing.B) {
	data := []byte(`{"id":12345,"name":"test","data":{"key":"value"}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := make(map[string]any)
		if err := JSONUnmarshal(data, &result); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkFieldStringWithoutPool tests field extraction without object pool
func BenchmarkFieldStringWithoutPool(b *testing.B) {
	data := []byte(`{"id":12345,"name":"test","data":{"key":"value"}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate original implementation
		var m map[string]any
		json.Unmarshal(data, &m)
		_ = m["name"].(string)
	}
}

// BenchmarkFieldStringWithPool tests field extraction with object pool
func BenchmarkFieldStringWithPool(b *testing.B) {
	data := []byte(`{"id":12345,"name":"test","data":{"key":"value"}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use pooled map to simulate field extraction
		m := GetMap()
		if err := json.Unmarshal(data, &m); err == nil {
			if v, ok := m["name"].(string); ok {
				_ = v
			}
		}
		PutMap(m)
	}
}

// BenchmarkMapAllocation tests map allocation frequency
func BenchmarkMapAllocation(b *testing.B) {
	b.Run("WithoutPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := make(map[string]any, 16)
			m["id"] = i
			m["name"] = "test"
			_ = m
		}
	})

	b.Run("WithPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := GetMap()
			m["id"] = i
			m["name"] = "test"
			PutMap(m)
		}
	})
}

// BenchmarkBufferAllocation tests buffer allocation frequency
func BenchmarkBufferAllocation(b *testing.B) {
	data := map[string]any{"id": 123, "name": "test"}

	b.Run("WithoutPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Always create new buffer
			_, _ = json.Marshal(data)
		}
	})

	b.Run("WithPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := GetBuffer()
			enc := json.NewEncoder(buf)
			_ = enc.Encode(data)
			buf.Reset()
			PutBuffer(buf)
		}
	})
}
