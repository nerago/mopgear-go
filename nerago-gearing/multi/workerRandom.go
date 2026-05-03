package multi

import (
	"math/rand"
	"sync"
)

func makeRandomThreads(waitGroup *sync.WaitGroup, commonOptions CommonComboOptions, threadCount uint64, eachThreadCount uint64, comboChannel chan commonCombo) {
	for threadNum := range threadCount {
		waitGroup.Go(func() {
			makeRandomWorker(commonOptions, eachThreadCount, threadNum, comboChannel)
		})
	}
}

func makeRandomWorker(commonOptions CommonComboOptions, loopCount uint64, threadNum uint64, comboChannel chan<- commonCombo) {
	rng := rand.New(rand.NewSource(int64(threadNum)))
	for range loopCount {
		combo := makeRandomCombo(commonOptions, rng)
		comboChannel <- combo
	}
}

func makeRandomCombo(commonOptions CommonComboOptions, rng *rand.Rand) commonCombo {
	combo := commonCombo_Make(len(commonOptions), comboType_random)
	for itemId, options := range commonOptions {
		index := rng.Intn(len(options))
		choice := &options[index]
		combo.addItem(itemId, choice, Force_Optional)
	}
	return combo
}
