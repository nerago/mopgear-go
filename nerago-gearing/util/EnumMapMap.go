package util

import (
	"iter"
)

type EnumMapMap[J ~uint8, K ~uint8, V any] struct {
	content   []V
	isSet     BitSet
	len       int
	enumType1 EnumType[J]
	enumType2 EnumType[K]
}

func EnumMapMapMake[J ~uint8, K ~uint8, V any](enumType1 EnumType[J], enumType2 EnumType[K]) EnumMapMap[J, K, V] {
	enumType1.Validate()
	enumType2.Validate()
	arraySize := len(enumType1.Values) * len(enumType2.Values)
	return EnumMapMap[J, K, V]{
		make([]V, arraySize),
		BitSetMake(arraySize - 1),
		0,
		enumType1,
		enumType2,
	}
}

func (em *EnumMapMap[J, K, V]) IsUninitialized() bool {
	return em.content == nil || em.isSet == nil || em.enumType1.Values == nil || em.enumType2.Values == nil
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
	index := len(em.enumType2.Values)*int(key1) + int(key2)
	return em.isSet.IsSet(index)
}

func (em *EnumMapMap[J, K, V]) Get(key1 J, key2 K) (V, bool) {
	index := len(em.enumType2.Values)*int(key1) + int(key2)

	if em.isSet.IsSet(index) {
		return em.content[index], true
	} else {
		var nilValue V
		return nilValue, false
	}
}

func (em *EnumMapMap[J, K, V]) GetOrPanic(key1 J, key2 K) V {
	index := len(em.enumType2.Values)*int(key1) + int(key2)

	if em.isSet.IsSet(index) {
		return em.content[index]
	} else {
		panic("key not found")
	}
}

func (em *EnumMapMap[J, K, V]) HasKey1(key1 J) bool {
	for key2 := range len(em.enumType2.Values) {
		index := len(em.enumType2.Values)*int(key1) + key2
		if em.isSet.IsSet(index) {
			return true
		}
	}
	return false
}

func (em *EnumMapMap[J, K, V]) HasKey2(key2 K) bool {
	for key1 := range len(em.enumType1.Values) {
		index := len(em.enumType2.Values)*key1 + int(key2)
		// TODO make a ranged isSet method
		if em.isSet.IsSet(index) {
			return true
		}
	}
	return false
}

func (em *EnumMapMap[J, K, V]) Clear() {
	var nilValue V
	for index := range em.isSet.SeqIsSet() {
		em.content[index] = nilValue
	}

	for i := range em.isSet {
		em.isSet[i] = 0
	}
}

func (em *EnumMapMap[J, K, V]) Put(key1 J, key2 K, value V) {
	index := len(em.enumType2.Values)*int(key1) + int(key2)
	em.isSet.Set(index)
	em.content[index] = value
}

func (em *EnumMapMap[J, K, V]) Delete(key1 J, key2 K) {
	var nilValue V
	index := len(em.enumType2.Values)*int(key1) + int(key2)
	em.isSet.Clear(index)
	em.content[index] = nilValue
}

func (em *EnumMapMap[J, K, V]) DeleteAllForKey1(key1 J) {
	var nilValue V
	for key2 := range len(em.enumType2.Values) {
		index := len(em.enumType2.Values)*int(key1) + key2
		em.isSet.Clear(index)
		em.content[index] = nilValue
	}
}

func (em *EnumMapMap[J, K, V]) DeleteAllForKey2(key2 K) {
	var nilValue V
	for key1 := range len(em.enumType1.Values) {
		index := len(em.enumType2.Values)*key1 + int(key2)
		em.isSet.Clear(index)
		em.content[index] = nilValue
	}
}

func (em *EnumMapMap[J, K, V]) Apply(key1 J, key2 K, apply func(oldValue V) V) {
	index := len(em.enumType2.Values)*int(key1) + int(key2)
	em.isSet.Set(index)
	em.content[index] = apply(em.content[index])
}

func (em *EnumMapMap[J, K, V]) FirstKey1() J {
	// TODO make a ranged isSet method
	for key1 := range len(em.enumType1.Values) {
		for key2 := range len(em.enumType2.Values) {
			index := len(em.enumType2.Values)*key1 + key2
			if em.isSet(index) {
				return J(key1)
			}
		}
	}
	panic("no key")
}

func (em *EnumMapMap[J, K, V]) FirstKey2() K {
	//TODO implement me
	panic("implement me")
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
