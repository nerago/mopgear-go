package bonus_set

import (
	"iter"
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
	"slices"
)

var g_itemIdLookup [c_maxItemId]*BonusLookup
var g_setNameLookup map[string]*BonusLookup

type BonusLookup struct {
	name  string
	items []items.ItemId
}

func (lk *BonusLookup) Equals(other *BonusLookup) bool {
	return lk.name == other.name && slices.Equal(lk.items, other.items)
}

func (lk *BonusLookup) Name() string {
	return lk.name
}

func (lk *BonusLookup) ItemIdSeq() iter.Seq[items.ItemId] {
	return slices.Values(lk.items)
}

func (lk *BonusLookup) IncludesItem(itemId items.ItemId) bool {
	return slices.Contains(lk.items, itemId)
}

func (lk *BonusLookup) ContainsItemSolve(item *items.SolvableItem) bool {
	if item != nil {
		return slices.Contains(lk.items, item.ItemId())
	}
	return false
}

func (lk *BonusLookup) ContainsItemFull(item *items.FullItem) bool {
	if item != nil {
		return slices.Contains(lk.items, item.ItemId())
	}
	return false
}

func (lk *BonusLookup) CountItemsSolve(equip *items.SolvableEquipMap) uint8 {
	return countItemsGeneric(equip, lk.items)
}

func (lk *BonusLookup) CountItemsFull(equip *items.FullEquipMap) uint8 {
	return countItemsGeneric(equip, lk.items)
}

func countItemsGeneric[T items.IEquipMap[M], M items.IItem](equip T, lookupItems []items.ItemId) uint8 {
	count := uint8(0)

	item := equip.Get(items.Equip_Head)
	if item != nil && slices.Contains(lookupItems, item.ItemId()) {
		count++
	}

	item = equip.Get(items.Equip_Shoulder)
	if item != nil && slices.Contains(lookupItems, item.ItemId()) {
		count++
	}

	item = equip.Get(items.Equip_Chest)
	if item != nil && slices.Contains(lookupItems, item.ItemId()) {
		count++
	}

	item = equip.Get(items.Equip_Hand)
	if item != nil && slices.Contains(lookupItems, item.ItemId()) {
		count++
	}

	item = equip.Get(items.Equip_Leg)
	if item != nil && slices.Contains(lookupItems, item.ItemId()) {
		count++
	}

	return count
}

func init() {
	g_setNameLookup = make(map[string]*BonusLookup, len(g_setData))
	for _, data := range g_setData {
		info := &BonusLookup{
			data.variants[0].name,
			data.items,
		}
		for i := 1; i < len(data.variants); i++ {
			name := data.variants[i].name
			g_setNameLookup[name] = &BonusLookup{
				name,
				data.items,
			}
		}
		g_setNameLookup[info.name] = info

		for _, itemId := range data.items {
			g_itemIdLookup[itemId] = info
		}
	}
}

func AllBonusesText(equipMap *items.FullEquipMap) string {
	counts := make(map[string]uint32)
	for item := range equipMap.AllItemSeq() {
		set := g_itemIdLookup[item.ItemId()]
		if set != nil {
			counts[set.name]++
		}
	}

	build := util.StringBuild2{}
	for name, num := range counts {
		if build.Len() > 0 {
			build.WriteString(", ")
		}
		build.WriteString(name)
		build.WriteString(" => ")
		build.WriteUint32(num)
	}
	return build.String()
}

func IsAnyKnownItem(itemId items.ItemId) bool {
	return g_itemIdLookup[itemId] != nil
}
