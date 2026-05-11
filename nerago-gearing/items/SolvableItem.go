package items

import (
	"paladin_gearing_go/stats"
	. "paladin_gearing_go/stats"
)

// /////////////////////////////////////////////////////////////
type SolvableItem struct {
	total  StatBlock
	itemId ItemId
}

func SolvableItem_Of(item *FullItem) SolvableItem {
	return SolvableItem{
		itemId: item.itemId,
		total:  item.total}
}

func SolvableItem_ForTest(itemid ItemId, block StatBlock) SolvableItem {
	return SolvableItem{
		itemId: itemid,
		total:  block}
}

func (item *SolvableItem) ItemId() ItemId {
	return item.itemId
}

func (item *SolvableItem) IsEmpty() bool {
	return item.itemId == 0
}

func (item *SolvableItem) TotalCap() *StatBlock {
	return &item.total
}

func (item *SolvableItem) TotalRated() *StatBlock {
	return &item.total
}

func (item *SolvableItem) Equals(other *SolvableItem) bool {
	return item.itemId == other.itemId && stats.StatBlock_Equals(&item.total, &other.total)
}

func (item *SolvableItem) EqualsFull(other *FullItem) bool {
	return item.itemId == other.ItemId() && stats.StatBlock_Equals(&item.total, &other.total)
}
