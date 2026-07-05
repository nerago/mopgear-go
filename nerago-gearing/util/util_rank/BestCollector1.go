package util_rank

import "paladin_gearing_go/util"

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

func (collect *BestCollector1[T]) GetBestOptional() util.Optional[T] {
	if collect.hasBest {
		return util.Optional_OfPointer(collect.BestObject)
	} else {
		return util.Optional_Empty[T]()
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
