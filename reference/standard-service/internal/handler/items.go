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
	codeItemListInvalidParam  = "item.list.invalid_param"
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

// noLimit is the sentinel value for parseQueryInt's maxVal parameter.
// Pass noLimit to allow any value above minVal without capping.
const noLimit = 0

// List handles GET /api/v1/items with basic limit/offset pagination.
// Items are returned in stable creation order.
// Pagination metadata (total, limit, offset) is carried in the response meta field,
// not embedded in the data payload, demonstrating the canonical meta parameter usage.
//
//	GET /api/v1/items                    → 200 data:[…] meta:{total:N,limit:20,offset:0}
//	GET /api/v1/items?limit=5            → first 5 items; limit capped at 100
//	GET /api/v1/items?limit=5&offset=10  → items 11-15
//	GET /api/v1/items?limit=abc          → 400 TypeBadRequest item.list.invalid_param
func (h ItemHandler) List(w http.ResponseWriter, r *http.Request) {
	const defaultLimit = 20
	const maxLimit = 100

	limit, ok := parseQueryInt(r, "limit", defaultLimit, 1, maxLimit)
	if !ok {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Code(codeItemListInvalidParam).
			Detail("field", "limit").
			Message("limit must be a positive integer").
			Build())
		return
	}

	offset, ok := parseQueryInt(r, "offset", 0, 0, noLimit)
	if !ok {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Code(codeItemListInvalidParam).
			Detail("field", "offset").
			Message("offset must be a non-negative integer").
			Build())
		return
	}

	all := h.Repo.List()
	total := len(all)
	if offset > len(all) {
		offset = len(all)
	}
	page := all[offset:]
	if len(page) > limit {
		page = page[:limit]
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, page, map[string]any{
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// parseQueryInt parses an integer query parameter.
// Returns (defaultVal, true) when the parameter is absent.
// Returns (0, false) when the parameter is present but not a valid integer or is below minVal;
// the caller should write a 400 error in that case.
// When maxVal > 0, values above maxVal are silently clamped to maxVal.
// Pass noLimit (0) for maxVal to allow any value above minVal without capping.
func parseQueryInt(r *http.Request, key string, defaultVal, minVal, maxVal int) (int, bool) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return defaultVal, true
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < minVal {
		return 0, false
	}
	if maxVal > 0 && n > maxVal {
		return maxVal, true
	}
	return n, true
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
