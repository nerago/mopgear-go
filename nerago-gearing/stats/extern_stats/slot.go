package extern_stats

import (
	"paladin_gearing_go/items"
)

func MapSlotToGear(itemType, handType int32) items.SlotItem {
	switch itemType {
	case 1:
		return items.Item_Head
	case 2:
		return items.Item_Neck
	case 3:
		return items.Item_Shoulder
	case 4:
		return items.Item_Back
	case 5:
		return items.Item_Chest
	case 6:
		return items.Item_Wrist
	case 7:
		return items.Item_Hand
	case 8:
		return items.Item_Belt
	case 9:
		return items.Item_Leg
	case 10:
		return items.Item_Foot
	case 11:
		return items.Item_Ring
	case 12:
		return items.Item_Trinket
	case 13, 14:
		switch handType {
		case 1, 2:
			return items.Item_Weapon1H
		case 0, 4:
			return items.Item_Weapon2H
		case 3:
			return items.Item_Offhand
		default:
			panic("unknown weapon")
		}

	default:
		panic("unknown slot")
	}
}
