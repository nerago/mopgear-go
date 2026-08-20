package util_collection

import (
	"iter"
	"sync"
)

type QueueStackFiloConcurrent[T any] struct {
	inner QueueStackFilo[T]
	mutex sync.Mutex
}

var _ IQueueLite[int] = &QueueStackFiloConcurrent[int]{}

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

func (stack *QueueStackFiloConcurrent[T]) Clear() {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	stack.inner.Clear()
}

func (stack *QueueStackFiloConcurrent[T]) SeqValues() iter.Seq[T] {
	return func(yield func(T) bool) {
		stack.mutex.Lock()
		defer stack.mutex.Unlock()
		for value := range stack.inner.SeqValues() {
			if !yield(value) {
				return
			}
		}
	}
}
