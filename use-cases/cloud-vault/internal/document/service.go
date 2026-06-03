package document

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"cloud-vault/internal/idgen"
	"cloud-vault/internal/markdown"
	"cloud-vault/internal/storage"
)

// IndexEvent carries information about a document change for the search indexer.
// Defined here to avoid an import cycle: document → search would be cyclic if
// search defines it and document references it.
type IndexEvent struct {
	DocID    string
	Content  string
	Version  int
	Hash     string
	Deleted  bool
	IsImport bool
}

// IndexHookFn is called after a successful create, update, or delete operation.
// It must not block the caller; implementations should handle errors internally.
type IndexHookFn func(ctx context.Context, event IndexEvent)

const (
	VersionPolicyBounded   = "bounded"
	VersionPolicyFull      = "full"
	VersionPolicyOverwrite = "overwrite"
)

// VersioningConfig controls how object-storage snapshots are retained.
type VersioningConfig struct {
	Policy     string
	KeepLatest int
}

// Service orchestrates document creation, retrieval, and versioning.
type Service struct {
	repo       Repository
	storage    storage.ObjectStorage
	versioning VersioningConfig
	indexHook  IndexHookFn
}

// NewService constructs a Service with the given repository and storage backend.
func NewService(repo Repository, store storage.ObjectStorage) *Service {
	return NewServiceWithVersioning(repo, store, VersioningConfig{})
}

// NewServiceWithVersioning constructs a Service with explicit version retention.
func NewServiceWithVersioning(repo Repository, store storage.ObjectStorage, versioning VersioningConfig) *Service {
	return &Service{repo: repo, storage: store, versioning: normalizeVersioningConfig(versioning)}
}

// SetIndexHook registers a callback invoked after every document mutation.
func (s *Service) SetIndexHook(fn IndexHookFn) {
	s.indexHook = fn
}

func (s *Service) fireIndexHook(ctx context.Context, event IndexEvent) {
	if s.indexHook != nil {
		s.indexHook(ctx, event)
	}
}

// Create creates a new document via the editor.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*SaveResult, error) {
	content := req.Content
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = markdown.ExtractTitle(content)
	}
	if title == "" {
		title = "Untitled"
	}

	meta := markdown.Parse(content)
	hash := sha256Content(content)
	now := time.Now().UTC()
	docID := idgen.New()

	currentKey := CurrentKey(docID)

	contentBytes := []byte(content)
	size := int64(len(contentBytes))

	if err := s.storage.Put(ctx, currentKey, bytes.NewReader(contentBytes), size, "text/markdown"); err != nil {
		return nil, fmt.Errorf("upload current: %w", err)
	}

	uploadedAt := now
	doc := &Document{
		ID:             docID,
		Title:          title,
		StorageKey:     currentKey,
		CurrentVersion: 1,
		ContentHash:    hash,
		SizeBytes:      size,
		WordCount:      meta.WordCount,
		LineCount:      meta.LineCount,
		Status:         StatusActive,
		SyncStatus:     SyncStatusSynced,
		SourceType:     SourceTypeManual,
		ReviewStatus:   ReviewStatusPending,
		Summary:        meta.Summary,
		HeadingText:    meta.HeadingText,
		CreatedAt:      now,
		UpdatedAt:      now,
		UploadedAt:     &uploadedAt,
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		return nil, fmt.Errorf("persist document: %w", err)
	}
	if s.shouldCreateAutoSnapshot() {
		if _, err := s.createVersionSnapshot(ctx, doc, contentBytes, now, VersionKindAuto, "", false); err != nil {
			return nil, err
		}
	}

	s.fireIndexHook(ctx, IndexEvent{
		DocID: docID, Content: content, Version: 1, Hash: hash, IsImport: false,
	})

	return &SaveResult{
		ID:        docID,
		Title:     title,
		Version:   1,
		Changed:   true,
		UpdatedAt: now,
	}, nil
}

