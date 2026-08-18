package util_collection

import (
	"iter"
	"sync"
)

type linkedQueueNode[E any] struct {
	next, prev *linkedQueueNode[E]
	value      E
}

type LinkedQueue[E any] struct {
	root linkedQueueNode[E]
	size int
	pool sync.Pool
}

var _ IDeque[int] = &LinkedQueue[int]{}

func (q *LinkedQueue[E]) Clear() {
	var nilValue E
	root := &q.root
	for node := root.next; node != root; {
		next := node.next
		node.next = nil
		node.prev = nil
		node.value = nilValue
		q.pool.Put(node)
		node = next
	}
	q.root.next = &q.root
	q.root.prev = &q.root
	q.size = 0
}

func (q *LinkedQueue[E]) Size() int {
	return q.size
}

func (q *LinkedQueue[E]) IsEmpty() bool {
	return q.root.next == &q.root
}

func (q *LinkedQueue[E]) GetFirst() (E, bool) {
	node := q.root.next
	if node == &q.root || node == nil {
		var nilValue E
		return nilValue, false
	} else {
		return node.value, true
	}
}

func (q *LinkedQueue[E]) GetLast() (E, bool) {
	node := q.root.prev
	if node == &q.root || node == nil {
		var nilValue E
		return nilValue, false
	} else {
		return node.value, true
	}
}

func (q *LinkedQueue[E]) InsertFirst(value E) {
	node := q.pool.Get().(*linkedQueueNode[E])
	if node == nil {
		node = new(linkedQueueNode[E])
	}
	node.next = q.root.next
	node.prev = &q.root
	node.value = value
	q.root.next = node
	node.next.prev = node
	q.size++
}

func (q *LinkedQueue[E]) AppendLast(value E) {
	node := q.pool.Get().(*linkedQueueNode[E])
	if node == nil {
		node = new(linkedQueueNode[E])
	}
	node.next = &q.root
	node.prev = q.root.prev
	node.value = value
	q.root.prev = node
	node.prev.next = node
	q.size++
}

func (q *LinkedQueue[E]) deleteNode(node *linkedQueueNode[E]) {
	var nilValue E

	node.prev.next = node.next
	node.next.prev = node.prev

	node.next = nil
	node.prev = nil
	node.value = nilValue
	q.pool.Put(node)

	q.size--
}

func (q *LinkedQueue[E]) RemoveFirstAndReturn() (E, bool) {
	var nilValue E
	node := q.root.next
	if node == &q.root || node == nil {
		return nilValue, false
	} else {
		value := node.value
		q.deleteNode(node)
		return value, true
	}
}

func (q *LinkedQueue[E]) RemoveLastAndReturn() (E, bool) {
	var nilValue E
	node := q.root.prev
	if node == &q.root || node == nil {
		return nilValue, false
	} else {
		value := node.value
		q.deleteNode(node)
		return value, true
	}
}

func (q *LinkedQueue[E]) Get(targetIndex int) E {
	root := &q.root
	currIndex := 0
	for node := root.next; node != root; node = node.next {
		if currIndex == targetIndex {
			return node.value
		}
		currIndex++
	}
	panic("invalid index")
}

func (q *LinkedQueue[E]) ContainsFunc(predicate func(*E) bool) bool {
	root := &q.root
	for node := root.next; node != root; node = node.next {
		if predicate(&node.value) {
			return true
		}
	}
	return false
}

func (q *LinkedQueue[E]) ContainsFuncNoPointer(predicate func(E) bool) bool {
	root := &q.root
	for node := root.next; node != root; node = node.next {
		if predicate(node.value) {
			return true
		}
	}
	return false
}

func (q *LinkedQueue[E]) FilterFunc(keep func(*E) bool) {
	root := &q.root
	for node := root.next; node != root; {
		next := node.next
		if !keep(&node.value) {
			q.deleteNode(node)
		}
		node = next
	}
}

func (q *LinkedQueue[E]) FilterFuncNoPointer(keep func(E) bool) {
	root := &q.root
	for node := root.next; node != root; {
		next := node.next
		if !keep(node.value) {
			q.deleteNode(node)
		}
		node = next
	}
}

func (q *LinkedQueue[E]) RemoveDuplicatesFunc(equals func(a, b *E) bool) {
	root := &q.root
	for a := root.next; a != root; a = a.next {
		for b := a.next; b != root; {
			next := b.next
			if equals(&a.value, &b.value) {
				q.deleteNode(b)
			}
			b = next
		}
	}
}

func (q *LinkedQueue[E]) SeqIndexAndValues() iter.Seq2[int, E] {
	return func(yield func(int, E) bool) {
		root := &q.root
		currIndex := 0
		for node := root.next; node != root; node = node.next {
			if !yield(currIndex, node.value) {
				return
			}
			currIndex++
		}
	}
}

func (q *LinkedQueue[E]) SeqValues() iter.Seq[E] {
	return func(yield func(E) bool) {
		root := &q.root
		for node := root.next; node != root; node = node.next {
			if !yield(node.value) {
				return
			}
		}
	}
}

type linkedQueueTraversalEntry[E any] struct {
	node   *linkedQueueNode[E]
	index  int
	delete func()
}

func (qe linkedQueueTraversalEntry[E]) Value() E {
	return qe.node.value
}

func (qe linkedQueueTraversalEntry[E]) SetValue(value E) {
	qe.node.value = value
}

func (qe linkedQueueTraversalEntry[E]) Index() int {
	return qe.index
}

func (qe linkedQueueTraversalEntry[E]) Delete() {
	qe.delete()
}

func (q *LinkedQueue[E]) SeqListTraversalEntry() iter.Seq[ListTraversalEntry[E]] {
	return func(yield func(ListTraversalEntry[E]) bool) {
		root := &q.root
		currIndex := 0
		for node := root.next; node != root; {
			next := node.next
			if !yield(linkedQueueTraversalEntry[E]{
				node:   node,
				index:  currIndex,
				delete: func() { q.deleteNode(node) },
			}) {
				return
			}
			node = next
			currIndex++
		}
	}
}
