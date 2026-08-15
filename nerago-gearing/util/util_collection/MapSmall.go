package util_collection

import (
	"iter"
	"slices"
)

type MapSmallFixed[K comparable, V comparable] struct {
	keys   []K
	values []V
	size   uint8
}

func MapSmallFixedMake[K comparable, V comparable](maxSize uint8) *MapSmallFixed[K, V] {
	return &MapSmallFixed[K, V]{
		make([]K, maxSize),
		make([]V, maxSize),
		0,
	}
}

func (ms *MapSmallFixed[K, V]) Clear() {
	clear(ms.keys)
	clear(ms.values)
	ms.size = 0
}

func (ms *MapSmallFixed[K, V]) Size() int {
	return int(ms.size)
}

func (ms *MapSmallFixed[K, V]) IsEmpty() bool {
	return ms.size == 0
}

func (ms *MapSmallFixed[K, V]) SeqValues() iter.Seq[V] {
	return func(yield func(V) bool) {
		for i := range ms.size {
			if !yield(ms.values[i]) {
				return
			}
		}
	}
}

func (ms *MapSmallFixed[K, V]) EqualsInterface(other IMap[K, V], elementEqual func(*V, *V) bool) bool {
	if otherSmall, isSmall := other.(*MapSmallFixed[K, V]); isSmall {
		return ms.Equals(otherSmall)
	} else {
		return IMapEquals(ms, other, elementEqual)
	}
}

func (ms *MapSmallFixed[K, V]) Equals(other *MapSmallFixed[K, V]) bool {
	if ms.size != other.size {
		return false
	}

outerLoop:
	for a := range ms.size {
		keyA := ms.keys[a]
		for b := range ms.size {
			keyB := other.keys[b]
			if keyA == keyB {
				if ms.values[a] == other.values[b] {
					continue outerLoop
				} else {
					return false
				}
			}
		}
		// key not found
		return false
	}

	return true
}

func (ms *MapSmallFixed[K, V]) Has(key K) bool {
	for i := range ms.size {
		if ms.keys[i] == key {
			return true
		}
	}
	return false
}

func (ms *MapSmallFixed[K, V]) FirstKey() K {
	if ms.size > 0 {
		return ms.keys[0]
	} else {
		panic("no keys")
	}
}

func (ms *MapSmallFixed[K, V]) Get(key K) (V, bool) {
	for i := range ms.size {
		if ms.keys[i] == key {
			return ms.values[i], true
		}
	}
	var nilValue V
	return nilValue, false
}

func (ms *MapSmallFixed[K, V]) GetOrNilValue(key K) V {
	for i := range ms.size {
		if ms.keys[i] == key {
			return ms.values[i]
		}
	}
	var nilValue V
	return nilValue
}

func (ms *MapSmallFixed[K, V]) GetOrDefault(key K, defaultValue V) V {
	for i := range ms.size {
		if ms.keys[i] == key {
			return ms.values[i]
		}
	}
	return defaultValue
}

func (ms *MapSmallFixed[K, V]) GetOrPanic(key K) V {
	for i := range ms.size {
		if ms.keys[i] == key {
			return ms.values[i]
		}
	}
	panic("key missing")
}

func (ms *MapSmallFixed[K, V]) Put(key K, value V) {
	for i := range ms.size {
		if ms.keys[i] == key {
			ms.values[i] = value
			return
		}
	}

	if len(ms.keys) == int(ms.size) {
		panic("map full")
	}

	ms.keys[ms.size] = key
	ms.values[ms.size] = value
	ms.size++
}

func (ms *MapSmallFixed[K, V]) Compute(key K, apply func(V) V) {
	for i := range ms.size {
		if ms.keys[i] == key {
			ms.values[i] = apply(ms.values[i])
			return
		}
	}

	if len(ms.keys) == int(ms.size) {
		panic("map full")
	}

	var nilValue V
	ms.keys[ms.size] = key
	ms.values[ms.size] = apply(nilValue)
	ms.size++
}

func (ms *MapSmallFixed[K, V]) Delete(key K) {
	for i := range ms.size {
		if ms.keys[i] == key {
			DeleteIndexInPlace(&ms.keys, int(i))
			DeleteIndexInPlace(&ms.values, int(i))
			ms.size--
			return
		}
	}
}

func (ms *MapSmallFixed[K, V]) Foreach(apply func(key K, value V)) {
	for i := range ms.size {
		apply(ms.keys[i], ms.values[i])
	}
}

func (ms *MapSmallFixed[K, V]) SeqKeyValue() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for i := range ms.size {
			if !yield(ms.keys[i], ms.values[i]) {
				return
			}
		}
	}
}

func (ms *MapSmallFixed[K, V]) SeqKey() iter.Seq[K] {
	return func(yield func(K) bool) {
		for i := range ms.size {
			if !yield(ms.keys[i]) {
				return
			}
		}
	}
}

func (ms *MapSmallFixed[K, V]) KeySlice() []K {
	return slices.Clone(ms.keys)
}

func (ms *MapSmallFixed[K, V]) ValueSlice() []V {
	return slices.Clone(ms.values)
}

type MapTinyFixedArray[K comparable, V comparable, KA ArrayTinyParam[K], KV ArrayTinyParam[V]] struct {
	keys   KA
	values KV
	size   uint8
}

type MapTinyFixedArray8[K comparable, V comparable] struct {
	MapTinyFixedArray[K, V, [8]K, [8]V]
}

type MapTinyFixedArray16[K comparable, V comparable] struct {
	MapTinyFixedArray[K, V, [16]K, [16]V]
}

