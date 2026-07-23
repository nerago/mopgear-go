package simrank

import (
	"cmp"
	"iter"
	"paladin_gearing_go/util"
	"slices"
)

// guarantees that each ranking number is used in range, even given duplicate numbers
// consider obsolete mostly, use sort/ranking methods on main slices
func calculateRankingRanges[T any](highGood bool, inputData []T, toScore func(T) float64) iter.Seq2[T, util.HiLoInt] {
	if len(inputData) == 0 {
		return func(yield func(T, util.HiLoInt) bool) {}
	}

	type internalEntry struct {
		score   float64
		pointer T
		hilo    *util.HiLoInt
	}

	rankArray := make([]internalEntry, len(inputData))
	for i := range len(inputData) {
		pointer := inputData[i]
		rankArray[i] = internalEntry{
			toScore(pointer),
			pointer,
			nil,
		}
	}

	if highGood {
		slices.SortFunc(rankArray, func(a, b internalEntry) int { return cmp.Compare(a.score, b.score) })
	} else {
		slices.SortFunc(rankArray, func(a, b internalEntry) int { return cmp.Compare(b.score, a.score) })
	}

	prevScore := rankArray[0].score
	prevHiLo := &util.HiLoInt{Lo: 0, Hi: 0}
	for index := range rankArray {
		entry := &rankArray[index]
		if util.FloatsApproxEquals(entry.score, prevScore) {
			prevHiLo.Hi = index
			entry.hilo = prevHiLo
		} else {
			prevHiLo = &util.HiLoInt{Lo: index, Hi: index}
			entry.hilo = prevHiLo
		}
		prevScore = entry.score
	}

	return func(yield func(T, util.HiLoInt) bool) {
		for i := range rankArray {
			entry := &rankArray[i]
			hiLo := entry.hilo
			if !yield(entry.pointer, *hiLo) {
				return
			}
		}
	}
}
