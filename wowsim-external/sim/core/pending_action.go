//go:build !bubblesim
// +build !bubblesim

package core

import (
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

type PendingActionQueue struct {
	PendingActionQueueDouble
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
	pa := queue.chain.nextLink
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
	for pa := queue.chain.nextLink; pa != queue.chain; pa = pa.nextLink {
		if pa.CleanUp != nil {
			pa.CleanUp(sim)
		}

		pa.dispose(sim)
	}
}
