package bonus_set

import (
	"fmt"
	"math/rand/v2"
	"testing"

	. "github.com/nerago/mopgear-go/items"
	. "github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
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

func makeEquipFetch2(a, b *SpecSetsEnable) loopedFetch {
	var allInfo []PreparedBonus
	allInfo = append(allInfo, a.EnabledSets...)
	allInfo = append(allInfo, b.EnabledSets...)

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

var priority = weight_types.SimPriorityBasic_Make(
	Sim_DPS, 0.01,
	Sim_DEATH, 0.32,
	Sim_TMI, 0.17,
	Sim_DTPS, 0.50,
)

func makeSpecSetsEnables() (*SpecSetsEnable, *SpecSetsEnable, *SpecSetsEnable) {
	return SpecSetsEnableNone(),
		SpecSetsEnableNamed("White Tiger Plate"),
		SpecSetsEnableForSpec_AllowFallback(Spec_PaladinRet, OptimiseGoal_Dps, true)
}

var resultFloat float64

func BenchmarkCalcBonusSolve(test *testing.B) {
	a, b, c := makeSpecSetsEnables()
	var v float64

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += a.CalcBonusSolveFlat(equip)
		v += b.CalcBonusSolveFlat(equip)
		v += c.CalcBonusSolveFlat(equip)
	}
	resultFloat = v
}

// func BenchmarkCalcBonusSolveC(test *testing.B) {
// 	a, b, c := makeSpecSetsEnablees()
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
// 	a, b, c := makeSpecSetsEnablees()
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
// 	a, b, c := makeSpecSetsEnablees()
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
// 	a, b, c := makeSpecSetsEnablees()
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
// 	a, b, c := makeSpecSetsEnablees()
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
// 	a, b, c := makeSpecSetsEnablees()
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
	a, b, c := makeSpecSetsEnables()

	equipFetch := makeEquipFetch2(b, c)

	sets := []*SpecSetsEnable{a, b, c}

	impls := []func(*SpecSetsEnable, *SolvableEquipMap) float64{
		(*SpecSetsEnable).CalcBonusSolveFlat,
		// (*SpecSetsEnable).CalcBonusSolveUseAssem,
		// (*SpecSetsEnable).CalcBonusSolveAssemAssumeNonNull,
		// (*SpecSetsEnable).CalcBonusSolveAssemAssumeNonNullWithCases,
		// func(s *SpecSetsEnable, e *SolvableEquipMap) float32 { return s.CalcBonusGeneric(e) },
	}

	for range 1000 {
		equip := equipFetch.next()

		for t, set := range sets {
			var expect float64
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

var g_SpecSetsEnableSlots = [5]SlotEquip{Equip_Head, Equip_Shoulder, Equip_Chest, Equip_Hand, Equip_Leg}

func makeEquipForBonus(numInSet int) *SolvableEquipMap {
	equip := SolvableEquipMap{}
	for slot := range equip {
		equip[slot] = makeItem()
	}

	setItems := []ItemId{86659, 86660, 86661, 86662, 86663}
	for i := range numInSet {
		slot := g_SpecSetsEnableSlots[i]
		equip[slot] = makeItemForBonus(setItems[i])
	}

	return &equip
}

func makeEquipForBonus2(a, b *PreparedBonus, x, y int) *SolvableEquipMap {
	equip := SolvableEquipMap{}
	for slot := range equip {
		equip[slot] = makeItem()
	}

	for _, slot := range g_SpecSetsEnableSlots {
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

func randChoice(slice []ItemId) ItemId {
	return slice[rand.IntN(len(slice))]
}

func makeItemForBonus(id ItemId) *SolvableItem {
	item := SolvableItem_ForTest(id, randStatBlock())
	return &item
}

func makeItem() *SolvableItem {
	id := rand.Uint32N(10000)
	item := SolvableItem_ForTest(ItemId(id), randStatBlock())
	return &item
}
