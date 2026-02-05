package bind

import (
	"fmt"
	"reflect"
	"strings"
)

// Redactor masks sensitive fields for logging.
type Redactor struct {
	Mask      string
	Fields    map[string]struct{}
	Tags      []string
	MaxDepth  int
	TagValues map[string]struct{}
}

// DefaultRedactor creates a redactor with common defaults.
func DefaultRedactor() *Redactor {
	return NewRedactor(RedactConfig{
		Mask:     "***",
		MaxDepth: 4,
		Fields: []string{
			"password",
			"pass",
			"secret",
			"token",
			"access_token",
			"refresh_token",
			"authorization",
			"apikey",
			"api_key",
			"signature",
			"sign",
		},
		Tags: []string{"mask", "sensitive", "secret", "redact"},
	})
}

// RedactConfig configures Redactor behavior.
type RedactConfig struct {
	Mask     string
	Fields   []string
	Tags     []string
	MaxDepth int
}

// NewRedactor builds a redactor from config.
func NewRedactor(cfg RedactConfig) *Redactor {
	mask := cfg.Mask
	if mask == "" {
		mask = "***"
	}
	maxDepth := cfg.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 4
	}

	fields := make(map[string]struct{}, len(cfg.Fields))
	for _, f := range cfg.Fields {
		if name := strings.ToLower(strings.TrimSpace(f)); name != "" {
			fields[name] = struct{}{}
		}
	}

	tags := cfg.Tags
	if len(tags) == 0 {
		tags = []string{"mask", "sensitive", "secret", "redact"}
	}
	tagValues := map[string]struct{}{
		"true":   {},
		"1":      {},
		"yes":    {},
		"on":     {},
		"y":      {},
		"mask":   {},
		"redact": {},
	}

	return &Redactor{
		Mask:      mask,
		Fields:    fields,
		Tags:      tags,
		MaxDepth:  maxDepth,
		TagValues: tagValues,
	}
}

// Redact returns a redacted copy of the input.
func (r *Redactor) Redact(value any) any {
	if r == nil {
		return value
	}
	seen := make(map[uintptr]struct{})
	return r.redactValue(reflect.ValueOf(value), 0, seen)
}

func (r *Redactor) redactValue(v reflect.Value, depth int, seen map[uintptr]struct{}) any {
	if !v.IsValid() {
		return nil
	}
	if depth >= r.MaxDepth {
		return r.Mask
	}

	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		ptr := v.Pointer()
		if _, ok := seen[ptr]; ok {
			return r.Mask
		}
		seen[ptr] = struct{}{}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		if v.Type().PkgPath() == "time" && v.Type().Name() == "Time" {
			return v.Interface()
		}
		return r.redactStruct(v, depth, seen)
	case reflect.Map:
		return r.redactMap(v, depth, seen)
	case reflect.Slice, reflect.Array:
		length := v.Len()
		items := make([]any, length)
		for i := 0; i < length; i++ {
			items[i] = r.redactValue(v.Index(i), depth+1, seen)
		}
		return items
	default:
		return v.Interface()
	}
}

func (r *Redactor) redactStruct(v reflect.Value, depth int, seen map[uintptr]struct{}) map[string]any {
	out := make(map[string]any)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}

		name := field.Name
		if tag := field.Tag.Get("json"); tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] == "-" {
				continue
			}
			if parts[0] != "" {
				name = parts[0]
			}
		}

		if r.shouldRedactField(name, field.Tag) {
			out[name] = r.Mask
			continue
		}

		out[name] = r.redactValue(v.Field(i), depth+1, seen)
	}

	return out
}

func (r *Redactor) redactMap(v reflect.Value, depth int, seen map[uintptr]struct{}) map[string]any {
	out := make(map[string]any)
	iter := v.MapRange()
	for iter.Next() {
		key := iter.Key()
		val := iter.Value()
		keyStr := ""
		if key.Kind() == reflect.String {
			keyStr = key.String()
		} else {
			keyStr = fmt.Sprint(key.Interface())
		}
		if keyStr != "" && r.isSensitiveKey(keyStr) {
			out[keyStr] = r.Mask
			continue
		}
		out[keyStr] = r.redactValue(val, depth+1, seen)
	}
	return out
}

func (r *Redactor) shouldRedactField(name string, tag reflect.StructTag) bool {
	if r.isSensitiveKey(name) {
		return true
	}
	for _, tagName := range r.Tags {
		tagValue := strings.ToLower(strings.TrimSpace(tag.Get(tagName)))
		if tagValue == "" {
			continue
		}
		if _, ok := r.TagValues[tagValue]; ok {
			return true
		}
	}
	return false
}

func (r *Redactor) isSensitiveKey(name string) bool {
	if name == "" {
		return false
	}
	_, ok := r.Fields[strings.ToLower(name)]
	return ok
}
