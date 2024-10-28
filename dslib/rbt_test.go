package dslib

import (
	"testing"
	"time"
)

func verifyRedBlackProperties(t *testing.T, node *RBTreeNode) (int, bool) {
	if node == nil {
		return 0, true
	}

	if node.color != red && node.color != black {
		t.Error("Invalid color found")
		return 0, false
	}

	if node.parent == nil && node.color != black {
		t.Error("Root must be black")
		return 0, false
	}

	if node.color == red {
		if node.left != nil && node.left.color == red {
			t.Error("Red node cannot have red children")
			return 0, false
		}
		if node.right != nil && node.right.color == red {
			t.Error("Red node cannot have red children")
			return 0, false
		}
	}

	if node.left != nil && node.left.parent != node {
		t.Error("Invalid parent pointer in left child")
		return 0, false
	}
	if node.right != nil && node.right.parent != node {
		t.Error("Invalid parent pointer in right child")
		return 0, false
	}

	leftBlackHeight, leftValid := verifyRedBlackProperties(t, node.left)
	rightBlackHeight, rightValid := verifyRedBlackProperties(t, node.right)

	if !leftValid || !rightValid {
		return 0, false
	}

	if leftBlackHeight != rightBlackHeight {
		t.Error("Black height mismatch")
		return 0, false
	}

	currentBlackHeight := leftBlackHeight
	if node.color == black {
		currentBlackHeight++
	}

	return currentBlackHeight, true
}

func TestNewRBTree(t *testing.T) {
	tree := NewRBTree()
	if tree == nil {
		t.Fatal("NewRBTree() returned nil")
	}
	if tree.Root != nil {
		t.Fatal("New tree should have nil root")
	}
	if tree.GetSize() != 0 {
		t.Fatal("New tree should have size 0")
	}
}

func TestInsert(t *testing.T) {
	tree := NewRBTree()

	tree.Insert("A", 1, time.Now().Unix()+1000, 0)
	if tree.Root == nil {
		t.Error("Root should not be nil after insertion")
	}
	if tree.Root.color != black {
		t.Error("Root should be black")
	}
	if tree.GetSize() != 1 {
		t.Error("Tree size should be 1")
	}

	testData := []struct {
		key   string
		value int
	}{
		{"B", 2},
		{"C", 3},
		{"D", 4},
		{"E", 5},
	}

	for _, data := range testData {
		tree.Insert(data.key, data.value, time.Now().Unix()+1000, 0)
	}

	if tree.GetSize() != 5 {
		t.Error("Tree size incorrect after multiple insertions")
	}

	verifyRedBlackProperties(t, tree.Root)
}

func TestGet(t *testing.T) {
	tree := NewRBTree()
	testData := map[string]int{
		"A": 1,
		"B": 2,
		"C": 3,
	}

	for k, v := range testData {
		tree.Insert(k, v, time.Now().Unix()+1000, 1)
	}

	for k, v := range testData {
		found, value, exp, state := tree.Get(k)
		if !found {
			t.Errorf("Key %s not found", k)
		}
		if value != v || state != 1 || exp < time.Now().Unix()+999 {
			t.Errorf("Wrong object values for key %s: [%v, %v, %v]", k, value, state, exp)
		}
	}

	found, _, _, _ := tree.Get("Z")
	if found {
		t.Error("Found non-existent key")
	}
}

func TestDelete(t *testing.T) {
	tree := NewRBTree()
	testData := []string{"F", "B", "G", "A", "D", "I", "C", "E", "H"}

	for _, key := range testData {
		tree.Insert(key, key, time.Now().Unix()+1000, 0)
	}

	initialSize := tree.GetSize()

	testCases := []struct {
		key          string
		shouldDelete bool
	}{
		{"A", true},  // Leaf node
		{"D", true},  // Node with one child
		{"B", true},  // Node with two children
		{"Z", false}, // Non-existent node
	}

	for _, tc := range testCases {
		deleted := tree.Delete(tc.key)
		if deleted != tc.shouldDelete {
			t.Errorf("Delete(%s) returned %v, expected %v", tc.key, deleted, tc.shouldDelete)
		}
		if deleted {
			initialSize--
			if tree.GetSize() != initialSize {
				t.Errorf("Wrong size after deletion: got %d, want %d", tree.GetSize(), initialSize)
			}
			found, _, _, _ := tree.Get(tc.key)
			if found {
				t.Errorf("Key %s still exists after deletion", tc.key)
			}
		}
		verifyRedBlackProperties(t, tree.Root)
	}
}

func TestGetAllRecords(t *testing.T) {
	tree := NewRBTree()
	testData := []struct {
		key   string
		value int
	}{
		{"D", 4},
		{"B", 2},
		{"F", 6},
		{"A", 1},
		{"C", 3},
		{"E", 5},
	}

	for _, data := range testData {
		tree.Insert(data.key, data.value, time.Now().Unix()+1000, 0)
	}

	records := tree.GetAllRecords()

	if len(records) != len(testData) {
		t.Errorf("GetAllRecords returned wrong number of records: got %d, want %d",
			len(records), len(testData))
	}

	for i := 1; i < len(records); i++ {
		if records[i].Key <= records[i-1].Key {
			t.Error("Records not in order")
		}
	}
}

func TestRotations(t *testing.T) {
	tree := NewRBTree()

	tree.Insert("B", 2, time.Now().Unix()+1000, 0)
	tree.Insert("A", 1, time.Now().Unix()+1000, 0)
	tree.Insert("C", 3, time.Now().Unix()+1000, 0)
	tree.Insert("D", 4, time.Now().Unix()+1000, 0)
	tree.Insert("E", 5, time.Now().Unix()+1000, 0)

	if tree.Root.key != "B" {
		t.Error("Unexpected root after left rotation")
	}

	tree = NewRBTree()
	tree.Insert("D", 4, time.Now().Unix()+1000, 0)
	tree.Insert("C", 3, time.Now().Unix()+1000, 0)
	tree.Insert("E", 5, time.Now().Unix()+1000, 0)
	tree.Insert("B", 2, time.Now().Unix()+1000, 0)
	tree.Insert("A", 1, time.Now().Unix()+1000, 0)

	if tree.Root.key != "D" {
		t.Error("Unexpected root after right rotation")
	}

	verifyRedBlackProperties(t, tree.Root)
}

func TestEdgeCases(t *testing.T) {
	tree := NewRBTree()

	if found, _, _, _ := tree.Get("A"); found {
		t.Error("Get on empty tree should return false")
	}
	if tree.Delete("A") {
		t.Error("Delete on empty tree should return false")
	}
	if len(tree.GetAllRecords()) != 0 {
		t.Error("GetAllRecords on empty tree should return empty slice")
	}

	tree.Insert("A", 1, time.Now().Unix()+1000, 0)
	tree.Insert("A", 2, time.Now().Unix()+1000, 0)
	_, value, _, _ := tree.Get("A")
	if value != 2 {
		t.Error("Duplicate key should overwrite old value")
	}

	if !tree.Delete("A") {
		t.Error("Failed to delete root")
	}
	if tree.Root != nil {
		t.Error("Root should be nil after deleting last node")
	}
}
