package util_collection

import "iter"

type QueueSlice[E any] struct {
	q []E
}

func (q QueueSlice[E]) Clear() {
	clear(q.q)
	q.q = q.q[:0]
}

func (q QueueSlice[E]) Size() int {
	return len(q.q)
}

func (q QueueSlice[E]) IsEmpty() bool {
	return len(q.q) == 0
}

func (q QueueSlice[E]) SeqValues() iter.Seq[E] {
	return func(yield func(E) bool) {
		for _, v := range q.q {
			if !yield(v) {
				return
			}
		}
	}
}

func (q QueueSlice[E]) Push(v E) {
	q.q = append(q.q, v)
}

func (q QueueSlice[E]) Pop() (E, bool) {
	var nilValue E
	if len(q.q) > 0 {
		v := q.q[len(q.q)-1]
		q.q[len(q.q)-1] = nilValue
		q.q = q.q[0 : len(q.q)-1]
		return v, true
	} else {
		return nilValue, false
	}
}

func (q QueueSlice[E]) UpdateContents(apply func([]E) []E) {
	q.q = apply(q.q)
}

func (q QueueSlice[E]) ExamineContents(apply func([]E)) {
	apply(q.q)
}
