package items

import (
	"paladin_gearing_go/types/common"
	"paladin_gearing_go/types/stats"
)

const LOW_HIGH_MOP_ITEM_LEVELS_THRESHOLD = 522
const LOW_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL = 4
const HIGH_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL = 3
const MAX_UPGRADE_LEVEL = 2

type ItemRef struct {
	itemId        int64
	itemLevel     int16
	itemLevelBase int16
}

func (item *ItemRef) UpgradeLevel() int16 {
	if item.itemLevelBase < LOW_HIGH_MOP_ITEM_LEVELS_THRESHOLD {
		return (item.itemLevel - item.itemLevelBase) / LOW_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL
	} else {
		return (item.itemLevel - item.itemLevelBase) / HIGH_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL
	}
}

type FullItem struct {
	// generally fixed from imports
	ref         ItemRef
	slot        common.SlotItem
	baseName    string
	armorType   common.ArmorType
	primaryStat common.PrimaryStatType
	socketSlots []common.SocketType
	socketBonus stats.StatBlock
	phase       int8

	// specific item instance choices
	reforge       stats.ReforgeRecipe
	gemChoice     []int32
	enchantChoice int32
	randomSuffix  int32

	// stats for different purposes
	statBase    stats.StatBlock // constant stats post reforge
	statEnchant stats.StatBlock // stats added from gems, enchant (or trinket model)
	totalCap    stats.StatBlock // constant total stats as they contribute to caps
	totalRated  stats.StatBlock // averaged variable total stats for rating purposes
}

func FullItemData_fromWowSim(ref ItemRef, slot common.SlotItem, baseName string, statBase stats.StatBlock,
	armorType common.ArmorType, socketSlots []common.SocketType,
	socketBonus stats.StatBlock, phase int8) FullItem {
	return FullItem{ref, slot, baseName, armorType, statBase.PrimaryStat(),
		socketSlots, socketBonus, phase,
		stats.ReforgeRecipe_empty, nil, 0, 0,
		statBase, stats.StatBlock_empty, statBase, statBase}
}

func (item *FullItem) ChangeForReforge(changedStat stats.StatBlock, reforge stats.ReforgeRecipe) FullItem {
	var totalRated, totalCap stats.StatBlock
	if item.statEnchant.IsEmpty() {
		totalRated = changedStat
		totalCap = changedStat
	} else if item.slot.AddEnchantToCap() {
		totalRated = changedStat.Add(&item.statEnchant)
		totalCap = totalRated
	} else {
		totalRated = changedStat.Add(&item.statEnchant)
		totalCap = changedStat
	}

	return FullItem{item.ref, item.slot, item.baseName, item.armorType, item.primaryStat, item.socketSlots, item.socketBonus, item.phase,
		reforge, item.gemChoice, item.enchantChoice, item.randomSuffix,
		changedStat, item.statEnchant, totalCap, totalRated}
}

func (item *FullItem) FullName() string {
	if item.reforge.IsEmpty() {
		return item.baseName
	} else {
		return item.baseName + " " + item.reforge.Str()
	}
}

func (item *FullItem) IsEmpty() bool {
	return item.ref.itemId == 0
}

func (item *FullItem) Equals(other *FullItem) bool {
	return item.ref.itemId == other.ref.itemId && item.ref.itemLevel == other.ref.itemLevel && item.slot == other.slot &&
		item.statBase == other.statBase && item.statEnchant == other.statEnchant
}

type FullEquipMap [16]FullItem

func (equipMap *FullEquipMap) Get(slot common.SlotEquip) FullItem {
	return equipMap[slot]
}

type FullItemSet struct {
	items      FullEquipMap
	totalCap   stats.StatBlock
	totalRated stats.StatBlock
}
