package util

import (
	"iter"
	"slices"
)

type MapSlice[K comparable, V any] struct {
	data map[K][]V
}

func (mapslice MapSlice[K, V]) Init(size int) {
	mapslice.data = make(map[K][]V, size)
}

func (mapslice *MapSlice[K, V]) Clear() {
	clear(mapslice.data)
}

func (mapslice *MapSlice[K, V]) Add(key K, value V) {
	if mapslice.data == nil {
		mapslice.data = make(map[K][]V)
	}

	inner, hasInner := mapslice.data[key]
	if hasInner {
		mapslice.data[key] = append(inner, value)
	} else {
		mapslice.data[key] = []V{value}
	}
}

func (mapslice *MapSlice[K, V]) ValuesForKeyAsSeq(key K) iter.Seq[V] {
	inner := mapslice.data[key]
	return slices.Values(inner)
}

func (mapslice *MapSlice[K, V]) GetInternalSlice(key K) []V {
	return mapslice.data[key]
}

func (mapslice *MapSlice[K, V]) SeqKeys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for key := range mapslice.data {
			if !yield(key) {
				return
			}
		}
	}
}

func (mapslice *MapSlice[K, V]) SeqKeysValues() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for key, inner := range mapslice.data {
			for _, value := range inner {
				if !yield(key, value) {
					return
				}
			}
		}
	}
}

func (mapslice *MapSlice[K, V]) SeqValues() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, inner := range mapslice.data {
			for _, value := range inner {
				if !yield(value) {
					return
				}
			}
		}
	}
}

func (mapslice *MapSlice[K, V]) SeqGroupsNestedKeyValue() iter.Seq2[K, iter.Seq[V]] {
	return func(yield func(K, iter.Seq[V]) bool) {
		for key, inner := range mapslice.data {
			if !yield(key, slices.Values(inner)) {
				return
			}
		}
	}
}

func (mapslice *MapSlice[K, V]) SeqGroupsInternalSlice() iter.Seq2[K, []V] {
	return func(yield func(K, []V) bool) {
		for key, inner := range mapslice.data {
			if !yield(key, inner) {
				return
			}
		}
	}
}

func (mapslice *MapSlice[K, V]) MapInternalSlice(key K, mapper func([]V) []V) {
	if mapslice.data != nil {
		mapslice.data[key] = mapper(mapslice.data[key])
	}
}

func (mapslice *MapSlice[K, V]) MapInternalSlicesAll(mapper func(K, []V) []V) {
	for key, inner := range mapslice.data {
		mapslice.data[key] = mapper(key, inner)
	}
}
