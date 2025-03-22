package core

import "math"

// DistanceFunc defines a generic function type for computing distance or similarity
// between two float32 vectors. Higher values are generally considered more similar.
type DistanceFunc func(a, b []float32) float64

// CosineSimilarity computes the cosine similarity between two vectors `a` and `b`.
//
// Cosine similarity is defined as:
//
//	cos(θ) = (a • b) / (||a|| * ||b||)
//
// Where:
//   - "•" denotes dot product
//   - ||a|| is the Euclidean norm (L2) of vector a
//
// It returns a value between -1 and 1, where 1 means perfectly similar,
// 0 means orthogonal (no similarity), and -1 means diametrically opposite.
//
// If either vector has zero magnitude, the function returns 0 to avoid division by zero.
func CosineSimilarity(a, b []float32) float64 {
	var dot, normA, normB float64

	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// EuclideanDistance computes the Euclidean (L2) distance between two vectors `a` and `b`,
// and returns its negative value.
//
// The formula used is:
//
//	distance = -sqrt(Σ (aᵢ - bᵢ)²)
//
// The result is negated so that higher values (i.e., less negative) represent closer vectors,
// which is useful when using a max-heap in nearest-neighbor algorithms.
func EuclideanDistance(a, b []float32) float64 {
	var sum float64

	for i := range a {
		diff := float64(a[i]) - float64(b[i])
		sum += diff * diff
	}

	return -math.Sqrt(sum)
}

// DotProductSimilarity computes the dot product of two vectors `a` and `b`.
//
// The dot product is defined as:
//
//	a • b = Σ (aᵢ * bᵢ)
//
// A higher dot product indicates greater similarity in direction and magnitude.
// This function is commonly used in recommendation systems and vector search.
func DotProductSimilarity(a, b []float32) float64 {
	var dot float64

	for i := range a {
		dot += float64(a[i]) * float64(b[i])
	}

	return dot
}
