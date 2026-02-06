package items

import . "paladin_gearing_go/types/common"

type SkinnyItem struct {
	slot SlotEquip
	a    uint32
	b    uint32
	c    uint32
}

type SkinnyEquipMap [16]SkinnyItem

type SkinnyItemSet struct {
	items [16]SkinnyItem
	a     uint32
	b     uint32
	c     uint32
}
