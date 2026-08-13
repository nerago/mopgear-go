package model_factory

import (
	"paladin_gearing_go/files"
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/gear_model/bonus_set"
	"paladin_gearing_go/gear_model/requirements"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/tools"
)

var bonusMake = bonus_set.ItemCountsRequiredMake

const (
	setNameT15Prot      = "Plate of the Lightning Emperor"
	setNameT16Prot      = "Plate of Winged Triumph"
	setNameT15Ret       = "Battlegear of the Lightning Emperor"
	setNameT16Ret       = "Battlegear of Winged Triumph"
	wordOfGlory15Prot   = "Plate of the Lightning Emperor - Word of Glory"
	eternalFlameT16Prot = "Plate of Winged Triumph - Eternal Flame Full"
)

// ////////// standard model builders
func Model_PallyProtSurvival() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_Mitigation
	weight := tools.StatRatingsWeights_ReadFile(files.WeightSurvivalFile)
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
		BonusEnabled:      bonus_set.SpecSetsEnableNamed(&priority, wordOfGlory15Prot, setNameT16Prot),
		BonusRequiredSolve: bonus_set.ItemCountsRequiredOptionsMake(
			bonus_set.CountMode_AllowPlusOne,
			bonusMake(wordOfGlory15Prot, 2, setNameT16Prot, 0),
			bonusMake(wordOfGlory15Prot, 2, setNameT16Prot, 2),
			bonusMake(wordOfGlory15Prot, 0, setNameT16Prot, 2),
			bonusMake(wordOfGlory15Prot, 0, setNameT16Prot, 4),
		),
		BonusRequiredWeight: new(bonusMake(wordOfGlory15Prot, 0, setNameT16Prot, 2)),
		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile:        files.GearFileProtSurvival,
		SpecificIncompatibleList: trinketsStrengthMeleeOnly,
	}
}

func Model_PallyProtHeal() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_HalfMitiHeal
	weight := tools.StatRatingsWeights_ReadFile(files.WeightHealFile)
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
			bonusMake(eternalFlameT16Prot, 2),
			bonusMake(eternalFlameT16Prot, 4),
		),
		BonusRequiredWeight: new(bonusMake(eternalFlameT16Prot, 4)),
		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile:        files.GearFileProtHeal,
		SpecificIncompatibleList: trinketsStrengthMeleeOnly,
	}
}

func Model_PallyProtMitigation() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_Mitigation
	weight := tools.StatRatingsWeights_ReadFile(files.WeightMitigationFile)
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
		BonusEnabled:      bonus_set.SpecSetsEnableNamed(&priority, setNameT15Prot, setNameT16Prot),
		BonusRequiredSolve: bonus_set.ItemCountsRequiredOptionsMake(
			bonus_set.CountMode_AllowPlusOne,
			bonusMake(setNameT16Prot, 2, setNameT15Prot, 0),
			bonusMake(setNameT16Prot, 2, setNameT15Prot, 2),
			bonusMake(setNameT16Prot, 4, setNameT15Prot, 0),
		),
		BonusRequiredWeight: new(bonusMake(setNameT16Prot, 2, setNameT15Prot, 0)),
		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile:        files.GearFileProtMitigation,
		SpecificIncompatibleList: trinketsStrengthMeleeOnly,
	}
}

func Model_PallyProtBalanced() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_HalfMitiDps
	weight := tools.StatRatingsWeights_ReadFile(files.WeightBalancedFile)
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
			bonus_set.CountMode_AllowPlusOne,
			bonusMake(setNameT16Prot, 0),
			bonusMake(setNameT16Prot, 2),
			bonusMake(setNameT16Prot, 4)),
		BonusRequiredWeight: new(bonusMake(setNameT16Prot, 0)),
		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile:        files.GearFileProtBalanced,
		SpecificIncompatibleList: trinketsStrengthMeleeOnly,
	}
}

func Model_PallyProtDamage() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_Dps
	weight := tools.StatRatingsWeights_ReadFile(files.WeightDamageFile)
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
		SpecificIncompatibleList: trinketsStrengthMeleeOnly,
	}
}

func Model_PallyRet() gear_model.SpecModel {
	spec := stats.Spec_PaladinRet
	goal := stats.OptimiseGoal_Dps
	weight := tools.StatRatingsWeights_ReadFile(files.WeightRetFile)
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
		BonusEnabled:      bonus_set.SpecSetsEnableNamed(&priority, setNameT15Ret, setNameT16Ret),
		BonusRequiredSolve: bonus_set.ItemCountsRequiredOptionsMake(
			bonus_set.CountMode_Exact,
			bonusMake(setNameT15Ret, 0, setNameT16Ret, 0),
			bonusMake(setNameT15Ret, 2, setNameT16Ret, 0),
			bonusMake(setNameT15Ret, 2, setNameT16Ret, 2),
			bonusMake(setNameT15Ret, 0, setNameT16Ret, 2),
			bonusMake(setNameT15Ret, 0, setNameT16Ret, 4),
		),
		BonusRequiredWeight: new(bonusMake(setNameT15Ret, 0, setNameT16Ret, 4)),
		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile:        files.GearFileRet,
		SpecificIncompatibleList: trinketsStrengthTankOnly,
	}
}
