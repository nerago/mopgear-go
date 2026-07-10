package solve_build

import (
	"math/rand"
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

func SolverBuildRandom_MakeN_FullAndValidate(itemOptions *items.FullOptionsMap, model *gear_model.SpecModel, targetCount int, minimumHaste uint32) []items.FullItemSet {
	results := make([]items.FullItemSet, 0, targetCount)
	rng := rand.New(rand.NewSource(int64(0)))
	for len(results) < targetCount {
		itemSet := makeSetFromRandomFull(itemOptions, rng)
		if model.CheckSetFull_ForWeightProcess(&itemSet) && checkPairedSlotsNoDuplicate(itemSet.Items()) && itemSet.Total().GetUInt(stats.Stat_Haste) >= minimumHaste {
			itemSet.DebugValidate()
			itemSet.ValidateItemRules()
			results = append(results, itemSet)
		}
	}
	results = util.RemoveDuplicatesFunc(results, (*items.FullItemSet).Equals)
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
		} else if items.UniqueEquipViolation(a, b) {
			return false
		}
	}
	return true
}
