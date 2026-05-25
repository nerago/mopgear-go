package util

import (
	"iter"
)

type MapMap[J comparable, K comparable, V any] struct {
	dataBy1 map[J]map[K]V
	dataBy2 map[K]map[J]V
}

type MapMapEntry[J comparable, K comparable, V any] struct {
	key1  J
	key2  K
	value V
}

func (mapmap *MapMap[J, K, V]) Get(key1 J, key2 K) (V, bool) {
	data := mapmap.dataBy1
	if data != nil {
		inner, hasInner := data[key1]
		if hasInner {
			value, hasValue := inner[key2]
			return value, hasValue
		}
	}

	var nilValue V
	return nilValue, false
}

func (mapmap *MapMap[J, K, V]) GetOrPanic(key1 J, key2 K) V {
	data := mapmap.dataBy1
	if data != nil {
		inner, hasInner := data[key1]
		if hasInner {
			value, hasValue := inner[key2]
			if hasValue {
				return value
			}
		}
	}

	panic("value not found")
}

func (mapmap *MapMap[J, K, V]) Has(key1 J, key2 K) bool {
	data := mapmap.dataBy1
	if data != nil {
		inner, hasInner := data[key1]
		if hasInner {
			_, hasValue := inner[key2]
			return hasValue
		}
	}
	return false
}

func (mapmap *MapMap[J, K, V]) Clear() {
	clear(mapmap.dataBy1)
	clear(mapmap.dataBy2)
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

func (mapmap *MapMap[J, K, V]) SeqWithKeys() iter.Seq[MapMapEntry[J, K, V]] {
	if mapmap.dataBy1 != nil {
		return func(yield func(MapMapEntry[J, K, V]) bool) {
			for key1, inner := range mapmap.dataBy1 {
				for key2, value := range inner {
					if !yield(MapMapEntry[J, K, V]{key1, key2, value}) {
						return
					}
				}
			}
		}
	} else {
		return func(yield func(MapMapEntry[J, K, V]) bool) {}
	}
}

func (mapmap *MapMap[J, K, V]) ForeachWithKeys(apply func(key1 J, key2 K, value V)) {
	if mapmap.dataBy1 != nil {
		for key1, inner := range mapmap.dataBy1 {
			for key2, value := range inner {
				apply(key1, key2, value)
			}
		}
	}
}

func (mapmap *MapMap[J, K, V]) ForeachGroupForKey1(apply func(key1 J, lookup func(K) V)) {
	if mapmap.dataBy1 != nil {
		for key1, inner := range mapmap.dataBy1 {
			lookup := func(key2 K) V {
				value, hasValue := inner[key2]
				if hasValue {
					return value
				} else {
					panic("value not found")
				}
			}
			apply(key1, lookup)
		}
	}
}

func (mapmap *MapMap[J, K, V]) ForeachGroupForKey2(apply func(key2 K, lookup func(J) V)) {
	if mapmap.dataBy2 != nil {
		for key2, inner := range mapmap.dataBy2 {
			lookup := func(key1 J) V {
				value, hasValue := inner[key1]
				if hasValue {
					return value
				} else {
					panic("value not found")
				}
			}
			apply(key2, lookup)
		}
	}
}

func (mapmap *MapMap[J, K, V]) ForeachInnerWithKey1Value(key1 J, apply func(key2 K, value V)) {
	if mapmap.dataBy1 != nil {
		inner, hasInner := mapmap.dataBy1[key1]
		if hasInner {
			for key2, value := range inner {
				apply(key2, value)
			}
		}
	}
}

func (mapmap *MapMap[J, K, V]) ForeachInnerWithKey2Value(key2 K, apply func(key1 J, value V)) {
	if mapmap.dataBy2 != nil {
		inner, hasInner := mapmap.dataBy2[key2]
		if hasInner {
			for key1, value := range inner {
				apply(key1, value)
			}
		}
	}
}

func (mapmap *MapMap[J, K, V]) SeqGroupsKey1() iter.Seq2[J, func(K) V] {
	if mapmap.dataBy1 != nil {
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
	} else {
		return func(yield func(J, func(K) V) bool) {}
	}
}

func (mapmap *MapMap[J, K, V]) SeqGroupsKey2() iter.Seq2[K, func(J) V] {
	if mapmap.dataBy2 != nil {
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
	} else {
		return func(yield func(K, func(J) V) bool) {}
	}
}
