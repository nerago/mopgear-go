package util_async

import "github.com/nerago/mopgear-go/util/util_collection"

type IFuture interface {
	SetResultEmpty()
}

type IFutureWithResult[T any] interface {
	IFuture
	SetResult(value T)

	GetResultNoWait() (T, bool)
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

type IFutureErrorMinimal interface {
	SetResultSuccess()
	AddResultError(err error)

	HasError() bool
	GetResultNoWait() (error, bool)
	WaitForResult() error

	ForwardErrorToOtherFuture(other IFutureErrorMinimal)
}

type IFutureWithResultOrError[T any] interface {
	IFuture
	SetResult(value T)
	SetResultError(err error)

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
