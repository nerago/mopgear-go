package model

import (
	"fmt"
	"math/rand/v2"
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
)

func MainModel() {
	bonus := SetBonus_Named("White Tiger Plate")
	equip := _makeEquipForBonus(3)
	v := bonus.CalcBonusSolveUseAssem(equip)
	fmt.Printf("result = %f", v)
}

var _setBonusSlots = [5]items.SlotEquip{items.Equip_Head, items.Equip_Shoulder, items.Equip_Chest, items.Equip_Hand, items.Equip_Leg}

func _makeEquipForBonus(numInSet int) *items.SolvableEquipMap {
	equip := items.SolvableEquipMap{}
	for slot := range equip {
		equip[slot] = _makeItem()
	}

	setItems := []uint32{86659, 86660, 86661, 86662, 86663}
	// numInSet := rand.IntN(6)
	for i := range numInSet {
		slot := _setBonusSlots[i]
		equip[slot] = _makeItemForBonus(setItems[i])
	}

	return &equip
}

func _makeItemForBonus(id uint32) *items.SolvableItem {
	item := items.SolvableItem_ForTest(id, _randStatBlock())
	return &item
}

func _makeItem() *items.SolvableItem {
	id := rand.Uint32N(10000)
	item := items.SolvableItem_ForTest(id, _randStatBlock())
	return &item
}

func _randStatBlock() stats.StatBlock {
	block := stats.StatBlock{}
	for i := range block {
		block[i] = rand.Uint32()
	}
	return block
}
