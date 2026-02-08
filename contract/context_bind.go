package contract

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

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

// ShouldBindJSON is an alias for BindJSON. It binds the request JSON body to dst
// and returns the error for the caller to handle. The naming follows the convention
// where "Should" methods return errors without writing a response.
func (c *Ctx) ShouldBindJSON(dst any) error {
	return c.BindJSON(dst)
}

// BindQuery binds URL query parameters to the provided struct using the "query" struct tag.
// It supports string, int, int64, float64, bool, and slice-of-string fields.
// Fields without a "query" tag are skipped. A tag value of "-" also skips the field.
func (c *Ctx) BindQuery(dst any) error {
	return bindQuery(c.Query, dst)
}

// bindQuery maps URL query values to struct fields using the "query" struct tag.
func bindQuery(values url.Values, dst any) error {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &BindError{Status: http.StatusBadRequest, Message: "bind destination must be a non-nil pointer to a struct"}
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return &BindError{Status: http.StatusBadRequest, Message: "bind destination must be a pointer to a struct"}
	}

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		tag := field.Tag.Get("query")
		if tag == "" || tag == "-" {
			continue
		}

		// Split tag to get the name (ignore options like omitempty for now)
		name := strings.SplitN(tag, ",", 2)[0]
		queryVal := values.Get(name)
		queryVals := values[name]

		fv := rv.Field(i)
		if !fv.CanSet() {
			continue
		}

		if err := setFieldFromQuery(fv, queryVal, queryVals); err != nil {
			return &BindError{
				Status:  http.StatusBadRequest,
				Message: fmt.Sprintf("invalid query parameter %q: %v", name, err),
				Err:     err,
			}
		}
	}

	return nil
}

// setFieldFromQuery assigns a query value to a struct field based on its type.
func setFieldFromQuery(fv reflect.Value, val string, vals []string) error {
	if val == "" && len(vals) == 0 {
		return nil
	}

	switch fv.Kind() {
	case reflect.String:
		fv.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		fv.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		fv.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		fv.SetFloat(n)
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		fv.SetBool(b)
	case reflect.Slice:
		if fv.Type().Elem().Kind() == reflect.String {
			fv.Set(reflect.ValueOf(vals))
		}
	}
	return nil
}

// FormFile returns the first file for the provided form key.
// It is a convenience wrapper around http.Request.FormFile.
func (c *Ctx) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := c.R.FormFile(name)
	return fh, err
}

// SaveUploadedFile copies an uploaded file to a destination path on disk.
// Parent directories are created automatically if they do not exist.
func (c *Ctx) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	if dir := filepath.Dir(dst); dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return err
		}
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
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
