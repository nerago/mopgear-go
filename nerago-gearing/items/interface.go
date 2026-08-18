package items

import (
	"github.com/nerago/mopgear-go/stats"
	"iter"
)

type IItem interface {
	*FullItem | *SolvableItem
	ItemId() ItemId
	IsEmpty() bool
	Total() *stats.StatBlock
}

type IEquipMap[M IItem] interface {
	*FullEquipMap | *SolvableEquipMap
	Get(slot SlotEquip) M
	Has(slot SlotEquip) bool
	GetAsId(slot SlotEquip) ItemId
	AllItemSeq() iter.Seq[M]
}

type IItemSet[E IEquipMap[M], M IItem] interface {
	*FullItemSet | *SolvableItemSet
	Total() *stats.StatBlock
	Items() E
}
