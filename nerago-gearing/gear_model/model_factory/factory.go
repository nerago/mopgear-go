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
func Model_PallyProtMitigation_WithSet() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_Mitigation
	weight := tools.StatRatingsWeights_ReadFile(files.WeightMitiWithSetFile)
	return gear_model.SpecModel{
		Spec:              spec,
		Goal:              goal,
		SimulateAs:        stats.Fight_Juggernaut_NoExternalHeal,
		SimPriority:       SimPriority_withSet,
		StatWeights:       weight,
		StatRequirements:  requirements.StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:      gear_model.ReforgeRules_tank,
		StatsForWeighting: StatsForWeighting_strengthTank,
		EnchantChoice:     gear_model.EnchantChoice_ForSpec(spec, goal),
		GemChoice:         gear_model.GemChoice_ForSpec(spec, goal),

		BonusEnabled:        bonus_set.SpecSetsEnableNamed(setNameT15Prot, setNameT16Prot),
		BonusRequiredSolve:  bonus_set.ItemCountsRequiredOptionsForFactory(BonusItems_Prot15_Prot16_2pcEach),
		BonusRequiredWeight: nil,

		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile:        files.GearFileProtMitigationWithSet,
		SpecificIncompatibleList: trinketsStrengthMeleeOnly,
	}
}

func Model_PallyProtHeal() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_HalfMitiHeal
	weight := tools.StatRatingsWeights_ReadFile(files.WeightHealFile)
	return gear_model.SpecModel{
		Spec:              spec,
		Goal:              goal,
		SimulateAs:        stats.Fight_Juggernaut_OffHealer,
		SimPriority:       SimPriority_heal,
		StatWeights:       weight,
		StatRequirements:  requirements.StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:      gear_model.ReforgeRules_tank,
		StatsForWeighting: StatsForWeighting_strengthTank,
		EnchantChoice:     gear_model.EnchantChoice_ForSpec(spec, goal),
		GemChoice:         gear_model.GemChoice_ForSpec(spec, goal),

		BonusEnabled:        bonus_set.SpecSetsEnableNamed(setNameT16Prot),
		BonusRequiredSolve:  bonus_set.ItemCountsRequiredOptionsForFactory(BonusItems_Prot16_4pc),
		BonusRequiredWeight: nil,

		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile:        files.GearFileProtHeal,
		SpecificIncompatibleList: trinketsStrengthMeleeOnly,
	}
}

func Model_PallyProtMitigation_NoSet() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_Mitigation
	weight := tools.StatRatingsWeights_ReadFile(files.WeightMitiNoSetFile)
	return gear_model.SpecModel{
		Spec:              spec,
		Goal:              goal,
		SimulateAs:        stats.Fight_Juggernaut_NoExternalHeal,
		SimPriority:       SimPriority_generalMiti,
		StatWeights:       weight,
		StatRequirements:  requirements.StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:      gear_model.ReforgeRules_tank,
		StatsForWeighting: StatsForWeighting_strengthTank,
		EnchantChoice:     gear_model.EnchantChoice_ForSpec(spec, goal),
		GemChoice:         gear_model.GemChoice_ForSpec(spec, goal),

		BonusEnabled:        bonus_set.SpecSetsEnableNone(),
		BonusRequiredSolve:  bonus_set.ItemCountsRequiredOptionsAny(),
		BonusRequiredWeight: &BonusItems_Prot16_2pcOnly,

		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile:        files.GearFileProtMitigationNoSet,
		SpecificIncompatibleList: trinketsStrengthMeleeOnly,
	}
}

func Model_PallyProtCompromise() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_HalfMitiDps
	weight := tools.StatRatingsWeights_ReadFile(files.WeightCompromiseFile)
	return gear_model.SpecModel{
		Spec:              spec,
		Goal:              goal,
		SimulateAs:        stats.Fight_Juggernaut_HighHeal,
		SimPriority:       SimPriority_compromise,
		StatWeights:       weight,
		StatRequirements:  requirements.StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:      gear_model.ReforgeRules_tank,
		StatsForWeighting: StatsForWeighting_strengthTank,
		EnchantChoice:     gear_model.EnchantChoice_ForSpec(spec, goal),
		GemChoice:         gear_model.GemChoice_ForSpec(spec, goal),

		BonusEnabled:        bonus_set.SpecSetsEnableNone(),
		BonusRequiredSolve:  bonus_set.ItemCountsRequiredOptionsAny(),
		BonusRequiredWeight: &BonusItems_Prot16_2pcOnly,

		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile:        files.GearFileProtCompromise,
		SpecificIncompatibleList: trinketsStrengthMeleeOnly,
	}
}

func Model_PallyProtDps() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_Dps
	weight := tools.StatRatingsWeights_ReadFile(files.WeightDpsFile)

	return gear_model.SpecModel{
		Spec:              spec,
		Goal:              goal,
		SimulateAs:        stats.Fight_Horridon_HighHeal,
		SimPriority:       SimPriority_dps,
		StatWeights:       weight,
		StatRequirements:  requirements.StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:      gear_model.ReforgeRules_tank,
		StatsForWeighting: StatsForWeighting_strengthTank,
		EnchantChoice:     gear_model.EnchantChoice_ForSpec(spec, goal),
		GemChoice:         gear_model.GemChoice_ForSpec(spec, goal),

		BonusEnabled:        bonus_set.SpecSetsEnableNone(),
		BonusRequiredSolve:  bonus_set.ItemCountsRequiredOptionsAny(),
		BonusRequiredWeight: &BonusItems_ZeroAll,

		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile:        files.GearFileProtDps,
		SpecificIncompatibleList: trinketsStrengthMeleeOnly,
	}
}

func Model_PallyRet() gear_model.SpecModel {
	spec := stats.Spec_PaladinRet
	goal := stats.OptimiseGoal_Dps
	weight := tools.StatRatingsWeights_ReadFile(files.WeightRetFile)

	return gear_model.SpecModel{
		Spec:              spec,
		Goal:              goal,
		SimulateAs:        stats.Fight_Horridon_HighHeal,
		SimPriority:       SimPriority_ret,
		SimSpeedUp:        8,
		StatWeights:       weight,
		StatRequirements:  requirements.StatRequirementsHitExpertise_RetWideCap(),
		StatsForWeighting: StatsForWeighting_strengthMelee,
		ReforgeRules:      gear_model.ReforgeRules_melee,
		EnchantChoice:     gear_model.EnchantChoice_ForSpec(spec, goal),
		GemChoice:         gear_model.GemChoice_ForSpec(spec, goal),

		BonusEnabled:        bonus_set.SpecSetsEnableNamed(setNameT15Ret, setNameT16Ret),
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
