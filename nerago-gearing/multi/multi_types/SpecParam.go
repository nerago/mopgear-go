package multi_types

import (
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/setup"
)

type SpecParam struct {
	// basic settings
	Label      string
	Model      gear_model.SpecModel
	ItemInputs ItemInputs
}

type ItemInputs struct {
	GearFile                  string
	RequestRatingPercent      float64
	ExtraUpgradeLevel         items.UpgradeLevel
	ForceUpgradeExistingItems items.UpgradeLevel
	MissingEnchant            setup.MissingEnchantMode
	ExtraItems                []items.ItemId
	ExtraFromBags             bool
	BlockedItems              []items.ItemId
	SemiFixedSlots            map[items.SlotEquip][]items.ItemId
	ReportVariant             map[items.SlotEquip]items.ItemId
}

func (param *SpecParam) AddExtraItems(extraItemIds []items.ItemId) {
	param.ItemInputs.ExtraItems = append(param.ItemInputs.ExtraItems, extraItemIds...)
}

func (param *SpecParam) AddExtraItem(extraItemId items.ItemId) {
	param.ItemInputs.ExtraItems = append(param.ItemInputs.ExtraItems, extraItemId)
}

func (param *SpecParam) AddBagsExtra() {
	param.ItemInputs.ExtraFromBags = true
}

func (param *SpecParam) BlockItem(itemId items.ItemId) {
	param.ItemInputs.BlockedItems = append(param.ItemInputs.BlockedItems, itemId)
}

func (param *SpecParam) ForceSingleSlot(slot items.SlotEquip, itemId items.ItemId) {
	if param.ItemInputs.SemiFixedSlots == nil {
		param.ItemInputs.SemiFixedSlots = make(map[items.SlotEquip][]items.ItemId)
	}
	if param.ItemInputs.SemiFixedSlots[slot] != nil {
		panic("slot already has forced item(s) set")
	}
	param.ItemInputs.SemiFixedSlots[slot] = []items.ItemId{itemId}
}

func (param *SpecParam) ForceTryAllSlot(slot items.SlotEquip, idList []items.ItemId) {
	if param.ItemInputs.SemiFixedSlots == nil {
		param.ItemInputs.SemiFixedSlots = make(map[items.SlotEquip][]items.ItemId)
	}
	if param.ItemInputs.SemiFixedSlots[slot] != nil {
		panic("slot already has forced item(s) set")
	}
	param.ItemInputs.SemiFixedSlots[slot] = idList
}

func (param *SpecParam) AddReportVariant(slot items.SlotEquip, id items.ItemId) {
	if param.ItemInputs.ReportVariant == nil {
		param.ItemInputs.ReportVariant = make(map[items.SlotEquip]items.ItemId)
	}
	param.ItemInputs.ReportVariant[slot] = id
}
