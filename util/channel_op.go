package util

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

func Channel_Map_Multi[T any, R any](inputChannel <-chan T, mapper func(T) R, threadCount int) <-chan R {
	outputChannel := make(chan R, bufferSize)
	doneChannel := make(chan any, threadCount)
	for range threadCount {
		go func() {
			for value := range inputChannel {
				outputChannel <- mapper(value)
			}
			doneChannel <- true
		}()
	}
	go func() {
		for range threadCount {
			_ = <-doneChannel
		}
		close(outputChannel)
	}()
	return outputChannel
}

func Channel_Filter_Multi[T any](inputChannel <-chan T, predicate func(T) bool, threadCount int) <-chan T {
	outputChannel := make(chan T, bufferSize)
	doneChannel := make(chan any, threadCount)
	for range threadCount {
		go func() {
			for value := range inputChannel {
				if predicate(value) {
					outputChannel <- value
				}
			}
			doneChannel <- true
		}()
	}
	go func() {
		for range threadCount {
			_ = <-doneChannel
		}
		close(outputChannel)
	}()
	return outputChannel
}

func Channel_Transform_Multi[T any, R any](inputChannel <-chan T, transform func(T, chan<- R), threadCount int) <-chan R {
	outputChannel := make(chan R, bufferSize)
	doneChannel := make(chan any, threadCount)
	for range threadCount {
		go func() {
			for value := range inputChannel {
				transform(value, outputChannel)
			}
			doneChannel <- true
		}()
	}
	go func() {
		for range threadCount {
			_ = <-doneChannel
		}
		close(outputChannel)
	}()
	return outputChannel
}
