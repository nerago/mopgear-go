package items

import (
	"github.com/nerago/mopgear-go/stats"
	. "github.com/nerago/mopgear-go/stats"
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

func (item *SolvableItem) Total() *StatBlock {
	return &item.total
}

func (item *SolvableItem) Equals(other *SolvableItem) bool {
	return item.itemId == other.itemId && stats.StatBlock_Equals(&item.total, &other.total)
}

func (item *SolvableItem) EqualsFull(other *FullItem) bool {
	return item.itemId == other.ItemId() && stats.StatBlock_Equals(&item.total, &other.total)
}
