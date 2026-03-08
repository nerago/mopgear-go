package util

import "sync"

// const bufferSize = 128

func makeOutputChannel[R any]() chan R {
	// return make(chan R, bufferSize)
	return make(chan R)
}

func Channel_Map_Single[T any, R any](inputChannel <-chan T, mapper func(T) R) <-chan R {
	outputChannel := makeOutputChannel[R]()
	go func() {
		for value := range inputChannel {
			outputChannel <- mapper(value)
		}
		close(outputChannel)
	}()
	return outputChannel
}

func Channel_Filter_Single[T any](inputChannel <-chan T, predicate func(T) bool) <-chan T {
	outputChannel := makeOutputChannel[T]()
	go func() {
		for value := range inputChannel {
			if predicate(value) {
				outputChannel <- value
			}
		}
		close(outputChannel)
	}()
	return outputChannel
}

func Channel_Map_Multi[T any, R any](threadCount int, inputChannel <-chan T, mapper func(T) R) <-chan R {
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

func Channel_Filter_Multi[T any](threadCount int, inputChannel <-chan T, predicate func(T) bool) <-chan T {
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

func Channel_TransformEach_Multi[T any, R any](threadCount int, inputChannel <-chan T, transform func(T, chan<- R)) <-chan R {
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

func Channel_TransformAll_Multi[T any, R any](threadCount int, inputChannel <-chan T, transformAll func(<-chan T, chan<- R)) <-chan R {
	var waitGroup sync.WaitGroup
	outputChannel := makeOutputChannel[R]()
	for range threadCount {
		waitGroup.Go(func() {
			transformAll(inputChannel, outputChannel)
		})
	}
	go func() {
		waitGroup.Wait()
		close(outputChannel)
	}()
	return outputChannel
}

func Channel_GenerateAll_Multi[R any](threadCount int, generateSubGroup func(int, chan<- R), after func()) <-chan R {
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
		if after != nil {
			after()
		}
	}()
	return outputChannel
}

func Channel_IterateEach_Multi[T any, R any](threadCount int, inputSlice []T, transform func(*T, chan<- R)) <-chan R {
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

func Void_IterateEach_Multi_Blocking[T any](threadCount int, inputSlice []T, process func(*T)) {
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

func Void_IterateEach_Multi_BlockingTracked[T any](threadCount int, inputSlice []T, process func(*T)) {
	var waitGroup sync.WaitGroup

	inputLength := len(inputSlice)
	splits := indexSplitsInt(inputLength, threadCount)

	counts := make([]uint64, threadCount)
	trackProgress := TrackProgress_Start()
	trackProgress.RunFromArray(&counts, uint64(inputLength))
	defer trackProgress.Stop()

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

func Channel_Collect[T any](inputChannel <-chan T) []T {
	slice := make([]T, 0)
	for item := range inputChannel {
		slice = append(slice, item)
	}
	return slice
}
