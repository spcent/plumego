package contract

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// BindJSON binds the request JSON body to the provided destination structure.
// It performs minimal decoding and returns a bindError on failure.
// The optional opts argument tightens per-call JSON behavior; omit it to use defaults.
func (c *Ctx) BindJSON(dst any, opts ...BindOptions) error {
	if err := c.requireRequest(); err != nil {
		return err
	}

	var opt BindOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	if opt.MaxBodySize < 0 {
		return invalidBodySizeError()
	}

	data, err := c.bodyBytes()
	if err != nil {
		if errors.Is(err, ErrInvalidParam) {
			return err
		}
		if errors.Is(err, ErrRequestBodyTooLarge) {
			// err IS ErrRequestBodyTooLarge (set directly by bodyBytes); pass through.
			return &bindError{Status: http.StatusRequestEntityTooLarge, Message: ErrRequestBodyTooLarge.Error(), Err: err}
		}
		return &bindError{Status: http.StatusBadRequest, Message: "failed to read request body", Err: err}
	}

	if opt.MaxBodySize > 0 && int64(len(data)) > opt.MaxBodySize {
		return &bindError{Status: http.StatusRequestEntityTooLarge, Message: ErrRequestBodyTooLarge.Error(), Err: ErrRequestBodyTooLarge}
	}
	// opt.MaxBodySize, if positive, enforces a stricter per-call cap on the
	// already-read body. RequestConfig.MaxBodySize enforces read-time limits.
	return decodeJSONBody(data, dst, opt.DisallowUnknownFields)
}

// joinSentinel wraps sentinel and cause together so that errors.Is(e, sentinel)
// is true while the original cause is still reachable for logging and diagnosis.
// If cause is nil or identical to sentinel, sentinel is returned directly.
func joinSentinel(sentinel, cause error) error {
	if cause == nil || cause == sentinel {
		return sentinel
	}
	return fmt.Errorf("%w: %w", sentinel, cause)
}

func decodeJSONBody(data []byte, dst any, disallowUnknown bool) error {
	if dst == nil {
		return &bindError{Status: http.StatusBadRequest, Message: "bind destination must not be nil", Err: ErrInvalidBindDst}
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return &bindError{Status: http.StatusBadRequest, Message: ErrEmptyRequestBody.Error(), Err: ErrEmptyRequestBody}
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	if disallowUnknown {
		decoder.DisallowUnknownFields()
	}

	if err := decoder.Decode(dst); err != nil {
		return &bindError{Status: http.StatusBadRequest, Message: ErrInvalidJSON.Error(), Err: joinSentinel(ErrInvalidJSON, err)}
	}

	if decoder.Decode(&struct{}{}) != io.EOF {
		return &bindError{Status: http.StatusBadRequest, Message: ErrUnexpectedExtraData.Error(), Err: ErrUnexpectedExtraData}
	}

	return nil
}

func invalidBodySizeError() error {
	return &bindError{Status: http.StatusBadRequest, Message: "invalid max body size", Err: ErrInvalidParam}
}

// BindQuery binds URL query parameters to the provided struct using the "query" struct tag.
// It supports scalar primitives, pointer-to-scalar fields, and primitive slices.
// Fields without a "query" tag are skipped. A tag value of "-" also skips the field.
// Validation is an explicit second step via ValidateStruct.
func (c *Ctx) BindQuery(dst any) error {
	if err := c.requireRequest(); err != nil {
		return err
	}
	if c.R.URL == nil {
		return &bindError{Status: http.StatusBadRequest, Message: ErrRequestNil.Error(), Err: ErrRequestNil}
	}
	return bindQuery(c.R.URL.Query(), dst)
}

func (c *Ctx) requireRequest() error {
	if c == nil {
		return &bindError{Status: http.StatusBadRequest, Message: ErrContextNil.Error(), Err: ErrContextNil}
	}
	if c.R == nil {
		return &bindError{Status: http.StatusBadRequest, Message: ErrRequestNil.Error(), Err: ErrRequestNil}
	}
	return nil
}

// bindQuery maps URL query values to struct fields using the "query" struct tag.
func bindQuery(values url.Values, dst any) error {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &bindError{Status: http.StatusBadRequest, Message: "bind destination must be a non-nil pointer to a struct", Err: ErrInvalidBindDst}
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return &bindError{Status: http.StatusBadRequest, Message: "bind destination must be a pointer to a struct", Err: ErrInvalidBindDst}
	}

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		tag := field.Tag.Get("query")
		if tag == "" || tag == "-" {
			continue
		}

		// query:"name,omitempty" is accepted for parity with common Go tags.
		// Query binding already behaves as omitempty-by-default for absent values.
		name := strings.SplitN(tag, ",", 2)[0]
		queryVal := values.Get(name)
		queryVals := values[name]

		fv := rv.Field(i)
		if !fv.CanSet() {
			continue
		}

		if err := setFieldFromQuery(fv, queryVal, queryVals); err != nil {
			return &bindError{
				Status:  http.StatusBadRequest,
				Message: fmt.Sprintf("invalid query parameter %q: %v", name, err),
				Err:     fmt.Errorf("%w: %w", ErrInvalidQueryParam, err),
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
	case reflect.Ptr:
		if fv.Type().Elem().Kind() == reflect.Ptr {
			return unsupportedQueryFieldError(fv.Type())
		}
		elem := reflect.New(fv.Type().Elem()).Elem()
		if err := setFieldFromQuery(elem, val, vals); err != nil {
			return err
		}
		ptr := reflect.New(fv.Type().Elem())
		ptr.Elem().Set(elem)
		fv.Set(ptr)
	case reflect.String:
		fv.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidParam, err)
		}
		fv.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidParam, err)
		}
		fv.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidParam, err)
		}
		fv.SetFloat(n)
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidParam, err)
		}
		fv.SetBool(b)
	case reflect.Slice:
		if fv.Type().Elem().Kind() == reflect.String {
			fv.Set(reflect.ValueOf(append([]string(nil), vals...)))
			return nil
		}
		result := reflect.MakeSlice(fv.Type(), 0, len(vals))
		for _, item := range vals {
			elem := reflect.New(fv.Type().Elem()).Elem()
			if err := setFieldFromQuery(elem, item, []string{item}); err != nil {
				return err
			}
			result = reflect.Append(result, elem)
		}
		fv.Set(result)
	default:
		return unsupportedQueryFieldError(fv.Type())
	}
	return nil
}

func unsupportedQueryFieldError(t reflect.Type) error {
	return fmt.Errorf("%w: unsupported query destination type %s", ErrInvalidBindDst, t)
}

func (c *Ctx) bodyBytes() ([]byte, error) {
	c.bodyReadOnce.Do(func() {
		if err := c.requireRequest(); err != nil {
			c.bodyErr = err
			return
		}
		if c.R.Body == nil {
			c.body = nil
			return
		}
		if c.config != nil && c.config.MaxBodySize < 0 {
			c.bodyErr = invalidBodySizeError()
			return
		}

		reader := io.Reader(c.R.Body)
		maxBodySize := int64(0)
		if c.config != nil && c.config.MaxBodySize > 0 {
			maxBodySize = c.config.MaxBodySize
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
		if c.config == nil || c.config.EnableBodyCache {
			c.R.Body = io.NopCloser(bytes.NewBuffer(c.body))
		}
	})
	return c.body, c.bodyErr
}
