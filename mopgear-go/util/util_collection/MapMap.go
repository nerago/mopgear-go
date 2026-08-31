package util_collection

import (
	"iter"
	"maps"
)

type MapMap[J comparable, K comparable, V any] struct {
	dataBy1 map[J]map[K]V
	dataBy2 map[K]map[J]V
}

var _ IMapMap[int, int, int] = &MapMap[int, int, int]{}

type MapMapEntry[J comparable, K comparable, V any] struct {
	Key1  J
	Key2  K
	Value V
}

// optional
func (mm *MapMap[J, K, V]) Init(size int) {
	mm.dataBy1 = make(map[J]map[K]V, size)
	mm.dataBy2 = make(map[K]map[J]V, size)
}

func (mm *MapMap[J, K, V]) Get(key1 J, key2 K) (V, bool) {
	value, hasValue := mm.dataBy1[key1][key2]
	return value, hasValue
}

func (mm *MapMap[J, K, V]) GetOrPanic(key1 J, key2 K) V {
	value, hasValue := mm.dataBy1[key1][key2]
	if hasValue {
		return value
	} else {
		panic("value not found")
	}
}

func (mm *MapMap[J, K, V]) Has(key1 J, key2 K) bool {
	_, hasValue := mm.dataBy1[key1][key2]
	return hasValue
}

func (mm *MapMap[J, K, V]) HasKey1(key1 J) bool {
	_, hasValue := mm.dataBy1[key1]
	return hasValue
}

func (mm *MapMap[J, K, V]) HasKey2(key2 K) bool {
	_, hasValue := mm.dataBy2[key2]
	return hasValue
}

func (mm *MapMap[J, K, V]) Clear() {
	clear(mm.dataBy1)
	clear(mm.dataBy2)
}

func (mm *MapMap[J, K, V]) Size() int {
	size := 0
	for _, inner := range mm.dataBy1 {
		size += len(inner)
	}
	return size
}

// uses assumption that empty inner maps don't exist, no maps therefore no data
func (mm *MapMap[J, K, V]) IsEmpty() bool {
	return len(mm.dataBy1) == 0
}

func (mm *MapMap[J, K, V]) Put(key1 J, key2 K, value V) {
	putInMap(key1, key2, value, &mm.dataBy1)
	putInMap(key2, key1, value, &mm.dataBy2)
}

func putInMap[J comparable, K comparable, V any](key1 J, key2 K, value V, dataPointer *map[J]map[K]V) {
	if *dataPointer == nil {
		data1 := make(map[J]map[K]V)
		inner1 := make(map[K]V)
		inner1[key2] = value
		data1[key1] = inner1
		*dataPointer = data1
	} else if inner1, hasInner1 := (*dataPointer)[key1]; !hasInner1 {
		inner1 = make(map[K]V)
		inner1[key2] = value
		(*dataPointer)[key1] = inner1
	} else {
		inner1[key2] = value
	}
}

// removes nested inner maps when last item removed, if changing this check IsEmpty etc
func (mm *MapMap[J, K, V]) Delete(key1 J, key2 K) {
	deleteInMap(key1, key2, mm.dataBy1)
	deleteInMap(key2, key1, mm.dataBy2)
}

func deleteInMap[J comparable, K comparable, V any](key1 J, key2 K, mapVar map[J]map[K]V) {
	inner1, hasInner1 := mapVar[key1]
	if hasInner1 {
		delete(inner1, key2)
		if len(inner1) == 0 {
			delete(mapVar, key1)
		}
	}
}

func (mm *MapMap[J, K, V]) DeleteAllForKey1(key1 J) {
	deleteAllNestedIn(key1, mm.dataBy1, mm.dataBy2)
	delete(mm.dataBy1, key1)
}

func (mm *MapMap[J, K, V]) DeleteAllForKey2(key2 K) {
	deleteAllNestedIn(key2, mm.dataBy2, mm.dataBy1)
	delete(mm.dataBy2, key2)
}

func deleteAllNestedIn[J comparable, K comparable, V any](key1 K, mapVar1 map[K]map[J]V, mapVar2 map[J]map[K]V) {
	for key2 := range mapVar1[key1] {
		inner2, hasInner2 := mapVar2[key2]
		if hasInner2 {
			delete(inner2, key1)
			if len(inner2) == 0 {
				delete(mapVar2, key2)
			}
		}
	}
}

func (mm *MapMap[J, K, V]) Apply(key1 J, key2 K, apply func(oldValue V) V) {
	var value V

	if mm.dataBy1 == nil {
		data1 := make(map[J]map[K]V)
		inner1 := make(map[K]V)

		value = apply(value)
		inner1[key2] = value

		data1[key1] = inner1
		mm.dataBy1 = data1
	} else if inner1, hasInner1 := mm.dataBy1[key1]; !hasInner1 {
		inner1 = make(map[K]V)

		value = apply(value)
		inner1[key2] = value

		mm.dataBy1[key1] = inner1
	} else {
		value = apply(inner1[key2])
		inner1[key2] = value
	}

	putInMap(key2, key1, value, &mm.dataBy2)
}

func (mm *MapMap[J, K, V]) FirstKey1() J {
	for x := range mm.dataBy1 {
		return x
	}
	panic("empty map")
}

func (mm *MapMap[J, K, V]) FirstKey2() K {
	for x := range mm.dataBy2 {
		return x
	}
	panic("empty map")
}

func (mm *MapMap[J, K, V]) SeqKey1() iter.Seq[J] {
	return maps.Keys(mm.dataBy1)
}

func (mm *MapMap[J, K, V]) SeqKey2() iter.Seq[K] {
	return maps.Keys(mm.dataBy2)
}

