package items

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
