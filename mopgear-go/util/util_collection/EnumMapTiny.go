package util_collection

import "iter"

type EnumMapTiny[E EnumBaseType, V any, A ArrayTinyParam[V]] struct {
	content A
	isSet   uint16
	len     uint8
}

func (em *EnumMapTiny[E, V, A]) Clone() EnumMapTiny[E, V, A] {
	return EnumMapTiny[E, V, A]{em.content, em.isSet, em.len}
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

func (em *EnumMapTiny[E, V, A]) Size() int {
	return int(em.len)
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

func (em *EnumMapTiny[E, V, A]) GetOrPanic(key E) V {
	if (em.isSet & (1 << key)) != 0 {
		return em.content[key]
	} else {
		panic("key not set")
	}
}

func (em *EnumMapTiny[E, V, A]) GetOrNilValue(key E) V {
	if (em.isSet & (1 << key)) != 0 {
		return em.content[key]
	} else {
		var nilValue V
		return nilValue
	}
}

func (em *EnumMapTiny[E, V, A]) GetOrDefault(key E, defaultValue V) V {
	if (em.isSet & (1 << key)) != 0 {
		return em.content[key]
	} else {
		return defaultValue
	}
}

func (em *EnumMapTiny[E, V, A]) GetOrUseFactory(key E, factory func() V) V {
	if (em.isSet & (1 << key)) == 0 {
		em.content[key] = factory()
		em.isSet |= 1 << key
		em.len++
	}
	return em.content[key]
}

func (em *EnumMapTiny[E, V, A]) Put(key E, value V) {
	if (em.isSet & (1 << key)) == 0 {
		em.isSet |= 1 << key
		em.len++
	}
	em.content[key] = value
}

func (em *EnumMapTiny[E, V, A]) Compute(key E, apply func(V) V) {
	if (em.isSet & (1 << key)) != 0 {
		em.content[key] = apply(em.content[key])
	} else {
		var nilValue V
		em.content[key] = apply(nilValue)
		em.len++
	}
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
	slice := make([]E, 0, em.len)
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
	slice := make([]V, 0, em.len)
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
