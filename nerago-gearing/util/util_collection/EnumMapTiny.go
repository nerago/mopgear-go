package util_collection

import "iter"

type sample uint8

func (s sample) Name() string {
	switch s {
	case Sample_One:
		return "one"
	case Sample_Two:
		return "two"
	case Sample_Three:
		return "three"
	default:
		return ""
	}
}

func (s sample) EnumNumValues() uint8 {
	return uint8(len(sampleList))
}

const (
	Sample_One      sample = iota
	Sample_Two      sample = iota
	Sample_Three    sample = iota
	Sample_EnumSize        = 3
)

var sampleList = []sample{Sample_One, Sample_Two, Sample_Three}
var sampleEnum = EnumTypeMake(sampleList)

type enumMap2Sample[V any] struct {
	EnumMapTiny[sample, V, [Sample_EnumSize]V]
}

var _ IMap[sample, int] = &enumMap2Sample[int]{}

type EnumMapTiny[E EnumBaseType, V any, A ArrayTinyParam[V]] struct {
	content  A
	isSet    uint16
	len      uint8
	enumType EnumType[E]
}

func (em *EnumMapTiny[E, V, A]) Clear() {
	var nilValue V
	for i := range len(em.content) {
		em.content[i] = nilValue
	}
	em.isSet = 0
	em.len = 0
}

func (em *EnumMapTiny[E, V, A]) IsEmpty() bool {
	return em.isSet == 0
}

func (em *EnumMapTiny[E, V, A]) Has(key E) bool {
	return (em.isSet & (1 << key)) != 0
}

func (em *EnumMapTiny[E, V, A]) FirstKey() E {
	bits := em.isSet
	index := 0
	for bits != 0 {
		if (bits & 0x1) != 0 {
			return E(index)
		}
		index++
		bits >>= 1
	}
	panic("no key")
}

func (em *EnumMapTiny[E, V, A]) Get(key E) (V, bool) {
	if (em.isSet & (1 << key)) != 0 {
		return em.content[key], true
	} else {
		var nilValue V
		return nilValue, false
	}
}

func (em *EnumMapTiny[E, V, A]) Size() int {
	return int(em.len)
}

func (em *EnumMapTiny[E, V, A]) GetOrPanic(key E) V {
	if (em.isSet & (1 << key)) != 0 {
		return em.content[key]
	} else {
		panic("key not set")
	}
}

func (em *EnumMapTiny[E, V, A]) Put(key E, value V) {
	if (em.isSet & (1 << key)) == 0 {
		em.isSet |= 1 << key
		em.len++
	}
	em.content[key] = value
}

func (em *EnumMapTiny[E, V, A]) Delete(key E) {
	if (em.isSet & (1 << key)) != 0 {
		var nilValue V
		em.content[key] = nilValue
		em.isSet &= ^(1 << key)
		em.len--
	}
}

func (em *EnumMapTiny[E, V, A]) EqualsInterface(other IMap[E, V], elementEqual func(*V, *V) bool) bool {
	if otherEnum, isOtherEnum := other.(*EnumMapTiny[E, V, A]); isOtherEnum {
		return em.Equals(otherEnum, elementEqual)
	} else {
		return IMapEquals(em, other, elementEqual)
	}
}

func (em *EnumMapTiny[E, V, A]) Equals(other *EnumMapTiny[E, V, A], equal func(*V, *V) bool) bool {
	if em.isSet != other.isSet {
		return false
	}
	bits := em.isSet
	index := 0
	for bits != 0 {
		if (bits & 0x1) != 0 {
			if !equal(&em.content[index], &other.content[index]) {
				return false
			}
		}
		index++
		bits >>= 1
	}
	return true
}

func (em *EnumMapTiny[E, V, A]) Foreach(apply func(key E, value V)) {
	bits := em.isSet
	index := 0
	for bits != 0 {
		if (bits & 0x1) != 0 {
			apply(E(index), em.content[index])
		}
		index++
		bits >>= 1
	}
}

func (em *EnumMapTiny[E, V, A]) SeqKeyValue() iter.Seq2[E, V] {
	return func(yield func(E, V) bool) {
		bits := em.isSet
		index := 0
		for bits != 0 {
			if (bits & 0x1) != 0 {
				if !yield(E(index), em.content[index]) {
					return
				}
			}
			index++
			bits >>= 1
		}
	}
}

func (em *EnumMapTiny[E, V, A]) SeqValues() iter.Seq[V] {
	return func(yield func(V) bool) {
		bits := em.isSet
		index := 0
		for bits != 0 {
			if (bits & 0x1) != 0 {
				if !yield(em.content[index]) {
					return
				}
			}
			index++
			bits >>= 1
		}
	}
}

func (em *EnumMapTiny[E, V, A]) SeqKey() iter.Seq[E] {
	return func(yield func(E) bool) {
		bits := em.isSet
		index := 0
		for bits != 0 {
			if (bits & 0x1) != 0 {
				if !yield(E(index)) {
					return
				}
			}
			index++
			bits >>= 1
		}
	}
}

func (em *EnumMapTiny[E, V, A]) KeySlice() []E {
	slice := make([]E, em.len)
	bits := em.isSet
	index := 0
	for bits != 0 {
		if (bits & 0x1) != 0 {
			slice = append(slice, E(index))
		}
		index++
		bits >>= 1
	}
	return slice
}

func (em *EnumMapTiny[E, V, A]) ValueSlice() []V {
	slice := make([]V, em.len)
	bits := em.isSet
	index := 0
	for bits != 0 {
		if (bits & 0x1) != 0 {
			slice = append(slice, em.content[index])
		}
		index++
		bits >>= 1
	}
	return slice
}
