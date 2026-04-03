package multi

import (
	"math/rand"
	"sync"
)

func makeRandomThreads(waitGroup *sync.WaitGroup, commonOptions commonComboOptions, threadCount uint64, eachThreadCount uint64, counters []uint64, comboChannel chan commonCombo) {
	for threadNum := range threadCount {
		waitGroup.Go(func() {
			makeRandomWorker(commonOptions, eachThreadCount, threadNum, &counters[threadNum], comboChannel)
		})
	}
}

func makeRandomWorker(commonOptions commonComboOptions, loopCount uint64, threadNum uint64, doneCounter *uint64, comboChannel chan<- commonCombo) {
	rng := rand.New(rand.NewSource(int64(threadNum)))
	for range loopCount {
		combo := makeRandomCombo(commonOptions, rng)
		comboChannel <- combo
		*doneCounter++
	}
}

func makeRandomCombo(commonOptions commonComboOptions, rng *rand.Rand) commonCombo {
	combo := commonCombo_Make(len(commonOptions), comboType_random)
	for itemId, options := range commonOptions {
		index := rng.Intn(len(options))
		choice := &options[index]
		combo.addItem(itemId, choice)
	}
	return combo
}
