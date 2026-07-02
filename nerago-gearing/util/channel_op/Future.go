package channel_op

import (
	"os"
	"paladin_gearing_go/util"
	"sync"
)

// ########### Future ###########

type Future[T any] struct {
	isComplete    bool
	lock          sync.Mutex
	signalChannel chan futureResult[T]
	hasWaiter     bool
}

type futureResult[T any] struct {
	value    T
	hasValue bool
}

func Future_Make[T any]() *Future[T] {
	return &Future[T]{
		signalChannel: make(chan futureResult[T], 1),
	}
}

func (future *Future[T]) SetResult(value T) {
	future.lock.Lock()
	if !future.isComplete {
		future.isComplete = true
		future.signalChannel <- futureResult[T]{value: value, hasValue: true}
	}
	future.lock.Unlock()
}

func (future *Future[T]) SetResultEmpty() {
	future.lock.Lock()
	if !future.isComplete {
		future.isComplete = true
		future.signalChannel <- futureResult[T]{hasValue: false}
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

	return result.value, result.hasValue
}

// ########### FutureCancellable ###########

type FutureCancellable[T any] struct {
	isComplete    bool
	isCancelled   bool
	lock          sync.Mutex
	signalChannel chan futureResult[T]
	hasWaiter     bool
	onCancel      []func()
}

func FutureCancellable_Make[T any]() *FutureCancellable[T] {
	return &FutureCancellable[T]{
		signalChannel: make(chan futureResult[T], 1),
	}
}

func (future *FutureCancellable[T]) SetResult(value T) {
	future.lock.Lock()
	if !future.isComplete {
		future.isComplete = true
		future.onCancel = nil
		future.signalChannel <- futureResult[T]{value: value, hasValue: true}
	}
	future.lock.Unlock()
}

func (future *FutureCancellable[T]) SetResultEmpty() {
	future.lock.Lock()
	if !future.isComplete {
		future.isComplete = true
		future.onCancel = nil
		future.signalChannel <- futureResult[T]{hasValue: false}
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
		future.isComplete = true
		future.isCancelled = true
		for i := range future.onCancel {
			future.onCancel[i]()
		}
		future.onCancel = nil
		future.signalChannel <- futureResult[T]{hasValue: false}
	}
	future.lock.Unlock()
}

func (future *FutureCancellable[T]) ShouldContinue() bool {
	return !future.isComplete
}

func (future *FutureCancellable[T]) ShouldFinish() bool {
	return future.isComplete
}

func (future *FutureCancellable[T]) WaitForResult() (T, bool) {
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

	return result.value, result.hasValue
}

func (future *FutureCancellable[T]) WaitForResult_AsOptional() util.Optional[T] {
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

	if result.hasValue {
		return util.Optional_OfValue(result.value)
	} else {
		return util.Optional_Empty[T]()
	}
}

func (future *FutureCancellable[T]) WaitForResultOrKeyPress() (T, bool) {
	if future == nil || future.signalChannel == nil {
		panic("invalid future")
	} else if future.hasWaiter {
		panic("duplicate waiter")
	}
	future.hasWaiter = true

	channelForKey := future.channelKeyPress()

	select {
	case result, channelOk := <-future.signalChannel:
		if !channelOk {
			panic("signal channel closed")
		}
		close(future.signalChannel)
		return result.value, result.hasValue

	case <-channelForKey:
		var nilValue T

		future.lock.Lock()
		defer future.lock.Unlock()

		// if we're here then no-one is listening for signal anymore, but someone could still have beaten us to lock
		if future.isCancelled {
			close(future.signalChannel)
			return nilValue, false
		} else if future.isComplete {
			result, channelOk := <-future.signalChannel
			if !channelOk {
				panic("signal channel closed")
			}
			close(future.signalChannel)
			return result.value, result.hasValue
		}

		_, err := os.Stdout.WriteString("Cancelling on key press\n")
		if err != nil {
			panic(err)
		}

		future.isComplete = true
		future.isCancelled = true
		for i := range future.onCancel {
			future.onCancel[i]()
		}
		future.onCancel = nil
		return nilValue, false
	}
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

func FutureCancellable_Map[T any, R any](innerFuture *FutureCancellable[T], mapper func(T) (R, bool)) *FutureCancellable[R] {
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
