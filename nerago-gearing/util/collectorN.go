package util

import (
	"cmp"
	"math"
	"slices"
)

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
	equals func(a, b *T) bool
}

func LowestCollector_ForN[T any](limit int, equals func(a, b *T) bool) LowestCollectorN[T] {
	return LowestCollectorN[T]{
		collectorNInternal[T]{
			array: make([]internalEntry[T], 0, limit),
			worst: math.MaxUint64,
			size:  0,
			limit: limit},
		equals}
}

func (collect *LowestCollectorN[T]) sortContent() {
	slices.SortFunc(collect.array, func(a, b internalEntry[T]) int {
		return cmp.Compare(b.value, a.value)
	})
	collect.array = slices.CompactFunc(collect.array, func(a, b internalEntry[T]) bool {
		return a.value == b.value && collect.equals(a.object, b.object)
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

func LowestCollectorN_OfChannel[T any](channel <-chan LowestCollectorN[T], expectNum int) []T {
	var best *LowestCollectorN[T] = nil
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
type HighestCollectorN[T any] struct {
	collectorNInternal[T]
	equals func(a, b *T) bool
}

func HighestCollector_ForN[T any](limit int, equals func(a, b *T) bool) HighestCollectorN[T] {
	return HighestCollectorN[T]{
		collectorNInternal[T]{
			array: make([]internalEntry[T], 0, limit),
			worst: 0,
			size:  0,
			limit: limit},
		equals}
}

func (collect *HighestCollectorN[T]) sortContent() {
	slices.SortFunc(collect.array, func(a, b internalEntry[T]) int {
		return cmp.Compare(a.value, b.value)
	})
	collect.array = slices.CompactFunc(collect.array, func(a, b internalEntry[T]) bool {
		return a.value == b.value && collect.equals(a.object, b.object)
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

func HighestCollectorN_OfChannel[T any](channel <-chan HighestCollectorN[T], expectNum int) []T {
	var best *HighestCollectorN[T] = nil
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
