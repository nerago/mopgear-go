package util_collection

import (
	"iter"
	"sync"
)

type QueueStackFilo[T any] struct {
	array []T
	top   int
}

func QueueStackFilo_Create[T any](allocateSize int) IQueueLite[T] {
	array := make([]T, allocateSize)
	return &QueueStackFilo[T]{array: array, top: 0}
}

func (stack *QueueStackFilo[T]) Clear() {
	var nilValue T
	for i := range stack.top {
		stack.array[i] = nilValue
	}
	stack.top = 0
}

func (stack *QueueStackFilo[T]) IsEmpty() bool {
	return stack.top == 0
}

func (stack *QueueStackFilo[T]) Size() int {
	return stack.top
}

func (stack *QueueStackFilo[T]) Push(value T) {
	writeIndex := stack.top
	if writeIndex < len(stack.array) {
		stack.top = writeIndex + 1
		stack.array[writeIndex] = value
	} else if len(stack.array) > 0 {
		newArray := make([]T, len(stack.array)*2)
		copy(newArray, stack.array)
		newArray[writeIndex] = value
		stack.top = writeIndex + 1
		stack.array = newArray
	} else {
		stack.top = 1
		stack.array = make([]T, 8)
		stack.array[writeIndex] = value
	}
}

func (stack *QueueStackFilo[T]) Pop() (T, bool) {
	var nilValue T
	readIndex := stack.top - 1
	if readIndex >= 0 {
		value := stack.array[readIndex]
		stack.array[readIndex] = nilValue
		stack.top = readIndex
		return value, true
	} else {
		return nilValue, false
	}
}

func (stack *QueueStackFilo[T]) SeqValues() iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := range stack.top {
			if !yield(stack.array[i]) {
				return
			}
		}
	}
}

func (stack *QueueStackFilo[T]) UpdateContents(apply func([]T) []T) {
	stack.resetFromSlice(apply(stack.array[0:stack.top]))
}

func (stack *QueueStackFilo[T]) ExamineContents(apply func([]T)) {
	apply(stack.array[0:stack.top])
}

func (stack *QueueStackFilo[T]) resetFromSlice(content []T) {
	if len(content) > len(stack.array) {
		stack.array = make([]T, len(content)*2)
	}

	copy(stack.array, content)
	stack.top = len(content)

	var nilValue T
	for i := stack.top; i < len(stack.array); i++ {
		stack.array[i] = nilValue
	}
}

type QueueStackFiloConcurrent[T any] struct {
	inner QueueStackFilo[T]
	mutex sync.Mutex
}

func (stack *QueueStackFiloConcurrent[T]) IsEmpty() bool {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	return stack.inner.IsEmpty()
}

func (stack *QueueStackFiloConcurrent[T]) Size() int {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	return stack.inner.Size()
}

func (stack *QueueStackFiloConcurrent[T]) Push(value T) {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	stack.inner.Push(value)
}

func (stack *QueueStackFiloConcurrent[T]) Pop() (T, bool) {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	return stack.inner.Pop()
}

type QueueStackFiloPoolParent[T any] struct {
	children    []*QueueStackFiloPoolChild[T]
	parentMutex sync.Mutex
}

func (parent *QueueStackFiloPoolParent[T]) MakeChild() *QueueStackFiloPoolChild[T] {
	parent.parentMutex.Lock()
	defer parent.parentMutex.Unlock()

	child := &QueueStackFiloPoolChild[T]{parent: parent}
	parent.children = append(parent.children, child)
	return child
}

type QueueStackFiloPoolChild[T any] struct {
	inner      QueueStackFilo[T]
	parent     *QueueStackFiloPoolParent[T]
	childMutex sync.Mutex
}

func (child *QueueStackFiloPoolChild[T]) CountLocal() int {
	child.childMutex.Lock()
	defer child.childMutex.Unlock()
	return child.inner.Size()
}

func (child *QueueStackFiloPoolChild[T]) Push(value T) {
	child.childMutex.Lock()
	child.inner.Push(value)
	child.childMutex.Unlock()
}

func (child *QueueStackFiloPoolChild[T]) Pop() (T, bool) {
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
			value, hasValue = other.inner.Pop()
			other.childMutex.Unlock()

			if hasValue {
				return value, true
			}
		}
	}

	var nilValue T
	return nilValue, false
}

func (child *QueueStackFiloPoolChild[T]) ExamineContents(apply func(content []T)) {
	child.childMutex.Lock()
	child.inner.ExamineContents(apply)
	child.childMutex.Unlock()
}
