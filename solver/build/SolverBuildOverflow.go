package build

import (
	"context"
	"math/big"
	. "paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/util"
)

func SolverBuildOverflow_Run(itemOptions *SolvableOptionsMap, model *model.Model, targetCount uint64, trackProgress bool, printer *util.PrintRecorder) util.Optional[SolvableItemSet] {
	printer.Printf("SOLVE OVERFLOW %d\n", targetCount)
	return evaluateOverflow(itemOptions, model, targetCount, trackProgress, defaultEvaluateThreadCount, emptyPeekFunc)
}

func evaluateOverflow(itemOptions *SolvableOptionsMap, model *model.Model, targetCount uint64, trackProgress bool, threadCount int, peekFunc func(*SolvableItemSet)) util.Optional[SolvableItemSet] {
	resultChannel := make(chan util.BestCollector1[SolvableItemSet], threadCount)
	eachThreadCount := max(targetCount/uint64(threadCount), 1)
	skip := chooseSkip_PrimeAndIsntSlotSize(itemOptions, targetCount)
	counters := make([]uint64, threadCount)

	// track progress with cancel
	if trackProgress {
		ctx, cancel := context.WithCancel(context.Background())
		go util.TrackProgressIntThreaded(ctx, &counters, targetCount)
		defer cancel()
	}

	for threadNum := range threadCount {
		go evaluateOverflowWorker(resultChannel, model, eachThreadCount, itemOptions, skip, uint64(threadNum), &counters[threadNum], peekFunc)
	}

	// combine each thread's best result
	return util.BestCollector1_OfChannel(resultChannel, threadCount)
}

func chooseSkip_PrimeAndIsntSlotSize(itemOptions *SolvableOptionsMap, targetCount uint64) uint64 {
	comboCount := itemOptions.TotalCombinationCount()
	skip := util.ChooseSkip_NextPrimeFromRatio(comboCount, big.NewInt(int64(targetCount)))

	if skip.Cmp(util.Int_One) == 0 {
		return 1
	} else if !skip.IsUint64() {
		panic("big num not handled")
	}

	for isASlotSize(itemOptions, skip.Uint64()) {
		skip = util.PrimeNextGreater(skip)
	}
	return skip.Uint64()
}

func isASlotSize(itemOptions *SolvableOptionsMap, skip uint64) bool {
	for _, options := range itemOptions {
		if uint64(len(options)) == skip {
			return true
		}
	}
	return false
}

func evaluateOverflowWorker(resultChannel chan util.BestCollector1[SolvableItemSet], model *model.Model, eachThreadCount uint64, itemOptions *SolvableOptionsMap, skip uint64, threadNum uint64, processedCounter *uint64, peekFunc func(*SolvableItemSet)) {
	best := util.BestCollector1[SolvableItemSet]{}

	indexes := [16]uint64{}
	slotSizes := makeSlotSizes(itemOptions)
	advanceArrays(&indexes, &slotSizes, skip*threadNum*eachThreadCount)

	for range eachThreadCount {
		itemSet := makeSetFromArraysDirect(itemOptions, &indexes)
		advanceArrays(&indexes, &slotSizes, skip)
		if peekFunc != nil {
			peekFunc(&itemSet)
		}
		if model.CheckSet(&itemSet) {
			rating := model.CalcRatingSolve(&itemSet)
			best.Offer(&itemSet, rating)
		}
		(*processedCounter)++
	}

	resultChannel <- best
}

func makeSetFromArraysDirect(slotOptions *SolvableOptionsMap, slotIndexes *[16]uint64) SolvableItemSet {
	equip := SolvableEquipMap{}
	for slot, options := range slotOptions {
		if options != nil {
			index := slotIndexes[slot]
			equip[slot] = &options[index]
		}
	}
	return SolvableItemSet_Of(equip)
}

func advanceArrays(indexes *[16]uint64, sizes *[16]uint64, skip uint64) {
	for slot := range indexes {
		slotSize := sizes[slot]
		if slotSize > 1 {
			// TODO this still has issues if skip==slotSize, or a factor thereof
			value := indexes[slot] + skip
			indexes[slot] = value % slotSize
			skip = value / slotSize
			if skip == 0 {
				return
			}
		}
	}
}

func makeSlotSizes(itemOptions *SolvableOptionsMap) [16]uint64 {
	slotSizes := [16]uint64{}
	for slot, options := range itemOptions {
		slotSizes[slot] = uint64(len(options))
	}
	return slotSizes
}
