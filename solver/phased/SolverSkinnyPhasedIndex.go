package phased

import (
	"math/big"
	"paladin_gearing_go/model"
	. "paladin_gearing_go/model"
	"paladin_gearing_go/solver/solve_util"
	. "paladin_gearing_go/types/common"
	. "paladin_gearing_go/types/items"
	"paladin_gearing_go/util"
	"time"
)

const (
	threadCount  = 6 // per thread type
	bufferSize   = 256
	filterTarget = 10000
)

func SolverSkinnyPhasedIndex_Run(itemOptions *SolvableOptionsMap, model *Model, targetCount uint64, printer *util.PrintRecorder) SolvableItemSet {
	skinnyOptions := toSkinnyOptions(itemOptions, model)

	max := skinnyOptions.TotalCombinationCount()
	targetCombination := big.NewInt(int64(targetCount))
	skip := util.ChooseSkip(max, targetCombination)

	printer.Printf("SOLVE PHASED %d %d %d\n", max, targetCombination, skip)

	if max.IsUint64() && skip.IsUint64() {
		skinnyComboChannel := makeSkinnyCombosMultiThread(&skinnyOptions, model, max.Uint64(), skip.Uint64())
		skinnyComboChannel = filterLowHitCombos(skinnyComboChannel)
		return findBestSolvedMultiThread(itemOptions, model, skinnyComboChannel)
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

func makeSkinnyCombosMultiThread(itemOptions *SkinnyOptionsMap, model *Model, max, skip uint64) <-chan SkinnyItemSet {
	skinnyCombos := make(chan SkinnyItemSet, bufferSize)
	doneSignal := make(chan any, bufferSize)
	counters := [threadCount]uint64{}

	// track progress
	go trackProgressIntThreaded(&counters, max/skip, skinnyCombos, doneSignal)

	// start up workers
	splits := solve_util.IndexSplitsInt(max, skip, threadCount)
	for i := range threadCount {
		go createWorkerRangeInt(itemOptions, model, splits[i], splits[i+1], skip, skinnyCombos, doneSignal, &counters[i])
	}

	return skinnyCombos
}

func trackProgressIntThreaded(threadCounters *[threadCount]uint64, expectedTotal uint64, skinnyCombos chan<- SkinnyItemSet, doneSignal <-chan any) {
	startTime := time.Now()
	remaining := threadCount
	for {
		select {
		case <-time.After(time.Second * 5):
			var totalCount uint64 = 0
			for _, value := range threadCounters {
				totalCount += value
			}
			percent := float64(totalCount) / float64(expectedTotal)

			util.PrintProgressInt(startTime, percent, totalCount)

		case <-doneSignal:
			remaining--
			if remaining == 0 {
				close(skinnyCombos)
				return
			}
		}
	}
}

func createWorkerRangeInt(itemOptions *SkinnyOptionsMap, model *model.Model, start, max, skip uint64, skinnyCombos chan<- SkinnyItemSet, doneSignal chan<- any, progressCounter *uint64) {
	index := start
	for index < max {
		set := makeSkinnySetInt(itemOptions, index)
		if model.CheckSetSkinny(&set) {
			skinnyCombos <- set
		}

		index += skip
		(*progressCounter)++
	}

	doneSignal <- true
}

func makeSkinnySetInt(itemOptions *SkinnyOptionsMap, mainIndex uint64) SkinnyItemSet {
	equip := SkinnyEquipMap{}
	var a, b uint32

	currIndex := mainIndex

	for slot, array := range itemOptions {
		size := uint64(len(array))

		slotIndex := currIndex % size
		currIndex /= size

		item := array[slotIndex]
		equip[slot] = item
		a += item.A
		b += item.B
	}

	return SkinnyItemSet{Items: equip, A: a, B: b}
}

func findBestSolvedMultiThread(itemOptions *SolvableOptionsMap, model *Model, skinnyComboChannel <-chan SkinnyItemSet) SolvableItemSet {
	resultChannel := make(chan util.BestCollector1[SolvableItemSet], threadCount)

	for range threadCount {
		go findBestSolvedWorker(itemOptions, model, skinnyComboChannel, resultChannel)
	}

	return util.BestCollector1_OfChannel(resultChannel, threadCount)
}

func findBestSolvedWorker(itemOptions *SolvableOptionsMap, model *Model, skinnyComboChannel <-chan SkinnyItemSet, resultChannel chan util.BestCollector1[SolvableItemSet]) {
	best := util.BestCollector1[SolvableItemSet]{}
	for skinnySet := range skinnyComboChannel {
		solveSet := makeFromSkinny(itemOptions, model, &skinnySet)

		// assert still matches requirement, should be redundant
		if !model.CheckSet(&solveSet) {
			panic("inconsistent cap calcuations")
		}

		rating := model.CalcRatingSolve(&solveSet)
		best.Offer(&solveSet, rating)
	}
	resultChannel <- best
}

func makeFromSkinny(itemOptions *SolvableOptionsMap, model *Model, skinnySet *SkinnyItemSet) SolvableItemSet {
	chosen := SolvableItemSet{}
	for slot := Equip_Head; slot <= Equip_Offhand; slot++ {
		skinny := &skinnySet.Items[slot]
		options := itemOptions[slot]

		best := util.BestCollector1[SolvableItem]{}
		for _, item := range options {
			if model.StatRequirements.SkinnyMatch(skinny, &item) {
				rating := model.CalcRatingSolveItem(&item)
				best.Offer(&item, rating)
			}
		}

		chosen.AddItem_Mutating(slot, best.GetBestPointer())
	}
	return chosen
}

func filterLowHitCombos(inputChannel <-chan SkinnyItemSet) <-chan SkinnyItemSet {
	return util.Channel_TransformAll_Multi(threadCount, inputChannel, func(inputChannel <-chan SkinnyItemSet, outputChannel chan<- SkinnyItemSet) {
		valueHeap := util.LowestNIntHeap_For(filterTarget)
		for itemSet := range inputChannel {
			rating := uint64(itemSet.A + itemSet.B)
			if valueHeap.Offer(rating) {
				outputChannel <- itemSet
			}
		}
	})
}

func filterLowHitCombos0(inputChannel <-chan SkinnyItemSet) <-chan SkinnyItemSet {
	collectedChannel := make(chan util.LowestCollectorN[SkinnyItemSet])
	for range threadCount {
		go filterWorker(inputChannel, collectedChannel)
	}
	bestArray := util.LowestCollectorN_OfChannel(collectedChannel, threadCount)

	outputChannel := make(chan SkinnyItemSet)
	go func() {
		for _, item := range bestArray {
			outputChannel <- item
		}
		close(outputChannel)
	}()
	return outputChannel
}

func filterWorker(inputChannel <-chan SkinnyItemSet, collectedChannel chan<- util.LowestCollectorN[SkinnyItemSet]) {
	best := util.LowestCollector_ForN[SkinnyItemSet](filterTarget)
	for itemSet := range inputChannel {
		rating := itemSet.A + itemSet.B
		best.Offer(&itemSet, uint64(rating))
	}
	collectedChannel <- best
}
