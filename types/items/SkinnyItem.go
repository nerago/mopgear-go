package items

// import . "paladin_gearing_go/types/common"

type SkinnyItem struct {
	A      uint32
	B      uint32
	Exists bool // just for zero value in collections
}

type SkinnyEquipMap [16]SkinnyItem

type SkinnyItemSet struct {
	items [16]SkinnyItem
	A     uint32
	B     uint32
}
