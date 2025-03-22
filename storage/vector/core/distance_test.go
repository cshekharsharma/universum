package core

import (
	"math"
	"testing"
)

func floatEquals(a, b float64, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []float32
		expected float64
	}{
		{
			name:     "Identical vectors",
			a:        []float32{1, 2, 3},
			b:        []float32{1, 2, 3},
			expected: 1.0,
		},
		{
			name:     "Orthogonal vectors",
			a:        []float32{1, 0},
			b:        []float32{0, 1},
			expected: 0.0,
		},
		{
			name:     "Zero vector case",
			a:        []float32{0, 0},
			b:        []float32{1, 1},
			expected: 0.0,
		},
		{
			name:     "Negative cosine similarity",
			a:        []float32{1, 0},
			b:        []float32{-1, 0},
			expected: -1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CosineSimilarity(tt.a, tt.b)
			if !floatEquals(got, tt.expected, 1e-6) {
				t.Errorf("CosineSimilarity() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestEuclideanDistance(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []float32
		expected float64
	}{
		{
			name:     "Same vectors",
			a:        []float32{1, 2, 3},
			b:        []float32{1, 2, 3},
			expected: 0.0, // same vector means zero distance, negated
		},
		{
			name:     "Simple distance",
			a:        []float32{0, 0},
			b:        []float32{3, 4},
			expected: -5.0, // 3-4-5 triangle
		},
		{
			name:     "Negative components",
			a:        []float32{-1, -2},
			b:        []float32{1, 2},
			expected: -math.Sqrt(20), // diff is (2,4)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EuclideanDistance(tt.a, tt.b)
			if !floatEquals(got, tt.expected, 1e-6) {
				t.Errorf("EuclideanDistance() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestDotProductSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []float32
		expected float64
	}{
		{
			name:     "Positive dot product",
			a:        []float32{1, 2, 3},
			b:        []float32{4, 5, 6},
			expected: 32.0,
		},
		{
			name:     "Zero vectors",
			a:        []float32{0, 0, 0},
			b:        []float32{1, 2, 3},
			expected: 0.0,
		},
		{
			name:     "Negative dot product",
			a:        []float32{-1, 0},
			b:        []float32{1, 0},
			expected: -1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DotProductSimilarity(tt.a, tt.b)
			if !floatEquals(got, tt.expected, 1e-6) {
				t.Errorf("DotProductSimilarity() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
