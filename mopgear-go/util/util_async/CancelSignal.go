package util_async

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/nerago/mopgear-go/util"
)

type CancelSignal interface {
	AddCancelHandler(func() error) error
	Cancel() error
	ShouldContinue() bool
	ShouldFinish() bool
	CancelSignalChannel() <-chan struct{}
}

func ChainCancel(outer, inner CancelSignal) error {
	return outer.AddCancelHandler(inner.Cancel)
}

type CancelSignalBasic struct {
	isCancelled   bool
	lock          sync.Mutex
	onCancel      []func() error
	signalChannel chan struct{}
}

func CancelSignal_Make() *CancelSignalBasic {
	return &CancelSignalBasic{
		signalChannel: make(chan struct{}),
	}
}

func (cancel *CancelSignalBasic) AddCancelHandler(onCancel func() error) error {
	cancel.lock.Lock()
	defer cancel.lock.Unlock()
	if cancel.isCancelled {
		return onCancel()
	} else {
		cancel.onCancel = append(cancel.onCancel, onCancel)
		return nil
	}
}

func (cancel *CancelSignalBasic) Cancel() error {
	cancel.lock.Lock()
	defer cancel.lock.Unlock()

	var resultError error
	if !cancel.isCancelled {
		cancel.isCancelled = true
		for i := range cancel.onCancel {
			err := cancel.onCancel[i]()
			resultError = errors.Join(resultError, err)
		}
		cancel.onCancel = nil
		close(cancel.signalChannel)
	}
	return resultError
}

func (cancel *CancelSignalBasic) ShouldContinue() bool {
	return !cancel.isCancelled
}

func (cancel *CancelSignalBasic) ShouldFinish() bool {
	return cancel.isCancelled
}

func (cancel *CancelSignalBasic) CancelSignalChannel() <-chan struct{} {
	return cancel.signalChannel
}

func CancelOnKeyPress(cancel CancelSignal) {
	go func() {
		waitForKeyPress()
		err := cancel.Cancel()
		if err != nil {
			util.GlobalErrorHandler(err)
		}
	}()
}

func waitForKeyPress() {
	_, err := os.Stdin.Read([]byte{0})
	if err != nil {
		panic(err)
	}
}

func CancelAfterTimeout(cancel CancelSignal, timeout time.Duration, printer *util.PrintRecorder) *time.Timer {
	return time.AfterFunc(timeout, func() {
		printer.Println("###################### TIME LIMIT EXPIRED ######################")
		err := cancel.Cancel()
		util.GlobalErrorHandler(err)
	})
}

type CancelSignalContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func CancelSignalContextMake(ctx context.Context, cancel context.CancelFunc) CancelSignalContext {
	return CancelSignalContext{ctx, cancel}
}

func CancelSignalContextMakeFromParent(parentContext context.Context) CancelSignalContext {
	ctx, cancel := context.WithCancel(parentContext)
	return CancelSignalContext{ctx, cancel}
}

func (c CancelSignalContext) AddCancelHandler(onCancel func()) {
	context.AfterFunc(c.ctx, onCancel)
}

func (c CancelSignalContext) Cancel() {
	c.cancel()
}

func (c CancelSignalContext) ShouldContinue() bool {
	select {
	case <-c.ctx.Done():
		return false
	default:
		return true
	}
}

func (c CancelSignalContext) ShouldFinish() bool {
	select {
	case <-c.ctx.Done():
		return true
	default:
		return false
	}
}

func (c CancelSignalContext) CancelSignalChannel() <-chan struct{} {
	return c.ctx.Done()
}
