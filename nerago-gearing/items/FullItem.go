package items

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

// /////////////////////////////////////////////////////////////
type FullItem struct {
	// generally fixed from imports
	itemId       ItemId
	itemLevel    uint16
	upgradeLevel int8
	slot         SlotItem
	baseName     string
	armorType    stats.ArmorType
	primaryStat  stats.PrimaryStatType
	socketSlots  []stats.SocketType
	socketBonus  stats.StatBlock
	phase        int8

	// specific item instance choices
	reforge       stats.ReforgeRecipe
	gemChoice     []stats.GemInfo
	enchantChoice uint32
	randomSuffix  int32

	// stats for different purposes
	statBase    stats.StatBlock // constant stats post reforge
	statEnchant stats.StatBlock // stats added from gems, enchant, or trinket model

	total stats.StatBlock // constant total stats as they contribute to caps
}

func FullItem_FromWowSim(itemId ItemId, itemLevel uint16, itemLevelBase uint16, upgradeLevel int8, slot SlotItem, baseName string, statBase stats.StatBlock, armorType stats.ArmorType, socketSlots []stats.SocketType, socketBonus stats.StatBlock, phase int8) FullItem {
	return FullItem{
		itemId, itemLevel, upgradeLevel, slot, baseName, armorType, statBase.PrimaryStat(),
		socketSlots, socketBonus, phase,
		stats.ReforgeRecipe_empty, nil, 0, 0,
		statBase, stats.StatBlock_empty,
		statBase}
}

func FullItem_ForTest(itemId ItemId, slot SlotItem, statBase stats.StatBlock) FullItem {
	return FullItem{
		itemId, 400, 1,
		slot, slot.Name(), stats.Armor_None, statBase.PrimaryStat(),
		nil, stats.StatBlock_empty, 0,
		stats.ReforgeRecipe_empty, nil, 0, 0,
		statBase, stats.StatBlock_empty,
		statBase}
}

func (item *FullItem) NewWithChangedStatsReforge(changeStats stats.StatBlock, changeReforge stats.ReforgeRecipe) *FullItem {
	var newItem FullItem = *item
	newItem.reforge = changeReforge
	newItem.statBase = changeStats
	newItem.changeDerivedStatFields()
	return &newItem
}

func (item *FullItem) NewWithChangedStatsSuffix(changeStats stats.StatBlock, randomSuffix int32) *FullItem {
	var newItem FullItem = *item
	newItem.randomSuffix = randomSuffix
	newItem.statBase = changeStats
	newItem.changeDerivedStatFields()
	return &newItem
}

func (item *FullItem) NewWithEnchantDetails(socketSlots []stats.SocketType, gemChoice []stats.GemInfo, enchantChoice uint32, statEnchant stats.StatBlock) *FullItem {
	var newItem FullItem = *item
	newItem.socketSlots = socketSlots
	newItem.gemChoice = gemChoice
	newItem.enchantChoice = enchantChoice
	newItem.statEnchant = statEnchant
	newItem.changeDerivedStatFields()
	return &newItem
}

func (item *FullItem) NewWithInstanceDetails(socketSlots []stats.SocketType, reforge stats.ReforgeRecipe, gemChoice []stats.GemInfo, enchantChoice uint32, randomSuffix int32, statEnchant stats.StatBlock) *FullItem {
	var newItem FullItem = *item
	newItem.socketSlots = socketSlots
	newItem.reforge = reforge
	newItem.gemChoice = gemChoice
	newItem.enchantChoice = enchantChoice
	newItem.randomSuffix = randomSuffix
	newItem.statEnchant = statEnchant
	newItem.changeDerivedStatFields()
	return &newItem
}

func (item *FullItem) changeDerivedStatFields() {
	stats.StatBlock_Add_Into(&item.statBase, &item.statEnchant, &item.total)
}

func (item *FullItem) Total() *stats.StatBlock {
	return &item.total
}

func (item *FullItem) StatBase() *stats.StatBlock {
	return &item.statBase
}

func (item *FullItem) StatEnchant() *stats.StatBlock {
	return &item.statEnchant
}

func (item *FullItem) AppendFullName(build *util.StringBuild2) {
	build.WriteString(item.baseName)
	if !item.reforge.IsEmpty() {
		build.WriteRune(' ')
		item.reforge.AppendString(build)
	}
}

func (item *FullItem) BaseName() string {
	return item.baseName
}

func (item *FullItem) ItemId() ItemId {
	return item.itemId
}

func (item *FullItem) ItemLevel() uint16 {
	return item.itemLevel
}

func (item *FullItem) UpgradeLevel() int8 {
	return item.upgradeLevel
}

func (item *FullItem) Phase() int8 {
	return item.phase
}

func (item *FullItem) SlotItem() SlotItem {
	return item.slot
}

func (item *FullItem) ArmorType() stats.ArmorType {
	return item.armorType
}

func (item *FullItem) PrimaryStat() stats.PrimaryStatType {
	return item.primaryStat
}

func (item *FullItem) Reforge() stats.ReforgeRecipe {
	return item.reforge
}

func (item *FullItem) SocketSlots() []stats.SocketType {
	return item.socketSlots
}

func (item *FullItem) SocketBonus() *stats.StatBlock {
	return &item.socketBonus
}

func (item *FullItem) GemChoice() []stats.GemInfo {
	return item.gemChoice
}

func (item *FullItem) EnchantChoice() uint32 {
	return item.enchantChoice
}

func (item *FullItem) RandomSuffix() int32 {
	return item.randomSuffix
}

func (item *FullItem) Equals(other *FullItem) bool {
	return item.itemId == other.itemId && item.itemLevel == other.itemLevel && item.slot == other.slot &&
		stats.StatBlock_Equals(&item.statBase, &other.statBase) && stats.StatBlock_Equals(&item.statEnchant, &other.statEnchant)
}

func (item *FullItem) CreateString() string {
	build := util.StringBuild2{}
	item.AppendString(&build)
	return build.String()
}

func (item *FullItem) CreateFullName() string {
	build := util.StringBuild2{}
	item.AppendFullName(&build)
	return build.String()
}

func (item *FullItem) AppendString(build *util.StringBuild2) {
	build.WriteString("{ ")
	build.WriteString(item.slot.Name())

	build.WriteString(" \"")
	item.AppendFullName(build)

	build.WriteString("\" id=")
	build.WriteUint32(uint32(item.itemId))

	build.WriteString(" lvl=")
	build.WriteUint16(item.itemLevel)
	build.WriteRune(' ')

	item.statBase.AppendString(build)

	if !item.statEnchant.IsEmpty() {
		build.WriteString(" ENCHANT ")
		item.statEnchant.AppendString(build)
	}

	if len(item.gemChoice) > 0 {
		build.WriteString(" GEMS ")
		for _, gem := range item.gemChoice {
			gem.AppendString(build)
		}
	}

	build.WriteString(" }")
}

func (item *FullItem) MakeItemWithRandomSuffix(randomSuffix int32) *FullItem {
	if randomSuffix != 0 {
		wowSimStats, suffix := item.randomStatsFromWowSim(randomSuffix)

		newStats := stats.StatBlock{}
		newStats.SetFromAddOthers(item.StatBase(), &wowSimStats)

		item = item.NewWithChangedStatsSuffix(newStats, randomSuffix)
		item.baseName += " " + suffix
	}

	return item
}
