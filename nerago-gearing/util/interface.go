package util

import "iter"

type IMapMap[J comparable, K comparable, V any] interface {
	Get(key1 J, key2 K) (V, bool)
	GetOrPanic(key1 J, key2 K) V
	Has(key1 J, key2 K) bool
	HasKey1(key1 J) bool
	HasKey2(key2 K) bool
	Clear()
	Size() int
	IsEmpty() bool
	Put(key1 J, key2 K, value V)
	Delete(key1 J, key2 K)
	DeleteAllForKey1(key1 J)
	DeleteAllForKey2(key2 K)
	Apply(key1 J, key2 K, apply func(oldValue V) V)
	FirstKey1() J
	FirstKey2() K
	SeqKey1() iter.Seq[J]
	SeqKey2() iter.Seq[K]
	SeqValues() iter.Seq[V]
	SeqWithKeys() iter.Seq[MapMapEntry[J, K, V]]
	SeqWithKeysOtherOrder() iter.Seq[MapMapEntry[J, K, V]]
	ForeachWithKeys(apply func(key1 J, key2 K, value V))
	SeqInnerWithKey1Value(key1 J) iter.Seq2[K, V]
	SeqInnerWithKey2Value(key2 K) iter.Seq2[J, V]
	SeqGroupsKey1Lookup() iter.Seq2[J, func(K) V]
	SeqGroupsKey2Lookup() iter.Seq2[K, func(J) V]
	SeqGroupsKey1NestedKeyValue() iter.Seq2[J, iter.Seq2[K, V]]
	SeqGroupsKey2NestedKeyValue() iter.Seq2[K, iter.Seq2[J, V]]
	SeqKey1Key2Nested() iter.Seq2[J, iter.Seq[K]]
	SeqKey2Key1Nested() iter.Seq2[K, iter.Seq[J]]
}
