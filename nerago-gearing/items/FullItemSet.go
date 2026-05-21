package items

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

// /////////////////////////////////////////////////////////////
type FullItemSet struct {
	items FullEquipMap
	total stats.StatBlock
}

func FullItemSet_FromMap(equipMap FullEquipMap) FullItemSet {
	itemSet := FullItemSet{equipMap, stats.StatBlock{}}
	for _, item := range equipMap {
		if item != nil {
			stats.StatBlock_Increment_Mutating(&itemSet.total, &item.total)
		}
	}
	return itemSet
}

func FullItemSet_FromSolved(solvedSet SolvableItemSet, optionsMap *FullOptionsMap) FullItemSet {
	fullMap := FullEquipMap{}
	for slot, solveItem := range solvedSet.items {
		if solveItem != nil {
			fullItem := findMatch(optionsMap[slot], solveItem)
			fullMap[slot] = fullItem
		}
	}
	return FullItemSet{items: fullMap, total: solvedSet.total}
}

func findMatch(fullItem []FullItem, solveItem *SolvableItem) *FullItem {
	for i := range fullItem {
		item := &fullItem[i]
		if isMatch(item, solveItem) {
			return item
		}
	}
	panic("match not found")
}

func (itemSet *FullItemSet) PrintStats(printer *util.PrintRecorder) {
	printer.Printf("STATS %s\n", itemSet.total.CreateString())
}

func (itemSet *FullItemSet) Total() *stats.StatBlock {
	return &itemSet.total
}

func (itemSet *FullItemSet) Items() *FullEquipMap {
	return &itemSet.items
}

func (itemSet *FullItemSet) Equals(other *FullItemSet) bool {
	return itemSet.items.Equals(&other.items)
}

func (itemSet *FullItemSet) EqualsAllowNil(other *FullItemSet) bool {
	if itemSet != nil && other != nil {
		return itemSet.items.Equals(&other.items)
	} else if itemSet == nil && other == nil {
		return true
	} else {
		return false
	}
}

func (itemSet *FullItemSet) ValidateItemRules() {
	weapon := itemSet.items.Get(Equip_Weapon)
	if weapon == nil {
		panic("no weapon in set")
	} else if weapon.slot == Item_Weapon2H && itemSet.items.Has(Equip_Offhand) {
		panic("weapon 2H with unexpected offhand")
	} else if weapon.slot == Item_Weapon1H && !itemSet.items.Has(Equip_Offhand) {
		panic("weapon 1H with missing offhand")
	}

	checkPairedSlotNoDuplicate(itemSet.items.Get(Equip_Ring1), itemSet.items.Get(Equip_Ring2))
	checkPairedSlotNoDuplicate(itemSet.items.Get(Equip_Trinket1), itemSet.items.Get(Equip_Trinket2))

	for slot := Equip_Head; slot <= Equip_Weapon; slot++ {
		if !itemSet.items.Has(slot) {
			panic("unexpected empty slot " + slot.Name())
		}
	}
}

func checkPairedSlotNoDuplicate(a, b *FullItem) {
	if a != nil && b != nil {
		if a.ItemId() == b.ItemId() {
			panic("duplicate item " + a.CreateString())
		} else if UniqueEquipViolation(a, b) {
			panic("unique equipped violation:\n" + a.CreateString() + "\n" + b.CreateString())
		}
	}
}
