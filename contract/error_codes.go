package contract

import "fmt"

// ErrorCode represents a standardized error code for the application.
// Each error code follows the pattern: CATEGORY_XXX where XXX is a unique number.
type ErrorCode string

const (
	// General Errors
	ErrInternal     ErrorCode = "GENERAL_001" // Internal server error
	ErrInvalidInput ErrorCode = "GENERAL_002" // Invalid input parameters
	ErrNotFound     ErrorCode = "GENERAL_003" // Resource not found
	ErrUnauthorized ErrorCode = "GENERAL_004" // Authentication required
	ErrForbidden    ErrorCode = "GENERAL_005" // Insufficient permissions
	ErrConflict     ErrorCode = "GENERAL_006" // Resource conflict
	ErrValidation   ErrorCode = "GENERAL_007" // Validation failed

	// Dependency Injection Errors
	ErrDICircularDependency ErrorCode = "DI_001" // Circular dependency detected
	ErrDINotFound           ErrorCode = "DI_002" // Service not found in container
	ErrDIInvalidType        ErrorCode = "DI_003" // Invalid type for dependency
	ErrDIMultipleMatches    ErrorCode = "DI_004" // Multiple services match interface
	ErrDIInjectionFailed    ErrorCode = "DI_005" // Dependency injection failed

	// Router Errors
	ErrRouterInvalidPath      ErrorCode = "ROUTER_001" // Invalid route path
	ErrRouterDuplicateRoute   ErrorCode = "ROUTER_002" // Duplicate route registration
	ErrRouterParamMismatch    ErrorCode = "ROUTER_003" // Path parameter mismatch
	ErrRouterMethodNotAllowed ErrorCode = "ROUTER_004" // HTTP method not allowed

	// Middleware Errors
	ErrMiddlewareTimeout      ErrorCode = "MIDDLEWARE_001" // Request timeout
	ErrMiddlewareRateLimit    ErrorCode = "MIDDLEWARE_002" // Rate limit exceeded
	ErrMiddlewareBodyTooLarge ErrorCode = "MIDDLEWARE_003" // Request body too large
	ErrMiddlewareAuthFailed   ErrorCode = "MIDDLEWARE_004" // Authentication failed
	ErrMiddlewareCORS         ErrorCode = "MIDDLEWARE_005" // CORS violation

	// Configuration Errors
	ErrConfigInvalidSource    ErrorCode = "CONFIG_001" // Invalid configuration source
	ErrConfigValidationFailed ErrorCode = "CONFIG_002" // Configuration validation failed
	ErrConfigParseError       ErrorCode = "CONFIG_003" // Configuration parsing error
	ErrConfigMissingRequired  ErrorCode = "CONFIG_004" // Required configuration missing

	// WebSocket Errors
	ErrWSInvalidHandshake ErrorCode = "WS_001" // Invalid WebSocket handshake
	ErrWSAuthFailed       ErrorCode = "WS_002" // WebSocket authentication failed
	ErrWSRoomFull         ErrorCode = "WS_003" // WebSocket room at capacity
	ErrWSHubFull          ErrorCode = "WS_004" // WebSocket hub at capacity
	ErrWSInvalidMessage   ErrorCode = "WS_005" // Invalid WebSocket message

	// Webhook Errors
	ErrWebhookInvalidSignature ErrorCode = "WEBHOOK_001" // Invalid webhook signature
	ErrWebhookDuplicate        ErrorCode = "WEBHOOK_002" // Duplicate webhook event
	ErrWebhookBodyTooLarge     ErrorCode = "WEBHOOK_003" // Webhook body too large
	ErrWebhookProcessingFailed ErrorCode = "WEBHOOK_004" // Webhook processing failed
	ErrWebhookInvalidPayload   ErrorCode = "WEBHOOK_005" // Invalid webhook payload

	// PubSub Errors
	ErrPubSubTopicNotFound   ErrorCode = "PUBSUB_001" // Topic not found
	ErrPubSubPublishFailed   ErrorCode = "PUBSUB_002" // Message publish failed
	ErrPubSubSubscribeFailed ErrorCode = "PUBSUB_003" // Subscription failed
	ErrPubSubBackpressure    ErrorCode = "PUBSUB_004" // Backpressure detected

	// Security Errors
	ErrSecurityInvalidToken     ErrorCode = "SECURITY_001" // Invalid security token
	ErrSecurityTokenExpired     ErrorCode = "SECURITY_002" // Security token expired
	ErrSecurityInvalidSignature ErrorCode = "SECURITY_003" // Invalid cryptographic signature
	ErrSecurityAbuseDetected    ErrorCode = "SECURITY_004" // Abuse pattern detected

	// Health Check Errors
	ErrHealthCheckFailed        ErrorCode = "HEALTH_001" // Health check failed
	ErrHealthComponentUnhealthy ErrorCode = "HEALTH_002" // Component unhealthy
	ErrHealthReadinessNotReady  ErrorCode = "HEALTH_003" // Service not ready

	// Metrics Errors
	ErrMetricsCollectorFull ErrorCode = "METRICS_001" // Metrics collector at capacity
	ErrMetricsInvalidRecord ErrorCode = "METRICS_002" // Invalid metrics record

	// File System Errors
	ErrFileNotFound         ErrorCode = "FS_001" // File not found
	ErrFilePermissionDenied ErrorCode = "FS_002" // File permission denied
	ErrFileReadFailed       ErrorCode = "FS_003" // File read failed
	ErrFileWriteFailed      ErrorCode = "FS_004" // File write failed
)

// ErrorDetail provides additional context for an error.
type ErrorDetail struct {
	Code        ErrorCode              `json:"code"`
	Message     string                 `json:"message"`
	Field       string                 `json:"field,omitempty"`
	Value       interface{}            `json:"value,omitempty"`
	Constraints map[string]interface{} `json:"constraints,omitempty"`
}

// StructuredError represents a standardized error response.
type StructuredError struct {
	Detail   ErrorDetail            `json:"error"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewStructuredError creates a new structured error.
func NewStructuredError(code ErrorCode, message string, metadata map[string]interface{}) StructuredError {
	return StructuredError{
		Detail: ErrorDetail{
			Code:    code,
			Message: message,
		},
		Metadata: metadata,
	}
}

// WithField adds field information to the error.
func (e StructuredError) WithField(field string, value interface{}) StructuredError {
	e.Detail.Field = field
	e.Detail.Value = value
	return e
}

// WithConstraints adds validation constraints to the error.
func (e StructuredError) WithConstraints(constraints map[string]interface{}) StructuredError {
	e.Detail.Constraints = constraints
	return e
}

// WithMetadata adds additional metadata to the error.
func (e StructuredError) WithMetadata(key string, value interface{}) StructuredError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// Error implements the error interface.
func (e StructuredError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Detail.Code, e.Detail.Message)
}