// CreateFromImport creates a document with import provenance metadata.
func (s *Service) CreateFromImport(ctx context.Context, req CreateFromImportRequest) (*SaveResult, error) {
	content := req.Content
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = markdown.ExtractTitle(content)
	}
	if title == "" {
		title = "Untitled"
	}

	meta := markdown.Parse(content)
	hash := sha256Content(content)
	now := time.Now().UTC()
	docID := idgen.New()

	currentKey := CurrentKey(docID)

	contentBytes := []byte(content)
	size := int64(len(contentBytes))

	if err := s.storage.Put(ctx, currentKey, bytes.NewReader(contentBytes), size, "text/markdown"); err != nil {
		return nil, fmt.Errorf("upload current: %w", err)
	}

	uploadedAt := now
	importedAt := now
	doc := &Document{
		ID:             docID,
		Title:          title,
		OriginalPath:   req.OriginalPath,
		StorageKey:     currentKey,
		CurrentVersion: 1,
		ContentHash:    hash,
		SizeBytes:      size,
		WordCount:      meta.WordCount,
		LineCount:      meta.LineCount,
		Status:         StatusActive,
		SyncStatus:     SyncStatusSynced,
		SourceType:     SourceTypeImported,
		ImportJobID:    req.ImportJobID,
		ReviewStatus:   ReviewStatusPending,
		Summary:        meta.Summary,
		HeadingText:    meta.HeadingText,
		CreatedAt:      now,
		UpdatedAt:      now,
		UploadedAt:     &uploadedAt,
		ImportedAt:     &importedAt,
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		return nil, fmt.Errorf("persist document: %w", err)
	}
	if s.shouldCreateAutoSnapshot() {
		if _, err := s.createVersionSnapshot(ctx, doc, contentBytes, now, VersionKindAuto, "", false); err != nil {
			return nil, err
		}
	}

	s.fireIndexHook(ctx, IndexEvent{
		DocID: docID, Content: content, Version: 1, Hash: hash, IsImport: true,
	})

	return &SaveResult{
		ID:        docID,
		Title:     title,
		Version:   1,
		Changed:   true,
		UpdatedAt: now,
	}, nil
}

// Get returns the document metadata and its current Markdown content.
func (s *Service) Get(ctx context.Context, id string) (*Detail, error) {
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	rc, err := s.storage.Get(ctx, doc.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("fetch content: %w", err)
	}
	defer rc.Close()

	raw, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("read content: %w", err)
	}

	d := toDetail(doc, string(raw))
	return &d, nil
}

// Update saves a new version of a document, enforcing optimistic concurrency.
func (s *Service) Update(ctx context.Context, id string, req UpdateRequest) (*SaveResult, error) {
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.BaseVersion != doc.CurrentVersion {
		return nil, ErrVersionConflict
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = doc.Title
	}

	content := req.Content
	hash := sha256Content(content)

	if hash == doc.ContentHash && title == doc.Title {
		return &SaveResult{
			ID:        doc.ID,
			Title:     doc.Title,
			Version:   doc.CurrentVersion,
			Changed:   false,
			UpdatedAt: doc.UpdatedAt,
		}, nil
	}

	newVersion := doc.CurrentVersion + 1
	now := time.Now().UTC()

	contentBytes := []byte(content)
	size := int64(len(contentBytes))
	meta := markdown.Parse(content)

	currentKey := CurrentKey(id)

	if err := s.storage.Put(ctx, currentKey, bytes.NewReader(contentBytes), size, "text/markdown"); err != nil {
		return nil, fmt.Errorf("upload current: %w", err)
	}

	uploadedAt := now
	doc.Title = title
	doc.StorageKey = currentKey
	doc.CurrentVersion = newVersion
	doc.ContentHash = hash
	doc.SizeBytes = size
	doc.WordCount = meta.WordCount
	doc.LineCount = meta.LineCount
	doc.SyncStatus = SyncStatusSynced
	doc.Summary = meta.Summary
	doc.HeadingText = meta.HeadingText
	doc.UpdatedAt = now
	doc.UploadedAt = &uploadedAt

	if err := s.repo.Update(ctx, doc); err != nil {
		return nil, fmt.Errorf("persist update: %w", err)
	}

	if s.shouldCreateAutoSnapshot() {
		if _, err := s.createVersionSnapshot(ctx, doc, contentBytes, now, VersionKindAuto, "", false); err != nil {
			return nil, err
		}
		if err := s.pruneVersions(ctx, id); err != nil {
			return nil, err
		}
	}

	s.fireIndexHook(ctx, IndexEvent{
		DocID: id, Content: content, Version: newVersion, Hash: hash, IsImport: false,
	})

	return &SaveResult{
		ID:        id,
		Title:     title,
		Version:   newVersion,
		Changed:   true,
		UpdatedAt: now,
	}, nil
}

