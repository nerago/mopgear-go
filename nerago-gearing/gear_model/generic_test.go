package gear_model

import (
	"math/rand/v2"
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"testing"
)

var resultFloat float64

func BenchmarkCalcSetSpecific(test *testing.B) {
	model := Model_PallyProtMitigation_WithSet()
	set := makeSet()
	var v float64
	for test.Loop() {
		v += model.CalcRatingSolve(set)
	}
	resultFloat = v
}

func makeSet() *items.SolvableItemSet {
	equip := items.SolvableEquipMap{}
	for slot := range equip {
		equip[slot] = makeItem()
	}
	set := items.SolvableItemSet_Of(equip)
	return &set
}

func makeItem() *items.SolvableItem {
	id := rand.Uint32N(10000)
	item := items.SolvableItem_ForTest(items.ItemId(id), randStatBlock())
	return &item
}

func randStatBlock() stats.StatBlock {
	block := stats.StatBlock{}
	for i := range block {
		block[i] = rand.Uint32()
	}
	return block
}
