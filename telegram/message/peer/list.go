package peer

import (
	"time"

	"github.com/nnqq/td/tg"
)

type nodeData struct {
	key       string
	value     tg.InputPeerClass
	expiresAt time.Time
}

type linkedNode struct {
	next, prev *linkedNode
	nodeData
}

// linkedList is a doubly linked list implementation.
// This implementation is highly inspired by container/list, but type safe.
type linkedList struct {
	root linkedNode
	len  int
}

// Front returns the first element of list l or nil if the list is empty.
func (l *linkedList) Front() *linkedNode {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// Back returns the last element of list l or nil if the list is empty.
func (l *linkedList) Back() *linkedNode {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// insert inserts e after at, increments l.len, and returns e.
func (l *linkedList) insert(e, at *linkedNode) *linkedNode {
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
	l.len++
	return e
}

func (l *linkedList) insertValue(v nodeData, at *linkedNode) *linkedNode {
	e := new(linkedNode)
	e.nodeData = v
	return l.insert(e, at)
}

// PushFront inserts a new element e with value v at the front of list l and returns e.
func (l *linkedList) PushFront(v nodeData) *linkedNode {
	l.lazyInit()
	return l.insertValue(v, &l.root)
}

// lazyInit lazily initializes a zero List value.
func (l *linkedList) lazyInit() {
	if l.root.next == nil {
		l.root.next = &l.root
		l.root.prev = &l.root
		l.len = 0
	}
}

// remove removes e from its list, decrements l.len, and returns e.
func (l *linkedList) remove(e *linkedNode) nodeData {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	l.len--

	value := e.nodeData
	return value
}

// move moves e to next to at and returns e.
func (l *linkedList) move(e, at *linkedNode) *linkedNode {
	if e == at {
		return e
	}
	e.prev.next = e.next
	e.next.prev = e.prev

	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e

	return e
}

// Remove removes e from l if e is an element of list l.
// It returns the element value e.Value.
// The element must not be nil.
func (l *linkedList) Remove(e *linkedNode) nodeData {
	return l.remove(e)
}

// MoveToFront moves element e to the front of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *linkedList) MoveToFront(e *linkedNode) {
	if l.root.next == e {
		return
	}
	// see comment in List.Remove about initialization of l
	l.move(e, &l.root)
}
