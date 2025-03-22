// Package hnsw provides an in-memory implementation of the Hierarchical Navigable Small World (HNSW) algorithm
// for approximate nearest neighbor (ANN) search in high-dimensional vector spaces.
//
// This package supports different similarity/distance functions (cosine, Euclidean, dot product)
// and allows you to tune key parameters like max neighbors per node and search exploration factor.
//
// Key Features:
//   - Insert high-dimensional vectors into the index with `Insert`
//   - Perform fast approximate nearest-neighbor search with `Search`
//   - Configurable distance metrics and search quality vs performance trade-offs
package hnsw

import (
	"errors"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"
	"universum/storage/vector/core"
)

// HNSW represents the Hierarchical Navigable Small World index.
// It stores nodes organized by levels with interconnections that allow fast
// approximate nearest neighbor searches.
type HNSW struct {
	Nodes        map[string]*core.VectorNode // All inserted nodes, indexed by ID
	MaxLevel     int                         // Highest level of any node in the index
	EntryPoint   *core.VectorNode            // Starting node for traversal
	MaxNeighbors int                         // Max neighbors per node per level
	EfSearch     int                         // Exploration factor during search
	EfConstruct  int                         // Exploration factor during construction
	DistanceFunc core.DistanceFunc           // Function used to measure similarity or distance
	random       *rand.Rand                  // Random generator used for level assignment
	mu           sync.RWMutex                // Read-write lock for concurrent access

	getRandomLevel func() int // Function to generate random levels
}

// NewHNSW initializes and returns a new HNSW index using the provided options.
// It selects the appropriate distance function based on the given distance metric.
func NewHNSW(opts HNSWOptions) *HNSW {
	var distFn core.DistanceFunc

	switch opts.DistanceMetric {
	case DistanceMetricEuclidean:
		distFn = core.EuclideanDistance
	case DistanceMetricDotProduct:
		distFn = core.DotProductSimilarity
	default:
		distFn = core.CosineSimilarity
	}

	return &HNSW{
		Nodes:        make(map[string]*core.VectorNode),
		MaxNeighbors: opts.MaxNeighbors,
		EfConstruct:  opts.EfConstruct,
		EfSearch:     opts.EfSearch,
		DistanceFunc: distFn,
		random:       rand.New(rand.NewSource(time.Now().UnixNano())),

		getRandomLevel: func() int {
			const prob = 1 / math.E
			level := 0

			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for r.Float64() < prob {
				level++
			}

			return level
		},
	}
}

// Insert adds a new vector into the HNSW index under the given ID.
// It assigns a random level to the new node and connects it to the graph.
// Returns an error if the node ID already exists.
func (h *HNSW) Insert(id string, vector []float32) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.Nodes[id]; exists {
		return errors.New("node already exists")
	}

	level := h.getRandomLevel()
	node := &core.VectorNode{
		ID:     id,
		Vector: vector,
		Level:  level,
		Links:  make(map[int][]*core.VectorNode),
	}

	h.Nodes[id] = node

	// First node becomes the entry point
	if h.EntryPoint == nil {
		h.EntryPoint = node
		h.MaxLevel = level
		return nil
	}

	// Greedy search from top level down to one above new node's level
	entry := h.EntryPoint
	for l := h.MaxLevel; l > level; l-- {
		entry = h.greedySearch(entry, vector, l)
	}

	// Connect in all levels from 0 to new node's level
	for l := min(level, h.MaxLevel); l >= 0; l-- {
		candidates := h.searchLayer(entry, vector, h.EfConstruct, l)
		neighbors := h.selectTopK(candidates, vector, h.MaxNeighbors)
		node.Links[l] = neighbors

		for _, n := range neighbors {
			n.Links[l] = append(n.Links[l], node)
		}
	}

	if level > h.MaxLevel {
		h.EntryPoint = node
		h.MaxLevel = level
	}

	return nil
}

// Search performs an approximate nearest neighbor search for the given query vector.
// It returns the top `k` closest nodes from the index using `EfSearch` as the exploration factor.
// Returns an error if the index is empty.
func (h *HNSW) Search(query []float32, topK int) ([]*core.VectorNode, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.EntryPoint == nil {
		return nil, errors.New("index is empty")
	}

	entry := h.EntryPoint
	for l := h.MaxLevel; l > 0; l-- {
		entry = h.greedySearch(entry, query, l)
	}

	candidates := h.searchLayer(entry, query, h.EfSearch, 0)
	return h.selectTopK(candidates, query, topK), nil
}

// greedySearch performs a greedy search starting from the given entry node at a specific level.
// It returns the node that is most similar (or closest) to the query at that level.
func (h *HNSW) greedySearch(entry *core.VectorNode, query []float32, level int) *core.VectorNode {
	best := entry
	bestScore := h.DistanceFunc(query, best.Vector)
	changed := true

	for changed {
		changed = false

		for _, neighbor := range best.Links[level] {
			score := h.DistanceFunc(query, neighbor.Vector)

			if score > bestScore {
				best = neighbor
				bestScore = score
				changed = true
			}
		}
	}

	return best
}

// searchLayer performs a breadth-first exploration of nodes at a given level,
// starting from the given entry node, and collects up to `ef` candidates.
func (h *HNSW) searchLayer(entry *core.VectorNode, query []float32, ef int, level int) []*core.VectorNode {
	visited := make(map[string]bool)

	pq := make(priorityQueue, 0)
	pq = pq.Push(entry, h.DistanceFunc(query, entry.Vector))

	visited[entry.ID] = true

	for i := range pq {
		curr := pq[i].node
		for _, neighbor := range curr.Links[level] {
			if visited[neighbor.ID] {
				continue
			}

			visited[neighbor.ID] = true
			score := h.DistanceFunc(query, neighbor.Vector)
			pq = pq.Push(neighbor, score)
		}

		if len(pq) > ef {
			pq = pq.Truncate(ef)
		}
	}

	return pq.Nodes()
}

// selectTopK sorts the given list of nodes by similarity to the query
// and returns the top `k` most similar nodes.
func (h *HNSW) selectTopK(nodes []*core.VectorNode, query []float32, k int) []*core.VectorNode {
	sort.Slice(nodes, func(i, j int) bool {
		return h.DistanceFunc(query, nodes[i].Vector) > h.DistanceFunc(query, nodes[j].Vector)
	})

	if len(nodes) > k {
		return nodes[:k]
	}

	return nodes
}
