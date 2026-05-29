package tag

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Service handles tag and document-tag business logic.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTag(ctx context.Context, req CreateTagRequest) (*TagResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("tag name must not be empty")
	}
	t := &Tag{
		ID:        newID(),
		Name:      name,
		Color:     req.Color,
		Source:    req.Source,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	resp := toResponse(t)
	return &resp, nil
}

func (s *Service) ListTags(ctx context.Context) (*ListTagsResponse, error) {
	tags, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]TagResponse, 0, len(tags))
	for _, t := range tags {
		items = append(items, toResponse(t))
	}
	return &ListTagsResponse{Items: items, Total: len(items)}, nil
}

func (s *Service) UpdateTag(ctx context.Context, id string, req UpdateTagRequest) (*TagResponse, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if name := strings.TrimSpace(req.Name); name != "" {
		t.Name = name
	}
	if req.Color != "" {
		t.Color = req.Color
	}
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, err
	}
	resp := toResponse(t)
	return &resp, nil
}

func (s *Service) DeleteTag(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) GetDocumentTags(ctx context.Context, docID string) (*ListTagsResponse, error) {
	tags, err := s.repo.GetDocumentTags(ctx, docID)
	if err != nil {
		return nil, err
	}
	items := make([]TagResponse, 0, len(tags))
	for _, t := range tags {
		items = append(items, toResponse(t))
	}
	return &ListTagsResponse{Items: items, Total: len(items)}, nil
}

func (s *Service) SetDocumentTags(ctx context.Context, docID string, req SetDocumentTagsRequest) error {
	return s.repo.SetDocumentTags(ctx, docID, req.TagIDs)
}

func (s *Service) RemoveDocumentTag(ctx context.Context, docID, tagID string) error {
	return s.repo.RemoveDocumentTag(ctx, docID, tagID)
}
