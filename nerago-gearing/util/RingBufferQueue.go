package util

// FIFO queue
type RingBufferQueue[T any] struct {
	array      []T
	readIndex  int
	writeIndex int
}

func RingBufferQueue_Create[T any](allocateSize int) RingBufferQueue[T] {
	array := make([]T, allocateSize)
	return RingBufferQueue[T]{array: array, readIndex: 0, writeIndex: 0}
}

func (ring *RingBufferQueue[T]) Push(value T) {
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

func (ring *RingBufferQueue[T]) Pop() (T, bool) {
	if ring.readIndex == ring.writeIndex {
		var nilValue T
		return nilValue, false
	}

	value := ring.array[ring.readIndex]
	ring.readIndex = (ring.readIndex + 1) % len(ring.array)
	return value, true
}
