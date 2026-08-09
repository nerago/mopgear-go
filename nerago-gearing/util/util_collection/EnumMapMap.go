package util_collection

import (
	"iter"
	"slices"
)

type EnumMapMap[J EnumBaseType, K EnumBaseType, V any] struct {
	content   []V
	isSet     BitSet
	len       int
	enumType1 EnumType[J]
	enumType2 EnumType[K]
}

func EnumMapMapMake[J EnumBaseType, K EnumBaseType, V any](enumType1 EnumType[J], enumType2 EnumType[K]) EnumMapMap[J, K, V] {
	arraySize := uint32(enumType1.NumValues()) * uint32(enumType2.NumValues())
	return EnumMapMap[J, K, V]{
		make([]V, arraySize),
		BitSetMake(arraySize - 1),
		0,
		enumType1,
		enumType2,
	}
}

func (em *EnumMapMap[J, K, V]) keyToIndex(key1 J, key2 K) uint32 {
	return uint32(em.enumType2.NumValues())*uint32(key1) + uint32(key2)
}

func (em *EnumMapMap[J, K, V]) indexToKeys(index uint32) (key1 J, key2 K) {
	blockSize := uint32(em.enumType2.NumValues())
	key1 = J(index / blockSize)
	key2 = K(index % blockSize)
	return key1, key2
}

func (em *EnumMapMap[J, K, V]) IsUninitialized() bool {
	return em.content == nil || em.isSet == nil
}

func (em *EnumMapMap[J, K, V]) Clone() *EnumMapMap[J, K, V] {
	return &EnumMapMap[J, K, V]{
		slices.Clone(em.content),
		slices.Clone(em.isSet),
		em.len,
		em.enumType1,
		em.enumType2,
	}
}

func (em *EnumMapMap[J, K, V]) Equals(other *EnumMapMap[J, K, V], elementEqual func(*V, *V) bool) bool {
	if em.len != other.len {
		return false
	}

	for i := range em.isSet {
		if em.isSet[i] != other.isSet[i] {
			return false
		}
	}

	for index := range em.isSet.SeqIsSet() {
		if !elementEqual(&em.content[index], &other.content[index]) {
			return false
		}
	}

	return true
}

func (em *EnumMapMap[J, K, V]) Size() int {
	return em.len
}

func (em *EnumMapMap[J, K, V]) IsEmpty() bool {
	return em.len == 0
}

func (em *EnumMapMap[J, K, V]) Has(key1 J, key2 K) bool {
	index := em.keyToIndex(key1, key2)
	return em.isSet.IsSet(index)
}

func (em *EnumMapMap[J, K, V]) Get(key1 J, key2 K) (V, bool) {
	index := em.keyToIndex(key1, key2)

	if em.isSet.IsSet(index) {
		return em.content[index], true
	} else {
		var nilValue V
		return nilValue, false
	}
}

func (em *EnumMapMap[J, K, V]) GetOrPanic(key1 J, key2 K) V {
	index := em.keyToIndex(key1, key2)

	if em.isSet.IsSet(index) {
		return em.content[index]
	} else {
		panic("key not found")
	}
}

func (em *EnumMapMap[J, K, V]) HasKey1(key1 J) bool {
	minIndex := em.keyToIndex(key1, em.enumType2.First())
	maxIndex := em.keyToIndex(key1, em.enumType2.Last())
	return em.isSet.IsAnySetInRange(minIndex, maxIndex)
}

func (em *EnumMapMap[J, K, V]) HasKey2(key2 K) bool {
	startIndex := em.keyToIndex(em.enumType1.First(), key2)
	for range em.isSet.SeqIsSetSkipScan(startIndex, uint32(em.enumType2.NumValues())) {
		return true
	}
	return false
}

func (em *EnumMapMap[J, K, V]) Clear() {
	var nilValue V
	for index := range em.isSet.SeqIsSet() {
		em.content[index] = nilValue
	}
	em.isSet.ClearAll()
}

func (em *EnumMapMap[J, K, V]) Put(key1 J, key2 K, value V) {
	index := em.keyToIndex(key1, key2)
	em.isSet.Set(index)
	em.content[index] = value
}

func (em *EnumMapMap[J, K, V]) Delete(key1 J, key2 K) {
	var nilValue V
	index := em.keyToIndex(key1, key2)
	em.isSet.Clear(index)
	em.content[index] = nilValue
}

