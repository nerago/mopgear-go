package util_collection

import (
	"iter"
)

type EnumSet[E EnumBaseType] struct {
	bits     BitSet
	len      uint8
	enumType EnumType[E]
}

func EnumSetMake[E EnumBaseType](enumType EnumType[E]) EnumSet[E] {
	return EnumSet[E]{
		BitSetMake(uint32(enumType.NumValues() - 1)),
		0,
		enumType,
	}
}

func (set *EnumSet[E]) Clear() {
	set.bits.ClearAll()
}

func (set *EnumSet[E]) Size() int {
	return int(set.len)
}

func (set *EnumSet[E]) IsEmpty() bool {
	return set.len == 0
}

func (set *EnumSet[E]) Has(value E) bool {
	return set.bits.IsSet(uint32(value))
}

func (set *EnumSet[E]) AddIfMissing(value E) (wasMember bool) {
	return set.bits.SetReturningOld(uint32(value))
}

func (set *EnumSet[E]) DeleteValue(value E) {
	set.bits.Clear(uint32(value))
}

func (set *EnumSet[E]) SeqValues() iter.Seq[E] {
	return func(yield func(E) bool) {
		for i := range set.bits.SeqIsSet() {
			if !yield(E(i)) {
				return
			}
		}
	}
}
