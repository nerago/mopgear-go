package util_collection

import (
	"iter"
)

// TODO test this properly

type QueueStackFilo[T any] struct {
	array []T
	size  int
}

var _ IQueueLiteInspectable[int] = &QueueStackFilo[int]{}

func QueueStackFilo_Create[T any](allocateSize int) IQueueLite[T] {
	array := make([]T, allocateSize)
	return &QueueStackFilo[T]{array: array, size: 0}
}

func (stack *QueueStackFilo[T]) Clear() {
	var nilValue T
	for i := range stack.size {
		stack.array[i] = nilValue
	}
	stack.size = 0
}

func (stack *QueueStackFilo[T]) IsEmpty() bool {
	return stack.size == 0
}

func (stack *QueueStackFilo[T]) Size() int {
	return stack.size
}

func (stack *QueueStackFilo[T]) Push(value T) {
	sz := stack.size
	if sz < len(stack.array) {
		stack.size = sz + 1
		stack.array[sz] = value
	} else if len(stack.array) > 0 {
		newArray := make([]T, len(stack.array)*2)
		copy(newArray, stack.array)
		newArray[sz] = value
		stack.size = sz + 1
		stack.array = newArray
	} else {
		stack.size = 1
		stack.array = make([]T, 8)
		stack.array[sz] = value
	}
}

func (stack *QueueStackFilo[T]) Pop() (T, bool) {
	var nilValue T
	readIndex := stack.size - 1
	if readIndex >= 0 {
		value := stack.array[readIndex]
		stack.array[readIndex] = nilValue
		stack.size = readIndex
		return value, true
	} else {
		return nilValue, false
	}
}

func (stack *QueueStackFilo[T]) SeqValues() iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := range stack.size {
			if !yield(stack.array[i]) {
				return
			}
		}
	}
}

func (stack *QueueStackFilo[T]) UpdateContents(apply func([]T) []T) {
	stack.resetFromSlice(apply(stack.array[0:stack.size]))
}

func (stack *QueueStackFilo[T]) ExamineContents(apply func([]T)) {
	apply(stack.array[0:stack.size])
}

func (stack *QueueStackFilo[T]) resetFromSlice(content []T) {
	if len(content) > len(stack.array) {
		stack.array = make([]T, len(content)*2)
	}

	copy(stack.array, content)
	stack.size = len(content)

	var nilValue T
	for i := stack.size; i < len(stack.array); i++ {
		stack.array[i] = nilValue
	}
}
