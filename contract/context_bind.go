package contract

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	logpkg "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/validator"
)

// BindJSON binds the request JSON body to the provided destination structure.
// It performs minimal validation and returns a BindError on failure.
func (c *Ctx) BindJSON(dst any) error {
	data, err := c.bodyBytes()
	if err != nil {
		if errors.Is(err, ErrRequestBodyTooLarge) {
			return &BindError{Status: http.StatusRequestEntityTooLarge, Message: ErrRequestBodyTooLarge.Error(), Err: err}
		}
		return &BindError{Status: http.StatusBadRequest, Message: "failed to read request body", Err: err}
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return &BindError{Status: http.StatusBadRequest, Message: "request body is empty"}
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	// DisallowUnknownFields could be enabled if you want strict mode:
	// decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return &BindError{Status: http.StatusBadRequest, Message: "invalid JSON payload", Err: err}
	}

	// Ensure no trailing data
	if decoder.Decode(&struct{}{}) != io.EOF {
		return &BindError{Status: http.StatusBadRequest, Message: "unexpected extra JSON data"}
	}

	return nil
}

// BindJSONWithOptions binds the request JSON body with configurable options.
func (c *Ctx) BindJSONWithOptions(dst any, opts BindOptions) error {
	data, err := c.bodyBytes()
	if err != nil {
		if errors.Is(err, ErrRequestBodyTooLarge) {
			return &BindError{Status: http.StatusRequestEntityTooLarge, Message: ErrRequestBodyTooLarge.Error(), Err: err}
		}
		return &BindError{Status: http.StatusBadRequest, Message: "failed to read request body", Err: err}
	}

	if opts.MaxBodyBytes > 0 && int64(len(data)) > opts.MaxBodyBytes {
		return &BindError{Status: http.StatusRequestEntityTooLarge, Message: ErrRequestBodyTooLarge.Error(), Err: ErrRequestBodyTooLarge}
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return &BindError{Status: http.StatusBadRequest, Message: ErrEmptyRequestBody.Error(), Err: ErrEmptyRequestBody}
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	if opts.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	if err := decoder.Decode(dst); err != nil {
		return &BindError{Status: http.StatusBadRequest, Message: ErrInvalidJSON.Error(), Err: err}
	}

	if decoder.Decode(&struct{}{}) != io.EOF {
		return &BindError{Status: http.StatusBadRequest, Message: "unexpected extra JSON data", Err: ErrUnexpectedExtraData}
	}

	return nil
}

// BindAndValidateJSON binds the request body to dst and validates it using struct tags.
func (c *Ctx) BindAndValidateJSON(dst any) error {
	if err := c.BindJSON(dst); err != nil {
		return err
	}

	if err := validator.Validate(dst); err != nil {
		return &BindError{Status: http.StatusBadRequest, Message: err.Error(), Err: err}
	}

	return nil
}

// BindAndValidateJSONWithOptions binds and validates JSON payloads with configurable options.
func (c *Ctx) BindAndValidateJSONWithOptions(dst any, opts BindOptions) error {
	if err := c.BindJSONWithOptions(dst, opts); err != nil {
		logBindError(c, dst, opts, err)
		return err
	}

	if opts.DisableValidation {
		return nil
	}

	validate := opts.Validator
	if validate == nil {
		validate = validator.Validate
	}
	if err := validate(dst); err != nil {
		bindErr := &BindError{Status: http.StatusBadRequest, Message: err.Error(), Err: err}
		logBindError(c, dst, opts, bindErr)
		return bindErr
	}

	return nil
}

func logBindError(c *Ctx, payload any, opts BindOptions, err error) {
	if c == nil {
		return
	}
	logger := opts.Logger
	if logger == nil {
		return
	}

	fields := logpkg.Fields{
		"error":    err.Error(),
		"method":   c.R.Method,
		"path":     c.R.URL.Path,
		"trace_id": TraceIDFromContext(c.R.Context()),
	}

	if v := FieldErrorsFrom(err); len(v) > 0 && payload != nil {
		if opts.Redact != nil {
			fields["payload"] = opts.Redact(payload)
		} else {
			fields["payload"] = payload
		}
		fields["validation_fields"] = v
	}

	logger.WarnCtx(c.R.Context(), "request binding failed", fields)
}

func (c *Ctx) bodyBytes() ([]byte, error) {
	c.bodyReadOnce.Do(func() {
		reader := io.Reader(c.R.Body)
		maxBodySize := int64(0)
		if c.Config != nil && c.Config.MaxBodySize > 0 {
			maxBodySize = c.Config.MaxBodySize
			if c.W != nil {
				reader = http.MaxBytesReader(c.W, c.R.Body, maxBodySize)
			} else {
				reader = io.LimitReader(reader, maxBodySize+1)
			}
		}

		c.body, c.bodyErr = io.ReadAll(reader)
		if c.bodyErr != nil {
			var maxErr *http.MaxBytesError
			if errors.As(c.bodyErr, &maxErr) {
				c.bodyErr = ErrRequestBodyTooLarge
				c.body = nil
			}
			return
		}
		if maxBodySize > 0 && int64(len(c.body)) > maxBodySize {
			c.bodyErr = ErrRequestBodyTooLarge
			c.body = nil
			return
		}
		c.bodySize.Store(int64(len(c.body)))
		if c.Config == nil || c.Config.EnableBodyCache {
			c.R.Body = io.NopCloser(bytes.NewBuffer(c.body))
		}
	})
	return c.body, c.bodyErr
}
