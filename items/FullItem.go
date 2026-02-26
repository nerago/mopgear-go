package items

import (
	"iter"
	"paladin_gearing_go/stats"
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
	Slot        stats.SlotItem
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
	TotalCap    stats.StatBlock // constant total stats as they contribute to caps
	TotalRated  stats.StatBlock // averaged variable total stats for rating purposes
}

func FullItem_FromWowSim(ref ItemRef, slot stats.SlotItem, baseName string, statBase stats.StatBlock,
	armorType stats.ArmorType, socketSlots []stats.SocketType,
	socketBonus stats.StatBlock, phase int8) FullItem {
	return FullItem{ref, slot, baseName, armorType, statBase.PrimaryStat(),
		socketSlots, socketBonus, phase,
		stats.ReforgeRecipe_empty, nil, 0, 0,
		statBase, stats.StatBlock_empty, statBase, statBase}
}

func (item *FullItem) ChangedForReforge(changedStat stats.StatBlock, reforge stats.ReforgeRecipe) *FullItem {
	newItem := item.ChangedBaseStats(changedStat)
	newItem.Reforge = reforge
	return newItem
}

func derivedStatFields(slot stats.SlotItem, statBase, statEnchant stats.StatBlock) (stats.StatBlock, stats.StatBlock) {
	if statEnchant.IsEmpty() {
		return statBase, statBase
	} else if slot.AddEnchantToCap() {
		sum := stats.StatBlock_Add_NoPointer(statBase, statEnchant)
		return sum, sum
	} else {
		sum := stats.StatBlock_Add_NoPointer(statBase, statEnchant)
		return statBase, sum
	}
}

func (item *FullItem) ChangedBaseStats(changedBase stats.StatBlock) *FullItem {
	totalCap, totalRated := derivedStatFields(item.Slot, changedBase, item.StatEnchant)
	return &FullItem{item.Ref, item.Slot, item.BaseName, item.ArmorType, item.PrimaryStat,
		item.SocketSlots, item.SocketBonus, item.Phase,
		item.Reforge, item.GemChoice, item.EnchantChoice, item.RandomSuffix,
		changedBase, item.StatEnchant, totalCap, totalRated}
}

func (item *FullItem) ChangedEnchantStats(changedEnchant stats.StatBlock) *FullItem {
	totalCap, totalRated := derivedStatFields(item.Slot, item.StatBase, changedEnchant)
	return &FullItem{item.Ref, item.Slot, item.BaseName, item.ArmorType, item.PrimaryStat,
		item.SocketSlots, item.SocketBonus, item.Phase,
		item.Reforge, item.GemChoice, item.EnchantChoice, item.RandomSuffix,
		item.StatBase, changedEnchant, totalCap, totalRated}
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

func (equipMap *FullEquipMap) Get(slot stats.SlotEquip) *FullItem {
	return equipMap[slot]
}

func (equipMap *FullEquipMap) AllItemSeq() iter.Seq[*FullItem] {
	return func(yield func(*FullItem) bool) {
		for _, item := range equipMap {
			if !item.IsEmpty() {
				if !yield(item) {
					return
				}
			}
		}
	}
}

type FullItemSet struct {
	Items      FullEquipMap
	TotalCap   stats.StatBlock
	TotalRated stats.StatBlock
}
