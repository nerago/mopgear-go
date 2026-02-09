package util

type BestCollector1[T any] struct {
	BestObject *T
	BestValue  uint64
}

func (collect *BestCollector1[T]) GetBest() T {
	if collect.BestObject == nil {
		panic("no best found")
	}

	return *collect.BestObject
}

func (collect *BestCollector1[T]) Add(object *T, value uint64) {
	if value > collect.BestValue {
		collect.BestObject = object
		collect.BestValue = value
	}
}
