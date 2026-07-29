package util_collection

import (
	"iter"
)

type ICollection interface {
	Clear()
	Size() int
	IsEmpty() bool
}

type IMap[K comparable, V any] interface {
	ICollection
	EqualsInterface(other IMap[K, V], elementEqual func(*V, *V) bool) bool
	Has(key K) bool
	Get(key K) (V, bool)
	GetOrPanic(key K) V
	Put(key K, value V)
	Delete(key K)
	SeqKeyValue() iter.Seq2[K, V]
	SeqValues() iter.Seq[V]
	SeqKey() iter.Seq[K]
	KeySlice() []K
}

type IMapMapCommon[J comparable, K comparable, V any] interface {
	ICollection
	Has(key1 J, key2 K) bool
	HasKey1(key1 J) bool
	HasKey2(key2 K) bool
	FirstKey1() J
	FirstKey2() K
	DeleteAllForKey1(key1 J)
	DeleteAllForKey2(key2 K)
	ForeachWithKeys(apply func(key1 J, key2 K, value V))
	SeqKey1() iter.Seq[J]
	SeqKey2() iter.Seq[K]
	SeqValues() iter.Seq[V]
	SeqKey1Key2() iter.Seq2[J, iter.Seq[K]]
	SeqKey2Key1() iter.Seq2[K, iter.Seq[J]]
	SeqValuesWithKey1(key1 J) iter.Seq[V]
	SeqValuesWithKey2(key2 K) iter.Seq[V]
	SeqKey1Key2ValueEntries() iter.Seq[MapMapEntry[J, K, V]]
	SeqKey2Key1ValueEntries() iter.Seq[MapMapEntry[J, K, V]]
}

type IMapMap[J comparable, K comparable, V any] interface {
	IMapMapCommon[J, K, V]
	Get(key1 J, key2 K) (V, bool)
	GetOrPanic(key1 J, key2 K) V
	Put(key1 J, key2 K, value V)
	Delete(key1 J, key2 K)
	Apply(key1 J, key2 K, apply func(oldValue V) V)
	SeqKey2ValueWithKey1(key1 J) iter.Seq2[K, V]
	SeqKey1ValueWithKey2(key2 K) iter.Seq2[J, V]
	SeqKey1NestedKey2Value() iter.Seq2[J, iter.Seq2[K, V]]
	SeqKey2NestedKey1Value() iter.Seq2[K, iter.Seq2[J, V]]
}

type IMapMapSlice[J comparable, K comparable, V any] interface {
	IMapMapCommon[J, K, V]
	GetAsSliceInternal(key1 J, key2 K) ([]V, bool)
	GetAsSliceClone(key1 J, key2 K) ([]V, bool)
	GetAsSeq(key1 J, key2 K) iter.Seq[V]
	Add(key1 J, key2 K, value V)
	DeleteAllForKey1Key2(key1 J, key2 K)
	MapInternalSliceOrPanic(key1 J, key2 K, mapper func([]V) []V)
	MapInternalSliceIfExists(key1 J, key2 K, mapper func([]V) []V) bool
	SeqKey1Key2ValueSeqEntries() iter.Seq[MapMapSliceEntry[J, K, V]]
	SeqKey2Key1ValueSeqEntries() iter.Seq[MapMapSliceEntry[J, K, V]]
}

var mm IMapMap[int, int, int] = &MapMap[int, int, int]{}
var mms IMapMapSlice[int, int, int] = &MapMapSlice[int, int, int]{}
var em IMap[Sample, int] = &EnumMap[Sample, int]{}

func IMapEquals[K comparable, V any](a IMap[K, V], b IMap[K, V], elementEqual func(*V, *V) bool) bool {
	if a.Size() != b.Size() {
		return false
	} else if a.Size() == 0 && b.Size() == 0 {
		return true
	}

	for k, v := range a.SeqKeyValue() {
		v2, hasV2 := b.Get(k)
		if !hasV2 || !elementEqual(&v, &v2) {
			return false
		}
	}

	return true
}
