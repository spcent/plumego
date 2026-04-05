package contract

import (
	"net/http"
	"strings"
	"time"

	"github.com/spcent/plumego/log"
)

const (
	// RequestIDHeader is the canonical request id header.
	RequestIDHeader = "X-Request-ID"
	// FallbackRequestIDHeader is the legacy request id header.
	FallbackRequestIDHeader = "X-Trace-ID"
)

// ObservabilityPolicy defines canonical middleware observability behavior.
type ObservabilityPolicy struct {
	mask          string
	sensitiveKeys map[string]struct{}
}

// DefaultObservabilityPolicy is the shared observability policy.
var DefaultObservabilityPolicy = NewObservabilityPolicy()

// NewObservabilityPolicy creates a policy with safe defaults and optional
// application-specific sensitive key patterns.
func NewObservabilityPolicy(extraSensitiveKeys ...string) ObservabilityPolicy {
	keys := map[string]struct{}{
		"token":     {},
		"secret":    {},
		"signature": {},
		"password":  {},
	}
	for _, key := range extraSensitiveKeys {
		key = strings.ToLower(strings.TrimSpace(key))
		if key == "" {
			continue
		}
		keys[key] = struct{}{}
	}

	return ObservabilityPolicy{
		mask:          "***",
		sensitiveKeys: keys,
	}
}

// RequestIDFromRequest resolves request id from context, canonical header, or fallback header.
func (p ObservabilityPolicy) RequestIDFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if id := TraceIDFromContext(r.Context()); id != "" {
		return id
	}
	if id := strings.TrimSpace(r.Header.Get(RequestIDHeader)); id != "" {
		return id
	}
	if id := strings.TrimSpace(r.Header.Get(FallbackRequestIDHeader)); id != "" {
		return id
	}
	return ""
}

// AttachRequestID writes request id to request context and response header.
func (p ObservabilityPolicy) AttachRequestID(w http.ResponseWriter, r *http.Request, id string, includeInRequest bool) *http.Request {
	if r == nil {
		return r
	}
	ctx := WithTraceIDString(r.Context(), id)
	if includeInRequest {
		r.Header.Set(RequestIDHeader, id)
	}
	if w != nil {
		w.Header().Set(RequestIDHeader, id)
	}
	return r.WithContext(ctx)
}

// MiddlewareLogFields returns the canonical structured log fields for middleware logs.
func (p ObservabilityPolicy) MiddlewareLogFields(r *http.Request, status int, duration time.Duration) log.Fields {
	fields := log.Fields{
		"method":     "",
		"path":       "",
		"status":     status,
		"duration":   duration.String(),
		"request_id": "",
	}
	if r == nil {
		return fields
	}
	fields["method"] = r.Method
	if r.URL != nil {
		fields["path"] = r.URL.Path
	}
	fields["request_id"] = p.RequestIDFromRequest(r)
	return fields
}

// SensitiveKeys returns sensitive key patterns used for redaction.
func (p ObservabilityPolicy) SensitiveKeys() []string {
	keys := make([]string, 0, len(p.sensitiveKeys))
	for key := range p.sensitiveKeys {
		keys = append(keys, key)
	}
	return keys
}

// RedactFields returns a deep copy with sensitive values masked.
func (p ObservabilityPolicy) RedactFields(fields map[string]any) map[string]any {
	if fields == nil {
		return nil
	}
	out := make(map[string]any, len(fields))
	for k, v := range fields {
		if p.isSensitiveKey(k) {
			out[k] = p.mask
			continue
		}
		out[k] = p.redactValue(v)
	}
	return out
}

func (p ObservabilityPolicy) redactValue(v any) any {
	switch value := v.(type) {
	case map[string]any:
		return p.RedactFields(value)
	case []any:
		items := make([]any, len(value))
		for i, item := range value {
			items[i] = p.redactValue(item)
		}
		return items
	default:
		return value
	}
}

func (p ObservabilityPolicy) isSensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	if normalized == "" {
		return false
	}
	for sensitive := range p.sensitiveKeys {
		if strings.Contains(normalized, sensitive) {
			return true
		}
	}
	return false
}
