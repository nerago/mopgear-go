package util

import (
	"iter"
	"maps"
	"sync"
)

type MapConcurrent[K comparable, V any] struct {
	data  map[K]V
	mutex sync.RWMutex
}

// optional
func (mc *MapConcurrent[K, V]) Init(size int) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if mc.data == nil {
		mc.data = make(map[K]V, size)
	}
}

func (mc *MapConcurrent[K, V]) Get(key K) (V, bool) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	value, hasValue := mc.data[key]
	return value, hasValue
}

func (mc *MapConcurrent[K, V]) GetOrNil(key K) V {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	value := mc.data[key]
	return value
}

func (mc *MapConcurrent[K, V]) GetOrPanic(key K) V {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	value, hasValue := mc.data[key]
	if hasValue {
		return value
	} else {
		panic("value not found")
	}
}

func (mc *MapConcurrent[K, V]) Has(key K) bool {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	_, hasValue := mc.data[key]
	return hasValue
}

func (mc *MapConcurrent[K, V]) Clear() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	clear(mc.data)
}

func (mc *MapConcurrent[K, V]) Size() int {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	return len(mc.data)
}

func (mc *MapConcurrent[K, V]) IsEmpty() bool {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	return len(mc.data) == 0
}

func (mc *MapConcurrent[K, V]) Put(key K, value V) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	data := mc.data
	if data == nil {
		data = make(map[K]V)
		mc.data = data
	}
	data[key] = value
}

func (mc *MapConcurrent[K, V]) Delete(key K) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	delete(mc.data, key)
}

func (mc *MapConcurrent[K, V]) Apply(key K, apply func(oldValue V) V) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

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

func (mc *MapConcurrent[K, V]) FirstKey() K {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	for x := range mc.data {
		return x
	}
	panic("empty map")
}

func (mc *MapConcurrent[K, V]) KeysAsSlice() []K {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	slice := make([]K, len(mc.data))
	for k := range mc.data {
		slice = append(slice, k)
	}
	return slice
}

func (mc *MapConcurrent[K, V]) ValuesAsSlice() []V {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	slice := make([]V, len(mc.data))
	for _, v := range mc.data {
		slice = append(slice, v)
	}
	return slice
}

func (mc *MapConcurrent[K, V]) SeqWithKeys_StaleInefficient() iter.Seq2[K, V] {
	mc.mutex.RLock()
	clone := maps.Clone(mc.data)
	mc.mutex.RUnlock()
	
	return maps.All(clone)
}
