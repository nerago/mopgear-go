package util_collection

import (
	"iter"
	"math/rand"
	"slices"
)

type List[E any] struct {
	inner []E
}

var _ IList[string] = &List[string]{nil}

func (lst *List[E]) Clear() {
	lst.inner = lst.inner[:0]
}

func (lst *List[E]) Size() int {
	return len(lst.inner)
}

func (lst *List[E]) IsEmpty() bool {
	return len(lst.inner) == 0
}

func (lst *List[E]) Get(index int) E {
	return lst.inner[index]
}

func (lst *List[E]) GetLast() (E, bool) {
	if len(lst.inner) > 0 {
		return lst.inner[len(lst.inner)-1], true
	} else {
		var nilValue E
		return nilValue, false
	}
}

func (lst *List[E]) GetFirst() (E, bool) {
	if len(lst.inner) > 0 {
		return lst.inner[0], true
	} else {
		var nilValue E
		return nilValue, false
	}
}

func (lst *List[E]) ContainsFunc(predicate func(*E) bool) bool {
	for i := range lst.inner {
		if predicate(&lst.inner[i]) {
			return true
		}
	}
	return false
}

func (lst *List[E]) ContainsFuncNoPointer(predicate func(E) bool) bool {
	for i := range lst.inner {
		if predicate(lst.inner[i]) {
			return true
		}
	}
	return false
}

func (lst *List[E]) SeqIndexAndValues() iter.Seq2[int, E] {
	return func(yield func(int, E) bool) {
		for i := range lst.inner {
			if !yield(i, lst.inner[i]) {
				return
			}
		}
	}
}

func (lst *List[E]) SeqValues() iter.Seq[E] {
	return func(yield func(E) bool) {
		for i := range lst.inner {
			if !yield(lst.inner[i]) {
				return
			}
		}
	}
}

func (lst *List[E]) SeqValuePointers() iter.Seq[*E] {
	return func(yield func(*E) bool) {
		for i := range lst.inner {
			if !yield(&lst.inner[i]) {
				return
			}
		}
	}
}

func (lst *List[E]) SortFunc(compare func(*E, *E) int) {
	slices.SortFunc(lst.inner, func(a, b E) int {
		return compare(&a, &b)
	})
}

func (lst *List[E]) Shuffle() {
	rand.Shuffle(len(lst.inner), func(a, b int) {
		lst.inner[a], lst.inner[b] = lst.inner[b], lst.inner[a]
	})
}

func (lst *List[E]) Swap(a, b int) {
	lst.inner[a], lst.inner[b] = lst.inner[b], lst.inner[a]
}

func (lst *List[E]) Put(index int, value E) {
	lst.inner[index] = value
}

func (lst *List[E]) AppendLast(value E) {
	lst.inner = append(lst.inner, value)
}

func (lst *List[E]) InsertFirst(value E) {
	n := len(lst.inner)
	if n == 0 {
		lst.AppendLast(value)
		return
	}
	lst.inner = append(lst.inner, lst.inner[n-1])
	copy(lst.inner[1:n], lst.inner[0:n-1])
	lst.inner[0] = value
}

func (lst *List[E]) RemoveFirstAndReturn() (E, bool) {
	if len(lst.inner) == 0 {
		var nilValue E
		return nilValue, false
	}
	value := lst.inner[0]
	lst.inner = lst.inner[1:]
	return value, true
}

func (lst *List[E]) RemoveLastAndReturn() (E, bool) {
	if len(lst.inner) == 0 {
		var nilValue E
		return nilValue, false
	}
	value := lst.inner[len(lst.inner)-1]
	lst.inner = lst.inner[0 : len(lst.inner)-1]
	return value, true
}

func (lst *List[E]) DeleteIndex(index int) {
	DeleteIndexInPlace(&lst.inner, index)
}

func (lst *List[E]) RemoveDuplicatesFunc(equals func(a *E, b *E) bool) {
	RemoveDuplicatesFunc_InPlace(&lst.inner, equals)
}

func (lst *List[E]) FilterFunc(predicate func(*E) bool) {
	FilterSliceInPlace(&lst.inner, predicate)
}

func (lst *List[E]) FilterFuncNoPointer(predicate func(E) bool) {
	FilterSliceInPlace_NoPointer(&lst.inner, predicate)
}

type listTraversalEntry[E any] struct {
	value E
	index int
	lst   *List[E]
}

func (te listTraversalEntry[E]) Value() E {
	return te.value
}

func (te listTraversalEntry[E]) Index() int {
	return te.index
}

func (te listTraversalEntry[E]) Delete() {
	te.lst.DeleteIndex(te.index)
}

func (lst *List[E]) SeqListTraversalEntry() iter.Seq[ListTraversalEntry[E]] {
	return func(yield func(ListTraversalEntry[E]) bool) {
		for i := range lst.inner {
			if !yield(listTraversalEntry[E]{lst.inner[i], i, lst}) {
				return
			}
		}
	}
}

func (lst *List[E]) SubList(firstIndex, lastIndex int) IListRead[E] {
	return &List[E]{lst.inner[firstIndex : lastIndex+1]}
}
