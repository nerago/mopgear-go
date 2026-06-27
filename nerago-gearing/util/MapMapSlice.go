package util

import (
	"iter"
	"slices"
)

type MapMapSlice[J comparable, K comparable, V any] struct {
	dataBy1 map[J]map[K][]V
	dataBy2 map[K]map[J][]V
}

type MapMapSliceEntry[J comparable, K comparable, V any] struct {
	Key1     J
	Key2     K
	ValueSeq iter.Seq[V]
}

func (mmapslice *MapMapSlice[J, K, V]) ValuesForKeyAsSlice(key1 J, key2 K) ([]V, bool) {
	value, hasValue := mmapslice.dataBy1[key1][key2]
	return value, hasValue
}

func (mmapslice *MapMapSlice[J, K, V]) ValuesForKeyAsSeq(key1 J, key2 K) iter.Seq[V] {
	value := mmapslice.dataBy1[key1][key2]
	return slices.Values(value)
}

func (mmapslice *MapMapSlice[J, K, V]) ValuesForKey1AsSeq(key1 J) iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, slice := range mmapslice.dataBy1[key1] {
			for _, value := range slice {
				if !yield(value) {
					return
				}
			}
		}
	}
}

func (mmapslice *MapMapSlice[J, K, V]) ValuesForKey2AsSeq(key2 K) iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, slice := range mmapslice.dataBy2[key2] {
			for _, value := range slice {
				if !yield(value) {
					return
				}
			}
		}
	}
}

func (mmapslice *MapMapSlice[J, K, V]) Clear() {
	clear(mmapslice.dataBy1)
	clear(mmapslice.dataBy2)
}

func (mmapslice *MapMapSlice[J, K, V]) Add(key1 J, key2 K, value V) {
	var slice []V

	data1 := mmapslice.dataBy1
	if data1 == nil {
		data1 = make(map[J]map[K][]V)
		mmapslice.dataBy1 = data1

		inner1 := make(map[K][]V)
		data1[key1] = inner1

		slice = []V{value}
		inner1[key2] = slice
	} else if inner1, hasInner1 := data1[key1]; !hasInner1 {
		inner1 = make(map[K][]V)
		data1[key1] = inner1

		slice = []V{value}
		inner1[key2] = slice
	} else {
		slice = inner1[key2]
		slice = append(slice, value)
		inner1[key2] = slice
	}

	data2 := mmapslice.dataBy2
	if data2 == nil {
		data2 = make(map[K]map[J][]V)
		mmapslice.dataBy2 = data2

		inner2 := make(map[J][]V)
		data2[key2] = inner2

		inner2[key1] = slice
	} else if inner2, hasInner2 := data2[key2]; !hasInner2 {
		inner2 = make(map[J][]V)
		data2[key2] = inner2

		inner2[key1] = slice
	} else {
		inner2[key1] = slice
	}
}

func (mmapslice *MapMapSlice[J, K, V]) MapInternalSlice(key1 J, key2 K, mapper func([]V) []V) {
	data1 := mmapslice.dataBy1
	if data1 != nil {
		inner1, hasInner1 := data1[key1]
		if hasInner1 {
			value1, hasValue1 := inner1[key2]
			if hasValue1 {
				newSlice := mapper(value1)
				inner1[key2] = newSlice

				data2 := mmapslice.dataBy2
				if data2 == nil {
					panic("keys not found")
				} else if inner2, hasInner2 := data2[key2]; !hasInner2 {
					panic("keys not found")
				} else {
					inner2[key1] = newSlice
				}

				return
			}
		}
	}
	panic("keys not found")
}

func (mmapslice *MapMapSlice[J, K, V]) SeqWithKeys() iter.Seq[MapMapEntry[J, K, V]] {
	return func(yield func(MapMapEntry[J, K, V]) bool) {
		for key1, inner := range mmapslice.dataBy1 {
			for key2, slice := range inner {
				for _, value := range slice {
					if !yield(MapMapEntry[J, K, V]{key1, key2, value}) {
						return
					}
				}
			}
		}
	}
}

func (mmapslice *MapMapSlice[J, K, V]) SeqValues() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, inner := range mmapslice.dataBy1 {
			for _, slice := range inner {
				for _, value := range slice {
					if !yield(value) {
						return
					}
				}
			}
		}
	}
}

func (mmapslice *MapMapSlice[J, K, V]) ForeachWithKeys(apply func(key1 J, key2 K, value V)) {
	for key1, inner := range mmapslice.dataBy1 {
		for key2, slice := range inner {
			for _, value := range slice {
				apply(key1, key2, value)
			}
		}
	}
}

func (mmapslice *MapMapSlice[J, K, V]) SeqGroupsKey1Lookup() iter.Seq2[J, func(K) iter.Seq[V]] {
	return func(yield func(J, func(K) iter.Seq[V]) bool) {
		for key1, inner := range mmapslice.dataBy1 {
			lookup := func(key2 K) iter.Seq[V] {
				slice := inner[key2]
				return slices.Values(slice)
			}

			if !yield(key1, lookup) {
				return
			}
		}
	}
}

func (mmapslice *MapMapSlice[J, K, V]) SeqGroupsKey2Lookup() iter.Seq2[K, func(J) iter.Seq[V]] {
	return func(yield func(K, func(J) iter.Seq[V]) bool) {
		for key2, inner := range mmapslice.dataBy2 {
			lookup := func(key1 J) iter.Seq[V] {
				slice := inner[key1]
				return slices.Values(slice)
			}

			if !yield(key2, lookup) {
				return
			}
		}
	}
}

func (mmapslice *MapMapSlice[J, K, V]) SeqGroupsKeysNestedValueSeq() iter.Seq[MapMapSliceEntry[J, K, V]] {
	return func(yield func(MapMapSliceEntry[J, K, V]) bool) {
		for key1, inner := range mmapslice.dataBy1 {
			for key2, slice := range inner {
				if !yield(MapMapSliceEntry[J, K, V]{key1, key2, slices.Values(slice)}) {
					return
				}
			}
		}
	}
}
