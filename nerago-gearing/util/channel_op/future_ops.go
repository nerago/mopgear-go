package channel_op

import (
	"sync"
)

func MapFuture_SliceToChannel_Cancellable[T any, R any](threadCount int, inputSlice []T, primaryCancel CancelSignal, mapper func(*T) *FutureCancellable[R]) <-chan R {
	indexChannel := make(chan int)
	loopCancelChannel := primaryCancel.CancelSignalChannel()

	go func() {
	indexLoop:
		for index := range inputSlice {
			select {
			case indexChannel <- index: // send index to next routine
			case <-loopCancelChannel: // break out on cancel
				break indexLoop
			}
		}
		close(indexChannel)
	}()

	outputChannel := makeOutputChannel[R]()
	var waitGroup sync.WaitGroup
	for range threadCount {
		waitGroup.Go(func() {
			for index := range indexChannel {
				future := mapper(&inputSlice[index])
				if future != nil {
					ChainCancel(primaryCancel, future)
					value, hasValue := future.WaitForResult()
					if hasValue {
						outputChannel <- value
					}
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

func MapFuture_SliceToSlice_FutureCancellable[T any, R any](threadCount int, inputSlice []T, mapper func(*T) *FutureCancellable[R]) *FutureCancellable[[]R] {
	primaryFuture := FutureCancellable_Make[[]R]()
	loopCancelChannel := primaryFuture.CancelSignalChannel()

	go func() {
		launchIndex := 0
		itemResultChannel := make(chan FutureResult[R], threadCount)
		outputSlice := make([]R, 0)
		activeFutureCount := 0

		for launchIndex < len(inputSlice) && primaryFuture.ShouldContinue() {
			// launch new futures
			for activeFutureCount < threadCount {
				entry := &inputSlice[launchIndex]
				launchIndex++

				// launch on its own goroutine in case setup takes time
				mapAndLaunchAsFuture(entry, mapper, primaryFuture, itemResultChannel)
				activeFutureCount++
			}

			// wait for at least one future to finish
			if waitForRoutineCompletionOrCancel(itemResultChannel, loopCancelChannel, &outputSlice, &activeFutureCount) {
				break
			}

			// drain any additional completed futures
			drainCompletedFutures(itemResultChannel, &outputSlice, &activeFutureCount)
		}

		waitForRemainingCompletionAndSetResult(itemResultChannel, &outputSlice, &activeFutureCount, primaryFuture)
	}()

	return primaryFuture
}

func mapAndLaunchAsFuture[T any, R any](entry *T, mapper func(*T) *FutureCancellable[R], primaryFuture *FutureCancellable[[]R], itemResultChannel chan FutureResult[R]) {
	go func() {
		itemFuture := mapper(entry)
		if itemFuture != nil {
			ChainCancel(primaryFuture, itemFuture)
			itemFuture.ForwardAnyResultToChannel(itemResultChannel)
		} else {
			itemResultChannel <- FutureResult[R]{HasValue: false}
		}
	}()
}

func waitForRoutineCompletionOrCancel[R any](itemResultChannel chan FutureResult[R], loopCancelChannel <-chan any, outputSlice *[]R, activeFutureCount *int) bool {
	select {
	case itemResult := <-itemResultChannel:
		if itemResult.HasValue {
			*outputSlice = append(*outputSlice, itemResult.Value)
		}
		*activeFutureCount--
		return false
	case <-loopCancelChannel:
		return true
	}
}

func drainCompletedFutures[R any](itemResultChannel chan FutureResult[R], outputSlice *[]R, activeFutureCount *int) {
	for {
		select {
		case itemResult := <-itemResultChannel:
			if itemResult.HasValue {
				*outputSlice = append(*outputSlice, itemResult.Value)
			}
			*activeFutureCount--
		default:
			return
		}
	}
}

func waitForRemainingCompletionAndSetResult[R any](itemResultChannel chan FutureResult[R], outputSlice *[]R, activeFutureCount *int, primaryFuture *FutureCancellable[[]R]) {
	for *activeFutureCount > 0 {
		itemResult := <-itemResultChannel
		if itemResult.HasValue {
			*outputSlice = append(*outputSlice, itemResult.Value)
		}
		*activeFutureCount--
	}

	primaryFuture.SetResult(*outputSlice)
}
