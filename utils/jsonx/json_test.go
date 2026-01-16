package jsonx

import (
	"reflect"
	"testing"
)

func TestFieldString(t *testing.T) {
	raw := []byte(`{"id":"evt_1","livemode":true,"obj":{"name":"x"}}`)
	if FieldString(raw, "id") != "evt_1" {
		t.Fatalf("id mismatch")
	}
	if FieldBool(raw, "livemode") != true {
		t.Fatalf("livemode mismatch")
	}
	if PathString(raw, "obj", "name") != "x" {
		t.Fatalf("path mismatch")
	}
}

func TestFieldInt(t *testing.T) {
	raw := []byte(`{"count":42,"str_count":"100","invalid":"abc"}`)
	if FieldInt(raw, "count") != 42 {
		t.Fatalf("count mismatch")
	}
	if FieldInt(raw, "str_count") != 100 {
		t.Fatalf("str_count mismatch")
	}
	if FieldInt(raw, "missing") != 0 {
		t.Fatalf("missing should return 0")
	}
	if FieldInt(raw, "invalid") != 0 {
		t.Fatalf("invalid should return 0")
	}
}

func TestFieldInt64(t *testing.T) {
	raw := []byte(`{"count":9223372036854775807,"str_count":"123"}`)
	if FieldInt64(raw, "count") != 9223372036854775807 {
		t.Fatalf("count mismatch")
	}
	if FieldInt64(raw, "str_count") != 123 {
		t.Fatalf("str_count mismatch")
	}
	if FieldInt64(raw, "missing") != 0 {
		t.Fatalf("missing should return 0")
	}
}

func TestFieldFloat64(t *testing.T) {
	raw := []byte(`{"price":99.99,"str_price":"123.45","invalid":"abc"}`)
	if FieldFloat64(raw, "price") != 99.99 {
		t.Fatalf("price mismatch")
	}
	if FieldFloat64(raw, "str_price") != 123.45 {
		t.Fatalf("str_price mismatch")
	}
	if FieldFloat64(raw, "missing") != 0 {
		t.Fatalf("missing should return 0")
	}
	if FieldFloat64(raw, "invalid") != 0 {
		t.Fatalf("invalid should return 0")
	}
}

func TestPathInt(t *testing.T) {
	raw := []byte(`{"obj":{"count":42,"str_count":"100"}}`)
	if PathInt(raw, "obj", "count") != 42 {
		t.Fatalf("count mismatch")
	}
	if PathInt(raw, "obj", "str_count") != 100 {
		t.Fatalf("str_count mismatch")
	}
	if PathInt(raw, "obj", "missing") != 0 {
		t.Fatalf("missing should return 0")
	}
}

func TestPathInt64(t *testing.T) {
	raw := []byte(`{"obj":{"count":9223372036854775807,"str_count":"123"}}`)
	if PathInt64(raw, "obj", "count") != 9223372036854775807 {
		t.Fatalf("count mismatch")
	}
	if PathInt64(raw, "obj", "str_count") != 123 {
		t.Fatalf("str_count mismatch")
	}
	if PathInt64(raw, "obj", "missing") != 0 {
		t.Fatalf("missing should return 0")
	}
}

func TestPathFloat64(t *testing.T) {
	raw := []byte(`{"obj":{"price":99.99,"str_price":"123.45"}}`)
	if PathFloat64(raw, "obj", "price") != 99.99 {
		t.Fatalf("price mismatch")
	}
	if PathFloat64(raw, "obj", "str_price") != 123.45 {
		t.Fatalf("str_price mismatch")
	}
	if PathFloat64(raw, "obj", "missing") != 0 {
		t.Fatalf("missing should return 0")
	}
}

func TestPathBool(t *testing.T) {
	raw := []byte(`{"obj":{"enabled":true,"disabled":false}}`)
	if PathBool(raw, "obj", "enabled") != true {
		t.Fatalf("enabled mismatch")
	}
	if PathBool(raw, "obj", "disabled") != false {
		t.Fatalf("disabled mismatch")
	}
	if PathBool(raw, "obj", "missing") != false {
		t.Fatalf("missing should return false")
	}
}

