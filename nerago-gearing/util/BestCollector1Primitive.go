package util

///////////////////////////////////////////////////////////
type BestCollector1Primitive[T any] struct {
	BestObject T
	BestValue  uint64
	hasValue   bool
}

func (collect *BestCollector1Primitive[T]) CheckValidOrPanic() {
	if !collect.hasValue {
		panic("no best found")
	}
}

func (collect *BestCollector1Primitive[T]) HasBest() bool {
	return collect.hasValue
}

func (collect *BestCollector1Primitive[T]) GetBestOrPanic() T {
	collect.CheckValidOrPanic()
	return collect.BestObject
}

func (collect *BestCollector1Primitive[T]) GetBestOptional() Optional[T] {
	if collect.hasValue {
		return Optional_OfValue(collect.BestObject)
	} else {
		return Optional_Empty[T]()
	}
}

func (collect *BestCollector1Primitive[T]) Offer(object T, value uint64) {
	if value > collect.BestValue {
		collect.BestObject = object
		collect.BestValue = value
		collect.hasValue = true
	}
}

func (collect *BestCollector1Primitive[T]) CombineOther(other BestCollector1Primitive[T]) {
	if other.hasValue && other.BestValue > collect.BestValue {
		collect.BestObject = other.BestObject
		collect.BestValue = other.BestValue
		collect.hasValue = true
	}
}
