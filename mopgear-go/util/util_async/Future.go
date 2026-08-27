package util_async

import (
	"os"
	"sync"
	"time"

	"github.com/nerago/mopgear-go/util/util_collection"
)

type IFuture interface {
	SetResultEmpty()
	AddCompletedHandler(onComplete func())
}

type IFutureWithResult[T any] interface {
	IFuture
	SetResult(value T)

	WaitForResult() (T, bool)
	WaitForResultOrPanic() T
	WaitForResultOrNilValue() T
	WaitForResultAsOptional() util_collection.Optional[T]

	ForwardSuccessfulResultToCallback(apply func(T))
	ForwardResultToRelevantCallback(onSuccess func(T), onFail func())
	ForwardResultToOtherFuture(other IFutureWithResult[T])
	ForwardSuccessfulResultToChannel(resultChannel chan<- T)
	ForwardAnyResultToChannel(resultChannel chan<- futureResult[T])
	//MapSameType(mapper func(T) (T, bool)) *Future[T]
}

type IFutureWithCancel interface {
	IFuture
	CancelSignal
}

type IFutureWithResultAndCancel[T any] interface {
	IFutureWithCancel
	IFutureWithResult[T]
	WaitForResultOrKeyPress() (T, bool)
}

type futureResult[T any] struct {
	Value    T
	HasValue bool
}

// ########### FutureVoid ###########

type FutureVoid struct {
	signalChannel chan any
	onComplete    []func()
	lock          sync.Mutex
	isComplete    bool
	hasWaiter     bool
}

var _ IFuture = &FutureVoid{}

func FutureVoid_Make() *FutureVoid {
	return &FutureVoid{
		signalChannel: make(chan any, 1),
	}
}

func (fv *FutureVoid) SetResultEmpty() {
	fv.lock.Lock()
	if !fv.isComplete {
		fv.performComplete()
		fv.signalChannel <- true
	}
	fv.lock.Unlock()
}

func (fv *FutureVoid) WaitForComplete() {
	if fv == nil || fv.signalChannel == nil {
		panic("invalid FutureVoid")
	}

	fv.lock.Lock()
	if fv.hasWaiter {
		panic("duplicate waiter")
	}
	fv.hasWaiter = true
	fv.lock.Unlock()

	_, channelOk := <-fv.signalChannel
	if !channelOk {
		panic("signal channel closed")
	}
	close(fv.signalChannel)
}

func (fv *FutureVoid) WaitForLimitedDuration(duration time.Duration) (timeout bool) {
	if fv == nil || fv.signalChannel == nil {
		panic("invalid FutureVoid")
	}

	fv.lock.Lock()
	if fv.hasWaiter {
		panic("duplicate waiter")
	}
	fv.hasWaiter = true
	fv.lock.Unlock()

	select {
	case _, channelOk := <-fv.signalChannel:
		if !channelOk {
			panic("signal channel closed")
		}
		close(fv.signalChannel)
		return false

	case <-time.After(duration):
		fv.lock.Lock()
		fv.hasWaiter = false
		fv.lock.Unlock()
		return true
	}
}

func (fv *FutureVoid) AddCompletedHandler(onComplete func()) {
	fv.lock.Lock()
	if fv.isComplete {
		onComplete()
	} else {
		fv.onComplete = append(fv.onComplete, onComplete)
	}
	fv.lock.Unlock()
}

func (fv *FutureVoid) performComplete() {
	fv.isComplete = true
	for i := range fv.onComplete {
		fv.onComplete[i]()
	}
	fv.onComplete = nil
}

// ########### futureCommon ###########

type futureCommon[T any] struct {
	isComplete    bool
	lock          sync.Mutex
	hasWaiter     bool
	resultChannel chan futureResult[T]
	onComplete    []func()
}

func (future *futureCommon[T]) AddCompletedHandler(onComplete func()) {
	future.lock.Lock()
	if future.isComplete {
		onComplete()
	} else {
		future.onComplete = append(future.onComplete, onComplete)
	}
	future.lock.Unlock()
}

func (future *futureCommon[T]) performComplete() {
	future.isComplete = true
	for i := range future.onComplete {
		future.onComplete[i]()
	}
	future.onComplete = nil
}

func (future *futureCommon[T]) verifyCanWait() {
	if future == nil || future.resultChannel == nil {
		panic("invalid future")
	} else if future.hasWaiter {
		panic("duplicate waiter")
	}
	future.hasWaiter = true
}

func (future *futureCommon[T]) resultFromChannel() (T, bool) {
	result, channelOk := <-future.resultChannel
	if !channelOk {
		panic("signal channel closed")
	}
	close(future.resultChannel)
	return result.Value, result.HasValue
}

func (future *futureCommon[T]) WaitForResult() (T, bool) {
	future.verifyCanWait()
	return future.resultFromChannel()
}

func (future *futureCommon[T]) WaitForResultOrPanic() T {
	future.verifyCanWait()
	value, hasValue := future.resultFromChannel()
	if !hasValue {
		panic("expected valid result")
	}
	return value
}

func (future *futureCommon[T]) WaitForResultOrNilValue() T {
	future.verifyCanWait()
	value, hasValue := future.resultFromChannel()
	if hasValue {
		return value
	} else {
		var nilValue T
		return nilValue
	}
}

func (future *futureCommon[T]) WaitForResultAsOptional() util_collection.Optional[T] {
	future.verifyCanWait()
	value, hasValue := future.resultFromChannel()
	if hasValue {
		return util_collection.Optional_OfValue(value)
	} else {
		return util_collection.Optional_Empty[T]()
	}
}

