package middleware

import (
	"encoding/json"
	"net/http"
)

// Response represents a standardized JSON response structure.
//
// This struct provides a consistent format for API responses.
// It can be used with the JSON helper function to send JSON responses.
//
// Example:
//
//	import "github.com/spcent/plumego/middleware"
//
//	response := middleware.Response{
//		Code: 200,
//		Msg:  "Success",
//		Data: map[string]string{"user": "john"},
//	}
//	middleware.JSON(w, http.StatusOK, response)
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
//	import "github.com/spcent/plumego/middleware"
//
//	// Send simple JSON response
//	data := map[string]string{"message": "Hello, World!"}
//	middleware.JSON(w, http.StatusOK, data)
//
//	// Send structured response
//	response := middleware.Response{
//		Code: 200,
//		Msg:  "Success",
//		Data: map[string]string{"user": "john"},
//	}
//	middleware.JSON(w, http.StatusOK, response)
//
// Note: This function ignores encoding errors. For production use,
// consider handling errors appropriately.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