func (mm *MapMap[J, K, V]) SeqValues() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, inner := range mm.dataBy1 {
			for _, value := range inner {
				if !yield(value) {
					return
				}
			}
		}
	}
}

func (mm *MapMap[J, K, V]) SeqValuesWithKey1(key1 J) iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, value := range mm.dataBy1[key1] {
			if !yield(value) {
				return
			}
		}
	}
}

func (mm *MapMap[J, K, V]) SeqValuesWithKey2(key2 K) iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, value := range mm.dataBy2[key2] {
			if !yield(value) {
				return
			}
		}
	}
}

func (mm *MapMap[J, K, V]) SeqKey1Key2ValueEntries() iter.Seq[MapMapEntry[J, K, V]] {
	return func(yield func(MapMapEntry[J, K, V]) bool) {
		for key1, inner := range mm.dataBy1 {
			for key2, value := range inner {
				if !yield(MapMapEntry[J, K, V]{key1, key2, value}) {
					return
				}
			}
		}
	}
}

func (mm *MapMap[J, K, V]) SeqKey2Key1ValueEntries() iter.Seq[MapMapEntry[J, K, V]] {
	return func(yield func(MapMapEntry[J, K, V]) bool) {
		for key2, inner := range mm.dataBy2 {
			for key1, value := range inner {
				if !yield(MapMapEntry[J, K, V]{key1, key2, value}) {
					return
				}
			}
		}
	}
}

func (mm *MapMap[J, K, V]) Foreach(apply func(key1 J, key2 K, value V)) {
	for key1, inner := range mm.dataBy1 {
		for key2, value := range inner {
			apply(key1, key2, value)
		}
	}
}

func (mm *MapMap[J, K, V]) SeqKey2ValueWithKey1(key1 J) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for key2, value := range mm.dataBy1[key1] {
			if !yield(key2, value) {
				return
			}
		}
	}
}

func (mm *MapMap[J, K, V]) SeqKey1ValueWithKey2(key2 K) iter.Seq2[J, V] {
	return func(yield func(J, V) bool) {
		for key1, value := range mm.dataBy2[key2] {
			if !yield(key1, value) {
				return
			}
		}
	}
}

func (mm *MapMap[J, K, V]) SeqKey1NestedKey2Value() iter.Seq2[J, iter.Seq2[K, V]] {
	return func(yield func(J, iter.Seq2[K, V]) bool) {
		for key1, inner := range mm.dataBy1 {
			if !yield(key1, maps.All(inner)) {
				return
			}
		}
	}
}

func (mm *MapMap[J, K, V]) SeqKey2NestedKey1Value() iter.Seq2[K, iter.Seq2[J, V]] {
	return func(yield func(K, iter.Seq2[J, V]) bool) {
		for key2, inner := range mm.dataBy2 {
			if !yield(key2, maps.All(inner)) {
				return
			}
		}
	}
}

func (mm *MapMap[J, K, V]) SeqKey1Key2() iter.Seq2[J, iter.Seq[K]] {
	return func(yield func(J, iter.Seq[K]) bool) {
		for key1, inner := range mm.dataBy1 {
			if !yield(key1, maps.Keys(inner)) {
				return
			}
		}
	}
}

func (mm *MapMap[J, K, V]) SeqKeysAll() iter.Seq2[J, K] {
	return func(yield func(J, K) bool) {
		for key1, inner := range mm.dataBy1 {
			for key2 := range inner {
				if !yield(key1, key2) {
					return
				}
			}
		}
	}
}

func (mm *MapMap[J, K, V]) SeqKey2Key1() iter.Seq2[K, iter.Seq[J]] {
	return func(yield func(K, iter.Seq[J]) bool) {
		for key2, inner := range mm.dataBy2 {
			if !yield(key2, maps.Keys(inner)) {
				return
			}
		}
	}
}

func (mm *MapMap[J, K, V]) Equals(other *MapMap[J, K, V], valueEqual func(*V, *V) bool) bool {
	if len(mm.dataBy1) != len(other.dataBy1) {
		return false
	}
	for key1, inner := range mm.dataBy1 {
		otherInner, hasOtherInner := other.dataBy1[key1]
		if hasOtherInner && len(inner) == len(otherInner) {
			for key2, value := range inner {
				otherValue, hasOtherValue := otherInner[key2]
				if !hasOtherValue || !valueEqual(&value, &otherValue) {
					return false
				}
			}
		} else {
			return false
		}
	}
	return true
}

// a "map operation" in map/reduce terms, but too many uses for the word "map" here
func MapMap_FromExistingMapMap_WithApply[J comparable, K comparable, V any, R any](mm *MapMap[J, K, V], apply func(V) R) *MapMap[J, K, R] {
	return MapMap_FromExistingMapMap_WithApplyPlusKeys[J, K, V, R](mm, func(j J, k K, v V) R {
		return apply(v)
	})
}

func MapMap_FromExistingMapMap_WithApplyPlusKeys[J comparable, K comparable, V any, R any](mm *MapMap[J, K, V], apply func(J, K, V) R) *MapMap[J, K, R] {
	resultMap := MapMap[J, K, R]{
		dataBy1: make(map[J]map[K]R, len(mm.dataBy1)),
		dataBy2: make(map[K]map[J]R, len(mm.dataBy2)),
	}

	for key2, inner := range mm.dataBy2 {
		resultMap.dataBy2[key2] = make(map[J]R, len(inner))
	}

	for key1, inner := range mm.dataBy1 {
		newInner := make(map[K]R, len(inner))
		for key2, value := range inner {
			newValue := apply(key1, key2, value)
			newInner[key2] = newValue
			resultMap.dataBy2[key2][key1] = newValue
		}
		resultMap.dataBy1[key1] = newInner
	}

	return &resultMap
}
