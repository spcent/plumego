package observability

import (
	"net/http"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
)

const sensitiveFieldMask = "***"

func RequestIDFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if id := contract.RequestIDFromContext(r.Context()); id != "" {
		return id
	}
	if id := strings.TrimSpace(r.Header.Get(contract.RequestIDHeader)); id != "" {
		return id
	}
	return ""
}

func AttachRequestID(w http.ResponseWriter, r *http.Request, id string, includeInRequest bool) *http.Request {
	if r == nil {
		return r
	}
	ctx := contract.WithRequestID(r.Context(), id)
	if includeInRequest {
		r.Header.Set(contract.RequestIDHeader, id)
	}
	if w != nil {
		w.Header().Set(contract.RequestIDHeader, id)
	}
	return r.WithContext(ctx)
}

func MiddlewareLogFields(r *http.Request, status int, duration time.Duration) map[string]any {
	fields := map[string]any{
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
	fields["request_id"] = RequestIDFromRequest(r)
	return fields
}

func RedactFields(fields map[string]any, extraSensitiveKeys ...string) map[string]any {
	if fields == nil {
		return nil
	}

	sensitiveKeys := map[string]struct{}{
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
		sensitiveKeys[key] = struct{}{}
	}

	out := make(map[string]any, len(fields))
	for k, v := range fields {
		if isSensitiveKey(k, sensitiveKeys) {
			out[k] = sensitiveFieldMask
			continue
		}
		out[k] = redactValue(v, sensitiveKeys)
	}
	return out
}

func redactValue(v any, sensitiveKeys map[string]struct{}) any {
	switch value := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(value))
		for k, nested := range value {
			if isSensitiveKey(k, sensitiveKeys) {
				out[k] = sensitiveFieldMask
				continue
			}
			out[k] = redactValue(nested, sensitiveKeys)
		}
		return out
	case []any:
		items := make([]any, len(value))
		for i, item := range value {
			items[i] = redactValue(item, sensitiveKeys)
		}
		return items
	default:
		return value
	}
}

func isSensitiveKey(key string, sensitiveKeys map[string]struct{}) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	if normalized == "" {
		return false
	}
	for sensitive := range sensitiveKeys {
		if strings.Contains(normalized, sensitive) {
			return true
		}
	}
	return false
}
