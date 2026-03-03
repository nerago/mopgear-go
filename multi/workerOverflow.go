package multi

import (
	"math/big"
	"paladin_gearing_go/util"
	"sync"
)

func makeOverflowThreads(waitGroup *sync.WaitGroup, commonOptions commonComboOptions, threadCount uint64, eachThreadCount uint64, counters []uint64, comboChannel chan<- commonCombo) {
	skip := chooseSkip_PrimeAndIsntSlotSize(commonOptions, threadCount*eachThreadCount)
	for threadNum := range threadCount {
		waitGroup.Go(func() {
			evaluateOverflowWorker(commonOptions, eachThreadCount, threadNum, skip, &counters[threadNum], comboChannel)
		})
	}
}

func evaluateOverflowWorker(commonOptions commonComboOptions, loopCount uint64, threadNum uint64, skip *big.Int, doneCounter *uint64, comboChannel chan<- commonCombo) {
	indexes := make(map[uint32]uint32, len(commonOptions))

	initialSkip := big.NewInt(int64(threadNum * loopCount))
	initialSkip.Mul(initialSkip, skip)
	advanceArrays(commonOptions, indexes, initialSkip)

	for range loopCount {
		combo := makeComboAndAdvance(commonOptions, indexes, skip)
		comboChannel <- combo
		*doneCounter++
	}
}

func makeComboAndAdvance(commonOptions commonComboOptions, slotIndexes map[uint32]uint32, skip *big.Int) commonCombo {
	combo := make(commonCombo, len(commonOptions))

	remainingSkip := big.NewInt(0).Set(skip)
	temp := big.NewInt(0)
	mod := big.NewInt(0)

	for itemId, options := range commonOptions {
		index := slotIndexes[itemId]
		combo[itemId] = options[index]

		slotSize := len(options)
		if slotSize > 1 && remainingSkip.Cmp(util.Int_Zero) > 0 {
			remainingSkip.Add(remainingSkip, temp.SetUint64(uint64(index)))
			remainingSkip.DivMod(remainingSkip, temp.SetUint64(uint64(slotSize)), mod)
			slotIndexes[itemId] = uint32(mod.Uint64())
		}
	}
	return combo
}

func advanceArrays(commonOptions commonComboOptions, slotIndexes map[uint32]uint32, skip *big.Int) {
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

func chooseSkip_PrimeAndIsntSlotSize(commonOptions commonComboOptions, targetCount uint64) *big.Int {
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

func isFactorOfSlotSize(commonOptions commonComboOptions, skip *big.Int) bool {
	mod := big.NewInt(0)
	for _, options := range commonOptions {
		slotSize := big.NewInt(int64(len(options)))

		if slotSize.Cmp(skip) == 0 {
			return true
		} else if slotSize.Cmp(skip) > 0 {
			mod.Mod(slotSize, skip)
		} else {
			mod.Mod(skip, slotSize)
		}

		if mod.Cmp(util.Int_Zero) == 0 {
			return true
		}
	}
	return false

}
