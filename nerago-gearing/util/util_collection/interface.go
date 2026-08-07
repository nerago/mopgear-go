package util_collection

import (
	"iter"
)

type IEquatable[T IEquatable[T]] interface {
	Equals(other *T) bool
}

type IEquatableByElement[T IEquatableByElement[T, E], E any] interface {
	Equals(other *T, valueEquals func(*E, *E) bool) bool
}

type ICollection[E any] interface {
	Clear()
	Size() int
	IsEmpty() bool
}

type ISet[E any] interface {
	ICollection[E]
	Has(value E) bool
	Add(value E) (wasMember bool)
	Delete(value E)
	SeqValues() iter.Seq[E]
}

type IQueue[E any] interface {
	ICollection[E]
	Push(E)
	Pop() (E, bool)
}

type IListRead[E any] interface {
	ICollection[E]
	Get(index int) E
	ContainsFunc(predicate func(*E) bool) bool
	ContainsFuncNoPointer(predicate func(E) bool) bool
	SeqIndexAndValues() iter.Seq2[int, E]
	SeqValues() iter.Seq[E]
}

type IListSort[E any] interface {
	ICollection[E]
	SortFunc(compare func(*E, *E) int)
	Shuffle()
	Swap(indexA, indexB int)
}

type IListReadWrite[E any] interface {
	ICollection[E]
	IListSort[E]
	Put(index int, value E)
	Append(E)
	DeleteIndex(index int)
	RemoveDuplicatesFunc(equals func(a, b *E) bool)
}

type IList[E any] interface {
	IListRead[E]
	IListSort[E]
	IListReadWrite[E]
	SubList(firstIndex, lastIndex int) IListRead[E]
}

type IMap[K comparable, V any] interface {
	ICollection[V]
	EqualsInterface(other IMap[K, V], elementEqual func(*V, *V) bool) bool
	Has(key K) bool
	FirstKey() K
	Get(key K) (V, bool)
	GetOrPanic(key K) V
	Put(key K, value V)
	Delete(key K)
	Foreach(apply func(key K, value V))
	SeqKeyValue() iter.Seq2[K, V]
	SeqValues() iter.Seq[V]
	SeqKey() iter.Seq[K]
	KeySlice() []K
	ValueSlice() []V
}

type IMapMapCommon[J comparable, K comparable, V any] interface {
	ICollection[V]
	Has(key1 J, key2 K) bool
	HasKey1(key1 J) bool
	HasKey2(key2 K) bool
	FirstKey1() J
	FirstKey2() K
	DeleteAllForKey1(key1 J)
	DeleteAllForKey2(key2 K)
	Foreach(apply func(key1 J, key2 K, value V))
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

var _ IMap[int, int] = &MapConcurrent[int, int]{}
var _ IMapMap[int, int, int] = &MapMap[int, int, int]{}
var _ IMapMapSlice[int, int, int] = &MapMapSlice[int, int, int]{}

func IMapEquals[K comparable, V any](a IMap[K, V], b IMap[K, V], elementEqual func(*V, *V) bool) bool {
	if concurA, isConcurA := a.(*MapConcurrent[K, V]); isConcurA {
		concurA.mutex.RLock()
		defer concurA.mutex.RUnlock()
	}
	if concurB, isConcurB := b.(*MapConcurrent[K, V]); isConcurB {
		concurB.mutex.RLock()
		defer concurB.mutex.RUnlock()
	}

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

func IdentityFunc[T any](value T) T {
	return value
}
