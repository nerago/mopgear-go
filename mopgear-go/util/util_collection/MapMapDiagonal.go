package util_collection

import (
	"iter"
	"maps"
)

type MapMapDiagonal[K comparable, V any] struct {
	data map[K]map[K]V
}

// optional
func (mmd *MapMapDiagonal[K, V]) Init(size int) {
	mmd.data = make(map[K]map[K]V, size)
}

func (mmd *MapMapDiagonal[K, V]) Get(key1 K, key2 K) (V, bool) {
	value, hasValue := mmd.data[key1][key2]
	if hasValue {
		return value, hasValue
	}

	value, hasValue = mmd.data[key2][key1]
	return value, hasValue
}

func (mmd *MapMapDiagonal[K, V]) GetOrPanic(key1 K, key2 K) V {
	value, hasValue := mmd.data[key1][key2]
	if hasValue {
		return value
	}

	value, hasValue = mmd.data[key2][key1]
	if hasValue {
		return value
	} else {
		panic("value not found")
	}
}

func (mmd *MapMapDiagonal[K, V]) Has(key1 K, key2 K) bool {
	_, hasValue := mmd.data[key1][key2]
	if hasValue {
		return true
	}

	_, hasValue = mmd.data[key2][key1]
	return hasValue
}

func (mmd *MapMapDiagonal[K, V]) Clear() {
	clear(mmd.data)
}

func (mmd *MapMapDiagonal[K, V]) Size() int {
	size := 0
	for _, inner := range mmd.data {
		size += len(inner)
	}
	return size
}

func (mmd *MapMapDiagonal[K, V]) Put(key1 K, key2 K, value V) {
	data := mmd.data
	if data == nil {
		data = make(map[K]map[K]V)
		mmd.data = data
	}

	inner1, hasInner1 := data[key1]
	inner2, hasInner2 := data[key2]

	if hasInner1 && hasInner2 {
		_, hasValue1 := inner1[key2]
		_, hasValue2 := inner2[key1]

		if hasValue1 && hasValue2 {
			if key1 == key2 {
				inner1[key2] = value
			} else {
				panic("duplicate storage")
			}
		} else if hasValue1 {
			inner1[key2] = value
		} else {
			inner2[key1] = value
		}
	} else if hasInner1 {
		inner1[key2] = value
	} else if hasInner2 {
		inner2[key1] = value
	} else {
		inner1 = make(map[K]V)
		data[key1] = inner1
		inner1[key2] = value
	}
}

func (mmd *MapMapDiagonal[K, V]) Apply(key1 K, key2 K, apply func(oldValue V) V) {
	data := mmd.data
	if data == nil {
		data = make(map[K]map[K]V)
		mmd.data = data
	}

	inner1, hasInner1 := data[key1]
	inner2, hasInner2 := data[key2]

	if hasInner1 && hasInner2 {
		value1, hasValue1 := inner1[key2]
		value2, hasValue2 := inner2[key1]
		if hasValue1 && hasValue2 {
			panic("duplicate")
		} else if hasValue1 {
			inner1[key2] = apply(value1)
		} else {
			inner2[key1] = apply(value2)
		}
	} else if hasInner1 {
		inner1[key2] = apply(inner1[key2])
	} else if hasInner2 {
		inner2[key1] = apply(inner2[key1])
	} else {
		inner1 = make(map[K]V)
		data[key1] = inner1

		var nilValue V
		inner1[key2] = apply(nilValue)
	}
}

func (mmd *MapMapDiagonal[K, V]) SeqValues() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, inner := range mmd.data {
			for _, value := range inner {
				if !yield(value) {
					return
				}
			}
		}
	}
}

func (mmd *MapMapDiagonal[K, V]) SeqWithKeys() iter.Seq[MapMapEntry[K, K, V]] {
	return func(yield func(MapMapEntry[K, K, V]) bool) {
		for key1, inner := range mmd.data {
			for key2, value := range inner {
				if !yield(MapMapEntry[K, K, V]{key1, key2, value}) {
					return
				}
			}
		}
	}
}

func (mmd *MapMapDiagonal[K, V]) ForeachWithKeys(apply func(key1 K, key2 K, value V)) {
	for key1, inner := range mmd.data {
		for key2, value := range inner {
			apply(key1, key2, value)
		}
	}
}

func (mmd *MapMapDiagonal[K, V]) SeqInnerWithKeyValue(key1 K) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for a, inner := range mmd.data {
			if a == key1 {
				for b, value := range inner {
					if !yield(b, value) {
						return
					}
				}
			} else {
				value, hasValue := inner[key1]
				if hasValue && !yield(a, value) {
					return
				}
			}
		}
	}
}

func (mmd *MapMapDiagonal[K, V]) SeqGroupsKey1NestedKeyValue() iter.Seq2[K, iter.Seq2[K, V]] {
	return func(yield func(K, iter.Seq2[K, V]) bool) {
		for key1, inner := range mmd.data {
			if !yield(key1, maps.All(inner)) {
				return
			}
		}
	}
}

func (mmd *MapMapDiagonal[K, V]) SeqKey1Key2Nested() iter.Seq2[K, iter.Seq[K]] {
	return func(yield func(K, iter.Seq[K]) bool) {
		for key1, inner := range mmd.data {
			if !yield(key1, maps.Keys(inner)) {
				return
			}
		}
	}
}
