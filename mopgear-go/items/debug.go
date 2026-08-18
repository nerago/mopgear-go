package items

import (
	. "github.com/nerago/mopgear-go/stats"
)

func (set *SolvableItemSet) DebugValidate() {
	equipMap := set.items
	previousTotal := set.total

	recalcTotal := StatBlock{}
	for item := range equipMap.AllItemSeq() {
		StatBlock_Increment_Mutating(&recalcTotal, &item.total)
	}

	if !StatBlock_Equals(&recalcTotal, &previousTotal) {
		panic("totals don't match\n" + recalcTotal.CreateString() + "\n" + previousTotal.CreateString())
	}

	if previousTotal.IsEmpty() {
		panic("set has no stats")
	}
}

func (itemSet *FullItemSet) DebugValidate() {
	equipMap := itemSet.items
	previousTotal := itemSet.total

	recalcTotal := StatBlock{}
	for item := range equipMap.AllItemSeq() {
		StatBlock_Increment_Mutating(&recalcTotal, item.Total())
	}

	if !StatBlock_Equals(&recalcTotal, &previousTotal) {
		panic("totals don't match\n" + recalcTotal.CreateString() + "\n" + previousTotal.CreateString())
	}
}
