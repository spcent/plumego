package semanticcache

import (
	"context"
	"testing"
)

func TestSimpleEmbeddingGenerator(t *testing.T) {
	ctx := context.Background()
	gen := NewSimpleEmbeddingGenerator(128)

	t.Run("Generate", func(t *testing.T) {
		text := "Hello, world!"
		emb, err := gen.Generate(ctx, text)
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}

		if emb == nil {
			t.Fatal("embedding is nil")
		}

		if len(emb.Vector) != 128 {
			t.Errorf("expected 128 dimensions, got %d", len(emb.Vector))
		}

		if emb.Model != "simple-hash" {
			t.Errorf("expected model 'simple-hash', got %s", emb.Model)
		}

		if emb.Text != text {
			t.Errorf("expected text %q, got %q", text, emb.Text)
		}

		if emb.Hash == "" {
			t.Error("hash is empty")
		}
	})

	t.Run("Deterministic", func(t *testing.T) {
		text := "Test text"
		emb1, _ := gen.Generate(ctx, text)
		emb2, _ := gen.Generate(ctx, text)

		if emb1.Hash != emb2.Hash {
			t.Error("embeddings should be deterministic")
		}

		if len(emb1.Vector) != len(emb2.Vector) {
			t.Error("vector dimensions should match")
		}

		for i := range emb1.Vector {
			if emb1.Vector[i] != emb2.Vector[i] {
				t.Errorf("vector mismatch at index %d", i)
				break
			}
		}
	})

	t.Run("DifferentTexts", func(t *testing.T) {
		emb1, _ := gen.Generate(ctx, "Hello")
		emb2, _ := gen.Generate(ctx, "World")

		if emb1.Hash == emb2.Hash {
			t.Error("different texts should have different hashes")
		}

		// Vectors should be different
		same := true
		for i := range emb1.Vector {
			if emb1.Vector[i] != emb2.Vector[i] {
				same = false
				break
			}
		}
		if same {
			t.Error("different texts should have different vectors")
		}
	})

	t.Run("GenerateBatch", func(t *testing.T) {
		texts := []string{"text1", "text2", "text3"}
		embeddings, err := gen.GenerateBatch(ctx, texts)
		if err != nil {
			t.Fatalf("GenerateBatch failed: %v", err)
		}

		if len(embeddings) != len(texts) {
			t.Errorf("expected %d embeddings, got %d", len(texts), len(embeddings))
		}

		for i, emb := range embeddings {
			if emb.Text != texts[i] {
				t.Errorf("embedding %d text mismatch", i)
			}
		}
	})

	t.Run("Dimensions", func(t *testing.T) {
		if gen.Dimensions() != 128 {
			t.Errorf("expected 128 dimensions, got %d", gen.Dimensions())
		}
	})

	t.Run("Model", func(t *testing.T) {
		if gen.Model() != "simple-hash" {
			t.Errorf("expected model 'simple-hash', got %s", gen.Model())
		}
	})
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        []float64
		b        []float64
		expected float64
		wantErr  bool
	}{
		{
			name:     "Identical vectors",
			a:        []float64{1.0, 0.0, 0.0},
			b:        []float64{1.0, 0.0, 0.0},
			expected: 1.0,
		},
		{
			name:     "Orthogonal vectors",
			a:        []float64{1.0, 0.0, 0.0},
			b:        []float64{0.0, 1.0, 0.0},
			expected: 0.0,
		},
		{
			name:     "Opposite vectors",
			a:        []float64{1.0, 0.0, 0.0},
			b:        []float64{-1.0, 0.0, 0.0},
			expected: -1.0,
		},
		{
			name:     "Similar vectors",
			a:        []float64{1.0, 1.0, 0.0},
			b:        []float64{1.0, 0.9, 0.1},
			expected: 0.9, // approximate
		},
		{
			name:    "Dimension mismatch",
			a:       []float64{1.0, 0.0},
			b:       []float64{1.0, 0.0, 0.0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CosineSimilarity(tt.a, tt.b)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Allow small floating point differences
			if !floatEquals(result, tt.expected, 0.1) {
				t.Errorf("expected similarity %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestEuclideanDistance(t *testing.T) {
	tests := []struct {
		name     string
		a        []float64
		b        []float64
		expected float64
		wantErr  bool
	}{
		{
			name:     "Identical vectors",
			a:        []float64{1.0, 0.0, 0.0},
			b:        []float64{1.0, 0.0, 0.0},
			expected: 0.0,
		},
		{
			name:     "Unit distance",
			a:        []float64{0.0, 0.0, 0.0},
			b:        []float64{1.0, 0.0, 0.0},
			expected: 1.0,
		},
		{
			name:     "3-4-5 triangle",
			a:        []float64{0.0, 0.0},
			b:        []float64{3.0, 4.0},
			expected: 5.0,
		},
		{
			name:    "Dimension mismatch",
			a:       []float64{1.0, 0.0},
			b:       []float64{1.0, 0.0, 0.0},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EuclideanDistance(tt.a, tt.b)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !floatEquals(result, tt.expected, 0.01) {
				t.Errorf("expected distance %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestNormalizeVector(t *testing.T) {
	tests := []struct {
		name     string
		input    []float64
		expected float64 // expected norm after normalization
	}{
		{
			name:     "Already normalized",
			input:    []float64{1.0, 0.0, 0.0},
			expected: 1.0,
		},
		{
			name:     "Needs normalization",
			input:    []float64{3.0, 4.0},
			expected: 1.0,
		},
		{
			name:     "All same values",
			input:    []float64{2.0, 2.0, 2.0},
			expected: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeVector(tt.input)

			// Calculate norm
			var norm float64
			for _, v := range result {
				norm += v * v
			}
			norm = sqrt(norm)

			if !floatEquals(norm, tt.expected, 0.01) {
				t.Errorf("expected norm %f, got %f", tt.expected, norm)
			}
		})
	}
}

func floatEquals(a, b, epsilon float64) bool {
	if a > b {
		return a-b < epsilon
	}
	return b-a < epsilon
}

func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z -= (z*z - x) / (2 * z)
	}
	return z
}
