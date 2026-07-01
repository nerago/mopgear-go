package util

type QueueStackFilo[T any] struct {
	array []T
	top   int
}

func QueueStackFilo_Create[T any](allocateSize int) Queue[T] {
	array := make([]T, allocateSize)
	return &QueueStackFilo[T]{array: array, top: 0}
}

func (ring *QueueStackFilo[T]) IsEmpty() bool {
	return ring.top == 0
}

func (ring *QueueStackFilo[T]) Push(value T) {
	writeIndex := ring.top
	if writeIndex < len(ring.array) {
		ring.top = writeIndex + 1
		ring.array[writeIndex] = value
	} else if len(ring.array) > 0 {
		newArray := make([]T, len(ring.array)*2)
		copy(newArray, ring.array)
		newArray[writeIndex] = value
		ring.top = writeIndex + 1
		ring.array = newArray
	} else {
		ring.top = 1
		ring.array = make([]T, 8)
	}
}

func (ring *QueueStackFilo[T]) Pop() (T, bool) {
	var nilValue T
	readIndex := ring.top - 1
	if readIndex >= 0 {
		value := ring.array[readIndex]
		ring.array[readIndex] = nilValue
		ring.top = readIndex
		return value, true
	} else {
		return nilValue, false
	}
}
