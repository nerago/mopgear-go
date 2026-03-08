package items

import (
	"paladin_gearing_go/stats"
)

type fullItem_common struct {
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
}

func (item *fullItem_common) FullName() string {
	if item.Reforge.IsEmpty() {
		return item.BaseName
	} else {
		return item.BaseName + " " + item.Reforge.Str()
	}
}

func (item *fullItem_common) ItemId() uint32 {
	return item.Ref.ItemId
}

func (item *fullItem_common) IsEmpty() bool {
	return item.Ref.ItemId == 0
}

func (item *FullItem) Equals(other *FullItem) bool {
	return item.Ref.ItemId == other.Ref.ItemId && item.Ref.ItemLevel == other.Ref.ItemLevel && item.Slot == other.Slot &&
		item.StatBase == other.StatBase && item.StatEnchant == other.StatEnchant
}
