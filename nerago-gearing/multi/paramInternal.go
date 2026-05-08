package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/solver"
	"sync"
	"sync/atomic"
)

type multiSetParamInternal struct {
	multi_types.MultiSetParam

	job *MultiSetJob

	// working data
	exactEquippedGear items.FullEquipMap
	itemOptions       items.FullOptionsMap
	addedFromBags     []items.ItemId
	seenInSolutions   *seenMap
	baselineResult    solver.SolveOutput
	ratingMultiply    float64 // derived

	//debug
	solveFailCount    atomic.Uint64
	solveSuccessCount atomic.Uint64
}

func (param *multiSetParamInternal) init() {
	param.seenInSolutions = &seenMap{content: make(map[items.ItemId]uint32)}
}

type seenMap struct {
	content map[items.ItemId]uint32
	mutex   sync.Mutex
}

func (seen *seenMap) Add(itemSet *items.FullItemSet) {
	seen.mutex.Lock()
	defer seen.mutex.Unlock()

	for item := range itemSet.Items().AllItemSeq() {
		seen.content[item.ItemId()]++
	}
}
