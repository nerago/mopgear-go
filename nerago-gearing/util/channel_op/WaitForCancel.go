package channel_op

import (
	"os"
	"sync"
)

// ########### Future ###########
type Future[T any] struct {
	signalChannel chan any
	result        T
	isValidResult bool
	isComplete    bool
	lock          sync.Mutex
}

func Future_Make[T any]() *Future[T] {
	return &Future[T]{
		signalChannel: make(chan any, 1),
	}
}

func (future *Future[T]) SetResult(value T) {
	future.lock.Lock()
	defer future.lock.Unlock()

	if !future.isComplete {
		future.result = value
		future.isValidResult = true
		future.isComplete = true
		future.signalChannel <- true
	}
}

func (future *Future[T]) SetEmpty() {
	future.lock.Lock()
	defer future.lock.Unlock()

	if !future.isComplete {
		future.isValidResult = false
		future.isComplete = true
		future.signalChannel <- true
	}
}

func (future *Future[T]) WaitForResult() (T, bool) {
	<-future.signalChannel
	return future.result, future.isValidResult
}

func (future *Future[T]) WaitForResultOrPanic() T {
	<-future.signalChannel
	if !future.isValidResult {
		panic("no result")
	}
	return future.result
}

// ########### FutureCancellable ###########
type FutureCancellable[T any] struct {
	signalChannel chan any
	result        T
	isValidResult bool
	isComplete    bool
	onCancel      func()
	lock          sync.Mutex
}

func FutureCancellable_Make[T any](onCancel func()) *FutureCancellable[T] {
	if onCancel == nil {
		panic("onCancel required")
	}
	return &FutureCancellable[T]{
		signalChannel: make(chan any, 1),
		onCancel:      onCancel,
	}
}

func (future *FutureCancellable[T]) SetResult(value T) {
	future.lock.Lock()
	defer future.lock.Unlock()

	if !future.isComplete {
		future.result = value
		future.isValidResult = true
		future.isComplete = true
		future.signalChannel <- true
	}
}

func (future *FutureCancellable[T]) SetEmpty() {
	future.lock.Lock()
	defer future.lock.Unlock()

	if !future.isComplete {
		future.isValidResult = false
		future.isComplete = true
		future.signalChannel <- true
	}
}

func (future *FutureCancellable[T]) Cancel(value T) {
	future.lock.Lock()
	defer future.lock.Unlock()

	if !future.isComplete {
		future.isValidResult = false
		future.isComplete = true
		future.onCancel()
		future.signalChannel <- true
	}
}

func (future *FutureCancellable[T]) WaitForResult() (T, bool) {
	<-future.signalChannel
	return future.result, future.isValidResult
}

func (future *FutureCancellable[T]) WaitForResultOrKeyPress() (T, bool) {
	channelForKey := future.channelKeyPress()

	select {
	case <-future.signalChannel:
		return future.result, future.isValidResult

	case <-channelForKey:
		future.lock.Lock()
		defer future.lock.Unlock()

		// if we're here then noone is listening for signal anymore, but someone could still have beaten us to lock
		if future.isComplete {
			return future.result, future.isValidResult
		} else {
			future.isValidResult = false
			future.isComplete = true
			future.onCancel()
			return future.result, false
		}
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
