package util_async

import (
	"errors"
	"math/rand"
	"slices"
	"sync"

	"github.com/nerago/mopgear-go/util/util_collection"
)

type NestedTaskPoolParent struct {
	mutex    sync.Mutex
	children []*NestedTaskPoolChild
	started  bool
	stopped  bool

	chanTaskAdded  chan *internalTask
	chanThreadExit chan any
	chanPoolStop   chan any
}

func (pool *NestedTaskPoolParent) NewChild() *NestedTaskPoolChild {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	child := &NestedTaskPoolChild{pool: pool, waitChannel: make(chan bool)}
	pool.children = append(pool.children, child)
	return child
}

func (pool *NestedTaskPoolParent) taskAdded(task *internalTask) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if pool.stopped {
		panic("stopped, not accepting new tasks")
	}
	if pool.started {
		pool.chanTaskAdded <- task
	}
}

func (pool *NestedTaskPoolParent) Start(threads int) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if pool.started {
		panic("already started")
	}

	pool.chanTaskAdded = make(chan *internalTask)
	pool.chanThreadExit = make(chan any)
	pool.chanPoolStop = make(chan any)
	go pool.controlThread(threads)

	pool.started = true
}

func (pool *NestedTaskPoolParent) Stop() {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if !pool.started {
		panic("never started")
	} else if pool.stopped {
		panic("already stopped")
	}

	pool.stopped = true
	close(pool.chanPoolStop)
}

func (pool *NestedTaskPoolParent) countQueuedTasks() int {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	count := 0
	for _, child := range pool.children {
		child.mutex.Lock()
		count += len(child.queue)
		child.mutex.Unlock()
	}
	return count
}

//func (pool *NestedTaskPoolParent) selectQueuedTask() *internalTask {
//	pool.mutex.Lock()
//	defer pool.mutex.Unlock()
//
//	taskList := make([]*internalTask, 0)
//	for _, child := range pool.children {
//		if child.mutex.TryLock() {
//			// we deliberately are choosing to leave it to end of function, after we do the notifyStarting
//			//goland:noinspection GoDeferInLoop
//			defer child.mutex.Unlock()
//			taskList = append(taskList, child.queue...)
//		}
//	}
//
//	if len(taskList) > 0 {
//		task := taskList[rand.Intn(len(taskList))]
//		task.owner.notifyStartingAlreadyLocked(task)
//		return task
//	} else {
//		return nil
//	}
//}

func (pool *NestedTaskPoolParent) selectQueuedTask() *internalTask {
	for {
		pool.mutex.Lock()
		childSliceClone := slices.Clone(pool.children)
		pool.mutex.Unlock()

		taskList := make([]*internalTask, 0)
		for _, child := range childSliceClone {
			// this version we're released our parent lock so should be no possible deadlock
			child.mutex.Lock()
			taskList = append(taskList, child.queue...)
			child.mutex.Unlock()
		}

		if len(taskList) == 0 {
			return nil
		}

		task := taskList[rand.Intn(len(taskList))]
		if task.owner.requestStartTaskNeedLock(task) {
			return task
		}
	}
}

func (pool *NestedTaskPoolParent) controlThread(maxThreads int) {
	currentThreads := max(maxThreads, pool.countQueuedTasks())
	for range currentThreads {
		go pool.workerThread()
	}

	stopped := pool.stopped
	for !stopped || currentThreads > 0 {
		select {
		case <-pool.chanTaskAdded:
			if currentThreads < maxThreads && !stopped {
				go pool.workerThread()
				currentThreads++
			}
		case <-pool.chanThreadExit:
			currentThreads--

			// possible with resource contention might have ended thread unnecessarily
			countQueued := pool.countQueuedTasks()
			for currentThreads < countQueued {
				go pool.workerThread()
				currentThreads++
			}
		case <-pool.chanPoolStop:
			stopped = true
		}
	}
}

