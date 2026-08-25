package util_collection

import (
	"iter"
	"sync"
)

type QueueStackFiloConcurrent[T any] struct {
	inner QueueStackFilo[T]
	mutex sync.RWMutex
}

var _ IQueueLite[int] = &QueueStackFiloConcurrent[int]{}

func (stack *QueueStackFiloConcurrent[T]) IsEmpty() bool {
	stack.mutex.RLock()
	defer stack.mutex.RUnlock()
	return stack.inner.IsEmpty()
}

func (stack *QueueStackFiloConcurrent[T]) Size() int {
	stack.mutex.RLock()
	defer stack.mutex.RUnlock()
	return stack.inner.Size()
}

func (stack *QueueStackFiloConcurrent[T]) Push(value T) {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	stack.inner.Push(value)
}

func (stack *QueueStackFiloConcurrent[T]) PushSeveral(valueSlice []T) {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	stack.inner.PushSeveral(valueSlice)
}

func (stack *QueueStackFiloConcurrent[T]) Pop() (T, bool) {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	return stack.inner.Pop()
}

func (stack *QueueStackFiloConcurrent[T]) PopSeveral(buffer []T) []T {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	return stack.inner.PopSeveral(buffer)
}

func (stack *QueueStackFiloConcurrent[T]) Clear() {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	stack.inner.Clear()
}

func (stack *QueueStackFiloConcurrent[T]) SeqValues() iter.Seq[T] {
	return func(yield func(T) bool) {
		stack.mutex.RLock()
		defer stack.mutex.RUnlock()
		for value := range stack.inner.SeqValues() {
			if !yield(value) {
				return
			}
		}
	}
}

func (stack *QueueStackFiloConcurrent[T]) ExamineContents(apply func([]T)) {
	stack.mutex.RLock()
	defer stack.mutex.RUnlock()

	stack.inner.ExamineContents(apply)
}

func (stack *QueueStackFiloConcurrent[T]) UpdateContents(apply func([]T) []T) {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()

	stack.inner.UpdateContents(apply)
}

const c_cacheSize = 4

type QueueStackFiloConcurrentCached[T any] struct {
	primary *QueueStackFiloConcurrent[T]
	read    QueueStackFilo[T]
	write   QueueStackFilo[T]
}

func QueueStackFiloConcurrentCachedMake[T any](primary *QueueStackFiloConcurrent[T]) *QueueStackFiloConcurrentCached[T] {
	return &QueueStackFiloConcurrentCached[T]{primary: primary}
}

func (cache *QueueStackFiloConcurrentCached[T]) Flush() {
	for !cache.write.IsEmpty() {
		cache.flush()
	}
}

func (cache *QueueStackFiloConcurrentCached[T]) flush() {
	array := [c_cacheSize]T{}
	slice := cache.write.PopSeveral(array[:])
	cache.primary.PushSeveral(slice)
}

func (cache *QueueStackFiloConcurrentCached[T]) Push(value T) {
	cache.write.Push(value)
	if cache.write.Size() >= c_cacheSize {
		cache.flush()
	}
}

func (cache *QueueStackFiloConcurrentCached[T]) Pop() (T, bool) {
	if value, hasValue := cache.read.Pop(); hasValue {
		return value, hasValue
	} else if value2, hasValue2 := cache.write.Pop(); hasValue2 {
		return value2, hasValue2
	} else {
		return cache.popFromPrimary()
	}
}

func (cache *QueueStackFiloConcurrentCached[T]) popFromPrimary() (T, bool) {
	array := [c_cacheSize]T{}
	slice := cache.primary.PopSeveral(array[:])
	if len(slice) == 0 {
		return makeNilValue[T](), false
	} else if len(slice) == 1 {
		return slice[0], true
	} else {
		cache.read.PushSeveral(slice[1:])
		return slice[0], true
	}
}

func (cache *QueueStackFiloConcurrentCached[T]) SizeVolatile() any {
	return cache.primary.inner.Size()
}

func (cache *QueueStackFiloConcurrentCached[T]) ExamineContents(apply func([]T)) {
	// really should concat the cached items too
	cache.primary.ExamineContents(apply)
}
