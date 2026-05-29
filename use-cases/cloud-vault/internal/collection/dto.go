package collection

// CreateCollectionRequest is the body for POST /api/v1/collections.
type CreateCollectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

// UpdateCollectionRequest is the body for PUT /api/v1/collections/:id.
type UpdateCollectionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AddDocumentRequest is the body for POST /api/v1/collections/:id/documents.
type AddDocumentRequest struct {
	DocumentID string `json:"document_id"`
	Note       string `json:"note"`
}

// ReorderRequest is the body for PUT /api/v1/collections/:id/documents/reorder.
type ReorderRequest struct {
	DocumentIDs []string `json:"document_ids"`
}

// CreateFromSearchRequest is the body for POST /api/v1/collections/from-search.
type CreateFromSearchRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	DocumentIDs []string `json:"document_ids"`
}

// CollectionListResult is the paginated result for GET /api/v1/collections.
type CollectionListResult struct {
	Items  []Collection `json:"items"`
	Total  int          `json:"total"`
}

// CollectionDetailResult includes documents.
type CollectionDetailResult struct {
	Collection Collection           `json:"collection"`
	Documents  []CollectionDocument `json:"documents"`
	Total      int                  `json:"total"`
}
