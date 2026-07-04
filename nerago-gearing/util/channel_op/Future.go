package channel_op

import (
	"os"
	"paladin_gearing_go/util"
	"sync"
)

// ########### Future ###########

type FutureVoid[T any] struct {
	isComplete    bool
	lock          sync.Mutex
	signalChannel chan any
	hasWaiter     bool
}

func FutureVoid_Make[T any]() *FutureVoid[T] {
	return &FutureVoid[T]{
		signalChannel: make(chan any, 1),
	}
}

func (fv *FutureVoid[T]) SetResultEmpty() {
	fv.lock.Lock()
	if !fv.isComplete {
		fv.isComplete = true
		fv.signalChannel <- true
	}
	fv.lock.Unlock()
}

func (fv *FutureVoid[T]) WaitForComplete() {
	if fv == nil || fv.signalChannel == nil {
		panic("invalid FutureVoid")
	} else if fv.hasWaiter {
		panic("duplicate waiter")
	}
	fv.hasWaiter = true

	_, channelOk := <-fv.signalChannel
	if !channelOk {
		panic("signal channel closed")
	}
	close(fv.signalChannel)
}

// ########### Future ###########

type Future[T any] struct {
	isComplete    bool
	lock          sync.Mutex
	signalChannel chan FutureResult[T]
	hasWaiter     bool
}

type FutureResult[T any] struct {
	Value    T
	HasValue bool
}

func Future_Make[T any]() *Future[T] {
	return &Future[T]{
		signalChannel: make(chan FutureResult[T], 1),
	}
}

func (future *Future[T]) SetResult(value T) {
	future.lock.Lock()
	if !future.isComplete {
		future.isComplete = true
		future.signalChannel <- FutureResult[T]{Value: value, HasValue: true}
	}
	future.lock.Unlock()
}

func (future *Future[T]) SetResultEmpty() {
	future.lock.Lock()
	if !future.isComplete {
		future.isComplete = true
		future.signalChannel <- FutureResult[T]{HasValue: false}
	}
	future.lock.Unlock()
}

func (future *Future[T]) WaitForResult() (T, bool) {
	if future == nil || future.signalChannel == nil {
		panic("invalid future")
	} else if future.hasWaiter {
		panic("duplicate waiter")
	}
	future.hasWaiter = true

	result, channelOk := <-future.signalChannel
	if !channelOk {
		panic("signal channel closed")
	}
	close(future.signalChannel)

	return result.Value, result.HasValue
}

// ########### FutureCancellable ###########

type FutureCancellable[T any] struct {
	isComplete          bool
	isCancelled         bool
	lock                sync.Mutex
	resultChannel       chan FutureResult[T]
	cancelSignalChannel chan any
	hasWaiter           bool
	onCancel            []func()
}

func FutureCancellable_Make[T any]() *FutureCancellable[T] {
	return &FutureCancellable[T]{
		resultChannel:       make(chan FutureResult[T], 1),
		cancelSignalChannel: make(chan any),
	}
}

func (future *FutureCancellable[T]) SetResult(value T) {
	future.lock.Lock()
	if !future.isComplete {
		future.isComplete = true
		future.onCancel = nil
		future.resultChannel <- FutureResult[T]{Value: value, HasValue: true}
	}
	future.lock.Unlock()
}

func (future *FutureCancellable[T]) SetResultEmpty() {
	future.lock.Lock()
	if !future.isComplete {
		future.isComplete = true
		future.onCancel = nil
		future.resultChannel <- FutureResult[T]{HasValue: false}
	}
	future.lock.Unlock()
}

func (future *FutureCancellable[T]) AddCancelHandler(onCancel func()) {
	future.lock.Lock()
	if future.isCancelled {
		onCancel()
	} else if !future.isComplete {
		future.onCancel = append(future.onCancel, onCancel)
	}
	future.lock.Unlock()
}

func (future *FutureCancellable[T]) Cancel() {
	future.lock.Lock()
	if !future.isComplete {
		future.performCancel()
		future.resultChannel <- FutureResult[T]{HasValue: false}
	}
	future.lock.Unlock()
}

func (future *FutureCancellable[T]) performCancel() {
	future.isComplete = true
	future.isCancelled = true
	for i := range future.onCancel {
		future.onCancel[i]()
	}
	future.onCancel = nil
	close(future.cancelSignalChannel)
}

func (future *FutureCancellable[T]) CancelSignalChannel() <-chan any {
	return future.cancelSignalChannel
}

func (future *FutureCancellable[T]) ShouldContinue() bool {
	return !future.isComplete
}

