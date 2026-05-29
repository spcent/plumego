package organize

import (
	"context"
	"path/filepath"
	"strings"
	"time"
)

// buildTopics creates rule-based topics from existing tags and path keywords.
func (s *Service) buildTopics(ctx context.Context) (int, error) {
	tagNames, err := s.repo.ListTagNames(ctx)
	if err != nil {
		return 0, err
	}

	docTags, err := s.repo.ListDocTagIDs(ctx)
	if err != nil {
		return 0, err
	}

	// Build topic for each existing tag that has at least 2 documents.
	tagDocCount := make(map[string]int)
	for _, tags := range docTags {
		for _, tid := range tags {
			tagDocCount[tid]++
		}
	}

	// Build doc → tagIDs reverse for scoring.
	topicDocMap := make(map[string][]string) // topicID → []docIDs
	topicIDs := make(map[string]string)       // tagID → topicID

	count := 0
	for tagID, cnt := range tagDocCount {
		if cnt < 2 {
			continue
		}
		tagName := tagNames[tagID]
		if tagName == "" {
			continue
		}

		topicID := "tag:" + tagID
		topicIDs[tagID] = topicID

		topic := &Topic{
			ID:          topicID,
			Name:        tagName,
			Description: "Documents tagged with " + tagName,
			Source:      "rule",
			Status:      "active",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}
		if err := s.repo.UpsertTopic(ctx, topic); err != nil {
			return count, err
		}
		count++
	}

	// Assign documents to tag-based topics.
	for docID, tags := range docTags {
		for _, tagID := range tags {
			topicID, ok := topicIDs[tagID]
			if !ok {
				continue
			}
			topicDocMap[topicID] = append(topicDocMap[topicID], docID)
		}
	}

	// Also create path-based topics for directories with ≥ 3 documents.
	docs, err := s.repo.ListAllDocsForFingerprint(ctx)
	if err != nil {
		return count, err
	}

	dirDocs := make(map[string][]string)
	for _, d := range docs {
		if d.OriginalPath == "" {
			continue
		}
		dir := filepath.ToSlash(filepath.Dir(d.OriginalPath))
		if dir == "." || dir == "" {
			continue
		}
		// Use the last meaningful path segment as topic name.
		parts := strings.Split(dir, "/")
		var meaningful []string
		for _, p := range parts {
			if p != "" && p != "." && p != ".." {
				meaningful = append(meaningful, p)
			}
		}
		if len(meaningful) == 0 {
			continue
		}
		topicKey := "dir:" + dir
		dirDocs[topicKey] = append(dirDocs[topicKey], d.ID)
		topicDocMap[topicKey] = append(topicDocMap[topicKey], d.ID)
		_ = meaningful // name is derived below
	}

	for topicKey, docIDs := range dirDocs {
		if len(docIDs) < 3 {
			continue
		}
		dirPath := strings.TrimPrefix(topicKey, "dir:")
		parts := strings.Split(dirPath, "/")
		var name string
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] != "" {
				name = titleCase(parts[i])
				break
			}
		}
		if name == "" {
			continue
		}

		topic := &Topic{
			ID:          topicKey,
			Name:        name,
			Description: "Documents in path: " + dirPath,
			Source:      "rule",
			Status:      "active",
		}
		if err := s.repo.UpsertTopic(ctx, topic); err != nil {
			return count, err
		}
		count++
	}

	// Write topic_documents rows.
	for topicID, docIDs := range topicDocMap {
		seen := make(map[string]struct{})
		for _, docID := range docIDs {
			if _, ok := seen[docID]; ok {
				continue
			}
			seen[docID] = struct{}{}
			td := &TopicDocument{
				TopicID:    topicID,
				DocumentID: docID,
				Score:      1.0,
				Source:     "rule",
			}
			if err := s.repo.AddTopicDocument(ctx, td); err != nil {
				return count, err
			}
		}
	}

	return count, nil
}
