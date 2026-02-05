package types

const LOW_HIGH_MOP_ITEM_LEVELS_THRESHOLD = 522
const LOW_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL = 4
const HIGH_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL = 3
const MAX_UPGRADE_LEVEL = 2

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

type SlotEquip int8

const (
	Equip_Head     SlotEquip = iota
	Equip_Neck     SlotEquip = iota
	Equip_Shoulder SlotEquip = iota
	Equip_Back     SlotEquip = iota
	Equip_Chest    SlotEquip = iota
	Equip_Wrist    SlotEquip = iota
	Equip_Hand     SlotEquip = iota
	Equip_Belt     SlotEquip = iota
	Equip_Leg      SlotEquip = iota
	Equip_Foot     SlotEquip = iota
	Equip_Ring1    SlotEquip = iota
	Equip_Ring2    SlotEquip = iota
	Equip_Trinket1 SlotEquip = iota
	Equip_Trinket2 SlotEquip = iota
	Equip_Weapon   SlotEquip = iota
	Equip_Offhand  SlotEquip = iota
)

func toSlotEquipOptions(slot SlotItem) []SlotEquip {
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
		panic("unknown slot " + string(slot))
	}
}

func pairedSlot(slot SlotEquip) SlotEquip {
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

func addEnchantToCap(slot SlotItem) bool {
	return slot != Item_Trinket
}

func possibleBlacksmith(slot SlotItem) bool {
	return slot == Item_Wrist || slot == Item_Hand || slot == Item_Belt
}

type ArmorType int8

const (
	Armor_None    ArmorType = -1
	Armor_Cloth             = 1
	Armor_Leather           = 2
	Armor_Mail              = 3
	Armor_Plate             = 4
)

type SocketType int8

const (
	Socket_Meta        SocketType = 1
	Socket_Red                    = 2
	Socket_Blue                   = 3
	Socket_Yellow                 = 4
	Socket_General                = 8
	Socket_Engineering            = 9
	Socket_Sha                    = 10
)

type PrimaryStatType int8

const (
	PrimaryStat_None      PrimaryStatType = iota
	PrimaryStat_Strength                  = iota
	PrimaryStat_Agility                   = iota
	PrimaryStat_Intellect                 = iota
)

type ItemRef struct {
	itemId        int64
	itemLevel     int16
	itemLevelBase int16
}

func upgradeLevel(item *ItemRef) int16 {
	if item.itemLevelBase < LOW_HIGH_MOP_ITEM_LEVELS_THRESHOLD {
		return (item.itemLevel - item.itemLevelBase) / LOW_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL
	} else {
		return (item.itemLevel - item.itemLevelBase) / HIGH_MOP_ITEM_LEVELS_PER_UPGRADE_LEVEL
	}
}

type ItemShared struct {
	ref         *ItemRef
	slot        SlotItem
	baseName    string
	armorType   ArmorType
	primaryStat PrimaryStatType
	socketSlots []SocketType
	socketBonus StatBlock
	phase       int8
}

type ReforgeRecipe struct {
	from, to StatType
}

type FullItemData struct {
	shared        *ItemShared
	reforge       ReforgeRecipe
	statBase      StatBlock
	statEnchant   StatBlock
	totalCap      StatBlock
	totalRated    StatBlock
	gemChoice     []int32
	enchantChoice int32
	randomSuffix  int32
}

func FullItemData_fromWowSim(ref *ItemRef, slot SlotItem, baseName string, statBase StatBlock,
	armorType ArmorType, socketSlots []SocketType, socketBonus *StatBlock,
	phase int8) {
	return FullItemData(shared, nil, statBase, StatBlock_empty)
}

func primaryStat(item *FullItemData) PrimaryStatType {
	str := item.statBase[Stat_Strength] != 0
	agi := item.statBase[Stat_Agility] != 0
	itl := item.statBase[Stat_Intellect] != 0

	primaryCount := 0
	if str {
		primaryCount++
	}
	if agi {
		primaryCount++
	}
	if itl {
		primaryCount++
	}

	if primaryCount > 1 {
		panic("conflicting primary stats")
	} else if primaryCount == 0 {
		return PrimaryStat_None
	} else if str {
		return PrimaryStat_Strength
	} else if agi {
		return PrimaryStat_Agility
	} else {
		return PrimaryStat_Intellect
	}
}
