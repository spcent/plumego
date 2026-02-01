package router

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/spcent/plumego/contract"
)

// Deprecated: use contract.Response or contract.WriteResponse instead.
// Response represents a standardized JSON response structure.
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Deprecated: use contract.APIError or contract.WriteError instead.
// ErrorResponse represents a standardized error response structure.
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// JSONWriter provides consistent JSON response writing.
// Deprecated: prefer contract.WriteResponse/WriteError for new code.
type JSONWriter struct {
	resourceName string
}

// NewJSONWriter creates a new JSON writer with resource name
func NewJSONWriter(resourceName string) *JSONWriter {
	return &JSONWriter{resourceName: resourceName}
}

// WriteJSON writes a standardized JSON response
func (jw *JSONWriter) WriteJSON(w http.ResponseWriter, status int, data any) {
	if err := contract.WriteResponse(w, nil, status, data, nil); err != nil {
		if jw != nil && jw.resourceName != "" {
			fmt.Printf("[router/resource] JSON encoding failed for %s: %v\n", jw.resourceName, err)
		} else {
			fmt.Printf("[router/resource] JSON encoding failed: %v\n", err)
		}
	}
}

// WriteError writes a standardized error response
func (jw *JSONWriter) WriteError(w http.ResponseWriter, status int, message string, details any) {
	apiErr := contract.APIError{
		Status:   status,
		Code:     http.StatusText(status),
		Message:  message,
		Category: contract.CategoryForStatus(status),
	}
	if details != nil {
		apiErr.Details = map[string]any{"details": details}
	}
	contract.WriteError(w, nil, apiErr)
}

// WriteResponse writes a contract response payload and includes trace id when available.
func (jw *JSONWriter) WriteResponse(w http.ResponseWriter, r *http.Request, status int, data any, meta map[string]any) error {
	return contract.WriteResponse(w, r, status, data, meta)
}

// WriteAPIError writes a contract error payload.
func (jw *JSONWriter) WriteAPIError(w http.ResponseWriter, r *http.Request, err contract.APIError) {
	contract.WriteError(w, r, err)
}

// WriteNotImplemented writes a standardized "Not Implemented" response
func (jw *JSONWriter) WriteNotImplemented(w http.ResponseWriter, method string) {
	resourceName := "resource"
	if jw != nil && jw.resourceName != "" {
		resourceName = jw.resourceName
	}

	jw.WriteError(w, http.StatusNotImplemented, "Not Implemented",
		fmt.Sprintf("The %s method is not implemented for the %s resource", method, resourceName))
}

// ResourceController defines the interface for RESTful resource controllers
type ResourceController interface {
	// Basic CRUD operations
	Index(http.ResponseWriter, *http.Request)  // GET /resource
	Show(http.ResponseWriter, *http.Request)   // GET /resource/:id
	Create(http.ResponseWriter, *http.Request) // POST /resource
	Update(http.ResponseWriter, *http.Request) // PUT /resource/:id
	Delete(http.ResponseWriter, *http.Request) // DELETE /resource/:id
	Patch(http.ResponseWriter, *http.Request)  // PATCH /resource/:id

	// Additional HTTP methods
	Options(http.ResponseWriter, *http.Request) // OPTIONS /resource
	Head(http.ResponseWriter, *http.Request)    // HEAD /resource

	// Batch operations (optional)
	BatchCreate(http.ResponseWriter, *http.Request) // POST /resource/batch
	BatchDelete(http.ResponseWriter, *http.Request) // DELETE /resource/batch
}

// BaseResourceController provides a default implementation of ResourceController
// All methods return "Not Implemented" error by default
type BaseResourceController struct {
	ResourceName string      // Resource name used for response messages
	jsonWriter   *JSONWriter // JSON writer for consistent responses
}

// NewBaseResourceController creates a new base resource controller with JSON writer
func NewBaseResourceController(resourceName string) *BaseResourceController {
	return &BaseResourceController{
		ResourceName: resourceName,
		jsonWriter:   NewJSONWriter(resourceName),
	}
}

