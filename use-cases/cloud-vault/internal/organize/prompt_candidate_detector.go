package organize

import (
	"context"
	"io"
	"strings"

	"cloud-vault/internal/storage"
)

var promptKeywords = []string{
	"你现在是",
	"请实现",
	"约束",
	"验收标准",
	"输出格式",
	"claude code",
	"codex",
	"prompt",
	"你是一个",
	"system prompt",
	"user:",
	"assistant:",
	"<system>",
	"<human>",
}

// detectPromptCandidates scans document content for prompt-like patterns.
func (s *Service) detectPromptCandidates(ctx context.Context) (int, error) {
	docs, err := s.repo.ListAllDocsForFingerprint(ctx)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, d := range docs {
		// Fetch content from object storage.
		storageKey := "docs/" + d.ID + "/current.md"
		rc, err := s.store.Get(ctx, storageKey)
		if err != nil {
			// Skip if not found (e.g., orphaned record).
			continue
		}
		contentBytes, readErr := io.ReadAll(rc)
		rc.Close()
		if readErr != nil {
			continue
		}

		lowerContent := strings.ToLower(string(contentBytes))
		matchCount := 0
		for _, kw := range promptKeywords {
			if strings.Contains(lowerContent, kw) {
				matchCount++
			}
		}

		if matchCount == 0 {
			continue
		}

		isCandidate := matchCount >= 2
		score := float64(matchCount) / float64(len(promptKeywords)) * 100

		if err := s.repo.UpdatePromptCandidate(ctx, d.ID, isCandidate, score); err != nil {
			return count, err
		}
		if isCandidate {
			count++
		}
	}
	return count, nil
}

// detectPromptCandidatesFromFTS uses document_fts content for prompt detection (faster, no storage calls).
func (s *Service) detectPromptCandidatesFromFTS(ctx context.Context) (int, error) {
	_ = s.store // store may be used for full content if FTS is not populated
	return s.detectPromptCandidatesFTSBased(ctx)
}

func (s *Service) detectPromptCandidatesFTSBased(ctx context.Context) (int, error) {
	docs, err := s.repo.ListAllDocsForFingerprint(ctx)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, d := range docs {
		// Use title + summary + headings as proxy (storage read is expensive).
		combined := strings.ToLower(d.Title + " " + d.Summary + " " + extractHeadingText(d.HeadingsJSON))
		matchCount := 0
		for _, kw := range promptKeywords {
			if strings.Contains(combined, kw) {
				matchCount++
			}
		}

		if matchCount == 0 {
			_ = s.repo.UpdatePromptCandidate(ctx, d.ID, false, 0)
			continue
		}

		isCandidate := matchCount >= 1
		score := float64(matchCount) / float64(len(promptKeywords)) * 100

		if err := s.repo.UpdatePromptCandidate(ctx, d.ID, isCandidate, score); err != nil {
			return count, err
		}
		if isCandidate {
			count++
		}
	}
	return count, nil
}

// ensure storage is referenced (used in detectPromptCandidates).
var _ = (*storage.LocalStorage)(nil)
