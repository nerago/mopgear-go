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
	extraItems        []uint32
	fixedSlots        map[items.SlotEquip]uint32

	// working data
	exactEquippedGear items.FullEquipMap
	itemOptions       items.FullOptionsMap
	seenInSolutions   *seenMap
	baselineResult    solver.SolveOutput
	ratingMultiply    uint64 // derived

	// stuff not ported
	// boolean upgradeCurrentItems;
	// boolean challengeScale;
	// double worstCommonPenalty;
	// Map<Integer, Integer> duplicatedItems;
	// List<Integer> removeItems;
	// double optimalRating;
	// FullItemSet optimalBaselineSet;
	// suppressSlotCheck
}

func (param *MultiSetParam) init(job *MultiSetJob) {
	param.job = job
	param.seenInSolutions = &seenMap{content: make(map[uint32]uint32)}
}

func (param *MultiSetParam) AddExtraItems(extraItemIds []uint32) *MultiSetParam {
	param.extraItems = append(param.extraItems, extraItemIds...)
	return param
}

func (param *MultiSetParam) AddExtraItem(extraItemId uint32) *MultiSetParam {
	param.extraItems = append(param.extraItems, extraItemId)
	return param
}

func (param *MultiSetParam) AddFixedSlot(slot items.SlotEquip, itemId uint32) *MultiSetParam {
	if param.fixedSlots == nil {
		param.fixedSlots = make(map[items.SlotEquip]uint32)
	}
	param.fixedSlots[slot] = itemId
	return param
}

type seenMap struct {
	content map[uint32]uint32
	mutex   sync.Mutex
	// or could use sync.Map
}

func (seen *seenMap) Add(itemSet *items.FullItemSet) {
	seen.mutex.Lock()
	for item := range itemSet.Items.AllItemSeq() {
		seen.content[item.ItemId()]++
	}
	seen.mutex.Unlock()
}
