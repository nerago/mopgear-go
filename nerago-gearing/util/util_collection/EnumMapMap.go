package util_collection

import (
	"iter"
	"slices"
)

type EnumMapMap[J ~uint8, K ~uint8, V any] struct {
	content   []V
	isSet     BitSet
	len       int
	enumType1 EnumType[J]
	enumType2 EnumType[K]
}

func EnumMapMapMake[J ~uint8, K ~uint8, V any](enumType1 EnumType[J], enumType2 EnumType[K]) EnumMapMap[J, K, V] {
	if enumType1.IsUninitialized() || enumType2.IsUninitialized() {
		panic("type not initialized")
	}

	arraySize := enumType1.NumValues() * enumType2.NumValues()
	return EnumMapMap[J, K, V]{
		make([]V, arraySize),
		BitSetMake(arraySize - 1),
		0,
		enumType1,
		enumType2,
	}
}

func (em *EnumMapMap[J, K, V]) keyToIndex(key1 J, key2 K) uint32 {
	return em.enumType2.NumValues()*uint32(key1) + uint32(key2)
}

func (em *EnumMapMap[J, K, V]) indexToKeys(index uint32) (key1 J, key2 K) {
	key1 = J(index / em.enumType2.NumValues())
	key2 = K(index % em.enumType2.NumValues())
	return
}

func (em *EnumMapMap[J, K, V]) IsUninitialized() bool {
	return em.content == nil || em.isSet == nil || em.enumType1.IsUninitialized() || em.enumType2.IsUninitialized()
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
	index := em.keyToIndex(em.enumType1.First(), key2)
	for range em.enumType1.NumValues() {
		if em.isSet.IsSet(index) {
			return true
		}
		index += em.enumType2.NumValues()
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
	for i := minIndex; i <= maxIndex; i++ {
		em.content[i] = nilValue
	}

	em.isSet.ClearInRange(minIndex, maxIndex)
}

func (em *EnumMapMap[J, K, V]) DeleteAllForKey2(key2 K) {
	var nilValue V
	index := em.keyToIndex(em.enumType1.First(), key2)
	for range em.enumType1.NumValues() {
		em.isSet.Clear(index)
		em.content[index] = nilValue
		index += em.enumType2.NumValues()
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

	//minIndex := em.keyIndex(em.enumType1.First(), em.enumType2.First())
	//maxIndex := em.keyIndex(em.enumType1.First(), em.enumType2.Last())
	//for key1 := range em.enumType1.NumValues() {
	//	if em.isSet.IsAnySetInRange(minIndex, maxIndex) {
	//		return J(key1)
	//	}
	//	minIndex += em.enumType2.NumValues()
	//	maxIndex += em.enumType2.NumValues()
	//}
	//panic("no key")
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
	//TODO implement me
	panic("implement me")
}

func (em *EnumMapMap[J, K, V]) SeqKey2() iter.Seq[K] {
	//TODO implement me
	panic("implement me")
}

func (em *EnumMapMap[J, K, V]) SeqValues() iter.Seq[V] {
	//TODO implement me
	panic("implement me")
}

func (em *EnumMapMap[J, K, V]) SeqWithKeys() iter.Seq[MapMapEntry[J, K, V]] {
	//TODO implement me
	panic("implement me")
}

func (em *EnumMapMap[J, K, V]) SeqWithKeysOtherOrder() iter.Seq[MapMapEntry[J, K, V]] {
	//TODO implement me
	panic("implement me")
}

func (em *EnumMapMap[J, K, V]) ForeachWithKeys(apply func(key1 J, key2 K, value V)) {
	//TODO implement me
	panic("implement me")
}

func (em *EnumMapMap[J, K, V]) SeqInnerWithKey1Value(key1 J) iter.Seq2[K, V] {
	//TODO implement me
	panic("implement me")
}

func (em *EnumMapMap[J, K, V]) SeqInnerWithKey2Value(key2 K) iter.Seq2[J, V] {
	//TODO implement me
	panic("implement me")
}

func (em *EnumMapMap[J, K, V]) SeqGroupsKey1Lookup() iter.Seq2[J, func(K) V] {
	//TODO implement me
	panic("implement me")
}

func (em *EnumMapMap[J, K, V]) SeqGroupsKey2Lookup() iter.Seq2[K, func(J) V] {
	//TODO implement me
	panic("implement me")
}

func (em *EnumMapMap[J, K, V]) SeqGroupsKey1NestedKeyValue() iter.Seq2[J, iter.Seq2[K, V]] {
	//TODO implement me
	panic("implement me")
}

func (em *EnumMapMap[J, K, V]) SeqGroupsKey2NestedKeyValue() iter.Seq2[K, iter.Seq2[J, V]] {
	//TODO implement me
	panic("implement me")
}

func (em *EnumMapMap[J, K, V]) SeqKey1Key2Nested() iter.Seq2[J, iter.Seq[K]] {
	//TODO implement me
	panic("implement me")
}

func (em *EnumMapMap[J, K, V]) SeqKey2Key1Nested() iter.Seq2[K, iter.Seq[J]] {
	//TODO implement me
	panic("implement me")
}
