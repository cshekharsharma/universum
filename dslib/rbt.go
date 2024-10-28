package dslib

import (
	"universum/entity"
	"universum/utils"
)

// red-black Tree Constants
const (
	red   = true
	black = false
)

// RBTreeNode represents a node in the red-black Tree
type RBTreeNode struct {
	key    string
	value  interface{}
	expiry int64
	state  uint8
	color  bool
	left   *RBTreeNode
	right  *RBTreeNode
	parent *RBTreeNode
}

// RBTree represents the red-black Tree structure
type RBTree struct {
	Root *RBTreeNode
	size int64
}

// NewRBTree creates a new red-black Tree
func NewRBTree() *RBTree {
	return &RBTree{}
}

// Insert inserts a key-value pair into the red-black Tree
func (t *RBTree) Insert(key string, value interface{}, expiry int64, state uint8) {
	if existingNode := t.findNode(key); existingNode != nil {
		existingNode.value = value
		existingNode.expiry = expiry
		existingNode.state = state
		return
	}

	newNode := &RBTreeNode{
		key:    key,
		value:  value,
		expiry: expiry,
		state:  state,
		color:  red,
	}

	t.insertNode(newNode)
	t.fixInsertion(newNode)
	t.size++
}

// insertNode handles the standard BST insertion logic
func (t *RBTree) insertNode(newNode *RBTreeNode) {
	var parent *RBTreeNode
	node := t.Root

	for node != nil {
		parent = node
		if newNode.key < node.key {
			node = node.left
		} else {
			node = node.right
		}
	}

	newNode.parent = parent

	if parent == nil {
		t.Root = newNode
	} else if newNode.key < parent.key {
		parent.left = newNode
	} else {
		parent.right = newNode
	}
}

// fixInsertion ensures the red-black Tree properties are maintained after insertion
func (t *RBTree) fixInsertion(node *RBTreeNode) {
	for node != t.Root && node.parent.color {
		if node.parent == node.parent.parent.left {
			uncle := node.parent.parent.right

			if uncle != nil && uncle.color {
				node.parent.color = black
				uncle.color = black
				node.parent.parent.color = red
				node = node.parent.parent

			} else {
				if node == node.parent.right {
					node = node.parent
					t.rotateLeft(node)
				}

				node.parent.color = black
				node.parent.parent.color = red
				t.rotateRight(node.parent.parent)
			}
		} else {
			uncle := node.parent.parent.left

			if uncle != nil && uncle.color {
				node.parent.color = black
				uncle.color = black
				node.parent.parent.color = red
				node = node.parent.parent

			} else {
				if node == node.parent.left {
					node = node.parent
					t.rotateRight(node)
				}

				node.parent.color = black
				node.parent.parent.color = red
				t.rotateLeft(node.parent.parent)
			}
		}
	}

	t.Root.color = black
}

// rotateLeft rotates the subtree rooted at node to the left
func (t *RBTree) rotateLeft(node *RBTreeNode) {
	rightNode := node.right
	node.right = rightNode.left

	if rightNode.left != nil {
		rightNode.left.parent = node
	}

	rightNode.parent = node.parent

	if node.parent == nil {
		t.Root = rightNode
	} else if node == node.parent.left {
		node.parent.left = rightNode
	} else {
		node.parent.right = rightNode
	}

	rightNode.left = node
	node.parent = rightNode
}

// rotateRight rotates the subtree rooted at node to the right
func (t *RBTree) rotateRight(node *RBTreeNode) {
	leftNode := node.left
	node.left = leftNode.right

	if leftNode.right != nil {
		leftNode.right.parent = node
	}

	leftNode.parent = node.parent

	if node.parent == nil {
		t.Root = leftNode
	} else if node == node.parent.right {
		node.parent.right = leftNode
	} else {
		node.parent.left = leftNode
	}

	leftNode.right = node
	node.parent = leftNode
}

// Delete removes a node from the tree by key
func (t *RBTree) Delete(key string) bool {
	node := t.findNode(key)
	if node == nil {
		return false
	}

	t.deleteNode(node)
	t.size--

	return true
}

