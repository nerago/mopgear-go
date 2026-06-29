package channel_op

import (
	"paladin_gearing_go/util"
	"sync"
	"testing"
	"time"
)

func TestFlowSetResult(t *testing.T) {
	waitGroup := sync.WaitGroup{}
	future := FutureCancellable_Make[int]()
	calledCancelHandler := false
	future.AddCancelHandler(func() {
		calledCancelHandler = true
	})
	waitGroup.Go(func() {
		timer := util.StopwatchMakeStarted()
		value, ok := future.WaitForResult()
		if !ok {
			t.Error("bad ok")
		}
		if value != 123 {
			t.Error("bad value")
		}
		taken := timer.Elapsed()
		if taken < 2*time.Second {
			t.Error("finished too fast")
		}
	})
	waitGroup.Go(func() {
		time.Sleep(2 * time.Second)
		future.SetResult(123)
	})
	waitGroup.Wait()
	if calledCancelHandler {
		t.Error("called cancel handler")
	}
}

func TestFlowSetResultLateWaiter(t *testing.T) {
	waitGroup := sync.WaitGroup{}
	future := FutureCancellable_Make[int]()
	calledCancelHandler := false
	future.AddCancelHandler(func() {
		calledCancelHandler = true
	})
	waitGroup.Go(func() {
		time.Sleep(2 * time.Second)
		value, ok := future.WaitForResult()
		if !ok {
			t.Error("bad ok")
		}
		if value != 123 {
			t.Error("bad value")
		}
	})
	waitGroup.Go(func() {
		time.Sleep(1 * time.Second)
		future.SetResult(123)
	})
	waitGroup.Wait()
	if calledCancelHandler {
		t.Error("called cancel handler")
	}
}

func TestFlowSetEmpty(t *testing.T) {
	waitGroup := sync.WaitGroup{}
	future := FutureCancellable_Make[int]()
	calledCancelHandler := false
	future.AddCancelHandler(func() {
		calledCancelHandler = true
	})
	waitGroup.Go(func() {
		timer := util.StopwatchMakeStarted()
		value, ok := future.WaitForResult()
		if ok {
			t.Error("bad ok")
		}
		if value != 0 {
			t.Error("bad value")
		}
		taken := timer.Elapsed()
		if taken < 2*time.Second {
			t.Error("finished too fast")
		}
	})
	waitGroup.Go(func() {
		time.Sleep(2 * time.Second)
		future.SetResultEmpty()
	})
	waitGroup.Wait()
	if calledCancelHandler {
		t.Error("called cancel handler")
	}
}

func TestFlowSetEmptyLateWaiter(t *testing.T) {
	waitGroup := sync.WaitGroup{}
	future := FutureCancellable_Make[int]()
	calledCancelHandler := false
	future.AddCancelHandler(func() {
		calledCancelHandler = true
	})
	waitGroup.Go(func() {
		time.Sleep(2 * time.Second)
		value, ok := future.WaitForResult()
		if ok {
			t.Error("bad ok")
		}
		if value != 0 {
			t.Error("bad value")
		}
	})
	waitGroup.Go(func() {
		time.Sleep(1 * time.Second)
		future.SetResultEmpty()
	})
	waitGroup.Wait()
	if calledCancelHandler {
		t.Error("called cancel handler")
	}
}

func TestFlowCancel(t *testing.T) {
	waitGroup := sync.WaitGroup{}
	future := FutureCancellable_Make[int]()
	calledCancelHandler := false
	future.AddCancelHandler(func() {
		calledCancelHandler = true
	})
	waitGroup.Go(func() {
		timer := util.StopwatchMakeStarted()
		value, ok := future.WaitForResult()
		if ok {
			t.Error("bad ok")
		}
		if value != 0 {
			t.Error("bad value")
		}
		taken := timer.Elapsed()
		if taken < 2*time.Second {
			t.Error("finished too fast")
		}
	})
	waitGroup.Go(func() {
		time.Sleep(2 * time.Second)
		future.Cancel()
	})
	waitGroup.Wait()
	if !calledCancelHandler {
		t.Error("didn't call cancel handler")
	}
}

