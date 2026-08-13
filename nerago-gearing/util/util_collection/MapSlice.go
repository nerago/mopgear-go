package util_collection

import (
	"iter"
	"slices"
)

type MapSlice[K comparable, V any] struct {
	data map[K][]V
}

func (ms *MapSlice[K, V]) Init(size int) {
	ms.data = make(map[K][]V, size)
}

func (ms *MapSlice[K, V]) Clear() {
	clear(ms.data)
}

func (ms *MapSlice[K, V]) Add(key K, value V) {
	if ms.data == nil {
		ms.data = make(map[K][]V)
	}

	inner, hasInner := ms.data[key]
	if hasInner {
		ms.data[key] = append(inner, value)
	} else {
		ms.data[key] = []V{value}
	}
}

func (ms *MapSlice[K, V]) RemoveAllForKey(key K) {
	delete(ms.data, key)
}

func (ms *MapSlice[K, V]) CountForKey(key K) int {
	inner := ms.data[key]
	return len(inner)
}

func (ms *MapSlice[K, V]) ValuesForKeyAsSeq(key K) iter.Seq[V] {
	inner := ms.data[key]
	return slices.Values(inner)
}

func (ms *MapSlice[K, V]) GetInternalSlice(key K) []V {
	return ms.data[key]
}

func (ms *MapSlice[K, V]) GetInternalSliceOptional(key K) ([]V, bool) {
	slice, found := ms.data[key]
	return slice, found
}

func (ms *MapSlice[K, V]) Has(key K) bool {
	slice := ms.data[key]
	return len(slice) > 0
}

func (ms *MapSlice[K, V]) SeqKeys() iter.Seq[K] {
	return func(yield func(K) bool) {
		for key := range ms.data {
			if !yield(key) {
				return
			}
		}
	}
}

func (ms *MapSlice[K, V]) SeqKeysValues() iter.Seq2[K, *V] {
	return func(yield func(K, *V) bool) {
		for key, inner := range ms.data {
			for i := range inner {
				if !yield(key, &inner[i]) {
					return
				}
			}
		}
	}
}

func (ms *MapSlice[K, V]) SeqValues() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, inner := range ms.data {
			for i := range inner {
				if !yield(inner[i]) {
					return
				}
			}
		}
	}
}

func (ms *MapSlice[K, V]) SeqValuesPointer() iter.Seq[*V] {
	return func(yield func(*V) bool) {
		for _, inner := range ms.data {
			for i := range inner {
				if !yield(&inner[i]) {
					return
				}
			}
		}
	}
}

func (ms *MapSlice[K, V]) SeqGroupsNestedKeyValue() iter.Seq2[K, iter.Seq[V]] {
	return func(yield func(K, iter.Seq[V]) bool) {
		for key, inner := range ms.data {
			if !yield(key, slices.Values(inner)) {
				return
			}
		}
	}
}

func (ms *MapSlice[K, V]) SeqGroupsInternalSlice() iter.Seq2[K, []V] {
	return func(yield func(K, []V) bool) {
		for key, inner := range ms.data {
			if !yield(key, inner) {
				return
			}
		}
	}
}

func (ms *MapSlice[K, V]) MapInternalSlice(key K, mapper func([]V) []V) {
	if ms.data == nil {
		ms.data = make(map[K][]V)
	}
	newSlice := mapper(ms.data[key])
	if len(newSlice) > 0 {
		ms.data[key] = newSlice
	} else {
		delete(ms.data, key)
	}
}

func (ms *MapSlice[K, V]) MapInternalSlicesAll(mapper func(K, []V) []V) {
	for key, inner := range ms.data {
		newSlice := mapper(key, inner)
		if len(newSlice) > 0 {
			ms.data[key] = newSlice
		} else {
			delete(ms.data, key)
		}
	}
}
