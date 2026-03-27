package phased

import (
	"math/big"
	"paladin_gearing_go/items"
	. "paladin_gearing_go/items"
	. "paladin_gearing_go/model"
	"paladin_gearing_go/solver/solve_util"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"paladin_gearing_go/util/util_rank"
)

const (
	threadCount  = 6 // per thread type
	bufferSize   = 256
	filterTarget = 10000
)

// var skinnyPool = sync.Pool{New: func() any { return new(SkinnyItemSet) }}
// var solvablePool = sync.Pool{New: func() any { return new(SolvableItemSet) }}

func SolverSkinnyPhasedIndex_Run(itemOptions *SolvableOptionsMap, model *Model, targetCount uint64, trackProgress *util.TrackProgress, printer *util.PrintRecorder) util.Optional[SolvableItemSet] {
	skinnyOptions := toSkinnyOptions(itemOptions, model)

	max := skinnyOptions.TotalCombinationCount()
	targetCombination := big.NewInt(int64(targetCount))
	skip := util.ChooseSkip_NextPrimeFromRatio(max, targetCombination)

	printer.Printf("SOLVE PHASED %d %d %d\n", max, targetCombination, skip)

	if max.IsUint64() && skip.IsUint64() {
		return makeSkinnyCombosMultiThread(&skinnyOptions, itemOptions, model, max.Uint64(), skip.Uint64(), trackProgress)
	} else {
		panic("too many combos for int")
	}
}

func toSkinnyOptions(itemOptions *SolvableOptionsMap, model *Model) SkinnyOptionsMap {
	skinnyOptions := SkinnyOptionsMap{}
	for slot, slotOptions := range itemOptions {
		uniqueMap := make(map[SkinnyItem]bool, len(slotOptions))

		for _, item := range slotOptions {
			skinny := model.StatRequirements.ToSkinny(&item)
			uniqueMap[skinny] = true
		}

		uniqueSet := toKeys(uniqueMap)
		skinnyOptions[slot] = uniqueSet
	}
	return skinnyOptions
}

func toKeys(uniqueMap map[SkinnyItem]bool) []SkinnyItem {
	keys := make([]SkinnyItem, 0, len(uniqueMap))
	for item := range uniqueMap {
		keys = append(keys, item)
	}
	return keys
}

func makeSkinnyCombosMultiThread(skinnyOptions *SkinnyOptionsMap, itemOptions *SolvableOptionsMap, model *Model, max, skip uint64, trackProgress *util.TrackProgress) util.Optional[SolvableItemSet] {
	counters := make([]uint64, threadCount)

	trackProgress.RunFromArray(&counters, max/skip)

	// start up workers
	splits := solve_util.IndexSplitsInt(max, skip, threadCount)
	resultChannel := channel_op.GenerateAll_ToChannel(threadCount, func(threadNum int, resultChannel chan<- util_rank.BestCollector1[SolvableItemSet]) {
		createWorkerRangeInt(skinnyOptions, itemOptions, model, splits[threadNum], splits[threadNum+1], skip, resultChannel, &counters[threadNum])
	})

	return util_rank.BestCollector1_OfChannel(resultChannel, threadCount)
}

func createWorkerRangeInt(skinnyOptions *SkinnyOptionsMap, itemOptions *SolvableOptionsMap, model *Model, start, max, skip uint64, resultChannel chan<- util_rank.BestCollector1[SolvableItemSet], progressCounter *uint64) {
	valueHeap := util_rank.LowestNIntHeap_For(filterTarget)
	best := util_rank.BestCollector1[SolvableItemSet]{}

	skinnySet := new(SkinnyItemSet)
	solveSet := new(SolvableItemSet)
	best.BestObject = new(SolvableItemSet) 

	index := start
	for index < max {
		makeSkinnySetInt(skinnyOptions, index, skinnySet)
		index += skip

		if model.CheckSetSkinny(skinnySet) {
			skinnyRating := uint64(skinnySet.A + skinnySet.B)
			if valueHeap.Offer(skinnyRating) {
				makeFromSkinny(itemOptions, model, skinnySet, solveSet)

				// assert still matches requirement, should be redundant
				if !model.CheckSet(solveSet) {
					panic("inconsistent cap calcuations")
				}

				solveRating := model.CalcRatingSolve(solveSet)
				best.OfferAndSwap(&solveSet, solveRating)
			}
		}

		*progressCounter++
	}

	resultChannel <- best
}

func makeSkinnySetInt(itemOptions *SkinnyOptionsMap, mainIndex uint64, set *SkinnyItemSet) {
	set.ClearTotals()

	currIndex := mainIndex

	for slot := range itemOptions {
		array := itemOptions[slot]
		size := uint64(len(array))

		if size > 0 {
			// TODO use overflow techniques instead?
			slotIndex := currIndex % size
			currIndex /= size

			item := array[slotIndex]
			set.Items[slot] = item
			set.A += item.A
			set.B += item.B
		}
	}
}

func makeFromSkinny(itemOptions *SolvableOptionsMap, model *Model, skinnySet *SkinnyItemSet, chosen *SolvableItemSet) {
	chosen.ClearTotals()

	for slot := Equip_Iter_First; slot <= Equip_Iter_Last; slot++ {
		skinny := &skinnySet.Items[slot]
		if skinny.Exists {
			options := itemOptions[slot]

			best := util_rank.BestCollector1[SolvableItem]{} // allow for zero rating
			for i := range len(options) {
				item := &options[i]
				if model.StatRequirements.SkinnyMatch(skinny, item) {
					rating := model.CalcRatingSolveItem(item)
					best.Offer(item, rating)
				}
			}

			chosen.AddItem_DeferCalc(slot, best.GetBestPointerOrPanic())
		}
	}

	items.SolvableItemSet_RecalculateTotal(chosen)
}
