package ai

import (
	"context"
	"fmt"
	"io"
	"strings"

	"cloud-vault/internal/storage"
)

// ContextBuilder assembles document text for an LLM prompt.
type ContextBuilder struct {
	repo     *Repository
	store    storage.ObjectStorage
	maxTokens int
}

func NewContextBuilder(repo *Repository, store storage.ObjectStorage, maxTokens int) *ContextBuilder {
	return &ContextBuilder{repo: repo, store: store, maxTokens: maxTokens}
}

// BuildContext loads the listed document IDs and returns a combined context
// string truncated to maxTokens. It only uses explicitly provided IDs — no
// whole-library scanning.
func (cb *ContextBuilder) BuildContext(ctx context.Context, documentIDs []string) (string, []DocMeta, error) {
	var metas []DocMeta
	for _, id := range documentIDs {
		meta, err := cb.repo.GetDocMeta(id)
		if err != nil {
			return "", nil, fmt.Errorf("get doc meta %s: %w", id, err)
		}
		if meta == nil {
			continue
		}
		metas = append(metas, *meta)
	}

	var sb strings.Builder
	usedTokens := 0

	for _, meta := range metas {
		header := fmt.Sprintf("## Document: %s (ID: %s)\n\n", meta.Title, meta.ID)
		sb.WriteString(header)
		usedTokens += estimateTokens(header)

		content, err := cb.loadContent(ctx, meta)
		if err != nil {
			// Fall back to summary/headings if storage unavailable.
			fallback := cb.fallbackText(meta)
			sb.WriteString(fallback)
			usedTokens += estimateTokens(fallback)
		} else {
			remaining := cb.maxTokens - usedTokens
			if remaining <= 0 {
				break
			}
			truncated := truncateToTokens(content, remaining)
			sb.WriteString(truncated)
			sb.WriteString("\n\n")
			usedTokens += estimateTokens(truncated)
		}

		if usedTokens >= cb.maxTokens {
			sb.WriteString("\n[Context truncated due to token limit]\n")
			break
		}
	}

	return sb.String(), metas, nil
}

func (cb *ContextBuilder) loadContent(ctx context.Context, meta DocMeta) (string, error) {
	rc, err := cb.store.Get(ctx, meta.StorageKey)
	if err != nil {
		return "", err
	}
	defer rc.Close()
	data, err := io.ReadAll(io.LimitReader(rc, 2*1024*1024))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (cb *ContextBuilder) fallbackText(meta DocMeta) string {
	var parts []string
	if meta.Summary != nil && *meta.Summary != "" {
		parts = append(parts, "Summary: "+*meta.Summary)
	}
	if meta.HeadingText != nil && *meta.HeadingText != "" {
		parts = append(parts, "Headings: "+*meta.HeadingText)
	}
	if len(parts) == 0 {
		return "(no content available)\n\n"
	}
	return strings.Join(parts, "\n") + "\n\n"
}

// truncateToTokens returns a prefix of s that fits within tokenLimit.
func truncateToTokens(s string, tokenLimit int) string {
	if estimateTokens(s) <= tokenLimit {
		return s
	}
	// Rough: 4 chars per token.
	cutBytes := tokenLimit * 4
	if cutBytes > len(s) {
		return s
	}
	// Walk back to a newline boundary.
	for cutBytes > 0 && s[cutBytes-1] != '\n' {
		cutBytes--
	}
	if cutBytes == 0 {
		cutBytes = tokenLimit * 4
		if cutBytes > len(s) {
			cutBytes = len(s)
		}
	}
	return s[:cutBytes] + "\n[...truncated]"
}
