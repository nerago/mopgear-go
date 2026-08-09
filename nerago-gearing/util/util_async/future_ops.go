package util_async

import (
	"paladin_gearing_go/util/util_collection"
	"reflect"
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

type FutureValueAdderInt struct {
	FutureValueAdder[int]
}

func FutureValueAdderIntMake(initialValue int) *FutureValueAdderInt {
	return &FutureValueAdderInt{
		FutureValueAdder: FutureValueAdder[int]{total: initialValue, combiner: func(a int, b int) int { return a + b }, channel: make(chan int)}}
}

func (fa *FutureValueAdderInt) AddValueImmediate(value int) {
	fa.FutureValueAdder.AddValueImmediate(func(x int) int { return x + value })
}

type FutureChannelMixer[T any] struct {
	mutex         sync.Mutex
	activeSources int
	outputChannel chan T
	ready         bool
}

func FutureChannelMixerMake[T any]() *FutureChannelMixer[T] {
	return &FutureChannelMixer[T]{outputChannel: make(chan T, 64)}
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

type FutureChannelMixer2[T any] struct {
	mutex               sync.Mutex
	activeSources       int
	outputChannel       chan T
	ready               bool
	sourceFutures       []*Future[T]
	sourceFuturesCancel []*FutureCancellable[T]
	sourceChannels      []<-chan T
	sourceValues        []T
}

func (fc *FutureChannelMixer2[T]) AddFuture(future *Future[T]) {
	if fc.ready {
		panic("can't add source after ready")
	}
	fc.mutex.Lock()
	fc.sourceFutures = append(fc.sourceFutures, future)
	fc.mutex.Unlock()
}

func (fc *FutureChannelMixer2[T]) AddFutureCancellable(future *FutureCancellable[T]) {
	if fc.ready {
		panic("can't add source after ready")
	}
	fc.mutex.Lock()
	fc.sourceFuturesCancel = append(fc.sourceFuturesCancel, future)
	fc.mutex.Unlock()
}

func (fc *FutureChannelMixer2[T]) AddChannel(inputChannel <-chan T) {
	if fc.ready {
		panic("can't add source after ready")
	}
	fc.mutex.Lock()
	fc.sourceChannels = append(fc.sourceChannels, inputChannel)
	fc.mutex.Unlock()
}

// adding directly to channel with a big buffer might be fairly reliable, but run inside a goroutine to ensure no block
func (fc *FutureChannelMixer2[T]) AddValue(value T) {
	if fc.ready {
		panic("can't add source after ready")
	}
	fc.mutex.Lock()
	fc.sourceValues = append(fc.sourceValues, value)
	fc.mutex.Unlock()
}

// confirm most of the expected futures have been added
// more can be added safely but may not be processed
func (fc *FutureChannelMixer2[T]) ReadyUpAndPrepareChannel() <-chan T {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()

	fc.ready = true
	if len(fc.sourceChannels) == 0 && len(fc.sourceFutures) == 0 && len(fc.sourceFuturesCancel) == 0 {
		fc.outputChannel = make(chan T, len(fc.sourceValues))
		fc.drainValues()
		close(fc.outputChannel)
		return fc.outputChannel
	}

	fc.outputChannel = make(chan T)
	go func() {
		fc.drainValues()
		fc.selectLoop()
		close(fc.outputChannel)
	}()
	return fc.outputChannel
}

func (fc *FutureChannelMixer2[T]) drainValues() {
	for _, value := range fc.sourceValues {
		fc.outputChannel <- value
	}
	fc.sourceValues = nil
}

func (fc *FutureChannelMixer2[T]) selectLoop() {
	count := len(fc.sourceChannels) + len(fc.sourceFutures) + len(fc.sourceFuturesCancel)
	cases := make([]reflect.SelectCase, 0, count)
	for _, channel := range fc.sourceChannels {
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(channel),
		})
	}
	for _, future := range fc.sourceFutures {
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(future.resultChannel),
		})
	}
	for _, future := range fc.sourceFuturesCancel {
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(future.resultChannel),
		})
	}

	for len(cases) > 0 {
		chosen, reflectValue, hasValue := reflect.Select(cases)
		if hasValue {
			switch value := reflectValue.Interface().(type) {
			case FutureResult[T]:
				if value.HasValue {
					fc.outputChannel <- value.Value
				}
				cases[chosen].Chan.Close()
				util_collection.DeleteIndexInPlace(&cases, chosen)
			case T:
				fc.outputChannel <- value
			}
		} else {
			util_collection.DeleteIndexInPlace(&cases, chosen)
		}
	}
}
