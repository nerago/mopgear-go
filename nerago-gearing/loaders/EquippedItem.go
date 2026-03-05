package loaders

type EquippedItem struct {
	ItemId        uint32
	GemChoice     []uint32
	EnchantChoice uint32
	RandomSuffix  int32
	UpgradeStep   int16
	Reforging     uint16
}
