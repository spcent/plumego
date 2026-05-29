package organize

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var reNonAlpha = regexp.MustCompile(`[^a-z0-9\p{Han}\p{Hiragana}\p{Katakana}]+`)

// normalizeTitle returns a lowercase, punctuation-stripped version of a title.
func normalizeTitle(title string) string {
	return strings.TrimSpace(reNonAlpha.ReplaceAllString(strings.ToLower(title), " "))
}

// tokenize splits text into lowercase words (ASCII + CJK chars).
func tokenize(text string) []string {
	text = strings.ToLower(text)
	parts := reNonAlpha.Split(text, -1)
	var tokens []string
	for _, p := range parts {
		if len(p) >= 2 {
			tokens = append(tokens, p)
		}
	}
	return tokens
}

// buildShingles returns a set of bi-gram token shingles from a token list.
func buildShingles(tokens []string) map[string]struct{} {
	set := make(map[string]struct{}, len(tokens))
	for i := 0; i < len(tokens)-1; i++ {
		set[tokens[i]+" "+tokens[i+1]] = struct{}{}
	}
	// Also add unigrams to improve recall on short documents.
	for _, t := range tokens {
		set[t] = struct{}{}
	}
	return set
}

// jaccardSimilarity returns the Jaccard similarity between two shingle sets.
func jaccardSimilarity(a, b map[string]struct{}) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}
	intersection := 0
	for k := range a {
		if _, ok := b[k]; ok {
			intersection++
		}
	}
	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0.0
	}
	return float64(intersection) / float64(union)
}

// simhashFromTokens returns a 64-bit simhash as a hex string.
func simhashFromTokens(tokens []string) string {
	counts := [64]int{}
	for _, t := range tokens {
		h := fnv.New64a()
		h.Write([]byte(t))
		v := h.Sum64()
		for i := 0; i < 64; i++ {
			if v&(1<<uint(i)) != 0 {
				counts[i]++
			} else {
				counts[i]--
			}
		}
	}
	var hash uint64
	for i := 0; i < 64; i++ {
		if counts[i] > 0 {
			hash |= 1 << uint(i)
		}
	}
	return fmt.Sprintf("%016x", hash)
}

// hamming64 returns the Hamming distance between two 64-bit simhash hex strings.
func hamming64(a, b string) int {
	if a == b {
		return 0
	}
	var va, vb uint64
	fmt.Sscanf(a, "%016x", &va)
	fmt.Sscanf(b, "%016x", &vb)
	diff := va ^ vb
	dist := 0
	for diff != 0 {
		dist += int(diff & 1)
		diff >>= 1
	}
	return dist
}

// simhashSimilarity converts Hamming distance to a [0,1] similarity score.
func simhashSimilarity(a, b string) float64 {
	d := hamming64(a, b)
	return 1.0 - float64(d)/64.0
}

// bucket groups document IDs by a dimension key.
type bucket struct {
	key  string
	docs []string
}

