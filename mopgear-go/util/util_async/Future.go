package util_async

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
)

type futureResult[T any] struct {
	Value    T
	HasValue bool
}

type valueOrError[T any] struct {
	Value *T
	Error error
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
	defer fv.lock.Unlock()
	if !fv.isComplete {
		fv.signalChannel <- true
		fv.performComplete()
	}
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

func (fv *FutureVoid) performComplete() {
	for i := range fv.onComplete {
		fv.onComplete[i]()
	}

	fv.isComplete = true
	fv.onComplete = nil
}

// ########### futureCommon ###########

type futureCommon[T any] struct {
	isComplete    bool
	lock          sync.Mutex
	hasWaiter     bool
	resultChannel chan futureResult[T]
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
	defer future.lock.Unlock()
	if !future.isComplete {
		future.resultChannel <- futureResult[T]{Value: value, HasValue: true}
		future.isComplete = true
	}
}

func (future *Future[T]) SetResultEmpty() {
	future.lock.Lock()
	defer future.lock.Unlock()
	if !future.isComplete {
		future.resultChannel <- futureResult[T]{HasValue: false}
		future.isComplete = true
	}
}

func (future *Future[T]) MapSameType(mapper func(T) (T, bool)) *Future[T] {
	return Future_MapValue(future, mapper)
}

// ########### FutureWithError ###########
type FutureWithError[T any] struct {
	FutureCancellable[valueOrError[T]]
}

func FutureWithError_Make[T any]() *FutureWithError[T] {
	return &FutureWithError[T]{
		*FutureCancellable_Make[valueOrError[T]](),
	}
}

// ########### FutureCancellable ###########

type FutureCancellable[T any] struct {
	futureCommon[T]
	isCancelled         bool
	cancelSignalChannel chan struct{}
	onCancel            []func() error
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
	defer future.lock.Unlock()
	if !future.isComplete {
		future.onCancel = nil
		future.resultChannel <- futureResult[T]{Value: value, HasValue: true}
		future.isComplete = true
	}
}

func (future *FutureCancellable[T]) SetResultEmpty() {
	future.lock.Lock()
	defer future.lock.Unlock()
	if !future.isComplete {
		future.onCancel = nil
		future.resultChannel <- futureResult[T]{HasValue: false}
		future.isComplete = true
	}
}

func (future *FutureCancellable[T]) AddCancelHandler(onCancel func() error) error {
	future.lock.Lock()
	defer future.lock.Unlock()
	if future.isCancelled {
		return onCancel()
	} else if !future.isComplete {
		future.onCancel = append(future.onCancel, onCancel)
	}
	return nil
}

func (future *FutureCancellable[T]) Cancel() error {
	future.lock.Lock()
	defer future.lock.Unlock()
	if !future.isComplete {
		future.resultChannel <- futureResult[T]{HasValue: false}
		return future.performCancelAndComplete()
	}
	return nil
}

func (future *FutureCancellable[T]) CancelOrPanic() {
	err := future.Cancel()
	util.GlobalFatalErrorHandler(err)
}

func (future *FutureCancellable[T]) performCancelAndComplete() error {
	future.isComplete = true

	future.isCancelled = true
	close(future.cancelSignalChannel)

	var resultError error
	for i := range future.onCancel {
		err := future.onCancel[i]()
		resultError = errors.Join(resultError, err)
	}
	future.onCancel = nil

	return resultError
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

func (future *FutureCancellable[T]) WaitForResultOrKeyPress() (T, bool, error) {
	future.verifyCanWait()

	channelForKey := future.channelKeyPress()

	select {
	case result, channelOk := <-future.resultChannel:
		if !channelOk {
			panic("signal channel closed")
		}
		close(future.resultChannel)
		return result.Value, result.HasValue, nil

	case <-channelForKey:
		var nilValue T

		future.lock.Lock()
		defer future.lock.Unlock()

		// if we're here then no-one is listening for signal anymore, but someone could still have beaten us to lock
		if future.isCancelled {
			close(future.resultChannel)
			return nilValue, false, nil
		} else if future.isComplete {
			result, channelOk := <-future.resultChannel
			if !channelOk {
				panic("signal channel closed")
			}
			close(future.resultChannel)
			return result.Value, result.HasValue, nil
		}

		_, err1 := os.Stdout.WriteString("Cancelling on key press\n")

		err2 := future.performCancelAndComplete()

		return nilValue, false, errors.Join(err1, err2)
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

func (future *FutureCancellable[T]) MapSameType(mapper func(T) T) (*FutureCancellable[T], error) {
	return FutureCancellable_MapValue(future, mapper)
}

func FutureCancellable_MapValue[T any, R any](innerFuture *FutureCancellable[T], mapper func(T) R) (*FutureCancellable[R], error) {
	outerFuture := FutureCancellable_Make[R]()
	err1 := ChainCancel(outerFuture, innerFuture)

	go func() {
		value, hasValue := innerFuture.WaitForResult()
		if hasValue {
			newValue := mapper(value)
			outerFuture.SetResult(newValue)
		} else {
			outerFuture.SetResultEmpty()
		}
	}()

	return outerFuture, err1
}

func FutureCancellable_MapValueOptional[T any, R any](innerFuture *FutureCancellable[T], mapper func(T) (R, bool)) (*FutureCancellable[R], error) {
	outerFuture := FutureCancellable_Make[R]()
	err1 := ChainCancel(outerFuture, innerFuture)

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

	return outerFuture, err1
}

func FutureCancellable_MapToFuture[T any, R any](innerFuture *FutureCancellable[T], mapper func(T) *FutureCancellable[R]) (*FutureCancellable[R], error) {
	outerFuture := FutureCancellable_Make[R]()
	errMain := ChainCancel(outerFuture, innerFuture)

	go func() {
		innerValue, innerHasValue := innerFuture.WaitForResult()
		if innerHasValue {
			middleFuture := mapper(innerValue)
			if middleFuture != nil {
				middleFuture.ForwardResultToOtherFuture(outerFuture)
				return
			}
		}
		outerFuture.SetResultEmpty()
	}()

	return outerFuture, errMain
}

// ########### FutureCancellableWithError ###########

type FutureCancellableWithError[T any] struct {
	*FutureCancellable[valueOrError[T]]
}

var _ IFutureWithResultAndCancelAndError[int] = &FutureCancellableWithError[int]{}

func FutureCancellableWithError_Make[T any]() *FutureCancellableWithError[T] {
	return &FutureCancellableWithError[T]{
		FutureCancellable_Make[valueOrError[T]](),
	}
}

func (future *FutureCancellableWithError[T]) SetResult(value T) {
	future.lock.Lock()
	defer future.lock.Unlock()
	if !future.isComplete {
		future.onCancel = nil
		future.resultChannel <- futureResult[valueOrError[T]]{
			Value:    valueOrError[T]{Value: new(value)},
			HasValue: true,
		}
		future.isComplete = true
	}
}

func (future *FutureCancellableWithError[T]) SetResultError(err error) {
	future.lock.Lock()
	defer future.lock.Unlock()
	if !future.isComplete {
		future.onCancel = nil
		future.resultChannel <- futureResult[valueOrError[T]]{
			Value:    valueOrError[T]{Error: err},
			HasValue: true,
		}
		future.isComplete = true
	}
}

func (future *FutureCancellableWithError[T]) finalError(value valueOrError[T]) error {
	if value.Error != nil {
		return value.Error
	} else if future.isCancelled {
		return errors.New("cancelled")
	} else {
		return errors.New("empty result")
	}
}

func (future *FutureCancellableWithError[T]) WaitForResultOrError() (T, error) {
	future.verifyCanWait()
	value, hasValue := future.resultFromChannel()
	if hasValue && value.Error == nil && value.Value != nil {
		return *value.Value, nil
	} else {
		return makeNilValue[T](), future.finalError(value)
	}
}

func (future *FutureCancellableWithError[T]) WaitForResultPointerOrError() (*T, error) {
	future.verifyCanWait()
	value, hasValue := future.resultFromChannel()
	if hasValue && value.Error == nil && value.Value != nil {
		return value.Value, nil
	} else {
		return nil, future.finalError(value)
	}
}

func (future *FutureCancellableWithError[T]) WaitForResultOrPanic() T {
	future.verifyCanWait()
	value, hasValue := future.resultFromChannel()
	if hasValue && value.Error == nil && value.Value != nil {
		return *value.Value
	} else {
		panic(future.finalError(value))
	}
}

//goland:noinspection GoDfaErrorMayBeNotNil
func (future *FutureCancellableWithError[T]) WaitForResultOrKeyPress() (T, bool, error) {
	value, hasValue, err := future.FutureCancellable.WaitForResultOrKeyPress()
	if hasValue && err == nil && value.Error == nil && value.Value != nil {
		return *value.Value, true, nil
	} else if value.Error != nil || err != nil {
		return makeNilValue[T](), false, errors.Join(err, value.Error)
	} else {
		return makeNilValue[T](), false, nil
	}
}

func (future *FutureCancellableWithError[T]) ForwardResultToOtherFuture(other IFutureWithResultOrError[T]) {
	future.verifyCanWait()
	go func() {
		value, hasValue := future.resultFromChannel()

		if hasValue && value.Error == nil && value.Value != nil {
			other.SetResult(*value.Value)
		} else if value.Error != nil {
			other.SetResultError(value.Error)
		} else {
			other.SetResultEmpty()
		}
	}()
}

func FutureCancellable_MapValueError[T any, R any](innerFuture *FutureCancellable[T], mapper func(T) (*R, error)) *FutureCancellableWithError[R] {
	outerFuture := FutureCancellableWithError_Make[R]()
	if err1 := ChainCancel(outerFuture, innerFuture); err1 != nil {
		outerFuture.SetResultError(err1)
	} else {
		go func() {
			value, hasValue := innerFuture.WaitForResult()

			if hasValue {
				newValue, errMapped := mapper(value)
				if errMapped == nil {
					outerFuture.SetResult(*newValue)
				} else {
					outerFuture.SetResultError(errMapped)
				}
			} else {
				outerFuture.SetResultEmpty()
			}
		}()
	}

	return outerFuture
}

func (future *FutureCancellableWithError[T]) ForwardResultOrErrorToCallback(onSuccess func(T), onFail func(error)) {
	future.verifyCanWait()
	go func() {
		value, hasValue := future.resultFromChannel()
		if hasValue && value.Error == nil && value.Value != nil {
			onSuccess(*value.Value)
		} else {
			onFail(value.Error)
		}
	}()
}

func makeNilValue[T any]() T {
	var nilValue T
	return nilValue
}
