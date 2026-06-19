package loaders

import (
	"paladin_gearing_go/db"
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"slices"
)

type EquippedItem struct {
	ItemId                 items.ItemId
	GemChoice              []uint32
	EnchantChoice          uint32
	RandomSuffix           items.RandomSuffix
	UpgradeStepOrItemLevel int32
	Reforging              uint16
}

func (eqi *EquippedItem) Equals(other *EquippedItem) bool {
	return eqi.ItemId == other.ItemId &&
		slices.Equal(eqi.GemChoice, other.GemChoice) &&
		eqi.EnchantChoice == other.EnchantChoice &&
		eqi.RandomSuffix == other.RandomSuffix &&
		eqi.UpgradeStepOrItemLevel == other.UpgradeStepOrItemLevel &&
		eqi.Reforging == other.Reforging
}

func EquippedItem_FromFull(full *items.FullItem) EquippedItem {
	var reforge uint16 = 0
	if !full.Reforge().IsEmpty() {
		reforge = db.WowSimDB_ReforgeToId(full.Reforge())
	}
	return EquippedItem{
		ItemId:                 full.ItemId(),
		GemChoice:              util.MapSliceAsNew(full.GemChoice(), func(x *stats.GemInfo) uint32 { return x.Id }),
		EnchantChoice:          full.EnchantChoice(),
		RandomSuffix:           full.RandomSuffix(),
		UpgradeStepOrItemLevel: int32(full.UpgradeLevel()),
		Reforging:              reforge,
	}
}
