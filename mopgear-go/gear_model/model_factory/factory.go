package model_factory

import (
	"github.com/nerago/mopgear-go/cmd/mygear"
	"github.com/nerago/mopgear-go/files"
	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/gear_model/bonus_set"
	"github.com/nerago/mopgear-go/gear_model/requirements"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/tools"
)

var bonusMake = bonus_set.ItemCountsRequiredMake

const (
	setNameT15Prot      = "Plate of the Lightning Emperor"
	setNameT16Prot      = "Plate of Winged Triumph"
	setNameT15Ret       = "Battlegear of the Lightning Emperor"
	setNameT16Ret       = "Battlegear of Winged Triumph"
	eternalFlameT16Prot = "Plate of Winged Triumph - Eternal Flame Full"
)

// ////////// standard model builders
func Model_PallyProtSurvival() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_Mitigation
	weight, err := tools.StatRatingsWeightsExtended_ReadFile(files.WeightProtSurvival)
	if err != nil {
		panic(err)
	}
	priority := SimPriority_survival
	return gear_model.SpecModel{
		Spec:              spec,
		Goal:              goal,
		SimulateAs:        stats.Fight_Juggernaut_SelfWordGlory,
		SimPriority:       priority,
		StatWeights:       weight,
		StatRequirements:  requirements.StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:      gear_model.ReforgeRules_tank,
		StatsForWeighting: StatsForWeighting_strengthTank,
		EnchantChoice:     gear_model.EnchantChoice_ForSpec(spec, goal),
		GemChoice:         gear_model.GemChoice_ForSpec(spec, goal),
		BonusEnabled:      bonus_set.SpecSetsEnableNamed(&priority, setNameT16Prot),
		BonusRequiredSolve: bonus_set.ItemCountsRequiredOptionsMake(
			bonus_set.CountMode_Minimum,
			bonusMake(setNameT16Prot, 2),
		),
		BonusRequiredWeight: new(bonusMake(setNameT15Prot, 0, setNameT16Prot, 2)),
		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile:        files.GearFileProtSurvival,
		SpecificIncompatibleList: mygear.TrinketsStrengthMeleeOnly,
	}
}

func Model_PallyProtHeal() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_HalfMitiHeal
	weight := tools.StatRatingsWeightsExtended_ReadFile(files.WeightProtHeal)
	priority := SimPriority_heal
	return gear_model.SpecModel{
		Spec:              spec,
		Goal:              goal,
		SimulateAs:        stats.Fight_Juggernaut_OffHealer,
		SimPriority:       priority,
		StatWeights:       weight,
		StatRequirements:  requirements.StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:      gear_model.ReforgeRules_tank,
		StatsForWeighting: StatsForWeighting_strengthTank,
		EnchantChoice:     gear_model.EnchantChoice_ForSpec(spec, goal),
		GemChoice:         gear_model.GemChoice_ForSpec(spec, goal),
		BonusEnabled:      bonus_set.SpecSetsEnableNamed(&priority, eternalFlameT16Prot),
		BonusRequiredSolve: bonus_set.ItemCountsRequiredOptionsMake(
			bonus_set.CountMode_AllowPlusOne,
			bonusMake(eternalFlameT16Prot, 4),
		),
		BonusRequiredWeight: new(bonusMake(eternalFlameT16Prot, 4)),
		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile:        files.GearFileProtHeal,
		SpecificIncompatibleList: mygear.TrinketsStrengthMeleeOnly,
	}
}

func Model_PallyProtMitigation() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_Mitigation
	weight := tools.StatRatingsWeightsExtended_ReadFile(files.WeightProtMitigation)
	priority := SimPriority_mitigation
	return gear_model.SpecModel{
		Spec:              spec,
		Goal:              goal,
		SimulateAs:        stats.Fight_Juggernaut_NoExternalHeal,
		SimPriority:       priority,
		StatWeights:       weight,
		StatRequirements:  requirements.StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:      gear_model.ReforgeRules_tank,
		StatsForWeighting: StatsForWeighting_strengthTank,
		EnchantChoice:     gear_model.EnchantChoice_ForSpec(spec, goal),
		GemChoice:         gear_model.GemChoice_ForSpec(spec, goal),
		BonusEnabled:      bonus_set.SpecSetsEnableNamed(&priority, setNameT16Prot),
		BonusRequiredSolve: bonus_set.ItemCountsRequiredOptionsMake(
			bonus_set.CountMode_AllowPlusOne, // no real justification for any restriction
			bonusMake(setNameT16Prot, 0),
			bonusMake(setNameT16Prot, 2),
		),
		BonusRequiredWeight: new(bonusMake(setNameT16Prot, 2, setNameT15Prot, 0)),
		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile:        files.GearFileProtMitigation,
		SpecificIncompatibleList: mygear.TrinketsStrengthMeleeOnly,
	}
}