func (future *futureCommon[T]) GetResultNoWait() (T, bool) {
	future.verifyCanWait()
	select {
	case result, channelOk := <-future.resultChannel:
		if !channelOk {
			panic("signal channel closed")
		}
		close(future.resultChannel)
		return result.Value, result.HasValue
	default:
		var nilValue T
		return nilValue, false
	}
}

func (future *futureCommon[T]) ForwardResultToOtherFuture(other IFutureWithResult[T]) {
	future.verifyCanWait()
	go func() {
		value, hasValue := future.resultFromChannel()
		if hasValue {
			other.SetResult(value)
		} else {
			other.SetResultEmpty()
		}
	}()
}

func (future *futureCommon[T]) ForwardSuccessfulResultToChannel(resultChannel chan<- T) {
	future.verifyCanWait()
	go func() {
		value, hasValue := future.resultFromChannel()
		if hasValue {
			resultChannel <- value
		}
	}()
}

func (future *futureCommon[T]) ForwardAnyResultToChannel(resultChannel chan<- futureResult[T]) {
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

func (future *futureCommon[T]) ForwardSuccessfulResultToCallback(apply func(T)) {
	future.verifyCanWait()
	go func() {
		value, hasValue := future.resultFromChannel()
		if hasValue {
			apply(value)
		}
	}()
}

func (future *futureCommon[T]) ForwardResultToRelevantCallback(onSuccess func(T), onFail func()) {
	future.verifyCanWait()
	go func() {
		value, hasValue := future.resultFromChannel()
		if hasValue {
			onSuccess(value)
		} else {
			onFail()
		}
	}()
}

// ########### Future ###########

type Future[T any] struct {
	futureCommon[T]
}

var _ IFutureWithResult[int] = &Future[int]{}

func Future_Make[T any]() *Future[T] {
	return &Future[T]{
		resultChannel: make(chan futureResult[T], 1),
	}
}

func (future *Future[T]) SetResult(value T) {
	future.lock.Lock()
	if !future.isComplete {
		future.performComplete()
		future.resultChannel <- futureResult[T]{Value: value, HasValue: true}
	}
	future.lock.Unlock()
}

func (future *Future[T]) SetResultEmpty() {
	future.lock.Lock()
	if !future.isComplete {
		future.performComplete()
		future.resultChannel <- futureResult[T]{HasValue: false}
	}
	future.lock.Unlock()
}

func (future *Future[T]) MapSameType(mapper func(T) (T, bool)) *Future[T] {
	return Future_MapValue(future, mapper)
}

// ########### FutureCancellable ###########

type FutureCancellable[T any] struct {
	futureCommon[T]
	isCancelled         bool
	cancelSignalChannel chan struct{}
	onCancel            []func()
}

var _ IFutureWithResultAndCancel[int] = &FutureCancellable[int]{}

func FutureCancellable_Make[T any]() *FutureCancellable[T] {
	return &FutureCancellable[T]{
		resultChannel:       make(chan futureResult[T], 1),
		cancelSignalChannel: make(chan struct{}),
	}
}

func (future *FutureCancellable[T]) SetResult(value T) {
	future.lock.Lock()
	if !future.isComplete {
		future.performComplete()
		future.onCancel = nil
		future.resultChannel <- futureResult[T]{Value: value, HasValue: true}
	}
	future.lock.Unlock()
}

func (future *FutureCancellable[T]) SetResultEmpty() {
	future.lock.Lock()
	if !future.isComplete {
		future.performComplete()
		future.onCancel = nil
		future.resultChannel <- futureResult[T]{HasValue: false}
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
		future.performCancelAndComplete()
		future.resultChannel <- futureResult[T]{HasValue: false}
	}
	future.lock.Unlock()
}

func (future *FutureCancellable[T]) performCancelAndComplete() {
	future.performComplete()
	future.isCancelled = true
	for i := range future.onCancel {
		future.onCancel[i]()
	}
	future.onCancel = nil
	close(future.cancelSignalChannel)
}

func (future *FutureCancellable[T]) CancelSignalChannel() <-chan struct{} {
	return future.cancelSignalChannel
}

func (future *FutureCancellable[T]) ShouldContinue() bool {
	return !future.isComplete
}

func (future *FutureCancellable[T]) ShouldFinish() bool {
	return future.isComplete
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

		future.performCancelAndComplete()
		return nilValue, false
	}
}

func (*FutureCancellable[T]) channelKeyPress() chan any {
	channelForKey := make(chan any)
	go func() {
		waitForKeyPress()
		channelForKey <- true
	}()
	return channelForKey
}

func (future *FutureCancellable[T]) MapSameType(mapper func(T) (T, bool)) *FutureCancellable[T] {
	return FutureCancellable_MapValue(future, mapper)
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

type ValueOrError[T any] struct {
	Value T
	Error error
}

type FutureCancellableWithError[T any] struct {
	*FutureCancellable[ValueOrError[T]]
}

func FutureCancellable_MapValueError[T any, R any](innerFuture *FutureCancellable[T], mapper func(T) (R, error)) *FutureCancellableWithError[R] {
	outerFuture := FutureCancellable_Make[ValueOrError[R]]()
	ChainCancel(outerFuture, innerFuture)

	go func() {
		value, hasValue := innerFuture.WaitForResult()
		if hasValue {
			newValue, err := mapper(value)
			if err == nil {
				outerFuture.SetResult(ValueOrError[R]{Value: newValue})
			} else {
				outerFuture.SetResult(ValueOrError[R]{Error: err})
			}
		} else {
			outerFuture.SetResultEmpty()
		}
	}()

	return &FutureCancellableWithError[R]{outerFuture}
}
