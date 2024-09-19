package dslib

import (
	"math/rand"
	"time"
)

const MaxLevel = 16
const MinString = ""

type SkipList struct {
	head  *Node
	level int
	size  int
	rand  *rand.Rand
}

type Node struct {
	key    string
	value  interface{}
	expiry int64
	next   []*Node
}

// NewNode creates a new Node for the skip list
func NewNode(key string, value interface{}, expiry int64, level int) *Node {
	return &Node{
		key:    key,
		value:  value,
		expiry: expiry,
		next:   make([]*Node, level),
	}
}

// NewSkipList initializes a new SkipList with the size field and dedicated random generator
func NewSkipList() *SkipList {
	return &SkipList{
		head:  NewNode(MinString, nil, 0, MaxLevel),
		level: 1,
		size:  0,
		rand:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Insert inserts a new element into the skip list or updates an existing one
func (sl *SkipList) Insert(key string, value interface{}, expiry int64) {
	update := make([]*Node, MaxLevel)
	current := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for current.next[i] != nil && current.next[i].key < key {
			current = current.next[i]
		}
		update[i] = current
	}

	current = current.next[0]
	if current != nil && current.key == key {
		current.value = value
		current.expiry = expiry
		return
	}

	level := sl.randomLevel()

	if level > sl.level {
		for i := sl.level; i < level; i++ {
			update[i] = sl.head
		}
		sl.level = level
	}

	newNode := NewNode(key, value, expiry, level)

	for i := 0; i < level; i++ {
		newNode.next[i] = update[i].next[i]
		update[i].next[i] = newNode
	}

	sl.size++
}

// Search returns the value for the specified key, if it exists, or nil if not found
func (sl *SkipList) Search(key string) (bool, interface{}, int64) {
	current := sl.head
	for i := sl.level - 1; i >= 0; i-- {
		for current.next[i] != nil && current.next[i].key < key {
			current = current.next[i]
		}
	}

	current = current.next[0]
	if current != nil && current.key == key {
		return true, current.value, current.expiry
	}

	return false, nil, 0
}

// Get retrieves a value from the skip list based on the given key
func (sl *SkipList) Get(key string) (bool, interface{}, int64) {
	return sl.Search(key)
}

// randomLevel generates a random level for a new node
func (sl *SkipList) randomLevel() int {
	level := 1
	for level < MaxLevel && sl.rand.Float64() < 0.5 {
		level++
	}
	return level
}

// Size returns the number of elements in the skip list
func (sl *SkipList) Size() int {
	return sl.size
}

// Remove deletes a node with the given key from the skip list
func (sl *SkipList) Remove(key string) bool {
	update := make([]*Node, MaxLevel)
	current := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for current.next[i] != nil && current.next[i].key < key {
			current = current.next[i]
		}
		update[i] = current
	}

	current = current.next[0]

	if current == nil || current.key != key {
		return false
	}

	for i := 0; i < sl.level; i++ {
		if update[i].next[i] != current {
			break
		}
		update[i].next[i] = current.next[i]
	}

	for sl.level > 1 && sl.head.next[sl.level-1] == nil {
		sl.level--
	}

	sl.size--
	return true
}
