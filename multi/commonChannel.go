package multi

import (
	"math/big"
	"math/rand"
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
	"sync"
)

const additionalSetEach uint64 = 64
const additionalThreads uint64 = 2

type commonCombo map[uint32]items.FullItem

func (job *MultiSetJob) makeCommonChannel(commonOptions commonComboOptions, targetCount uint64, trackProgress *util.TrackProgress) <-chan commonCombo {
	counters := make([]uint64, generateThreadCount+additionalThreads)
	additionalCount := additionalSetEach * additionalThreads
	eachThreadCount := max((targetCount-additionalCount)/generateThreadCount, 1)

	trackProgress.RunFromArray(&counters, targetCount)

	var waitGroup sync.WaitGroup
	comboChannel := make(chan commonCombo)
	waitGroup.Go(func() { makeBaselineWorker(&job.params, commonOptions, &counters[0], comboChannel) })
	waitGroup.Go(func() { makeEquippedWorker(&job.params, commonOptions, &counters[1], comboChannel) })

	makeRandomThreads(&waitGroup, commonOptions, generateThreadCount/2, eachThreadCount, counters[2:2+generateThreadCount/2], comboChannel)
	makeOverflowThreads(&waitGroup, commonOptions, generateThreadCount/2, eachThreadCount, counters[2+generateThreadCount/2:], comboChannel)

	go func() {
		waitGroup.Wait()
		close(comboChannel)
	}()
	return comboChannel
}

func makeBaselineWorker(params *[]MultiSetParam, commonOptions commonComboOptions, doneCounter *uint64, comboChannel chan<- commonCombo) {
	rng := rand.New(rand.NewSource(0xBA5E))
	for paramIndex := range *params {
		param := &(*params)[paramIndex]
		for range additionalSetEach {
			combo := make(commonCombo)

			// copy what items are in baseline set
			for item := range param.baselineResult.FullSet.Items.AllItemSeq() {
				combo[item.ItemId()] = *item
			}

			fillOutRemainingOptions(commonOptions, combo, rng)

			comboChannel <- combo
			*doneCounter++
		}
	}
}

func makeEquippedWorker(params *[]MultiSetParam, commonOptions commonComboOptions, doneCounter *uint64, comboChannel chan<- commonCombo) {
	rng := rand.New(rand.NewSource(0xE819))
	for paramIndex := range *params {
		param := &(*params)[paramIndex]
		for range additionalSetEach {
			combo := make(commonCombo)

			// copy what items are in equipped set
			for item := range param.exactEquippedGear.AllItemSeq() {
				combo[item.ItemId()] = *item
			}

			fillOutRemainingOptions(commonOptions, combo, rng)

			comboChannel <- combo
			*doneCounter++
		}
	}
}

func fillOutRemainingOptions(commonOptions commonComboOptions, combo commonCombo, rng *rand.Rand) {
	for itemId, options := range commonOptions {
		_, alreadySet := combo[itemId]
		if !alreadySet {
			index := rng.Intn(len(options))
			combo[itemId] = options[index]
		}
	}
}

func combinationCount(options commonComboOptions) *big.Int {
	valueCount := 0
	total := big.NewInt(1)
	for _, slotArray := range options {
		slotSize := int64(len(slotArray))
		if slotSize > 0 {
			total.Mul(total, big.NewInt(slotSize))
			valueCount++
		}
	}
	if valueCount == 0 {
		panic("empty options")
	}
	return total
}
