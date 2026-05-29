package tag

import "time"

// CreateTagRequest is the payload for creating a tag.
type CreateTagRequest struct {
	Name   string `json:"name"`
	Color  string `json:"color,omitempty"`
	Source string `json:"source,omitempty"`
}

// UpdateTagRequest is the payload for updating a tag.
type UpdateTagRequest struct {
	Name  string `json:"name,omitempty"`
	Color string `json:"color,omitempty"`
}

// SetDocumentTagsRequest replaces all tags on a document.
type SetDocumentTagsRequest struct {
	TagIDs []string `json:"tag_ids"`
}

// TagResponse is the response representation of a tag.
type TagResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color,omitempty"`
	Source    string    `json:"source,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ListTagsResponse is the response for tag list operations.
type ListTagsResponse struct {
	Items []TagResponse `json:"items"`
	Total int           `json:"total"`
}

func toResponse(t *Tag) TagResponse {
	return TagResponse{
		ID:        t.ID,
		Name:      t.Name,
		Color:     t.Color,
		Source:    t.Source,
		CreatedAt: t.CreatedAt,
	}
}
