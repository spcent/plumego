package handler

import (
	"encoding/json"
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
	"standard-service/internal/domain/item"
)

// ItemRepository is the minimal persistence contract that ItemHandler depends on.
// Pass a concrete implementation from routes.go; pass a stub in tests.
type ItemRepository interface {
	Create(name string) item.Item
	Get(id string) (item.Item, bool)
	List() []item.Item
	Delete(id string) bool
}

// ItemHandler demonstrates constructor injection: declare the dependency as an
// interface field and wire the concrete implementation in routes.go.
type ItemHandler struct {
	Repo ItemRepository
}

type createItemReq struct {
	Name string `json:"name"`
}

// Create handles POST /api/v1/items.
// It demonstrates the canonical request body decode path and structured validation errors.
//
//	POST /api/v1/items {"name":"widget"}  → 201 {"id":"item-1","name":"widget",...}
//	POST /api/v1/items {}                 → 400 TypeRequired
//	POST /api/v1/items (bad JSON)         → 400 TypeBadRequest
func (h ItemHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createItemReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message("request body must be valid JSON").
			Build())
		return
	}
	if req.Name == "" {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeRequired).
			Detail("field", "name").
			Message("name is required").
			Build())
		return
	}
	item := h.Repo.Create(req.Name)
	_ = contract.WriteResponse(w, r, http.StatusCreated, item, nil)
}

// List handles GET /api/v1/items.
// It returns all items as a JSON array; an empty store returns an empty array.
//
//	GET /api/v1/items → 200 [items…]
func (h ItemHandler) List(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteResponse(w, r, http.StatusOK, h.Repo.List(), nil)
}

// GetByID handles GET /api/v1/items/:id.
// It demonstrates path parameter extraction and the canonical 404 error path.
//
//	GET /api/v1/items/item-1  → 200 item
//	GET /api/v1/items/missing → 404 TypeNotFound
func (h ItemHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	item, ok := h.Repo.Get(id)
	if !ok {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Detail("id", id).
			Message("item not found").
			Build())
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, item, nil)
}

// Delete handles DELETE /api/v1/items/:id.
// A successful delete returns 204 No Content with no body.
//
//	DELETE /api/v1/items/item-1  → 204
//	DELETE /api/v1/items/missing → 404 TypeNotFound
func (h ItemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	if !h.Repo.Delete(id) {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Detail("id", id).
			Message("item not found").
			Build())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
