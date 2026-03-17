package modelassem

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
)

// need to repeat stuff into local package so goes in assembly header
const (
	Equip_Head     = items.Equip_Head
	Equip_Shoulder = items.Equip_Shoulder
	Equip_Chest    = items.Equip_Chest
	Equip_Hand     = items.Equip_Hand
	Equip_Leg      = items.Equip_Leg
	// offset_itemId  = 48
	offset_itemId  = 96
)

// not picking up well on the correct type from build flags

type SolvableItem items.SolvableItem

// type SolvableItem struct {
// 	total  stats.StatBlock
// 	totalb  stats.StatBlock
// 	itemId uint32
// }

// can't access private fields in Ref otherwise
type SolvableItemRefOnly struct {
	totalCap   stats.StatBlock
	totalRated stats.StatBlock
	itemId     uint32
}

//	func CalcBonusSolveRef(equip *[16]*SolvableItemRefOnly, itemToSet []int8, activeSets [][6]float32) float32 {
//		var counts [10]uint8
//		addToSpecificSet(&counts, itemToSet, equip[Equip_Head])
//		addToSpecificSet(&counts, itemToSet, equip[Equip_Shoulder])
//		addToSpecificSet(&counts, itemToSet, equip[Equip_Chest])
//		addToSpecificSet(&counts, itemToSet, equip[Equip_Hand])
//		addToSpecificSet(&counts, itemToSet, equip[Equip_Leg])
//		var value float32 = 1.0
//		for index := range activeSets {
//			value *= activeSets[index][counts[index+1]]
//		}
//		return value
//	}
//
//	func CalcBonusSolveRef(equip *[16]*items.SolvableItem, itemToSet []int8, activeSets [][6]float32) float32 {
//		var counts [10]uint8
//		addToSpecificSet2(&counts, itemToSet, equip[Equip_Head])
//		addToSpecificSet2(&counts, itemToSet, equip[Equip_Shoulder])
//		addToSpecificSet2(&counts, itemToSet, equip[Equip_Chest])
//		addToSpecificSet2(&counts, itemToSet, equip[Equip_Hand])
//		addToSpecificSet2(&counts, itemToSet, equip[Equip_Leg])
//		var value float32 = 1.0
//		for index := range activeSets {
//			value *= activeSets[index][counts[index+1]]
//		}
//		return value
//	}
func CalcBonusSolveRef(equip *items.SolvableEquipMap, itemToSet []int8, activeSets [][6]float32) float32 {
	var counts [10]uint8
	addToSpecificSet2(&counts, itemToSet, equip[Equip_Head])
	addToSpecificSet2(&counts, itemToSet, equip[Equip_Shoulder])
	addToSpecificSet2(&counts, itemToSet, equip[Equip_Chest])
	addToSpecificSet2(&counts, itemToSet, equip[Equip_Hand])
	addToSpecificSet2(&counts, itemToSet, equip[Equip_Leg])
	var value float32 = 1.0
	for index := range activeSets {
		value *= activeSets[index][counts[index+1]]
	}
	return value
}

func addToSpecificSet(counts *[10]uint8, itemToSet []int8, item *SolvableItemRefOnly) {
	if item != nil {
		entry := itemToSet[item.itemId]
		counts[entry]++
	}
}
func addToSpecificSet2(counts *[10]uint8, itemToSet []int8, item *items.SolvableItem) {
	if item != nil {
		entry := itemToSet[item.ItemId()]
		counts[entry]++
	}
}

func CalcBonusSolveAssem(equip *items.SolvableEquipMap, itemToSet []int8, activeSets []float32) float32

// func CalcBonusSolveAssem(equip *items.SolvableEquipMap, itemToSet []int8, activeSets [][6]float32) float32

func CalcBonusSolveAssemAssumeNonNull(equip *items.SolvableEquipMap, itemToSet []int8, activeSets []float32) float32
