package validator

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Rule func(value any) error

// Required ensures the value is non-nil and non-empty (for strings).
func Required() Rule {
	return func(value any) error {
		switch v := value.(type) {
		case nil:
			return errors.New("field is required")
		case string:
			if strings.TrimSpace(v) == "" {
				return errors.New("field is required")
			}
		default:
			// For unsupported types, rely on reflect zero check.
			rv := reflect.ValueOf(value)
			if rv.IsZero() {
				return errors.New("field is required")
			}
		}
		return nil
	}
}

// Email performs a lightweight email format check.
func Email() Rule {
	return func(value any) error {
		if value == nil {
			return errors.New("must be string")
		}

		str, ok := value.(string)
		if !ok {
			return errors.New("must be string")
		}
		str = strings.TrimSpace(str)
		if str == "" || !strings.Contains(str, "@") {
			return errors.New("invalid email format")
		}
		return nil
	}
}

// Min enforces a minimum length for strings using rune count.
func Min(min int) Rule {
	return func(value any) error {
		str, ok := value.(string)
		if ok && utf8.RuneCountInString(str) < min {
			return fmt.Errorf("must be at least %d characters", min)
		}
		return nil
	}
}

// Max enforces a maximum length for strings using rune count.
func Max(max int) Rule {
	return func(value any) error {
		str, ok := value.(string)
		if ok && utf8.RuneCountInString(str) > max {
			return fmt.Errorf("must be at most %d characters", max)
		}
		return nil
	}
}

// Validate applies validation rules defined via struct tags.
// Supported tag syntax: `validate:"required,email,min=3,max=20"`.
func Validate(data any) error {
	if data == nil {
		return errors.New("validator: nil data")
	}

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return errors.New("validator: nil data")
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return errors.New("validator: expected struct")
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}

		rawValue := v.Field(i)
		if rawValue.Kind() == reflect.Pointer {
			if rawValue.IsNil() {
				if !strings.Contains(tag, "required") {
					continue
				}
				// keep nil for required rule to catch
				rawValue = reflect.Zero(rawValue.Type())
			} else {
				rawValue = rawValue.Elem()
			}
		}

		value := rawValue.Interface()
		rules := strings.Split(tag, ",")
		for _, rule := range rules {
			trimmed := strings.TrimSpace(rule)
			if trimmed == "" {
				continue
			}
			if err := applyRule(trimmed, value); err != nil {
				return fmt.Errorf("%s: %w", field.Name, err)
			}
		}
	}

	return nil
}

func applyRule(rule string, value any) error {
	switch {
	case rule == "required":
		return Required()(value)
	case rule == "email":
		return Email()(value)
	case strings.HasPrefix(rule, "min="):
		n, err := strconv.Atoi(strings.TrimPrefix(rule, "min="))
		if err != nil {
			return fmt.Errorf("invalid min value")
		}
		return Min(n)(value)
	case strings.HasPrefix(rule, "max="):
		n, err := strconv.Atoi(strings.TrimPrefix(rule, "max="))
		if err != nil {
			return fmt.Errorf("invalid max value")
		}
		return Max(n)(value)
	default:
		return fmt.Errorf("unknown rule: %s", rule)
	}
}

// BindJSON decodes a JSON request body into v and applies validation tags.
func BindJSON(r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}
	return Validate(v)
}
