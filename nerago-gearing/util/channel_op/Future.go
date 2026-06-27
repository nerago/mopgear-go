package channel_op

import (
	"fmt"
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

func (future *Future[T]) SetResultEmpty() {
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
	signalChannel chan futureResult[T]
	isComplete    bool
	isCancelled   bool
	hasWaiter     bool
	onCancel      []func()
	lock          sync.Mutex
}

type futureResult[T any] struct {
	value    T
	hasValue bool
}

func FutureCancellable_Make[T any]() *FutureCancellable[T] {
	sc := make(chan futureResult[T], 1)
	return &FutureCancellable[T]{
		signalChannel: sc,
	}
}

func (future *FutureCancellable[T]) SetResult(value T) {
	future.lock.Lock()
	if !future.isComplete {
		future.isComplete = true
		future.onCancel = nil
		future.lock.Unlock()
		future.signalChannel <- futureResult[T]{value: value, hasValue: true}
	} else {
		future.lock.Unlock()
	}
}

func (future *FutureCancellable[T]) SetResultEmpty() {
	future.lock.Lock()
	if !future.isComplete {
		future.isComplete = true
		future.onCancel = nil
		future.lock.Unlock()
		future.signalChannel <- futureResult[T]{hasValue: false}
	} else {
		future.lock.Unlock()
	}
}

func (future *FutureCancellable[T]) AddCancelHandler(onCancel func()) {
	future.lock.Lock()

	if future.isCancelled {
		future.lock.Unlock()
		onCancel()
	} else {
		future.onCancel = append(future.onCancel, onCancel)
		future.lock.Unlock()
	}
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
		future.lock.Unlock()
		future.signalChannel <- futureResult[T]{hasValue: false}
	} else {
		future.lock.Unlock()
	}
}

// IsCancelled also includes normal completion, intended for use in loop conditions etc
func (future *FutureCancellable[T]) IsCancelled() bool {
	return future.isComplete || future.isCancelled
}

func (future *FutureCancellable[T]) WaitForResult() (T, bool) {
	if future == nil || future.signalChannel == nil {
		panic("invalid future")
	} else if future.hasWaiter {
		panic("duplicate waiter")
	}
	future.hasWaiter = true

	os.Stdout.WriteString("WaitForResult before\n")
	result, channelOk := <-future.signalChannel
	if !channelOk {
		panic("signal channel closed")
	}
	close(future.signalChannel)

	fmt.Printf("WaitForResult after %v %v\n", result.value, result.hasValue)
	return result.value, result.hasValue
}

//func (future *FutureCancellable[T]) WaitForResultOrKeyPress() (T, bool) {
//	if future == nil || future.signalChannel == nil {
//		panic("bad future")
//	}
//
//	channelForKey := future.channelKeyPress()
//
//	select {
//	case <-future.signalChannel:
//		return future.result, future.isValidResult
//
//	case <-channelForKey:
//		future.lock.Lock()
//		defer future.lock.Unlock()
//
//		// if we're here then noone is listening for signal anymore, but someone could still have beaten us to lock
//		if future.isComplete {
//			return future.result, future.isValidResult
//		} else {
//			_, err := os.Stdout.WriteString("Cancelling on key press\n")
//			if err != nil {
//				panic(err)
//			}
//
//			future.isValidResult = false
//			future.isComplete = true
//			for i := range future.onCancel {
//				future.onCancel[i]()
//			}
//			future.onCancel = nil
//			return future.result, false
//		}
//	}
//}

//func (*FutureCancellable[T]) channelKeyPress() chan any {
//	channelForKey := make(chan any)
//	go func() {
//		_, err := os.Stdin.Read([]byte{0})
//		if err != nil {
//			panic(err)
//		}
//		channelForKey <- true
//	}()
//	return channelForKey
//}