func (future *FutureCancellable[T]) ShouldFinish() bool {
	return future.isComplete
}

func (future *FutureCancellable[T]) verifyCanWait() {
	if future == nil || future.resultChannel == nil {
		panic("invalid future")
	} else if future.hasWaiter {
		panic("duplicate waiter")
	}
	future.hasWaiter = true
}

func (future *FutureCancellable[T]) resultFromChannel() (T, bool) {
	result, channelOk := <-future.resultChannel
	if !channelOk {
		panic("signal channel closed")
	}
	close(future.resultChannel)
	return result.Value, result.HasValue
}

func (future *FutureCancellable[T]) WaitForResult() (T, bool) {
	future.verifyCanWait()
	return future.resultFromChannel()
}

func (future *FutureCancellable[T]) WaitForResultOrPanic() T {
	future.verifyCanWait()
	value, hasValue := future.resultFromChannel()
	if !hasValue {
		panic("expected valid result")
	}
	return value
}

func (future *FutureCancellable[T]) WaitForResultAsOptional() util.Optional[T] {
	future.verifyCanWait()
	value, hasValue := future.resultFromChannel()
	if hasValue {
		return util.Optional_OfValue(value)
	} else {
		return util.Optional_Empty[T]()
	}
}

func (future *FutureCancellable[T]) WaitForResultThenRun(onSuccess func(T), onFail func()) {
	future.verifyCanWait()
	value, hasValue := future.resultFromChannel()
	if hasValue && onSuccess != nil {
		onSuccess(value)
	} else if !hasValue && onFail != nil {
		onFail()
	}
}

func (future *FutureCancellable[T]) WaitForResultOrKeyPress() (T, bool) {
	future.verifyCanWait()

	channelForKey := future.channelKeyPress()

	select {
	case result, channelOk := <-future.resultChannel:
		if !channelOk {
			panic("signal channel closed")
		}
		close(future.resultChannel)
		return result.Value, result.HasValue

	case <-channelForKey:
		var nilValue T

		future.lock.Lock()
		defer future.lock.Unlock()

		// if we're here then no-one is listening for signal anymore, but someone could still have beaten us to lock
		if future.isCancelled {
			close(future.resultChannel)
			return nilValue, false
		} else if future.isComplete {
			result, channelOk := <-future.resultChannel
			if !channelOk {
				panic("signal channel closed")
			}
			close(future.resultChannel)
			return result.Value, result.HasValue
		}

		_, err := os.Stdout.WriteString("Cancelling on key press\n")
		if err != nil {
			panic(err)
		}

		future.performCancel()
		return nilValue, false
	}
}

func (future *FutureCancellable[T]) ForwardSuccessfulResultToChannel(resultChannel chan<- T) {
	future.verifyCanWait()
	go func() {
		value, hasValue := future.resultFromChannel()
		if hasValue {
			resultChannel <- value
		}
	}()
}

func (future *FutureCancellable[T]) ForwardAnyResultToChannel(resultChannel chan<- FutureResult[T]) {
	future.verifyCanWait()
	go func() {
		result, channelOk := <-future.resultChannel
		if !channelOk {
			panic("signal channel closed")
		}
		close(future.resultChannel)
		resultChannel <- result
	}()
}

func (*FutureCancellable[T]) channelKeyPress() chan any {
	channelForKey := make(chan any)
	go func() {
		_, err := os.Stdin.Read([]byte{0})
		if err != nil {
			panic(err)
		}
		channelForKey <- true
	}()
	return channelForKey
}

func FutureCancellable_MapValue[T any, R any](innerFuture *FutureCancellable[T], mapper func(T) (R, bool)) *FutureCancellable[R] {
	outerFuture := FutureCancellable_Make[R]()
	ChainCancel(outerFuture, innerFuture)

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

func FutureCancellable_MapToFuture[T any, R any](innerFuture *FutureCancellable[T], mapper func(T) *FutureCancellable[R]) *FutureCancellable[R] {
	outerFuture := FutureCancellable_Make[R]()
	ChainCancel(outerFuture, innerFuture)

	go func() {
		innerValue, innerHasValue := innerFuture.WaitForResult()
		if innerHasValue {
			middleFuture := mapper(innerValue)
			if middleFuture != nil {
				ChainCancel(outerFuture, middleFuture)
				middleValue, middleHasValue := middleFuture.WaitForResult()
				if middleHasValue {
					outerFuture.SetResult(middleValue)
					return
				}
			}
		}
		outerFuture.SetResultEmpty()
	}()

	return outerFuture
}
