package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"strconv"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"
	"standard-service/internal/domain/item"
)

const (
	codeItemCreateInvalidJSON    = "item.create.invalid_json"
	codeItemCreateBodyRequired   = "item.create.body_required"
	codeItemCreateFailed         = "item.create.failed"
	codeItemFieldsRequired       = "item.fields.required"
	codeItemUpdateFieldsRequired = "item.update.fields_required"
	codeItemUpdateFailed         = "item.update.failed"
	codeItemNotFound             = "item.not_found"
	codeItemUpdateInvalidJSON    = "item.update.invalid_json"
	codeItemUpdateBodyRequired   = "item.update.body_required"
	codeItemListFailed           = "item.list.failed"
	codeItemListInvalidParam     = "item.list.invalid_param"
	codeItemPatchInvalidJSON     = "item.patch.invalid_json"
	codeItemPatchBodyRequired    = "item.patch.body_required"
	codeItemPatchNoFields        = "item.patch.no_fields"
	codeItemPatchFailed          = "item.patch.failed"
	codeItemDeleteFailed         = "item.delete.failed"
)

// noLimit is passed as maxVal to parseQueryInt when no upper bound is needed.
// Using math.MaxInt means the clamp condition (n > maxVal) will never trigger
// for any realistic query parameter value.
const noLimit = math.MaxInt

// ItemRepository is the minimal persistence contract that ItemHandler depends on.
// All methods accept a context so callers can propagate request deadlines and
// cancellation to real storage backends. Pass a concrete implementation from
// routes.go; pass a stub in tests.
//
// The error return on mutating methods distinguishes storage failures (non-nil error)
// from "not found" (bool false, nil error) so handlers can return 500 vs 404.
// Real backends propagate ctx to the underlying driver; the in-memory implementation
// checks ctx.Err() before acquiring its lock so the contract holds for both.
type ItemRepository interface {
	Create(ctx context.Context, name, description string) (item.Item, error)
	Get(ctx context.Context, id string) (item.Item, bool)
	List(ctx context.Context, offset, limit int) ([]item.Item, int, error)
	Update(ctx context.Context, id, name, description string) (item.Item, bool, error)
	Patch(ctx context.Context, id, name, description string) (item.Item, bool, error)
	Delete(ctx context.Context, id string) (bool, error)
}

// ItemHandler demonstrates constructor injection: declare the dependency as an
// interface field and wire the concrete implementation in routes.go.
//
// Logger must not be nil; pass a.Core.Logger() from routes.go.
// Use plumelog.NewLogger(plumelog.LoggerConfig{Format: plumelog.LoggerFormatDiscard}) in tests.
type ItemHandler struct {
	Repo   ItemRepository
	Logger plumelog.StructuredLogger
}

// createItemReq, updateItemReq, and patchItemReq carry the same fields today but
// are kept as distinct types because their validation semantics differ: Create and
// Update require both fields; Patch requires at least one. Keeping them separate
// also leaves room to diverge — for example, Patch could add pointer fields for
// explicit-null semantics without changing the Create or Update contracts.
type createItemReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type updateItemReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type patchItemReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// requireItemFields checks that name and description are both non-empty.
// It adds a per-field detail entry to eb for each empty field and returns false
// when at least one is missing, so the caller can return a single error response
// that names every problem the client needs to fix in one round trip.
func requireItemFields(eb *contract.ErrorBuilder, name, description string) (*contract.ErrorBuilder, bool) {
	valid := true
	if name == "" {
		eb = eb.Detail("name", "field is required")
		valid = false
	}
	if description == "" {
		eb = eb.Detail("description", "field is required")
		valid = false
	}
	return eb, valid
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
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeRequired).
				Code(codeItemCreateBodyRequired).
				Message("request body is required").
				Build()))
			return
		}
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Code(codeItemCreateInvalidJSON).
			Message("request body must be valid JSON").
			Build()))
		return
	}

	eb := contract.NewErrorBuilder().
		Type(contract.TypeRequired).
		Code(codeItemFieldsRequired).
		Message("one or more required fields are missing")
	if eb, ok := requireItemFields(eb, req.Name, req.Description); !ok {
		logWriteErr(h.Logger, contract.WriteError(w, r, eb.Build()))
		return
	}

	created, err := h.Repo.Create(r.Context(), req.Name, req.Description)
	if err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Code(codeItemCreateFailed).
			Message("failed to create item").
			Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusCreated, created, nil))
}

