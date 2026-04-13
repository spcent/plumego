package contract

// Canonical API error codes used by the contract module.
const (
	CodeBadRequest       = "BAD_REQUEST"
	CodeValidationError  = "VALIDATION_ERROR"
	CodeResourceNotFound = "RESOURCE_NOT_FOUND"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeForbidden        = "FORBIDDEN"
	CodeTimeout          = "TIMEOUT"
	CodeInternalError    = "INTERNAL_ERROR"
	CodeRateLimited      = "RATE_LIMITED"

	// Validation sub-codes
	CodeRequired      = "REQUIRED_FIELD_MISSING"
	CodeInvalidFormat = "INVALID_FORMAT"
	CodeOutOfRange    = "VALUE_OUT_OF_RANGE"
	CodeDuplicate     = "DUPLICATE_VALUE"

	// Auth sub-codes
	CodeInvalidToken = "INVALID_TOKEN"
	CodeExpiredToken = "EXPIRED_TOKEN"

	// Resource sub-codes
	CodeConflict      = "RESOURCE_CONFLICT"
	CodeAlreadyExists = "RESOURCE_ALREADY_EXISTS"
	CodeGone          = "RESOURCE_GONE"

	// System sub-codes
	CodeUnavailable      = "SERVICE_UNAVAILABLE"
	CodeMaintenance      = "MAINTENANCE_MODE"
	CodeMethodNotAllowed = "METHOD_NOT_ALLOWED"
	CodeNotImplemented   = "NOT_IMPLEMENTED"
	CodeBadGateway       = "BAD_GATEWAY"
	CodeInvalidRequest   = "INVALID_REQUEST"

	// Business sub-codes
	CodeInvalidState        = "INVALID_STATE"
	CodeInsufficientFunds   = "INSUFFICIENT_FUNDS"
	CodeOperationNotAllowed = "OPERATION_NOT_ALLOWED"

	// Request binding codes
	CodeRequestBindError      = "REQUEST_BIND_ERROR"
	CodeRequestBodyTooLarge   = "REQUEST_BODY_TOO_LARGE"
	CodeRequestBodyReadFailed = "REQUEST_BODY_READ_FAILED"
	CodeEmptyBody             = "EMPTY_BODY"
	CodeInvalidJSON           = "INVALID_JSON"
	CodeUnexpectedExtraData   = "UNEXPECTED_EXTRA_DATA"
	CodeInvalidQuery          = "INVALID_QUERY"
	CodeInvalidBindDst        = "INVALID_BIND_DESTINATION"
)
