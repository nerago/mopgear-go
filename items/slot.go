package items

import (
	"strconv"
)

type SlotItem int8

const (
	Item_Head     SlotItem = 1
	Item_Neck     SlotItem = 2
	Item_Shoulder SlotItem = 3
	Item_Back     SlotItem = 16
	Item_Chest    SlotItem = 5 // also 20
	Item_Wrist    SlotItem = 9
	Item_Hand     SlotItem = 10
	Item_Belt     SlotItem = 6
	Item_Leg      SlotItem = 7
	Item_Foot     SlotItem = 8
	Item_Ring     SlotItem = 11
	Item_Trinket  SlotItem = 12
	Item_Weapon2H SlotItem = 17
	Item_Weapon1H SlotItem = 13
	Item_Offhand  SlotItem = 14 // also 23
)

func (slot SlotItem) Name() string {
	switch slot {
	case Item_Back:
		return "Back"
	case Item_Belt:
		return "Belt"
	case Item_Chest:
		return "Chest"
	case Item_Foot:
		return "Foot"
	case Item_Hand:
		return "Hand"
	case Item_Head:
		return "Head"
	case Item_Leg:
		return "Leg"
	case Item_Neck:
		return "Neck"
	case Item_Offhand:
		return "Offhand"
	case Item_Ring:
		return "Ring"
	case Item_Shoulder:
		return "Shoulder"
	case Item_Trinket:
		return "Trinket"
	case Item_Weapon1H:
		return "Weapon1H"
	case Item_Weapon2H:
		return "Weapon2H"
	case Item_Wrist:
		return "Wrist"
	default:
		panic("unexpected common.SlotItem")
	}
}

type SlotEquip int8

const (
	Equip_Head     SlotEquip = iota
	Equip_Neck     SlotEquip = iota
	Equip_Shoulder SlotEquip = iota
	Equip_Back     SlotEquip = iota

	Equip_Chest SlotEquip = iota
	Equip_Wrist SlotEquip = iota
	Equip_Hand  SlotEquip = iota
	Equip_Belt  SlotEquip = iota

	Equip_Leg   SlotEquip = iota
	Equip_Foot  SlotEquip = iota
	Equip_Ring1 SlotEquip = iota
	Equip_Ring2 SlotEquip = iota

	Equip_Trinket1 SlotEquip = iota
	Equip_Trinket2 SlotEquip = iota
	Equip_Weapon   SlotEquip = iota
	Equip_Offhand  SlotEquip = iota
)

func (slot SlotItem) ToSlotEquipOptions() []SlotEquip {
	switch slot {
	case Item_Head:
		return []SlotEquip{Equip_Head}
	case Item_Neck:
		return []SlotEquip{Equip_Neck}
	case Item_Shoulder:
		return []SlotEquip{Equip_Shoulder}
	case Item_Back:
		return []SlotEquip{Equip_Back}
	case Item_Chest:
		return []SlotEquip{Equip_Chest}
	case Item_Wrist:
		return []SlotEquip{Equip_Wrist}
	case Item_Hand:
		return []SlotEquip{Equip_Hand}
	case Item_Belt:
		return []SlotEquip{Equip_Belt}
	case Item_Leg:
		return []SlotEquip{Equip_Leg}
	case Item_Foot:
		return []SlotEquip{Equip_Foot}
	case Item_Ring:
		return []SlotEquip{Equip_Ring1, Equip_Ring1}
	case Item_Trinket:
		return []SlotEquip{Equip_Trinket1, Equip_Trinket2}
	case Item_Weapon2H:
		return []SlotEquip{Equip_Weapon}
	case Item_Weapon1H:
		return []SlotEquip{Equip_Weapon}
	case Item_Offhand:
		return []SlotEquip{Equip_Offhand}
	default:
		panic("unknown slot " + strconv.Itoa(int(slot)))
	}
}

func (slot SlotEquip) PairedSlot() SlotEquip {
	switch slot {
	case Equip_Ring1:
		return Equip_Ring2
	case Equip_Ring2:
		return Equip_Ring1
	case Equip_Trinket1:
		return Equip_Trinket2
	case Equip_Trinket2:
		return Equip_Trinket1
	default:
		return -1
	}
}

func (slot SlotItem) AddEnchantToCap() bool {
	return slot != Item_Trinket
}

func (slot SlotItem) PossibleBlacksmith() bool {
	return slot == Item_Wrist || slot == Item_Hand || slot == Item_Belt
}
