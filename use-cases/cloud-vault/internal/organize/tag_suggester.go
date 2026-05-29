package organize

import (
	"context"
	"path/filepath"
	"strings"
	"time"
)

// suggestTags generates tag suggestions for all non-deleted documents.
func (s *Service) suggestTags(ctx context.Context) (int, error) {
	docs, err := s.repo.ListAllDocsForFingerprint(ctx)
	if err != nil {
		return 0, err
	}

	tagNames, err := s.repo.ListTagNames(ctx)
	if err != nil {
		return 0, err
	}

	docTags, err := s.repo.ListDocTagIDs(ctx)
	if err != nil {
		return 0, err
	}

	// Build reverse map: tagName → tagID.
	nameToID := make(map[string]string, len(tagNames))
	for id, name := range tagNames {
		nameToID[strings.ToLower(name)] = id
	}

	// Build set of existing tag names (lowercase) per document.
	existingTags := make(map[string]map[string]struct{}, len(docs))
	for docID, tagIDs := range docTags {
		set := make(map[string]struct{})
		for _, tid := range tagIDs {
			set[strings.ToLower(tagNames[tid])] = struct{}{}
		}
		existingTags[docID] = set
	}

	count := 0
	for _, d := range docs {
		candidates := extractTagCandidates(d)
		existing := existingTags[d.ID]

		for _, c := range candidates {
			nameLow := strings.ToLower(c.name)
			// Skip if the document already has this tag.
			if _, has := existing[nameLow]; has {
				continue
			}
			tagID := nameToID[nameLow]
			sugg := &TagSuggestion{
				ID:         newID(),
				DocumentID: d.ID,
				TagID:      tagID,
				TagName:    c.name,
				Source:     c.source,
				Confidence: c.confidence,
				Status:     TagSuggestionStatusPending,
				CreatedAt:  time.Now().UTC(),
				UpdatedAt:  time.Now().UTC(),
			}
			if err := s.repo.UpsertTagSuggestion(ctx, sugg); err != nil {
				return count, err
			}
			count++
		}
	}
	return count, nil
}

type tagCandidate struct {
	name       string
	source     string
	confidence float64
}

func extractTagCandidates(d docFingerprintRow) []tagCandidate {
	seen := make(map[string]struct{})
	var candidates []tagCandidate

	add := func(name, source string, conf float64) {
		if name == "" {
			return
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		candidates = append(candidates, tagCandidate{name: name, source: source, confidence: conf})
	}

	// From path components.
	if d.OriginalPath != "" {
		parts := strings.FieldsFunc(filepath.ToSlash(d.OriginalPath), func(r rune) bool {
			return r == '/' || r == '\\'
		})
		for i, part := range parts {
			part = strings.TrimSuffix(part, filepath.Ext(part))
			if isStopWord(part) || len(part) < 2 {
				continue
			}
			conf := 0.6
			if i < len(parts)-1 {
				conf = 0.75
			}
			add(titleCase(part), "path", conf)
		}
	}

	// From title words.
	for _, word := range strings.Fields(d.Title) {
		word = strings.Trim(word, ".,!?:;\"'")
		if len(word) >= 3 && !isStopWord(strings.ToLower(word)) {
			add(word, "title", 0.5)
		}
	}

	// From headings (up to first 3).
	headings := strings.Fields(extractHeadingText(d.HeadingsJSON))
	for i, word := range headings {
		if i >= 20 {
			break
		}
		word = strings.Trim(word, ".,!?:;\"'")
		if len(word) >= 3 && !isStopWord(strings.ToLower(word)) {
			add(word, "heading", 0.4)
		}
	}

	return candidates
}

func isStopWord(w string) bool {
	switch w {
	case "a", "an", "the", "and", "or", "but", "in", "on", "at", "to", "for",
		"of", "with", "is", "are", "was", "were", "be", "been", "being",
		"it", "its", "this", "that", "these", "those", "my", "your",
		"index", "readme", "docs", "doc", "file", "files", "src", "lib",
		"untitled", "new", "note", "notes", "page", "pages", "md":
		return true
	}
	return false
}

func titleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}