func (em *EnumMapMap[J, K, V]) DeleteAllForKey1(key1 J) {
	var nilValue V

	minIndex := em.keyToIndex(key1, em.enumType2.First())
	maxIndex := em.keyToIndex(key1, em.enumType2.Last())
	for index := range em.isSet.SeqIsSetBetweenClearing(minIndex, maxIndex) {
		em.content[index] = nilValue
	}
}

func (em *EnumMapMap[J, K, V]) DeleteAllForKey2(key2 K) {
	var nilValue V

	startIndex := em.keyToIndex(em.enumType1.First(), key2)
	for index := range em.isSet.SeqIsSetSkipScanClearing(startIndex, uint32(em.enumType2.NumValues())) {
		em.content[index] = nilValue
	}
}

func (em *EnumMapMap[J, K, V]) Apply(key1 J, key2 K, apply func(oldValue V) V) {
	index := em.keyToIndex(key1, key2)
	em.isSet.Set(index)
	em.content[index] = apply(em.content[index])
}

func (em *EnumMapMap[J, K, V]) FirstKey1() J {
	firstIndex, hasIndex := em.isSet.FirstIndexIsSet()
	if hasIndex {
		key1, _ := em.indexToKeys(firstIndex)
		return key1
	} else {
		panic("no key")
	}
}

func (em *EnumMapMap[J, K, V]) FirstKey2() K {
	firstIndex, hasIndex := em.isSet.FirstIndexIsSet()
	if hasIndex {
		_, key2 := em.indexToKeys(firstIndex)
		return key2
	} else {
		panic("no key")
	}
}

func (em *EnumMapMap[J, K, V]) SeqKey1() iter.Seq[J] {
	return func(yield func(J) bool) {
		minIndex := em.keyToIndex(em.enumType1.First(), em.enumType2.First())
		maxIndex := em.keyToIndex(em.enumType1.First(), em.enumType2.Last())
		for key1 := range em.enumType1.NumValues() {
			if em.isSet.IsAnySetInRange(minIndex, maxIndex) {
				if !yield(J(key1)) {
					return
				}
			}
			minIndex += uint32(em.enumType2.NumValues())
			maxIndex += uint32(em.enumType2.NumValues())
		}
	}
}

func (em *EnumMapMap[J, K, V]) SeqKey2() iter.Seq[K] {
	return func(yield func(K) bool) {
		startIndex := em.keyToIndex(em.enumType1.First(), em.enumType2.First())
		for key2 := range em.enumType2.NumValues() {
			for range em.isSet.SeqIsSetSkipScan(startIndex, uint32(em.enumType2.NumValues())) {
				if !yield(K(key2)) {
					return
				}
				break
			}
			startIndex++
		}
	}
}

func (em *EnumMapMap[J, K, V]) SeqValues() iter.Seq[V] {
	return func(yield func(V) bool) {
		for index := range em.isSet.SeqIsSet() {
			value := em.content[index]
			if !yield(value) {
				return
			}
		}
	}
}

func (em *EnumMapMap[J, K, V]) SeqKey1Key2ValueEntries() iter.Seq[MapMapEntry[J, K, V]] {
	return func(yield func(MapMapEntry[J, K, V]) bool) {
		for index := range em.isSet.SeqIsSet() {
			value := em.content[index]
			key1, key2 := em.indexToKeys(index)
			if !yield(MapMapEntry[J, K, V]{Key1: key1, Key2: key2, Value: value}) {
				return
			}
		}
	}
}

func (em *EnumMapMap[J, K, V]) SeqKey2Key1ValueEntries() iter.Seq[MapMapEntry[J, K, V]] {
	return func(yield func(MapMapEntry[J, K, V]) bool) {
		startIndex := em.keyToIndex(em.enumType1.First(), em.enumType2.First())
		for key2 := range em.enumType2.NumValues() {
			for index := range em.isSet.SeqIsSetSkipScan(startIndex, uint32(em.enumType2.NumValues())) {
				key1, _ := em.indexToKeys(index)
				value := em.content[index]
				if !yield(MapMapEntry[J, K, V]{Key1: key1, Key2: K(key2), Value: value}) {
					return
				}
				break
			}
			startIndex++
		}
	}
}

