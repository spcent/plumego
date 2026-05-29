package tag

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeRepo is an in-memory Repository for service tests.
type fakeRepo struct {
	tags    map[string]*Tag
	docTags map[string][]string // docID → []tagID
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		tags:    make(map[string]*Tag),
		docTags: make(map[string][]string),
	}
}

func (r *fakeRepo) Create(_ context.Context, t *Tag) error {
	for _, existing := range r.tags {
		if existing.Name == t.Name {
			return ErrDuplicate
		}
	}
	r.tags[t.ID] = t
	return nil
}

func (r *fakeRepo) GetByID(_ context.Context, id string) (*Tag, error) {
	t, ok := r.tags[id]
	if !ok {
		return nil, ErrNotFound
	}
	return t, nil
}

func (r *fakeRepo) List(_ context.Context) ([]*Tag, error) {
	out := make([]*Tag, 0, len(r.tags))
	for _, t := range r.tags {
		out = append(out, t)
	}
	return out, nil
}

func (r *fakeRepo) Update(_ context.Context, t *Tag) error {
	if _, ok := r.tags[t.ID]; !ok {
		return ErrNotFound
	}
	r.tags[t.ID] = t
	return nil
}

func (r *fakeRepo) Delete(_ context.Context, id string) error {
	if _, ok := r.tags[id]; !ok {
		return ErrNotFound
	}
	delete(r.tags, id)
	delete(r.docTags, id)
	return nil
}

func (r *fakeRepo) GetDocumentTags(_ context.Context, docID string) ([]*Tag, error) {
	ids := r.docTags[docID]
	out := make([]*Tag, 0, len(ids))
	for _, id := range ids {
		if t, ok := r.tags[id]; ok {
			out = append(out, t)
		}
	}
	return out, nil
}

func (r *fakeRepo) SetDocumentTags(_ context.Context, docID string, tagIDs []string) error {
	r.docTags[docID] = tagIDs
	return nil
}

func (r *fakeRepo) RemoveDocumentTag(_ context.Context, docID, tagID string) error {
	ids := r.docTags[docID]
	updated := ids[:0]
	for _, id := range ids {
		if id != tagID {
			updated = append(updated, id)
		}
	}
	r.docTags[docID] = updated
	return nil
}

func (r *fakeRepo) EnsureTag(_ context.Context, name, color, source string) (*Tag, error) {
	for _, t := range r.tags {
		if t.Name == name {
			return t, nil
		}
	}
	t := &Tag{ID: newID(), Name: name, Color: color, Source: source, CreatedAt: time.Now()}
	r.tags[t.ID] = t
	return t, nil
}

// --- tests ---

func TestCreateTag_ok(t *testing.T) {
	svc := NewService(newFakeRepo())
	resp, err := svc.CreateTag(context.Background(), CreateTagRequest{Name: "go"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Name != "go" {
		t.Errorf("name: want 'go', got %q", resp.Name)
	}
	if resp.ID == "" {
		t.Error("id must not be empty")
	}
}

func TestCreateTag_emptyName(t *testing.T) {
	svc := NewService(newFakeRepo())
	_, err := svc.CreateTag(context.Background(), CreateTagRequest{Name: "  "})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestCreateTag_withColor(t *testing.T) {
	svc := NewService(newFakeRepo())
	resp, err := svc.CreateTag(context.Background(), CreateTagRequest{Name: "urgent", Color: "#ff0000"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Color != "#ff0000" {
		t.Errorf("color: want '#ff0000', got %q", resp.Color)
	}
}

func TestListTags(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo)
	svc.CreateTag(context.Background(), CreateTagRequest{Name: "a"})
	svc.CreateTag(context.Background(), CreateTagRequest{Name: "b"})

	list, err := svc.ListTags(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if list.Total != 2 || len(list.Items) != 2 {
		t.Errorf("total/items: want 2/2, got %d/%d", list.Total, len(list.Items))
	}
}

func TestUpdateTag_nameAndColor(t *testing.T) {
	svc := NewService(newFakeRepo())
	resp, _ := svc.CreateTag(context.Background(), CreateTagRequest{Name: "old"})

	updated, err := svc.UpdateTag(context.Background(), resp.ID, UpdateTagRequest{Name: "new", Color: "#abc"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "new" {
		t.Errorf("name: want 'new', got %q", updated.Name)
	}
	if updated.Color != "#abc" {
		t.Errorf("color: want '#abc', got %q", updated.Color)
	}
}

func TestUpdateTag_emptyNamePreservesOld(t *testing.T) {
	svc := NewService(newFakeRepo())
	resp, _ := svc.CreateTag(context.Background(), CreateTagRequest{Name: "keep"})

	updated, err := svc.UpdateTag(context.Background(), resp.ID, UpdateTagRequest{Name: ""})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "keep" {
		t.Errorf("name should be preserved: got %q", updated.Name)
	}
}

func TestUpdateTag_notFound(t *testing.T) {
	svc := NewService(newFakeRepo())
	_, err := svc.UpdateTag(context.Background(), "no-such-id", UpdateTagRequest{Name: "x"})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestDeleteTag(t *testing.T) {
	svc := NewService(newFakeRepo())
	resp, _ := svc.CreateTag(context.Background(), CreateTagRequest{Name: "tmp"})

	if err := svc.DeleteTag(context.Background(), resp.ID); err != nil {
		t.Fatal(err)
	}
	list, _ := svc.ListTags(context.Background())
	if list.Total != 0 {
		t.Errorf("tag should be deleted, got %d", list.Total)
	}
}

func TestSetAndGetDocumentTags(t *testing.T) {
	svc := NewService(newFakeRepo())
	a, _ := svc.CreateTag(context.Background(), CreateTagRequest{Name: "a"})
	b, _ := svc.CreateTag(context.Background(), CreateTagRequest{Name: "b"})

	err := svc.SetDocumentTags(context.Background(), "doc-1", SetDocumentTagsRequest{
		TagIDs: []string{a.ID, b.ID},
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := svc.GetDocumentTags(context.Background(), "doc-1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 2 {
		t.Errorf("want 2 doc tags, got %d", result.Total)
	}
}

func TestRemoveDocumentTag(t *testing.T) {
	svc := NewService(newFakeRepo())
	a, _ := svc.CreateTag(context.Background(), CreateTagRequest{Name: "a"})
	b, _ := svc.CreateTag(context.Background(), CreateTagRequest{Name: "b"})

	svc.SetDocumentTags(context.Background(), "doc-1", SetDocumentTagsRequest{TagIDs: []string{a.ID, b.ID}})
	svc.RemoveDocumentTag(context.Background(), "doc-1", a.ID)

	result, _ := svc.GetDocumentTags(context.Background(), "doc-1")
	if result.Total != 1 || result.Items[0].ID != b.ID {
		t.Errorf("expected only tag b remaining, got %+v", result.Items)
	}
}
