package util_collection

type QueueRingBufferFifo[T any] struct {
	array      []T
	readIndex  int
	writeIndex int
}

func QueueRingBufferFifo_Create[T any](allocateSize int) IQueue[T] {
	array := make([]T, allocateSize)
	return &QueueRingBufferFifo[T]{array: array, readIndex: 0, writeIndex: 0}
}

func (ring *QueueRingBufferFifo[T]) Push(value T) {
	ring.array[ring.writeIndex] = value
	ring.writeIndex = (ring.writeIndex + 1) % len(ring.array)

	if ring.writeIndex == ring.readIndex {
		newArray := make([]T, len(ring.array)*2)
		newArray[0] = ring.array[ring.readIndex]
		write := 1
		for read := (ring.readIndex + 1) % len(ring.array); read != ring.writeIndex; read = (read + 1) % len(ring.array) {
			newArray[write] = ring.array[read]
			write++
		}
		ring.array = newArray
		ring.readIndex = 0
		ring.writeIndex = write
	}
}

func (ring *QueueRingBufferFifo[T]) Pop() (T, bool) {
	if ring.readIndex == ring.writeIndex {
		var nilValue T
		return nilValue, false
	}

	value := ring.array[ring.readIndex]
	ring.readIndex = (ring.readIndex + 1) % len(ring.array)
	return value, true
}

func (ring *QueueRingBufferFifo[T]) Clear() {
	clear(ring.array)
	ring.readIndex = 0
	ring.writeIndex = 0
}

func (ring *QueueRingBufferFifo[T]) IsEmpty() bool {
	return ring.readIndex == ring.writeIndex
}

func (ring *QueueRingBufferFifo[T]) Size() int {
	if ring.readIndex == ring.writeIndex {
		return 0
	} else if ring.readIndex < ring.writeIndex {
		return ring.writeIndex - ring.readIndex
	} else {
		return ring.writeIndex - ring.readIndex + len(ring.array)
	}
}
