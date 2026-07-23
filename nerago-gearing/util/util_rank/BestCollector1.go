package util_rank

import (
	"paladin_gearing_go/util/util_collection"
	"sync"
)

// /////////////////////////////////////////////////////////
type BestCollector1[T any] struct {
	BestObject *T
	BestValue  float64
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

func (collect *BestCollector1[T]) GetBestOrPanic() T {
	collect.CheckValidOrPanic()
	return *collect.BestObject
}

func (collect *BestCollector1[T]) GetBestOrDefault(defaultValue T) T {
	if collect.hasBest {
		return *collect.BestObject
	} else {
		return defaultValue
	}
}

func (collect *BestCollector1[T]) GetBestPointerOrPanic() *T {
	collect.CheckValidOrPanic()
	return collect.BestObject
}

func (collect *BestCollector1[T]) GetBestOptional() util_collection.Optional[T] {
	if collect.hasBest {
		return util_collection.Optional_OfPointer(collect.BestObject)
	} else {
		return util_collection.Optional_Empty[T]()
	}
}

func (collect *BestCollector1[T]) GetBestOrNilValue() T {
	if collect.hasBest {
		return *collect.BestObject
	} else {
		var nilValue T
		return nilValue
	}
}

func (collect *BestCollector1[T]) Offer(object *T, value float64) {
	if collect.isBetter(value) || !collect.hasBest {
		collect.BestObject = object
		collect.BestValue = value
		collect.hasBest = true
	}
}

func (collect *BestCollector1[T]) OfferAndIsBetter(object *T, value float64) bool {
	if collect.isBetter(value) || !collect.hasBest {
		collect.BestObject = object
		collect.BestValue = value
		collect.hasBest = true
		return true
	} else {
		return false
	}
}

func (collect *BestCollector1[T]) OfferAndSwap(pointer **T, value float64) {
	if collect.isBetter(value) || !collect.hasBest {
		var prev *T = collect.BestObject
		collect.BestObject = *pointer
		collect.BestValue = value
		collect.hasBest = true
		*pointer = prev
	}
}

func (collect *BestCollector1[T]) isBetter(value float64) bool {
	if collect.Minimise {
		return value < collect.BestValue
	} else {
		return value > collect.BestValue
	}
}

type BestCollector1Concurrent[T any] struct {
	inner BestCollector1[T]
	mutex sync.RWMutex
}

func (bc *BestCollector1Concurrent[T]) GetBestValue() float64 {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	return bc.inner.BestValue
}

func (bc *BestCollector1Concurrent[T]) Offer(object *T, value float64) {
	bc.mutex.Lock()
	bc.inner.Offer(object, value)
	bc.mutex.Unlock()
}

func (bc *BestCollector1Concurrent[T]) GetBestOrNilValue() T {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	return bc.inner.GetBestOrNilValue()
}
