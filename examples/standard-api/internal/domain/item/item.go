// Package item contains the reference application's item domain model,
// repository, and service. The three files implement the canonical three-layer
// structure described in docs/reference/canonical-style-guide.md §3:
//
//	item.go    — domain model (the Item type)
//	store.go   — persistence layer (Repository interface and MemoryStore)
//	service.go — business logic layer (Service interface and ItemService)
package item

import "time"

// Item is the canonical item resource in this reference application.
type Item struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