func (pool *NestedTaskPoolParent) workerThread() {
	for !pool.stopped {
		task := pool.selectQueuedTask()
		if task != nil {
			err := task.run()
			task.owner.notifyTaskComplete(task, err)
		} else {
			break
		}
	}
	pool.chanThreadExit <- true
}

type internalTask struct {
	owner *NestedTaskPoolChild
	run   func() error
}

type NestedTaskPoolChild struct {
	mutex           sync.Mutex
	pool            *NestedTaskPoolParent
	queue           []*internalTask
	running         []*internalTask
	waitChannel     chan bool
	continueOnError bool
	errors          []error
}

func (child *NestedTaskPoolChild) SetContinueOnError(continueOnError bool) {
	child.mutex.Lock()
	defer child.mutex.Unlock()
	child.continueOnError = continueOnError
}

func (child *NestedTaskPoolChild) Go(run func() error) {
	child.mutex.Lock()
	defer child.mutex.Unlock()

	if len(child.errors) > 0 && !child.continueOnError {
		child.errors = append(child.errors, errors.New("attempted to enqueue another routine after error"))
		return
	}

	task := &internalTask{child, run}
	child.queue = append(child.queue, task)
	child.pool.taskAdded(task)
}

func (child *NestedTaskPoolChild) requestStartTaskNeedLock(task *internalTask) bool {
	child.mutex.Lock()
	defer child.mutex.Unlock()
	if removeFromSliceIfExists(&child.queue, task) {
		child.running = append(child.running, task)
		return true
	}
	return false
}

func (child *NestedTaskPoolChild) notifyStartingAlreadyLocked(task *internalTask) error {
	err := removeFromSlice(&child.queue, task)
	child.running = append(child.running, task)
	return err
}

func (child *NestedTaskPoolChild) notifyTaskComplete(task *internalTask, err error) {
	child.mutex.Lock()
	defer child.mutex.Unlock()

	removeFromSlice(&child.running, task)

	if err != nil {
		child.errors = append(child.errors, err)
	}

	if err != nil && !child.continueOnError {
		child.queue = nil
	}

	if len(child.queue) == 0 && len(child.running) == 0 && child.waitChannel != nil {
		close(child.waitChannel)
		child.waitChannel = nil
	}
}

// only clears wait if at least one child has been added and finished
//func (child *NestedTaskPoolChild) WaitAllCompleteAtLeastOne() error {
//	waitChan := child.waitChannel
//	if waitChan != nil {
//		<-waitChan
//	}
//
//	child.mutex.Lock()
//	defer child.mutex.Unlock()
//	return child.makeErrorResult()
//}

// can complete immediately if nothing has been queued
func (child *NestedTaskPoolChild) WaitAllComplete() error {
	if initialEmpty, err := child.checkQueueEmptyAndMakeError(); initialEmpty {
		return err
	}

	waitChan := child.waitChannel
	if waitChan != nil {
		<-waitChan
	}

	child.mutex.Lock()
	defer child.mutex.Unlock()
	return child.makeErrorResult()
}

func (child *NestedTaskPoolChild) checkQueueEmptyAndMakeError() (bool, error) {
	child.mutex.Lock()
	defer child.mutex.Unlock()
	if len(child.queue) == 0 && len(child.running) == 0 {
		return true, child.makeErrorResult()
	} else {
		return false, nil
	}
}

func (child *NestedTaskPoolChild) makeErrorResult() error {
	if len(child.errors) > 0 {
		return errors.Join(child.errors...)
	} else {
		return nil
	}
}

func removeFromSlice(slice *[]*internalTask, task *internalTask) error {
	index := slices.Index(*slice, task)
	if index == -1 {
		return errors.New("not in slice")
	}
	util_collection.DeleteIndexInPlace(slice, index)
	return nil
}

func removeFromSliceIfExists(slice *[]*internalTask, task *internalTask) bool {
	index := slices.Index(*slice, task)
	if index == -1 {
		return false
	}
	util_collection.DeleteIndexInPlace(slice, index)
	return true
}
