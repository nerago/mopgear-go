package build

import (
	"math/rand"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/util"
)

func SolverBuildRandom_MakeN(itemOptions *items.SolvableOptionsMap, model *model.Model, targetCount int, printer *util.PrintRecorder) []items.SolvableItemSet {
	results := make([]items.SolvableItemSet, 0, targetCount)
	rng := rand.New(rand.NewSource(int64(0)))
	for len(results) < targetCount {
		itemSet := makeSetFromRandom(itemOptions, rng)
		if model.CheckSet(&itemSet) {
			itemSet.DebugValidate()
			results = append(results, itemSet)
		}
	}
	return results
}

func makeSetFromRandom(itemOptions *items.SolvableOptionsMap, rng *rand.Rand) items.SolvableItemSet {
	equip := items.SolvableEquipMap{}
	for slot, options := range itemOptions {
		optionSize := len(options)
		if optionSize > 0 {
			index := rng.Intn(optionSize)
			equip[slot] = &options[index]
		}
	}
	return items.SolvableItemSet_Of(equip)
}

func SolverBuildRandom_MakeN_FullAndValidate(itemOptions *items.FullOptionsMap, model *model.Model, targetCount int, printer *util.PrintRecorder) []items.FullItemSet {
	results := make([]items.FullItemSet, 0, targetCount)
	rng := rand.New(rand.NewSource(int64(0)))
	for len(results) < targetCount {
		itemSet := makeSetFromRandomFull(itemOptions, rng)
		if model.StatRequirements.CheckSet(itemSet.Total()) && checkPairedSlotsNoDuplicate(itemSet.Items()) {
			itemSet.DebugValidate()
			itemSet.ValidateItemRules()
			results = append(results, itemSet)
		}
	}
	return results
}

func makeSetFromRandomFull(itemOptions *items.FullOptionsMap, rng *rand.Rand) items.FullItemSet {
	equip := items.FullEquipMap{}
	for slot, options := range itemOptions {
		optionSize := len(options)
		if optionSize > 0 {
			index := rng.Intn(optionSize)
			equip[slot] = &options[index]
		}
	}
	return items.FullItemSet_FromMap(equip)
}

func checkPairedSlotsNoDuplicate(equip *items.FullEquipMap) bool {
	return checkPairedSlotNoDuplicate(equip.Get(items.Equip_Ring1), equip.Get(items.Equip_Ring2)) &&
		checkPairedSlotNoDuplicate(equip.Get(items.Equip_Trinket1), equip.Get(items.Equip_Trinket2))
}

func checkPairedSlotNoDuplicate(a, b *items.FullItem) bool {
	if a != nil && b != nil {
		if a.ItemId() == b.ItemId() {
			return false
		} else if items.UniqueEquipViolation(a.BaseName(), b.BaseName()) {
			return false
		}
	}
	return true
}
