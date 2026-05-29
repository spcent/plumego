package collection

import (
	"errors"
	"time"
)

var (
	ErrNotFound  = errors.New("collection not found")
	ErrDuplicate = errors.New("document already in collection")
)

// Collection is a named set of document references.
type Collection struct {
	ID          string
	Name        string
	Description string
	Type        string // manual | search | topic
	Status      string // active | archived
	DocCount    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CollectionDocument is a document entry inside a collection.
type CollectionDocument struct {
	CollectionID string
	DocumentID   string
	Title        string
	Status       string
	SortOrder    int
	Note         string
	CreatedAt    time.Time
}