func TestArrayString(t *testing.T) {
	raw := []byte(`{"tags":["a","b","c"],"invalid":123}`)
	expected := []string{"a", "b", "c"}
	result := ArrayString(raw, "tags")
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("tags mismatch: got %v, want %v", result, expected)
	}
	if ArrayString(raw, "invalid") != nil {
		t.Fatalf("invalid should return nil")
	}
	if ArrayString(raw, "missing") != nil {
		t.Fatalf("missing should return nil")
	}
}

func TestArrayInt(t *testing.T) {
	raw := []byte(`{"nums":[1,2,3],"str_nums":["10","20","30"],"invalid":"abc"}`)
	expected := []int{1, 2, 3}
	result := ArrayInt(raw, "nums")
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("nums mismatch: got %v, want %v", result, expected)
	}
	expectedStr := []int{10, 20, 30}
	resultStr := ArrayInt(raw, "str_nums")
	if !reflect.DeepEqual(resultStr, expectedStr) {
		t.Fatalf("str_nums mismatch: got %v, want %v", resultStr, expectedStr)
	}
	if ArrayInt(raw, "invalid") != nil {
		t.Fatalf("invalid should return nil")
	}
	if ArrayInt(raw, "missing") != nil {
		t.Fatalf("missing should return nil")
	}
}

func TestArrayInt64(t *testing.T) {
	raw := []byte(`{"nums":[1,2,3],"str_nums":["10","20","30"]}`)
	expected := []int64{1, 2, 3}
	result := ArrayInt64(raw, "nums")
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("nums mismatch: got %v, want %v", result, expected)
	}
	expectedStr := []int64{10, 20, 30}
	resultStr := ArrayInt64(raw, "str_nums")
	if !reflect.DeepEqual(resultStr, expectedStr) {
		t.Fatalf("str_nums mismatch: got %v, want %v", resultStr, expectedStr)
	}
}

func TestArrayFloat64(t *testing.T) {
	raw := []byte(`{"prices":[1.1,2.2,3.3],"str_prices":["10.5","20.5","30.5"]}`)
	expected := []float64{1.1, 2.2, 3.3}
	result := ArrayFloat64(raw, "prices")
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("prices mismatch: got %v, want %v", result, expected)
	}
	expectedStr := []float64{10.5, 20.5, 30.5}
	resultStr := ArrayFloat64(raw, "str_prices")
	if !reflect.DeepEqual(resultStr, expectedStr) {
		t.Fatalf("str_prices mismatch: got %v, want %v", resultStr, expectedStr)
	}
}

func TestArrayBool(t *testing.T) {
	raw := []byte(`{"flags":[true,false,true],"invalid":123}`)
	expected := []bool{true, false, true}
	result := ArrayBool(raw, "flags")
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("flags mismatch: got %v, want %v", result, expected)
	}
	if ArrayBool(raw, "invalid") != nil {
		t.Fatalf("invalid should return nil")
	}
	if ArrayBool(raw, "missing") != nil {
		t.Fatalf("missing should return nil")
	}
}

func TestMapString(t *testing.T) {
	raw := []byte(`{"map":{"a":"1","b":"2"},"invalid":123}`)
	expected := map[string]string{"a": "1", "b": "2"}
	result := MapString(raw, "map")
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("map mismatch: got %v, want %v", result, expected)
	}
	if MapString(raw, "invalid") != nil {
		t.Fatalf("invalid should return nil")
	}
	if MapString(raw, "missing") != nil {
		t.Fatalf("missing should return nil")
	}
}

func TestMapInt(t *testing.T) {
	raw := []byte(`{"map":{"a":1,"b":2},"str_map":{"a":"10","b":"20"}}`)
	expected := map[string]int{"a": 1, "b": 2}
	result := MapInt(raw, "map")
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("map mismatch: got %v, want %v", result, expected)
	}
	expectedStr := map[string]int{"a": 10, "b": 20}
	resultStr := MapInt(raw, "str_map")
	if !reflect.DeepEqual(resultStr, expectedStr) {
		t.Fatalf("str_map mismatch: got %v, want %v", resultStr, expectedStr)
	}
}

