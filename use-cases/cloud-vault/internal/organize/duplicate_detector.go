package organize

import (
	"context"
	"time"
)

// detectDuplicates finds exact duplicates by content_hash and writes pairs to document_similarity.
func (s *Service) detectDuplicates(ctx context.Context) (int, error) {
	groups, err := s.repo.ListDuplicateGroups(ctx)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, g := range groups {
		if len(g.Documents) < 2 {
			continue
		}
		// The first document (oldest) is treated as the canonical one.
		canonical := g.Documents[0]
		for _, dup := range g.Documents[1:] {
			// Ensure canonical ID < dup ID to maintain a stable pair ordering.
			idA, idB := canonical.ID, dup.ID
			if idA > idB {
				idA, idB = idB, idA
			}
			sim := &DocumentSimilarity{
				ID:              newID(),
				DocumentIDA:     idA,
				DocumentIDB:     idB,
				SimilarityType:  SimilarityTypeExact,
				SimilarityScore: 1.0,
				Reason:          "identical content_hash: " + g.ContentHash[:8],
				Status:          SimilarityStatusPending,
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
			}
			if err := s.repo.UpsertSimilarity(ctx, sim); err != nil {
				return count, err
			}
			count++
		}
	}
	return count, nil
}
