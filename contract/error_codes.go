package contract

// Canonical API error codes used by the contract module.
const (
	CodeValidationError  = "VALIDATION_ERROR"
	CodeResourceNotFound = "RESOURCE_NOT_FOUND"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeForbidden        = "FORBIDDEN"
	CodeTimeout          = "TIMEOUT"
	CodeInternalError    = "INTERNAL_ERROR"
	CodeRateLimited      = "RATE_LIMITED"

	CodeRequestBindError      = "REQUEST_BIND_ERROR"
	CodeRequestBodyTooLarge   = "REQUEST_BODY_TOO_LARGE"
	CodeRequestBodyReadFailed = "REQUEST_BODY_READ_FAILED"
	CodeEmptyBody             = "EMPTY_BODY"
	CodeInvalidJSON           = "INVALID_JSON"
	CodeUnexpectedExtraData   = "UNEXPECTED_EXTRA_DATA"
	CodeInvalidQuery          = "INVALID_QUERY"
	CodeInvalidBindDst        = "INVALID_BIND_DESTINATION"
)
