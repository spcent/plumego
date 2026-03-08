package bind

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/spcent/plumego/contract"
	logpkg "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/middleware"
	"github.com/spcent/plumego/validator"
)

// ErrorHandler allows customizing bind/validation error responses.
type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error, fields []contract.FieldError)

// JSONOptions configures BindJSON behavior.
type JSONOptions struct {
	MaxBodyBytes          int64
	DisallowUnknownFields bool
	DisableValidation     bool
	Validator             func(any) error
	Logger                logpkg.StructuredLogger
	Redactor              *Redactor
	OnError               ErrorHandler
}

// BindJSON parses JSON into a struct and validates it.
//
// # Canonical style note
//
// This middleware-first binding pattern is NOT canonical for new Plumego code.
// Per the style guide (§9.3), canonical handlers should decode JSON explicitly:
//
//	var req MyRequest
//	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
//	    contract.WriteError(w, r, contract.NewError(...))
//	    return
//	}
//
// BindJSON is provided as an optional compatibility helper for advanced cases
// where centralized validation middleware is strongly preferred. It must not be
// used in canonical examples or generated scaffolding.
func BindJSON[T any](opts JSONOptions) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			payload, body, err := decodeJSON[T](w, r, opts)
			if err != nil {
				fields := contract.FieldErrorsFrom(err)
				logBindError(r, opts, err, payload, body, fields)
				writeBindError(w, r, opts, err, fields)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func decodeJSON[T any](w http.ResponseWriter, r *http.Request, opts JSONOptions) (T, []byte, error) {
	var payload T
	if r == nil {
		return payload, nil, contract.ErrRequestNil
	}

	body, err := readBody(w, r, opts.MaxBodyBytes)
	if err != nil {
		return payload, body, err
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return payload, body, contract.ErrEmptyRequestBody
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	if opts.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	if err := decoder.Decode(&payload); err != nil {
		return payload, body, contract.ErrInvalidJSON
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return payload, body, contract.ErrUnexpectedExtraData
	}

	validate := opts.Validator
	if validate == nil && !opts.DisableValidation {
		validate = validator.Validate
	}
	if validate != nil {
		if err := validate(&payload); err != nil {
			return payload, body, err
		}
	}

	return payload, body, nil
}

func readBody(w http.ResponseWriter, r *http.Request, maxBytes int64) ([]byte, error) {
	reader := io.Reader(r.Body)
	if maxBytes > 0 {
		reader = http.MaxBytesReader(w, r.Body, maxBytes)
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return body, contract.ErrRequestBodyTooLarge
		}
		return body, err
	}
	// Restore body for downstream consumers.
	r.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

func writeBindError(w http.ResponseWriter, r *http.Request, opts JSONOptions, err error, fields []contract.FieldError) {
	if opts.OnError != nil {
		opts.OnError(w, r, err, fields)
		return
	}
	contract.WriteBindError(w, r, err)
}

func logBindError(r *http.Request, opts JSONOptions, err error, payload any, body []byte, fields []contract.FieldError) {
	if opts.Logger == nil {
		return
	}

	fieldsLog := logpkg.Fields{
		"error":    err.Error(),
		"method":   "",
		"path":     "",
		"trace_id": contract.TraceIDFromContext(r.Context()),
	}
	if r != nil {
		fieldsLog["method"] = r.Method
		fieldsLog["path"] = r.URL.Path
	}
	if len(fields) > 0 {
		fieldsLog["validation_fields"] = fields
	}

	if len(fields) > 0 {
		redactor := opts.Redactor
		if redactor == nil {
			redactor = DefaultRedactor()
		}
		fieldsLog["payload"] = contract.DefaultObservabilityPolicy.RedactFields(map[string]any{"payload": redactor.Redact(payload)})["payload"]
	} else if len(body) > 0 {
		fieldsLog["body_bytes"] = len(body)
	}

	opts.Logger.WarnCtx(r.Context(), "request binding failed", logpkg.Fields(contract.DefaultObservabilityPolicy.RedactFields(fieldsLog)))
}
