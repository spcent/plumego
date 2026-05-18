package admin

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
)

type Handler struct {
	store *InMemoryStore
}

type CreateTenantRequest struct {
	Name string `json:"name"`
}

func NewHandler(store *InMemoryStore) *Handler {
	if store == nil {
		store = NewInMemoryStore()
	}
	return &Handler{store: store}
}

func (h *Handler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	var req CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, contract.TypeValidation, "invalid request body")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeError(w, r, contract.TypeValidation, "tenant name is required")
		return
	}

	record, err := h.store.Create(r.Context(), TenantRecord{
		ID:        newTenantID(),
		Name:      name,
		Status:    StatusActive,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		writeError(w, r, contract.TypeInternal, "create tenant failed")
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusCreated, record, nil)
}

func (h *Handler) GetTenant(w http.ResponseWriter, r *http.Request) {
	record, err := h.store.Get(r.Context(), tenantIDFromRequest(r))
	if errors.Is(err, ErrNotFound) {
		writeError(w, r, contract.TypeNotFound, "tenant not found")
		return
	}
	if err != nil {
		writeError(w, r, contract.TypeInternal, "get tenant failed")
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, record, nil)
}

func (h *Handler) SuspendTenant(w http.ResponseWriter, r *http.Request) {
	record, err := h.store.Suspend(r.Context(), tenantIDFromRequest(r))
	if errors.Is(err, ErrNotFound) {
		writeError(w, r, contract.TypeNotFound, "tenant not found")
		return
	}
	if err != nil {
		writeError(w, r, contract.TypeInternal, "suspend tenant failed")
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, record, nil)
}

func (h *Handler) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	err := h.store.Delete(r.Context(), tenantIDFromRequest(r))
	if errors.Is(err, ErrNotFound) {
		writeError(w, r, contract.TypeNotFound, "tenant not found")
		return
	}
	if err != nil {
		writeError(w, r, contract.TypeInternal, "delete tenant failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func tenantIDFromRequest(r *http.Request) string {
	if rc := contract.RequestContextFromContext(r.Context()); rc.Params != nil {
		if id := strings.TrimSpace(rc.Params["id"]); id != "" {
			return id
		}
	}
	path := strings.TrimPrefix(r.URL.Path, "/admin/tenants/")
	path = strings.TrimSuffix(path, "/suspend")
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

func newTenantID() string {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "tenant-" + time.Now().UTC().Format("20060102150405.000000000")
	}
	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80
	encoded := hex.EncodeToString(raw[:])
	return encoded[0:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:]
}
