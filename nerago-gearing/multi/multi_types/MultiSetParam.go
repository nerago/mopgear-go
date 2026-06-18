package multi_types

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
)

type MultiSetParam struct {
	// basic settings
	Label    string
	GearFile string
	Model    model.Model

	// solve settings
	// IncludeInFirstPass   bool // TODO consider reintroducing in hights solver
	RequestRatingPercent float64

	// extra item settings
	ExtraUpgradeLevel         int8
	ForceUpgradeExistingItems int8

	// intended to use accessor funcs
	ExtraItems     []items.ItemId
	ExtraFromBags  bool
	BlockedItems   []items.ItemId
	SemiFixedSlots map[items.SlotEquip][]items.ItemId
	ReportVariant  map[items.SlotEquip]items.ItemId
}

func (param *MultiSetParam) AddExtraItems(extraItemIds []items.ItemId) {
	param.ExtraItems = append(param.ExtraItems, extraItemIds...)
}

func (param *MultiSetParam) AddExtraItem(extraItemId items.ItemId) {
	param.ExtraItems = append(param.ExtraItems, extraItemId)
}

func (param *MultiSetParam) AddBagsExtra() {
	param.ExtraFromBags = true
}

func (param *MultiSetParam) BlockItem(itemId items.ItemId) {
	param.BlockedItems = append(param.BlockedItems, itemId)
}

func (param *MultiSetParam) ForceSingleSlot(slot items.SlotEquip, itemId items.ItemId) {
	if param.SemiFixedSlots == nil {
		param.SemiFixedSlots = make(map[items.SlotEquip][]items.ItemId)
	}
	if param.SemiFixedSlots[slot] != nil {
		panic("slot already has forced item(s) set")
	}
	param.SemiFixedSlots[slot] = []items.ItemId{itemId}
}

func (param *MultiSetParam) ForceTryAllSlot(slot items.SlotEquip, idList []items.ItemId) {
	if param.SemiFixedSlots == nil {
		param.SemiFixedSlots = make(map[items.SlotEquip][]items.ItemId)
	}
	if param.SemiFixedSlots[slot] != nil {
		panic("slot already has forced item(s) set")
	}
	param.SemiFixedSlots[slot] = idList
}

func (param *MultiSetParam) AddReportVariant(slot items.SlotEquip, id items.ItemId) {
	if param.ReportVariant == nil {
		param.ReportVariant = make(map[items.SlotEquip]items.ItemId)
	}
	param.ReportVariant[slot] = id
}
