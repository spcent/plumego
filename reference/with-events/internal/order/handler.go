package order

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
)

type CreateOrderRequest struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	TotalCents int64  `json:"total_cents"`
}

type Order struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	TotalCents int64  `json:"total_cents"`
	Status     string `json:"status"`
}

type Handler struct {
	publisher *OrderPublisher
	mu        sync.RWMutex
	orders    map[string]Order
}

func NewHandler(publisher *OrderPublisher) *Handler {
	return &Handler{publisher: publisher, orders: make(map[string]Order)}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeValidation).
			Code(contract.CodeInvalidJSON).
			Message("invalid request body").
			Build())
		return
	}
	if strings.TrimSpace(req.ID) == "" || strings.TrimSpace(req.CustomerID) == "" || req.TotalCents <= 0 {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeValidation).
			Code(contract.CodeValidationError).
			Message("invalid order").
			Build())
		return
	}

	order := Order{ID: req.ID, CustomerID: req.CustomerID, TotalCents: req.TotalCents, Status: "accepted"}
	h.mu.Lock()
	h.orders[order.ID] = order
	h.mu.Unlock()

	if err := h.publisher.Publish(r.Context(), OrderCreated{
		ID:         "order-created-" + order.ID,
		OrderID:    order.ID,
		CustomerID: order.CustomerID,
		TotalCents: order.TotalCents,
		CreatedAt:  time.Now().UTC(),
	}); err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("publish order event failed").
			Build())
		return
	}

	_ = contract.WriteResponse(w, r, http.StatusAccepted, order, nil)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")

	h.mu.RLock()
	order, ok := h.orders[id]
	h.mu.RUnlock()
	if !ok {
		order = Order{ID: id, Status: "stub"}
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, order, nil)
}
