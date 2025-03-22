package hnsw

const (
	DistanceMetricCosine     uint8 = 1
	DistanceMetricEuclidean  uint8 = 1 << 1 // 2
	DistanceMetricDotProduct uint8 = 1 << 2 // 4
)

type HNSWOptions struct {
	MaxNeighbors   int
	EfSearch       int
	EfConstruct    int
	DistanceMetric uint8
}
