package loaders

import "paladin_gearing_go/items"

type EquippedItem struct {
	ItemId        items.ItemId
	GemChoice     []uint32
	EnchantChoice uint32
	RandomSuffix  int32
	UpgradeStep   int8
	Reforging     uint16
}