// Index handles GET /resource requests
func (c *BaseResourceController) Index(w http.ResponseWriter, r *http.Request) {
	c.jsonWriter.WriteNotImplemented(w, "Index")
}

// Show handles GET /resource/:id requests
func (c *BaseResourceController) Show(w http.ResponseWriter, r *http.Request) {
	c.jsonWriter.WriteNotImplemented(w, "Show")
}

// Create handles POST /resource requests
func (c *BaseResourceController) Create(w http.ResponseWriter, r *http.Request) {
	c.jsonWriter.WriteNotImplemented(w, "Create")
}

// Update handles PUT /resource/:id requests
func (c *BaseResourceController) Update(w http.ResponseWriter, r *http.Request) {
	c.jsonWriter.WriteNotImplemented(w, "Update")
}

// Delete handles DELETE /resource/:id requests
func (c *BaseResourceController) Delete(w http.ResponseWriter, r *http.Request) {
	c.jsonWriter.WriteNotImplemented(w, "Delete")
}

// Patch handles PATCH /resource/:id requests
func (c *BaseResourceController) Patch(w http.ResponseWriter, r *http.Request) {
	c.jsonWriter.WriteNotImplemented(w, "Patch")
}

// Options handles OPTIONS /resource requests for CORS and method negotiation
func (c *BaseResourceController) Options(w http.ResponseWriter, r *http.Request) {
	// Set default CORS headers
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
	w.WriteHeader(http.StatusNoContent)
}

// Head handles HEAD /resource requests (returns headers only, no body)
func (c *BaseResourceController) Head(w http.ResponseWriter, r *http.Request) {
	// Default implementation: just return 200 OK with no body
	w.WriteHeader(http.StatusOK)
}

// BatchCreate handles batch creation of resources
func (c *BaseResourceController) BatchCreate(w http.ResponseWriter, r *http.Request) {
	c.jsonWriter.WriteNotImplemented(w, "BatchCreate")
}

// BatchDelete handles batch deletion of resources
func (c *BaseResourceController) BatchDelete(w http.ResponseWriter, r *http.Request) {
	c.jsonWriter.WriteNotImplemented(w, "BatchDelete")
}

// JSON is a helper method to write JSON responses
func (c *BaseResourceController) JSON(w http.ResponseWriter, status int, data any) {
	c.jsonWriter.WriteJSON(w, status, data)
}

// Error is a helper method to write standardized error responses
func (c *BaseResourceController) Error(w http.ResponseWriter, status int, message string, details any) {
	c.jsonWriter.WriteError(w, status, message, details)
}

// Response is a helper method to write contract-based JSON responses.
func (c *BaseResourceController) Response(w http.ResponseWriter, r *http.Request, status int, data any, meta map[string]any) {
	_ = c.jsonWriter.WriteResponse(w, r, status, data, meta)
}

// APIError is a helper method to write contract-based error responses.
func (c *BaseResourceController) APIError(w http.ResponseWriter, r *http.Request, err contract.APIError) {
	c.jsonWriter.WriteAPIError(w, r, err)
}

// ================================================
// Enhanced CRUD Framework - Query Builder
// ================================================

// QueryParams represents parsed query parameters for list operations
type QueryParams struct {
	// Pagination
	Page     int
	PageSize int
	Offset   int
	Limit    int

	// Sorting
	Sort      []SortField
	SortBy    string // Deprecated: use Sort instead
	SortOrder string // Deprecated: use Sort instead

	// Filtering
	Filters map[string]string
	Search  string

	// Field selection
	Fields []string

	// Include related resources
	Include []string
}

// SortField represents a field to sort by
type SortField struct {
	Field string
	Desc  bool
}

// QueryBuilder builds query parameters from HTTP request
type QueryBuilder struct {
	defaultPageSize int
	maxPageSize     int
	allowedSorts    map[string]bool
	allowedFilters  map[string]bool
}

// NewQueryBuilder creates a new QueryBuilder with sensible defaults
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		defaultPageSize: 20,
		maxPageSize:     100,
		allowedSorts:    make(map[string]bool),
		allowedFilters:  make(map[string]bool),
	}
}

