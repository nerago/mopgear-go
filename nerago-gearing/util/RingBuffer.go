package util

type RingBuffer[T any] struct {
	array                 []T
	readIndex, writeIndex int
}

func RingBuffer_Create[T any](size int, initialEntry T) RingBuffer[T] {
	array := make([]T, size)
	array[0] = initialEntry
	return RingBuffer[T]{array: array, readIndex: 0, writeIndex: 1}
}

func (ring *RingBuffer[T]) Write(value T) {
	if ring.readIndex == ring.writeIndex {
		ring.readIndex = (ring.readIndex + 1) % len(ring.array)
	}

	ring.array[ring.writeIndex] = value
	ring.writeIndex = (ring.writeIndex + 1) % len(ring.array)
}

func (ring *RingBuffer[T]) ReadOldest() T {
	return ring.array[ring.readIndex]
}