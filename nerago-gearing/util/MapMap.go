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

// uses assumtion that empty inner maps don't exist, no maps therefore no data
func (mm *MapMap[J, K, V]) IsEmpty() bool {
	return len(mm.dataBy1) == 0
}

func (mm *MapMap[J, K, V]) Put(key1 J, key2 K, value V) {
	if mm.dataBy1 == nil {
		data1 := make(map[J]map[K]V)
		inner1 := make(map[K]V)
		inner1[key2] = value
		data1[key1] = inner1
		mm.dataBy1 = data1
	} else if inner1, hasInner1 := mm.dataBy1[key1]; !hasInner1 {
		inner1 = make(map[K]V)
		inner1[key2] = value
		mm.dataBy1[key1] = inner1
	} else {
		inner1[key2] = value
	}

	if mm.dataBy2 == nil {
		data2 := make(map[K]map[J]V)
		inner2 := make(map[J]V)
		inner2[key1] = value
		data2[key2] = inner2
		mm.dataBy2 = data2
	} else if inner2, hasInner2 := mm.dataBy2[key2]; !hasInner2 {
		inner2 = make(map[J]V)
		inner2[key1] = value
		mm.dataBy2[key2] = inner2
	} else {
		inner2[key1] = value
	}
}

// removes nested inner maps when last item removed, if changing this check IsEmpty etc
func (mm *MapMap[J, K, V]) Delete(key1 J, key2 K) {
	if mm.dataBy1 != nil {
		inner1, hasInner1 := mm.dataBy1[key1]
		if hasInner1 {
			delete(inner1, key2)
			if len(inner1) == 0 {
				delete(mm.dataBy1, key1)
			}
		}
	}

	if mm.dataBy2 != nil {
		inner2, hasInner2 := mm.dataBy2[key2]
		if hasInner2 {
			delete(inner2, key1)
			if len(inner2) == 0 {
				delete(mm.dataBy2, key2)
			}
		}
	}
}

func (mm *MapMap[J, K, V]) DeleteAllForKey1(key1 J) {
	for key2 := range mm.dataBy1[key1] {
		inner2, hasInner2 := mm.dataBy2[key2]
		if hasInner2 {
			delete(inner2, key1)
			if len(inner2) == 0 {
				delete(mm.dataBy2, key2)
			}
		}
	}
	delete(mm.dataBy1, key1)
}

func (mm *MapMap[J, K, V]) DeleteAllForKey2(key2 K) {
	for key1 := range mm.dataBy2[key2] {
		inner1, hasInner1 := mm.dataBy1[key1]
		if hasInner1 {
			delete(inner1, key2)
			if len(inner1) == 0 {
				delete(mm.dataBy1, key1)
			}
		}
	}
	delete(mm.dataBy2, key2)
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

	if mm.dataBy2 == nil {
		data2 := make(map[K]map[J]V)
		inner2 := make(map[J]V)
		inner2[key1] = value
		data2[key2] = inner2
		mm.dataBy2 = data2
	} else if inner2, hasInner2 := mm.dataBy2[key2]; !hasInner2 {
		inner2 = make(map[J]V)
		inner2[key1] = value
		mm.dataBy2[key2] = inner2
	} else {
		inner2[key1] = value
	}
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

func (mm *MapMap[J, K, V]) SeqWithKeys() iter.Seq[MapMapEntry[J, K, V]] {
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

func (mm *MapMap[J, K, V]) SeqWithKeysOtherOrder() iter.Seq[MapMapEntry[J, K, V]] {
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

func (mm *MapMap[J, K, V]) ForeachWithKeys(apply func(key1 J, key2 K, value V)) {
	for key1, inner := range mm.dataBy1 {
		for key2, value := range inner {
			apply(key1, key2, value)
		}
	}
}

func (mm *MapMap[J, K, V]) SeqInnerWithKey1Value(key1 J) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for key2, value := range mm.dataBy1[key1] {
			if !yield(key2, value) {
				return
			}
		}
	}

}

func (mm *MapMap[J, K, V]) SeqInnerWithKey2Value(key2 K) iter.Seq2[J, V] {
	return func(yield func(J, V) bool) {
		for key1, value := range mm.dataBy2[key2] {
			if !yield(key1, value) {
				return
			}
		}
	}
}

func (mm *MapMap[J, K, V]) SeqGroupsKey1Lookup() iter.Seq2[J, func(K) V] {
	return func(yield func(J, func(K) V) bool) {
		for key1, inner := range mm.dataBy1 {
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

func (mm *MapMap[J, K, V]) SeqGroupsKey2Lookup() iter.Seq2[K, func(J) V] {
	return func(yield func(K, func(J) V) bool) {
		for key2, inner := range mm.dataBy2 {
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

func (mm *MapMap[J, K, V]) SeqGroupsKey1NestedKeyValue() iter.Seq2[J, iter.Seq2[K, V]] {
	return func(yield func(J, iter.Seq2[K, V]) bool) {
		for key1, inner := range mm.dataBy1 {
			if !yield(key1, maps.All(inner)) {
				return
			}
		}
	}
}

func (mm *MapMap[J, K, V]) SeqGroupsKey2NestedKeyValue() iter.Seq2[K, iter.Seq2[J, V]] {
	return func(yield func(K, iter.Seq2[J, V]) bool) {
		for key2, inner := range mm.dataBy2 {
			if !yield(key2, maps.All(inner)) {
				return
			}
		}
	}
}

func (mm *MapMap[J, K, V]) SeqKey1Key2Nested() iter.Seq2[J, iter.Seq[K]] {
	return func(yield func(J, iter.Seq[K]) bool) {
		for key1, inner := range mm.dataBy1 {
			if !yield(key1, maps.Keys(inner)) {
				return
			}
		}
	}
}

func (mm *MapMap[J, K, V]) SeqKey2Key1Nested() iter.Seq2[K, iter.Seq[J]] {
	return func(yield func(K, iter.Seq[J]) bool) {
		for key2, inner := range mm.dataBy2 {
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
