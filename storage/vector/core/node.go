package core

type VectorNode struct {
	ID     string
	Vector []float32
	Level  int                   // Max level this node appears in
	Links  map[int][]*VectorNode // level -> list of neighbors
}
