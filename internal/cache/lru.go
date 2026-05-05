package cache

// LRUNode represents a node in the doubly linked list.
type LRUNode struct {
	Key  string
	prev *LRUNode
	next *LRUNode
}

// LRUList is a simple doubly linked list used for tracking item recency.
type LRUList struct {
	head *LRUNode
	tail *LRUNode
}

// NewLRUList creates a new LRU list.
func NewLRUList() *LRUList {
	return &LRUList{}
}

// PushFront adds a new key to the front of the list and returns the node.
func (l *LRUList) PushFront(key string) *LRUNode {
	node := &LRUNode{Key: key}
	if l.head == nil {
		l.head = node
		l.tail = node
	} else {
		node.next = l.head
		l.head.prev = node
		l.head = node
	}
	return node
}

// MoveToFront moves an existing node to the front of the list.
func (l *LRUList) MoveToFront(node *LRUNode) {
	if node == nil || node == l.head {
		return
	}

	l.Remove(node)
	
	node.prev = nil
	node.next = l.head
	if l.head != nil {
		l.head.prev = node
	}
	l.head = node
	if l.tail == nil {
		l.tail = node
	}
}

// Remove removes a specific node from the list.
func (l *LRUList) Remove(node *LRUNode) {
	if node == nil {
		return
	}

	if node.prev != nil {
		node.prev.next = node.next
	} else {
		l.head = node.next // Node was head
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		l.tail = node.prev // Node was tail
	}

	// Clean up pointers to avoid memory leaks
	node.prev = nil
	node.next = nil
}

// RemoveTail removes the least recently used node (tail) and returns it.
func (l *LRUList) RemoveTail() *LRUNode {
	if l.tail == nil {
		return nil
	}

	node := l.tail
	l.Remove(node)
	return node
}
