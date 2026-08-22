package bonus_set

import (
	"math/rand/v2"

	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/stats"
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

var g_setBonusSlots = []items.SlotEquip{
	items.Equip_Head,
	items.Equip_Chest,
	items.Equip_Hand,
	items.Equip_Leg,
	items.Equip_Shoulder,
}

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
	item := items.FullItem_ForTest(items.ItemId(id), items.Item_Head, randStatBlock())
	return &item
}

func makeItemForFullBonus(id uint32) *items.FullItem {
	item := items.FullItem_ForTest(items.ItemId(id), items.Item_Head, randStatBlock())
	return &item
}

func randStatBlock() stats.StatBlock {
	block := stats.StatBlock{}
	for i := range block {
		block[i] = rand.Uint32()
	}
	return block
}