func TestMapInt64(t *testing.T) {
	raw := []byte(`{"map":{"a":1,"b":2},"str_map":{"a":"10","b":"20"}}`)
	expected := map[string]int64{"a": 1, "b": 2}
	result := MapInt64(raw, "map")
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("map mismatch: got %v, want %v", result, expected)
	}
	expectedStr := map[string]int64{"a": 10, "b": 20}
	resultStr := MapInt64(raw, "str_map")
	if !reflect.DeepEqual(resultStr, expectedStr) {
		t.Fatalf("str_map mismatch: got %v, want %v", resultStr, expectedStr)
	}
}

func TestMapFloat64(t *testing.T) {
	raw := []byte(`{"map":{"a":1.1,"b":2.2},"str_map":{"a":"10.5","b":"20.5"}}`)
	expected := map[string]float64{"a": 1.1, "b": 2.2}
	result := MapFloat64(raw, "map")
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("map mismatch: got %v, want %v", result, expected)
	}
	expectedStr := map[string]float64{"a": 10.5, "b": 20.5}
	resultStr := MapFloat64(raw, "str_map")
	if !reflect.DeepEqual(resultStr, expectedStr) {
		t.Fatalf("str_map mismatch: got %v, want %v", resultStr, expectedStr)
	}
}

func TestMapBool(t *testing.T) {
	raw := []byte(`{"map":{"a":true,"b":false},"invalid":123}`)
	expected := map[string]bool{"a": true, "b": false}
	result := MapBool(raw, "map")
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("map mismatch: got %v, want %v", result, expected)
	}
	if MapBool(raw, "invalid") != nil {
		t.Fatalf("invalid should return nil")
	}
	if MapBool(raw, "missing") != nil {
		t.Fatalf("missing should return nil")
	}
}

func TestArrayMapString(t *testing.T) {
	raw := []byte(`{"items":[{"a":"1","b":"2"},{"c":"3","d":"4"}]}`)
	expected := []map[string]string{
		{"a": "1", "b": "2"},
		{"c": "3", "d": "4"},
	}
	result := ArrayMapString(raw, "items")
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("items mismatch: got %v, want %v", result, expected)
	}
	if ArrayMapString(raw, "missing") != nil {
		t.Fatalf("missing should return nil")
	}
}

func TestArrayMapInt(t *testing.T) {
	raw := []byte(`{"items":[{"a":1,"b":2},{"c":3,"d":4}],"str_items":[{"a":"10","b":"20"}]}`)
	expected := []map[string]int{
		{"a": 1, "b": 2},
		{"c": 3, "d": 4},
	}
	result := ArrayMapInt(raw, "items")
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("items mismatch: got %v, want %v", result, expected)
	}
	expectedStr := []map[string]int{
		{"a": 10, "b": 20},
	}
	resultStr := ArrayMapInt(raw, "str_items")
	if !reflect.DeepEqual(resultStr, expectedStr) {
		t.Fatalf("str_items mismatch: got %v, want %v", resultStr, expectedStr)
	}
}

func TestArrayMapInt64(t *testing.T) {
	raw := []byte(`{"items":[{"a":1,"b":2},{"c":3,"d":4}],"str_items":[{"a":"10","b":"20"}]}`)
	expected := []map[string]int64{
		{"a": 1, "b": 2},
		{"c": 3, "d": 4},
	}
	result := ArrayMapInt64(raw, "items")
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("items mismatch: got %v, want %v", result, expected)
	}
	expectedStr := []map[string]int64{
		{"a": 10, "b": 20},
	}
	resultStr := ArrayMapInt64(raw, "str_items")
	if !reflect.DeepEqual(resultStr, expectedStr) {
		t.Fatalf("str_items mismatch: got %v, want %v", resultStr, expectedStr)
	}
}

func TestArrayMapFloat64(t *testing.T) {
	raw := []byte(`{"items":[{"a":1.1,"b":2.2},{"c":3.3,"d":4.4}],"str_items":[{"a":"10.5","b":"20.5"}]}`)
	expected := []map[string]float64{
		{"a": 1.1, "b": 2.2},
		{"c": 3.3, "d": 4.4},
	}
	result := ArrayMapFloat64(raw, "items")
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("items mismatch: got %v, want %v", result, expected)
	}
	expectedStr := []map[string]float64{
		{"a": 10.5, "b": 20.5},
	}
	resultStr := ArrayMapFloat64(raw, "str_items")
	if !reflect.DeepEqual(resultStr, expectedStr) {
		t.Fatalf("str_items mismatch: got %v, want %v", resultStr, expectedStr)
	}
}

