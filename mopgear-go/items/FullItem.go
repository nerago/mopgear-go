package items

import (
	"strings"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
)

// /////////////////////////////////////////////////////////////
type FullItem struct {
	fullItemStatic
	fullItemInstance
	statBase    stats.StatBlock // constant stats post reforge
	statEnchant stats.StatBlock // stats added from gems, enchant, or trinket model
	total       stats.StatBlock // constant total stats as they contribute to caps
}

type fullItemStatic struct {
	itemId      ItemId
	slot        SlotItem
	armorType   stats.ArmorType
	primaryStat stats.PrimaryStatType
	phase       int8
	baseName    string
	socketSlots []stats.SocketType
	socketBonus stats.StatBlock
}

type fullItemInstance struct {
	itemLevel     uint16
	upgradeLevel  UpgradeLevel
	reforge       stats.ReforgeRecipe
	gemChoice     []stats.GemInfo
	enchantChoice uint32
	randomSuffix  RandomSuffix
	tagName       string
}

//goland:noinspection GoFixEmbedLit
func FullItem_FromWowSim(itemId ItemId, itemLevel uint16, upgradeLevel UpgradeLevel, slot SlotItem, baseName string, statBase stats.StatBlock, armorType stats.ArmorType, socketSlots []stats.SocketType, socketBonus stats.StatBlock, phase int8) FullItem {
	return FullItem{
		fullItemStatic: fullItemStatic{
			itemId, slot, armorType, statBase.PrimaryStat(),
			phase, baseName, socketSlots, socketBonus},
		fullItemInstance: fullItemInstance{
			itemLevel: itemLevel, upgradeLevel: upgradeLevel,
		},
		statBase: statBase,
		total:    statBase,
	}
}

func FullItem_ForTest(itemId ItemId, slot SlotItem, statBase stats.StatBlock) FullItem {
	return FullItem{
		itemId: itemId, itemLevel: 600, upgradeLevel: 1,
		slot: slot, baseName: slot.Name(), armorType: stats.Armor_None,
		primaryStat: statBase.PrimaryStat(),
		statBase:    statBase, total: statBase}
}

func (item *FullItem) NewWithChangedStatsReforge(changeStats stats.StatBlock, changeReforge stats.ReforgeRecipe) *FullItem {
	newItem := &FullItem{
		fullItemStatic:   item.fullItemStatic,
		fullItemInstance: item.fullItemInstance,
		statBase:         changeStats,
		statEnchant:      item.statEnchant,
	}
	newItem.fullItemInstance.reforge = changeReforge
	newItem.updateDerivedStatTotal()
	return newItem
}

func (item *FullItem) newWithChangedStatsSuffix(changeStats stats.StatBlock, randomSuffix RandomSuffix) *FullItem {
	newItem := &FullItem{
		fullItemStatic:   item.fullItemStatic,
		fullItemInstance: item.fullItemInstance,
		statBase:         changeStats,
		statEnchant:      item.statEnchant,
	}
	newItem.fullItemInstance.randomSuffix = randomSuffix
	newItem.updateDerivedStatTotal()
	return newItem
}

func (item *FullItem) NewWithEnchantDetails(socketSlots []stats.SocketType, gemChoice []stats.GemInfo, enchantChoice uint32, statEnchant stats.StatBlock) *FullItem {
	instance := fullItemInstance{
		item.itemLevel,
		item.upgradeLevel,
		item.reforge,
		gemChoice,
		enchantChoice,
		item.randomSuffix,
		item.tagName,
	}
	newItem := &FullItem{
		fullItemStatic:   item.fullItemStatic,
		fullItemInstance: instance,
		statBase:         item.statBase,
		statEnchant:      statEnchant,
	}
	newItem.fullItemStatic.socketSlots = socketSlots
	newItem.updateDerivedStatTotal()
	return newItem
}

func (item *FullItem) NewWithInstanceDetails(socketSlots []stats.SocketType, reforge stats.ReforgeRecipe, gemChoice []stats.GemInfo, enchantChoice uint32, randomSuffix RandomSuffix, statEnchant stats.StatBlock) *FullItem {
	instance := fullItemInstance{
		item.itemLevel,
		item.upgradeLevel,
		reforge,
		gemChoice,
		enchantChoice,
		randomSuffix,
		"",
	}
	newItem := &FullItem{
		fullItemStatic:   item.fullItemStatic,
		fullItemInstance: instance,
		statBase:         item.statBase,
		statEnchant:      statEnchant,
	}
	newItem.fullItemStatic.socketSlots = socketSlots
	newItem.updateDerivedStatTotal()
	return newItem
}

func (item *FullItem) SetNameTag(add string) {
	item.tagName += add
}

func (item *FullItem) updateDerivedStatTotal() {
	stats.StatBlock_Add_Into(&item.statBase, &item.statEnchant, &item.total)
}

func (item *FullItem) IsEmpty() bool {
	return item.itemId == 0
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
	if item.tagName != "" {
		build.WriteRune(' ')
		build.WriteString(item.tagName)
	}
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

func (item *FullItem) UpgradeLevel() UpgradeLevel {
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

func (item *FullItem) RandomSuffix() RandomSuffix {
	return item.randomSuffix
}

func (item *FullItem) Equals(other *FullItem) bool {
	return item.itemId == other.itemId && item.itemLevel == other.itemLevel && item.slot == other.slot &&
		stats.StatBlock_Equals(&item.statBase, &other.statBase) && stats.StatBlock_Equals(&item.statEnchant, &other.statEnchant) &&
		item.randomSuffix == other.randomSuffix
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

	if !item.statBase.IsEmpty() {
		item.statBase.AppendString(build)
	}

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

func (item *FullItem) MakeItemWithRandomSuffix(randomSuffix RandomSuffix) *FullItem {
	if randomSuffix != 0 {
		wowSimStats, suffix := item.randomStatsFromWowSim(randomSuffix)

		newStats := stats.StatBlock{}
		newStats.SetFromAddOthers(item.StatBase(), &wowSimStats)

		result := item.newWithChangedStatsSuffix(newStats, randomSuffix)
		result.baseName += " " + suffix
		return result
	}

	return item
}

const ReGem_GemAlternate = "re-gem-alternate"
const ReGem_GemDefault = "re-gem-default"

func (item *FullItem) HasBeenRegemmed() bool {
	return strings.Contains(item.tagName, ReGem_GemAlternate) || strings.Contains(item.tagName, ReGem_GemDefault)
}