func (ma *MapTinyFixedArray[K, V, KA, VA]) Clear() {
	var nilKey K
	var nilValue V
	for i := range ma.size {
		ma.keys[i] = nilKey
		ma.values[i] = nilValue
	}
	ma.size = 0
}

func (ma *MapTinyFixedArray[K, V, KA, VA]) Size() int {
	return int(ma.size)
}

func (ma *MapTinyFixedArray[K, V, KA, VA]) IsEmpty() bool {
	return ma.size == 0
}

func (ma *MapTinyFixedArray[K, V, KA, VA]) SeqValues() iter.Seq[V] {
	return func(yield func(V) bool) {
		for i := range ma.size {
			if !yield(ma.values[i]) {
				return
			}
		}
	}
}

func (ma *MapTinyFixedArray[K, V, KA, VA]) EqualsInterface(other IMap[K, V], elementEqual func(*V, *V) bool) bool {
	if otherSmall, isSmall := other.(*MapTinyFixedArray[K, V, KA, VA]); isSmall {
		return ma.Equals(otherSmall)
	} else {
		return IMapEquals(ma, other, elementEqual)
	}
}

func (ma *MapTinyFixedArray[K, V, KA, VA]) Equals(other *MapTinyFixedArray[K, V, KA, VA]) bool {
	if ma.size != other.size {
		return false
	}

outerLoop:
	for a := range ma.size {
		keyA := ma.keys[a]
		for b := range ma.size {
			keyB := other.keys[b]
			if keyA == keyB {
				if ma.values[a] == other.values[b] {
					continue outerLoop
				} else {
					return false
				}
			}
		}
		// key not found
		return false
	}

	return true
}

func (ma *MapTinyFixedArray[K, V, KA, VA]) Has(key K) bool {
	for i := range ma.size {
		if ma.keys[i] == key {
			return true
		}
	}
	return false
}

func (ma *MapTinyFixedArray[K, V, KA, VA]) FirstKey() K {
	if ma.size > 0 {
		return ma.keys[0]
	} else {
		panic("no keys")
	}
}

func (ma *MapTinyFixedArray[K, V, KA, VA]) Get(key K) (V, bool) {
	for i := range ma.size {
		if ma.keys[i] == key {
			return ma.values[i], true
		}
	}
	var nilValue V
	return nilValue, false
}

func (ma *MapTinyFixedArray[K, V, KA, VA]) GetOrNilValue(key K) V {
	for i := range ma.size {
		if ma.keys[i] == key {
			return ma.values[i]
		}
	}
	var nilValue V
	return nilValue
}

func (ma *MapTinyFixedArray[K, V, KA, VA]) GetOrDefault(key K, defaultValue V) V {
	for i := range ma.size {
		if ma.keys[i] == key {
			return ma.values[i]
		}
	}
	return defaultValue
}

func (ma *MapTinyFixedArray[K, V, KA, VA]) GetOrPanic(key K) V {
	for i := range ma.size {
		if ma.keys[i] == key {
			return ma.values[i]
		}
	}
	panic("key missing")
}

func (ma *MapTinyFixedArray[K, V, KA, VA]) Put(key K, value V) {
	for i := range ma.size {
		if ma.keys[i] == key {
			ma.values[i] = value
			return
		}
	}

	if len(ma.keys) == int(ma.size) {
		panic("map full")
	}

	ma.keys[ma.size] = key
	ma.values[ma.size] = value
	ma.size++
}

func (ma *MapTinyFixedArray[K, V, KA, VA]) Compute(key K, apply func(V) V) {
	for i := range ma.size {
		if ma.keys[i] == key {
			ma.values[i] = apply(ma.values[i])
			return
		}
	}

	if len(ma.keys) == int(ma.size) {
		panic("map full")
	}

	var nilValue V
	ma.keys[ma.size] = key
	ma.values[ma.size] = apply(nilValue)
	ma.size++
}

func (ma *MapTinyFixedArray[K, V, KA, VA]) Delete(key K) {
	for i := range ma.size {
		if ma.keys[i] == key {
			removeIndexAndShiftTinyArray[K, KA](&ma.keys, int(i))
			removeIndexAndShiftTinyArray[V, VA](&ma.values, int(i))
			ma.size--
			return
		}
	}
}

func removeIndexAndShiftTinyArray[T any, A ArrayTinyParam[T]](array *A, index int) {
	if index < 0 || index >= len(*array) {
		panic("invalid index")
	}

	var nilValue T
	for i := index; i < len(*array)-1; i++ {
		(*array)[i] = (*array)[i+1]
	}
	(*array)[len(*array)-1] = nilValue
}

func (ma *MapTinyFixedArray[K, V, KA, VA]) Foreach(apply func(key K, value V)) {
	for i := range ma.size {
		apply(ma.keys[i], ma.values[i])
	}
}

func (ma *MapTinyFixedArray[K, V, KA, VA]) SeqKeyValue() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for i := range ma.size {
			if !yield(ma.keys[i], ma.values[i]) {
				return
			}
		}
	}
}

func (ma *MapTinyFixedArray[K, V, KA, VA]) SeqKey() iter.Seq[K] {
	return func(yield func(K) bool) {
		for i := range ma.size {
			if !yield(ma.keys[i]) {
				return
			}
		}
	}
}

func (ma *MapTinyFixedArray[K, V, KA, VA]) KeySlice() []K {
	slice := make([]K, ma.size)
	for i := range ma.size {
		slice[i] = ma.keys[i]
	}
	return slice
}

func (ma *MapTinyFixedArray[K, V, KA, VA]) ValueSlice() []V {
	slice := make([]V, ma.size)
	for i := range ma.size {
		slice[i] = ma.values[i]
	}
	return slice
}
