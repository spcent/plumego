package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorBuilder(t *testing.T) {
	builder := NewErrorBuilder()

	err := builder.
		Status(http.StatusBadRequest).
		Category(CategoryValidation).
		Type(TypeValidation).
		Code(CodeValidationError).
		Message("test error message").
		Detail("field", "email").
		Detail("value", "invalid").
		Build()

	if err.Status != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, err.Status)
	}

	if err.Category != CategoryValidation {
		t.Fatalf("expected category %s, got %s", CategoryValidation, err.Category)
	}

	if err.Code != CodeValidationError {
		t.Fatalf("expected code %s, got %s", CodeValidationError, err.Code)
	}

	if err.Message != "test error message" {
		t.Fatalf("expected message %s, got %s", "test error message", err.Message)
	}

	if err.Details["field"] != "email" {
		t.Fatalf("expected field detail, got %v", err.Details["field"])
	}
}

func TestBuilderTypeOverwritesPriorFields(t *testing.T) {
	got := NewErrorBuilder().
		Status(999).
		Code(CodeInternalError).
		Category(CategoryServer).
		Type(TypeNotFound).
		Build()

	if got.Status != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, got.Status)
	}
	if got.Code != CodeResourceNotFound {
		t.Fatalf("expected code %q, got %q", CodeResourceNotFound, got.Code)
	}
	if got.Category != CategoryClient {
		t.Fatalf("expected category %q, got %q", CategoryClient, got.Category)
	}
}

func TestBuilderTypeStatusAndCategoryRemainCanonical(t *testing.T) {
	got := NewErrorBuilder().
		Type(TypeNotFound).
		Status(http.StatusUnprocessableEntity).
		Category(CategoryValidation).
		Build()

	if got.Status != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, got.Status)
	}
	if got.Category != CategoryClient {
		t.Fatalf("expected category %q, got %q", CategoryClient, got.Category)
	}
}

func TestErrorTypeMetaIsNameable(t *testing.T) {
	var meta ErrorTypeMeta = TypeNotFound.Meta()
	if meta.Status != http.StatusNotFound {
		t.Fatalf("status=%d, want %d", meta.Status, http.StatusNotFound)
	}
	if meta.Code != CodeResourceNotFound {
		t.Fatalf("code=%q, want %q", meta.Code, CodeResourceNotFound)
	}
	if meta.Category != CategoryClient {
		t.Fatalf("category=%q, want %q", meta.Category, CategoryClient)
	}
}

func TestGatewayTimeoutErrorTypeMeta(t *testing.T) {
	meta := TypeGatewayTimeout.Meta()
	if meta.Status != http.StatusGatewayTimeout {
		t.Fatalf("status=%d, want %d", meta.Status, http.StatusGatewayTimeout)
	}
	if meta.Code != CodeGatewayTimeout {
		t.Fatalf("code=%q, want %q", meta.Code, CodeGatewayTimeout)
	}
	if meta.Category != CategoryTimeout {
		t.Fatalf("category=%q, want %q", meta.Category, CategoryTimeout)
	}
}

