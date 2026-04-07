package tenant

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"
)

// ErrInvalidTenantID indicates a tenant ID failed validation.
var ErrInvalidTenantID = errors.New("invalid tenant ID")

// maxTenantIDLength is the maximum allowed length for a tenant ID.
const maxTenantIDLength = 255

// ValidateTenantID checks that a tenant ID is well-formed.
// Valid tenant IDs contain only alphanumeric characters, hyphens,
// underscores, and periods, with a maximum length of 255 bytes.
func ValidateTenantID(id string) error {
	if id == "" {
		return ErrInvalidTenantID
	}
	if len(id) > maxTenantIDLength {
		return fmt.Errorf("%w: exceeds maximum length of %d", ErrInvalidTenantID, maxTenantIDLength)
	}
	if !utf8.ValidString(id) {
		return fmt.Errorf("%w: contains invalid UTF-8", ErrInvalidTenantID)
	}
	for _, c := range id {
		if !isValidTenantIDChar(c) {
			return fmt.Errorf("%w: contains invalid character %q", ErrInvalidTenantID, c)
		}
	}
	return nil
}

func isValidTenantIDChar(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '_' || c == '.'
}

// TenantExtractor extracts tenant ID from a request.
type TenantExtractor func(*http.Request) (string, error)

// fromRequestValue creates a TenantExtractor from a function that reads
// a string value from the request.
func fromRequestValue(extract func(*http.Request) string) TenantExtractor {
	return func(r *http.Request) (string, error) {
		v := extract(r)
		if v == "" {
			return "", ErrTenantNotFound
		}
		return v, nil
	}
}

// FromHeader creates an extractor that reads tenant ID from an HTTP header.
func FromHeader(headerName string) TenantExtractor {
	return fromRequestValue(func(r *http.Request) string {
		return r.Header.Get(headerName)
	})
}

// FromQuery creates an extractor that reads tenant ID from a query parameter.
func FromQuery(paramName string) TenantExtractor {
	return fromRequestValue(func(r *http.Request) string {
		return r.URL.Query().Get(paramName)
	})
}

// FromCookie creates an extractor that reads tenant ID from a cookie.
func FromCookie(cookieName string) TenantExtractor {
	return fromRequestValue(func(r *http.Request) string {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			return ""
		}
		return cookie.Value
	})
}

// FromSubdomain creates an extractor that parses tenant ID from the request subdomain.
// Example: tenant1.example.com → "tenant1".
func FromSubdomain() TenantExtractor {
	return func(r *http.Request) (string, error) {
		host := r.Host
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}
		parts := strings.Split(host, ".")
		if len(parts) < 2 {
			return "", ErrTenantNotFound
		}
		return parts[0], nil
	}
}

// FromContextValue creates an extractor that reads tenant ID from a context value.
// Useful for integrating with authentication middleware that stores parsed claims
// (e.g., JWT) in the request context.
//
// Example:
//
//	tenant.FromContextValue(jwtClaimsKey, func(v any) string {
//		claims, ok := v.(map[string]any)
//		if !ok { return "" }
//		id, _ := claims["tenant_id"].(string)
//		return id
//	})
func FromContextValue(key any, extract func(any) string) TenantExtractor {
	return func(r *http.Request) (string, error) {
		v := r.Context().Value(key)
		if v == nil {
			return "", ErrTenantNotFound
		}
		id := extract(v)
		if id == "" {
			return "", ErrTenantNotFound
		}
		return id, nil
	}
}

// Chain chains multiple extractors, trying each in order until one succeeds.
func Chain(extractors ...TenantExtractor) TenantExtractor {
	return func(r *http.Request) (string, error) {
		for _, extractor := range extractors {
			tenantID, err := extractor(r)
			if err == nil && tenantID != "" {
				return tenantID, nil
			}
		}
		return "", ErrTenantNotFound
	}
}
