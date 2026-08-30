package util_async

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
)

type IFuture interface {
	SetResultEmpty() error
	AddCompletedHandler(onComplete func() error) error
}

type IFutureWithResult[T any] interface {
	IFuture
	SetResult(value T) error

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

type IFutureWithResultOrError[T any] interface {
	IFuture
	SetResult(value T) error
	SetResultError(err error) error

	WaitForResultOrError() (T, error)
	WaitForResultOrPanic() T

	ForwardResultOrErrorToCallback(onSuccess func(T), onFail func(error))
	ForwardResultToOtherFuture(other IFutureWithResultOrError[T])
}

type IFutureWithResultAndCancel[T any] interface {
	IFutureWithCancel
	IFutureWithResult[T]
	WaitForResultOrKeyPress() (T, bool, error)
}

type IFutureWithResultAndCancelAndError[T any] interface {
	IFutureWithCancel
	IFutureWithResultOrError[T]
	WaitForResultOrKeyPress() (T, bool, error)
}

type futureResult[T any] struct {
	Value    T
	HasValue bool
}

// ########### FutureVoid ###########

type FutureVoid struct {
	signalChannel chan any
	onComplete    []func() error
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

func (fv *FutureVoid) SetResultEmpty() error {
	fv.lock.Lock()
	defer fv.lock.Unlock()
	if !fv.isComplete {
		fv.signalChannel <- true
		return fv.performComplete()
	}
	return nil
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

func (fv *FutureVoid) AddCompletedHandler(onComplete func() error) error {
	fv.lock.Lock()
	defer fv.lock.Unlock()
	if fv.isComplete {
		return onComplete()
	} else {
		fv.onComplete = append(fv.onComplete, onComplete)
		return nil
	}
}

func (fv *FutureVoid) performComplete() error {
	var resultError error
	for i := range fv.onComplete {
		err := fv.onComplete[i]()
		resultError = errors.Join(resultError, err)
	}

	fv.isComplete = true
	fv.onComplete = nil

	return resultError
}

// ########### futureCommon ###########

type futureCommon[T any] struct {
	isComplete    bool
	lock          sync.Mutex
	hasWaiter     bool
	resultChannel chan futureResult[T]
	onComplete    []func() error
}

func (future *futureCommon[T]) AddCompletedHandler(onComplete func() error) error {
	future.lock.Lock()
	defer future.lock.Unlock()
	if future.isComplete {
		return onComplete()
	} else {
		future.onComplete = append(future.onComplete, onComplete)
		return nil
	}
}

func (future *futureCommon[T]) performComplete() error {
	future.isComplete = true

	var resultError error
	for i := range future.onComplete {
		err := future.onComplete[i]()
		resultError = errors.Join(resultError, err)
	}
	future.onComplete = nil

	return resultError
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

		var err error
		if hasValue {
			err = other.SetResult(value)
		} else {
			err = other.SetResultEmpty()
		}
		util.GlobalErrorHandler(err)
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

func (future *Future[T]) SetResult(value T) error {
	future.lock.Lock()
	defer future.lock.Unlock()
	if !future.isComplete {
		future.resultChannel <- futureResult[T]{Value: value, HasValue: true}
		return future.performComplete()
	}
	return nil
}

func (future *Future[T]) SetResultEmpty() error {
	future.lock.Lock()
	defer future.lock.Unlock()
	if !future.isComplete {
		future.resultChannel <- futureResult[T]{HasValue: false}
		return future.performComplete()
	}
	return nil
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

func (future *FutureCancellable[T]) SetResult(value T) error {
	future.lock.Lock()
	defer future.lock.Unlock()
	if !future.isComplete {
		future.onCancel = nil
		future.resultChannel <- futureResult[T]{Value: value, HasValue: true}
		return future.performComplete()
	}
	return nil
}

func (future *FutureCancellable[T]) SetResultEmpty() error {
	future.lock.Lock()
	defer future.lock.Unlock()
	if !future.isComplete {
		future.onCancel = nil
		future.resultChannel <- futureResult[T]{HasValue: false}
		return future.performComplete()
	}
	return nil
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

func (future *FutureCancellable[T]) performCancelAndComplete() error {
	resultError := future.performComplete()

	future.isCancelled = true
	close(future.cancelSignalChannel)

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
		var err2 error
		if hasValue {
			newValue := mapper(value)
			err2 = outerFuture.SetResult(newValue)
		} else {
			err2 = outerFuture.SetResultEmpty()
		}
		util.GlobalErrorHandler(err2)
	}()

	return outerFuture, err1
}

func FutureCancellable_MapValueOptional[T any, R any](innerFuture *FutureCancellable[T], mapper func(T) (R, bool)) (*FutureCancellable[R], error) {
	outerFuture := FutureCancellable_Make[R]()
	err1 := ChainCancel(outerFuture, innerFuture)

	go func() {
		value, hasValue := innerFuture.WaitForResult()
		var err2 error
		if hasValue {
			newValue, hasNew := mapper(value)
			if hasNew {
				err2 = outerFuture.SetResult(newValue)
			} else {
				err2 = outerFuture.SetResultEmpty()
			}
		} else {
			err2 = outerFuture.SetResultEmpty()
		}
		util.GlobalErrorHandler(err2)
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
		err := outerFuture.SetResultEmpty()
		util.GlobalErrorHandler(err)
	}()

	return outerFuture, errMain
}

// ########### FutureCancellableWithError ###########

type valueOrError[T any] struct {
	Value *T
	Error error
}

type FutureCancellableWithError[T any] struct {
	*FutureCancellable[valueOrError[T]]
}

var _ IFutureWithResultAndCancelAndError[int] = &FutureCancellableWithError[int]{}

func FutureCancellableWithError_Make[T any]() *FutureCancellableWithError[T] {
	return &FutureCancellableWithError[T]{
		FutureCancellable_Make[valueOrError[T]](),
	}
}

func (future *FutureCancellableWithError[T]) SetResult(value T) error {
	future.lock.Lock()
	defer future.lock.Unlock()
	if !future.isComplete {
		future.onCancel = nil
		future.resultChannel <- futureResult[valueOrError[T]]{
			Value:    valueOrError[T]{Value: new(value)},
			HasValue: true,
		}
		return future.performComplete()
	}
	return nil
}

func (future *FutureCancellableWithError[T]) SetResultError(err error) error {
	future.lock.Lock()
	defer future.lock.Unlock()
	if !future.isComplete {
		future.onCancel = nil
		future.resultChannel <- futureResult[valueOrError[T]]{
			Value:    valueOrError[T]{Error: err},
			HasValue: true,
		}
		return future.performComplete()
	}
	return nil
}

func (future *FutureCancellableWithError[T]) WaitForResultOrError() (T, error) {
	future.verifyCanWait()
	value, hasValue := future.resultFromChannel()
	if hasValue && value.Error == nil && value.Value != nil {
		return *value.Value, nil
	} else if value.Error != nil {
		return makeNilValue[T](), value.Error
	} else {
		return makeNilValue[T](), errors.New("empty result")
	}
}

func (future *FutureCancellableWithError[T]) WaitForResultPointerOrError() (*T, error) {
	future.verifyCanWait()
	value, hasValue := future.resultFromChannel()
	if hasValue && value.Error == nil && value.Value != nil {
		return value.Value, nil
	} else if value.Error != nil {
		return nil, value.Error
	} else {
		return nil, errors.New("empty result")
	}
}

func (future *FutureCancellableWithError[T]) WaitForResultOrPanic() T {
	future.verifyCanWait()
	value, hasValue := future.resultFromChannel()
	if hasValue && value.Error == nil && value.Value != nil {
		return *value.Value
	} else if value.Error != nil {
		panic(value.Error)
	} else {
		panic("empty result")
	}
}

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

		var err error
		if hasValue && value.Error == nil && value.Value != nil {
			err = other.SetResult(*value.Value)
		} else if value.Error != nil {
			err = other.SetResultError(value.Error)
		} else {
			err = other.SetResultEmpty()
		}
		util.GlobalErrorHandler(err)
	}()
}

func FutureCancellable_MapValueError[T any, R any](innerFuture *FutureCancellable[T], mapper func(T) (*R, error)) *FutureCancellableWithError[R] {
	outerFuture := FutureCancellableWithError_Make[R]()
	if err1 := ChainCancel(outerFuture, innerFuture); err1 != nil {
		err2 := outerFuture.SetResultError(err1)
		if err2 != nil {
			util.GlobalErrorHandler(errors.Join(err1, err2))
		}
	} else {
		go func() {
			value, hasValue := innerFuture.WaitForResult()

			var errInner error
			if hasValue {
				newValue, errMapped := mapper(value)
				if errMapped == nil {
					errInner = outerFuture.SetResult(*newValue)
				} else {
					errInner = outerFuture.SetResultError(errMapped)
				}
			} else {
				errInner = outerFuture.SetResultEmpty()
			}
			util.GlobalErrorHandler(errInner)
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