func TestErrorTypeTaxonomyMatrix(t *testing.T) {
	cases := []struct {
		errorType ErrorType
		category  ErrorCategory
		code      string
		status    int
	}{
		{TypeValidation, CategoryValidation, CodeValidationError, http.StatusBadRequest},
		{TypeRequired, CategoryValidation, CodeRequired, http.StatusBadRequest},
		{TypeInvalidFormat, CategoryValidation, CodeInvalidFormat, http.StatusBadRequest},
		{TypeOutOfRange, CategoryValidation, CodeOutOfRange, http.StatusBadRequest},
		{TypeDuplicate, CategoryValidation, CodeDuplicate, http.StatusBadRequest},
		{TypeUnauthorized, CategoryAuth, CodeUnauthorized, http.StatusUnauthorized},
		{TypeForbidden, CategoryAuth, CodeForbidden, http.StatusForbidden},
		{TypeInvalidToken, CategoryAuth, CodeInvalidToken, http.StatusUnauthorized},
		{TypeExpiredToken, CategoryAuth, CodeExpiredToken, http.StatusUnauthorized},
		{TypeNotFound, CategoryClient, CodeResourceNotFound, http.StatusNotFound},
		{TypeConflict, CategoryClient, CodeConflict, http.StatusConflict},
		{TypeAlreadyExists, CategoryClient, CodeAlreadyExists, http.StatusConflict},
		{TypeGone, CategoryClient, CodeGone, http.StatusGone},
		{TypeInternal, CategoryServer, CodeInternalError, http.StatusInternalServerError},
		{TypeUnavailable, CategoryServer, CodeUnavailable, http.StatusServiceUnavailable},
		{TypeTimeout, CategoryTimeout, CodeTimeout, http.StatusRequestTimeout},
		{TypeRateLimited, CategoryRateLimit, CodeRateLimited, http.StatusTooManyRequests},
		{TypeMaintenance, CategoryServer, CodeMaintenance, http.StatusServiceUnavailable},
		{TypeMethodNotAllowed, CategoryClient, CodeMethodNotAllowed, http.StatusMethodNotAllowed},
		{TypeNotImplemented, CategoryServer, CodeNotImplemented, http.StatusNotImplemented},
		{TypeBadGateway, CategoryServer, CodeBadGateway, http.StatusBadGateway},
		{TypeGatewayTimeout, CategoryTimeout, CodeGatewayTimeout, http.StatusGatewayTimeout},
	}

	if len(cases) != len(errorTypeLookup) {
		t.Fatalf("taxonomy cases=%d, lookup entries=%d", len(cases), len(errorTypeLookup))
	}

	for _, tc := range cases {
		t.Run(string(tc.errorType), func(t *testing.T) {
			meta := tc.errorType.Meta()
			if meta.Category != tc.category || meta.Code != tc.code || meta.Status != tc.status {
				t.Fatalf("Meta() = {category:%q code:%q status:%d}, want {category:%q code:%q status:%d}",
					meta.Category, meta.Code, meta.Status, tc.category, tc.code, tc.status)
			}

			got := NewErrorBuilder().Type(tc.errorType).Build()
			if got.Category != tc.category || got.Code != tc.code || got.Status != tc.status {
				t.Fatalf("builder = {category:%q code:%q status:%d}, want {category:%q code:%q status:%d}",
					got.Category, got.Code, got.Status, tc.category, tc.code, tc.status)
			}
		})
	}
}

func TestUnknownErrorTypeMetaFailsClosed(t *testing.T) {
	meta := ErrorType("extension_unknown").Meta()
	if meta.Category != CategoryServer {
		t.Fatalf("category=%q, want %q", meta.Category, CategoryServer)
	}
	if meta.Code != CodeInternalError {
		t.Fatalf("code=%q, want %q", meta.Code, CodeInternalError)
	}
	if meta.Status != http.StatusInternalServerError {
		t.Fatalf("status=%d, want %d", meta.Status, http.StatusInternalServerError)
	}
}

func TestNormalizeTypedAPIErrorKeepsCanonicalStatusAndCategory(t *testing.T) {
	got := normalizeAPIError(APIError{
		Status:   http.StatusConflict,
		Code:     "CUSTOM_NOT_FOUND",
		Message:  "custom not found",
		Category: CategoryServer,
		Type:     TypeNotFound,
	})

	if got.Status != http.StatusNotFound {
		t.Fatalf("status=%d, want %d", got.Status, http.StatusNotFound)
	}
	if got.Category != CategoryClient {
		t.Fatalf("category=%s, want %s", got.Category, CategoryClient)
	}
	if got.Code != "CUSTOM_NOT_FOUND" {
		t.Fatalf("code=%s, want %s", got.Code, "CUSTOM_NOT_FOUND")
	}
	if got.Type != TypeNotFound {
		t.Fatalf("type=%s, want %s", got.Type, TypeNotFound)
	}
}

func TestErrorBuilderChaining(t *testing.T) {
	err := NewErrorBuilder().
		Status(http.StatusNotFound).
		Category(CategoryClient).
		Type(TypeNotFound).
		Code(CodeResourceNotFound).
		Message("resource not found").
		Detail("resource", "user").
		Detail("id", "123").
		Build()

	if err.Status != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, err.Status)
	}

	if err.Details["resource"] != "user" || err.Details["id"] != "123" {
		t.Fatalf("expected details to be set, got %v", err.Details)
	}
}

