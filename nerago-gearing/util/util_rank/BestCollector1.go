package util_rank

import "paladin_gearing_go/util"

///////////////////////////////////////////////////////////
type BestCollector1[T any] struct {
	BestObject *T
	BestValue  uint64
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

func (collect *BestCollector1[T]) GetBestPointer() *T {
	if collect.hasBest {
		return collect.BestObject
	} else {
		return nil
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

func (collect *BestCollector1[T]) Offer(object *T, value uint64) {
	if value > collect.BestValue || !collect.hasBest {
		collect.BestObject = object
		collect.BestValue = value
		collect.hasBest = true
	}
}

func (collect *BestCollector1[T]) OfferWithResult(object *T, value uint64) bool {
	if value > collect.BestValue || !collect.hasBest {
		collect.BestObject = object
		collect.BestValue = value
		collect.hasBest = true
		return true
	}
	return false
}

func (collect *BestCollector1[T]) OfferAndSwap(pointer **T, value uint64) {
	if value > collect.BestValue || !collect.hasBest {
		var prev *T = collect.BestObject
		collect.BestObject = *pointer
		collect.BestValue = value
		collect.hasBest = true
		*pointer = prev
	}
}

func (collect *BestCollector1[T]) CombineOther(other BestCollector1[T]) {
	if other.hasBest && other.BestValue >= collect.BestValue {
		collect.BestObject = other.BestObject
		collect.BestValue = other.BestValue
		collect.hasBest = true
	}
}

func BestCollector1_OfChannel[T any](channel <-chan BestCollector1[T], expectNum int) util.Optional[T] {
	best := BestCollector1[T]{}
	for range expectNum {
		threadResult := <-channel
		best.CombineOther(threadResult)
	}
	return best.GetBestOptional()
}
