package util_collection

import (
	"iter"
	"math/rand"
	"slices"
)

type ListFlatWithPointers[T any] struct {
	data  []T
	index []uint16
}

var _ IListRead[*int] = &ListFlatWithPointers[int]{}

func ListFlatWithPointersMake[T any](fixedSize uint16) ListFlatWithPointers[T] {
	sl := ListFlatWithPointers[T]{
		make([]T, fixedSize),
		make([]uint16, fixedSize),
	}
	for i := range fixedSize {
		sl.index[i] = i
	}
	return sl
}

func (sl *ListFlatWithPointers[T]) Clear() {
	clear(sl.data)
	for i := range sl.index {
		sl.index[i] = uint16(i)
	}
}

func (sl *ListFlatWithPointers[T]) Size() int {
	return len(sl.index)
}

func (sl *ListFlatWithPointers[T]) IsEmpty() bool {
	return len(sl.index) == 0
}

func (sl *ListFlatWithPointers[T]) Get(idx int) *T {
	return &sl.data[sl.index[idx]]
}

func (sl *ListFlatWithPointers[T]) GetFirst() (*T, bool) {
	if len(sl.index) > 0 {
		return &sl.data[sl.index[0]], true
	} else {
		return nil, false
	}
}

func (sl *ListFlatWithPointers[T]) GetLast() (*T, bool) {
	if len(sl.index) > 0 {
		return &sl.data[sl.index[len(sl.index)-1]], true
	} else {
		return nil, false
	}
}

func (sl *ListFlatWithPointers[T]) ContainsFuncNoPointer(predicate func(*T) bool) bool {
	for i := range sl.data {
		if predicate(&sl.data[i]) {
			return true
		}
	}
	return false
}

func (sl *ListFlatWithPointers[T]) ContainsFunc(predicate func(**T) bool) bool {
	for i := range sl.data {
		element := &sl.data[i]
		if predicate(&element) {
			return true
		}
	}
	return false
}

func (sl *ListFlatWithPointers[T]) SortFunc(compare func(*T, *T) int) {
	slices.SortFunc(sl.index, func(a, b uint16) int {
		return compare(&sl.data[a], &sl.data[b])
	})
}

func (sl *ListFlatWithPointers[T]) Shuffle() {
	rand.Shuffle(len(sl.index), func(a, b int) {
		sl.index[a], sl.index[b] = sl.index[b], sl.index[a]
	})
}

func (sl *ListFlatWithPointers[T]) SeqIndexAndValues() iter.Seq2[int, *T] {
	return func(yield func(int, *T) bool) {
		for _, idx := range sl.index {
			if !yield(int(idx), &sl.data[idx]) {
				return
			}
		}
	}
}

func (sl *ListFlatWithPointers[T]) SeqValues() iter.Seq[*T] {
	return func(yield func(*T) bool) {
		for _, idx := range sl.index {
			if !yield(&sl.data[idx]) {
				return
			}
		}
	}
}

func (sl *ListFlatWithPointers[T]) SeqValuesUnordered() iter.Seq[*T] {
	return func(yield func(*T) bool) {
		for i := range sl.data {
			if !yield(&sl.data[i]) {
				return
			}
		}
	}
}