// List handles GET /api/v1/items with basic limit/offset pagination.
// Items are returned in stable creation order.
// Pagination metadata (total, limit, offset) is carried in the response meta field,
// not embedded in the data payload, demonstrating the canonical meta parameter usage.
// The meta offset field reflects the requested offset, not a clamped internal value,
// so clients always know exactly where their page begins relative to the full list.
// Pagination is delegated to the repository so real storage backends can issue a
// single LIMIT/OFFSET query rather than loading every row into memory.
//
//	GET /api/v1/items                    → 200 data:[…] meta:{total:N,limit:20,offset:0}
//	GET /api/v1/items?limit=5            → first 5 items; limit capped at 100
//	GET /api/v1/items?limit=5&offset=10  → items 11-15
//	GET /api/v1/items?offset=999         → empty data; meta:{total:N,limit:20,offset:999}
//	GET /api/v1/items?limit=abc          → 400 TypeBadRequest item.list.invalid_param
func (h ItemHandler) List(w http.ResponseWriter, r *http.Request) {
	const defaultLimit = 20
	const maxLimit = 100

	limit, ok := parseQueryInt(r, "limit", defaultLimit, 1, maxLimit)
	if !ok {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Code(codeItemListInvalidParam).
			Detail("field", "limit").
			Message("limit must be a positive integer").
			Build()))
		return
	}

	offset, ok := parseQueryInt(r, "offset", 0, 0, noLimit)
	if !ok {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Code(codeItemListInvalidParam).
			Detail("field", "offset").
			Message("offset must be a non-negative integer").
			Build()))
		return
	}

	// Delegate pagination to the repository. The meta offset field carries the
	// requested offset (not a clamped value) so clients always know exactly where
	// their page was requested relative to the full list.
	page, total, err := h.Repo.List(r.Context(), offset, limit)
	if err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Code(codeItemListFailed).
			Message("failed to list items").
			Build()))
		return
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, page, map[string]any{
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}))
}

// parseQueryInt parses an integer query parameter.
// Returns (defaultVal, true) when the parameter is absent.
// Returns (0, false) when the parameter is present but not a valid integer or is below minVal;
// the caller should write a 400 error in that case.
// When n > maxVal, the value is silently clamped to maxVal.
// Pass noLimit (math.MaxInt) for maxVal to allow any value above minVal without capping.
func parseQueryInt(r *http.Request, key string, defaultVal, minVal, maxVal int) (int, bool) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return defaultVal, true
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < minVal {
		return 0, false
	}
	if n > maxVal {
		return maxVal, true
	}
	return n, true
}

// GetByID handles GET /api/v1/items/:id.
// It demonstrates path parameter extraction and the canonical 404 error path.
//
//	GET /api/v1/items/<id>    → 200 item
//	GET /api/v1/items/missing → 404 TypeNotFound
func (h ItemHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	found, ok := h.Repo.Get(r.Context(), id)
	if !ok {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Code(codeItemNotFound).
			Detail("id", id).
			Message("item not found").
			Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, found, nil))
}

