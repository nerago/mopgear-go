//go:build !bubblesim
// +build !bubblesim

package core

import (
	"slices"
	"time"
)

type ActionPriority int32

const (
	ActionPriorityPrePull ActionPriority = -20

	ActionPriorityGCD ActionPriority = 0

	// Higher than GCD because regen can cause GCD actions (if we were waiting
	// for mana).
	ActionPriorityRegen ActionPriority = 1

	// Autos can cause regen (JoW, rage, energy procs, etc) so they should be
	// higher prio so that we never go backwards in the priority order.
	ActionPriorityAuto ActionPriority = 2

	// DOTs need to be higher than anything else so that dots can properly expire before we take other actions.
	ActionPriorityDOT ActionPriority = 3

	ActionPriorityHigh ActionPriority = 10
)

type PendingAction struct {
	NextActionAt time.Duration
	Priority     ActionPriority

	// Action to perform (required).
	OnAction func(sim *Simulation)
	// Cleanup when the action is cancelled (optional).
	CleanUp func(sim *Simulation)

	cancelled bool
	consumed  bool
	canPool   bool // Flags the PA as safe to use in shared object pools.

	nextLink, prevLink *PendingAction
}

func (pa *PendingAction) IsConsumed() bool {
	return pa == nil || pa.consumed
}

func (pa *PendingAction) Cancel(sim *Simulation) {
	if pa.cancelled {
		return
	}

	if pa.CleanUp != nil {
		pa.CleanUp(sim)
		pa.CleanUp = nil
	}

	pa.cancelled = true

	sim.pendingActionQueue.cancel(pa)
}

func (pa *PendingAction) dispose(sim *Simulation) {
	if pa.canPool && pa.consumed {
		sim.pendingActionPool.Put(pa)
	}
}

// type PendingActionQueue interface {
// 	add(add *PendingAction)
// 	getNext() *PendingAction
// 	popNext()
// 	cancel(*PendingAction)
// 	reset()
// 	cleanup(*Simulation)
// }

// type PendingActionQueue PendingActionQueueSingle

type PendingActionQueue struct {
	PendingActionQueueSingle
}

type PendingActionQueueSingle struct {
	first *PendingAction
}

func (queue *PendingActionQueueSingle) add(add *PendingAction) {
	curr := queue.first
	if add.NextActionAt < curr.NextActionAt || (add.NextActionAt == curr.NextActionAt && add.Priority > curr.Priority) {
		queue.first = add
		add.nextLink = curr
		return
	}

	prev := curr
	curr = curr.nextLink
	for curr != nil {
		if add.NextActionAt < curr.NextActionAt || (add.NextActionAt == curr.NextActionAt && add.Priority > curr.Priority) {
			prev.nextLink = add
			add.nextLink = curr
			return
		}
		prev = curr
		curr = curr.nextLink
	}
}

func (queue *PendingActionQueueSingle) getNext() *PendingAction {
	return queue.first
}

func (queue *PendingActionQueueSingle) popNext() {
	next := queue.first
	queue.first = next.nextLink
	next.nextLink = nil
}

func (queue *PendingActionQueueSingle) cancel(pa *PendingAction) {
	if pa == queue.first {
		queue.first = pa.nextLink
		pa.nextLink = nil
	} else {
		for curr := queue.first; curr != nil; curr = curr.nextLink {
			if curr.nextLink == pa {
				curr.nextLink = pa.nextLink
				pa.nextLink = nil
				return
			}
		}
	}
}

func (queue *PendingActionQueueSingle) reset() {
	queue.first = &PendingAction{
		NextActionAt: NeverExpires,
		OnAction: func(sim *Simulation) {
			panic("running sentinel pending action")
		},
	}
}

func (queue *PendingActionQueueSingle) cleanup(sim *Simulation) {
	for pa := queue.first; pa != nil; pa = pa.nextLink {
		if pa.CleanUp != nil {
			pa.CleanUp(sim)
		}

		pa.dispose(sim)
	}
}

type PendingActionQueueDouble struct {
	chain *PendingAction
}

