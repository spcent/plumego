package jsonx

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/spcent/plumego/utils/pool"
)

// FieldString extracts a top-level string field from JSON (best-effort).
// Returns "" if missing, invalid JSON, or not a string.
func FieldString(raw []byte, key string) string {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

// FieldBool extracts a top-level bool field from JSON (best-effort).
// Returns false if missing/invalid/not bool.
func FieldBool(raw []byte, key string) bool {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return false
	}
	v, ok := m[key]
	if !ok {
		return false
	}
	b, _ := v.(bool)
	return b
}

// FieldInt extracts a top-level int field from JSON (best-effort).
// Returns 0 if missing, invalid JSON, or not an int.
func FieldInt(raw []byte, key string) int {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return 0
	}
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return 0
}

// FieldInt64 extracts a top-level int64 field from JSON (best-effort).
// Returns 0 if missing, invalid JSON, or not an int64.
func FieldInt64(raw []byte, key string) int64 {
	m := make(map[string]json.Number)
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&m); err != nil {
		return 0
	}
	v, ok := m[key]
	if !ok {
		return 0
	}
	if i, err := v.Int64(); err == nil {
		return i
	}
	return 0
}

// FieldFloat64 extracts a top-level float64 field from JSON (best-effort).
// Returns 0 if missing, invalid JSON, or not a float64.
func FieldFloat64(raw []byte, key string) float64 {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return 0
	}
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return 0
}

// PathString extracts a nested string field: m[objKey][fieldKey] (best-effort).
// Returns "" if missing, invalid JSON, or type mismatch.
func PathString(raw []byte, objKey, fieldKey string) string {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return ""
	}
	obj, ok := m[objKey].(map[string]any)
	if !ok || obj == nil {
		return ""
	}
	v, ok := obj[fieldKey]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

// PathInt extracts a nested int field: m[objKey][fieldKey] (best-effort).
// Returns 0 if missing, invalid JSON, or type mismatch.
func PathInt(raw []byte, objKey, fieldKey string) int {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return 0
	}
	obj, ok := m[objKey].(map[string]any)
	if !ok || obj == nil {
		return 0
	}
	v, ok := obj[fieldKey]
	if !ok {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return 0
}

// PathInt64 extracts a nested int64 field: m[objKey][fieldKey] (best-effort).
// Returns 0 if missing, invalid JSON, or type mismatch.
func PathInt64(raw []byte, objKey, fieldKey string) int64 {
	var m map[string]map[string]json.Number
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&m); err != nil {
		return 0
	}
	obj, ok := m[objKey]
	if !ok || obj == nil {
		return 0
	}
	v, ok := obj[fieldKey]
	if !ok {
		return 0
	}
	if i, err := v.Int64(); err == nil {
		return i
	}
	return 0
}

// PathFloat64 extracts a nested float64 field: m[objKey][fieldKey] (best-effort).
// Returns 0 if missing, invalid JSON, or type mismatch.
func PathFloat64(raw []byte, objKey, fieldKey string) float64 {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return 0
	}
	obj, ok := m[objKey].(map[string]any)
	if !ok || obj == nil {
		return 0
	}
	v, ok := obj[fieldKey]
	if !ok {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return 0
}

// PathBool extracts a nested bool field: m[objKey][fieldKey] (best-effort).
// Returns false if missing/invalid/not bool.
func PathBool(raw []byte, objKey, fieldKey string) bool {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return false
	}
	obj, ok := m[objKey].(map[string]any)
	if !ok || obj == nil {
		return false
	}
	v, ok := obj[fieldKey]
	if !ok {
		return false
	}
	b, _ := v.(bool)
	return b
}

// ArrayString extracts a top-level string array field from JSON (best-effort).
// Returns nil if missing, invalid JSON, or not a string array.
func ArrayString(raw []byte, key string) []string {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// ArrayInt extracts a top-level int array field from JSON (best-effort).
// Returns nil if missing, invalid JSON, or not an int array.
func ArrayInt(raw []byte, key string) []int {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]int, 0, len(arr))
	for _, item := range arr {
		switch val := item.(type) {
		case float64:
			result = append(result, int(val))
		case string:
			if i, err := strconv.Atoi(val); err == nil {
				result = append(result, i)
			}
		}
	}
	return result
}

// ArrayInt64 extracts a top-level int64 array field from JSON (best-effort).
// Returns nil if missing, invalid JSON, or not an int64 array.
func ArrayInt64(raw []byte, key string) []int64 {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]int64, 0, len(arr))
	for _, item := range arr {
		switch val := item.(type) {
		case float64:
			result = append(result, int64(val))
		case string:
			if i, err := strconv.ParseInt(val, 10, 64); err == nil {
				result = append(result, i)
			}
		}
	}
	return result
}

