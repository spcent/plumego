package tag

import (
	"errors"
	"time"
)

var (
	ErrNotFound   = errors.New("tag not found")
	ErrDuplicate  = errors.New("tag name already exists")
)

// Tag is a label that can be attached to documents.
type Tag struct {
	ID        string
	Name      string
	Color     string
	Source    string // manual | imported
	CreatedAt time.Time
}

// DocumentTag is the association between a document and a tag.
type DocumentTag struct {
	DocumentID string
	TagID      string
}
