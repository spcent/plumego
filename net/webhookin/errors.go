package webhookin

import (
	"errors"
	"net/http"
)

// ErrorCode represents a standardized webhook verification error.
type ErrorCode string

const (
	CodeMissingSignature ErrorCode = "missing_signature"
	CodeInvalidSignature ErrorCode = "invalid_signature"
	CodeInvalidEncoding  ErrorCode = "invalid_encoding"
	CodeMissingTimestamp ErrorCode = "missing_timestamp"
	CodeInvalidTimestamp ErrorCode = "invalid_timestamp"
	CodeTimestampExpired ErrorCode = "timestamp_expired"
	CodeMissingNonce     ErrorCode = "missing_nonce"
	CodeReplayDetected   ErrorCode = "replay_detected"
	CodeIPDenied         ErrorCode = "ip_denied"
	CodeBodyRead         ErrorCode = "body_read_error"
	CodeConfigError      ErrorCode = "config_error"
	CodeInternalError    ErrorCode = "internal_error"
)

// VerifyError is a structured error returned by the generic verifier.
type VerifyError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *VerifyError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return string(e.Code)
}

func (e *VerifyError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// ErrorCodeOf extracts the code from a verification error.
func ErrorCodeOf(err error) ErrorCode {
	var verr *VerifyError
	if errors.As(err, &verr) {
		return verr.Code
	}
	return ""
}

// HTTPStatus maps verification errors to HTTP status codes.
func HTTPStatus(err error) int {
	switch ErrorCodeOf(err) {
	case CodeMissingSignature, CodeMissingTimestamp, CodeMissingNonce, CodeInvalidTimestamp, CodeInvalidEncoding:
		return http.StatusBadRequest
	case CodeInvalidSignature:
		return http.StatusUnauthorized
	case CodeReplayDetected:
		return http.StatusConflict
	case CodeIPDenied:
		return http.StatusForbidden
	default:
		return http.StatusBadRequest
	}
}

func verifyError(code ErrorCode, msg string, err error) *VerifyError {
	return &VerifyError{Code: code, Message: msg, Err: err}
}