// ArrayFloat64 extracts a top-level float64 array field from JSON (best-effort).
// Returns nil if missing, invalid JSON, or not a float64 array.
func ArrayFloat64(raw []byte, key string) []float64 {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]float64, 0, len(arr))
	for _, item := range arr {
		switch val := item.(type) {
		case float64:
			result = append(result, val)
		case string:
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				result = append(result, f)
			}
		}
	}
	return result
}

// ArrayBool extracts a top-level bool array field from JSON (best-effort).
// Returns nil if missing, invalid JSON, or not a bool array.
func ArrayBool(raw []byte, key string) []bool {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]bool, 0, len(arr))
	for _, item := range arr {
		if b, ok := item.(bool); ok {
			result = append(result, b)
		}
	}
	return result
}

// MapString extracts a top-level map[string]string field from JSON (best-effort).
// Returns nil if missing, invalid JSON, or not a map[string]string.
func MapString(raw []byte, key string) map[string]string {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	result := make(map[string]string, len(obj))
	for k, val := range obj {
		if s, ok := val.(string); ok {
			result[k] = s
		}
	}
	return result
}

// MapInt extracts a top-level map[string]int field from JSON (best-effort).
// Returns nil if missing, invalid JSON, or not a map[string]int.
func MapInt(raw []byte, key string) map[string]int {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	result := make(map[string]int, len(obj))
	for k, val := range obj {
		switch val := val.(type) {
		case float64:
			result[k] = int(val)
		case string:
			if i, err := strconv.Atoi(val); err == nil {
				result[k] = i
			}
		}
	}
	return result
}

// MapInt64 extracts a top-level map[string]int64 field from JSON (best-effort).
// Returns nil if missing, invalid JSON, or not a map[string]int64.
func MapInt64(raw []byte, key string) map[string]int64 {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	result := make(map[string]int64, len(obj))
	for k, val := range obj {
		switch val := val.(type) {
		case float64:
			result[k] = int64(val)
		case string:
			if i, err := strconv.ParseInt(val, 10, 64); err == nil {
				result[k] = i
			}
		}
	}
	return result
}

// MapFloat64 extracts a top-level map[string]float64 field from JSON (best-effort).
// Returns nil if missing, invalid JSON, or not a map[string]float64.
func MapFloat64(raw []byte, key string) map[string]float64 {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	result := make(map[string]float64, len(obj))
	for k, val := range obj {
		switch val := val.(type) {
		case float64:
			result[k] = val
		case string:
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				result[k] = f
			}
		}
	}
	return result
}

// MapBool extracts a top-level map[string]bool field from JSON (best-effort).
// Returns nil if missing, invalid JSON, or not a map[string]bool.
func MapBool(raw []byte, key string) map[string]bool {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	result := make(map[string]bool, len(obj))
	for k, val := range obj {
		if b, ok := val.(bool); ok {
			result[k] = b
		}
	}
	return result
}

// ArrayMapString extracts a top-level array of map[string]string field from JSON (best-effort).
// Returns nil if missing, invalid JSON, or not an array of map[string]string.
func ArrayMapString(raw []byte, key string) []map[string]string {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]string, 0, len(arr))
	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		m := make(map[string]string, len(obj))
		for k, val := range obj {
			if s, ok := val.(string); ok {
				m[k] = s
			}
		}
		result = append(result, m)
	}
	return result
}

// ArrayMapInt extracts a top-level array of map[string]int field from JSON (best-effort).
// Returns nil if missing, invalid JSON, or not an array of map[string]int.
func ArrayMapInt(raw []byte, key string) []map[string]int {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]int, 0, len(arr))
	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		m := make(map[string]int, len(obj))
		for k, val := range obj {
			switch val := val.(type) {
			case float64:
				m[k] = int(val)
			case string:
				if i, err := strconv.Atoi(val); err == nil {
					m[k] = i
				}
			}
		}
		result = append(result, m)
	}
	return result
}

// ArrayMapInt64 extracts a top-level array of map[string]int64 field from JSON (best-effort).
// Returns nil if missing, invalid JSON, or not an array of map[string]int64.
func ArrayMapInt64(raw []byte, key string) []map[string]int64 {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]int64, 0, len(arr))
	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		m := make(map[string]int64, len(obj))
		for k, val := range obj {
			switch val := val.(type) {
			case float64:
				m[k] = int64(val)
			case string:
				if i, err := strconv.ParseInt(val, 10, 64); err == nil {
					m[k] = i
				}
			}
		}
		result = append(result, m)
	}
	return result
}

