package jsonx

import "encoding/json"

// FieldString extracts a top-level string field from JSON (best-effort).
// Returns "" if missing, invalid JSON, or not a string.
func FieldString(raw []byte, key string) string {
	var m map[string]any
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
	var m map[string]any
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

// PathString extracts a nested string field: m[objKey][fieldKey] (best-effort).
// Returns "" if missing, invalid JSON, or type mismatch.
func PathString(raw []byte, objKey, fieldKey string) string {
	var m map[string]any
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
