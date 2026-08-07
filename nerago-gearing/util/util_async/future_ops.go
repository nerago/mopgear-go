package util_async

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

func waitForRoutineCompletionOrCancel[R any](itemResultChannel chan FutureResult[R], loopCancelChannel <-chan struct{}, outputSlice *[]R, activeFutureCount *int) bool {
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

func Future_MapValue[T any, R any](innerFuture *Future[T], mapper func(T) (R, bool)) *Future[R] {
	outerFuture := Future_Make[R]()

	go func() {
		value, hasValue := innerFuture.WaitForResult()
		if hasValue {
			newValue, hasNew := mapper(value)
			if hasNew {
				outerFuture.SetResult(newValue)
			} else {
				outerFuture.SetResultEmpty()
			}
		} else {
			outerFuture.SetResultEmpty()
		}
	}()

	return outerFuture
}

type FutureValueAdder[T any] struct {
	mutex          sync.Mutex
	waitingFutures int
	total          T
	combiner       func(T, T) T
	channel        chan T
	ready          bool
}

func FutureValueAdderMake[T any](initialValue T, combiner func(T, T) T) *FutureValueAdder[T] {
	return &FutureValueAdder[T]{total: initialValue, combiner: combiner, channel: make(chan T)}
}

func (fa *FutureValueAdder[T]) AddFuture(future IFutureWithResult[T]) {
	fa.mutex.Lock()
	fa.waitingFutures++
	fa.mutex.Unlock()

	future.ForwardResultToRelevantCallback(fa.processValue, fa.processFail)
}

func (fa *FutureValueAdder[T]) AddValueImmediate(apply func(T) T) {
	fa.mutex.Lock()
	defer fa.mutex.Unlock()

	if fa.ready {
		panic("AddValueImmediate should only be called before ready")
	}

	fa.total = apply(fa.total)
}

// confirm most of the expected futures have been added
// more can be added safely but may not be processed
func (fa *FutureValueAdder[T]) ReadyUpAndPrepareChannel() <-chan T {
	fa.mutex.Lock()
	defer fa.mutex.Unlock()

	fa.ready = true
	if fa.waitingFutures > 0 {
		return fa.channel
	} else {
		channelResult := fa.channel
		close(fa.channel)
		fa.channel = nil
		return channelResult
	}
}

func (fa *FutureValueAdder[T]) processValue(value T) {
	fa.mutex.Lock()
	defer fa.mutex.Unlock()

	fa.total = fa.combiner(fa.total, value)
	if fa.channel != nil {
		fa.channel <- fa.total
	}

	fa.waitingFutures--
	if fa.waitingFutures <= 0 && fa.ready {
		close(fa.channel)
		fa.channel = nil
	}
}

func (fa *FutureValueAdder[T]) processFail() {
	fa.mutex.Lock()
	defer fa.mutex.Unlock()

	fa.waitingFutures--
	if fa.waitingFutures <= 0 && fa.ready {
		close(fa.channel)
		fa.channel = nil
	}
}

type FutureChannelMixer[T any] struct {
	mutex         sync.Mutex
	activeSources int
	outputChannel chan T
	ready         bool
}

func FutureChannelMixerMake[T any]() *FutureChannelMixer[T] {
	return &FutureChannelMixer[T]{outputChannel: make(chan T)}
}

func (fc *FutureChannelMixer[T]) AddFuture(future IFutureWithResult[T]) {
	fc.mutex.Lock()
	fc.activeSources++
	fc.mutex.Unlock()

	future.ForwardResultToRelevantCallback(fc.processValue, fc.processFail)
}

func (fc *FutureChannelMixer[T]) AddChannel(inputChannel <-chan T) {
	fc.mutex.Lock()
	fc.activeSources++
	fc.mutex.Unlock()

	go func() {
		for value := range inputChannel {
			fc.outputChannel <- value
		}

		fc.mutex.Lock()
		fc.sourceFinished()
		fc.mutex.Unlock()
	}()
}

// adding directly to channel with a big buffer might be fairly reliable, but run inside a goroutine to ensure no block
func (fc *FutureChannelMixer[T]) AddValue(value T) {
	fc.mutex.Lock()
	fc.activeSources++
	fc.mutex.Unlock()

	go func() {
		fc.processValue(value)
	}()
}

// confirm most of the expected futures have been added
// more can be added safely but may not be processed
func (fc *FutureChannelMixer[T]) ReadyUpAndPrepareChannel() <-chan T {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	fc.ready = true
	if fc.activeSources > 0 {
		return fc.outputChannel
	} else {
		channelResult := fc.outputChannel
		close(fc.outputChannel)
		fc.outputChannel = nil
		return channelResult
	}
}

func (fc *FutureChannelMixer[T]) processValue(value T) {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	if fc.outputChannel != nil {
		fc.outputChannel <- value
	}

	fc.sourceFinished()
}

func (fc *FutureChannelMixer[T]) processFail() {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	fc.sourceFinished()
}

func (fc *FutureChannelMixer[T]) sourceFinished() {
	fc.activeSources--
	if fc.activeSources <= 0 && fc.ready {
		close(fc.outputChannel)
		fc.outputChannel = nil
	}
}
