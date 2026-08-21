package util_collection

import (
	"sync"
)

type QueueStackStealingPool[T any] struct {
	children    []*QueueStackPoolChild[T]
	parentMutex sync.Mutex
}

func (parent *QueueStackStealingPool[T]) MakeChild() *QueueStackPoolChild[T] {
	parent.parentMutex.Lock()
	defer parent.parentMutex.Unlock()

	child := &QueueStackPoolChild[T]{parent: parent}
	parent.children = append(parent.children, child)
	return child
}

type QueueStackPoolChild[T any] struct {
	inner      QueueStackFilo[T]
	parent     *QueueStackStealingPool[T]
	childMutex sync.Mutex
}

var _ IQueueThreadLocal[int] = &QueueStackPoolChild[int]{}

func (child *QueueStackPoolChild[T]) CountLocal() int {
	child.childMutex.Lock()
	defer child.childMutex.Unlock()
	return child.inner.Size()
}

func (child *QueueStackPoolChild[T]) Push(value T) {
	child.childMutex.Lock()
	child.inner.Push(value)
	child.childMutex.Unlock()
}

func (child *QueueStackPoolChild[T]) Pop() (T, bool) {
	// make sure  to let go of our own lock in case another thread holding parent lock and trying to acquire our child lock = deadlock
	child.childMutex.Lock()
	value, hasValue := child.inner.Pop()
	child.childMutex.Unlock()
	if hasValue {
		return value, true
	}

	child.parent.parentMutex.Lock()
	defer child.parent.parentMutex.Unlock()

	for _, other := range child.parent.children {
		if other != child {
			other.childMutex.Lock()
			value, hasValue = other.inner.PopBottom()
			other.childMutex.Unlock()

			if hasValue {
				return value, true
			}
		}
	}

	var nilValue T
	return nilValue, false
}

func (child *QueueStackPoolChild[T]) ExamineContents(apply func(content []T)) {
	child.childMutex.Lock()
	child.inner.ExamineContents(apply)
	child.childMutex.Unlock()
}