func TestCommonErrorBuilders(t *testing.T) {
	// Test validation error via builder
	valErr := NewErrorBuilder().
		Status(http.StatusBadRequest).
		Category(CategoryValidation).
		Type(TypeValidation).
		Code(CodeValidationError).
		Message("validation failed for field 'email': invalid format").
		Detail("field", "email").
		Detail("validation_message", "invalid format").
		Build()
	if valErr.Category != CategoryValidation {
		t.Fatalf("expected validation category")
	}
	if valErr.Details["field"] != "email" {
		t.Fatalf("expected field detail")
	}

	// Test not found error via builder
	notFoundErr := NewErrorBuilder().
		Status(http.StatusNotFound).
		Category(CategoryClient).
		Type(TypeNotFound).
		Code(CodeResourceNotFound).
		Message("resource 'user' not found").
		Detail("resource", "user").
		Build()
	if notFoundErr.Category != CategoryClient {
		t.Fatalf("expected client category for not found")
	}
	if notFoundErr.Details["resource"] != "user" {
		t.Fatalf("expected resource detail")
	}

	// Test unauthorized error via builder
	authErr := NewErrorBuilder().
		Status(http.StatusUnauthorized).
		Category(CategoryAuth).
		Type(TypeUnauthorized).
		Code(CodeUnauthorized).
		Message("invalid token").
		Build()
	if authErr.Category != CategoryAuth {
		t.Fatalf("expected authentication category")
	}

	// Test timeout error via builder
	timeoutErr := NewErrorBuilder().
		Status(http.StatusRequestTimeout).
		Category(CategoryTimeout).
		Type(TypeTimeout).
		Code(CodeTimeout).
		Message("database timeout").
		Build()
	if timeoutErr.Category != CategoryTimeout {
		t.Fatalf("expected timeout category")
	}

	// Test rate limit error via builder
	rateLimitErr := NewErrorBuilder().
		Status(http.StatusTooManyRequests).
		Category(CategoryRateLimit).
		Type(TypeRateLimited).
		Code(CodeRateLimited).
		Message("too many requests").
		Build()
	if rateLimitErr.Category != CategoryRateLimit {
		t.Fatalf("expected rate limit category")
	}
}

func TestErrorValidation(t *testing.T) {
	// Valid error
	validErr := APIError{
		Status:   http.StatusBadRequest,
		Code:     CodeValidationError,
		Message:  "validation failed",
		Category: CategoryValidation,
	}

	validationErrors := validateAPIError(validErr)
	if len(validationErrors) > 0 {
		t.Fatalf("expected no validation errors, got %v", validationErrors)
	}

	// Invalid errors
	invalidCases := []APIError{
		{
			Status:   999, // Invalid status
			Code:     CodeInternalError,
			Message:  "message",
			Category: CategoryClient,
		},
		{
			Status:   http.StatusOK,
			Code:     CodeInternalError,
			Message:  "message",
			Category: CategoryClient,
		},
		{
			Status:   http.StatusFound,
			Code:     CodeInternalError,
			Message:  "message",
			Category: CategoryClient,
		},
		{
			Status:   http.StatusBadRequest,
			Code:     "", // Empty code
			Message:  "message",
			Category: CategoryClient,
		},
		{
			Status:   http.StatusBadRequest,
			Code:     CodeInternalError,
			Message:  "", // Empty message
			Category: CategoryClient,
		},
		{
			Status:   http.StatusBadRequest,
			Code:     CodeInternalError,
			Message:  "message",
			Category: "", // Empty category
		},
	}

	for i, invalidErr := range invalidCases {
		errs := validateAPIError(invalidErr)
		if len(errs) == 0 {
			t.Fatalf("case %d: expected validation errors", i)
		}
	}
}

func TestHTTPStatusFromCategory(t *testing.T) {
	tests := []struct {
		category ErrorCategory
		expected int
	}{
		{CategoryClient, http.StatusBadRequest},
		{CategoryValidation, http.StatusBadRequest},
		{CategoryAuth, http.StatusUnauthorized},
		{CategoryRateLimit, http.StatusTooManyRequests},
		{CategoryServer, http.StatusInternalServerError},
		{CategoryTimeout, http.StatusRequestTimeout},
		{"unknown", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		if got := HTTPStatusFromCategory(tt.category); got != tt.expected {
			t.Fatalf("category %s: expected status %d, got %d", tt.category, tt.expected, got)
		}
	}
}

func TestErrorResponseWriting(t *testing.T) {
	// Test WriteError function
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	err := APIError{
		Status:   http.StatusBadRequest,
		Code:     CodeValidationError,
		Message:  "validation failed",
		Category: CategoryValidation,
		Details:  map[string]any{"field": "email"},
	}

	WriteError(recorder, req, err)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	if ct := recorder.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected content type application/json, got %s", ct)
	}

	var response ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error.Code != CodeValidationError {
		t.Fatalf("expected code in response")
	}
}

