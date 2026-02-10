package items

import (
	"iter"
	"paladin_gearing_go/types/common"
	"paladin_gearing_go/types/stats"
)

const LOW_HIGH_MOP_ITEM_LEVELS_THRESHOLD = 522
const LOW_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL = 4
const HIGH_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL = 3
const MAX_UPGRADE_LEVEL = 2

type ItemRef struct {
	ItemId        uint32
	ItemLevel     uint16
	ItemLevelBase uint16
}

func (ref ItemRef) UpgradeLevel() int16 {
	if ref.ItemLevel < ref.ItemLevelBase {
		return -1
	} else if ref.ItemLevelBase < LOW_HIGH_MOP_ITEM_LEVELS_THRESHOLD {
		return int16((ref.ItemLevel - ref.ItemLevelBase) / LOW_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL)
	} else {
		return int16((ref.ItemLevel - ref.ItemLevelBase) / HIGH_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL)
	}
}

type FullItem struct {
	// generally fixed from imports
	Ref         ItemRef
	Slot        common.SlotItem
	BaseName    string
	ArmorType   common.ArmorType
	PrimaryStat common.PrimaryStatType
	SocketSlots []common.SocketType
	SocketBonus stats.StatBlock
	Phase       int8

	// specific item instance choices
	Reforge       stats.ReforgeRecipe
	GemChoice     []int32
	EnchantChoice int32
	RandomSuffix  int32

	// stats for different purposes
	StatBase    stats.StatBlock // constant stats post reforge
	StatEnchant stats.StatBlock // stats added from gems, enchant, or trinket model
	TotalCap    stats.StatBlock // constant total stats as they contribute to caps
	TotalRated  stats.StatBlock // averaged variable total stats for rating purposes
}

func FullItem_FromWowSim(ref ItemRef, slot common.SlotItem, baseName string, statBase stats.StatBlock,
	armorType common.ArmorType, socketSlots []common.SocketType,
	socketBonus stats.StatBlock, phase int8) FullItem {
	return FullItem{ref, slot, baseName, armorType, statBase.PrimaryStat(),
		socketSlots, socketBonus, phase,
		stats.ReforgeRecipe_empty, nil, 0, 0,
		statBase, stats.StatBlock_empty, statBase, statBase}
}

func (item *FullItem) ChangeForReforge(changedStat stats.StatBlock, reforge stats.ReforgeRecipe) FullItem {
	var totalRated, totalCap stats.StatBlock
	if item.StatEnchant.IsEmpty() {
		totalRated = changedStat
		totalCap = changedStat
	} else if item.Slot.AddEnchantToCap() {
		totalRated = changedStat.Add(&item.StatEnchant)
		totalCap = totalRated
	} else {
		totalRated = changedStat.Add(&item.StatEnchant)
		totalCap = changedStat
	}

	return FullItem{item.Ref, item.Slot, item.BaseName, item.ArmorType, item.PrimaryStat,
		item.SocketSlots, item.SocketBonus, item.Phase,
		reforge, item.GemChoice, item.EnchantChoice, item.RandomSuffix,
		changedStat, item.StatEnchant, totalCap, totalRated}
}

func (item *FullItem) FullName() string {
	if item.Reforge.IsEmpty() {
		return item.BaseName
	} else {
		return item.BaseName + " " + item.Reforge.Str()
	}
}

func (item *FullItem) ItemId() uint32 {
	return item.Ref.ItemId
}

func (item *FullItem) IsEmpty() bool {
	return item.Ref.ItemId == 0
}

func (item *FullItem) Equals(other *FullItem) bool {
	return item.Ref.ItemId == other.Ref.ItemId && item.Ref.ItemLevel == other.Ref.ItemLevel && item.Slot == other.Slot &&
		item.StatBase == other.StatBase && item.StatEnchant == other.StatEnchant
}

type FullEquipMap [16]*FullItem

func (equipMap *FullEquipMap) Get(slot common.SlotEquip) *FullItem {
	return equipMap[slot]
}

func (equipMap *FullEquipMap) AllItems() iter.Seq[*FullItem] {
	return func(yield func(*FullItem) bool) {
		for _, item := range equipMap {
			if !yield(item) {
				return
			}
		}
	}
}

type FullItemSet struct {
	Items      FullEquipMap
	TotalCap   stats.StatBlock
	TotalRated stats.StatBlock
}
