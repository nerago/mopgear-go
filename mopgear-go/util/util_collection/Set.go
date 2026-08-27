package util_collection

import (
	"iter"
	"maps"
	"slices"
)

type SetComparable[E comparable] struct {
	inner map[E]bool
}

func (sc *SetComparable[E]) Clear() {
	clear(sc.inner)
}

func (sc *SetComparable[E]) Size() int {
	return len(sc.inner)
}

func (sc *SetComparable[E]) IsEmpty() bool {
	return len(sc.inner) == 0
}

func (sc *SetComparable[E]) SeqValues() iter.Seq[E] {
	return maps.Keys(sc.inner)
}

func (sc *SetComparable[E]) ContainsFunc(predicate func(*E) bool) bool {
	for k := range sc.inner {
		if predicate(new(k)) {
			return true
		}
	}
	return false
}

func (sc *SetComparable[E]) ContainsFuncNoPointer(predicate func(E) bool) bool {
	for k := range sc.inner {
		if predicate(k) {
			return true
		}
	}
	return false
}

func (sc *SetComparable[E]) FilterFunc(predicate func(*E) bool) {
	for k := range sc.inner {
		if !predicate(new(k)) {
			delete(sc.inner, k)
		}
	}
}

func (sc *SetComparable[E]) FilterFuncNoPointer(predicate func(E) bool) {
	for k := range sc.inner {
		if !predicate(k) {
			delete(sc.inner, k)
		}
	}
}

func (sc *SetComparable[E]) RemoveDuplicatesFunc(_ func(a *E, b *E) bool) {
	// nop for set
}

func (sc *SetComparable[E]) AddIfMissing(value E) (hadValue bool) {
	if sc.inner == nil {
		sc.inner = make(map[E]bool)
	}

	_, hasKey := sc.inner[value]
	if !hasKey {
		sc.inner[value] = true
	}
	return hasKey
}

func (sc *SetComparable[E]) HasValue(value E) bool {
	_, hasKey := sc.inner[value]
	return hasKey
}

func (sc *SetComparable[E]) DeleteValue(value E) (hadValue bool) {
	_, hasKey := sc.inner[value]
	if hasKey {
		delete(sc.inner, value)
	}
	return hasKey
}

type SetGeneral[E any] struct {
	slice []E
	equal func(*E, *E) bool
}

func SetGeneralMake[E any](equal func(*E, *E) bool) SetGeneral[E] {
	return SetGeneral[E]{equal: equal}
}

func (sg *SetGeneral[E]) Clear() {
	if sg.slice != nil {
		sg.slice = sg.slice[:0]
	}
}

func (sg *SetGeneral[E]) Size() int {
	return len(sg.slice)
}

func (sg *SetGeneral[E]) IsEmpty() bool {
	return len(sg.slice) == 0
}

func (sg *SetGeneral[E]) SeqValues() iter.Seq[E] {
	return slices.Values(sg.slice)
}

func (sg *SetGeneral[E]) ContainsFunc(predicate func(*E) bool) bool {
	for i := range sg.slice {
		if predicate(&sg.slice[i]) {
			return true
		}
	}
	return false
}

func (sg *SetGeneral[E]) ContainsFuncNoPointer(predicate func(E) bool) bool {
	for i := range sg.slice {
		if predicate(sg.slice[i]) {
			return true
		}
	}
	return false
}

func (sg *SetGeneral[E]) FilterFunc(predicate func(*E) bool) {
	FilterSliceInPlace(&sg.slice, predicate)
}

func (sg *SetGeneral[E]) FilterFuncNoPointer(predicate func(E) bool) {
	FilterSliceInPlace_NoPointer(&sg.slice, predicate)
}

func (sg *SetGeneral[E]) RemoveDuplicatesFunc(_ func(a *E, b *E) bool) {
	// nop for set
}

func (sg *SetGeneral[E]) AddIfMissing(value E) (hadValue bool) {
	for i := range sg.slice {
		if sg.equal(&sg.slice[i], &value) {
			return true
		}
	}

	sg.slice = append(sg.slice, value)
	return false
}

func (sg *SetGeneral[E]) HasValue(value E) bool {
	for i := range sg.slice {
		if sg.equal(&sg.slice[i], &value) {
			return true
		}
	}
	return false
}

func (sg *SetGeneral[E]) DeleteValue(value E) (hadValue bool) {
	for i := range sg.slice {
		if sg.equal(&sg.slice[i], &value) {
			DeleteIndexInPlace(&sg.slice, i)
			return true
		}
	}
	return false
}
