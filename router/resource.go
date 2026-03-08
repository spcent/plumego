package router

// This file re-exports REST resource helpers from the rest package as type aliases
// so that existing code continues to compile unchanged.
//
// All types here are deprecated for use via this package.
// The canonical location is github.com/spcent/plumego/rest.
//
// Migration: replace import "github.com/spcent/plumego/router" with
//            import "github.com/spcent/plumego/rest" and update type names.

import "github.com/spcent/plumego/rest"

// Deprecated: use rest.Response instead.
type Response = rest.Response

// Deprecated: use rest.ErrorResponse instead.
type ErrorResponse = rest.ErrorResponse

// Deprecated: use rest.JSONWriter instead. Prefer contract.WriteResponse / contract.WriteError for new code.
type JSONWriter = rest.JSONWriter

// NewJSONWriter creates a new JSONWriter.
//
// Deprecated: prefer contract.WriteResponse / contract.WriteError for new code.
func NewJSONWriter(resourceName string) *JSONWriter { return rest.NewJSONWriter(resourceName) }

// ResourceController defines the interface for RESTful resource controllers.
//
// Deprecated: use rest.ResourceController instead.
type ResourceController = rest.ResourceController

// BaseResourceController provides a default implementation of ResourceController.
//
// Deprecated: use rest.BaseResourceController instead.
type BaseResourceController = rest.BaseResourceController

// NewBaseResourceController creates a new BaseResourceController.
//
// Deprecated: use rest.NewBaseResourceController instead.
func NewBaseResourceController(resourceName string) *BaseResourceController {
	return rest.NewBaseResourceController(resourceName)
}

// QueryParams represents parsed query parameters for list operations.
//
// Deprecated: use rest.QueryParams instead.
type QueryParams = rest.QueryParams

// SortField represents a sort specification.
//
// Deprecated: use rest.SortField instead.
type SortField = rest.SortField

// QueryBuilder builds QueryParams from an *http.Request.
//
// Deprecated: use rest.QueryBuilder instead.
type QueryBuilder = rest.QueryBuilder

// NewQueryBuilder creates a QueryBuilder with sensible defaults.
//
// Deprecated: use rest.NewQueryBuilder instead.
func NewQueryBuilder() *QueryBuilder { return rest.NewQueryBuilder() }

// PaginationMeta holds pagination metadata for list responses.
//
// Deprecated: use rest.PaginationMeta instead.
type PaginationMeta = rest.PaginationMeta

// NewPaginationMeta builds PaginationMeta from page, page size, and total item count.
//
// Deprecated: use rest.NewPaginationMeta instead.
func NewPaginationMeta(page, pageSize int, totalItems int64) *PaginationMeta {
	return rest.NewPaginationMeta(page, pageSize, totalItems)
}

// PaginatedResponse is the canonical envelope for paginated list responses.
//
// Deprecated: use rest.PaginatedResponse instead.
type PaginatedResponse = rest.PaginatedResponse

// ResourceHooks defines lifecycle callbacks for resource CRUD operations.
//
// Deprecated: use rest.ResourceHooks instead.
type ResourceHooks = rest.ResourceHooks

// NoOpResourceHooks is a no-op implementation of ResourceHooks.
//
// Deprecated: use rest.NoOpResourceHooks instead.
type NoOpResourceHooks = rest.NoOpResourceHooks

// ResourceTransformer transforms resource entities before they are serialized.
//
// Deprecated: use rest.ResourceTransformer instead.
type ResourceTransformer = rest.ResourceTransformer

// IdentityTransformer returns resources unchanged (pass-through).
//
// Deprecated: use rest.IdentityTransformer instead.
type IdentityTransformer = rest.IdentityTransformer

// ParamExtractor provides utilities for extracting path and query parameters from requests.
//
// Deprecated: use rest.ParamExtractor instead.
type ParamExtractor = rest.ParamExtractor

// NewParamExtractor creates a new ParamExtractor.
//
// Deprecated: use rest.NewParamExtractor instead.
func NewParamExtractor() *ParamExtractor { return rest.NewParamExtractor() }

// ContextResourceController is a context-aware version of ResourceController.
//
// Deprecated: use rest.ContextResourceController instead.
type ContextResourceController = rest.ContextResourceController

// BaseContextResourceController provides a default context-aware resource controller.
//
// Deprecated: use rest.BaseContextResourceController instead.
type BaseContextResourceController = rest.BaseContextResourceController

// NewBaseContextResourceController creates a context-aware resource controller with defaults.
//
// Deprecated: use rest.NewBaseContextResourceController instead.
func NewBaseContextResourceController(resourceName string) *BaseContextResourceController {
	return rest.NewBaseContextResourceController(resourceName)
}

// ResourceOptions configures resource controller behavior.
//
// Deprecated: use rest.ResourceOptions instead.
type ResourceOptions = rest.ResourceOptions

// DefaultResourceOptions returns sensible defaults.
//
// Deprecated: use rest.DefaultResourceOptions instead.
func DefaultResourceOptions() *ResourceOptions { return rest.DefaultResourceOptions() }

// BatchResult is the result of a batch operation.
//
// Deprecated: use rest.BatchResult instead.
type BatchResult = rest.BatchResult

// BatchError describes an individual failure within a batch operation.
//
// Deprecated: use rest.BatchError instead.
type BatchError = rest.BatchError

// BatchProcessor handles batch operations with a configurable item limit.
//
// Deprecated: use rest.BatchProcessor instead.
type BatchProcessor = rest.BatchProcessor

// NewBatchProcessor creates a BatchProcessor. maxBatchSize ≤ 0 defaults to 100.
//
// Deprecated: use rest.NewBatchProcessor instead.
func NewBatchProcessor(maxBatchSize int) *BatchProcessor {
	return rest.NewBatchProcessor(maxBatchSize)
}
