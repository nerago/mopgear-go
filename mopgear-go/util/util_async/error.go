package util_async

import (
	"errors"
	"fmt"
	"sync"

	"github.com/nerago/mopgear-go/util"
)

type PossibleFutureError struct {
	errorValue    error
	signalChannel chan error
	lock          sync.Mutex
	isComplete    bool
	hasWaiter     bool
}

var _ IFutureErrorMinimal = &PossibleFutureError{}

func PossibleFutureErrorMake() *PossibleFutureError {
	return &PossibleFutureError{
		signalChannel: make(chan error, 1),
	}
}

func (pf *PossibleFutureError) SetResultSuccess() {
	pf.lock.Lock()
	defer pf.lock.Unlock()
	if !pf.isComplete {
		pf.signalChannel <- nil
		pf.isComplete = true
	}
}

func (pf *PossibleFutureError) SetResultError(err error) {
	util.GlobalWarnHandler(err)
	
	pf.lock.Lock()
	defer pf.lock.Unlock()
	if !pf.isComplete {
		pf.errorValue = err
		pf.signalChannel <- err
		pf.isComplete = true
	} else if !pf.hasWaiter {
		pf.errorValue = errors.Join(pf.errorValue, err)
	} else {
		util.GlobalFatalErrorHandler(fmt.Errorf("PossibleFutureError additional error after Wait; %w", err))
	}
}

func (pf *PossibleFutureError) HasError() bool {
	pf.lock.Lock()
	defer pf.lock.Unlock()
	return pf.isComplete && pf.errorValue != nil
}

func (pf *PossibleFutureError) verifyCanWaitNoLock() error {
	if pf == nil || pf.signalChannel == nil {
		return errors.New("invalid future")
	} else if pf.hasWaiter {
		return errors.New("duplicate waiter")
	}
	pf.hasWaiter = true
	return nil
}

func (pf *PossibleFutureError) verifyCanWaitLocked() error {
	pf.lock.Lock()
	defer pf.lock.Unlock()
	return pf.verifyCanWaitNoLock()
}

func (pf *PossibleFutureError) GetResultNoWait() (error, bool) {
	pf.lock.Lock()
	defer pf.lock.Unlock()

	err := pf.verifyCanWaitNoLock()
	if err != nil {
		return err, true
	}

	if pf.isComplete {
		return pf.errorValue, pf.errorValue != nil
	} else {
		return nil, false
	}
}

func (pf *PossibleFutureError) WaitForResult() error {
	err := pf.verifyCanWaitLocked()
	if err != nil {
		return err
	}

	result, channelOk := <-pf.signalChannel
	if !channelOk {
		return errors.New("signal channel closed")
	}
	close(pf.signalChannel)

	return result
}

func (pf *PossibleFutureError) ForwardErrorToOtherFuture(other IFutureErrorMinimal) {
	err := pf.verifyCanWaitLocked()
	if err != nil {
		other.SetResultError(err)
		return
	}

	go func() {
		result, channelOk := <-pf.signalChannel
		if !channelOk {
			other.SetResultError(errors.New("signal channel closed"))
			return
		}
		close(pf.signalChannel)

		if result != nil {
			other.SetResultError(result)
		}
	}()
}
