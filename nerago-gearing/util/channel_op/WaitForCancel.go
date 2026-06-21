package channel_op

import (
	"os"
	"paladin_gearing_go/util"
	"sync"
)

type Waitable struct {
	completionChannel chan any
	onCancel          func()

	waitGroup *sync.WaitGroup
}

type WaitableWithResult[T any] struct {
	Waitable
	resultChannel <-chan T
	result        T
}

func Waitable_ForWaitGroup(waitGroup *sync.WaitGroup) *Waitable {
	waitable := new(Waitable)
	waitable.completionChannel = make(chan any)
	go func() {
		waitGroup.Wait()
		waitable.completionChannel <- true
	}()
	return waitable
}

func WaitableWithResult_ForWaitGroup[T any](waitGroup *sync.WaitGroup, resultChannel <-chan T) *WaitableWithResult[T] {
	waitable := new(WaitableWithResult[T])
	waitable.resultChannel = resultChannel
	go func() {
		waitGroup.Wait()
		waitable.result = <-resultChannel
		waitable.completionChannel <- true
	}()
	return waitable
}

func WaitableWithResult_ForChannel[T any](resultChannel <-chan T) *WaitableWithResult[T] {
	waitable := new(WaitableWithResult[T])
	waitable.resultChannel = resultChannel
	go func() {
		waitable.result = <-resultChannel
		waitable.completionChannel <- true
	}()
	return waitable
}

func WaitableWithResult_SendSupply[T any]() (*WaitableWithResult[T], func(T)) {
	waitable := new(WaitableWithResult[T])
	supplyFunc := func(result T) {
		waitable.result = result
		waitable.completionChannel <- true
	}
	return waitable, supplyFunc
}

func (wait *Waitable) AddCancelAction(onCancel func()) {
	wait.onCancel = onCancel
}

func (wait *Waitable) AddCancelTracker(track *util.TrackProgress) {
	wait.onCancel = track.CancelAll
}

func (wait *Waitable) WaitCompletionOrKeyPress() {
	channelForKey := wait.channelKeyPress()

	select {
	case <-wait.completionChannel:
	case <-channelForKey:
	}
}

func (wait *WaitableWithResult[T]) WaitCompletionOrKeyPress() (T, bool) {

	channelForKey := wait.channelKeyPress()

	select {
	case <-wait.completionChannel:
		return wait.result, true
	case <-channelForKey:
		return wait.result, false
	}
}

func (*Waitable) channelKeyPress() chan bool {
	channelForKey := make(chan bool)
	go func() {
		bytes := make([]byte, 1)
		os.Stdin.Read(bytes)
		channelForKey <- true
	}()
	return channelForKey
}
