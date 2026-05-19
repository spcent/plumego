package usage

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/spcent/plumego/contract"
	tenantadmin "with-tenant-admin/internal/tenant/admin"
)

type TenantLookup interface {
	Get(context.Context, string) (tenantadmin.TenantRecord, error)
}

type Store interface {
	Record(context.Context, string, string, int64) error
	Report(context.Context, string) ([]UsageRecord, error)
}

type Handler struct {
	tenants TenantLookup
	store   Store
}

type RecordUsageRequest struct {
	Resource string `json:"resource"`
	Count    int64  `json:"count"`
}

func NewHandler(tenants TenantLookup, store Store) *Handler {
	if store == nil {
		store = NewInMemoryUsageStore()
	}
	return &Handler{tenants: tenants, store: store}
}

func (h *Handler) RecordUsage(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantIDFromRequest(r)
	if err := h.requireTenant(r.Context(), tenantID); err != nil {
		writeError(w, r, contract.TypeNotFound, "tenant not found")
		return
	}

	var req RecordUsageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, contract.TypeValidation, "invalid request body")
		return
	}
	resource := strings.TrimSpace(req.Resource)
	if resource == "" {
		writeError(w, r, contract.TypeValidation, "usage resource is required")
		return
	}
	if req.Count <= 0 {
		writeError(w, r, contract.TypeValidation, "usage count must be positive")
		return
	}

	if err := h.store.Record(r.Context(), tenantID, resource, req.Count); err != nil {
		writeError(w, r, contract.TypeInternal, "record usage failed")
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusAccepted, nil, nil)
}

func (h *Handler) GetUsageReport(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantIDFromRequest(r)
	if err := h.requireTenant(r.Context(), tenantID); err != nil {
		writeError(w, r, contract.TypeNotFound, "tenant not found")
		return
	}

	records, err := h.store.Report(r.Context(), tenantID)
	if errors.Is(err, ErrNotFound) {
		records = []UsageRecord{}
	} else if err != nil {
		writeError(w, r, contract.TypeInternal, "get usage report failed")
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, records, nil)
}

func (h *Handler) requireTenant(ctx context.Context, tenantID string) error {
	if tenantID == "" {
		return tenantadmin.ErrNotFound
	}
	if h.tenants == nil {
		return nil
	}
	_, err := h.tenants.Get(ctx, tenantID)
	return err
}

func tenantIDFromRequest(r *http.Request) string {
	if rc := contract.RequestContextFromContext(r.Context()); rc.Params != nil {
		if id := strings.TrimSpace(rc.Params["tenantID"]); id != "" {
			return id
		}
	}
	path := strings.TrimPrefix(r.URL.Path, "/admin/usage/")
	if path == r.URL.Path {
		return ""
	}
	return strings.TrimSpace(path)
}

func writeError(w http.ResponseWriter, r *http.Request, typ contract.ErrorType, message string) {
	_ = contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(typ).
		Message(message).
		Build())
}