func TestErrorResponseWithRequestID(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := WithRequestID(req.Context(), "req-123")
	req = req.WithContext(ctx)

	err := APIError{
		Status:   http.StatusInternalServerError,
		Code:     CodeInternalError,
		Message:  "internal server error",
		Category: CategoryServer,
	}

	WriteError(recorder, req, err)

	var response ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.RequestID != "req-123" {
		t.Fatalf("expected request id in response")
	}
}

func TestWriteErrorPreservesRequestID(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := WithRequestID(req.Context(), "context-req-id")
	req = req.WithContext(ctx)

	err := APIError{
		Status:    http.StatusBadRequest,
		Code:      CodeValidationError,
		Message:   "validation failed",
		Category:  CategoryValidation,
		RequestID: "explicit-req-id",
	}

	WriteError(recorder, req, err)

	var response ErrorResponse
	if decodeErr := json.NewDecoder(recorder.Body).Decode(&response); decodeErr != nil {
		t.Fatalf("failed to decode response: %v", decodeErr)
	}

	if response.RequestID != "explicit-req-id" {
		t.Fatalf("expected explicit request id to be preserved")
	}
}

func TestWriteErrorNormalizesExplicitRequestID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{name: "trim", id: " explicit-req-id ", want: "explicit-req-id"},
		{name: "unsafe", id: "bad\nrequest", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			WriteError(recorder, req, APIError{
				Status:    http.StatusBadRequest,
				Code:      CodeValidationError,
				Message:   "validation failed",
				Category:  CategoryValidation,
				RequestID: tt.id,
			})

			var response ErrorResponse
			if decodeErr := json.NewDecoder(recorder.Body).Decode(&response); decodeErr != nil {
				t.Fatalf("failed to decode response: %v", decodeErr)
			}
			if response.RequestID != tt.want {
				t.Fatalf("expected request id %q, got %q", tt.want, response.RequestID)
			}
		})
	}
}

func TestErrorBuilderRequestIDNormalizesSafety(t *testing.T) {
	got := NewErrorBuilder().RequestID(" req-123 ").Build()
	if got.RequestID != "req-123" {
		t.Fatalf("expected trimmed request id, got %q", got.RequestID)
	}

	got = NewErrorBuilder().RequestID("bad\trequest").Build()
	if got.RequestID != "" {
		t.Fatalf("expected unsafe request id to be dropped, got %q", got.RequestID)
	}
}

func TestWriteErrorDefaults(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Test with minimal error (no status, code, or category)
	err := APIError{Message: "test message"}

	WriteError(recorder, req, err)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected default status to be internal server error")
	}

	var response ErrorResponse
	json.NewDecoder(recorder.Body).Decode(&response)

	if response.Error.Code != CodeInternalError {
		t.Fatalf("expected default code %q, got %q", CodeInternalError, response.Error.Code)
	}

	if response.Error.Category != CategoryServer {
		t.Fatalf("expected default category to be server")
	}
}

func TestWriteErrorDefaultsMessage(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	WriteError(recorder, req, APIError{})

	var response ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.Error.Message != http.StatusText(http.StatusInternalServerError) {
		t.Fatalf("expected default message %q, got %q", http.StatusText(http.StatusInternalServerError), response.Error.Message)
	}
}

func TestWriteErrorDefaultCodeUsesCanonicalMachineCode(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   string
	}{
		{name: "bad request", status: http.StatusBadRequest, want: CodeBadRequest},
		{name: "unprocessable", status: http.StatusUnprocessableEntity, want: CodeInvalidRequest},
		{name: "unknown client", status: http.StatusTeapot, want: CodeInvalidRequest},
		{name: "gateway timeout", status: http.StatusGatewayTimeout, want: CodeGatewayTimeout},
		{name: "service unavailable", status: http.StatusServiceUnavailable, want: CodeUnavailable},
		{name: "unknown server", status: http.StatusLoopDetected, want: CodeInternalError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			if err := WriteError(recorder, req, APIError{
				Status:   tt.status,
				Message:  "error",
				Category: CategoryForStatus(tt.status),
			}); err != nil {
				t.Fatalf("unexpected write error: %v", err)
			}

			var response ErrorResponse
			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if response.Error.Code != tt.want {
				t.Fatalf("expected code %q, got %q", tt.want, response.Error.Code)
			}
		})
	}
}