func TestFlowCancelLateWaiter(t *testing.T) {
	waitGroup := sync.WaitGroup{}
	future := FutureCancellable_Make[int]()
	calledCancelHandler := false
	future.AddCancelHandler(func() {
		calledCancelHandler = true
	})
	waitGroup.Go(func() {
		time.Sleep(2 * time.Second)
		value, ok := future.WaitForResult()
		if ok {
			t.Error("bad ok")
		}
		if value != 0 {
			t.Error("bad value")
		}
	})
	waitGroup.Go(func() {
		time.Sleep(1 * time.Second)
		future.Cancel()
	})
	waitGroup.Wait()
	if !calledCancelHandler {
		t.Error("didn't call cancel handler")
	}
}

func TestFlowCancelThenDone(t *testing.T) {
	waitGroup := sync.WaitGroup{}
	future := FutureCancellable_Make[int]()
	calledCancelHandler := false
	future.AddCancelHandler(func() {
		calledCancelHandler = true
	})
	waitGroup.Go(func() {
		timer := util.StopwatchMakeStarted()
		value, ok := future.WaitForResult()
		if ok {
			t.Error("bad ok")
		}
		if value != 0 {
			t.Error("bad value")
		}
		taken := timer.Elapsed()
		if taken < 1*time.Second {
			t.Error("finished too fast")
		}
	})
	waitGroup.Go(func() {
		time.Sleep(2 * time.Second)
		future.SetResult(123)
	})
	waitGroup.Go(func() {
		time.Sleep(1 * time.Second)
		future.Cancel()
	})
	waitGroup.Wait()
	if !calledCancelHandler {
		t.Error("didn't call cancel handler")
	}
}

func TestFlowCancelThenDoneLateWaiter(t *testing.T) {
	waitGroup := sync.WaitGroup{}
	future := FutureCancellable_Make[int]()
	calledCancelHandler := false
	future.AddCancelHandler(func() {
		calledCancelHandler = true
	})
	waitGroup.Go(func() {
		time.Sleep(3 * time.Second)
		value, ok := future.WaitForResult()
		if ok {
			t.Error("bad ok")
		}
		if value != 0 {
			t.Error("bad value")
		}
	})
	waitGroup.Go(func() {
		time.Sleep(2 * time.Second)
		future.SetResult(123)
	})
	waitGroup.Go(func() {
		time.Sleep(1 * time.Second)
		future.Cancel()
	})
	waitGroup.Wait()
	if !calledCancelHandler {
		t.Error("didn't call cancel handler")
	}
}

func TestFlowDoneThenCancel(t *testing.T) {
	waitGroup := sync.WaitGroup{}
	future := FutureCancellable_Make[int]()
	calledCancelHandler := false
	future.AddCancelHandler(func() {
		calledCancelHandler = true
	})
	waitGroup.Go(func() {
		timer := util.StopwatchMakeStarted()
		value, ok := future.WaitForResult()
		if !ok {
			t.Error("bad ok")
		}
		if value != 123 {
			t.Error("bad value")
		}
		taken := timer.Elapsed()
		if taken < 1*time.Second {
			t.Error("finished too fast")
		}
	})
	waitGroup.Go(func() {
		time.Sleep(1 * time.Second)
		future.SetResult(123)
	})
	waitGroup.Go(func() {
		time.Sleep(2 * time.Second)
		future.Cancel()
	})
	waitGroup.Wait()
	if calledCancelHandler {
		t.Error("called cancel handler")
	}
}

func TestFlowDoneThenCancelLaterWaiter(t *testing.T) {
	waitGroup := sync.WaitGroup{}
	future := FutureCancellable_Make[int]()
	calledCancelHandler := false
	future.AddCancelHandler(func() {
		calledCancelHandler = true
	})
	waitGroup.Go(func() {
		time.Sleep(3 * time.Second)
		value, ok := future.WaitForResult()
		if !ok {
			t.Error("bad ok")
		}
		if value != 123 {
			t.Error("bad value")
		}
	})
	waitGroup.Go(func() {
		time.Sleep(1 * time.Second)
		future.SetResult(123)
	})
	waitGroup.Go(func() {
		time.Sleep(2 * time.Second)
		future.Cancel()
	})
	waitGroup.Wait()
	if calledCancelHandler {
		t.Error("called cancel handler")
	}
}
