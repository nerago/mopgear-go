package util_rank

import (
	"cmp"
	"math"
	"slices"
)

type internalFloatEntry[T any] struct {
	object *T
	value  float64
}

type collectorFloatNInternal[T any] struct {
	array       []internalFloatEntry[T]
	worst       float64
	size, limit uint64
}

type LowestCollectorFloatN[T any] struct {
	collectorFloatNInternal[T]
	equals func(a, b *T) bool
}

func LowestCollectorFloat_ForN[T any](limit uint64, equals func(a, b *T) bool) LowestCollectorFloatN[T] {
	return LowestCollectorFloatN[T]{
		collectorFloatNInternal[T]{
			array: make([]internalFloatEntry[T], 0, limit),
			worst: math.MaxFloat64,
			size:  0,
			limit: limit},
		equals}
}

func (collect *LowestCollectorFloatN[T]) sortContent() {
	slices.SortFunc(collect.array, func(a, b internalFloatEntry[T]) int {
		return cmp.Compare(b.value, a.value)
	})
	collect.array = slices.CompactFunc(collect.array, func(a, b internalFloatEntry[T]) bool {
		return a.value == b.value && collect.equals(a.object, b.object)
	})
}

func (collect *LowestCollectorFloatN[T]) Offer(object *T, value float64) {
	if collect.size < collect.limit {
		entry := internalFloatEntry[T]{object, value}
		collect.array = append(collect.array, entry)
		collect.sortContent()
		collect.worst = collect.array[0].value
		collect.size++
	} else if value < collect.worst {
		entry := internalFloatEntry[T]{object, value}
		collect.array[0] = entry
		collect.sortContent()
		collect.worst = collect.array[0].value
	}
}

func (collect *LowestCollectorFloatN[T]) Merge_Mutating(other *LowestCollectorFloatN[T]) {
	if other.size > 0 {
		collect.array = append(collect.array, other.array...)
		collect.sortContent()

		arrayTotal := uint64(len(collect.array))
		if arrayTotal > collect.limit {
			collect.array = collect.array[arrayTotal-collect.limit : arrayTotal]
			collect.size = collect.limit
		} else {
			collect.size = arrayTotal
		}

		collect.worst = collect.array[0].value
	}
}

func (collect *collectorFloatNInternal[T]) ResultsFlat() []T {
	result := make([]T, 0, collect.size)
	for _, entry := range collect.array {
		result = append(result, *entry.object)
	}
	return result
}

func (collect *collectorFloatNInternal[T]) ResultsPointers() []*T {
	result := make([]*T, 0, collect.size)
	for _, entry := range collect.array {
		result = append(result, entry.object)
	}
	return result
}

func LowestCollectorFloatN_OfChannel[T any](channel <-chan LowestCollectorFloatN[T], expectNum int) []T {
	var best *LowestCollectorFloatN[T] = nil
	for range expectNum {
		threadResult := <-channel
		if best == nil {
			best = &threadResult
		} else {
			best.Merge_Mutating(&threadResult)
		}
	}
	return best.ResultsFlat()
}

// ///////////////////////////////////////////////////////////
type HighestCollectorFloatN[T any] struct {
	collectorFloatNInternal[T]
	equals func(a, b *T) bool
}

func HighestCollectorFloat_ForN[T any](limit uint64, equals func(a, b *T) bool) HighestCollectorFloatN[T] {
	return HighestCollectorFloatN[T]{
		collectorFloatNInternal[T]{
			array: make([]internalFloatEntry[T], 0, limit),
			worst: 0,
			size:  0,
			limit: limit},
		equals}
}

func (collect *HighestCollectorFloatN[T]) sortContent() {
	slices.SortFunc(collect.array, func(a, b internalFloatEntry[T]) int {
		return cmp.Compare(a.value, b.value)
	})
	collect.array = slices.CompactFunc(collect.array, func(a, b internalFloatEntry[T]) bool {
		return a.value == b.value && collect.equals(a.object, b.object)
	})
	// TODO consider util.RemoveDuplicatesFunc()
}

func (collect *HighestCollectorFloatN[T]) Offer(object *T, value float64) {
	if collect.size < collect.limit {
		entry := internalFloatEntry[T]{object, value}
		collect.array = append(collect.array, entry)
		collect.sortContent()
		collect.worst = collect.array[0].value
		collect.size++
	} else if value > collect.worst {
		entry := internalFloatEntry[T]{object, value}
		collect.array[0] = entry
		collect.sortContent()
		collect.worst = collect.array[0].value
	}
}

func (collect *HighestCollectorFloatN[T]) Merge_Mutating(other *HighestCollectorFloatN[T]) {
	if other.size > 0 {
		collect.array = append(collect.array, other.array...)
		collect.sortContent()

		arrayTotal := uint64(len(collect.array))
		if arrayTotal > collect.limit {
			collect.array = collect.array[arrayTotal-collect.limit : arrayTotal]
			collect.size = collect.limit
		} else {
			collect.size = arrayTotal
		}

		collect.worst = collect.array[0].value
	}
}

func HighestCollectorFloatN_OfChannel[T any](channel <-chan HighestCollectorFloatN[T], expectNum int) []T {
	var best *HighestCollectorFloatN[T] = nil
	for range expectNum {
		threadResult := <-channel
		if best == nil {
			best = &threadResult
		} else {
			best.Merge_Mutating(&threadResult)
		}
	}
	return best.ResultsFlat()
}
