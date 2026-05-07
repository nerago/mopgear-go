package channel_op

import (
	"paladin_gearing_go/util"
	"sync"
)

func makeOutputChannel[R any]() chan R {
	return make(chan R)
}

func Map_ChannelToChannel[T any, R any](threadCount int, inputChannel <-chan T, mapper func(T) R) <-chan R {
	var waitGroup sync.WaitGroup
	outputChannel := makeOutputChannel[R]()
	for range threadCount {
		waitGroup.Go(func() {
			for value := range inputChannel {
				outputChannel <- mapper(value)
			}
		})
	}
	go func() {
		waitGroup.Wait()
		close(outputChannel)
	}()
	return outputChannel
}

func Filter_ChannelToChannel[T any](threadCount int, inputChannel <-chan T, predicate func(T) bool) <-chan T {
	var waitGroup sync.WaitGroup
	outputChannel := makeOutputChannel[T]()
	for range threadCount {
		waitGroup.Go(func() {
			for value := range inputChannel {
				if predicate(value) {
					outputChannel <- value
				}
			}
		})
	}
	go func() {
		waitGroup.Wait()
		close(outputChannel)
	}()
	return outputChannel
}

func MapMulti_ChannelToChannel[T any, R any](threadCount int, inputChannel <-chan T, transform func(T, chan<- R)) <-chan R {
	var waitGroup sync.WaitGroup
	outputChannel := makeOutputChannel[R]()
	for range threadCount {
		waitGroup.Go(func() {
			for value := range inputChannel {
				transform(value, outputChannel)
			}
		})
	}
	go func() {
		waitGroup.Wait()
		close(outputChannel)
	}()
	return outputChannel
}

func TransformAll_ChannelToChannel[T any, R any](threadCount int, inputChannel <-chan T, transformAll func(int, <-chan T, chan<- R)) <-chan R {
	var waitGroup sync.WaitGroup
	outputChannel := makeOutputChannel[R]()
	for threadNum := range threadCount {
		waitGroup.Go(func() {
			transformAll(threadNum, inputChannel, outputChannel)
		})
	}
	go func() {
		waitGroup.Wait()
		close(outputChannel)
	}()
	return outputChannel
}

func GenerateAll_ToChannel[R any](threadCount int, generateSubGroup func(int, chan<- R)) <-chan R {
	var waitGroup sync.WaitGroup
	outputChannel := makeOutputChannel[R]()
	for threadNum := range threadCount {
		waitGroup.Go(func() {
			generateSubGroup(threadNum, outputChannel)
		})
	}
	go func() {
		waitGroup.Wait()
		close(outputChannel)
	}()
	return outputChannel
}

func Map_SliceToChannel[T any, R any](threadCount int, inputSlice []T, transform func(*T, chan<- R)) <-chan R {
	var waitGroup sync.WaitGroup
	outputChannel := makeOutputChannel[R]()

	inputLength := len(inputSlice)
	splits := indexSplitsInt(inputLength, threadCount)

	for threadNum := range threadCount {
		waitGroup.Go(func() {
			start := splits[threadNum]
			end := splits[threadNum+1]
			for index := start; index < end; index++ {
				transform(&inputSlice[index], outputChannel)
			}
		})
	}
	go func() {
		waitGroup.Wait()
		close(outputChannel)
	}()
	return outputChannel
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

func ForEach_BlockingTracked_Void[T any](trackProgress *util.TrackProgress, threadCount int, inputSlice []T, process func(*T)) {
	var waitGroup sync.WaitGroup

	inputLength := len(inputSlice)
	splits := indexSplitsInt(inputLength, threadCount)

	counts := make([]uint64, threadCount)
	trackProgress.RunFromArray(&counts, uint64(inputLength))

	for threadNum := range threadCount {
		waitGroup.Go(func() {
			start := splits[threadNum]
			end := splits[threadNum+1]
			for index := start; index < end; index++ {
				process(&inputSlice[index])
				counts[threadNum]++
			}
		})
	}

	waitGroup.Wait()
	trackProgress.Stop()
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
