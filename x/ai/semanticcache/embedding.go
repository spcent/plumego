// Package semanticcache provides semantic caching for LLM responses using embeddings.
package semanticcache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
)

// Embedding represents a vector embedding.
type Embedding struct {
	// Vector representation
	Vector []float64

	// Metadata
	Model      string
	Dimensions int
	Text       string
	Hash       string
}

// EmbeddingGenerator generates embeddings from text.
type EmbeddingGenerator interface {
	// Generate creates an embedding from text.
	Generate(ctx context.Context, text string) (*Embedding, error)

	// GenerateBatch creates embeddings for multiple texts.
	GenerateBatch(ctx context.Context, texts []string) ([]*Embedding, error)

	// Dimensions returns the embedding dimension count.
	Dimensions() int

	// Model returns the model name.
	Model() string
}

// SimpleEmbeddingGenerator creates simple hash-based embeddings for testing.
// In production, use a real embedding model (e.g., Claude embeddings, OpenAI embeddings).
type SimpleEmbeddingGenerator struct {
	model      string
	dimensions int
}

// NewSimpleEmbeddingGenerator creates a simple embedding generator.
// This is for testing only - use real embeddings in production.
func NewSimpleEmbeddingGenerator(dimensions int) *SimpleEmbeddingGenerator {
	if dimensions == 0 {
		dimensions = 128
	}
	return &SimpleEmbeddingGenerator{
		model:      "simple-hash",
		dimensions: dimensions,
	}
}

// Generate implements EmbeddingGenerator.
func (g *SimpleEmbeddingGenerator) Generate(ctx context.Context, text string) (*Embedding, error) {
	// Hash the text
	hash := sha256.Sum256([]byte(text))
	hashStr := hex.EncodeToString(hash[:])

	// Convert hash to pseudo-vector
	vector := make([]float64, g.dimensions)
	for i := 0; i < g.dimensions; i++ {
		// Use hash bytes to generate deterministic "embedding"
		byteIdx := i % len(hash)
		vector[i] = float64(hash[byteIdx]) / 255.0
	}

	// Normalize
	vector = normalizeVector(vector)

	return &Embedding{
		Vector:     vector,
		Model:      g.model,
		Dimensions: g.dimensions,
		Text:       text,
		Hash:       hashStr,
	}, nil
}

// GenerateBatch implements EmbeddingGenerator.
func (g *SimpleEmbeddingGenerator) GenerateBatch(ctx context.Context, texts []string) ([]*Embedding, error) {
	embeddings := make([]*Embedding, len(texts))
	for i, text := range texts {
		emb, err := g.Generate(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = emb
	}
	return embeddings, nil
}

// Dimensions implements EmbeddingGenerator.
func (g *SimpleEmbeddingGenerator) Dimensions() int {
	return g.dimensions
}

// Model implements EmbeddingGenerator.
func (g *SimpleEmbeddingGenerator) Model() string {
	return g.model
}

// CosineSimilarity computes cosine similarity between two vectors.
func CosineSimilarity(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vector dimensions mismatch: %d != %d", len(a), len(b))
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0, nil
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB)), nil
}

// EuclideanDistance computes Euclidean distance between two vectors.
func EuclideanDistance(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vector dimensions mismatch: %d != %d", len(a), len(b))
	}

	var sum float64
	for i := 0; i < len(a); i++ {
		diff := a[i] - b[i]
		sum += diff * diff
	}

	return math.Sqrt(sum), nil
}

// normalizeVector normalizes a vector to unit length.
func normalizeVector(v []float64) []float64 {
	var norm float64
	for _, val := range v {
		norm += val * val
	}
	norm = math.Sqrt(norm)

	if norm == 0 {
		return v
	}

	normalized := make([]float64, len(v))
	for i, val := range v {
		normalized[i] = val / norm
	}
	return normalized
}
