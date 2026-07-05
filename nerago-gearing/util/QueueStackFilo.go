package util

type QueueStackFilo[T any] struct {
	array []T
	top   int
}

func QueueStackFilo_Create[T any](allocateSize int) Queue[T] {
	array := make([]T, allocateSize)
	return &QueueStackFilo[T]{array: array, top: 0}
}

func (stack *QueueStackFilo[T]) IsEmpty() bool {
	return stack.top == 0
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