// Delete soft-deletes a document and fires the index hook.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return err
	}
	s.fireIndexHook(ctx, IndexEvent{DocID: id, Deleted: true})
	return nil
}

// List returns a paginated list of documents.
func (s *Service) List(ctx context.Context, q ListQuery) (*ListResult, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	q.Limit = limit

	docs, total, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, err
	}

	items := make([]Summary, 0, len(docs))
	for _, d := range docs {
		items = append(items, toSummary(d))
	}

	return &ListResult{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: q.Offset,
	}, nil
}

// ExistsByHash reports whether a non-deleted document with the given hash exists.
func (s *Service) ExistsByHash(ctx context.Context, hash string) (bool, error) {
	return s.repo.ExistsByHash(ctx, hash)
}

// UpdateFavorite toggles the favourite state of a document.
func (s *Service) UpdateFavorite(ctx context.Context, id string, favorite bool) error {
	return s.repo.UpdateFavorite(ctx, id, favorite)
}

// UpdateStatus archives or restores a document.
func (s *Service) UpdateStatus(ctx context.Context, id string, status string) error {
	if status != StatusActive && status != StatusArchived {
		return fmt.Errorf("invalid status %q: must be active or archived", status)
	}
	return s.repo.UpdateStatus(ctx, id, status)
}

// BatchUpdateStatus applies a status change to multiple documents.
func (s *Service) BatchUpdateStatus(ctx context.Context, ids []string, status string) error {
	if status != StatusActive && status != StatusArchived && status != StatusDeleted {
		return fmt.Errorf("invalid status %q", status)
	}
	return s.repo.BatchUpdateStatus(ctx, ids, status)
}

// UpdateReviewStatus updates the review status of a document.
func (s *Service) UpdateReviewStatus(ctx context.Context, id string, reviewStatus string) error {
	if reviewStatus != ReviewStatusPending && reviewStatus != ReviewStatusReviewed {
		return fmt.Errorf("invalid review_status %q", reviewStatus)
	}
	return s.repo.UpdateReviewStatus(ctx, id, reviewStatus)
}

// GetVersions returns all version summaries for a document.
func (s *Service) GetVersions(ctx context.Context, docID string) (*VersionListResult, error) {
	if _, err := s.repo.GetByID(ctx, docID); err != nil {
		return nil, err
	}

	versions, err := s.repo.GetVersions(ctx, docID)
	if err != nil {
		return nil, err
	}

	items := make([]VersionSummary, 0, len(versions))
	for _, v := range versions {
		items = append(items, VersionSummary{
			Version:   v.Version,
			SizeBytes: v.SizeBytes,
			CreatedAt: v.CreatedAt,
			Kind:      v.Kind,
			Pinned:    v.Pinned,
			Note:      v.Note,
		})
	}
	return &VersionListResult{Items: items}, nil
}

// GetVersion returns the full content for a specific document version.
func (s *Service) GetVersion(ctx context.Context, docID string, version int) (*VersionDetail, error) {
	if _, err := s.repo.GetByID(ctx, docID); err != nil {
		return nil, err
	}

	ver, err := s.repo.GetVersion(ctx, docID, version)
	if err != nil {
		return nil, err
	}

	rc, err := s.storage.Get(ctx, ver.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("fetch version content: %w", err)
	}
	defer rc.Close()

	raw, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("read version content: %w", err)
	}

	return &VersionDetail{
		ID:        ver.ID,
		Version:   ver.Version,
		Content:   string(raw),
		CreatedAt: ver.CreatedAt,
		Kind:      ver.Kind,
		Pinned:    ver.Pinned,
		Note:      ver.Note,
	}, nil
}