func TestArrayMapBool(t *testing.T) {
	raw := []byte(`{"items":[{"a":true,"b":false},{"c":true,"d":false}]}`)
	expected := []map[string]bool{
		{"a": true, "b": false},
		{"c": true, "d": false},
	}
	result := ArrayMapBool(raw, "items")
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("items mismatch: got %v, want %v", result, expected)
	}
	if ArrayMapBool(raw, "missing") != nil {
		t.Fatalf("missing should return nil")
	}
}

func TestPathArrayMapString(t *testing.T) {
	raw := []byte(`{"obj":{"items":[{"a":"1","b":"2"},{"c":"3","d":"4"}]}}`)
	expected := []map[string]string{
		{"a": "1", "b": "2"},
		{"c": "3", "d": "4"},
	}
	result := PathArrayMapString(raw, "obj", "items")
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("items mismatch: got %v, want %v", result, expected)
	}
	if PathArrayMapString(raw, "obj", "missing") != nil {
		t.Fatalf("missing should return nil")
	}
}

func TestPathArrayMapInt(t *testing.T) {
	raw := []byte(`{"obj":{"items":[{"a":1,"b":2},{"c":3,"d":4}],"str_items":[{"a":"10","b":"20"}]}}`)
	expected := []map[string]int{
		{"a": 1, "b": 2},
		{"c": 3, "d": 4},
	}
	result := PathArrayMapInt(raw, "obj", "items")
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("items mismatch: got %v, want %v", result, expected)
	}
	expectedStr := []map[string]int{
		{"a": 10, "b": 20},
	}
	resultStr := PathArrayMapInt(raw, "obj", "str_items")
	if !reflect.DeepEqual(resultStr, expectedStr) {
		t.Fatalf("str_items mismatch: got %v, want %v", resultStr, expectedStr)
	}
}

func TestPathArrayMapInt64(t *testing.T) {
	raw := []byte(`{"obj":{"items":[{"a":1,"b":2},{"c":3,"d":4}],"str_items":[{"a":"10","b":"20"}]}}`)
	expected := []map[string]int64{
		{"a": 1, "b": 2},
		{"c": 3, "d": 4},
	}
	result := PathArrayMapInt64(raw, "obj", "items")
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("items mismatch: got %v, want %v", result, expected)
	}
	expectedStr := []map[string]int64{
		{"a": 10, "b": 20},
	}
	resultStr := PathArrayMapInt64(raw, "obj", "str_items")
	if !reflect.DeepEqual(resultStr, expectedStr) {
		t.Fatalf("str_items mismatch: got %v, want %v", resultStr, expectedStr)
	}
}

func TestPathArrayMapFloat64(t *testing.T) {
	raw := []byte(`{"obj":{"items":[{"a":1.1,"b":2.2},{"c":3.3,"d":4.4}],"str_items":[{"a":"10.5","b":"20.5"}]}}`)
	expected := []map[string]float64{
		{"a": 1.1, "b": 2.2},
		{"c": 3.3, "d": 4.4},
	}
	result := PathArrayMapFloat64(raw, "obj", "items")
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("items mismatch: got %v, want %v", result, expected)
	}
	expectedStr := []map[string]float64{
		{"a": 10.5, "b": 20.5},
	}
	resultStr := PathArrayMapFloat64(raw, "obj", "str_items")
	if !reflect.DeepEqual(resultStr, expectedStr) {
		t.Fatalf("str_items mismatch: got %v, want %v", resultStr, expectedStr)
	}
}

func TestPathArrayMapBool(t *testing.T) {
	raw := []byte(`{"obj":{"items":[{"a":true,"b":false},{"c":true,"d":false}]}}`)
	expected := []map[string]bool{
		{"a": true, "b": false},
		{"c": true, "d": false},
	}
	result := PathArrayMapBool(raw, "obj", "items")
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("items mismatch: got %v, want %v", result, expected)
	}
	if PathArrayMapBool(raw, "obj", "missing") != nil {
		t.Fatalf("missing should return nil")
	}
}
