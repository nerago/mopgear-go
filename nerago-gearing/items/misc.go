package items

const LOW_HIGH_MOP_ITEM_LEVELS_THRESHOLD = 522
const LOW_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL = 4
const HIGH_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL = 3
const MAX_UPGRADE_LEVEL = 2

type ItemId uint32

func calcUpgradeLevel(itemLevel uint16, itemLevelBase uint16) int8 {
	diff := int16(itemLevel) - int16(itemLevelBase)
	if diff < 0 {
		return -1
	} else if diff == 0 {
		return 0
	} else if diff == 4 || diff == 7 {
		return 1
	} else if diff == 8 || diff == 14 {
		return 2
	} else if diff == 12 {
		return 3
	} else if diff == 16 {
		return 4
	} else {
		panic("unknown upgrade level")
	}
}