// buildBuckets assigns documents to buckets by tag, import job, and path prefix.
func buildBuckets(docs []docFingerprintRow, docTags map[string][]string) []bucket {
	tagBuckets := make(map[string][]string)
	for _, d := range docs {
		for _, tagID := range docTags[d.ID] {
			tagBuckets["tag:"+tagID] = append(tagBuckets["tag:"+tagID], d.ID)
		}
	}

	jobBuckets := make(map[string][]string)
	for _, d := range docs {
		if d.ImportJobID != "" {
			jobBuckets["job:"+d.ImportJobID] = append(jobBuckets["job:"+d.ImportJobID], d.ID)
		}
	}

	dirBuckets := make(map[string][]string)
	for _, d := range docs {
		if d.OriginalPath != "" {
			dir := filepath.Dir(d.OriginalPath)
			if dir != "." {
				dirBuckets["dir:"+dir] = append(dirBuckets["dir:"+dir], d.ID)
			}
		}
	}

	// Title prefix buckets (first 3 words of normalised title).
	prefixBuckets := make(map[string][]string)
	for _, d := range docs {
		norm := normalizeTitle(d.Title)
		words := strings.Fields(norm)
		if len(words) >= 2 {
			prefix := strings.Join(words[:min(3, len(words))], " ")
			prefixBuckets["prefix:"+prefix] = append(prefixBuckets["prefix:"+prefix], d.ID)
		}
	}

	var buckets []bucket
	for k, ids := range tagBuckets {
		if len(ids) >= 2 {
			buckets = append(buckets, bucket{key: k, docs: ids})
		}
	}
	for k, ids := range jobBuckets {
		if len(ids) >= 2 {
			buckets = append(buckets, bucket{key: k, docs: ids})
		}
	}
	for k, ids := range dirBuckets {
		if len(ids) >= 2 {
			buckets = append(buckets, bucket{key: k, docs: ids})
		}
	}
	for k, ids := range prefixBuckets {
		if len(ids) >= 2 {
			buckets = append(buckets, bucket{key: k, docs: ids})
		}
	}
	return buckets
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// detectSimilarity computes near-duplicate pairs within each bucket.
func (s *Service) detectSimilarity(ctx context.Context) (int, error) {
	docs, err := s.repo.ListAllDocsForFingerprint(ctx)
	if err != nil {
		return 0, err
	}

	docTags, err := s.repo.ListDocTagIDs(ctx)
	if err != nil {
		return 0, err
	}

	// Build fingerprints.
	type fp struct {
		id       string
		shingles map[string]struct{}
		simhash  string
	}
	fpMap := make(map[string]fp, len(docs))
	for _, d := range docs {
		text := buildDocText(d)
		tokens := tokenize(text)
		fpMap[d.ID] = fp{
			id:       d.ID,
			shingles: buildShingles(tokens),
			simhash:  simhashFromTokens(tokens),
		}
		// Persist fingerprint.
		_ = s.repo.UpsertFingerprint(ctx, &DocumentFingerprint{
			DocumentID:  d.ID,
			TitleNorm:   normalizeTitle(d.Title),
			Simhash:     simhashFromTokens(tokenize(strings.ToLower(d.Title + " " + d.Summary))),
			HeadingHash: simhashFromTokens(tokenize(extractHeadingText(d.HeadingsJSON))),
			KeywordHash: simhashFromTokens(tokens),
		})
	}

	buckets := buildBuckets(docs, docTags)

	// Track pairs already compared to avoid duplicate work.
	seen := make(map[string]struct{})
	count := 0
	limit := s.cfg.MaxComparePerBucket

	for _, b := range buckets {
		compared := 0
		for i := 0; i < len(b.docs) && compared < limit; i++ {
			for j := i + 1; j < len(b.docs) && compared < limit; j++ {
				idA, idB := b.docs[i], b.docs[j]
				if idA > idB {
					idA, idB = idB, idA
				}
				pairKey := idA + ":" + idB
				if _, ok := seen[pairKey]; ok {
					continue
				}
				seen[pairKey] = struct{}{}
				compared++

				fpA, okA := fpMap[idA]
				fpB, okB := fpMap[idB]
				if !okA || !okB {
					continue
				}

				score := jaccardSimilarity(fpA.shingles, fpB.shingles)
				if score < s.cfg.RelatedThreshold {
					continue
				}

				simType := SimilarityTypeRelated
				if score >= s.cfg.NearDuplicateThreshold {
					simType = SimilarityTypeNear
				}

				sim := &DocumentSimilarity{
					ID:              newID(),
					DocumentIDA:     idA,
					DocumentIDB:     idB,
					SimilarityType:  simType,
					SimilarityScore: score,
					Reason:          fmt.Sprintf("bucket=%s jaccard=%.2f", b.key, score),
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
	}
	return count, nil
}

func buildDocText(d docFingerprintRow) string {
	return d.Title + " " + extractHeadingText(d.HeadingsJSON) + " " + d.Summary
}

func extractHeadingText(headingsJSON string) string {
	if headingsJSON == "" {
		return ""
	}
	var headings []struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(headingsJSON), &headings); err != nil {
		return ""
	}
	var parts []string
	for _, h := range headings {
		parts = append(parts, h.Text)
	}
	return strings.Join(parts, " ")
}
