package channel_op

import (
	"paladin_gearing_go/util"
	"sync"
)

func makeOutputChannel[R any]() chan R {
	return make(chan R)
}

func Map_ChannelToChannel[T any, R any](threadCount int, inputChannel <-chan T, mapper func(T, chan<- R)) <-chan R {
	var waitGroup sync.WaitGroup
	outputChannel := make(chan R)

	for range threadCount {
		waitGroup.Go(func() {
			for value := range inputChannel {
				mapper(value, outputChannel)
			}
		})
	}

	go func() {
		waitGroup.Wait()
		close(outputChannel)
	}()
	return outputChannel
}

func Map_ChannelToSlice[T any, R any](threadCount int, inputChannel <-chan T, mapper func(T, chan<- R)) []R {
	var waitGroup sync.WaitGroup
	tempChannel := make(chan R)

	for range threadCount {
		waitGroup.Go(func() {
			for value := range inputChannel {
				mapper(value, tempChannel)
			}
		})
	}

	go func() {
		waitGroup.Wait()
		close(tempChannel)
	}()

	outputSlice := make([]R, 0)
	for item := range tempChannel {
		outputSlice = append(outputSlice, item)
	}
	return outputSlice
}

func Map_SliceToChannel[T any, R any](threadCount int, inputSlice []T, mapper func(*T, chan<- R)) <-chan R {
	var waitGroup sync.WaitGroup

	inputLength := len(inputSlice)
	splits := indexSplitsInt(inputLength, threadCount)

	outputChannel := make(chan R)
	for threadNum := range threadCount {
		waitGroup.Go(func() {
			start := splits[threadNum]
			end := splits[threadNum+1]
			for index := start; index < end; index++ {
				mapper(&inputSlice[index], outputChannel)
			}
		})
	}

	go func() {
		waitGroup.Wait()
		close(outputChannel)
	}()
	return outputChannel
}

func Map_SliceToSlice[T any, R any](threadCount int, inputSlice []T, mapper func(*T, chan<- R)) []R {
	var waitGroup sync.WaitGroup

	inputLength := len(inputSlice)
	splits := indexSplitsInt(inputLength, threadCount)

	tempChannel := make(chan R)
	for threadNum := range threadCount {
		waitGroup.Go(func() {
			start := splits[threadNum]
			end := splits[threadNum+1]
			for index := start; index < end; index++ {
				mapper(&inputSlice[index], tempChannel)
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

func PermuteAsChannel[T any](listsOfOptions [][]T) <-chan []T {
	stepChannel := make(chan []T, 8)
	go func() {
		for _, value := range listsOfOptions[0] {
			stepChannel <- []T{value}
		}
		close(stepChannel)
	}()

	for i := 1; i < len(listsOfOptions); i++ {
		stepChannel = permuteStep(stepChannel, listsOfOptions[i])
	}

	return stepChannel
}

func permuteStep[T any](inChannel chan []T, options []T) chan []T {
	outputChannel := make(chan []T, 8)
	go func() {
		for currSlice := range inChannel {
			for _, value := range options {
				outputChannel <- util.CopyAndAppend(currSlice, value)
			}
		}
		close(outputChannel)
	}()
	return outputChannel
}
