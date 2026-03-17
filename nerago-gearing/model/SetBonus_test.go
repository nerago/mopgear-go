package model

import (
	. "paladin_gearing_go/items"
	. "paladin_gearing_go/stats"
	"testing"
)

type loopedFetch struct {
	items []*SolvableEquipMap
	index int
}

func (fetch *loopedFetch) add(item *SolvableEquipMap) {
	fetch.items = append(fetch.items, item)
}

func (fetch *loopedFetch) next() *SolvableEquipMap {
	result := fetch.items[fetch.index]
	fetch.index = (fetch.index + 1) % len(fetch.items)
	return result
}

func makeEquipFetch() loopedFetch {
	fetch := loopedFetch{}
	for i := range 6 {
		fetch.add(makeEquipForBonus(i))
	}
	return fetch
}

var resultFloat float32

func makeSetBonuses() (SetBonus, SetBonus, SetBonus) {
	return SetBonus_Empty(), SetBonus_Named("White Tiger Plate"), SetBonus_ForSpec(Spec_PaladinRet)
}

func BenchmarkCalcBonusSolve(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += a.CalcBonusSolve(equip)
		v += b.CalcBonusSolve(equip)
		v += c.CalcBonusSolve(equip)
	}
	resultFloat = v
}

func BenchmarkCalcBonusGeneric(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += a.CalcBonusGeneric(equip)
		v += b.CalcBonusGeneric(equip)
		v += c.CalcBonusGeneric(equip)
	}
	resultFloat = v
}

func BenchmarkCalcBonusSolveUseAssem(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += a.CalcBonusSolveUseAssem(equip)
		v += b.CalcBonusSolveUseAssem(equip)
		v += c.CalcBonusSolveUseAssem(equip)
	}
	resultFloat = v
}

var g_setBonusSlots = [5]SlotEquip{Equip_Head, Equip_Shoulder, Equip_Chest, Equip_Hand, Equip_Leg}

func makeEquipForBonus(numInSet int) *SolvableEquipMap {
	equip := SolvableEquipMap{}
	for slot := range equip {
		equip[slot] = makeItem()
	}

	setItems := []uint32{86659, 86660, 86661, 86662, 86663}
	// numInSet := rand.IntN(6)
	for i := range numInSet {
		slot := g_setBonusSlots[i]
		equip[slot] = makeItemForBonus(setItems[i])
	}

	return &equip
}

func makeItemForBonus(id uint32) *SolvableItem {
	item := SolvableItem_ForTest(id, randStatBlock())
	return &item
}
