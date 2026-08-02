package util_collection

import (
	"iter"
	"maps"
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

func (mms *MapMapSlice[J, K, V]) Size() int {
	size := 0
	for _, inner := range mms.dataBy1 {
		for _, slice := range inner {
			size += len(slice)
		}
	}
	return size
}

func (mms *MapMapSlice[J, K, V]) IsEmpty() bool {
	return len(mms.dataBy1) == 0
}

func (mms *MapMapSlice[J, K, V]) Has(key1 J, key2 K) bool {
	_, hasValue := mms.dataBy1[key1][key2]
	return hasValue
}

func (mms *MapMapSlice[J, K, V]) GetAsSliceInternal(key1 J, key2 K) ([]V, bool) {
	value, hasValue := mms.dataBy1[key1][key2]
	return value, hasValue
}

func (mms *MapMapSlice[J, K, V]) GetAsSliceClone(key1 J, key2 K) ([]V, bool) {
	value, hasValue := mms.dataBy1[key1][key2]
	return slices.Clone(value), hasValue
}

func (mms *MapMapSlice[J, K, V]) GetAsSeq(key1 J, key2 K) iter.Seq[V] {
	value := mms.dataBy1[key1][key2]
	return slices.Values(value)
}

func (mms *MapMapSlice[J, K, V]) HasKey1(key1 J) bool {
	_, hasInner := mms.dataBy1[key1]
	return hasInner
}

func (mms *MapMapSlice[J, K, V]) HasKey2(key2 K) bool {
	_, hasInner := mms.dataBy2[key2]
	return hasInner
}

func (mms *MapMapSlice[J, K, V]) FirstKey1() J {
	for k := range mms.dataBy1 {
		return k
	}
	panic("empty map")
}

func (mms *MapMapSlice[J, K, V]) FirstKey2() K {
	for k := range mms.dataBy2 {
		return k
	}
	panic("empty map")
}

func (mms *MapMapSlice[J, K, V]) DeleteAllForKey1(key1 J) {
	deleteAllNestedIn(key1, mms.dataBy1, mms.dataBy2)
	delete(mms.dataBy1, key1)
}

func (mms *MapMapSlice[J, K, V]) DeleteAllForKey2(key2 K) {
	deleteAllNestedIn(key2, mms.dataBy2, mms.dataBy1)
	delete(mms.dataBy2, key2)
}

func (mms *MapMapSlice[J, K, V]) DeleteAllForKey1Key2(key1 J, key2 K) {
	deleteInMap(key1, key2, mms.dataBy1)
	deleteInMap(key2, key1, mms.dataBy2)
}

func (mms *MapMapSlice[J, K, V]) SeqKey1() iter.Seq[J] {
	return maps.Keys(mms.dataBy1)
}

func (mms *MapMapSlice[J, K, V]) SeqKey2() iter.Seq[K] {
	return maps.Keys(mms.dataBy2)
}

func (mms *MapMapSlice[J, K, V]) SeqKey1Key2() iter.Seq2[J, iter.Seq[K]] {
	return func(yield func(J, iter.Seq[K]) bool) {
		for key1, inner := range mms.dataBy1 {
			if !yield(key1, maps.Keys(inner)) {
				return
			}
		}
	}
}

func (mms *MapMapSlice[J, K, V]) SeqKey2Key1() iter.Seq2[K, iter.Seq[J]] {
	return func(yield func(K, iter.Seq[J]) bool) {
		for key2, inner := range mms.dataBy2 {
			if !yield(key2, maps.Keys(inner)) {
				return
			}
		}
	}
}

func (mms *MapMapSlice[J, K, V]) SeqKey1ValueWithKey2(key2 K) iter.Seq2[J, V] {
	return seqSlicedKeyValuesWithKey(key2, mms.dataBy2)
}

func (mms *MapMapSlice[J, K, V]) SeqKey2ValueWithKey1(key1 J) iter.Seq2[K, V] {
	return seqSlicedKeyValuesWithKey(key1, mms.dataBy1)
}

func (mms *MapMapSlice[J, K, V]) SeqValuesWithKey1(key1 J) iter.Seq[V] {
	return seqSlicedValuesWithKey(key1, mms.dataBy1)
}

func (mms *MapMapSlice[J, K, V]) SeqValuesWithKey2(key2 K) iter.Seq[V] {
	return seqSlicedValuesWithKey(key2, mms.dataBy2)
}

func seqSlicedValuesWithKey[J comparable, K comparable, V any](key1 J, mapVar map[J]map[K][]V) iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, slice := range mapVar[key1] {
			for _, value := range slice {
				if !yield(value) {
					return
				}
			}
		}
	}
}

