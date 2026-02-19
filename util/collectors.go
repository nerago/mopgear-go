package util

import (
	"cmp"
	"math"
	"slices"
)

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

// ///////////////////////////////////////////////////////////
type internalEntry[T any] struct {
	object *T
	value  uint64
}

type collectorNInternal[T any] struct {
	array       []internalEntry[T]
	worst       uint64
	size, limit int
}

type CollectorN[T any] interface {
	Offer(object *T, value uint64)
	ResultsFlat() []T
	ResultsPointers() []*T
	// Merge_Mutating(other *CollectorN[T])
}

type LowestCollectorN[T any] struct {
	collectorNInternal[T]
}

func LowestCollector_ForN[T any](limit int) LowestCollectorN[T] {
	return LowestCollectorN[T]{
		collectorNInternal[T]{
			array: make([]internalEntry[T], 0, limit),
			worst: math.MaxUint64,
			size:  0,
			limit: limit}}
}

func (collect *LowestCollectorN[T]) sortContent() {
	slices.SortFunc(collect.array, func(a, b internalEntry[T]) int {
		return cmp.Compare(b.value, a.value)
	})
}

func (collect *LowestCollectorN[T]) Offer(object *T, value uint64) {
	if collect.size < collect.limit {
		entry := internalEntry[T]{object, value}
		collect.array = append(collect.array, entry)
		collect.sortContent()
		collect.worst = collect.array[0].value
		collect.size++
	} else if value < collect.worst {
		entry := internalEntry[T]{object, value}
		collect.array[0] = entry
		collect.sortContent()
		collect.worst = collect.array[0].value
	}
}

func (collect *LowestCollectorN[T]) Merge_Mutating(other *LowestCollectorN[T]) {
	if other.size > 0 {
		collect.array = append(collect.array, other.array...)
		collect.sortContent()

		arrayTotal := len(collect.array)
		if arrayTotal > collect.limit {
			collect.array = collect.array[arrayTotal-collect.limit : arrayTotal]
			collect.size = collect.limit
		} else {
			collect.size = arrayTotal
		}

		collect.worst = collect.array[0].value
	}
}

func (collect *collectorNInternal[T]) ResultsFlat() []T {
	result := make([]T, 0, collect.size)
	for _, entry := range collect.array {
		result = append(result, *entry.object)
	}
	return result
}

func (collect *collectorNInternal[T]) ResultsPointers() []*T {
	result := make([]*T, 0, collect.size)
	for _, entry := range collect.array {
		result = append(result, entry.object)
	}
	return result
}

// ///////////////////////////////////////////////////////////
type HighestCollectorN[T any] struct {
	collectorNInternal[T]
}

func HighestCollector_ForN[T any](limit int) HighestCollectorN[T] {
	return HighestCollectorN[T]{
		collectorNInternal[T]{
			array: make([]internalEntry[T], 0, limit),
			worst: 0,
			size:  0,
			limit: limit}}
}

func (collect *HighestCollectorN[T]) sortContent() {
	slices.SortFunc(collect.array, func(a, b internalEntry[T]) int {
		return cmp.Compare(a.value, b.value)
	})
}

func (collect *HighestCollectorN[T]) Offer(object *T, value uint64) {
	if collect.size < collect.limit {
		entry := internalEntry[T]{object, value}
		collect.array = append(collect.array, entry)
		collect.sortContent()
		collect.worst = collect.array[0].value
		collect.size++
	} else if value > collect.worst {
		entry := internalEntry[T]{object, value}
		collect.array[0] = entry
		collect.sortContent()
		collect.worst = collect.array[0].value
	}
}

func (collect *HighestCollectorN[T]) Merge_Mutating(other *HighestCollectorN[T]) {
	if other.size > 0 {
		collect.array = append(collect.array, other.array...)
		collect.sortContent()

		arrayTotal := len(collect.array)
		if arrayTotal > collect.limit {
			collect.array = collect.array[arrayTotal-collect.limit : arrayTotal]
			collect.size = collect.limit
		} else {
			collect.size = arrayTotal
		}

		collect.worst = collect.array[0].value
	}
}
