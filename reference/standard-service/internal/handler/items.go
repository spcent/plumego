package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
	"standard-service/internal/domain/item"
)

const (
	codeItemCreateInvalidJSON = "item.create.invalid_json"
	codeItemNameRequired      = "item.name.required"
	codeItemNotFound          = "item.not_found"
	codeItemUpdateInvalidJSON = "item.update.invalid_json"
)

// ItemRepository is the minimal persistence contract that ItemHandler depends on.
// Pass a concrete implementation from routes.go; pass a stub in tests.
type ItemRepository interface {
	Create(name string) item.Item
	Get(id string) (item.Item, bool)
	List() []item.Item
	Update(id, name string) (item.Item, bool)
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

type updateItemReq struct {
	Name string `json:"name"`
}

// listResponse is the paginated list envelope returned by GET /api/v1/items.
type listResponse struct {
	Items  []item.Item `json:"items"`
	Total  int         `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
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
			Code(codeItemCreateInvalidJSON).
			Message("request body must be valid JSON").
			Build())
		return
	}
	if req.Name == "" {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeRequired).
			Code(codeItemNameRequired).
			Detail("field", "name").
			Message("name is required").
			Build())
		return
	}
	item := h.Repo.Create(req.Name)
	_ = contract.WriteResponse(w, r, http.StatusCreated, item, nil)
}

// List handles GET /api/v1/items with basic limit/offset pagination.
// Items are returned in stable creation order.
//
//	GET /api/v1/items                    → 200 {items:[…], total:N, limit:20, offset:0}
//	GET /api/v1/items?limit=5            → first 5 items; limit capped at 100
//	GET /api/v1/items?limit=5&offset=10  → items 11-15
func (h ItemHandler) List(w http.ResponseWriter, r *http.Request) {
	const defaultLimit = 20
	const maxLimit = 100

	limit := listQueryInt(r, "limit", defaultLimit, 1, maxLimit)
	offset := listQueryInt(r, "offset", 0, 0, 0) // 0 maxVal = uncapped

	all := h.Repo.List()
	total := len(all)
	if offset > len(all) {
		offset = len(all)
	}
	page := all[offset:]
	if len(page) > limit {
		page = page[:limit]
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, listResponse{
		Items: page, Total: total, Limit: limit, Offset: offset,
	}, nil)
}

// listQueryInt parses a non-negative integer query parameter.
// Falls back to defaultVal when absent, non-numeric, or below minVal.
// Clamps to maxVal when maxVal > 0.
func listQueryInt(r *http.Request, key string, defaultVal, minVal, maxVal int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < minVal {
		return defaultVal
	}
	if maxVal > 0 && n > maxVal {
		return maxVal
	}
	return n
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
			Code(codeItemNotFound).
			Detail("id", id).
			Message("item not found").
			Build())
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, item, nil)
}

// Update handles PUT /api/v1/items/:id.
// PUT is idempotent: repeated calls with the same body yield the same result.
// Only the name field is replaceable; id and created_at are immutable.
//
//	PUT /api/v1/items/item-1 {"name":"renamed"} → 200 updated item
//	PUT /api/v1/items/item-1 {}                 → 400 TypeRequired
//	PUT /api/v1/items/missing {"name":"x"}      → 404 TypeNotFound
func (h ItemHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")

	var req updateItemReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Code(codeItemUpdateInvalidJSON).
			Message("request body must be valid JSON").
			Build())
		return
	}
	if req.Name == "" {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeRequired).
			Code(codeItemNameRequired).
			Detail("field", "name").
			Message("name is required").
			Build())
		return
	}

	item, ok := h.Repo.Update(id, req.Name)
	if !ok {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Code(codeItemNotFound).
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
			Code(codeItemNotFound).
			Detail("id", id).
			Message("item not found").
			Build())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