func TestWriteErrorPreservesExplicitCode(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if err := WriteError(recorder, req, APIError{
		Status:   http.StatusBadRequest,
		Code:     "CUSTOM_STABLE_CODE",
		Message:  "bad request",
		Category: CategoryClient,
	}); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	var response ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.Error.Code != "CUSTOM_STABLE_CODE" {
		t.Fatalf("expected explicit code to be preserved, got %q", response.Error.Code)
	}
}

func TestErrorBuilderBuildDefaultsMessage(t *testing.T) {
	err := NewErrorBuilder().Build()
	if err.Message != http.StatusText(http.StatusInternalServerError) {
		t.Fatalf("expected default message %q, got %q", http.StatusText(http.StatusInternalServerError), err.Message)
	}
}

func TestErrorBuilderWithSeverityAndType(t *testing.T) {
	err := NewErrorBuilder().
		Status(http.StatusBadRequest).
		Category(CategoryValidation).
		Type(TypeValidation).
		Severity(SeverityWarning).
		Code(CodeValidationError).
		Message("validation warning").
		Build()

	if err.Severity != SeverityWarning {
		t.Fatalf("expected severity to be set on APIError")
	}

	if err.Type != TypeValidation {
		t.Fatalf("expected type to be set on APIError")
	}
}

func TestErrorBuilderDropsInvalidTypeAndSeverity(t *testing.T) {
	err := NewErrorBuilder().
		Status(http.StatusBadRequest).
		Code(CodeBadRequest).
		Message("bad request").
		Type(ErrorType("unknown_type")).
		Severity(ErrorSeverity("loud")).
		Build()

	if err.Type != "" {
		t.Fatalf("expected invalid type to be dropped, got %q", err.Type)
	}
	if err.Severity != "" {
		t.Fatalf("expected invalid severity to be dropped, got %q", err.Severity)
	}
}

func TestWriteErrorDropsInvalidTypeAndSeverity(t *testing.T) {
	recorder := httptest.NewRecorder()

	if err := WriteError(recorder, nil, APIError{
		Status:   http.StatusBadRequest,
		Code:     CodeBadRequest,
		Message:  "bad request",
		Category: CategoryClient,
		Type:     ErrorType("unknown_type"),
		Severity: ErrorSeverity("loud"),
	}); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	var response ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.Error.Type != "" {
		t.Fatalf("expected invalid type to be omitted, got %q", response.Error.Type)
	}
	if response.Error.Severity != "" {
		t.Fatalf("expected invalid severity to be omitted, got %q", response.Error.Severity)
	}
}

func TestErrorBuilderDetails(t *testing.T) {
	builder := NewErrorBuilder()
	details := map[string]any{
		"field":  "email",
		"value":  "invalid",
		"reason": "format",
	}

	err := builder.
		Status(http.StatusBadRequest).
		Category(CategoryValidation).
		Code(CodeValidationError).
		Message("validation failed").
		Details(details).
		Build()

	if len(err.Details) != 3 {
		t.Fatalf("expected 3 details, got %d", len(err.Details))
	}

	if err.Details["field"] != "email" || err.Details["value"] != "invalid" {
		t.Fatalf("expected details to be copied correctly")
	}
}

func TestErrorBuilderDetailsAreIsolatedAfterBuild(t *testing.T) {
	builder := NewErrorBuilder().
		Code(CodeValidationError).
		Message("validation failed").
		Detail("field", "email")

	first := builder.Build()
	builder.Detail("field", "name").Detail("other", "value")
	second := builder.Build()

	if first.Details["field"] != "email" {
		t.Fatalf("expected first build details to stay isolated, got %v", first.Details["field"])
	}
	if _, ok := first.Details["other"]; ok {
		t.Fatalf("expected first build not to observe later detail, got %+v", first.Details)
	}
	if second.Details["field"] != "name" || second.Details["other"] != "value" {
		t.Fatalf("expected second build to include current details, got %+v", second.Details)
	}
}

