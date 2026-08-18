package items

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/stats/extern_stats"

	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/proto"
)

func (item *FullItem) getAsWowSimItem() (simBaseItem *core.Item, simScale *proto.ScalingItemProperties) {
	simBaseItem = core.GetItemByID(int32(item.itemId))
	for _, scale := range simBaseItem.ScalingOptions {
		if scale.Ilvl == int32(item.itemLevel) {
			return simBaseItem, scale
		}
	}
	panic("level scale not found")
}

func (item *FullItem) randomStatsFromWowSim(randomSuffix RandomSuffix) (stats.StatBlock, string) {
	_, simScale := item.getAsWowSimItem()

	suffixObject, knownSuffix := core.RandomSuffixesByID[int32(randomSuffix)]
	if !knownSuffix {
		panic("unknown item suffix")
	}

	simStats := suffixObject.Stats.Multiply(float64(simScale.RandPropPoints) / 10000.).Floor()
	statBlock := extern_stats.SimStatsToGearStatBlock(simStats)
	return statBlock, suffixObject.Name
}

func MapSlotToGear(itemType, handType int32) SlotItem {
	switch itemType {
	case 1:
		return Item_Head
	case 2:
		return Item_Neck
	case 3:
		return Item_Shoulder
	case 4:
		return Item_Back
	case 5:
		return Item_Chest
	case 6:
		return Item_Wrist
	case 7:
		return Item_Hand
	case 8:
		return Item_Belt
	case 9:
		return Item_Leg
	case 10:
		return Item_Foot
	case 11:
		return Item_Ring
	case 12:
		return Item_Trinket
	case 13, 14:
		switch handType {
		case 1, 2:
			return Item_Weapon1H
		case 0, 4:
			return Item_Weapon2H
		case 3:
			return Item_Offhand
		default:
			panic("unknown weapon")
		}

	default:
		panic("unknown slot")
	}
}
