package multi

import (
	"context"
	"math/big"
	"math/rand"
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
	"sync"
)

// const bufferSize = 256
const additionalSetCount = 10

type commonCombo map[uint32]items.FullItem

func (job *MultiSetJob) makeCommonChannel(commonOptions commonComboOptions, targetCount uint64) <-chan commonCombo {
	eachThreadCount := max(targetCount/generateThreadCount, 1)
	counters := make([]uint64, generateThreadCount)
	// comboCount := combinationCount(commonOptions)

	ctx, cancel := context.WithCancel(context.Background())
	go util.TrackProgressIntThreaded(ctx, &counters, targetCount)

	var waitGroup sync.WaitGroup
	comboChannel := make(chan commonCombo)
	// comboChannel := make(chan commonCombo, bufferSize)
	for threadNum := range generateThreadCount {
		waitGroup.Go(func() {
			makeCommonWorker(commonOptions, eachThreadCount, threadNum, &counters[threadNum], comboChannel)
		})
	}

	// waitGroup.Go(func() { makeBaselineWorker(job.params, commonOptions, comboChannel) })
	// waitGroup.Go(func() { makeEquippedWorker(job.params, commonOptions, comboChannel) })

	go func() {
		waitGroup.Wait()
		close(comboChannel)
		cancel()
	}()
	return comboChannel
}

func makeCommonWorker(commonOptions commonComboOptions, loopCount uint64, threadNum int, doneCounter *uint64, comboChannel chan<- commonCombo) {
	rng := rand.New(rand.NewSource(int64(threadNum)))
	for range loopCount {
		combo := makeRandomCombo(commonOptions, rng)
		comboChannel <- combo
		*doneCounter++
	}
}

func makeRandomCombo(commonOptions commonComboOptions, rng *rand.Rand) commonCombo {
	combo := make(commonCombo)
	for itemId, options := range commonOptions {
		index := rng.Intn(len(options))
		combo[itemId] = options[index]
	}
	return combo
}

func makeBaselineWorker(params []MultiSetParam, commonOptions commonComboOptions, comboChannel chan<- commonCombo) {
	rng := rand.New(rand.NewSource(0xBA5E))
	for _, param := range params {
		for range additionalSetCount {
			combo := make(commonCombo)

			// copy what items in baseline set
			for item := range param.baselineResult.FullSet.Items.AllItemSeq() {
				combo[item.ItemId()] = *item
			}

			fillOutRemainingOptions(commonOptions, combo, rng)

			comboChannel <- combo
		}
	}
}

func makeEquippedWorker(params []MultiSetParam, commonOptions commonComboOptions, comboChannel chan<- commonCombo) {
	rng := rand.New(rand.NewSource(0xE819))
	for _, param := range params {
		for range additionalSetCount {
			combo := make(commonCombo)

			// copy what items in equipped set
			for item := range param.exactEquippedGear.AllItemSeq() {
				combo[item.ItemId()] = *item
			}

			fillOutRemainingOptions(commonOptions, combo, rng)

			comboChannel <- combo
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
