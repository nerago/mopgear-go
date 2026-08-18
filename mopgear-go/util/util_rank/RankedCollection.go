package util_rank

import (
	"iter"
	"maps"
	"slices"
)

type RankedCollection[T comparable] struct {
	groups map[float64][]T
}

func (coll *RankedCollection[T]) Add(object T, rating float64) {
	if coll.groups == nil {
		coll.groups = make(map[float64][]T)
	}

	coll.groups[rating] = append(coll.groups[rating], object)
}

func (coll *RankedCollection[T]) OrderedResult() iter.Seq2[T, float64] {
	if coll.groups == nil {
		return func(yield func(T, float64) bool) {}
	}

	keys := slices.AppendSeq([]float64{}, maps.Keys(coll.groups))
	slices.Sort(keys)
	return func(yield func(T, float64) bool) {
		for _, rating := range keys {
			for _, item := range coll.groups[rating] {
				if !yield(item, rating) {
					return
				}
			}
		}
	}
}
