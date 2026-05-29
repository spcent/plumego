package collection

import (
	"context"
	"errors"
	"strings"
)

// Service handles collection business logic.
type Service struct {
	repo *Repository
}

// NewService constructs a Service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Create creates a new collection.
func (s *Service) Create(ctx context.Context, req CreateCollectionRequest) (*Collection, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("name is required")
	}
	collType := req.Type
	if collType == "" {
		collType = "manual"
	}
	c := &Collection{
		ID:          newID(),
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Type:        collType,
		Status:      "active",
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Update updates a collection's name and description.
func (s *Service) Update(ctx context.Context, id string, req UpdateCollectionRequest) (*Collection, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("name is required")
	}
	if err := s.repo.Update(ctx, id, strings.TrimSpace(req.Name), strings.TrimSpace(req.Description)); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

// Delete removes a collection and all its document associations.
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// GetByID returns a collection with its documents.
func (s *Service) GetByID(ctx context.Context, id string) (*CollectionDetailResult, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	docs, err := s.repo.ListDocuments(ctx, id)
	if err != nil {
		return nil, err
	}
	return &CollectionDetailResult{
		Collection: *c,
		Documents:  docs,
		Total:      len(docs),
	}, nil
}

// List returns all active collections.
func (s *Service) List(ctx context.Context) (*CollectionListResult, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	return &CollectionListResult{Items: items, Total: len(items)}, nil
}

// AddDocument adds a document to a collection.
func (s *Service) AddDocument(ctx context.Context, collectionID string, req AddDocumentRequest) error {
	if req.DocumentID == "" {
		return errors.New("document_id is required")
	}
	return s.repo.AddDocument(ctx, collectionID, req.DocumentID, req.Note)
}

// RemoveDocument removes a document from a collection.
func (s *Service) RemoveDocument(ctx context.Context, collectionID, documentID string) error {
	return s.repo.RemoveDocument(ctx, collectionID, documentID)
}

// Reorder reorders documents in a collection.
func (s *Service) Reorder(ctx context.Context, collectionID string, req ReorderRequest) error {
	return s.repo.Reorder(ctx, collectionID, req.DocumentIDs)
}

// CreateFromSearch creates a collection from a list of selected document IDs.
func (s *Service) CreateFromSearch(ctx context.Context, req CreateFromSearchRequest) (*Collection, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("name is required")
	}
	c := &Collection{
		ID:          newID(),
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Type:        "search",
		Status:      "active",
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	for _, docID := range req.DocumentIDs {
		_ = s.repo.AddDocument(ctx, c.ID, docID, "")
	}
	return c, nil
}
