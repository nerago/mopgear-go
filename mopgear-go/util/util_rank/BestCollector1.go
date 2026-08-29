package util_rank

import (
	"sync"

	"github.com/nerago/mopgear-go/util/util_collection"
)

// ######################## BestCollector1 ########################
type BestCollector1[T any] struct {
	bestObject *T
	bestScore  float64
	Minimise   bool
	hasBest    bool
}

func (collect *BestCollector1[T]) CheckValidOrPanic() {
	if !collect.hasBest {
		panic("no best found")
	}
}

func (collect *BestCollector1[T]) HasBest() bool {
	return collect.hasBest
}

func (collect *BestCollector1[T]) GetBest() (T, bool) {
	return *collect.bestObject, collect.hasBest
}

func (collect *BestCollector1[T]) GetBestPointer() (*T, bool) {
	return collect.bestObject, collect.hasBest
}

func (collect *BestCollector1[T]) GetBestScore() float64 {
	return collect.bestScore
}

func (collect *BestCollector1[T]) GetBestOrPanic() T {
	collect.CheckValidOrPanic()
	return *collect.bestObject
}

func (collect *BestCollector1[T]) GetBestOrDefault(defaultValue T) T {
	if collect.hasBest {
		return *collect.bestObject
	} else {
		return defaultValue
	}
}

func (collect *BestCollector1[T]) GetBestPointerOrPanic() *T {
	collect.CheckValidOrPanic()
	return collect.bestObject
}

func (collect *BestCollector1[T]) GetBestOptional() util_collection.Optional[T] {
	if collect.hasBest {
		return util_collection.Optional_OfPointer(collect.bestObject)
	} else {
		return util_collection.Optional_Empty[T]()
	}
}

func (collect *BestCollector1[T]) GetBestOrNilValue() T {
	if collect.hasBest {
		return *collect.bestObject
	} else {
		var nilValue T
		return nilValue
	}
}

func (collect *BestCollector1[T]) Offer(object *T, value float64) {
	if collect.isBetter(value) || !collect.hasBest {
		collect.bestObject = object
		collect.bestScore = value
		collect.hasBest = true
	}
}

func (collect *BestCollector1[T]) OfferAndIsBetter(object *T, value float64) bool {
	if collect.isBetter(value) || !collect.hasBest {
		collect.bestObject = object
		collect.bestScore = value
		collect.hasBest = true
		return true
	} else {
		return false
	}
}

func (collect *BestCollector1[T]) OfferAndSwap(pointer **T, value float64) {
	if collect.isBetter(value) || !collect.hasBest {
		prev := collect.bestObject
		collect.bestObject = *pointer
		collect.bestScore = value
		collect.hasBest = true
		*pointer = prev
	}
}

func (collect *BestCollector1[T]) isBetter(value float64) bool {
	if collect.Minimise {
		return value < collect.bestScore
	} else {
		return value > collect.bestScore
	}
}

// ######################## BestCollector1Lite ########################
type BestCollector1Lite[T any] struct {
	bestValue T
	bestScore float64
	hasBest   bool
}

func (collect *BestCollector1Lite[T]) CheckValidOrPanic() {
	if !collect.hasBest {
		panic("no best found")
	}
}

func (collect *BestCollector1Lite[T]) HasBest() bool {
	return collect.hasBest
}

func (collect *BestCollector1Lite[T]) GetBestOrPanic() T {
	collect.CheckValidOrPanic()
	return collect.bestValue
}

func (collect *BestCollector1Lite[T]) GetBestOrDefault(defaultValue T) T {
	if collect.hasBest {
		return collect.bestValue
	} else {
		return defaultValue
	}
}

func (collect *BestCollector1Lite[T]) GetBestOrNilValue() T {
	if collect.hasBest {
		return collect.bestValue
	} else {
		var nilValue T
		return nilValue
	}
}

func (collect *BestCollector1Lite[T]) Offer(value T, score float64) {
	if score > collect.bestScore || !collect.hasBest {
		collect.bestValue = value
		collect.bestScore = score
		collect.hasBest = true
	}
}

// ######################## BestCollector1Concurrent ########################
type BestCollector1Concurrent[T any] struct {
	inner BestCollector1[T]
	mutex sync.RWMutex
}

func (bc *BestCollector1Concurrent[T]) GetBestValue() float64 {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	return bc.inner.bestScore
}

func (bc *BestCollector1Concurrent[T]) Offer(object *T, value float64) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	bc.inner.Offer(object, value)
}

func (bc *BestCollector1Concurrent[T]) OfferAndIsBetter(object *T, value float64) bool {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	return bc.inner.OfferAndIsBetter(object, value)
}

func (bc *BestCollector1Concurrent[T]) GetBestOrNilValue() T {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	return bc.inner.GetBestOrNilValue()
}

func (bc *BestCollector1Concurrent[T]) GetBest() (T, bool) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	return bc.inner.GetBest()
}

func (bc *BestCollector1Concurrent[T]) GetBestPointer() (*T, bool) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	return bc.inner.GetBestPointer()
}

func (bc *BestCollector1Concurrent[T]) GetBestOptional() util_collection.Optional[T] {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	return bc.inner.GetBestOptional()
}
