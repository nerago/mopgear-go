package channel_op

import "sync"

type CancelSignal interface {
	AddCancelHandler(func())
	Cancel()
	IsCancelled() bool
}

func ChainCancel(outer, inner CancelSignal) {
	outer.AddCancelHandler(inner.Cancel)
}

type CancelSignalBasic struct {
	isCancelled bool
	onCancel    []func()
	lock        sync.Mutex
}

func CancelSignal_Make() CancelSignal {
	return &CancelSignalBasic{}
}

func (cancel *CancelSignalBasic) AddCancelHandler(onCancel func()) {
	cancel.lock.Lock()
	defer cancel.lock.Unlock()

	if cancel.isCancelled {
		onCancel()
	} else {
		cancel.onCancel = append(cancel.onCancel, onCancel)
	}
}

func (cancel *CancelSignalBasic) Cancel() {
	cancel.lock.Lock()
	defer cancel.lock.Unlock()

	if !cancel.isCancelled {
		for i := range cancel.onCancel {
			cancel.onCancel[i]()
		}
		cancel.onCancel = nil
		cancel.isCancelled = true
	}
}

func (cancel *CancelSignalBasic) IsCancelled() bool {
	return cancel.isCancelled
}
