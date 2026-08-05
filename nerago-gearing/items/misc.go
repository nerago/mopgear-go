package items

import (
	"paladin_gearing_go/util"
	"strconv"
)

const (
	MOP_DUNGEON_ITEM_LEVELS_THRESHOLD                   = 463
	LOW_HIGH_MOP_ITEM_LEVELS_THRESHOLD                  = 522
	LOW_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL               = 4
	HIGH_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL              = 3
	MAX_UPGRADE_LEVEL                      UpgradeLevel = 2
	NO_RANDOM_SUFFIX                       RandomSuffix = 0
)

type UpgradeLevel int8

type RandomSuffix int32

type ItemId uint32

func (id ItemId) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

func CalcUpgradeLevel(itemLevel uint16, itemLevelBase uint16) UpgradeLevel {
	diff := int16(itemLevel) - int16(itemLevelBase)
	if diff < 0 {
		return -1
	} else if diff == 0 {
		return 0
	} else if itemLevelBase <= MOP_DUNGEON_ITEM_LEVELS_THRESHOLD && diff == 8 {
		return 1
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

type ItemRef struct {
	ItemId       ItemId
	RandomSuffix RandomSuffix
}

func ItemRef_Of(item *FullItem) ItemRef {
	return ItemRef{ItemId: item.ItemId(), RandomSuffix: item.RandomSuffix()}
}

func (ref ItemRef) String() string {
	build := util.StringBuild2{}
	build.WriteInt32(int32(ref.ItemId))
	if ref.RandomSuffix != 0 {
		build.WriteString("[")
		build.WriteInt32(int32(ref.RandomSuffix))
		build.WriteString("]")
	}
	return build.String()
}
