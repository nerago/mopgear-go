package util

import (
	"iter"
	"maps"
)

type MapMap[J comparable, K comparable, V any] struct {
	dataBy1 map[J]map[K]V
	dataBy2 map[K]map[J]V
}


type MapMapEntry[J comparable, K comparable, V any] struct {
	Key1  J
	Key2  K
	Value V
}

// optional
func (mapmap *MapMap[J, K, V]) Init(size int) {
	mapmap.dataBy1 = make(map[J]map[K]V, size)
	mapmap.dataBy2 = make(map[K]map[J]V, size)
}

func (mapmap *MapMap[J, K, V]) Get(key1 J, key2 K) (V, bool) {
	value, hasValue := mapmap.dataBy1[key1][key2]
	return value, hasValue
}

func (mapmap *MapMap[J, K, V]) GetOrPanic(key1 J, key2 K) V {
	value, hasValue := mapmap.dataBy1[key1][key2]
	if hasValue {
		return value
	} else {
		panic("value not found")
	}
}

func (mapmap *MapMap[J, K, V]) Has(key1 J, key2 K) bool {
	_, hasValue := mapmap.dataBy1[key1][key2]
	return hasValue
}

func (mapmap *MapMap[J, K, V]) Clear() {
	clear(mapmap.dataBy1)
	clear(mapmap.dataBy2)
}

func (mapmap MapMap[J, K, V]) Size() int {
	size := 0
	for _, inner := range mapmap.dataBy1 {
		size += len(inner)
	}
	return size
}

func (mapmap *MapMap[J, K, V]) Put(key1 J, key2 K, value V) {
	data1 := mapmap.dataBy1
	if data1 == nil {
		data1 = make(map[J]map[K]V)
		mapmap.dataBy1 = data1
	}
	inner1, hasInner1 := data1[key1]
	if !hasInner1 {
		inner1 = make(map[K]V)
		data1[key1] = inner1
	}
	inner1[key2] = value

	data2 := mapmap.dataBy2
	if data2 == nil {
		data2 = make(map[K]map[J]V)
		mapmap.dataBy2 = data2
	}
	inner2, hasInner2 := data2[key2]
	if !hasInner2 {
		inner2 = make(map[J]V)
		data2[key2] = inner2
	}
	inner2[key1] = value
}

func (mapmap *MapMap[J, K, V]) Apply(key1 J, key2 K, apply func(oldValue V) V) {
	data1 := mapmap.dataBy1
	if data1 == nil {
		data1 = make(map[J]map[K]V)
		mapmap.dataBy1 = data1
	}
	inner1, hasInner1 := data1[key1]
	if !hasInner1 {
		inner1 = make(map[K]V)
		data1[key1] = inner1
	}
	value, hasValue := inner1[key2]
	if hasValue {
		value = apply(value)
	} else {
		var nilValue V
		value = apply(nilValue)
	}
	inner1[key2] = value

	data2 := mapmap.dataBy2
	if data2 == nil {
		data2 = make(map[K]map[J]V)
		mapmap.dataBy2 = data2
	}
	inner2, hasInner2 := data2[key2]
	if !hasInner2 {
		inner2 = make(map[J]V)
		data2[key2] = inner2
	}
	inner2[key1] = value
}

func (mapmap *MapMap[J, K, V]) SeqKey1() iter.Seq[J] {
	return maps.Keys(mapmap.dataBy1)
}

func (mapmap *MapMap[J, K, V]) SeqKey2() iter.Seq[K] {
	return maps.Keys(mapmap.dataBy2)
}

func (mapmap *MapMap[J, K, V]) SeqValues() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, inner := range mapmap.dataBy1 {
			for _, value := range inner {
				if !yield(value) {
					return
				}
			}
		}
	}
}

func (mapmap *MapMap[J, K, V]) SeqWithKeys() iter.Seq[MapMapEntry[J, K, V]] {
	return func(yield func(MapMapEntry[J, K, V]) bool) {
		for key1, inner := range mapmap.dataBy1 {
			for key2, value := range inner {
				if !yield(MapMapEntry[J, K, V]{key1, key2, value}) {
					return
				}
			}
		}
	}
}

func (mapmap *MapMap[J, K, V]) SeqWithKeysOtherOrder() iter.Seq[MapMapEntry[J, K, V]] {
	return func(yield func(MapMapEntry[J, K, V]) bool) {
		for key2, inner := range mapmap.dataBy2 {
			for key1, value := range inner {
				if !yield(MapMapEntry[J, K, V]{key1, key2, value}) {
					return
				}
			}
		}
	}
}

