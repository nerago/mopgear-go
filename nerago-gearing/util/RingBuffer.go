package util

type RingBuffer[T any] struct {
	array                 []T
	readOldIndex, readNewIndex int
	writeIndex int
}

func RingBuffer_Create[T any](size int, initialEntry T) RingBuffer[T] {
	array := make([]T, size)
	array[0] = initialEntry
	return RingBuffer[T]{array: array, readOldIndex: 0, readNewIndex: 0, writeIndex: 1}
}

func (ring *RingBuffer[T]) Write(value T) {
	if ring.readOldIndex == ring.writeIndex {
		ring.readOldIndex = (ring.readOldIndex + 1) % len(ring.array)
	}

	ring.readNewIndex = ring.writeIndex

	ring.array[ring.writeIndex] = value
	ring.writeIndex = (ring.writeIndex + 1) % len(ring.array)
}

func (ring *RingBuffer[T]) ReadOldest() T {
	return ring.array[ring.readOldIndex]
}

func (ring *RingBuffer[T]) ReadNewest() T {
	return ring.array[ring.readNewIndex]
}