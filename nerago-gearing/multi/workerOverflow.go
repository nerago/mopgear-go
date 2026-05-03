package multi

import (
	"math/big"
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
	"sync"
)

func makeOverflowThreads(waitGroup *sync.WaitGroup, commonOptions CommonComboOptions, threadCount uint64, eachThreadCount uint64, comboChannel chan<- commonCombo) {
	skip := chooseSkip_PrimeAndIsntSlotSize(commonOptions, threadCount*eachThreadCount)
	for threadNum := range threadCount {
		waitGroup.Go(func() {
			evaluateOverflowWorker(commonOptions, eachThreadCount, threadNum, skip, comboChannel)
		})
	}
}

func evaluateOverflowWorker(commonOptions CommonComboOptions, loopCount uint64, threadNum uint64, skip *big.Int, comboChannel chan<- commonCombo) {
	indexes := make(map[items.ItemId]uint32, len(commonOptions))

	initialSkip := big.NewInt(int64(threadNum * loopCount))
	initialSkip.Mul(initialSkip, skip)
	advanceArrays(commonOptions, indexes, initialSkip)

	for range loopCount {
		combo := makeComboAndAdvance(commonOptions, indexes, skip)
		comboChannel <- combo
	}
}

func makeComboAndAdvance(commonOptions CommonComboOptions, slotIndexes map[items.ItemId]uint32, skip *big.Int) commonCombo {
	combo := commonCombo_Make(len(commonOptions), comboType_overflow)

	remainingSkip := big.NewInt(0).Set(skip)
	temp := big.NewInt(0)
	mod := big.NewInt(0)

	for itemId, options := range commonOptions {
		index := slotIndexes[itemId]
		choice := &options[index]
		combo.addItem(itemId, choice, Force_Optional)

		slotSize := len(options)
		if slotSize > 1 && remainingSkip.Cmp(util.Int_Zero) > 0 {
			remainingSkip.Add(remainingSkip, temp.SetUint64(uint64(index)))
			remainingSkip.DivMod(remainingSkip, temp.SetUint64(uint64(slotSize)), mod)
			slotIndexes[itemId] = uint32(mod.Uint64())
		}
	}
	return combo
}

func advanceArrays(commonOptions CommonComboOptions, slotIndexes map[items.ItemId]uint32, skip *big.Int) {
	remainingSkip := big.NewInt(0).Set(skip)
	temp := big.NewInt(0)
	mod := big.NewInt(0)

	for itemId, options := range commonOptions {
		index := slotIndexes[itemId]
		slotSize := len(options)

		if slotSize > 1 {
			remainingSkip.Add(remainingSkip, temp.SetUint64(uint64(index)))
			remainingSkip.DivMod(remainingSkip, temp.SetUint64(uint64(slotSize)), mod)
			slotIndexes[itemId] = uint32(mod.Uint64())
			if remainingSkip.Cmp(util.Int_Zero) == 0 {
				return
			}
		}
	}
}

func chooseSkip_PrimeAndIsntSlotSize(commonOptions CommonComboOptions, targetCount uint64) *big.Int {
	comboCount := commonOptions.TotalCombinationCount()
	skip := util.ChooseSkip_NextPrimeFromRatio(comboCount, big.NewInt(int64(targetCount)))

	if skip.Cmp(util.Int_One) <= 0 {
		return util.Int_One
	}

	for isFactorOfSlotSize(commonOptions, skip) {
		skip = util.PrimeNextGreater(skip)
	}
	return skip
}

func isFactorOfSlotSize(commonOptions CommonComboOptions, skip *big.Int) bool {
	mod := big.NewInt(0)
	for _, options := range commonOptions {
		slotSizePrim := int64(len(options))
		if slotSizePrim <= 1 {
			continue // 0 or 1 is normal
		}

		slotSizeBig := big.NewInt(slotSizePrim)
		if slotSizeBig.Cmp(skip) == 0 {
			return true
		} else if slotSizeBig.Cmp(skip) > 0 {
			mod.Mod(slotSizeBig, skip)
		} else {
			mod.Mod(skip, slotSizeBig)
		}

		if mod.Cmp(util.Int_Zero) == 0 {
			return true
		}
	}
	return false

}
