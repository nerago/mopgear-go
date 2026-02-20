package util

// ///////////////////////////////////////////////////////////
type BestCollector1[T any] struct {
	BestObject *T
	BestValue  uint64
}

func (collect *BestCollector1[T]) CheckValidOrPanic() {
	if collect.BestObject == nil {
		panic("no best found")
	}
}

func (collect *BestCollector1[T]) GetBest() T {
	if collect.BestObject == nil {
		panic("no best found")
	}

	return *collect.BestObject
}

func (collect *BestCollector1[T]) Offer(object *T, value uint64) {
	if value > collect.BestValue {
		collect.BestObject = object
		collect.BestValue = value
	}
}

func (collect *BestCollector1[T]) CombineOther(other BestCollector1[T]) {
	collect.Offer(other.BestObject, other.BestValue)
}

func BestCollector1_OfChannel[T any](channel <-chan BestCollector1[T], expectNum int) T {
	best := BestCollector1[T]{}
	for range expectNum {
		threadResult := <-channel
		best.CombineOther(threadResult)
	}
	return best.GetBest()
}
