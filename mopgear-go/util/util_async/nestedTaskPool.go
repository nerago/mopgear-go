package util_async

import (
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

func (pool *NestedTaskPoolParent) selectQueuedTask() *internalTask {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	taskList := make([]*internalTask, 0)
	for _, child := range pool.children {
		child.mutex.Lock()
		taskList = append(taskList, child.queue...)
		child.mutex.Unlock()
	}

	if len(taskList) > 0 {
		task := taskList[rand.Intn(len(taskList))]
		return task
	} else {
		return nil
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
		case <-pool.chanPoolStop:
			stopped = true
		}
	}
}

func (pool *NestedTaskPoolParent) workerThread() {
	for !pool.stopped {
		task := pool.selectQueuedTask()
		if task != nil {
			task.owner.notifyStarting(task)
			task.run()
			task.owner.notifyTaskComplete(task)
		} else {
			break
		}
	}
	pool.chanThreadExit <- true
}

type internalTask struct {
	owner *NestedTaskPoolChild
	run   func()
}

type NestedTaskPoolChild struct {
	mutex       sync.Mutex
	pool        *NestedTaskPoolParent
	queue       []*internalTask
	running     []*internalTask
	waitChannel chan bool
}

func (child *NestedTaskPoolChild) Go(run func()) {
	child.mutex.Lock()
	defer child.mutex.Unlock()

	task := &internalTask{child, run}
	child.queue = append(child.queue, task)
	child.pool.taskAdded(task)
}

func (child *NestedTaskPoolChild) notifyStarting(task *internalTask) {
	child.mutex.Lock()
	defer child.mutex.Unlock()

	removeFromSlice(&child.queue, task)
	child.running = append(child.running, task)
}

func (child *NestedTaskPoolChild) notifyTaskComplete(task *internalTask) {
	child.mutex.Lock()
	defer child.mutex.Unlock()

	removeFromSlice(&child.running, task)

	if len(child.queue) == 0 && len(child.running) == 0 && child.waitChannel != nil {
		close(child.waitChannel)
		child.waitChannel = nil
	}
}

// only clears wait if at least one child has been added and finished
func (child *NestedTaskPoolChild) Wait() {
	waitChan := child.waitChannel
	if waitChan != nil {
		<-waitChan
	}
}

func removeFromSlice(slice *[]*internalTask, task *internalTask) {
	index := slices.Index(*slice, task)
	if index == -1 {
		panic("not in slice")
	}
	util_collection.DeleteIndexInPlace(slice, index)
}
