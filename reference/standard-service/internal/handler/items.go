package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
	"standard-service/internal/domain/item"
)

const (
	codeItemCreateInvalidJSON  = "item.create.invalid_json"
	codeItemCreateBodyRequired = "item.create.body_required"
	codeItemFieldsRequired     = "item.fields.required"
	codeItemUpdateFieldsRequired = "item.update.fields.required"
	codeItemNotFound           = "item.not_found"
	codeItemUpdateInvalidJSON  = "item.update.invalid_json"
	codeItemUpdateBodyRequired = "item.update.body_required"
	codeItemListInvalidParam   = "item.list.invalid_param"
)

// ItemRepository is the minimal persistence contract that ItemHandler depends on.
// All methods accept a context so callers can propagate request deadlines and
// cancellation to real storage backends. Pass a concrete implementation from
// routes.go; pass a stub in tests.
type ItemRepository interface {
	Create(ctx context.Context, name, description string) item.Item
	Get(ctx context.Context, id string) (item.Item, bool)
	List(ctx context.Context) []item.Item
	Update(ctx context.Context, id, name, description string) (item.Item, bool)
	Delete(ctx context.Context, id string) bool
}

// ItemHandler demonstrates constructor injection: declare the dependency as an
// interface field and wire the concrete implementation in routes.go.
type ItemHandler struct {
	Repo ItemRepository
}

type createItemReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type updateItemReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Create handles POST /api/v1/items.
// It demonstrates two canonical error patterns:
//   - body/JSON errors are caught first and returned immediately
//   - field validation collects ALL missing required fields before writing an
//     error response, so clients receive a single actionable reply that names
//     every field they need to fix rather than discovering failures one at a time
//
// Examples:
//
//	POST /api/v1/items {"name":"widget","description":"a widget"}  → 201 item
//	POST /api/v1/items                                             → 400 TypeRequired  (empty body)
//	POST /api/v1/items (bad JSON)                                  → 400 TypeBadRequest
//	POST /api/v1/items {}                                          → 400 TypeRequired  details: name + description
//	POST /api/v1/items {"name":"widget"}                           → 400 TypeRequired  detail: description
func (h ItemHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createItemReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if errors.Is(err, io.EOF) {
			_ = contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeRequired).
				Code(codeItemCreateBodyRequired).
				Message("request body is required").
				Build())
			return
		}
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Code(codeItemCreateInvalidJSON).
			Message("request body must be valid JSON").
			Build())
		return
	}

	// Collect all missing required fields before writing any error response.
	// Each field gets its own key in the details map so clients can fix
	// every problem in a single round trip.
	eb := contract.NewErrorBuilder().
		Type(contract.TypeRequired).
		Code(codeItemFieldsRequired).
		Message("one or more required fields are missing")
	valid := true
	if req.Name == "" {
		eb = eb.Detail("name", "field is required")
		valid = false
	}
	if req.Description == "" {
		eb = eb.Detail("description", "field is required")
		valid = false
	}
	if !valid {
		_ = contract.WriteError(w, r, eb.Build())
		return
	}

	created := h.Repo.Create(r.Context(), req.Name, req.Description)
	_ = contract.WriteResponse(w, r, http.StatusCreated, created, nil)
}

// noLimit is the sentinel value for parseQueryInt's maxVal parameter.
// Pass noLimit (-1) to allow any value above minVal without capping.
const noLimit = -1

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

	all := h.Repo.List(r.Context())
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
// When maxVal >= 0, values above maxVal are silently clamped to maxVal.
// Pass noLimit (-1) for maxVal to allow any value above minVal without capping.
func parseQueryInt(r *http.Request, key string, defaultVal, minVal, maxVal int) (int, bool) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return defaultVal, true
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < minVal {
		return 0, false
	}
	if maxVal >= 0 && n > maxVal {
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
	found, ok := h.Repo.Get(r.Context(), id)
	if !ok {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Code(codeItemNotFound).
			Detail("id", id).
			Message("item not found").
			Build())
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, found, nil)
}

// Update handles PUT /api/v1/items/:id.
// PUT is idempotent and replaces all mutable fields (name, description).
// id and created_at are immutable and are never replaced.
// Collect all missing required fields before writing an error response,
// matching the same pattern as Create so clients fix everything in one round trip.
//
//	PUT /api/v1/items/item-1 {"name":"renamed","description":"new"}  → 200 updated item
//	PUT /api/v1/items/item-1                                         → 400 TypeRequired (empty body)
//	PUT /api/v1/items/item-1 {}                                      → 400 TypeRequired (name + description)
//	PUT /api/v1/items/missing {"name":"x","description":"y"}         → 404 TypeNotFound
func (h ItemHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")

	var req updateItemReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if errors.Is(err, io.EOF) {
			_ = contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeRequired).
				Code(codeItemUpdateBodyRequired).
				Message("request body is required").
				Build())
			return
		}
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Code(codeItemUpdateInvalidJSON).
			Message("request body must be valid JSON").
			Build())
		return
	}

	eb := contract.NewErrorBuilder().
		Type(contract.TypeRequired).
		Code(codeItemUpdateFieldsRequired).
		Message("one or more required fields are missing")
	valid := true
	if req.Name == "" {
		eb = eb.Detail("name", "field is required")
		valid = false
	}
	if req.Description == "" {
		eb = eb.Detail("description", "field is required")
		valid = false
	}
	if !valid {
		_ = contract.WriteError(w, r, eb.Build())
		return
	}

	updated, ok := h.Repo.Update(r.Context(), id, req.Name, req.Description)
	if !ok {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Code(codeItemNotFound).
			Detail("id", id).
			Message("item not found").
			Build())
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, updated, nil)
}

// Delete handles DELETE /api/v1/items/:id.
// A successful delete returns 204 No Content with no body.
//
//	DELETE /api/v1/items/item-1  → 204
//	DELETE /api/v1/items/missing → 404 TypeNotFound
func (h ItemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	if !h.Repo.Delete(r.Context(), id) {
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
