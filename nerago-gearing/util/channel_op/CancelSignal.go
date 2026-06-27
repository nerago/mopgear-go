package channel_op

import "sync"

type CancelSignal interface {
	AddCancelHandler(func())
	Cancel()
	ShouldContinue() bool
	ShouldFinish() bool
}

func ChainCancel(outer, inner CancelSignal) {
	outer.AddCancelHandler(inner.Cancel)
}

type CancelSignalBasic struct {
	isCancelled bool
	lock        sync.Mutex
	onCancel    []func()
}

func CancelSignal_Make() CancelSignal {
	return &CancelSignalBasic{}
}

func (cancel *CancelSignalBasic) AddCancelHandler(onCancel func()) {
	cancel.lock.Lock()
	if cancel.isCancelled {
		onCancel()
	} else {
		cancel.onCancel = append(cancel.onCancel, onCancel)
	}
	cancel.lock.Unlock()
}

func (cancel *CancelSignalBasic) Cancel() {
	cancel.lock.Lock()
	if !cancel.isCancelled {
		cancel.isCancelled = true
		for i := range cancel.onCancel {
			cancel.onCancel[i]()
		}
		cancel.onCancel = nil
	}
	cancel.lock.Unlock()
}

func (cancel *CancelSignalBasic) ShouldContinue() bool {
	return !cancel.isCancelled
}

func (cancel *CancelSignalBasic) ShouldFinish() bool {
	return cancel.isCancelled
}