// CreateSnapshot pins the current document content as a key version.
func (s *Service) CreateSnapshot(ctx context.Context, docID, note string) (*VersionSummary, error) {
	doc, err := s.repo.GetByID(ctx, docID)
	if err != nil {
		return nil, err
	}

	if ver, err := s.repo.GetVersion(ctx, docID, doc.CurrentVersion); err == nil {
		if err := s.repo.UpdateVersionMetadata(ctx, docID, doc.CurrentVersion, VersionKindManual, strings.TrimSpace(note), true); err != nil {
			return nil, err
		}
		ver.Kind = VersionKindManual
		ver.Note = strings.TrimSpace(note)
		ver.Pinned = true
		return versionSummary(ver), nil
	} else if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	rc, err := s.storage.Get(ctx, doc.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("fetch current content: %w", err)
	}
	defer rc.Close()

	raw, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("read current content: %w", err)
	}

	ver, err := s.createVersionSnapshot(ctx, doc, raw, time.Now().UTC(), VersionKindManual, strings.TrimSpace(note), true)
	if err != nil {
		return nil, err
	}
	return versionSummary(ver), nil
}

// RestoreVersion restores a historical snapshot as a new current document version.
func (s *Service) RestoreVersion(ctx context.Context, docID string, version int) (*SaveResult, error) {
	doc, err := s.repo.GetByID(ctx, docID)
	if err != nil {
		return nil, err
	}
	if version == doc.CurrentVersion {
		return &SaveResult{
			ID:        doc.ID,
			Title:     doc.Title,
			Version:   doc.CurrentVersion,
			Changed:   false,
			UpdatedAt: doc.UpdatedAt,
		}, nil
	}

	target, err := s.repo.GetVersion(ctx, docID, version)
	if err != nil {
		return nil, err
	}

	rc, err := s.storage.Get(ctx, target.StorageKey)
	if err != nil {
		return nil, fmt.Errorf("fetch version content: %w", err)
	}
	targetBytes, err := io.ReadAll(rc)
	closeErr := rc.Close()
	if err != nil {
		return nil, fmt.Errorf("read version content: %w", err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close version content: %w", closeErr)
	}

	if err := s.ensureCurrentSnapshot(ctx, doc, version); err != nil {
		return nil, err
	}

	newVersion := doc.CurrentVersion + 1
	now := time.Now().UTC()
	content := string(targetBytes)
	hash := sha256Content(content)
	meta := markdown.Parse(content)
	title := markdown.ExtractTitle(content)
	if title == "" {
		title = doc.Title
	}

	currentKey := CurrentKey(docID)
	if err := s.storage.Put(ctx, currentKey, bytes.NewReader(targetBytes), int64(len(targetBytes)), "text/markdown"); err != nil {
		return nil, fmt.Errorf("upload restored current: %w", err)
	}

	uploadedAt := now
	doc.Title = title
	doc.StorageKey = currentKey
	doc.CurrentVersion = newVersion
	doc.ContentHash = hash
	doc.SizeBytes = int64(len(targetBytes))
	doc.WordCount = meta.WordCount
	doc.LineCount = meta.LineCount
	doc.SyncStatus = SyncStatusSynced
	doc.Summary = meta.Summary
	doc.HeadingText = meta.HeadingText
	doc.UpdatedAt = now
	doc.UploadedAt = &uploadedAt

	if err := s.repo.Update(ctx, doc); err != nil {
		return nil, fmt.Errorf("persist restored document: %w", err)
	}
	if _, err := s.createVersionSnapshot(ctx, doc, targetBytes, now, VersionKindManual, fmt.Sprintf("restored from v%d", version), true); err != nil {
		return nil, err
	}
	if err := s.pruneVersions(ctx, docID); err != nil {
		return nil, err
	}

	s.fireIndexHook(ctx, IndexEvent{
		DocID: docID, Content: content, Version: newVersion, Hash: hash, IsImport: false,
	})

	return &SaveResult{
		ID:        docID,
		Title:     title,
		Version:   newVersion,
		Changed:   true,
		UpdatedAt: now,
	}, nil
}

// HashContent returns the SHA-256 hex digest of content.
// Exported so the importer can check for duplicates before calling CreateFromImport.
func HashContent(content string) string {
	return sha256Content(content)
}

func sha256Content(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}

func (s *Service) shouldCreateAutoSnapshot() bool {
	return s.versioning.Policy == VersionPolicyFull || s.versioning.Policy == VersionPolicyBounded
}

func (s *Service) createVersionSnapshot(ctx context.Context, doc *Document, contentBytes []byte, createdAt time.Time, kind, note string, pinned bool) (*DocumentVersion, error) {
	versionKey := VersionKey(doc.ID, doc.CurrentVersion)
	if err := s.storage.Put(ctx, versionKey, bytes.NewReader(contentBytes), int64(len(contentBytes)), "text/markdown"); err != nil {
		return nil, fmt.Errorf("upload version %d: %w", doc.CurrentVersion, err)
	}
	ver := &DocumentVersion{
		ID:          idgen.New(),
		DocumentID:  doc.ID,
		Version:     doc.CurrentVersion,
		StorageKey:  versionKey,
		ContentHash: doc.ContentHash,
		SizeBytes:   int64(len(contentBytes)),
		CreatedAt:   createdAt,
		Note:        note,
		Kind:        kind,
		Pinned:      pinned,
	}
	if err := s.repo.CreateVersion(ctx, ver); err != nil {
		return nil, fmt.Errorf("persist version: %w", err)
	}
	return ver, nil
}

func (s *Service) ensureCurrentSnapshot(ctx context.Context, doc *Document, targetVersion int) error {
	if _, err := s.repo.GetVersion(ctx, doc.ID, doc.CurrentVersion); err == nil {
		return nil
	} else if !errors.Is(err, ErrNotFound) {
		return err
	}

	rc, err := s.storage.Get(ctx, doc.StorageKey)
	if err != nil {
		return fmt.Errorf("fetch current content before restore: %w", err)
	}
	currentBytes, err := io.ReadAll(rc)
	closeErr := rc.Close()
	if err != nil {
		return fmt.Errorf("read current content before restore: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("close current content before restore: %w", closeErr)
	}
	_, err = s.createVersionSnapshot(ctx, doc, currentBytes, time.Now().UTC(), VersionKindManual, fmt.Sprintf("before restore to v%d", targetVersion), true)
	return err
}

func (s *Service) pruneVersions(ctx context.Context, docID string) error {
	if s.versioning.Policy != VersionPolicyBounded {
		return nil
	}

	versions, err := s.repo.GetVersions(ctx, docID)
	if err != nil {
		return err
	}

	keepLatest := s.versioning.KeepLatest
	var kept int
	var prune []*DocumentVersion
	for _, ver := range versions {
		if ver.Pinned {
			continue
		}
		if kept < keepLatest {
			kept++
			continue
		}
		prune = append(prune, ver)
	}
	if len(prune) == 0 {
		return nil
	}

	prunedVersions := make([]int, 0, len(prune))
	for _, ver := range prune {
		if err := s.storage.Delete(ctx, ver.StorageKey); err != nil {
			return fmt.Errorf("prune version %d: %w", ver.Version, err)
		}
		prunedVersions = append(prunedVersions, ver.Version)
	}
	return s.repo.DeleteVersions(ctx, docID, prunedVersions)
}

func normalizeVersioningConfig(cfg VersioningConfig) VersioningConfig {
	policy := strings.ToLower(strings.TrimSpace(cfg.Policy))
	switch policy {
	case "", "bounded", "limited":
		policy = VersionPolicyBounded
	case "full", "all", "always":
		policy = VersionPolicyFull
	case "overwrite":
		policy = VersionPolicyOverwrite
	default:
		policy = VersionPolicyBounded
	}
	keepLatest := cfg.KeepLatest
	if keepLatest <= 0 {
		keepLatest = 5
	}
	return VersioningConfig{Policy: policy, KeepLatest: keepLatest}
}

func versionSummary(v *DocumentVersion) *VersionSummary {
	return &VersionSummary{
		Version:   v.Version,
		SizeBytes: v.SizeBytes,
		CreatedAt: v.CreatedAt,
		Kind:      v.Kind,
		Pinned:    v.Pinned,
		Note:      v.Note,
	}
}
