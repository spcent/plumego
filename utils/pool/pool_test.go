package pool

import (
	"encoding/json"
	"reflect"
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

func TestStringSlicePool(t *testing.T) {
	// Test GetStringSlice
	s := GetStringSlice()
	if s == nil {
		t.Error("GetStringSlice returned nil")
	}

	// Test slice is pre-allocated
	if cap(s) < 16 {
		t.Errorf("Expected slice capacity >= 16, got %d", cap(s))
	}

	// Test PutStringSlice
	s = append(s, "a", "b", "c")
	PutStringSlice(s)

	// Get another slice and verify it was reset
	s2 := GetStringSlice()
	if len(s2) != 0 {
		t.Errorf("Expected slice to be reset, got length %d", len(s2))
	}

	// Test PutStringSlice with nil
	PutStringSlice(nil) // Should not panic
}

func TestIntSlicePool(t *testing.T) {
	// Test GetIntSlice
	i := GetIntSlice()
	if i == nil {
		t.Error("GetIntSlice returned nil")
	}

	// Test slice is pre-allocated
	if cap(i) < 16 {
		t.Errorf("Expected slice capacity >= 16, got %d", cap(i))
	}

	// Test PutIntSlice
	i = append(i, 1, 2, 3)
	PutIntSlice(i)

	// Get another slice and verify it was reset
	i2 := GetIntSlice()
	if len(i2) != 0 {
		t.Errorf("Expected slice to be reset, got length %d", len(i2))
	}

	// Test PutIntSlice with nil
	PutIntSlice(nil) // Should not panic
}

func TestInt64SlicePool(t *testing.T) {
	// Test GetInt64Slice
	i := GetInt64Slice()
	if i == nil {
		t.Error("GetInt64Slice returned nil")
	}

	// Test slice is pre-allocated
	if cap(i) < 16 {
		t.Errorf("Expected slice capacity >= 16, got %d", cap(i))
	}

	// Test PutInt64Slice
	i = append(i, 1, 2, 3)
	PutInt64Slice(i)

	// Get another slice and verify it was reset
	i2 := GetInt64Slice()
	if len(i2) != 0 {
		t.Errorf("Expected slice to be reset, got length %d", len(i2))
	}

	// Test PutInt64Slice with nil
	PutInt64Slice(nil) // Should not panic
}

func TestFloat64SlicePool(t *testing.T) {
	// Test GetFloat64Slice
	f := GetFloat64Slice()
	if f == nil {
		t.Error("GetFloat64Slice returned nil")
	}

	// Test slice is pre-allocated
	if cap(f) < 16 {
		t.Errorf("Expected slice capacity >= 16, got %d", cap(f))
	}

	// Test PutFloat64Slice
	f = append(f, 1.1, 2.2, 3.3)
	PutFloat64Slice(f)

	// Get another slice and verify it was reset
	f2 := GetFloat64Slice()
	if len(f2) != 0 {
		t.Errorf("Expected slice to be reset, got length %d", len(f2))
	}

	// Test PutFloat64Slice with nil
	PutFloat64Slice(nil) // Should not panic
}

func TestBoolSlicePool(t *testing.T) {
	// Test GetBoolSlice
	b := GetBoolSlice()
	if b == nil {
		t.Error("GetBoolSlice returned nil")
	}

	// Test slice is pre-allocated
	if cap(b) < 16 {
		t.Errorf("Expected slice capacity >= 16, got %d", cap(b))
	}

	// Test PutBoolSlice
	b = append(b, true, false, true)
	PutBoolSlice(b)

	// Get another slice and verify it was reset
	b2 := GetBoolSlice()
	if len(b2) != 0 {
		t.Errorf("Expected slice to be reset, got length %d", len(b2))
	}

	// Test PutBoolSlice with nil
	PutBoolSlice(nil) // Should not panic
}

func TestMapStringSlicePool(t *testing.T) {
	// Test GetMapStringSlice
	m := GetMapStringSlice()
	if m == nil {
		t.Error("GetMapStringSlice returned nil")
	}

	// Test slice is pre-allocated
	if cap(m) < 8 {
		t.Errorf("Expected slice capacity >= 8, got %d", cap(m))
	}

	// Test PutMapStringSlice
	m = append(m, map[string]any{"key": "value"})
	PutMapStringSlice(m)

	// Get another slice and verify it was reset
	m2 := GetMapStringSlice()
	if len(m2) != 0 {
		t.Errorf("Expected slice to be reset, got length %d", len(m2))
	}

	// Test PutMapStringSlice with nil
	PutMapStringSlice(nil) // Should not panic
}

func TestExtractStringSlice(t *testing.T) {
	data := []byte(`{"tags":["a","b","c"],"invalid":123}`)

	// Extract string slice
	tags, err := ExtractStringSlice(data, "tags")
	if err != nil {
		t.Errorf("ExtractStringSlice failed: %v", err)
	}
	expected := []string{"a", "b", "c"}
	if !reflect.DeepEqual(tags, expected) {
		t.Errorf("Expected %v, got %v", expected, tags)
	}

	// Extract non-string slice
	invalid, err := ExtractStringSlice(data, "invalid")
	if err != nil {
		t.Errorf("ExtractStringSlice failed: %v", err)
	}
	if invalid != nil {
		t.Errorf("Expected nil for non-string slice, got %v", invalid)
	}

	// Extract missing field
	missing, err := ExtractStringSlice(data, "missing")
	if err != nil {
		t.Errorf("ExtractStringSlice failed: %v", err)
	}
	if missing != nil {
		t.Errorf("Expected nil for missing field, got %v", missing)
	}
}

func TestExtractIntSlice(t *testing.T) {
	data := []byte(`{"nums":[1,2,3],"str_nums":["10","20","30"],"invalid":"abc"}`)

	// Extract int slice
	nums, err := ExtractIntSlice(data, "nums")
	if err != nil {
		t.Errorf("ExtractIntSlice failed: %v", err)
	}
	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(nums, expected) {
		t.Errorf("Expected %v, got %v", expected, nums)
	}

	// Extract string int slice
	strNums, err := ExtractIntSlice(data, "str_nums")
	if err != nil {
		t.Errorf("ExtractIntSlice failed: %v", err)
	}
	expectedStr := []int{10, 20, 30}
	if !reflect.DeepEqual(strNums, expectedStr) {
		t.Errorf("Expected %v, got %v", expectedStr, strNums)
	}

	// Extract non-int slice
	invalid, err := ExtractIntSlice(data, "invalid")
	if err != nil {
		t.Errorf("ExtractIntSlice failed: %v", err)
	}
	if invalid != nil {
		t.Errorf("Expected nil for non-int slice, got %v", invalid)
	}
}

func TestExtractInt64Slice(t *testing.T) {
	data := []byte(`{"nums":[1,2,3],"str_nums":["10","20","30"]}`)

	// Extract int64 slice
	nums, err := ExtractInt64Slice(data, "nums")
	if err != nil {
		t.Errorf("ExtractInt64Slice failed: %v", err)
	}
	expected := []int64{1, 2, 3}
	if !reflect.DeepEqual(nums, expected) {
		t.Errorf("Expected %v, got %v", expected, nums)
	}

	// Extract string int64 slice
	strNums, err := ExtractInt64Slice(data, "str_nums")
	if err != nil {
		t.Errorf("ExtractInt64Slice failed: %v", err)
	}
	expectedStr := []int64{10, 20, 30}
	if !reflect.DeepEqual(strNums, expectedStr) {
		t.Errorf("Expected %v, got %v", expectedStr, strNums)
	}
}

func TestExtractFloat64Slice(t *testing.T) {
	data := []byte(`{"prices":[1.1,2.2,3.3],"str_prices":["10.5","20.5","30.5"]}`)

	// Extract float64 slice
	prices, err := ExtractFloat64Slice(data, "prices")
	if err != nil {
		t.Errorf("ExtractFloat64Slice failed: %v", err)
	}
	expected := []float64{1.1, 2.2, 3.3}
	if !reflect.DeepEqual(prices, expected) {
		t.Errorf("Expected %v, got %v", expected, prices)
	}

	// Extract string float64 slice
	strPrices, err := ExtractFloat64Slice(data, "str_prices")
	if err != nil {
		t.Errorf("ExtractFloat64Slice failed: %v", err)
	}
	expectedStr := []float64{10.5, 20.5, 30.5}
	if !reflect.DeepEqual(strPrices, expectedStr) {
		t.Errorf("Expected %v, got %v", expectedStr, strPrices)
	}
}

func TestExtractBoolSlice(t *testing.T) {
	data := []byte(`{"flags":[true,false,true],"invalid":123}`)

	// Extract bool slice
	flags, err := ExtractBoolSlice(data, "flags")
	if err != nil {
		t.Errorf("ExtractBoolSlice failed: %v", err)
	}
	expected := []bool{true, false, true}
	if !reflect.DeepEqual(flags, expected) {
		t.Errorf("Expected %v, got %v", expected, flags)
	}

	// Extract non-bool slice
	invalid, err := ExtractBoolSlice(data, "invalid")
	if err != nil {
		t.Errorf("ExtractBoolSlice failed: %v", err)
	}
	if invalid != nil {
		t.Errorf("Expected nil for non-bool slice, got %v", invalid)
	}
}

func TestExtractMapString(t *testing.T) {
	data := []byte(`{"map":{"a":"1","b":"2"},"invalid":123}`)

	// Extract map[string]string
	m, err := ExtractMapString(data, "map")
	if err != nil {
		t.Errorf("ExtractMapString failed: %v", err)
	}
	expected := map[string]string{"a": "1", "b": "2"}
	if !reflect.DeepEqual(m, expected) {
		t.Errorf("Expected %v, got %v", expected, m)
	}

	// Extract non-map
	invalid, err := ExtractMapString(data, "invalid")
	if err != nil {
		t.Errorf("ExtractMapString failed: %v", err)
	}
	if invalid != nil {
		t.Errorf("Expected nil for non-map, got %v", invalid)
	}
}

func TestExtractMapInt(t *testing.T) {
	data := []byte(`{"map":{"a":1,"b":2},"str_map":{"a":"10","b":"20"}}`)

	// Extract map[string]int
	m, err := ExtractMapInt(data, "map")
	if err != nil {
		t.Errorf("ExtractMapInt failed: %v", err)
	}
	expected := map[string]int{"a": 1, "b": 2}
	if !reflect.DeepEqual(m, expected) {
		t.Errorf("Expected %v, got %v", expected, m)
	}

	// Extract string map[string]int
	strMap, err := ExtractMapInt(data, "str_map")
	if err != nil {
		t.Errorf("ExtractMapInt failed: %v", err)
	}
	expectedStr := map[string]int{"a": 10, "b": 20}
	if !reflect.DeepEqual(strMap, expectedStr) {
		t.Errorf("Expected %v, got %v", expectedStr, strMap)
	}
}

func TestExtractMapInt64(t *testing.T) {
	data := []byte(`{"map":{"a":1,"b":2},"str_map":{"a":"10","b":"20"}}`)

	// Extract map[string]int64
	m, err := ExtractMapInt64(data, "map")
	if err != nil {
		t.Errorf("ExtractMapInt64 failed: %v", err)
	}
	expected := map[string]int64{"a": 1, "b": 2}
	if !reflect.DeepEqual(m, expected) {
		t.Errorf("Expected %v, got %v", expected, m)
	}

	// Extract string map[string]int64
	strMap, err := ExtractMapInt64(data, "str_map")
	if err != nil {
		t.Errorf("ExtractMapInt64 failed: %v", err)
	}
	expectedStr := map[string]int64{"a": 10, "b": 20}
	if !reflect.DeepEqual(strMap, expectedStr) {
		t.Errorf("Expected %v, got %v", expectedStr, strMap)
	}
}

func TestExtractMapFloat64(t *testing.T) {
	data := []byte(`{"map":{"a":1.1,"b":2.2},"str_map":{"a":"10.5","b":"20.5"}}`)

	// Extract map[string]float64
	m, err := ExtractMapFloat64(data, "map")
	if err != nil {
		t.Errorf("ExtractMapFloat64 failed: %v", err)
	}
	expected := map[string]float64{"a": 1.1, "b": 2.2}
	if !reflect.DeepEqual(m, expected) {
		t.Errorf("Expected %v, got %v", expected, m)
	}

	// Extract string map[string]float64
	strMap, err := ExtractMapFloat64(data, "str_map")
	if err != nil {
		t.Errorf("ExtractMapFloat64 failed: %v", err)
	}
	expectedStr := map[string]float64{"a": 10.5, "b": 20.5}
	if !reflect.DeepEqual(strMap, expectedStr) {
		t.Errorf("Expected %v, got %v", expectedStr, strMap)
	}
}

func TestExtractMapBool(t *testing.T) {
	data := []byte(`{"map":{"a":true,"b":false},"invalid":123}`)

	// Extract map[string]bool
	m, err := ExtractMapBool(data, "map")
	if err != nil {
		t.Errorf("ExtractMapBool failed: %v", err)
	}
	expected := map[string]bool{"a": true, "b": false}
	if !reflect.DeepEqual(m, expected) {
		t.Errorf("Expected %v, got %v", expected, m)
	}

	// Extract non-map
	invalid, err := ExtractMapBool(data, "invalid")
	if err != nil {
		t.Errorf("ExtractMapBool failed: %v", err)
	}
	if invalid != nil {
		t.Errorf("Expected nil for non-map, got %v", invalid)
	}
}

func TestExtractArrayMapString(t *testing.T) {
	data := []byte(`{"items":[{"a":"1","b":"2"},{"c":"3","d":"4"}]}`)

	// Extract array of map[string]string
	items, err := ExtractArrayMapString(data, "items")
	if err != nil {
		t.Errorf("ExtractArrayMapString failed: %v", err)
	}
	expected := []map[string]string{
		{"a": "1", "b": "2"},
		{"c": "3", "d": "4"},
	}
	if !reflect.DeepEqual(items, expected) {
		t.Errorf("Expected %v, got %v", expected, items)
	}
}

func TestExtractArrayMapInt(t *testing.T) {
	data := []byte(`{"items":[{"a":1,"b":2},{"c":3,"d":4}],"str_items":[{"a":"10","b":"20"}]}`)

	// Extract array of map[string]int
	items, err := ExtractArrayMapInt(data, "items")
	if err != nil {
		t.Errorf("ExtractArrayMapInt failed: %v", err)
	}
	expected := []map[string]int{
		{"a": 1, "b": 2},
		{"c": 3, "d": 4},
	}
	if !reflect.DeepEqual(items, expected) {
		t.Errorf("Expected %v, got %v", expected, items)
	}

	// Extract string array of map[string]int
	strItems, err := ExtractArrayMapInt(data, "str_items")
	if err != nil {
		t.Errorf("ExtractArrayMapInt failed: %v", err)
	}
	expectedStr := []map[string]int{
		{"a": 10, "b": 20},
	}
	if !reflect.DeepEqual(strItems, expectedStr) {
		t.Errorf("Expected %v, got %v", expectedStr, strItems)
	}
}

func TestExtractArrayMapInt64(t *testing.T) {
	data := []byte(`{"items":[{"a":1,"b":2},{"c":3,"d":4}],"str_items":[{"a":"10","b":"20"}]}`)

	// Extract array of map[string]int64
	items, err := ExtractArrayMapInt64(data, "items")
	if err != nil {
		t.Errorf("ExtractArrayMapInt64 failed: %v", err)
	}
	expected := []map[string]int64{
		{"a": 1, "b": 2},
		{"c": 3, "d": 4},
	}
	if !reflect.DeepEqual(items, expected) {
		t.Errorf("Expected %v, got %v", expected, items)
	}

	// Extract string array of map[string]int64
	strItems, err := ExtractArrayMapInt64(data, "str_items")
	if err != nil {
		t.Errorf("ExtractArrayMapInt64 failed: %v", err)
	}
	expectedStr := []map[string]int64{
		{"a": 10, "b": 20},
	}
	if !reflect.DeepEqual(strItems, expectedStr) {
		t.Errorf("Expected %v, got %v", expectedStr, strItems)
	}
}

func TestExtractArrayMapFloat64(t *testing.T) {
	data := []byte(`{"items":[{"a":1.1,"b":2.2},{"c":3.3,"d":4.4}],"str_items":[{"a":"10.5","b":"20.5"}]}`)

	// Extract array of map[string]float64
	items, err := ExtractArrayMapFloat64(data, "items")
	if err != nil {
		t.Errorf("ExtractArrayMapFloat64 failed: %v", err)
	}
	expected := []map[string]float64{
		{"a": 1.1, "b": 2.2},
		{"c": 3.3, "d": 4.4},
	}
	if !reflect.DeepEqual(items, expected) {
		t.Errorf("Expected %v, got %v", expected, items)
	}

	// Extract string array of map[string]float64
	strItems, err := ExtractArrayMapFloat64(data, "str_items")
	if err != nil {
		t.Errorf("ExtractArrayMapFloat64 failed: %v", err)
	}
	expectedStr := []map[string]float64{
		{"a": 10.5, "b": 20.5},
	}
	if !reflect.DeepEqual(strItems, expectedStr) {
		t.Errorf("Expected %v, got %v", expectedStr, strItems)
	}
}

func TestExtractArrayMapBool(t *testing.T) {
	data := []byte(`{"items":[{"a":true,"b":false},{"c":true,"d":false}]}`)

	// Extract array of map[string]bool
	items, err := ExtractArrayMapBool(data, "items")
	if err != nil {
		t.Errorf("ExtractArrayMapBool failed: %v", err)
	}
	expected := []map[string]bool{
		{"a": true, "b": false},
		{"c": true, "d": false},
	}
	if !reflect.DeepEqual(items, expected) {
		t.Errorf("Expected %v, got %v", expected, items)
	}
}
