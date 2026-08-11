package model_factory

import (
	"paladin_gearing_go/files"
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/gear_model/bonus_set"
	"paladin_gearing_go/gear_model/requirements"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/tools"
)

// ////////// standard model builders
func Model_PallyProtSurvival() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_Mitigation
	weight := tools.StatRatingsWeights_ReadFile(files.WeightSurvivalFile)
	priority := SimPriority_survival
	wordOfGlory15Prot := "Plate of the Lightning Emperor - Word of Glory"
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
			bonus_set.ItemCountsRequiredMake(wordOfGlory15Prot, 0, setNameT16Prot, 0),
			bonus_set.ItemCountsRequiredMake(wordOfGlory15Prot, 0, setNameT16Prot, 2),
			bonus_set.ItemCountsRequiredMake(wordOfGlory15Prot, 2, setNameT16Prot, 0),
			bonus_set.ItemCountsRequiredMake(wordOfGlory15Prot, 2, setNameT16Prot, 2),
			bonus_set.ItemCountsRequiredMake(setNameT16Prot, 4),
		),
		BonusRequiredWeight: new(BonusItems_Prot16_2pcOnly),
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
	eternalFlameT16Prot := "Plate of Winged Triumph - Eternal Flame Full"
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
		BonusEnabled:      bonus_set.SpecSetsEnableNamed(&priority, setNameT15Prot, eternalFlameT16Prot),
		BonusRequiredSolve: bonus_set.ItemCountsRequiredOptionsMake(
			bonus_set.CountMode_AllowPlusOne,
			bonus_set.ItemCountsRequiredMake(setNameT15Prot, 0, eternalFlameT16Prot, 2),
			bonus_set.ItemCountsRequiredMake(eternalFlameT16Prot, 4),
		),
		BonusRequiredWeight: new(BonusItems_Prot16_4pc),
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
			BonusItems_Prot16_2pcOnly,
			BonusItems_Prot16_4pc,
			BonusItems_Prot15_Prot16_2pcEach,
			BonusItems_Prot15_2pcOnly,
			BonusItems_ProtZero,
		),
		BonusRequiredWeight: &BonusItems_Prot16_2pcOnly,
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
		Spec:                spec,
		Goal:                goal,
		SimulateAs:          stats.Fight_Juggernaut_HighHeal,
		SimPriority:         priority,
		StatWeights:         weight,
		StatRequirements:    requirements.StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:        gear_model.ReforgeRules_tank,
		StatsForWeighting:   StatsForWeighting_strengthTank,
		EnchantChoice:       gear_model.EnchantChoice_ForSpec(spec, goal),
		GemChoice:           gear_model.GemChoice_ForSpec(spec, goal),
		BonusEnabled:        bonus_set.SpecSetsEnableNamed(&priority, setNameT15Prot, setNameT16Prot),
		BonusRequiredSolve:  bonus_set.ItemCountsRequiredOptionsAny(),
		BonusRequiredWeight: &BonusItems_Prot16_2pcOnly,
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
		BonusRequiredWeight: &BonusItems_ProtZero,
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
		Spec:                spec,
		Goal:                goal,
		SimulateAs:          stats.Fight_Horridon_HighHeal,
		SimPriority:         priority,
		SimSpeedUp:          6,
		StatWeights:         weight,
		StatRequirements:    requirements.StatRequirementsHitExpertise_RetWideCap(),
		StatsForWeighting:   StatsForWeighting_strengthMelee,
		ReforgeRules:        gear_model.ReforgeRules_melee,
		EnchantChoice:       gear_model.EnchantChoice_ForSpec(spec, goal),
		GemChoice:           gear_model.GemChoice_ForSpec(spec, goal),
		BonusEnabled:        bonus_set.SpecSetsEnableNamed(&priority, setNameT15Ret, setNameT16Ret),
		BonusRequiredSolve:  bonus_set.ItemCountsRequiredOptionsAny(),
		BonusRequiredWeight: &BonusItems_Ret15_Ret16_2pcEach,
		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile:        files.GearFileRet,
		SpecificIncompatibleList: trinketsStrengthTankOnly,
	}
}
