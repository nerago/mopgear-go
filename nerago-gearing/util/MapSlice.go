package util

import (
	"iter"
	"slices"
)

type MapSlice[K comparable, V any] struct {
	data map[K][]V
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

func (mapslice *MapSlice[K, V]) ValuesForKeyAsSlice(key K) ([]V, bool) {
	inner, hasInner := mapslice.data[key]
	if hasInner {
		return inner, hasInner
	} else {
		return nil, false
	}
}

func (mapslice *MapSlice[K, V]) ValuesForKeyAsSeq(key K) iter.Seq[V] {
	inner, hasInner := mapslice.data[key]
	if hasInner {
		return slices.Values(inner)
	} else {
		return func(yield func(V) bool) {}
	}
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

func (mapslice *MapSlice[K, V]) ForeachWithKeys(apply func(key K, value V)) {
	for key, inner := range mapslice.data {
		for _, value := range inner {
			apply(key, value)
		}
	}
}

func (mapslice *MapSlice[K, V]) ForeachValues(apply func(value V)) {
	for _, inner := range mapslice.data {
		for _, value := range inner {
			apply(value)
		}
	}
}
