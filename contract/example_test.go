package contract_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/spcent/plumego/contract"
)

func ExampleWriteError() {
	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	req = req.WithContext(contract.WithRequestID(req.Context(), "req-123"))
	rec := httptest.NewRecorder()

	err := contract.NewErrorBuilder().
		Type(contract.TypeNotFound).
		Message("user not found").
		Build()

	_ = contract.WriteError(rec, req, err)

	fmt.Println(rec.Code)
	fmt.Print(rec.Body.String())

	// Output:
	// 404
	// {"error":{"code":"RESOURCE_NOT_FOUND","message":"user not found","category":"client_error","type":"resource_not_found"},"request_id":"req-123"}
}

func ExampleWriteResponse() {
	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	req = req.WithContext(contract.WithRequestID(req.Context(), "req-123"))
	rec := httptest.NewRecorder()

	_ = contract.WriteResponse(rec, req, http.StatusOK, map[string]string{
		"id": "42",
	}, map[string]any{
		"count": 1,
	})

	fmt.Println(rec.Code)
	fmt.Print(rec.Body.String())

	// Output:
	// 200
	// {"data":{"id":"42"},"meta":{"count":1},"request_id":"req-123"}
}
