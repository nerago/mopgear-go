package items

import (
	. "github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
)

func (set *SolvableItemSet) DebugValidate() error {
	equipMap := set.items
	previousTotal := set.total

	recalcTotal := StatBlock{}
	for item := range equipMap.AllItemSeq() {
		StatBlock_Increment_Mutating(&recalcTotal, &item.total)
	}

	if !StatBlock_Equals(&recalcTotal, &previousTotal) {
		return util.ErrorTracedNew("totals don't match\n" + recalcTotal.CreateString() + "\n" + previousTotal.CreateString())
	}

	if previousTotal.IsEmpty() {
		return util.ErrorTracedNew("set has no stats")
	}

	return nil
}

func (itemSet *FullItemSet) DebugValidate() error {
	equipMap := itemSet.items
	previousTotal := itemSet.total

	recalcTotal := StatBlock{}
	for item := range equipMap.AllItemSeq() {
		StatBlock_Increment_Mutating(&recalcTotal, item.Total())
	}

	if !StatBlock_Equals(&recalcTotal, &previousTotal) {
		return util.ErrorTracedNew("totals don't match\n" + recalcTotal.CreateString() + "\n" + previousTotal.CreateString())
	}

	return nil
}