// WithPageSize sets the default and maximum page size
func (qb *QueryBuilder) WithPageSize(defaultSize, maxSize int) *QueryBuilder {
	qb.defaultPageSize = defaultSize
	qb.maxPageSize = maxSize
	return qb
}

// WithAllowedSorts sets the allowed sort fields
func (qb *QueryBuilder) WithAllowedSorts(fields ...string) *QueryBuilder {
	qb.allowedSorts = make(map[string]bool, len(fields))
	for _, field := range fields {
		qb.allowedSorts[field] = true
	}
	return qb
}

// WithAllowedFilters sets the allowed filter fields
func (qb *QueryBuilder) WithAllowedFilters(fields ...string) *QueryBuilder {
	qb.allowedFilters = make(map[string]bool, len(fields))
	for _, field := range fields {
		qb.allowedFilters[field] = true
	}
	return qb
}

// Parse parses query parameters from an HTTP request
func (qb *QueryBuilder) Parse(r *http.Request) *QueryParams {
	query := r.URL.Query()
	params := &QueryParams{
		Page:     1,
		PageSize: qb.defaultPageSize,
		Filters:  make(map[string]string),
	}

	// Parse pagination
	if pageStr := query.Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
		}
	}
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			params.PageSize = pageSize
			if params.PageSize > qb.maxPageSize {
				params.PageSize = qb.maxPageSize
			}
		}
	}
	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			params.Limit = limit
			if params.Limit > qb.maxPageSize {
				params.Limit = qb.maxPageSize
			}
		}
	}
	if offsetStr := query.Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			params.Offset = offset
		}
	}

	// Calculate offset from page if not explicitly set
	if params.Offset == 0 && params.Page > 1 {
		params.Offset = (params.Page - 1) * params.PageSize
	}
	if params.Limit == 0 {
		params.Limit = params.PageSize
	}

	// Parse sorting - support multiple sort fields
	// Format: sort=field1,-field2,field3 (- prefix means descending)
	if sortStr := query.Get("sort"); sortStr != "" {
		fields := strings.Split(sortStr, ",")
		for _, field := range fields {
			field = strings.TrimSpace(field)
			if field == "" {
				continue
			}
			desc := strings.HasPrefix(field, "-")
			if desc {
				field = field[1:]
			}
			// Check if sort field is allowed
			if len(qb.allowedSorts) > 0 && !qb.allowedSorts[field] {
				continue
			}
			params.Sort = append(params.Sort, SortField{Field: field, Desc: desc})
		}
	}

	// Legacy sorting support
	if sortBy := query.Get("sort_by"); sortBy != "" {
		params.SortBy = sortBy
		params.SortOrder = "asc"
		if order := query.Get("sort_order"); order == "desc" {
			params.SortOrder = "desc"
		}
		// Add to Sort array if allowed
		if len(qb.allowedSorts) == 0 || qb.allowedSorts[sortBy] {
			params.Sort = append(params.Sort, SortField{
				Field: sortBy,
				Desc:  params.SortOrder == "desc",
			})
		}
	}

	// Parse search
	params.Search = query.Get("search")
	if params.Search == "" {
		params.Search = query.Get("q")
	}

	// Parse filters
	for key, values := range query {
		if len(values) == 0 {
			continue
		}
		// Skip pagination, sorting, and search parameters
		if key == "page" || key == "page_size" || key == "limit" || key == "offset" ||
			key == "sort" || key == "sort_by" || key == "sort_order" ||
			key == "search" || key == "q" || key == "fields" || key == "include" {
			continue
		}
		// Check if filter is allowed
		if len(qb.allowedFilters) > 0 && !qb.allowedFilters[key] {
			continue
		}
		params.Filters[key] = values[0]
	}

	// Parse field selection
	if fieldsStr := query.Get("fields"); fieldsStr != "" {
		fields := strings.Split(fieldsStr, ",")
		for _, field := range fields {
			field = strings.TrimSpace(field)
			if field != "" {
				params.Fields = append(params.Fields, field)
			}
		}
	}

	// Parse includes
	if includeStr := query.Get("include"); includeStr != "" {
		includes := strings.Split(includeStr, ",")
		for _, inc := range includes {
			inc = strings.TrimSpace(inc)
			if inc != "" {
				params.Include = append(params.Include, inc)
			}
		}
	}

	return params
}

