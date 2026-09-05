package util_async

import (
	"errors"
	"fmt"
	"sync"

	"github.com/nerago/mopgear-go/util"
)

type PossibleFutureErrors struct {
	lock              sync.Mutex
	errorSlice        []error
	forwardTo         IFutureErrorMinimal
	hasReturnedResult bool
	isForwarding      bool
}

var _ IFutureErrorMinimal = &PossibleFutureErrors{}

func PossibleFutureErrorMake() *PossibleFutureErrors {
	return &PossibleFutureErrors{}
}

func (pf *PossibleFutureErrors) AddError(err error) {
	pf.lock.Lock()
	defer pf.lock.Unlock()
	if err == nil {
		// no error
	} else if pf.hasReturnedResult {
		util.GlobalWarnHandler(fmt.Errorf("PossibleFutureError additional error after GetResult; %w", err))
	} else if pf.isForwarding {
		pf.forwardTo.AddError(err)
	} else {
		pf.errorSlice = append(pf.errorSlice, err)
	}
}

func (pf *PossibleFutureErrors) HasError() bool {
	pf.lock.Lock()
	defer pf.lock.Unlock()
	return len(pf.errorSlice) > 0
}

func (pf *PossibleFutureErrors) GetResultNoWait() (err error, hasErrors bool) {
	pf.lock.Lock()
	defer pf.lock.Unlock()

	if pf.forwardTo != nil {
		panic("inconsistent result destination")
	}
	// flag return but don't actually block multiple returns; pretty harmless
	pf.hasReturnedResult = true

	if len(pf.errorSlice) > 0 {
		err = errors.Join(pf.errorSlice...)
		if err != nil {
			return err, true
		}
	}

	return nil, false
}

func (pf *PossibleFutureErrors) ForwardErrorToOtherFuture(other IFutureErrorMinimal) {
	pf.lock.Lock()
	defer pf.lock.Unlock()

	if pf.hasReturnedResult || pf.forwardTo != nil {
		panic("inconsistent result destination")
	}

	for _, err := range pf.errorSlice {
		other.AddError(err)
	}
	pf.errorSlice = nil

	pf.forwardTo = other
}
