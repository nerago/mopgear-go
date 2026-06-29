package util

type RingBufferCache[T any] struct {
	array                      []T
	readOldIndex, readNewIndex int
	writeIndex                 int
}

func RingBufferCache_Create[T any](size int, initialEntry T) RingBufferCache[T] {
	array := make([]T, size)
	array[0] = initialEntry
	return RingBufferCache[T]{array: array, readOldIndex: 0, readNewIndex: 0, writeIndex: 1}
}

func (ring *RingBufferCache[T]) Write(value T) {
	if ring.readOldIndex == ring.writeIndex {
		ring.readOldIndex = (ring.readOldIndex + 1) % len(ring.array)
	}

	ring.readNewIndex = ring.writeIndex

	ring.array[ring.writeIndex] = value
	ring.writeIndex = (ring.writeIndex + 1) % len(ring.array)
}

func (ring *RingBufferCache[T]) ReadOldest() T {
	return ring.array[ring.readOldIndex]
}

func (ring *RingBufferCache[T]) ReadNewest() T {
	return ring.array[ring.readNewIndex]
}
