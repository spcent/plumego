package organize

import (
	"context"
	"encoding/json"
	"strings"
)

// scoreQuality computes a quality_score for all non-deleted documents.
// Score range: 0–100.
func (s *Service) scoreQuality(ctx context.Context) (int, error) {
	rows, err := s.repo.ListDocumentsForScoring(ctx)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, row := range rows {
		score := computeQualityScore(row)
		isDup, _ := s.repo.IsDuplicateHash(ctx, row.ContentHash)
		if isDup {
			score -= 20
		}
		if score < 0 {
			score = 0
		}
		if err := s.repo.UpdateQualityScore(ctx, row.ID, score); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func computeQualityScore(row docScoringRow) float64 {
	score := 50.0

	// Title quality.
	title := strings.TrimSpace(row.Title)
	if len(title) < 3 {
		score -= 15
	} else if len(title) >= 10 {
		score += 10
	}

	// Word count.
	switch {
	case row.WordCount < 30:
		score -= 20
	case row.WordCount < 100:
		score -= 5
	case row.WordCount >= 300:
		score += 10
	case row.WordCount >= 1000:
		score += 15
	}

	// Headings.
	headingCount := countHeadings(row.HeadingsJSON)
	switch {
	case headingCount == 0:
		score -= 5
	case headingCount >= 3:
		score += 10
	case headingCount >= 8:
		score += 15
	}

	// Code blocks.
	if row.CodeBlockCount > 0 {
		score += 10
	}

	// Favorites signal value.
	if row.IsFavorite {
		score += 10
	}

	// Review status.
	if row.ReviewStatus == "valuable" {
		score += 15
	}

	if score > 100 {
		score = 100
	}
	return score
}

func countHeadings(headingsJSON string) int {
	if headingsJSON == "" {
		return 0
	}
	var headings []struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(headingsJSON), &headings); err != nil {
		return 0
	}
	return len(headings)
}
