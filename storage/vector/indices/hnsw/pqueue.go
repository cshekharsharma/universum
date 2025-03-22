package hnsw

import (
	"sort"
	"universum/storage/vector/core"
)

type candidate struct {
	node  *core.VectorNode
	score float64
}

type priorityQueue []candidate

func (pq priorityQueue) Len() int {
	return len(pq)
}

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].score > pq[j].score // max-heap
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *priorityQueue) Push(node *core.VectorNode, score float64) priorityQueue {
	*pq = append(*pq, candidate{node, score})
	sort.Sort(*pq)
	return *pq
}

func (pq *priorityQueue) Pop() candidate {
	c := (*pq)[0]
	*pq = (*pq)[1:]
	return c
}

func (pq *priorityQueue) Truncate(n int) priorityQueue {
	if len(*pq) <= n {
		return *pq
	}

	*pq = (*pq)[:n]
	return *pq
}

func (pq *priorityQueue) Nodes() []*core.VectorNode {
	nodes := make([]*core.VectorNode, len(*pq))

	for i, c := range *pq {
		nodes[i] = c.node
	}

	return nodes
}
