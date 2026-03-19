package items

import (
	"paladin_gearing_go/stats"
	"strings"
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

func (item *fullItem_common) AppendFullName(build *strings.Builder) {
	build.WriteString(item.BaseName)
	if !item.Reforge.IsEmpty() {
		build.WriteRune(' ')
		item.Reforge.AppendString(build)
	}
}

func (item *fullItem_common) ItemId() ItemId {
	return item.Ref.ItemId
}

func (item *fullItem_common) IsEmpty() bool {
	return item.Ref.ItemId == 0
}

func (item *FullItem) Equals(other *FullItem) bool {
	return item.Ref.ItemId == other.Ref.ItemId && item.Ref.ItemLevel == other.Ref.ItemLevel && item.Slot == other.Slot &&
		stats.StatBlock_Equals(&item.StatBase, &other.StatBase) && stats.StatBlock_Equals(&item.StatEnchant, &other.StatEnchant)
}

func (itemSet *FullItemSet) Equals(other FullItemSet) bool {
	return itemSet.items.Equals(&other.items)
}
