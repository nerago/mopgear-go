package items

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

// /////////////////////////////////////////////////////////////
type FullItem struct {
	// generally fixed from imports
	Ref         ItemRef
	Slot        SlotItem
	BaseName    string
	ArmorType   stats.ArmorType
	PrimaryStat stats.PrimaryStatType
	SocketSlots []stats.SocketType
	SocketBonus stats.StatBlock
	Phase       int8

	// specific item instance choices
	Reforge       stats.ReforgeRecipe
	GemChoice     []stats.GemInfo
	EnchantChoice uint32
	RandomSuffix  int32

	// stats for different purposes
	StatBase    stats.StatBlock // constant stats post reforge
	StatEnchant stats.StatBlock // stats added from gems, enchant, or trinket model

	total stats.StatBlock // constant total stats as they contribute to caps
}

func FullItem_FromWowSim(ref ItemRef, slot SlotItem, baseName string, statBase stats.StatBlock, armorType stats.ArmorType, socketSlots []stats.SocketType, socketBonus stats.StatBlock, phase int8) FullItem {
	return FullItem{
		ref, slot, baseName, armorType, statBase.PrimaryStat(),
		socketSlots, socketBonus, phase,
		stats.ReforgeRecipe_empty, nil, 0, 0,
		statBase, stats.StatBlock_empty,
		statBase}
}

func FullItem_ForTest(itemId ItemId, slot SlotItem, statBase stats.StatBlock) FullItem {
	return FullItem{
		ItemRef_Make(itemId, 400, 404),
		slot, slot.Name(), stats.Armor_None, statBase.PrimaryStat(),
		nil, stats.StatBlock_empty, 0,
		stats.ReforgeRecipe_empty, nil, 0, 0,
		statBase, stats.StatBlock_empty,
		statBase}
}

func (item *FullItem) ChangeDerivedStatFields() {
	stats.StatBlock_Add_Into(&item.StatBase, &item.StatEnchant, &item.total)
}

func (item *FullItem) Total() *stats.StatBlock {
	return &item.total
}

func (item *FullItem) AppendFullName(build *util.StringBuild2) {
	build.WriteString(item.BaseName)
	if !item.Reforge.IsEmpty() {
		build.WriteRune(' ')
		item.Reforge.AppendString(build)
	}
}

func (item *FullItem) ItemId() ItemId {
	return item.Ref.ItemId
}

func (item *FullItem) IsEmpty() bool {
	return item.Ref.ItemId == 0
}

func (item *FullItem) Equals(other *FullItem) bool {
	return item.Ref.ItemId == other.Ref.ItemId && item.Ref.ItemLevel == other.Ref.ItemLevel && item.Slot == other.Slot &&
		stats.StatBlock_Equals(&item.StatBase, &other.StatBase) && stats.StatBlock_Equals(&item.StatEnchant, &other.StatEnchant)
}

func (item *FullItem) EqualsExceptEnchant(other *FullItem) bool {
	return item.Ref.ItemId == other.Ref.ItemId && item.Ref.ItemLevel == other.Ref.ItemLevel && item.Slot == other.Slot &&
		stats.StatBlock_Equals(&item.StatBase, &other.StatBase)
}