func (t *RBTree) findNode(key string) *RBTreeNode {
	node := t.Root

	for node != nil {
		if key == node.key {
			return node
		} else if key < node.key {
			node = node.left
		} else {
			node = node.right
		}
	}

	return nil
}

// deleteNode deletes a node from the red-black Tree and rebalances it
func (t *RBTree) deleteNode(node *RBTreeNode) {
	var child, parent *RBTreeNode
	var color bool

	if node.left != nil && node.right != nil {
		successor := node.right
		for successor.left != nil {
			successor = successor.left
		}

		node.key = successor.key
		node.value = successor.value
		node.expiry = successor.expiry
		node.state = successor.state
		node = successor
	}

	if node.left == nil {
		child = node.right
	} else {
		child = node.left
	}

	parent = node.parent
	color = node.color

	if child != nil {
		child.parent = parent
	}

	if parent == nil {
		t.Root = child
	} else if node == parent.left {
		parent.left = child
	} else {
		parent.right = child
	}

	if !color {
		t.fixDeletion(child, parent)
	}

	// If we've deleted the last node, ensure Root is nil
	if t.size == 1 {
		t.Root = nil
	}
}

func (t *RBTree) fixDeletion(node, parent *RBTreeNode) {
	for node != t.Root && (node == nil || !node.color) {

		if node == parent.left {
			sibling := parent.right

			if sibling.color {
				sibling.color = black
				parent.color = red

				t.rotateLeft(parent)
				sibling = parent.right
			}

			if (sibling.left == nil || !sibling.left.color) &&
				(sibling.right == nil || !sibling.right.color) {
				sibling.color = red
				node = parent
				parent = node.parent
			} else {
				if sibling.right == nil || !sibling.right.color {
					sibling.left.color = black
					sibling.color = red

					t.rotateRight(sibling)
					sibling = parent.right
				}

				sibling.color = parent.color
				parent.color = black

				if sibling.right != nil {
					sibling.right.color = black
				}

				t.rotateLeft(parent)
				node = t.Root
			}

		} else {
			sibling := parent.left
			if sibling.color {
				sibling.color = black
				parent.color = red

				t.rotateRight(parent)
				sibling = parent.left
			}

			if (sibling.left == nil || !sibling.left.color) &&
				(sibling.right == nil || !sibling.right.color) {
				sibling.color = red
				node = parent
				parent = node.parent

			} else {
				if sibling.left == nil || !sibling.left.color {
					sibling.right.color = black
					sibling.color = red

					t.rotateLeft(sibling)
					sibling = parent.left
				}

				sibling.color = parent.color
				parent.color = black

				if sibling.left != nil {
					sibling.left.color = black
				}

				t.rotateRight(parent)
				node = t.Root
			}
		}
	}

	if node != nil {
		node.color = black
	}
}

// Get retrieves the value associated with a given key
func (t *RBTree) Get(key string) (bool, interface{}, int64, uint8) {
	node := t.Root

	for node != nil {
		if key == node.key {
			return true, node.value, node.expiry, node.state
		} else if key < node.key {
			node = node.left
		} else {
			node = node.right
		}
	}

	return false, nil, 0, 0
}

// GetSize returns the size of the red-black Tree
func (t *RBTree) GetSize() int64 {
	return t.size
}

// GetAllRecords retrieves all non-expired records from the red-black tree
// in sorted order, returning a slice of `entity.RecordKV`.
func (t *RBTree) GetAllRecords() []*entity.RecordKV {
	var result []*entity.RecordKV
	currentTime := utils.GetCurrentEPochTime()
	t.inOrderCollect(t.Root, &result, currentTime)
	return result
}

// inOrderCollect performs an in-order traversal of the tree,
// appending only non-expired nodes to the result slice.
func (t *RBTree) inOrderCollect(node *RBTreeNode, result *[]*entity.RecordKV, currentTime int64) {
	if node == nil {
		return
	}

	t.inOrderCollect(node.left, result, currentTime)

	if node.expiry > currentTime {
		*result = append(*result, &entity.RecordKV{
			Key: node.key,
			Record: &entity.ScalarRecord{
				Value:  node.value,
				Expiry: node.expiry,
				State:  node.state,
			},
		})
	}

	t.inOrderCollect(node.right, result, currentTime)
}
