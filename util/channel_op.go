package util

import "sync"

const bufferSize = 128

func Channel_Map_Single[T any, R any](inputChannel <-chan T, mapper func(T) R) <-chan R {
	outputChannel := make(chan R, bufferSize)
	go func() {
		for value := range inputChannel {
			outputChannel <- mapper(value)
		}
		close(outputChannel)
	}()
	return outputChannel
}

func Channel_Filter_Single[T any](inputChannel <-chan T, predicate func(T) bool) <-chan T {
	outputChannel := make(chan T, bufferSize)
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
	outputChannel := make(chan R, bufferSize)
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
	outputChannel := make(chan T, bufferSize)
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
	outputChannel := make(chan R, bufferSize)
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
	outputChannel := make(chan R, bufferSize)
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
	outputChannel := make(chan R, bufferSize)
	for threadNum := range threadCount {
		waitGroup.Go(func() {
			generateSubGroup(threadNum, outputChannel)
		})
	}
	go func() {
		waitGroup.Wait()
		close(outputChannel)
		after()
	}()
	return outputChannel
}
