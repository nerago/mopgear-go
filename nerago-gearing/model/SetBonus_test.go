package model

import (
	"fmt"
	"math/rand/v2"
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

func makeEquipFetch2(a, b *SetBonus) loopedFetch {
	allInfo := []setInfo{}
	allInfo = append(allInfo, a.activeSets...)
	allInfo = append(allInfo, b.activeSets...)

	fetch := loopedFetch{}
	for index1 := range allInfo {
		for index2 := index1 + 1; index2 < len(allInfo); index2++ {
			for x := range 6 {
				for y := range 6 {
					fetch.add(makeEquipForBonus2(&allInfo[index1], &allInfo[index2], x, y))
				}
			}
		}
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

// func BenchmarkCalcBonusSolveC(test *testing.B) {
// 	a, b, c := makeSetBonuses()
// 	var v float32

// 	equipFetch := makeEquipFetch()

// 	for test.Loop() {
// 		equip := equipFetch.next()
// 		v += a.CalcBonusSolveC(equip)
// 		v += b.CalcBonusSolveC(equip)
// 		v += c.CalcBonusSolveC(equip)
// 	}
// 	resultFloat = v
// }
// func BenchmarkCalcBonusSolveC0(test *testing.B) {
// 	a, b, c := makeSetBonuses()
// 	var v float32

// 	equipFetch := makeEquipFetch()

// 	for test.Loop() {
// 		equip := equipFetch.next()
// 		v += a.CalcBonusSolveC0(equip)
// 		v += b.CalcBonusSolveC0(equip)
// 		v += c.CalcBonusSolveC0(equip)
// 	}
// 	resultFloat = v
// }

// func BenchmarkCalcBonusGeneric(test *testing.B) {
// 	a, b, c := makeSetBonuses()
// 	var v float32

// 	equipFetch := makeEquipFetch()

// 	for test.Loop() {
// 		equip := equipFetch.next()
// 		v += a.CalcBonusGeneric(equip)
// 		v += b.CalcBonusGeneric(equip)
// 		v += c.CalcBonusGeneric(equip)
// 	}
// 	resultFloat = v
// }

// func BenchmarkCalcBonusSolveUseAssem(test *testing.B) {
// 	a, b, c := makeSetBonuses()
// 	var v float32

// 	equipFetch := makeEquipFetch()

// 	for test.Loop() {
// 		equip := equipFetch.next()
// 		v += a.CalcBonusSolveUseAssem(equip)
// 		v += b.CalcBonusSolveUseAssem(equip)
// 		v += c.CalcBonusSolveUseAssem(equip)
// 	}
// 	resultFloat = v
// }
// func BenchmarkCalcBonusSolveAssemAssumeNonNull(test *testing.B) {
// 	a, b, c := makeSetBonuses()
// 	var v float32

// 	equipFetch := makeEquipFetch()

// 	for test.Loop() {
// 		equip := equipFetch.next()
// 		v += a.CalcBonusSolveAssemAssumeNonNull(equip)
// 		v += b.CalcBonusSolveAssemAssumeNonNull(equip)
// 		v += c.CalcBonusSolveAssemAssumeNonNull(equip)
// 	}
// 	resultFloat = v
// }

// func BenchmarkCalcBonusSolveAssemAssumeNonNullWithCases(test *testing.B) {
// 	a, b, c := makeSetBonuses()
// 	var v float32

// 	equipFetch := makeEquipFetch()

// 	for test.Loop() {
// 		equip := equipFetch.next()
// 		v += a.CalcBonusSolveAssemAssumeNonNullWithCases(equip)
// 		v += b.CalcBonusSolveAssemAssumeNonNullWithCases(equip)
// 		v += c.CalcBonusSolveAssemAssumeNonNullWithCases(equip)
// 	}
// 	resultFloat = v
// }

func TestCalcBonusCompared(test *testing.T) {
	a, b, c := makeSetBonuses()

	equipFetch := makeEquipFetch2(&b, &c)

	sets := []*SetBonus{&a, &b, &c}

	impls := []func(*SetBonus, *SolvableEquipMap) float32{
		(*SetBonus).CalcBonusSolve,
		// (*SetBonus).CalcBonusSolveUseAssem,
		// (*SetBonus).CalcBonusSolveAssemAssumeNonNull,
		// (*SetBonus).CalcBonusSolveAssemAssumeNonNullWithCases,
		// func(s *SetBonus, e *SolvableEquipMap) float32 { return s.CalcBonusGeneric(e) },
	}

	for range 1000 {
		equip := equipFetch.next()

		for t, set := range sets {
			var expect float32
			for i, call := range impls {
				val := call(set, equip)
				fmt.Printf("set=%d func=%d %f\n", t, i, val)
				if i == 0 {
					expect = val
				} else if val != expect {
					test.Fatalf("mismatched results")
				} else {
					test.Log("ok")
				}
			}
		}
	}
}

var g_setBonusSlots = [5]SlotEquip{Equip_Head, Equip_Shoulder, Equip_Chest, Equip_Hand, Equip_Leg}

func makeEquipForBonus(numInSet int) *SolvableEquipMap {
	equip := SolvableEquipMap{}
	for slot := range equip {
		equip[slot] = makeItem()
	}

	setItems := []uint32{86659, 86660, 86661, 86662, 86663}
	for i := range numInSet {
		slot := g_setBonusSlots[i]
		equip[slot] = makeItemForBonus(setItems[i])
	}

	return &equip
}

func makeEquipForBonus2(a, b *setInfo, x, y int) *SolvableEquipMap {
	equip := SolvableEquipMap{}
	for slot := range equip {
		equip[slot] = makeItem()
	}

	for _, slot := range g_setBonusSlots {
		if x > 0 {
			id := randChoice(a.items)
			equip[slot] = makeItemForBonus(id)
			x--
		} else if y > 0 {
			id := randChoice(b.items)
			equip[slot] = makeItemForBonus(id)
			y--
		}
	}

	return &equip
}

func randChoice(slice []uint32) uint32 {
	return slice[rand.IntN(len(slice))]
}

func makeItemForBonus(id uint32) *SolvableItem {
	item := SolvableItem_ForTest(ItemId(id), randStatBlock())
	return &item
}
