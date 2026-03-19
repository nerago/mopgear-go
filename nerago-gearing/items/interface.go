package items

import (
	"iter"
	"paladin_gearing_go/stats"
)

type IHasStats interface {
	TotalCap() *stats.StatBlock
	TotalRated() *stats.StatBlock
}

type IItem interface {
	ItemId() ItemId
	IsEmpty() bool
	IHasStats
}

type IEquipMap interface {
	GetGeneric(SlotEquip) IItem
	GetAsId(SlotEquip) ItemId
	AllItemSeqGeneric() iter.Seq[IItem]
}

type IItemSet interface {
	ItemsGeneric() IEquipMap
	IHasStats
}

type IEquipMapArrays[Item IItem] interface {
	IEquipMap
	~[16]Item
}

type IEquipMapArraysLazy interface {
	FullEquipMap | SolvableEquipMap
}

type IItemSetArrays[Item IItem, EquipMap IEquipMapArrays[Item]] interface {
	IItemSet
	ItemsArray() EquipMap
}
