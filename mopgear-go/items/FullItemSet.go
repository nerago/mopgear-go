package items

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
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

func FullItemSet_FromSolved(solvedSet SolvableItemSet, optionsMap *FullOptionsMap) (FullItemSet, error) {
	fullMap := FullEquipMap{}
	for slot, solveItem := range solvedSet.items {
		if solveItem != nil {
			fullItem, err := findMatch(optionsMap[slot], solveItem)
			if err != nil {
				return FullItemSet{}, err
			}
			fullMap[slot] = fullItem
		}
	}
	return FullItemSet{items: fullMap, total: solvedSet.total}, nil
}

func findMatch(fullItem []FullItem, solveItem *SolvableItem) (*FullItem, error) {
	for i := range fullItem {
		item := &fullItem[i]
		if isMatch(item, solveItem) {
			return item, nil
		}
	}
	return nil, util.ErrorTracedNew("match not found")
}

func (itemSet *FullItemSet) PrintStats(printer util.Printable) {
	printer.Printf("STATS %s\n", itemSet.total.CreateString())
}

func (itemSet *FullItemSet) Total() *stats.StatBlock {
	return &itemSet.total
}

func (itemSet *FullItemSet) Items() *FullEquipMap {
	return &itemSet.items
}

func (itemSet *FullItemSet) Equals(other *FullItemSet) bool {
	return stats.StatBlock_Equals(&itemSet.total, &other.total) && itemSet.items.Equals(&other.items)
}

func (itemSet *FullItemSet) Clone() FullItemSet {
	return FullItemSet{
		itemSet.items,
		itemSet.total,
	}
}

func (itemSet *FullItemSet) ValidateItemRules() error {
	weapon := itemSet.items.Get(Equip_Weapon)
	if weapon == nil {
		return util.ErrorTracedNew("no weapon in set")
	} else if weapon.slot == Item_Weapon2H && itemSet.items.Has(Equip_Offhand) {
		return util.ErrorTracedNew("weapon 2H with unexpected offhand")
	} else if weapon.slot == Item_Weapon1H && !itemSet.items.Has(Equip_Offhand) {
		return util.ErrorTracedNew("weapon 1H with missing offhand")
	}

	if err := checkPairedSlotNoDuplicate(itemSet.items.Get(Equip_Ring1), itemSet.items.Get(Equip_Ring2)); err != nil {
		return err
	}
	if err := checkPairedSlotNoDuplicate(itemSet.items.Get(Equip_Trinket1), itemSet.items.Get(Equip_Trinket2)); err != nil {
		return err
	}

	for slot := Equip_Head; slot <= Equip_Weapon; slot++ {
		if !itemSet.items.Has(slot) {
			return util.ErrorTracedNew("unexpected empty slot " + slot.Name())
		}
	}

	return nil
}

func checkPairedSlotNoDuplicate(a, b *FullItem) error {
	if a != nil && b != nil {
		if a.ItemId() == b.ItemId() {
			return util.ErrorTracedNew("duplicate item " + a.CreateString())
		} else if UniqueEquipViolation(a, b) {
			return util.ErrorTracedNew("unique equipped violation:\n" + a.CreateString() + "\n" + b.CreateString())
		}
	}
	return nil
}
