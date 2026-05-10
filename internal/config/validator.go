package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

// Validator interface for configuration validation
type Validator interface {
	Validate(value any, key string) error
	Name() string
}

// Required validator ensures a value is not empty.
type Required struct{}

func (r *Required) Validate(value any, key string) error {
	if value == nil {
		return fmt.Errorf("config %s is required", key)
	}

	strValue := toString(value)
	if strings.TrimSpace(strValue) == "" {
		return fmt.Errorf("config %s is required and cannot be empty", key)
	}

	return nil
}

func (r *Required) Name() string {
	return "required"
}

// MinLength validator ensures string length is at least minimum
type MinLength struct {
	Min int
}

func (m *MinLength) Validate(value any, key string) error {
	str := toString(value)
	n := utf8.RuneCountInString(str)
	if n < m.Min {
		return fmt.Errorf("config %s must be at least %d characters, got %d", key, m.Min, n)
	}
	return nil
}

func (m *MinLength) Name() string {
	return "min_length"
}

// MaxLength validator ensures string length is at most maximum
type MaxLength struct {
	Max int
}

func (m *MaxLength) Validate(value any, key string) error {
	str := toString(value)
	if utf8.RuneCountInString(str) > m.Max {
		return fmt.Errorf("config %s must be at most %d characters, got %d", key, m.Max, utf8.RuneCountInString(str))
	}
	return nil
}

func (m *MaxLength) Name() string {
	return "max_length"
}

// Pattern validator ensures string matches regex pattern
type Pattern struct {
	Pattern    string
	once       sync.Once
	regex      *regexp.Regexp
	compileErr error
}

func (p *Pattern) compile() error {
	p.once.Do(func() {
		compiled, err := regexp.Compile(p.Pattern)
		if err != nil {
			p.compileErr = fmt.Errorf("invalid pattern %s: %w", p.Pattern, err)
			return
		}
		p.regex = compiled
	})
	return p.compileErr
}

func (p *Pattern) Validate(value any, key string) error {
	if err := p.compile(); err != nil {
		return err
	}

	str := toString(value)
	if !p.regex.MatchString(str) {
		return fmt.Errorf("config %s must match pattern %s, got %s", key, p.Pattern, str)
	}
	return nil
}

func (p *Pattern) Name() string {
	return "pattern"
}

// URL validator ensures string is a valid URL
type URL struct{}

func (u *URL) Validate(value any, key string) error {
	str := toString(value)
	if str == "" {
		return nil // Empty string is allowed (can use default)
	}
	parsed, err := url.Parse(str)
	if err != nil {
		return fmt.Errorf("config %s must be a valid URL, got %s: %w", key, str, err)
	}
	if parsed.Scheme == "" {
		return fmt.Errorf("config %s must be a valid URL with scheme (e.g., http://, https://), got %s", key, str)
	}
	if parsed.Host == "" {
		return fmt.Errorf("config %s must be a valid URL with host, got %s", key, str)
	}
	return nil
}

func (u *URL) Name() string {
	return "url"
}

// Range validator ensures numeric values are within range
type Range struct {
	Min float64
	Max float64
}

func (r *Range) Validate(value any, key string) error {
	num, err := parseFloatForValidation(value)
	if err != nil {
		return err
	}
	if num < r.Min || num > r.Max {
		return fmt.Errorf("config %s must be between %g and %g, got %g", key, r.Min, r.Max, num)
	}
	return nil
}

func (r *Range) Name() string {
	return "range"
}

// OneOf validator ensures value is one of the allowed values
type OneOf struct {
	Values []string
}

func (o *OneOf) Validate(value any, key string) error {
	str := toString(value)
	for _, allowed := range o.Values {
		if str == allowed {
			return nil
		}
	}
	return fmt.Errorf("config %s must be one of %v, got %s", key, o.Values, str)
}

func (o *OneOf) Name() string {
	return "one_of"
}

// parseFloatForValidation converts any value to float64 for validation purposes.
func parseFloatForValidation(value any) (float64, error) {
	str := toString(value)
	if str == "" {
		return 0, fmt.Errorf("cannot convert empty value to float64")
	}
	return strconv.ParseFloat(str, 64)
}
