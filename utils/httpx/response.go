package httpx

import (
	"net/http"

	"github.com/spcent/plumego/contract"
)

// Response represents a standardized JSON response structure.
//
// This struct provides a consistent format for API responses.
// It can be used with the JSON helper function to send JSON responses.
//
// Example:
//
//	import "github.com/spcent/plumego/utils/httpx"
//
//	response := httpx.Response{
//		Code: 200,
//		Msg:  "Success",
//		Data: map[string]string{"user": "john"},
//	}
//	httpx.JSON(w, http.StatusOK, response)
//
// Deprecated: use contract.Response or contract.WriteResponse instead.
type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

// JSON sends a JSON response with the given status code and data.
//
// This helper function sets the Content-Type header to application/json
// and encodes the data as JSON.
//
// Example:
//
//	import "github.com/spcent/plumego/utils/httpx"
//
//	// Send simple JSON response
//	data := map[string]string{"message": "Hello, World!"}
//	httpx.JSON(w, http.StatusOK, data)
//
//	// Send structured response
//	response := httpx.Response{
//		Code: 200,
//		Msg:  "Success",
//		Data: map[string]string{"user": "john"},
//	}
//	httpx.JSON(w, http.StatusOK, response)
//
// Note: This function ignores encoding errors. For production use,
// consider handling errors appropriately.
// Deprecated: use contract.WriteJSON or contract.WriteResponse instead.
func JSON(w http.ResponseWriter, status int, data any) {
	_ = contract.WriteJSON(w, status, data)
}
