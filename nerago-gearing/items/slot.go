package items

import "paladin_gearing_go/util/util_collection"

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

func (slot SlotItem) CanEnchant() bool {
	switch slot {
	case Item_Shoulder:
		return true
	case Item_Back:
		return true
	case Item_Chest:
		return true
	case Item_Wrist:
		return true
	case Item_Hand:
		return true
	case Item_Leg:
		return true
	case Item_Foot:
		return true
	case Item_Weapon2H:
		return true
	case Item_Weapon1H:
		return true
	case Item_Offhand:
		return true
	default:
		return false
	}
}

type SlotEquip uint8

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

	Equip_Iter_First  = Equip_Head
	Equip_Iter_Last   = Equip_Offhand
	SlotEquip_Count   = 16
	SlotEquip_Invalid = 255
)

func (slot SlotEquip) Name() string {
	switch slot {
	case Equip_Back:
		return "Back"
	case Equip_Belt:
		return "Belt"
	case Equip_Chest:
		return "Chest"
	case Equip_Foot:
		return "Foot"
	case Equip_Hand:
		return "Hand"
	case Equip_Head:
		return "Head"
	case Equip_Leg:
		return "Leg"
	case Equip_Neck:
		return "Neck"
	case Equip_Offhand:
		return "Offhand"
	case Equip_Ring1:
		return "Ring1"
	case Equip_Ring2:
		return "Ring2"
	case Equip_Shoulder:
		return "Shoulder"
	case Equip_Trinket1:
		return "Trinket1"
	case Equip_Trinket2:
		return "Trinket2"
	case Equip_Weapon:
		return "Weapon"
	case Equip_Wrist:
		return "Wrist"
	default:
		panic("unexpected common.SlotItem")
	}
}

var SlotEquip_List = []SlotEquip{
	Equip_Head, Equip_Neck, Equip_Shoulder, Equip_Back,
	Equip_Chest, Equip_Wrist, Equip_Hand, Equip_Belt,
	Equip_Leg, Equip_Foot, Equip_Ring1, Equip_Ring2,
	Equip_Trinket1, Equip_Trinket2, Equip_Weapon, Equip_Offhand,
}

var SlotEquipEnum = util_collection.EnumTypeMake[SlotEquip](SlotEquip_List)

func (slot SlotEquip) EnumNumValues() uint8 {
	return SlotEquip_Count
}

type SlotEquipMap[V any] struct {
	util_collection.EnumMapTiny[SlotEquip, V, [SlotEquip_Count]V]
}

var equipOptionsMap [20][]SlotEquip = makeToSlotEquipOptions()

func makeToSlotEquipOptions() [20][]SlotEquip {
	var opts [20][]SlotEquip
	opts[Item_Head] = []SlotEquip{Equip_Head}
	opts[Item_Neck] = []SlotEquip{Equip_Neck}
	opts[Item_Shoulder] = []SlotEquip{Equip_Shoulder}
	opts[Item_Back] = []SlotEquip{Equip_Back}
	opts[Item_Chest] = []SlotEquip{Equip_Chest}
	opts[Item_Wrist] = []SlotEquip{Equip_Wrist}
	opts[Item_Hand] = []SlotEquip{Equip_Hand}
	opts[Item_Belt] = []SlotEquip{Equip_Belt}
	opts[Item_Leg] = []SlotEquip{Equip_Leg}
	opts[Item_Foot] = []SlotEquip{Equip_Foot}
	opts[Item_Ring] = []SlotEquip{Equip_Ring1, Equip_Ring2}
	opts[Item_Trinket] = []SlotEquip{Equip_Trinket1, Equip_Trinket2}
	opts[Item_Weapon2H] = []SlotEquip{Equip_Weapon}
	opts[Item_Weapon1H] = []SlotEquip{Equip_Weapon}
	opts[Item_Offhand] = []SlotEquip{Equip_Offhand}
	return opts
}

func (slot SlotItem) ToSlotEquipOptions() []SlotEquip {
	return equipOptionsMap[slot]
}

func (slot SlotItem) ForEachEquip(apply func(equip SlotEquip)) {
	switch slot {
	case Item_Head:
		apply(Equip_Head)
	case Item_Neck:
		apply(Equip_Neck)
	case Item_Shoulder:
		apply(Equip_Shoulder)
	case Item_Back:
		apply(Equip_Back)
	case Item_Chest:
		apply(Equip_Chest)
	case Item_Wrist:
		apply(Equip_Wrist)
	case Item_Hand:
		apply(Equip_Hand)
	case Item_Belt:
		apply(Equip_Belt)
	case Item_Leg:
		apply(Equip_Leg)
	case Item_Foot:
		apply(Equip_Foot)
	case Item_Ring:
		apply(Equip_Ring1)
		apply(Equip_Ring2)
	case Item_Trinket:
		apply(Equip_Trinket1)
		apply(Equip_Trinket2)
	case Item_Weapon2H:
		apply(Equip_Weapon)
	case Item_Weapon1H:
		apply(Equip_Weapon)
	case Item_Offhand:
		apply(Equip_Offhand)
	}
}

var PairedSlotList = []SlotEquip{Equip_Ring1, Equip_Ring2, Equip_Trinket1, Equip_Trinket2}

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
		return SlotEquip_Invalid
	}
}

func (slot SlotItem) AddEnchantToCap() bool {
	return slot != Item_Trinket
}

func (slot SlotItem) PossibleBlacksmith() bool {
	return slot == Item_Wrist || slot == Item_Hand
}

func (slot SlotItem) AlwaysBlacksmith() bool {
	return slot == Item_Belt
}
