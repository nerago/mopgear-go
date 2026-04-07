package loaders

import (
	"paladin_gearing_go/items"
	"slices"
)

type EquippedItem struct {
	ItemId        items.ItemId
	GemChoice     []uint32
	EnchantChoice uint32
	RandomSuffix  int32
	UpgradeStep   int8
	Reforging     uint16
}

func (eqi *EquippedItem) Equals(other *EquippedItem) bool {
	return eqi.ItemId == other.ItemId &&
		slices.Equal(eqi.GemChoice, other.GemChoice) &&
		eqi.EnchantChoice == other.EnchantChoice &&
		eqi.RandomSuffix == other.RandomSuffix &&
		eqi.UpgradeStep == other.UpgradeStep &&
		eqi.Reforging == other.Reforging
}
