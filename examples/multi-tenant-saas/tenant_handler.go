package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/spcent/plumego"
	"github.com/spcent/plumego/store/db"
)

// AdminHandler handles tenant administration
type AdminHandler struct {
	manager *db.DBTenantConfigManager
}

// CreateTenantRequest represents the request body for creating a tenant
type CreateTenantRequest struct {
	TenantID             string            `json:"tenant_id"`
	QuotaRequestsPerMin  int               `json:"quota_requests_per_minute"`
	QuotaTokensPerMin    int               `json:"quota_tokens_per_minute"`
	AllowedModels        []string          `json:"allowed_models"`
	AllowedTools         []string          `json:"allowed_tools"`
	Metadata             map[string]string `json:"metadata,omitempty"`
}

// TenantResponse represents the response for tenant operations
type TenantResponse struct {
	TenantID    string                       `json:"tenant_id"`
	Quota       plumego.TenantQuotaConfig    `json:"quota"`
	Policy      plumego.TenantPolicyConfig   `json:"policy"`
	Metadata    map[string]string            `json:"metadata,omitempty"`
	UpdatedAt   time.Time                    `json:"updated_at"`
}

// CreateTenant creates a new tenant
func (h *AdminHandler) CreateTenant(ctx *plumego.Context) {
	var req CreateTenantRequest
	if err := json.NewDecoder(ctx.R.Body).Decode(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Validate tenant ID
	if req.TenantID == "" {
		ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "tenant_id is required",
		})
		return
	}

	// Check if tenant already exists
	existing, _ := h.manager.GetTenantConfig(context.Background(), req.TenantID)
	if existing.TenantID != "" {
		ctx.JSON(http.StatusConflict, map[string]string{
			"error": "Tenant already exists",
		})
		return
	}

	// Create tenant configuration
	config := plumego.TenantConfig{
		TenantID: req.TenantID,
		Quota: plumego.TenantQuotaConfig{
			RequestsPerMinute: req.QuotaRequestsPerMin,
			TokensPerMinute:   req.QuotaTokensPerMin,
		},
		Policy: plumego.TenantPolicyConfig{
			AllowedModels: req.AllowedModels,
			AllowedTools:  req.AllowedTools,
		},
		Metadata:  req.Metadata,
		UpdatedAt: time.Now(),
	}

	// Save tenant
	if err := h.manager.SetTenantConfig(context.Background(), config); err != nil {
		log.Printf("Failed to create tenant: %v", err)
		ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create tenant",
		})
		return
	}

	log.Printf("Created tenant: %s (quota: %d req/min, %d tok/min)",
		req.TenantID, req.QuotaRequestsPerMin, req.QuotaTokensPerMin)

	ctx.JSON(http.StatusCreated, toTenantResponse(config))
}

// GetTenant retrieves a tenant by ID
func (h *AdminHandler) GetTenant(ctx *plumego.Context) {
	tenantID, _ := ctx.Param("id")
	if tenantID == "" {
		ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "tenant_id is required",
		})
		return
	}

	config, err := h.manager.GetTenantConfig(context.Background(), tenantID)
	if err != nil {
		if err == plumego.ErrTenantNotFound {
			ctx.JSON(http.StatusNotFound, map[string]string{
				"error": "Tenant not found",
			})
			return
		}
		log.Printf("Failed to get tenant: %v", err)
		ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve tenant",
		})
		return
	}

	ctx.JSON(http.StatusOK, toTenantResponse(config))
}

// UpdateTenant updates an existing tenant
func (h *AdminHandler) UpdateTenant(ctx *plumego.Context) {
	tenantID, _ := ctx.Param("id")
	if tenantID == "" {
		ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "tenant_id is required",
		})
		return
	}

	var req CreateTenantRequest
	if err := json.NewDecoder(ctx.R.Body).Decode(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Verify tenant exists
	_, err := h.manager.GetTenantConfig(context.Background(), tenantID)
	if err != nil {
		if err == plumego.ErrTenantNotFound {
			ctx.JSON(http.StatusNotFound, map[string]string{
				"error": "Tenant not found",
			})
			return
		}
		log.Printf("Failed to get tenant: %v", err)
		ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve tenant",
		})
		return
	}

	// Update configuration
	config := plumego.TenantConfig{
		TenantID: tenantID,
		Quota: plumego.TenantQuotaConfig{
			RequestsPerMinute: req.QuotaRequestsPerMin,
			TokensPerMinute:   req.QuotaTokensPerMin,
		},
		Policy: plumego.TenantPolicyConfig{
			AllowedModels: req.AllowedModels,
			AllowedTools:  req.AllowedTools,
		},
		Metadata:  req.Metadata,
		UpdatedAt: time.Now(),
	}

	if err := h.manager.SetTenantConfig(context.Background(), config); err != nil {
		log.Printf("Failed to update tenant: %v", err)
		ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update tenant",
		})
		return
	}

	log.Printf("Updated tenant: %s", tenantID)

	ctx.JSON(http.StatusOK, toTenantResponse(config))
}

// DeleteTenant deletes a tenant
func (h *AdminHandler) DeleteTenant(ctx *plumego.Context) {
	tenantID, _ := ctx.Param("id")
	if tenantID == "" {
		ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "tenant_id is required",
		})
		return
	}

	if err := h.manager.DeleteTenantConfig(context.Background(), tenantID); err != nil {
		if err == plumego.ErrTenantNotFound {
			ctx.JSON(http.StatusNotFound, map[string]string{
				"error": "Tenant not found",
			})
			return
		}
		log.Printf("Failed to delete tenant: %v", err)
		ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete tenant",
		})
		return
	}

	log.Printf("Deleted tenant: %s", tenantID)

	ctx.JSON(http.StatusOK, map[string]string{
		"message": "Tenant deleted successfully",
	})
}

// ListTenants lists all tenants
func (h *AdminHandler) ListTenants(ctx *plumego.Context) {
	tenants, err := h.manager.ListTenants(context.Background(), 100, 0)
	if err != nil {
		log.Printf("Failed to list tenants: %v", err)
		ctx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to list tenants",
		})
		return
	}

	response := make([]TenantResponse, 0, len(tenants))
	for _, config := range tenants {
		response = append(response, toTenantResponse(config))
	}

	ctx.JSON(http.StatusOK, map[string]interface{}{
		"tenants": response,
		"count":   len(response),
	})
}

// toTenantResponse converts a TenantConfig to a TenantResponse
func toTenantResponse(config plumego.TenantConfig) TenantResponse {
	return TenantResponse{
		TenantID:  config.TenantID,
		Quota:     config.Quota,
		Policy:    config.Policy,
		Metadata:  config.Metadata,
		UpdatedAt: config.UpdatedAt,
	}
}