func seqSlicedKeyValuesWithKey[J comparable, K comparable, V any](key1 J, mapVar map[J]map[K][]V) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for key2, slice := range mapVar[key1] {
			for _, value := range slice {
				if !yield(key2, value) {
					return
				}
			}
		}
	}
}

func (mms *MapMapSlice[J, K, V]) Clear() {
	clear(mms.dataBy1)
	clear(mms.dataBy2)
}

func (mms *MapMapSlice[J, K, V]) Add(key1 J, key2 K, value V) {
	var slice []V

	data1 := mms.dataBy1
	if data1 == nil {
		data1 = make(map[J]map[K][]V)
		mms.dataBy1 = data1

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

	putInMap(key2, key1, slice, &mms.dataBy2)
}

func (mms *MapMapSlice[J, K, V]) MapInternalSliceOrPanic(key1 J, key2 K, mapper func([]V) []V) {
	if !mms.MapInternalSliceIfExists(key1, key2, mapper) {
		panic("keys not found")
	}
}

func (mms *MapMapSlice[J, K, V]) MapInternalSliceIfExists(key1 J, key2 K, mapper func([]V) []V) bool {
	data1 := mms.dataBy1
	if data1 != nil {
		inner1, hasInner1 := data1[key1]
		if hasInner1 {
			value1, hasValue1 := inner1[key2]
			if hasValue1 {
				newSlice := mapper(value1)
				if len(newSlice) > 0 {
					inner1[key2] = newSlice

					data2 := mms.dataBy2
					if data2 == nil {
						panic("inconsistent internals")
					} else if inner2, hasInner2 := data2[key2]; !hasInner2 {
						panic("inconsistent internals")
					} else {
						inner2[key1] = newSlice
					}
				} else {
					delete(inner1, key2)

					data2 := mms.dataBy2
					if data2 != nil {
						if inner2, hasInner2 := data2[key2]; hasInner2 {
							delete(inner2, key1)
						}
					}
				}

				return true
			}
		}
	}
	return false
}

// these don't extract into common nicely since struct values would swap
func (mms *MapMapSlice[J, K, V]) SeqKey1Key2ValueEntries() iter.Seq[MapMapEntry[J, K, V]] {
	return func(yield func(MapMapEntry[J, K, V]) bool) {
		for key1, inner := range mms.dataBy1 {
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

// these don't extract into common nicely since struct values would swap
func (mms *MapMapSlice[J, K, V]) SeqKey2Key1ValueEntries() iter.Seq[MapMapEntry[J, K, V]] {
	return func(yield func(MapMapEntry[J, K, V]) bool) {
		for key2, inner := range mms.dataBy2 {
			for key1, slice := range inner {
				for _, value := range slice {
					if !yield(MapMapEntry[J, K, V]{key1, key2, value}) {
						return
					}
				}
			}
		}
	}
}

func (mms *MapMapSlice[J, K, V]) SeqValues() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, inner := range mms.dataBy1 {
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

func (mms *MapMapSlice[J, K, V]) Foreach(apply func(key1 J, key2 K, value V)) {
	for key1, inner := range mms.dataBy1 {
		for key2, slice := range inner {
			for _, value := range slice {
				apply(key1, key2, value)
			}
		}
	}
}

func (mms *MapMapSlice[J, K, V]) SeqKey1Key2ValueSeqEntries() iter.Seq[MapMapSliceEntry[J, K, V]] {
	return func(yield func(MapMapSliceEntry[J, K, V]) bool) {
		for key1, inner := range mms.dataBy1 {
			for key2, slice := range inner {
				if !yield(MapMapSliceEntry[J, K, V]{key1, key2, slices.Values(slice)}) {
					return
				}
			}
		}
	}
}

func (mms *MapMapSlice[J, K, V]) SeqKey2Key1ValueSeqEntries() iter.Seq[MapMapSliceEntry[J, K, V]] {
	return func(yield func(MapMapSliceEntry[J, K, V]) bool) {
		for key2, inner := range mms.dataBy2 {
			for key1, slice := range inner {
				if !yield(MapMapSliceEntry[J, K, V]{key1, key2, slices.Values(slice)}) {
					return
				}
			}
		}
	}
}

func (mms *MapMapSlice[J, K, V]) SeqKey1ValueSeqWithKey2(key2 K) iter.Seq2[J, iter.Seq[V]] {
	return func(yield func(J, iter.Seq[V]) bool) {
		for key1, slice := range mms.dataBy2[key2] {
			if !yield(key1, slices.Values(slice)) {
				return
			}
		}
	}
}

func (mms *MapMapSlice[J, K, V]) SeqKey2ValueSeqWithKey1(key1 J) iter.Seq2[K, iter.Seq[V]] {
	return func(yield func(K, iter.Seq[V]) bool) {
		for key2, slice := range mms.dataBy1[key1] {
			if !yield(key2, slices.Values(slice)) {
				return
			}
		}
	}
}
