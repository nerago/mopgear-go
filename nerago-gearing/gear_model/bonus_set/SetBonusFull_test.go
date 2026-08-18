package bonus_set

import (
	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/items"
	"math/rand/v2"
)

// func BenchmarkCalcBonusFull(test *testing.B) {
// 	a, b, c := makeSetBonuses()
// 	var v float32

// 	for test.Loop() {
// 		test.StopTimer()
// 		equip := makeEquipFullForBonus()
// 		test.StartTimer()

// 		v += a.CalcBonus(equip)
// 		v += b.CalcBonus(equip)
// 		v += c.CalcBonus(equip)
// 	}
// 	resultFloat = v
// }

// func BenchmarkCalcBonus1GenericFull(test *testing.B) {
// 	a, b, c := makeSetBonuses()
// 	var v float32

// 	for test.Loop() {
// 		test.StopTimer()
// 		equip := makeEquipFullForBonus()
// 		test.StartTimer()

// 		v += a.CalcBonus1Generic(equip)
// 		v += b.CalcBonus1Generic(equip)
// 		v += c.CalcBonus1Generic(equip)
// 	}
// 	resultFloat = v
// }

func makeEquipFullForBonus() *items.FullEquipMap {
	equip := items.FullEquipMap{}
	for slot := range equip {
		equip[slot] = makeItemFull()
	}

	setItems := []uint32{86659, 86660, 86661, 86662, 86663}
	numInSet := rand.IntN(6)
	for i := range numInSet {
		slot := g_setBonusSlots[i]
		equip[slot] = makeItemForFullBonus(setItems[i])
	}

	return &equip
}

func makeItemFull() *items.FullItem {
	id := rand.Uint32N(10000)
	item := items.FullItem_ForTest(items.ItemId(id), items.Item_Head, gear_model.randStatBlock())
	return &item
}

func makeItemForFullBonus(id uint32) *items.FullItem {
	item := items.FullItem_ForTest(items.ItemId(id), items.Item_Head, gear_model.randStatBlock())
	return &item
}
