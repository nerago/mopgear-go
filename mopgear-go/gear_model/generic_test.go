package gear_model

import (
	"math/rand/v2"
	"testing"

	"github.com/nerago/mopgear-go/cmd/mygear"
	"github.com/nerago/mopgear-go/files"
	"github.com/nerago/mopgear-go/gear_model/bonus_set"
	"github.com/nerago/mopgear-go/gear_model/ratings"
	"github.com/nerago/mopgear-go/gear_model/requirements"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

var resultFloat float64

func BenchmarkCalcSetSpecific(test *testing.B) {
	model := modelTest()
	set := makeSet()
	var v float64
	for test.Loop() {
		v += model.CalcRatingSolve(set, 1)
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

func modelTest() SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_Mitigation
	weight := ratings.StatRatingsWeightsExtended{
		Weight1: weight_types.Weight1Basic_Of(
			[]float64{1.0000000000, 3.3054613752, 1.3553929822, 2.5269080431, -1.3481523198, 0.2858798437, 1.6104384233, -0.6888153486},
			[]stats.StatType{stats.Stat_Strength, stats.Stat_Stamina, stats.Stat_Haste, stats.Stat_Mastery, stats.Stat_Crit, stats.Stat_Dodge, stats.Stat_Parry, stats.Stat_Expertise},
		),
	}
	priority := weight_types.SimPriorityBasic_Make(
		stats.Sim_DPS, 0.01,
		stats.Sim_DEATH, 0.32,
		stats.Sim_TMI, 0.17,
		stats.Sim_DTPS, 0.50,
	)
	return SpecModel{
		Spec:             spec,
		Goal:             goal,
		SimulateAs:       stats.Fight_Juggernaut_SelfWordGlory,
		SimPriority:      priority,
		StatWeights:      weight,
		StatRequirements: requirements.StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:     ReforgeRules_tank,
		StatsForWeighting: []stats.StatType{
			stats.Stat_Strength,
			stats.Stat_Stamina,
			stats.Stat_Haste,
			stats.Stat_Mastery,
			stats.Stat_Crit,
			stats.Stat_Dodge,
			stats.Stat_Parry,
			stats.Stat_Expertise,
		},
		EnchantChoice: EnchantChoice_ForSpec(spec, goal),
		GemChoice:     GemChoice_ForSpec(spec, goal),
		BonusEnabled:  bonus_set.SpecSetsEnableNamed("Plate of Winged Triumph"),
		BonusRequiredSolve: bonus_set.ItemCountsRequiredOptionsMake(
			bonus_set.CountMode_AllowPlusOne, // no real justification for any restriction
			bonus_set.ItemCountsRequiredMake("Plate of Winged Triumph", 2),
		),
		Professions: ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile:  files.GearFileProtSurvival,
		BlockSpecificItems: mygear.TrinketsStrengthMeleeOnly,
	}
}
