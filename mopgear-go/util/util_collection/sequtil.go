package util_collection

import "iter"

func ConcatSeq2[T any](seq1 iter.Seq[T], seq2 iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq1 {
			if !yield(v) {
				return
			}
		}
		for v := range seq2 {
			if !yield(v) {
				return
			}
		}
	}
}

func ConcatSeqMany[T any](seqParam ...iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := range seqParam {
			for v := range seqParam[i] {
				if !yield(v) {
					return
				}
			}
		}
	}
}
