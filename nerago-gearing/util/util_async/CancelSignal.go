package util_async

import (
	"context"
	"os"
	"paladin_gearing_go/util"
	"sync"
	"time"
)

type CancelSignal interface {
	AddCancelHandler(func())
	Cancel()
	ShouldContinue() bool
	ShouldFinish() bool
	CancelSignalChannel() <-chan struct{}
}

func ChainCancel(outer, inner CancelSignal) {
	outer.AddCancelHandler(inner.Cancel)
}

type CancelSignalBasic struct {
	isCancelled   bool
	lock          sync.Mutex
	onCancel      []func()
	signalChannel chan struct{}
}

func CancelSignal_Make() CancelSignal {
	return &CancelSignalBasic{
		signalChannel: make(chan struct{}),
	}
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
		close(cancel.signalChannel)
	}
	cancel.lock.Unlock()
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
		cancel.Cancel()
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
		cancel.Cancel()
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
