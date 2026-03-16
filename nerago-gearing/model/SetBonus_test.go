package model

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"testing"
)

type loopedFetch struct {
	items []*items.SolvableEquipMap
	index int
}

func (fetch *loopedFetch) add(item *items.SolvableEquipMap) {
	fetch.items = append(fetch.items, item)
}

func (fetch *loopedFetch) next() *items.SolvableEquipMap {
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
	return SetBonus_Empty(), SetBonus_Named("White Tiger Plate"), SetBonus_ForSpec(stats.Spec_PaladinRet)
}

func BenchmarkCalcBonusSolveUseFuncLoopy(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += a.CalcBonusSolveUseFuncLoopy(equip)
		v += b.CalcBonusSolveUseFuncLoopy(equip)
		v += c.CalcBonusSolveUseFuncLoopy(equip)
	}
	resultFloat = v
}
func BenchmarkCalcBonusSolveUseFuncLoopy_Member(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += a.CalcBonusSolveUseFuncLoopy_Member(equip)
		v += b.CalcBonusSolveUseFuncLoopy_Member(equip)
		v += c.CalcBonusSolveUseFuncLoopy_Member(equip)
	}
	resultFloat = v
}
func BenchmarkCalcBonusSolveUseFuncUnrollAlways(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += a.CalcBonusSolveUseFuncUnrollAlways(equip)
		v += b.CalcBonusSolveUseFuncUnrollAlways(equip)
		v += c.CalcBonusSolveUseFuncUnrollAlways(equip)
	}
	resultFloat = v
}
func BenchmarkCalcBonusSolveUseFuncUnrollAndSwitch(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += a.CalcBonusSolveUseFuncUnrollAndSwitch(equip)
		v += b.CalcBonusSolveUseFuncUnrollAndSwitch(equip)
		v += c.CalcBonusSolveUseFuncUnrollAndSwitch(equip)
	}
	resultFloat = v
}
func BenchmarkCalcBonusSolveUseFuncUnrollAndSwitchBranchy(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += a.CalcBonusSolveUseFuncUnrollAndSwitchBranchy(equip)
		v += b.CalcBonusSolveUseFuncUnrollAndSwitchBranchy(equip)
		v += c.CalcBonusSolveUseFuncUnrollAndSwitchBranchy(equip)
	}
	resultFloat = v
}

func BenchmarkCalcBonusSolveUseFuncLoopyGenericTypeParam(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += CalcBonusSolveUseFuncLoopyGenericTypeParam(&a, *equip)
		v += CalcBonusSolveUseFuncLoopyGenericTypeParam(&b, *equip)
		v += CalcBonusSolveUseFuncLoopyGenericTypeParam(&c, *equip)
	}
	resultFloat = v
}
func BenchmarkCalcBonusSolveUseFuncLoopyGenericInterface(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += CalcBonusSolveUseFuncLoopyGenericInterface(&a, equip)
		v += CalcBonusSolveUseFuncLoopyGenericInterface(&b, equip)
		v += CalcBonusSolveUseFuncLoopyGenericInterface(&c, equip)
	}
	resultFloat = v
}
func BenchmarkCalcBonusSolveUseFuncUnrollAndSwitchGeneric(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += a.CalcBonusSolveUseFuncUnrollAndSwitchGeneric(equip)
		v += b.CalcBonusSolveUseFuncUnrollAndSwitchGeneric(equip)
		v += c.CalcBonusSolveUseFuncUnrollAndSwitchGeneric(equip)
	}
	resultFloat = v
}

func BenchmarkCalcBonusGenericInterfaceLoopySomeInlined(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += a.CalcBonusGenericInterfaceLoopySomeInlined(equip)
		v += b.CalcBonusGenericInterfaceLoopySomeInlined(equip)
		v += c.CalcBonusGenericInterfaceLoopySomeInlined(equip)
	}
	resultFloat = v
}

