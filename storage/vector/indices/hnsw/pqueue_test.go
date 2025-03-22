package hnsw

import (
	"testing"
	"universum/storage/vector/core"
)

func TestPriorityQueue_PushAndSort(t *testing.T) {
	var pq priorityQueue

	node1 := &core.VectorNode{ID: "n1"}
	node2 := &core.VectorNode{ID: "n2"}
	node3 := &core.VectorNode{ID: "n3"}

	pq = pq.Push(node1, 0.5)
	pq = pq.Push(node2, 0.8)
	pq = pq.Push(node3, 0.2)

	if pq[0].node.ID != "n2" || pq[1].node.ID != "n1" || pq[2].node.ID != "n3" {
		t.Errorf("PriorityQueue not sorted as max-heap: got order [%s %s %s]",
			pq[0].node.ID, pq[1].node.ID, pq[2].node.ID)
	}
}

func TestPriorityQueue_Pop(t *testing.T) {
	var pq priorityQueue
	pq = pq.Push(&core.VectorNode{ID: "a"}, 0.9)
	pq = pq.Push(&core.VectorNode{ID: "b"}, 0.4)

	top := pq.Pop()

	if top.node.ID != "a" {
		t.Errorf("Expected top node to be 'a', got %s", top.node.ID)
	}

	if len(pq) != 1 {
		t.Errorf("Expected queue length 1 after pop, got %d", len(pq))
	}
}

func TestPriorityQueue_Truncate(t *testing.T) {
	var pq priorityQueue
	pq = pq.Push(&core.VectorNode{ID: "x"}, 0.9)
	pq = pq.Push(&core.VectorNode{ID: "y"}, 0.7)
	pq = pq.Push(&core.VectorNode{ID: "z"}, 0.5)

	pq = pq.Truncate(2)

	if len(pq) != 2 {
		t.Errorf("Expected queue to be truncated to 2, got %d", len(pq))
	}
}

func TestPriorityQueue_Truncate_NoOp(t *testing.T) {
	var pq priorityQueue
	pq = pq.Push(&core.VectorNode{ID: "x"}, 0.1)
	pq = pq.Truncate(5)

	if len(pq) != 1 {
		t.Errorf("Expected no-op truncate to leave length 1, got %d", len(pq))
	}
}

func TestPriorityQueue_Nodes(t *testing.T) {
	var pq priorityQueue
	n1 := &core.VectorNode{ID: "1"}
	n2 := &core.VectorNode{ID: "2"}
	pq = pq.Push(n1, 0.3)
	pq = pq.Push(n2, 0.7)

	nodes := pq.Nodes()

	if len(nodes) != 2 {
		t.Fatalf("Expected 2 nodes, got %d", len(nodes))
	}
	if nodes[0].ID != "2" || nodes[1].ID != "1" {
		t.Errorf("Expected order [2, 1], got [%s, %s]", nodes[0].ID, nodes[1].ID)
	}
}