// Update handles PUT /api/v1/items/:id.
// PUT is idempotent and replaces all mutable fields (name, description).
// id and created_at are immutable and are never replaced.
// Collect all missing required fields before writing an error response,
// matching the same pattern as Create so clients fix everything in one round trip.
//
//	PUT /api/v1/items/<id> {"name":"renamed","description":"new"}  → 200 updated item
//	PUT /api/v1/items/<id>                                         → 400 TypeRequired (empty body)
//	PUT /api/v1/items/<id> {}                                      → 400 TypeRequired (name + description)
//	PUT /api/v1/items/missing {"name":"x","description":"y"}       → 404 TypeNotFound
func (h ItemHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")

	var req updateItemReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if errors.Is(err, io.EOF) {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeRequired).
				Code(codeItemUpdateBodyRequired).
				Message("request body is required").
				Build()))
			return
		}
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Code(codeItemUpdateInvalidJSON).
			Message("request body must be valid JSON").
			Build()))
		return
	}

	eb := contract.NewErrorBuilder().
		Type(contract.TypeRequired).
		Code(codeItemUpdateFieldsRequired).
		Message("one or more required fields are missing")
	if eb, ok := requireItemFields(eb, req.Name, req.Description); !ok {
		logWriteErr(h.Logger, contract.WriteError(w, r, eb.Build()))
		return
	}

	updated, ok, err := h.Repo.Update(r.Context(), id, req.Name, req.Description)
	if err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Code(codeItemUpdateFailed).
			Message("failed to update item").
			Build()))
		return
	}
	if !ok {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Code(codeItemNotFound).
			Detail("id", id).
			Message("item not found").
			Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, updated, nil))
}

// Patch handles PATCH /api/v1/items/:id.
// Unlike PUT, PATCH applies a partial update: only the non-empty fields in the
// request body are written; fields absent or empty in the body are left unchanged.
// At least one of name or description must be non-empty.
//
//	PATCH /api/v1/items/<id> {"name":"renamed"}                          → 200 (name changed, description preserved)
//	PATCH /api/v1/items/<id> {"description":"new desc"}                  → 200 (description changed, name preserved)
//	PATCH /api/v1/items/<id> {"name":"renamed","description":"new desc"} → 200 (both changed)
//	PATCH /api/v1/items/<id>                                             → 400 TypeRequired (empty body)
//	PATCH /api/v1/items/<id> {}                                          → 400 TypeRequired (no fields)
//	PATCH /api/v1/items/missing {"name":"x"}                             → 404 TypeNotFound
func (h ItemHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")

	var req patchItemReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if errors.Is(err, io.EOF) {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeRequired).
				Code(codeItemPatchBodyRequired).
				Message("request body is required").
				Build()))
			return
		}
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Code(codeItemPatchInvalidJSON).
			Message("request body must be valid JSON").
			Build()))
		return
	}

	if req.Name == "" && req.Description == "" {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeRequired).
			Code(codeItemPatchNoFields).
			Message("at least one of name or description is required").
			Build()))
		return
	}

	updated, ok, err := h.Repo.Patch(r.Context(), id, req.Name, req.Description)
	if err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Code(codeItemPatchFailed).
			Message("failed to patch item").
			Build()))
		return
	}
	if !ok {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Code(codeItemNotFound).
			Detail("id", id).
			Message("item not found").
			Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, updated, nil))
}

// Delete handles DELETE /api/v1/items/:id.
// A successful delete returns 204 No Content with no body.
//
//	DELETE /api/v1/items/<id>    → 204
//	DELETE /api/v1/items/missing → 404 TypeNotFound
func (h ItemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	ok, err := h.Repo.Delete(r.Context(), id)
	if err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Code(codeItemDeleteFailed).
			Message("failed to delete item").
			Build()))
		return
	}
	if !ok {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Code(codeItemNotFound).
			Detail("id", id).
			Message("item not found").
			Build()))
		return
	}
	// WriteResponse with 204 writes only the status line — contract.writeJSON
	// skips the body for statusDisallowsBody statuses (1xx, 204, 304).
	// Using WriteResponse keeps the response path consistent with all other handlers.
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusNoContent, nil, nil))
}
