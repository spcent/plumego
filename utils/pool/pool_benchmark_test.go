package pool

import (
	"encoding/json"
	"testing"
)

// BenchmarkJSONMarshalWithoutPool 测试没有对象池的JSON序列化
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

// BenchmarkJSONMarshalWithPool 测试使用对象池的JSON序列化
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

// BenchmarkJSONUnmarshalWithoutPool 测试没有对象池的JSON反序列化
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

// BenchmarkJSONUnmarshalWithPool 测试使用对象池的JSON反序列化
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

// BenchmarkFieldStringWithoutPool 测试没有对象池的字段提取
func BenchmarkFieldStringWithoutPool(b *testing.B) {
	data := []byte(`{"id":12345,"name":"test","data":{"key":"value"}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 模拟原始实现
		var m map[string]any
		json.Unmarshal(data, &m)
		_ = m["name"].(string)
	}
}

// BenchmarkFieldStringWithPool 测试使用对象池的字段提取
func BenchmarkFieldStringWithPool(b *testing.B) {
	data := []byte(`{"id":12345,"name":"test","data":{"key":"value"}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 使用池化的map来模拟字段提取
		m := GetMap()
		if err := json.Unmarshal(data, &m); err == nil {
			if v, ok := m["name"].(string); ok {
				_ = v
			}
		}
		PutMap(m)
	}
}

// BenchmarkMapAllocation 测试map分配频率
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

// BenchmarkBufferAllocation 测试buffer分配频率
func BenchmarkBufferAllocation(b *testing.B) {
	data := map[string]any{"id": 123, "name": "test"}

	b.Run("WithoutPool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// 每次都创建新的buffer
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