func TestErrorBuilderDetailsDeepCloneJSONLikeValues(t *testing.T) {
	fields := []any{
		map[string]any{"field": "email", "codes": []string{"required"}},
	}
	details := map[string]any{
		"fields": fields,
		"labels": []string{"alpha"},
	}

	got := NewErrorBuilder().
		Status(http.StatusBadRequest).
		Code(CodeValidationError).
		Message("validation failed").
		Details(details).
		Build()

	fields[0].(map[string]any)["field"] = "mutated"
	fields[0].(map[string]any)["codes"].([]string)[0] = "mutated"
	details["labels"].([]string)[0] = "mutated"

	gotFields := got.Details["fields"].([]any)
	gotField := gotFields[0].(map[string]any)
	if gotField["field"] != "email" {
		t.Fatalf("expected nested map detail to be isolated, got %+v", gotField)
	}
	if gotField["codes"].([]string)[0] != "required" {
		t.Fatalf("expected nested slice detail to be isolated, got %+v", gotField["codes"])
	}
	if got.Details["labels"].([]string)[0] != "alpha" {
		t.Fatalf("expected top-level slice detail to be isolated, got %+v", got.Details["labels"])
	}
}

func TestWriteErrorDeepClonesDetailsBeforeEncoding(t *testing.T) {
	details := map[string]any{
		"fields": []any{map[string]any{"field": "email"}},
	}
	apiErr := APIError{
		Status:   http.StatusBadRequest,
		Code:     CodeValidationError,
		Message:  "validation failed",
		Category: CategoryValidation,
		Details:  details,
	}

	rec := httptest.NewRecorder()
	if err := WriteError(rec, nil, apiErr); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}
	details["fields"].([]any)[0].(map[string]any)["field"] = "mutated"

	var resp ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	fields, ok := resp.Error.Details["fields"].([]any)
	if !ok || len(fields) != 1 {
		t.Fatalf("expected encoded fields detail, got %+v", resp.Error.Details)
	}
	field, ok := fields[0].(map[string]any)
	if !ok || field["field"] != "email" {
		t.Fatalf("expected encoded detail to preserve original value, got %+v", fields[0])
	}
}

func TestZeroValueErrorBuilderDetailDoesNotPanic(t *testing.T) {
	var builder ErrorBuilder

	got := builder.
		Status(http.StatusBadRequest).
		Code(CodeBadRequest).
		Message("bad request").
		Detail("field", "name").
		Build()

	if got.Details["field"] != "name" {
		t.Fatalf("expected zero-value builder detail to be set, got %+v", got.Details)
	}
}

func TestWriteErrorZeroValueDefaults(t *testing.T) {
	w := httptest.NewRecorder()
	_ = WriteError(w, nil, APIError{})

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestErrorBuilderIgnoresEmptyDetailKeys(t *testing.T) {
	got := NewErrorBuilder().
		Status(http.StatusBadRequest).
		Code(CodeBadRequest).
		Message("bad request").
		Detail("", "ignored").
		Detail("field", "name").
		Details(map[string]any{
			"":       "also ignored",
			"reason": "missing",
		}).
		Build()

	if _, ok := got.Details[""]; ok {
		t.Fatalf("expected empty detail key to be omitted, got %+v", got.Details)
	}
	if got.Details["field"] != "name" || got.Details["reason"] != "missing" {
		t.Fatalf("expected non-empty details to remain, got %+v", got.Details)
	}
}

func TestZeroValueErrorBuilderIgnoresEmptyDetailKey(t *testing.T) {
	var builder ErrorBuilder

	got := builder.
		Status(http.StatusBadRequest).
		Code(CodeBadRequest).
		Message("bad request").
		Detail("", "ignored").
		Build()

	if len(got.Details) != 0 {
		t.Fatalf("expected no details for empty key, got %+v", got.Details)
	}
}

func TestWriteErrorNormalizesInvalidStatus(t *testing.T) {
	w := httptest.NewRecorder()

	if err := WriteError(w, nil, APIError{
		Status:   700,
		Message:  "invalid status",
		Category: CategoryClient,
	}); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected invalid status to normalize to 500, got %d", w.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Error.Category != CategoryServer {
		t.Fatalf("expected category %q, got %q", CategoryServer, response.Error.Category)
	}
	if response.Error.Code != CodeInternalError {
		t.Fatalf("expected code %q, got %q", CodeInternalError, response.Error.Code)
	}
}

