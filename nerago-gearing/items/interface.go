package items

import (
	"iter"
	"paladin_gearing_go/stats"
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
