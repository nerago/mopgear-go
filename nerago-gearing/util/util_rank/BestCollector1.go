package util_rank

import "paladin_gearing_go/util"

// /////////////////////////////////////////////////////////
type BestCollector1[T any] struct {
	BestObject *T
	BestValue  float64
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

func (collect *BestCollector1[T]) GetBestOptional() util.Optional[T] {
	if collect.hasBest {
		return util.Optional_OfPointer(collect.BestObject)
	} else {
		return util.Optional_Empty[T]()
	}
}

func (collect *BestCollector1[T]) Offer(object *T, value float64) {
	if value > collect.BestValue || !collect.hasBest {
		collect.BestObject = object
		collect.BestValue = value
		collect.hasBest = true
	}
}

func (collect *BestCollector1[T]) OfferAndSwap(pointer **T, value float64) {
	if value > collect.BestValue || !collect.hasBest {
		var prev *T = collect.BestObject
		collect.BestObject = *pointer
		collect.BestValue = value
		collect.hasBest = true
		*pointer = prev
	}
}