func Model_PallyProtBalanced() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_HalfMitiDps
	weight := tools.StatRatingsWeightsExtended_ReadFile(files.WeightProtBalanced)
	priority := SimPriority_balanced
	return gear_model.SpecModel{
		Spec:              spec,
		Goal:              goal,
		SimulateAs:        stats.Fight_Juggernaut_HighHeal,
		SimPriority:       priority,
		StatWeights:       weight,
		StatRequirements:  requirements.StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:      gear_model.ReforgeRules_tank,
		StatsForWeighting: StatsForWeighting_strengthTank,
		EnchantChoice:     gear_model.EnchantChoice_ForSpec(spec, goal),
		GemChoice:         gear_model.GemChoice_ForSpec(spec, goal),
		BonusEnabled:      bonus_set.SpecSetsEnableNamed(&priority, setNameT16Prot),
		BonusRequiredSolve: bonus_set.ItemCountsRequiredOptionsMake(
			bonus_set.CountMode_Exact,
			bonusMake(setNameT16Prot, 0),
			bonusMake(setNameT16Prot, 2)),
		BonusRequiredWeight: new(bonusMake(setNameT16Prot, 0)),
		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile:        files.GearFileProtBalanced,
		SpecificIncompatibleList: mygear.TrinketsStrengthMeleeOnly,
	}
}

func Model_PallyProtDamage() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_Dps
	weight := tools.StatRatingsWeightsExtended_ReadFile(files.WeightProtDamage)
	return gear_model.SpecModel{
		Spec:                spec,
		Goal:                goal,
		SimulateAs:          stats.Fight_Horridon_HighHeal,
		SimPriority:         SimPriority_dps,
		StatWeights:         weight,
		StatRequirements:    requirements.StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:        gear_model.ReforgeRules_tank,
		StatsForWeighting:   StatsForWeighting_strengthTank,
		EnchantChoice:       gear_model.EnchantChoice_ForSpec(spec, goal),
		GemChoice:           gear_model.GemChoice_ForSpec(spec, goal),
		BonusEnabled:        bonus_set.SpecSetsEnableNone(),
		BonusRequiredSolve:  bonus_set.ItemCountsRequiredOptionsAny(),
		BonusRequiredWeight: nil,
		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile:        files.GearFileProtDamage,
		SpecificIncompatibleList: mygear.TrinketsStrengthMeleeOnly,
	}
}

func Model_PallyRet() gear_model.SpecModel {
	spec := stats.Spec_PaladinRet
	goal := stats.OptimiseGoal_Dps
	weight := tools.StatRatingsWeightsExtended_ReadFile(files.WeightRet)
	priority := SimPriority_ret
	return gear_model.SpecModel{
		Spec:              spec,
		Goal:              goal,
		SimulateAs:        stats.Fight_Horridon_HighHeal,
		SimPriority:       priority,
		SimSpeedUp:        6,
		StatWeights:       weight,
		StatRequirements:  requirements.StatRequirementsHitExpertise_RetWideCap(),
		StatsForWeighting: StatsForWeighting_strengthMelee,
		ReforgeRules:      gear_model.ReforgeRules_melee,
		EnchantChoice:     gear_model.EnchantChoice_ForSpec(spec, goal),
		GemChoice:         gear_model.GemChoice_ForSpec(spec, goal),
		BonusEnabled:      bonus_set.SpecSetsEnableNamed(&priority, setNameT16Ret),
		BonusRequiredSolve: bonus_set.ItemCountsRequiredOptionsMake(
			bonus_set.CountMode_Minimum,
			bonusMake(setNameT16Ret, 2),
		),
		BonusRequiredWeight: new(bonusMake(setNameT15Ret, 0, setNameT16Ret, 2)),
		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile:        files.GearFileRet,
		SpecificIncompatibleList: mygear.TrinketsStrengthTankOnly,
	}
}
