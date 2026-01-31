package core

import (
	"net/http"

	"github.com/spcent/plumego/contract"
)

func writeContractResponse(ctx *contract.Ctx, status int, data any) {
	if ctx == nil {
		return
	}
	_ = ctx.Response(status, data, nil)
}

func writeContractError(ctx *contract.Ctx, status int, code, message string) {
	if ctx == nil {
		return
	}
	apiErr := contract.APIError{
		Status:   status,
		Code:     code,
		Message:  message,
		Category: contract.CategoryForStatus(status),
	}
	contract.WriteError(ctx.W, ctx.R, apiErr)
}

func writeHTTPResponse(w http.ResponseWriter, r *http.Request, status int, data any) {
	_ = contract.WriteResponse(w, r, status, data, nil)
}

func writeHTTPError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	apiErr := contract.APIError{
		Status:   status,
		Code:     code,
		Message:  message,
		Category: contract.CategoryForStatus(status),
	}
	contract.WriteError(w, r, apiErr)
}
