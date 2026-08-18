package gear_model

import (
	"github.com/nerago/mopgear-go/gear_model/model_factory"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/stats"
	"math/rand/v2"
	"testing"
)

var resultFloat float64

func BenchmarkCalcSetSpecific(test *testing.B) {
	model := model_factory.Model_PallyProtSurvival()
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
