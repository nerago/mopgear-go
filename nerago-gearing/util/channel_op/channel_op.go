package channel_op

import (
	"sync"
)

func makeOutputChannel[R any]() chan R {
	return make(chan R)
}

func Map_SliceToSlice[T any, R any](threadCount int, inputSlice []T, transform func(*T, chan<- R)) []R {
	var waitGroup sync.WaitGroup

	inputLength := len(inputSlice)
	splits := indexSplitsInt(inputLength, threadCount)

	tempChannel := makeOutputChannel[R]()
	for threadNum := range threadCount {
		waitGroup.Go(func() {
			start := splits[threadNum]
			end := splits[threadNum+1]
			for index := start; index < end; index++ {
				transform(&inputSlice[index], tempChannel)
			}
		})
	}

	go func() {
		waitGroup.Wait()
		close(tempChannel)
	}()

	outputSlice := make([]R, 0, inputLength)
	for item := range tempChannel {
		outputSlice = append(outputSlice, item)
	}
	return outputSlice
}

func ForEach_Blocking_Void[T any](threadCount int, inputSlice []T, process func(*T)) {
	var waitGroup sync.WaitGroup

	inputLength := len(inputSlice)
	splits := indexSplitsInt(inputLength, threadCount)

	for threadNum := range threadCount {
		waitGroup.Go(func() {
			start := splits[threadNum]
			end := splits[threadNum+1]
			for index := start; index < end; index++ {
				process(&inputSlice[index])
			}
		})
	}

	waitGroup.Wait()
}

func indexSplitsInt(sliceLength int, threadCount int) []int {
	indexPerThread := sliceLength / threadCount

	splitArray := make([]int, 0, threadCount+1)
	start := 0
	for range threadCount {
		splitArray = append(splitArray, start)
		start += indexPerThread
	}
	splitArray = append(splitArray, sliceLength)

	return splitArray
}