func TestWriteErrorNormalizesNonErrorStatus(t *testing.T) {
	w := httptest.NewRecorder()

	if err := WriteError(w, nil, APIError{
		Status:   http.StatusOK,
		Code:     CodeBadRequest,
		Message:  "bad request",
		Category: CategoryClient,
	}); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected non-error status to normalize to 500, got %d", w.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Error.Category != CategoryServer {
		t.Fatalf("expected category %q, got %q", CategoryServer, response.Error.Category)
	}
	if response.Error.Code != CodeBadRequest {
		t.Fatalf("expected explicit code to be preserved, got %q", response.Error.Code)
	}
}

func TestWriteErrorEncodingFailureDoesNotCommitHeaders(t *testing.T) {
	recorder := httptest.NewRecorder()

	err := WriteError(recorder, nil, NewErrorBuilder().
		Status(http.StatusBadRequest).
		Category(CategoryClient).
		Code(CodeInternalError).
		Message("bad detail").
		Detail("bad", make(chan int)).
		Build())
	if err == nil {
		t.Fatal("expected encoding error, got nil")
	}

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected headers to remain uncommitted, got status %d", recorder.Code)
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("expected no body on encoding failure, got %q", recorder.Body.String())
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "" {
		t.Fatalf("expected no content type on encoding failure, got %q", contentType)
	}
}

func TestWriteErrorIgnoresEmptyDetailKeys(t *testing.T) {
	w := httptest.NewRecorder()

	if err := WriteError(w, nil, APIError{
		Status:   http.StatusBadRequest,
		Code:     CodeBadRequest,
		Message:  "bad request",
		Category: CategoryClient,
		Details: map[string]any{
			"":      "ignored",
			"field": "name",
		},
	}); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := response.Error.Details[""]; ok {
		t.Fatalf("expected empty detail key to be omitted, got %+v", response.Error.Details)
	}
	if response.Error.Details["field"] != "name" {
		t.Fatalf("expected non-empty detail to remain, got %+v", response.Error.Details)
	}
}

func TestErrorBuilderBuildFillsDefaults(t *testing.T) {
	// A builder with no explicit Status/Code/Category should produce a fully-
	// populated APIError after Build().
	got := NewErrorBuilder().Message("something went wrong").Build()

	if got.Status == 0 {
		t.Error("Build() must set a non-zero Status")
	}
	if got.Code == "" {
		t.Error("Build() must set a non-empty Code")
	}
	if got.Category == "" {
		t.Error("Build() must set a non-empty Category")
	}

	w := httptest.NewRecorder()
	_ = WriteError(w, nil, got)
	if w.Code != got.Status {
		t.Fatalf("expected status %d, got %d", got.Status, w.Code)
	}
}

func TestErrorBuilderStatusOnlyDerivesCategory(t *testing.T) {
	got := NewErrorBuilder().
		Status(http.StatusBadRequest).
		Code(CodeBadRequest).
		Message("bad request").
		Build()

	if got.Category != CategoryClient {
		t.Fatalf("expected category %q, got %q", CategoryClient, got.Category)
	}
}

func TestErrorBuilderNormalizesInvalidStatus(t *testing.T) {
	got := NewErrorBuilder().
		Status(42).
		Message("invalid status").
		Build()

	if got.Status != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", got.Status)
	}
	if got.Category != CategoryServer {
		t.Fatalf("expected category %q, got %q", CategoryServer, got.Category)
	}
	if got.Code != CodeInternalError {
		t.Fatalf("expected code %q, got %q", CodeInternalError, got.Code)
	}
}

func TestErrorBuilderNormalizesNonErrorStatus(t *testing.T) {
	got := NewErrorBuilder().
		Status(http.StatusFound).
		Code(CodeBadRequest).
		Message("redirect is not an error status").
		Build()

	if got.Status != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", got.Status)
	}
	if got.Category != CategoryServer {
		t.Fatalf("expected category %q, got %q", CategoryServer, got.Category)
	}
	if got.Code != CodeBadRequest {
		t.Fatalf("expected explicit code to be preserved, got %q", got.Code)
	}
}

func TestWriteJSONNormalizesInvalidStatus(t *testing.T) {
	w := httptest.NewRecorder()

	if err := WriteJSON(w, 42, map[string]string{"ok": "true"}); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
