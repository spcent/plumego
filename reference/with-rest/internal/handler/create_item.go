package handler

import (
	"errors"
	"net/http"

	"github.com/spcent/plumego/contract"
	plumego "github.com/spcent/plumego/x/validate"
	"github.com/spcent/plumego/x/validate/playground"
)

// CreateItemRequest is the validated request body for the demo create handler.
type CreateItemRequest struct {
	Name string `json:"name" validate:"required"`
}

// CreateItemResponse is the demo create response.
type CreateItemResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CreateItem handles POST /api/items with explicit binding and validation.
type CreateItem struct {
	validator plumego.Validator
}

// NewCreateItem constructs the handler with its validator dependency visible.
func NewCreateItem(validator plumego.Validator) *CreateItem {
	return &CreateItem{validator: validator}
}

// ServeHTTP implements http.Handler.
func (h *CreateItem) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req, err := plumego.Bind[CreateItemRequest](r, h.validator)
	if err != nil {
		writeBindError(w, r, err)
		return
	}

	_ = contract.WriteResponse(w, r, http.StatusCreated, CreateItemResponse{
		ID:   "demo-item",
		Name: req.Name,
	}, nil)
}

func writeBindError(w http.ResponseWriter, r *http.Request, err error) {
	var validationErr plumego.ValidationError
	if errors.As(err, &validationErr) {
		_ = contract.WriteError(w, r, validationAPIError(validationErr))
		return
	}

	_ = contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeBadRequest).
		Code(contract.CodeInvalidJSON).
		Message("invalid request body").
		Build())
}

func validationAPIError(err plumego.ValidationError) contract.APIError {
	builder := contract.NewErrorBuilder().
		Type(err.Type()).
		Code(err.Code()).
		Message(err.Message())

	var fieldErr playground.Error
	if errors.As(err.Err, &fieldErr) {
		builder.Detail("fields", fieldErr.Fields)
	}
	return builder.Build()
}
