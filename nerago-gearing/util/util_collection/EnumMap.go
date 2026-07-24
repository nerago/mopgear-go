package util_collection

import (
	"iter"
	"slices"
)

type EnumType[E ~uint8] struct {
	numValues uint8
}

func (et EnumType[E]) IsUninitialized() bool {
	return et.numValues == 0
}

func EnumTypeMake[E ~uint8](values []E) EnumType[E] {
	for i := 0; i < len(values); i++ {
		if values[i] != E(i) {
			panic("expected iota enum in order")
		}
	}
	return EnumType[E]{uint8(len(values))}
}

func (et EnumType[E]) NumValues() uint32 {
	return uint32(et.numValues)
}

func (et EnumType[E]) First() E {
	return E(0)
}

func (et EnumType[E]) Last() E {
	return E(et.numValues - 1)
}

type EnumMap[E ~uint8, V any] struct {
	content []V
	isSet   []bool
	len     int
}

func EnumMapMake[E ~uint8, V any](enumType EnumType[E]) EnumMap[E, V] {
	if enumType.IsUninitialized() {
		panic("type not initialized")
	}
	return EnumMap[E, V]{
		make([]V, enumType.NumValues()),
		make([]bool, enumType.NumValues()),
		0,
	}
}

func (em *EnumMap[E, V]) IsUninitialized() bool {
	return em.content == nil || em.isSet == nil /*|| em.enumType.Values == nil*/
}

func (em *EnumMap[E, V]) Clone() *EnumMap[E, V] {
	return &EnumMap[E, V]{
		slices.Clone(em.content),
		slices.Clone(em.isSet),
		em.len,
	}
}

func (em *EnumMap[E, V]) Equals(other *EnumMap[E, V], elementEqual func(*V, *V) bool) bool {
	if em.len != other.len {
		return false
	}

	for i := range em.isSet {
		if em.isSet[i] != other.isSet[i] || !elementEqual(&em.content[i], &other.content[i]) {
			return false
		}
	}
	return true
}

func (em *EnumMap[E, V]) Len() int {
	return em.len
}

func (em *EnumMap[E, V]) Has(key E) bool {
	return em.isSet[key]
}

func (em *EnumMap[E, V]) Get(key E) (V, bool) {
	return em.content[key], em.isSet[key]
}

func (em *EnumMap[E, V]) GetOrPanic(key E) V {
	if em.isSet[key] {
		return em.content[key]
	} else {
		panic("key not set")
	}
}

func (em *EnumMap[E, V]) Put(key E, value V) {
	if !em.isSet[key] {
		em.isSet[key] = true
		em.len++
	}
	em.content[key] = value
}

func (em *EnumMap[E, V]) Delete(key E) {
	if em.isSet[key] {
		var nilValue V
		em.isSet[key] = false
		em.content[key] = nilValue
		em.len--
	}
}

func (em *EnumMap[E, V]) SeqKeyValue() iter.Seq2[E, V] {
	return func(yield func(E, V) bool) {
		for i, isSet := range em.isSet {
			if isSet {
				if !yield(E(i), em.content[i]) {
					return
				}
			}
		}
	}
}

func (em *EnumMap[E, V]) SeqValues() iter.Seq[V] {
	return func(yield func(V) bool) {
		for i, isSet := range em.isSet {
			if isSet {
				if !yield(em.content[i]) {
					return
				}
			}
		}
	}
}

func (em *EnumMap[E, V]) SeqKey() iter.Seq[E] {
	return func(yield func(E) bool) {
		for i, isSet := range em.isSet {
			if isSet {
				if !yield(E(i)) {
					return
				}
			}
		}
	}
}

func (em *EnumMap[E, V]) KeySlice() []E {
	slice := make([]E, em.len)
	write := 0
	for i, isSet := range em.isSet {
		if isSet {
			slice[write] = E(i)
			write++
		}
	}
	return slice
}