func (em *EnumMapMap[J, K, V]) SeqKey1Key2() iter.Seq2[J, iter.Seq[K]] {
	return func(yield func(J, iter.Seq[K]) bool) {
		for key1 := range em.enumType1.NumValues() {
			key1J := J(key1)
			if !yield(key1J, func(yieldInner func(K) bool) {
				minIndex := em.keyToIndex(key1J, em.enumType2.First())
				maxIndex := em.keyToIndex(key1J, em.enumType2.Last())
				for index := range em.isSet.SeqIsSetBetween(minIndex, maxIndex) {
					_, key2 := em.indexToKeys(index)
					if !yieldInner(key2) {
						return
					}
				}
			}) {
				return
			}
		}
	}
}

func (em *EnumMapMap[J, K, V]) SeqKey2Key1() iter.Seq2[K, iter.Seq[J]] {
	return func(yield func(K, iter.Seq[J]) bool) {
		for key2 := range em.enumType2.NumValues() {
			if !yield(K(key2), func(yieldInner func(J) bool) {
				startIndex := em.keyToIndex(em.enumType1.First(), K(key2))
				for index := range em.isSet.SeqIsSetSkipScan(startIndex, uint32(em.enumType2.NumValues())) {
					key1, _ := em.indexToKeys(index)
					if !yieldInner(key1) {
						return
					}
				}
			}) {
				return
			}
		}
	}
}

func (em *EnumMapMap[J, K, V]) Foreach(apply func(key1 J, key2 K, value V)) {
	for index := range em.isSet.SeqIsSet() {
		value := em.content[index]
		key1, key2 := em.indexToKeys(index)
		apply(key1, key2, value)
	}
}

func (em *EnumMapMap[J, K, V]) SeqKey2ValueWithKey1(key1 J) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		minIndex := em.keyToIndex(key1, em.enumType2.First())
		maxIndex := em.keyToIndex(key1, em.enumType2.Last())
		for index := range em.isSet.SeqIsSetBetween(minIndex, maxIndex) {
			_, key2 := em.indexToKeys(index)
			value := em.content[index]
			if !yield(key2, value) {
				return
			}
		}
	}
}

func (em *EnumMapMap[J, K, V]) SeqKey1ValueWithKey2(key2 K) iter.Seq2[J, V] {
	return func(yield func(J, V) bool) {
		startIndex := em.keyToIndex(em.enumType1.First(), key2)
		for index := range em.isSet.SeqIsSetSkipScan(startIndex, uint32(em.enumType2.NumValues())) {
			key1, _ := em.indexToKeys(index)
			value := em.content[index]
			if !yield(key1, value) {
				return
			}
		}
	}
}

func (em *EnumMapMap[J, K, V]) SeqValuesWithKey1(key1 J) iter.Seq[V] {
	return func(yield func(V) bool) {
		minIndex := em.keyToIndex(key1, em.enumType2.First())
		maxIndex := em.keyToIndex(key1, em.enumType2.Last())
		for index := range em.isSet.SeqIsSetBetween(minIndex, maxIndex) {
			value := em.content[index]
			if !yield(value) {
				return
			}
		}
	}
}

func (em *EnumMapMap[J, K, V]) SeqValuesWithKey2(key2 K) iter.Seq[V] {
	return func(yield func(V) bool) {
		startIndex := em.keyToIndex(em.enumType1.First(), key2)
		for index := range em.isSet.SeqIsSetSkipScan(startIndex, uint32(em.enumType2.NumValues())) {
			value := em.content[index]
			if !yield(value) {
				return
			}
		}
	}
}

func (em *EnumMapMap[J, K, V]) SeqKey1NestedKey2Value() iter.Seq2[J, iter.Seq2[K, V]] {
	return func(yield func(J, iter.Seq2[K, V]) bool) {
		for key1 := range em.enumType1.NumValues() {
			j := J(key1)
			if !yield(j, em.SeqKey2ValueWithKey1(j)) {
				return
			}
		}
	}
}

func (em *EnumMapMap[J, K, V]) SeqKey2NestedKey1Value() iter.Seq2[K, iter.Seq2[J, V]] {
	return func(yield func(K, iter.Seq2[J, V]) bool) {
		for key2 := range em.enumType2.NumValues() {
			k := K(key2)
			if !yield(k, em.SeqKey1ValueWithKey2(k)) {
				return
			}
		}
	}
}
