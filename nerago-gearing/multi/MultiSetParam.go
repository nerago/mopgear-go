package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/solver"
	"sync"
)

type MultiSetParam struct {
	// basic settings
	Label    string
	GearFile string
	Model    model.Model
	job      *MultiSetJob

	// solve settings
	IncludeInFirstPass   bool
	RequestRatingPercent float64
	PhasedAcceptable     bool

	// extra item settings
	ExtraUpgradeLevel int16
	extraItems        []items.ItemId
	extraFromBags     bool
	fixedSlots        map[items.SlotEquip]items.ItemId

	// working data
	exactEquippedGear items.FullEquipMap
	itemOptions       items.FullOptionsMap
	addedFromBags     []items.ItemId
	seenInSolutions   *seenMap
	baselineResult    solver.SolveOutput
	ratingMultiply    uint64 // derived

	// stuff not ported
	// boolean upgradeCurrentItems;
	// boolean challengeScale;
	// double worstCommonPenalty;
	// Map<Integer, Integer> duplicatedItems;
	// List<Integer> removeItems;
	// suppressSlotCheck
}

func (param *MultiSetParam) init(job *MultiSetJob) {
	param.job = job
	param.seenInSolutions = &seenMap{content: make(map[items.ItemId]uint32)}
}

func (param *MultiSetParam) AddExtraItems(extraItemIds []items.ItemId) *MultiSetParam {
	param.extraItems = append(param.extraItems, extraItemIds...)
	return param
}

func (param *MultiSetParam) AddBagsExtra() {
	param.extraFromBags = true
}

func (param *MultiSetParam) AddExtraItem(extraItemId items.ItemId) *MultiSetParam {
	param.extraItems = append(param.extraItems, extraItemId)
	return param
}

func (param *MultiSetParam) AddFixedSlot(slot items.SlotEquip, itemId items.ItemId) *MultiSetParam {
	if param.fixedSlots == nil {
		param.fixedSlots = make(map[items.SlotEquip]items.ItemId)
	}
	param.fixedSlots[slot] = itemId
	return param
}

type seenMap struct {
	content map[items.ItemId]uint32
	mutex   sync.Mutex
}

func (seen *seenMap) Add(itemSet *items.FullItemSet) {
	seen.mutex.Lock()
	for item := range itemSet.Items().AllItemSeq() {
		seen.content[item.ItemId()]++
	}
	seen.mutex.Unlock()
}