// ArrayMapFloat64 extracts a top-level array of map[string]float64 field from JSON (best-effort).
// Returns nil if missing, invalid JSON, or not an array of map[string]float64.
func ArrayMapFloat64(raw []byte, key string) []map[string]float64 {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]float64, 0, len(arr))
	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		m := make(map[string]float64, len(obj))
		for k, val := range obj {
			switch val := val.(type) {
			case float64:
				m[k] = val
			case string:
				if f, err := strconv.ParseFloat(val, 64); err == nil {
					m[k] = f
				}
			}
		}
		result = append(result, m)
	}
	return result
}

// ArrayMapBool extracts a top-level array of map[string]bool field from JSON (best-effort).
// Returns nil if missing, invalid JSON, or not an array of map[string]bool.
func ArrayMapBool(raw []byte, key string) []map[string]bool {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]bool, 0, len(arr))
	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		m := make(map[string]bool, len(obj))
		for k, val := range obj {
			if b, ok := val.(bool); ok {
				m[k] = b
			}
		}
		result = append(result, m)
	}
	return result
}

// ArrayMapString extracts a nested array of map[string]string field from JSON (best-effort).
// Returns nil if missing, invalid JSON, or not an array of map[string]string.
func PathArrayMapString(raw []byte, objKey, fieldKey string) []map[string]string {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	obj, ok := m[objKey].(map[string]any)
	if !ok || obj == nil {
		return nil
	}
	v, ok := obj[fieldKey]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]string, 0, len(arr))
	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		m := make(map[string]string, len(obj))
		for k, val := range obj {
			if s, ok := val.(string); ok {
				m[k] = s
			}
		}
		result = append(result, m)
	}
	return result
}

// PathArrayMapInt extracts a nested array of map[string]int field from JSON (best-effort).
// Returns nil if missing, invalid JSON, or not an array of map[string]int.
func PathArrayMapInt(raw []byte, objKey, fieldKey string) []map[string]int {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	obj, ok := m[objKey].(map[string]any)
	if !ok || obj == nil {
		return nil
	}
	v, ok := obj[fieldKey]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]int, 0, len(arr))
	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		m := make(map[string]int, len(obj))
		for k, val := range obj {
			switch val := val.(type) {
			case float64:
				m[k] = int(val)
			case string:
				if i, err := strconv.Atoi(val); err == nil {
					m[k] = i
				}
			}
		}
		result = append(result, m)
	}
	return result
}

// PathArrayMapInt64 extracts a nested array of map[string]int64 field from JSON (best-effort).
// Returns nil if missing, invalid JSON, or not an array of map[string]int64.
func PathArrayMapInt64(raw []byte, objKey, fieldKey string) []map[string]int64 {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	obj, ok := m[objKey].(map[string]any)
	if !ok || obj == nil {
		return nil
	}
	v, ok := obj[fieldKey]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]int64, 0, len(arr))
	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		m := make(map[string]int64, len(obj))
		for k, val := range obj {
			switch val := val.(type) {
			case float64:
				m[k] = int64(val)
			case string:
				if i, err := strconv.ParseInt(val, 10, 64); err == nil {
					m[k] = i
				}
			}
		}
		result = append(result, m)
	}
	return result
}

// PathArrayMapFloat64 extracts a nested array of map[string]float64 field from JSON (best-effort).
// Returns nil if missing, invalid JSON, or not an array of map[string]float64.
func PathArrayMapFloat64(raw []byte, objKey, fieldKey string) []map[string]float64 {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	obj, ok := m[objKey].(map[string]any)
	if !ok || obj == nil {
		return nil
	}
	v, ok := obj[fieldKey]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]float64, 0, len(arr))
	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		m := make(map[string]float64, len(obj))
		for k, val := range obj {
			switch val := val.(type) {
			case float64:
				m[k] = val
			case string:
				if f, err := strconv.ParseFloat(val, 64); err == nil {
					m[k] = f
				}
			}
		}
		result = append(result, m)
	}
	return result
}

// PathArrayMapBool extracts a nested array of map[string]bool field from JSON (best-effort).
// Returns nil if missing, invalid JSON, or not an array of map[string]bool.
func PathArrayMapBool(raw []byte, objKey, fieldKey string) []map[string]bool {
	m := pool.GetMap()
	defer pool.PutMap(m)

	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	obj, ok := m[objKey].(map[string]any)
	if !ok || obj == nil {
		return nil
	}
	v, ok := obj[fieldKey]
	if !ok {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]map[string]bool, 0, len(arr))
	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		m := make(map[string]bool, len(obj))
		for k, val := range obj {
			if b, ok := val.(bool); ok {
				m[k] = b
			}
		}
		result = append(result, m)
	}
	return result
}
