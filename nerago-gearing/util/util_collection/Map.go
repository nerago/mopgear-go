package util_collection

import (
	"iter"
)

type Map[K comparable, V any] struct {
	data map[K]V
}

var _ IMap[int, int] = &Map[int, int]{}

// optional
func (mc *Map[K, V]) Init(size int) {
	if mc.data == nil {
		mc.data = make(map[K]V, size)
	}
}

func (mc *Map[K, V]) Get(key K) (V, bool) {
	value, hasValue := mc.data[key]
	return value, hasValue
}

func (mc *Map[K, V]) GetOrNilValue(key K) V {
	value := mc.data[key]
	return value
}

func (mc *Map[E, V]) GetOrDefault(key E, defaultValue V) V {
	value, hasValue := mc.data[key]
	if hasValue {
		return value
	} else {
		return defaultValue
	}
}

func (mc *Map[K, V]) GetOrPanic(key K) V {
	value, hasValue := mc.data[key]
	if hasValue {
		return value
	} else {
		panic("value not found")
	}
}

func (mc *Map[K, V]) Has(key K) bool {
	_, hasValue := mc.data[key]
	return hasValue
}

func (mc *Map[K, V]) Clear() {
	clear(mc.data)
}

func (mc *Map[K, V]) Size() int {
	return len(mc.data)
}

func (mc *Map[K, V]) IsEmpty() bool {
	return len(mc.data) == 0
}

func (mc *Map[K, V]) EqualsInterface(other IMap[K, V], elementEqual func(*V, *V) bool) bool {
	return IMapEquals(mc, other, elementEqual)
}

func (mc *Map[K, V]) Put(key K, value V) {
	data := mc.data
	if data == nil {
		data = make(map[K]V)
		mc.data = data
	}
	data[key] = value
}

func (mc *Map[K, V]) Compute(key K, apply func(V) V) {
	data := mc.data
	if data == nil {
		data = make(map[K]V)
		mc.data = data
	}

	value, hasValue := data[key]
	if hasValue {
		value = apply(value)
	} else {
		var nilValue V
		value = apply(nilValue)
	}
	data[key] = value
}

func (mc *Map[K, V]) Delete(key K) {
	delete(mc.data, key)
}

func (mc *Map[K, V]) FirstKey() K {
	for x := range mc.data {
		return x
	}
	panic("empty map")
}

func (mc *Map[K, V]) KeySlice() []K {
	slice := make([]K, 0, len(mc.data))
	for k := range mc.data {
		slice = append(slice, k)
	}
	return slice
}

func (mc *Map[K, V]) ValueSlice() []V {
	slice := make([]V, len(mc.data))
	for _, v := range mc.data {
		slice = append(slice, v)
	}
	return slice
}

func (mc *Map[K, V]) Foreach(apply func(key K, value V)) {
	for k, v := range mc.data {
		apply(k, v)
	}
}

func (mc *Map[K, V]) SeqKeyValue() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range mc.data {
			if !yield(k, v) {
				return
			}
		}
	}
}

func (mc *Map[K, V]) SeqValues() iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range mc.data {
			if !yield(v) {
				return
			}
		}
	}
}

func (mc *Map[K, V]) SeqKey() iter.Seq[K] {
	return func(yield func(K) bool) {
		for k := range mc.data {
			if !yield(k) {
				return
			}
		}
	}
}