func (mapmap *MapMap[J, K, V]) ForeachWithKeys(apply func(key1 J, key2 K, value V)) {
	for key1, inner := range mapmap.dataBy1 {
		for key2, value := range inner {
			apply(key1, key2, value)
		}
	}
}

func (mapmap *MapMap[J, K, V]) SeqInnerWithKey1Value(key1 J) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for key2, value := range mapmap.dataBy1[key1] {
			if !yield(key2, value) {
				return
			}
		}
	}

}

func (mapmap *MapMap[J, K, V]) SeqInnerWithKey2Value(key2 K) iter.Seq2[J, V] {
	return func(yield func(J, V) bool) {
		for key1, value := range mapmap.dataBy2[key2] {
			if !yield(key1, value) {
				return
			}
		}
	}
}

func (mapmap *MapMap[J, K, V]) SeqGroupsKey1Lookup() iter.Seq2[J, func(K) V] {
	return func(yield func(J, func(K) V) bool) {
		for key1, inner := range mapmap.dataBy1 {
			lookup := func(key2 K) V {
				value, hasValue := inner[key2]
				if hasValue {
					return value
				} else {
					panic("value not found")
				}
			}

			if !yield(key1, lookup) {
				return
			}
		}
	}
}

func (mapmap *MapMap[J, K, V]) SeqGroupsKey2Lookup() iter.Seq2[K, func(J) V] {
	return func(yield func(K, func(J) V) bool) {
		for key2, inner := range mapmap.dataBy2 {
			lookup := func(key1 J) V {
				value, hasValue := inner[key1]
				if hasValue {
					return value
				} else {
					panic("value not found")
				}
			}

			if !yield(key2, lookup) {
				return
			}
		}
	}
}

func (mapmap *MapMap[J, K, V]) SeqGroupsKey1NestedKeyValue() iter.Seq2[J, iter.Seq2[K, V]] {
	return func(yield func(J, iter.Seq2[K, V]) bool) {
		for key1, inner := range mapmap.dataBy1 {
			if !yield(key1, maps.All(inner)) {
				return
			}
		}
	}
}

func (mapmap *MapMap[J, K, V]) SeqGroupsKey2NestedKeyValue() iter.Seq2[K, iter.Seq2[J, V]] {
	return func(yield func(K, iter.Seq2[J, V]) bool) {
		for key2, inner := range mapmap.dataBy2 {
			if !yield(key2, maps.All(inner)) {
				return
			}
		}
	}
}

func (mapmap *MapMap[J, K, V]) SeqKey1Key2Nested() iter.Seq2[J, iter.Seq[K]] {
	return func(yield func(J, iter.Seq[K]) bool) {
		for key1, inner := range mapmap.dataBy1 {
			if !yield(key1, maps.Keys(inner)) {
				return
			}
		}
	}
}

func (mapmap *MapMap[J, K, V]) SeqKey2Key1Nested() iter.Seq2[K, iter.Seq[J]] {
	return func(yield func(K, iter.Seq[J]) bool) {
		for key2, inner := range mapmap.dataBy2 {
			if !yield(key2, maps.Keys(inner)) {
				return
			}
		}
	}
}

// a "map operation" in map/reduce terms, but too many uses for the word "map" here
func MapMap_FromExitingMapMap_WithApply[J comparable, K comparable, V any, R any](mapmap *MapMap[J, K, V], apply func(V) R) *MapMap[J, K, R] {
	resultMap := MapMap[J, K, R]{}

	resultMap.dataBy2 = make(map[K]map[J]R, len(mapmap.dataBy2))
	for key2, inner := range mapmap.dataBy2 {
		resultMap.dataBy2[key2] = make(map[J]R, len(inner))
	}

	resultMap.dataBy1 = make(map[J]map[K]R, len(mapmap.dataBy1))
	for key1, inner := range mapmap.dataBy1 {
		newInner := make(map[K]R, len(inner))
		for key2, value := range inner {
			newValue := apply(value)
			newInner[key2] = newValue
			resultMap.dataBy2[key2][key1] = newValue
		}
		resultMap.dataBy1[key1] = newInner
	}

	return &resultMap
}

func MapMap_FromExitingMapMap_WithApplyPlusKeys[J comparable, K comparable, V any, R any](mapmap *MapMap[J, K, V], apply func(J, K, V) R) *MapMap[J, K, R] {
	resultMap := MapMap[J, K, R]{}

	resultMap.dataBy2 = make(map[K]map[J]R, len(mapmap.dataBy2))
	for key2, inner := range mapmap.dataBy2 {
		resultMap.dataBy2[key2] = make(map[J]R, len(inner))
	}

	resultMap.dataBy1 = make(map[J]map[K]R, len(mapmap.dataBy1))
	for key1, inner := range mapmap.dataBy1 {
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
