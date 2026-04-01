package items

const LOW_HIGH_MOP_ITEM_LEVELS_THRESHOLD = 522
const LOW_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL = 4
const HIGH_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL = 3
const MAX_UPGRADE_LEVEL = 2

type ItemId uint32

type ItemRef struct {
	ItemId       ItemId
	ItemLevel    uint16
	UpgradeLevel int8
}

func ItemRef_Make(itemId ItemId, itemLevel uint16, itemLevelBase uint16) ItemRef {
	upgrade := calcUpgradeLevel(itemLevel, itemLevelBase)
	return ItemRef{itemId, itemLevel, upgrade}
}

func ItemRef_Challenge(itemId ItemId, itemLevel uint16) ItemRef {
	return ItemRef{itemId, itemLevel, -1}
}

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

// some high items are +4, but ones that were alredy upgradeable are +7

// OLDOLDOLDOLDOLDOLDOLD

// const LOW_HIGH_MOP_ITEM_LEVELS_THRESHOLD = 522
// const LOW_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL = 4
// const HIGH_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL = 3
// const MAX_UPGRADE_LEVEL = 2

// type ItemId uint32

// type ItemRef struct {
// 	ItemId        ItemId
// 	ItemLevel     uint16
// 	ItemLevelBase uint16
// }

// func (ref ItemRef) UpgradeLevel() int16 {
// 	if ref.ItemLevel < ref.ItemLevelBase {
// 		return -1
// 	} else if ref.ItemLevelBase < LOW_HIGH_MOP_ITEM_LEVELS_THRESHOLD {
// 		return int16((ref.ItemLevel - ref.ItemLevelBase) / LOW_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL)
// 	} else {
// 		return int16((ref.ItemLevel - ref.ItemLevelBase) / HIGH_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL)
// 	}
// }
