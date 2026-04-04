package contract

import (
	"net/http"
	"runtime/debug"

	"github.com/spcent/plumego/log"
)

// HandleError wraps err with operation context, logs it, and writes a JSON error
// response. It is a drop-in replacement for the old ErrorHandler.Handle method
// without requiring a handler object to be constructed first.
func HandleError(
	w http.ResponseWriter,
	r *http.Request,
	logger log.StructuredLogger,
	err error,
	operation string,
	module string,
	params map[string]any,
) {
	if err == nil {
		return
	}

	wrappedErr := WrapError(err, operation, module, params)
	logErrorWithContext(logger, r, wrappedErr)
	_ = WriteError(w, r, toAPIError(wrappedErr))
}

// toAPIError converts any error to a fully-populated APIError.
func toAPIError(err error) APIError {
	if err == nil {
		return NewInternalError("unknown error")
	}

	if apiErr, ok := err.(APIError); ok {
		if apiErr.Status == 0 {
			apiErr.Status = http.StatusInternalServerError
		}
		if apiErr.Code == "" {
			apiErr.Code = http.StatusText(apiErr.Status)
		}
		if apiErr.Category == "" {
			apiErr.Category = CategoryForStatus(apiErr.Status)
			if apiErr.Category == "" {
				apiErr.Category = CategoryServer
			}
		}
		return apiErr
	}

	if wrapped, ok := err.(*WrappedErrorWithContext); ok {
		if apiErr, ok := wrapped.Err.(APIError); ok {
			return apiErr
		}
		if wrapped.Err != nil {
			return toAPIError(wrapped.Err)
		}
		return NewInternalError(wrapped.Error())
	}

	return NewInternalError(err.Error())
}

func logErrorWithContext(logger log.StructuredLogger, r *http.Request, err error) {
	if logger == nil || err == nil {
		return
	}

	details := GetErrorDetails(err)
	fields := log.Fields{
		"operation": details["operation"],
		"module":    details["module"],
	}

	if params, ok := details["params"].(map[string]any); ok {
		for k, v := range params {
			fields[k] = v
		}
	}

	if wrappedErr, ok := err.(*WrappedErrorWithContext); ok {
		for k, v := range wrappedErr.Context.Params {
			fields[k] = v
		}
	}

	if r != nil {
		if traceID := TraceIDFromContext(r.Context()); traceID != "" {
			fields["trace_id"] = traceID
		}
	}

	logger.WithFields(fields).Error(FormatError(err))
}

// HandlePanic converts a recovered panic value to a wrapped error.
func HandlePanic(logger log.StructuredLogger, r any) error {
	if r == nil {
		return nil
	}

	err := PanicToError(r)

	if logger != nil {
		logger.WithFields(log.Fields{
			"panic":     r,
			"stack":     string(debug.Stack()),
			"operation": "panic_recovery",
		}).Error("Recovered from panic")
	}

	return err
}

// SafeExecute runs fn and wraps any returned error or panic with the given context.
func SafeExecute(fn func() error, operation, module string, params map[string]any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = WrapError(PanicToError(r), operation, module, params)
		}
	}()

	err = fn()
	if err != nil {
		return WrapError(err, operation, module, params)
	}
	return nil
}
