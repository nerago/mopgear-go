package items

type SkinnyItem struct {
	A      uint32
	B      uint32
	Exists bool // just for zero value in collections
}

type SkinnyEquipMap [16]SkinnyItem

type SkinnyItemSet struct {
	Items SkinnyEquipMap
	A     uint32
	B     uint32
}