func BenchmarkCalcBonusSolveDirectValueBasedLoopy_Inlinable(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += a.CalcBonusSolveDirectValueBasedLoopy_Inlinable(*equip)
		v += b.CalcBonusSolveDirectValueBasedLoopy_Inlinable(*equip)
		v += c.CalcBonusSolveDirectValueBasedLoopy_Inlinable(*equip)
	}
	resultFloat = v
}

func BenchmarkCalcBonusSolveDirectUnroll(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += a.CalcBonusSolveDirectUnroll(equip)
		v += b.CalcBonusSolveDirectUnroll(equip)
		v += c.CalcBonusSolveDirectUnroll(equip)
	}
	resultFloat = v
}
func BenchmarkCalcBonusSolveUnrollAndSpecial0(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += a.CalcBonusSolveUnrollAndSpecial0(equip)
		v += b.CalcBonusSolveUnrollAndSpecial0(equip)
		v += c.CalcBonusSolveUnrollAndSpecial0(equip)
	}
	resultFloat = v
}
func BenchmarkCalcBonusSolveDirectUnrollAndSwitch(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += a.CalcBonusSolveDirectUnrollAndSwitch(equip)
		v += b.CalcBonusSolveDirectUnrollAndSwitch(equip)
		v += c.CalcBonusSolveDirectUnrollAndSwitch(equip)
	}
	resultFloat = v
}

func BenchmarkCalcBonusGenericInterfaceGettersUnroll(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += a.CalcBonusGenericInterfaceGettersUnroll(equip)
		v += b.CalcBonusGenericInterfaceGettersUnroll(equip)
		v += c.CalcBonusGenericInterfaceGettersUnroll(equip)
	}
	resultFloat = v
}

func BenchmarkCalcBonusGenericTypeParamGettersUnroll(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += CalcBonusGenericTypeParamGettersUnroll(&a, equip)
		v += CalcBonusGenericTypeParamGettersUnroll(&b, equip)
		v += CalcBonusGenericTypeParamGettersUnroll(&c, equip)
	}
	resultFloat = v
}

func BenchmarkCalcBonusGenericTypeParamArrayUnroll(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += CalcBonusGenericTypeParamArrayUnroll[*items.SolvableItem, items.SolvableEquipMap](&a, equip)
		v += CalcBonusGenericTypeParamArrayUnroll(&b, equip)
		v += CalcBonusGenericTypeParamArrayUnroll(&c, equip)
	}
	resultFloat = v
}

func BenchmarkCalcBonusGenericInterfaceGettersLoopy(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += a.CalcBonusGenericInterfaceGettersLoopy(equip)
		v += b.CalcBonusGenericInterfaceGettersLoopy(equip)
		v += c.CalcBonusGenericInterfaceGettersLoopy(equip)
	}
	resultFloat = v
}

func BenchmarkCalcBonusGenericTypeParamGettersLoopy(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += CalcBonusGenericTypeParamGettersLoopy(&a, equip)
		v += CalcBonusGenericTypeParamGettersLoopy(&b, equip)
		v += CalcBonusGenericTypeParamGettersLoopy(&c, equip)
	}
	resultFloat = v
}

func BenchmarkCalcBonusGenericTypeParamArrayLoopy(test *testing.B) {
	a, b, c := makeSetBonuses()
	var v float32

	equipFetch := makeEquipFetch()

	for test.Loop() {
		equip := equipFetch.next()
		v += CalcBonusGenericTypeParamArrayLoopy(&a, *equip)
		v += CalcBonusGenericTypeParamArrayLoopy(&b, *equip)
		v += CalcBonusGenericTypeParamArrayLoopy(&c, *equip)
	}
	resultFloat = v
}

func makeEquipForBonus(numInSet int) *items.SolvableEquipMap {
	equip := items.SolvableEquipMap{}
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

func makeItemForBonus(id uint32) *items.SolvableItem {
	item := items.SolvableItem_ForTest(id, randStatBlock())
	return &item
}
