package hnsw

import (
	"testing"
	"universum/storage/vector/core"
)

func testOptions(metric uint8) HNSWOptions {
	return HNSWOptions{
		DistanceMetric: metric,
		MaxNeighbors:   2,
		EfConstruct:    4,
		EfSearch:       4,
	}
}

func f32(v ...float32) []float32 {
	return v
}

func TestInsertAndSearchSingleVector(t *testing.T) {
	h := NewHNSW(testOptions(DistanceMetricCosine))

	err := h.Insert("vec1", f32(1, 0))
	if err != nil {
		t.Fatalf("unexpected error on insert: %v", err)
	}

	results, err := h.Search(f32(1, 0), 1)
	if err != nil {
		t.Fatalf("unexpected error on search: %v", err)
	}

	if len(results) != 1 || results[0].ID != "vec1" {
		t.Errorf("expected vec1 as result, got %+v", results)
	}
}

func TestInsertDuplicateID(t *testing.T) {
	h := NewHNSW(testOptions(DistanceMetricCosine))

	_ = h.Insert("vec1", f32(1, 0))
	err := h.Insert("vec1", f32(0, 1))

	if err == nil {
		t.Error("expected error on duplicate insert, got nil")
	}
}

func TestSearchOnEmptyIndex(t *testing.T) {
	h := NewHNSW(testOptions(DistanceMetricCosine))

	_, err := h.Search(f32(1, 0), 1)
	if err == nil {
		t.Error("expected error on empty search, got nil")
	}
}

func TestNearestNeighborSelection(t *testing.T) {
	h := NewHNSW(testOptions(DistanceMetricDotProduct))

	_ = h.Insert("v1", f32(1, 0))   // dot = 1
	_ = h.Insert("v2", f32(0, 1))   // dot = 0
	_ = h.Insert("v3", f32(0.5, 0)) // dot = 0.5

	results, err := h.Search(f32(1, 0), 2)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if results[0].ID != "v1" || results[1].ID != "v3" {
		t.Errorf("unexpected search results: got %s, %s", results[0].ID, results[1].ID)
	}
}

func TestLevelPromotionAndEntryPointChange(t *testing.T) {
	h := NewHNSW(testOptions(DistanceMetricCosine))

	h.getRandomLevel = func() int {
		return h.MaxLevel + 1
	}

	_ = h.Insert("a", f32(1, 0)) // level 0
	if h.EntryPoint == nil || h.EntryPoint.ID != "a" {
		t.Errorf("EntryPoint not set correctly after first insert")
	}

	_ = h.Insert("b", f32(0, 1)) // promoted level
	if h.EntryPoint.ID != "b" {
		t.Errorf("Expected new entry point to be 'b', got %v", h.EntryPoint.ID)
	}
}

func TestMultipleDistanceMetrics(t *testing.T) {
	metrics := []uint8{
		DistanceMetricCosine,
		DistanceMetricDotProduct,
		DistanceMetricEuclidean,
	}

	for _, metric := range metrics {
		t.Run(string(metric), func(t *testing.T) {
			h := NewHNSW(testOptions(metric))

			_ = h.Insert("x", f32(1, 2))
			results, err := h.Search(f32(1, 2), 1)
			if err != nil {
				t.Fatalf("Search failed for metric %v: %v", metric, err)
			}
			if len(results) != 1 || results[0].ID != "x" {
				t.Errorf("Unexpected result for metric %v: %+v", metric, results)
			}
		})
	}
}

func TestGreedySearchSelectsBetterNeighbor(t *testing.T) {
	h := NewHNSW(testOptions(DistanceMetricDotProduct))

	a := &core.VectorNode{
		ID:     "a",
		Vector: f32(1, 0),
		Level:  1,
		Links:  map[int][]*core.VectorNode{},
	}
	b := &core.VectorNode{
		ID:     "b",
		Vector: f32(2, 0), // Closer in dot product to query
		Level:  1,
		Links:  map[int][]*core.VectorNode{},
	}

	// Wire node a to b at level 1
	a.Links[1] = []*core.VectorNode{b}

	result := h.greedySearch(a, f32(2, 0), 1)

	if result.ID != "b" {
		t.Errorf("Expected greedySearch to return b, got %v", result.ID)
	}
}
