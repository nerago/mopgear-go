package indexed

import (
	"math/big"
	. "paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/util"
)

const threadCount = 12

var int_one = big.NewInt(1)

func SolverIndexed_RunSkipping(itemOptions *SolvableOptionsMap, model *model.Model, targetCount uint64, printer *util.PrintRecorder) SolvableItemSet {
	max, skip := initSkipValues(itemOptions, targetCount)

	printer.Printf("SOLVE SKIP %d %d %d\n", max, targetCount, skip)

	return mainLoop(itemOptions, max, skip, model)
}

func initSkipValues(itemOptions *SolvableOptionsMap, targetCount uint64) (*big.Int, *big.Int) {
	max := itemOptions.TotalCombinationCount()
	targetCombination := big.NewInt(int64(targetCount))
	skip := util.ChooseSkip_NextPrimeFromRatio(max, targetCombination)
	return max, skip
}

func SolverIndexed_RunFull(itemOptions *SolvableOptionsMap, model *model.Model, printer *util.PrintRecorder) SolvableItemSet {
	max := itemOptions.TotalCombinationCount()
	printer.Printf("SOLVE FULL %d\n", max)
	return mainLoop(itemOptions, max, int_one, model)
}

func mainLoop(itemOptions *SolvableOptionsMap, max, skip *big.Int, model *model.Model) SolvableItemSet {
	// TODO consider partitioning some slots until under limit
	if max.IsUint64() && skip.IsUint64() {
		return mainLoop_multiThread_int(itemOptions, max.Uint64(), skip.Uint64(), model, emptyPeekFunc)
	} else {
		return mainLoop_multiThread_big(itemOptions, max, skip, model, emptyPeekFunc)
	}

	// if max.IsUint64() && skip.IsUint64() {
	// 	return mainLoop_singleThread_int(itemOptions, max.Uint64(), skip.Uint64(), model)
	// } else {
	// 	return mainLoop_singleThread_big(itemOptions, max, skip, model)
	// }
}

func slotSizesBig(itemOptions *SolvableOptionsMap) [16]*big.Int {
	slotSizes := [16]*big.Int{}
	for i, array := range itemOptions {
		if len(array) > 0 {
			slotSizes[i] = big.NewInt(int64(len(array)))
		}
	}
	return slotSizes
}

func makeSetBig(itemOptions *SolvableOptionsMap, slotSizes *[16]*big.Int, mainIndex *big.Int) SolvableItemSet {
	equip := SolvableEquipMap{}

	currIndex := big.NewInt(0)
	currIndex.Set(mainIndex)
	mod := big.NewInt(0)

	for slot, array := range itemOptions {
		size := slotSizes[slot]
		if size != nil {
			currIndex.DivMod(currIndex, size, mod)
			slotIndex := mod.Int64()

			equip[slot] = &array[slotIndex]
		}
	}

	return SolvableItemSet_Of(equip)
}

func makeSetInt(itemOptions *SolvableOptionsMap, mainIndex uint64) SolvableItemSet {
	equip := SolvableEquipMap{}

	currIndex := mainIndex

	for slot, array := range itemOptions {
		size := uint64(len(array))
		if size > 0 {
			slotIndex := currIndex % size
			currIndex /= size

			equip[slot] = &array[slotIndex]
		}
	}

	return SolvableItemSet_Of(equip)
}

func emptyPeekFunc(*SolvableItemSet) {
}