// ================================================
// Enhanced CRUD Framework - Pagination
// ================================================

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       any             `json:"data"`
	Pagination *PaginationMeta `json:"pagination,omitempty"`
}

// NewPaginationMeta creates pagination metadata
func NewPaginationMeta(page, pageSize int, totalItems int64) *PaginationMeta {
	totalPages := int((totalItems + int64(pageSize) - 1) / int64(pageSize))
	if totalPages < 0 {
		totalPages = 0
	}
	return &PaginationMeta{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// ================================================
// Enhanced CRUD Framework - Resource Hooks
// ================================================

// ResourceHooks defines lifecycle hooks for resource operations
type ResourceHooks interface {
	// BeforeCreate is called before creating a resource
	BeforeCreate(ctx context.Context, data any) error

	// AfterCreate is called after creating a resource
	AfterCreate(ctx context.Context, data any) error

	// BeforeUpdate is called before updating a resource
	BeforeUpdate(ctx context.Context, id string, data any) error

	// AfterUpdate is called after updating a resource
	AfterUpdate(ctx context.Context, id string, data any) error

	// BeforeDelete is called before deleting a resource
	BeforeDelete(ctx context.Context, id string) error

	// AfterDelete is called after deleting a resource
	AfterDelete(ctx context.Context, id string) error

	// BeforeList is called before listing resources
	BeforeList(ctx context.Context, params *QueryParams) error

	// AfterList is called after listing resources
	AfterList(ctx context.Context, params *QueryParams, data any) error
}

// NoOpResourceHooks provides a no-op implementation of ResourceHooks
type NoOpResourceHooks struct{}

func (h *NoOpResourceHooks) BeforeCreate(ctx context.Context, data any) error         { return nil }
func (h *NoOpResourceHooks) AfterCreate(ctx context.Context, data any) error          { return nil }
func (h *NoOpResourceHooks) BeforeUpdate(ctx context.Context, id string, data any) error {
	return nil
}
func (h *NoOpResourceHooks) AfterUpdate(ctx context.Context, id string, data any) error {
	return nil
}
func (h *NoOpResourceHooks) BeforeDelete(ctx context.Context, id string) error      { return nil }
func (h *NoOpResourceHooks) AfterDelete(ctx context.Context, id string) error       { return nil }
func (h *NoOpResourceHooks) BeforeList(ctx context.Context, params *QueryParams) error { return nil }
func (h *NoOpResourceHooks) AfterList(ctx context.Context, params *QueryParams, data any) error {
	return nil
}

// ================================================
// Enhanced CRUD Framework - Parameter Extraction
// ================================================

// ParamExtractor provides utilities for extracting parameters from requests
type ParamExtractor struct{}

// NewParamExtractor creates a new ParamExtractor
func NewParamExtractor() *ParamExtractor {
	return &ParamExtractor{}
}

// GetID extracts the ID parameter from the request context
func (pe *ParamExtractor) GetID(r *http.Request) string {
	reqCtx := contract.RequestContextFrom(r.Context())
	return reqCtx.Params["id"]
}

// GetParam extracts a named parameter from the request context
func (pe *ParamExtractor) GetParam(r *http.Request, name string) string {
	reqCtx := contract.RequestContextFrom(r.Context())
	return reqCtx.Params[name]
}

// GetQueryParam extracts a query parameter from the request
func (pe *ParamExtractor) GetQueryParam(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

// GetQueryInt extracts an integer query parameter
func (pe *ParamExtractor) GetQueryInt(r *http.Request, name string, defaultValue int) int {
	if str := r.URL.Query().Get(name); str != "" {
		if val, err := strconv.Atoi(str); err == nil {
			return val
		}
	}
	return defaultValue
}

// GetQueryBool extracts a boolean query parameter
func (pe *ParamExtractor) GetQueryBool(r *http.Request, name string, defaultValue bool) bool {
	if str := r.URL.Query().Get(name); str != "" {
		if val, err := strconv.ParseBool(str); err == nil {
			return val
		}
	}
	return defaultValue
}

// ================================================
// Enhanced CRUD Framework - Resource Transformer
// ================================================

// ResourceTransformer defines an interface for transforming resources
type ResourceTransformer interface {
	// Transform transforms a single resource for API response
	Transform(ctx context.Context, resource any) (any, error)

	// TransformCollection transforms a collection of resources
	TransformCollection(ctx context.Context, resources any) (any, error)
}

// IdentityTransformer is a pass-through transformer that returns resources unchanged
type IdentityTransformer struct{}

func (t *IdentityTransformer) Transform(ctx context.Context, resource any) (any, error) {
	return resource, nil
}

func (t *IdentityTransformer) TransformCollection(ctx context.Context, resources any) (any, error) {
	return resources, nil
}

// ================================================
// Enhanced CRUD Framework - Context-Aware Controllers
// ================================================

// ContextResourceController is a context-aware version of ResourceController
// that uses contract.Ctx for enhanced request handling
type ContextResourceController interface {
	// Basic CRUD operations with context
	IndexCtx(ctx *contract.Ctx)  // GET /resource
	ShowCtx(ctx *contract.Ctx)   // GET /resource/:id
	CreateCtx(ctx *contract.Ctx) // POST /resource
	UpdateCtx(ctx *contract.Ctx) // PUT /resource/:id
	DeleteCtx(ctx *contract.Ctx) // DELETE /resource/:id
	PatchCtx(ctx *contract.Ctx)  // PATCH /resource/:id

	// Batch operations
	BatchCreateCtx(ctx *contract.Ctx) // POST /resource/batch
	BatchDeleteCtx(ctx *contract.Ctx) // DELETE /resource/batch
}

// BaseContextResourceController provides a default implementation with context support
type BaseContextResourceController struct {
	ResourceName string
	QueryBuilder *QueryBuilder
	ParamExtractor *ParamExtractor
	Hooks        ResourceHooks
	Transformer  ResourceTransformer
}

// NewBaseContextResourceController creates a new context-aware resource controller
func NewBaseContextResourceController(resourceName string) *BaseContextResourceController {
	return &BaseContextResourceController{
		ResourceName:   resourceName,
		QueryBuilder:   NewQueryBuilder(),
		ParamExtractor: NewParamExtractor(),
		Hooks:          &NoOpResourceHooks{},
		Transformer:    &IdentityTransformer{},
	}
}

// IndexCtx handles GET /resource requests with context
func (c *BaseContextResourceController) IndexCtx(ctx *contract.Ctx) {
	ctx.JSON(http.StatusNotImplemented, map[string]string{
		"error":   "Not Implemented",
		"message": fmt.Sprintf("The Index method is not implemented for the %s resource", c.ResourceName),
	})
}

// ShowCtx handles GET /resource/:id requests with context
func (c *BaseContextResourceController) ShowCtx(ctx *contract.Ctx) {
	ctx.JSON(http.StatusNotImplemented, map[string]string{
		"error":   "Not Implemented",
		"message": fmt.Sprintf("The Show method is not implemented for the %s resource", c.ResourceName),
	})
}

// CreateCtx handles POST /resource requests with context
func (c *BaseContextResourceController) CreateCtx(ctx *contract.Ctx) {
	ctx.JSON(http.StatusNotImplemented, map[string]string{
		"error":   "Not Implemented",
		"message": fmt.Sprintf("The Create method is not implemented for the %s resource", c.ResourceName),
	})
}

// UpdateCtx handles PUT /resource/:id requests with context
func (c *BaseContextResourceController) UpdateCtx(ctx *contract.Ctx) {
	ctx.JSON(http.StatusNotImplemented, map[string]string{
		"error":   "Not Implemented",
		"message": fmt.Sprintf("The Update method is not implemented for the %s resource", c.ResourceName),
	})
}

// DeleteCtx handles DELETE /resource/:id requests with context
func (c *BaseContextResourceController) DeleteCtx(ctx *contract.Ctx) {
	ctx.JSON(http.StatusNotImplemented, map[string]string{
		"error":   "Not Implemented",
		"message": fmt.Sprintf("The Delete method is not implemented for the %s resource", c.ResourceName),
	})
}

// PatchCtx handles PATCH /resource/:id requests with context
func (c *BaseContextResourceController) PatchCtx(ctx *contract.Ctx) {
	ctx.JSON(http.StatusNotImplemented, map[string]string{
		"error":   "Not Implemented",
		"message": fmt.Sprintf("The Patch method is not implemented for the %s resource", c.ResourceName),
	})
}

// BatchCreateCtx handles batch creation of resources with context
func (c *BaseContextResourceController) BatchCreateCtx(ctx *contract.Ctx) {
	ctx.JSON(http.StatusNotImplemented, map[string]string{
		"error":   "Not Implemented",
		"message": fmt.Sprintf("The BatchCreate method is not implemented for the %s resource", c.ResourceName),
	})
}

// BatchDeleteCtx handles batch deletion of resources with context
func (c *BaseContextResourceController) BatchDeleteCtx(ctx *contract.Ctx) {
	ctx.JSON(http.StatusNotImplemented, map[string]string{
		"error":   "Not Implemented",
		"message": fmt.Sprintf("The BatchDelete method is not implemented for the %s resource", c.ResourceName),
	})
}

// ================================================
// Enhanced CRUD Framework - Resource Options
// ================================================

// ResourceOptions configures resource behavior
type ResourceOptions struct {
	// DefaultPageSize is the default number of items per page
	DefaultPageSize int

	// MaxPageSize is the maximum number of items per page
	MaxPageSize int

	// AllowedSorts are the fields that can be used for sorting
	AllowedSorts []string

	// AllowedFilters are the fields that can be used for filtering
	AllowedFilters []string

	// SoftDelete enables soft delete functionality
	SoftDelete bool

	// SoftDeleteField is the field name for soft delete timestamp
	SoftDeleteField string

	// EnableHooks enables lifecycle hooks
	EnableHooks bool

	// EnableTransformer enables resource transformation
	EnableTransformer bool
}

// DefaultResourceOptions returns default resource options
func DefaultResourceOptions() *ResourceOptions {
	return &ResourceOptions{
		DefaultPageSize: 20,
		MaxPageSize:     100,
		AllowedSorts:    []string{},
		AllowedFilters:  []string{},
		SoftDelete:      false,
		SoftDeleteField: "deleted_at",
		EnableHooks:     true,
		EnableTransformer: false,
	}
}

// ================================================
// Enhanced CRUD Framework - Batch Processor
// ================================================

// BatchResult represents the result of a batch operation
type BatchResult struct {
	Successful int                    `json:"successful"`
	Failed     int                    `json:"failed"`
	Total      int                    `json:"total"`
	Errors     []BatchError           `json:"errors,omitempty"`
	Results    []map[string]any       `json:"results,omitempty"`
}

// BatchError represents an error in a batch operation
type BatchError struct {
	Index   int    `json:"index"`
	ID      string `json:"id,omitempty"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// BatchProcessor handles batch operations
type BatchProcessor struct {
	maxBatchSize int
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(maxBatchSize int) *BatchProcessor {
	if maxBatchSize <= 0 {
		maxBatchSize = 100
	}
	return &BatchProcessor{
		maxBatchSize: maxBatchSize,
	}
}

// Process processes a batch operation
func (bp *BatchProcessor) Process(
	ctx context.Context,
	items []any,
	fn func(ctx context.Context, item any) error,
) *BatchResult {
	result := &BatchResult{
		Total:   len(items),
		Errors:  []BatchError{},
		Results: []map[string]any{},
	}

	if len(items) > bp.maxBatchSize {
		result.Failed = len(items)
		result.Errors = append(result.Errors, BatchError{
			Message: fmt.Sprintf("Batch size exceeds maximum of %d", bp.maxBatchSize),
			Code:    "BATCH_TOO_LARGE",
		})
		return result
	}

	for i, item := range items {
		if err := fn(ctx, item); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, BatchError{
				Index:   i,
				Message: err.Error(),
				Code:    "PROCESSING_ERROR",
			})
		} else {
			result.Successful++
		}
	}

	return result
}