func (queue *PendingActionQueueDouble) add(add *PendingAction) {
	// start next after sentinal, we should always find something
	curr := queue.chain.nextLink
	for {
		if add.NextActionAt < curr.NextActionAt || (add.NextActionAt == curr.NextActionAt && add.Priority > curr.Priority) {
			add.prevLink = curr.prevLink
			add.prevLink.nextLink = add
			add.nextLink = curr
			curr.prevLink = add
			return
		}
		curr = curr.nextLink
	}
}

func (queue *PendingActionQueueDouble) getNext() *PendingAction {
	return queue.chain.nextLink
}

func (queue *PendingActionQueueDouble) popNext() {
	pa := queue.chain
	pa.prevLink.nextLink = pa.nextLink
	pa.nextLink.prevLink = pa.prevLink
	pa.prevLink = nil
	pa.nextLink = nil
}

func (queue *PendingActionQueueDouble) cancel(pa *PendingAction) {
	if pa.prevLink != nil {
		pa.prevLink.nextLink = pa.nextLink
	}
	if pa.nextLink != nil {
		pa.nextLink.prevLink = pa.prevLink
	}
	pa.prevLink = nil
	pa.nextLink = nil
}

func (queue *PendingActionQueueDouble) reset() {
	sentinel := &PendingAction{
		NextActionAt: NeverExpires,
		OnAction: func(sim *Simulation) {
			panic("running sentinel pending action")
		},
	}
	sentinel.prevLink = sentinel
	sentinel.nextLink = sentinel
	queue.chain = sentinel
}

func (queue *PendingActionQueueDouble) cleanup(sim *Simulation) {
	for pa := queue.chain; pa != queue.chain; pa = pa.nextLink {
		if pa.CleanUp != nil {
			pa.CleanUp(sim)
		}

		pa.dispose(sim)
	}
}

type PendingActionQueueArray struct {
	pendingActions []*PendingAction
}

func (queue *PendingActionQueueArray) addAlternate(pa *PendingAction) {
	index := len(queue.pendingActions) - 1
	v := queue.pendingActions[index]
	if v.NextActionAt >= pa.NextActionAt || (v.NextActionAt == pa.NextActionAt && v.Priority < pa.Priority) {
		queue.pendingActions = append(queue.pendingActions, pa)
		return
	}

	index--
	for index >= 1 {
		v := queue.pendingActions[index]
		if v.NextActionAt >= pa.NextActionAt || (v.NextActionAt == pa.NextActionAt && v.Priority < pa.Priority) {
			queue.pendingActions = append(queue.pendingActions, nil)
			copy(queue.pendingActions[index+1:], queue.pendingActions[index:])
			queue.pendingActions[index] = pa
			return
		}
	}

	panic("shouldn't get here")
}

func (queue *PendingActionQueueArray) add(pa *PendingAction) {
	for index, v := range queue.pendingActions[1:] {
		if v.NextActionAt < pa.NextActionAt || (v.NextActionAt == pa.NextActionAt && v.Priority >= pa.Priority) {
			queue.pendingActions = append(queue.pendingActions, pa)
			copy(queue.pendingActions[index+2:], queue.pendingActions[index+1:])
			queue.pendingActions[index+1] = pa
			return
		}
	}
	queue.pendingActions = append(queue.pendingActions, pa)
}

func (queue *PendingActionQueueArray) getNext() *PendingAction {
	last := len(queue.pendingActions) - 1
	return queue.pendingActions[last]
}

func (queue *PendingActionQueueArray) popNext() {
	last := len(queue.pendingActions) - 1
	queue.pendingActions = queue.pendingActions[:last]
}

func (queue *PendingActionQueueArray) cancel(pa *PendingAction) {
	if i := slices.Index(queue.pendingActions, pa); i != -1 {
		queue.pendingActions = append(queue.pendingActions[:i], queue.pendingActions[i+1:]...)
	}
}

func (queue *PendingActionQueueArray) reset() {
	sentinelPendingAction := &PendingAction{
		NextActionAt: NeverExpires,
		OnAction: func(sim *Simulation) {
			panic("running sentinel pending action")
		},
	}
	queue.pendingActions = queue.pendingActions[:0]
	queue.pendingActions = append(queue.pendingActions, sentinelPendingAction)
}

func (queue *PendingActionQueueArray) cleanup(sim *Simulation) {
	for _, pa := range queue.pendingActions {
		if pa.CleanUp != nil {
			pa.CleanUp(sim)
		}

		pa.dispose(sim)
	}
}
