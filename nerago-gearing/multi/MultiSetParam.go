package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/solver"
	"sync"
	"sync/atomic"
)

type MultiSetParam struct {
	// basic settings
	Label    string
	GearFile string
	Model    model.Model

	// solve settings
	IncludeInFirstPass   bool
	RequestRatingPercent float64
	PhasedAcceptable     bool

	// extra item settings
	ExtraUpgradeLevel         int8
	ForceUpgradeExistingItems int8
	extraItems                []items.ItemId
	extraFromBags             bool
	blockedItems              []items.ItemId
	fixedSlots                map[items.SlotEquip]items.ItemId
	reportVariant             map[items.SlotEquip]items.ItemId

	// stuff not ported
	// boolean challengeScale;
	// Map<Integer, Integer> duplicatedItems;
	// suppressSlotCheck
}

type multiSetParamInternal struct {
	MultiSetParam

	job *MultiSetJob

	// working data
	exactEquippedGear   items.FullEquipMap
	itemOptions         items.FullOptionsMap
	addedFromBags       []items.ItemId
	seenInSolutions     *seenMap
	baselineResult      solver.SolveOutput
	baselineResultHighs solver.SolveOutput
	ratingMultiply      float64 // derived

	//debug
	solveFailCount    atomic.Uint64
	solveSuccessCount atomic.Uint64
}

func (param *multiSetParamInternal) init() {
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

func (param *MultiSetParam) BlockItem(itemId items.ItemId) *MultiSetParam {
	param.blockedItems = append(param.blockedItems, itemId)
	return param
}

func (param *MultiSetParam) AddFixedSlot(slot items.SlotEquip, itemId items.ItemId) *MultiSetParam {
	if param.fixedSlots == nil {
		param.fixedSlots = make(map[items.SlotEquip]items.ItemId)
	}
	param.fixedSlots[slot] = itemId
	return param
}

func (param *MultiSetParam) AddReportVariant(slot items.SlotEquip, id items.ItemId) {
	if param.reportVariant == nil {
		param.reportVariant = make(map[items.SlotEquip]items.ItemId)
	}
	param.reportVariant[slot] = id
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

func (seen *seenMap) Add1000(itemSet *items.FullItemSet) {
	seen.mutex.Lock()
	defer seen.mutex.Unlock()

	for item := range itemSet.Items().AllItemSeq() {
		seen.content[item.ItemId()] += 1000
	}
}
