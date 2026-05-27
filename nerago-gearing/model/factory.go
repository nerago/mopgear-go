package model

import (
	"paladin_gearing_go/files"
	. "paladin_gearing_go/model/ratings"
	. "paladin_gearing_go/model/requirements"
	. "paladin_gearing_go/stats"
)

// ////////// standard model builders
func Model_PallyProtMitigation_WithSet() Model {
	spec := Spec_PaladinProt
	goal := OptimiseGoal_Mitigation
	weight := StatRatingsWeights_ReadFile(files.WeightMitiWithSetFile, false, true, false)
	return Model{
		Spec:        spec,
		Goal:        goal,
		SimulateAs:  Fight_Horridon_LowHeal, // TODO really a raden set
		StatRatings: weight,
		StatRequirements: StatRequirementsHitExpertise_ProtFlexibleParry_PlusAdditional(
			// &StatAndValue{StatType: Stat_Haste, Value: 13500}),
			&StatAndValue{StatType: Stat_Haste, Value: 13000}),
		// &StatAndValue{StatType: Stat_Haste, Value: 12500}),
		ReforgeRules:     ReforgeRules_tank,
		EnchantChoice:    EnchantChoice_ForSpec(spec, goal),
		GemChoice:        GemChoice_ForSpec(spec, goal),
		SetBonus:         SetBonus_Named("Plate of the Lightning Emperor"),
		// SetBonus:         SetBonus_Named("Plate of the Lightning Emperor", "Plate of Winged Triumph"),
		SetBonusRequired: 4,
		Professions: ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile: files.GearFileProtMitigationSet,
	}
}

func Model_PallyProtMitigation_NoSet() Model {
	spec := Spec_PaladinProt
	goal := OptimiseGoal_Mitigation
	weight := StatRatingsWeights_ReadFile(files.WeightMitiNoSetFile, false, true, false)
	return Model{
		Spec:             spec,
		Goal:             goal,
		SimulateAs:       Fight_Horridon_LowHeal,
		StatRatings:      weight,
		StatRequirements: StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:     ReforgeRules_tank,
		EnchantChoice:    EnchantChoice_ForSpec(spec, goal),
		GemChoice:        GemChoice_ForSpec(spec, goal),
		SetBonus:         SetBonus_Empty(),
		Professions: ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile: files.GearFileProtMitigationNoSet,
	}
}

func Model_PallyProtCompromise() Model {
	spec := Spec_PaladinProt
	goal := OptimiseGoal_HalfMitiDps
	weight := StatRatingsWeights_ReadFile(files.WeightCompromiseFile, false, true, false)
	return Model{
		Spec:             spec,
		Goal:             goal,
		SimulateAs:       Fight_Animus,
		StatRatings:      weight,
		StatRequirements: StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:     ReforgeRules_tank,
		EnchantChoice:    EnchantChoice_ForSpec(spec, goal),
		GemChoice:        GemChoice_ForSpec(spec, goal),
		SetBonus:         SetBonus_Named("Plate of the Lightning Emperor Prot Damage"),
		Professions: ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile: files.GearFileProtCompromise,
	}
}

func Model_PallyProtDps() Model {
	spec := Spec_PaladinProt
	goal := OptimiseGoal_Dps
	weight := StatRatingsWeights_ReadFile(files.WeightDpsFile, false, true, false)
	return Model{
		Spec:             spec,
		Goal:             goal,
		SimulateAs:       Fight_Horridon_HighHeal,
		StatRatings:      weight,
		StatRequirements: StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:     ReforgeRules_tank,
		EnchantChoice:    EnchantChoice_ForSpec(spec, goal),
		GemChoice:        GemChoice_ForSpec(spec, goal),
		SetBonus:         SetBonus_Named("Plate of the Lightning Emperor Prot Damage"),
		Professions: ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile: files.GearFileProtDps,
	}
}

func Model_PallyRet() Model {
	spec := Spec_PaladinRet
	goal := OptimiseGoal_Dps
	weight := StatRatingsWeights_ReadFile(files.WeightRetFile, false, false, false)
	return Model{
		Spec:             spec,
		Goal:             goal,
		SimulateAs:       Fight_Horridon_HighHeal,
		StatRatings:      weight,
		StatRequirements: StatRequirementsHitExpertise_RetWideCap(),
		ReforgeRules:     ReforgeRules_melee,
		EnchantChoice:    EnchantChoice_ForSpec(spec, goal),
		GemChoice:        GemChoice_ForSpec(spec, goal),
		SetBonus:         SetBonus_ForSpec(spec, goal),
		Professions: ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile: files.GearFileRet,
	}
}

func Model_Testing() Model {
	spec := Spec_PaladinProt
	goal := OptimiseGoal_Dps
	return Model{
		Spec:             spec,
		Goal:             goal,
		SimulateAs:       Fight_Horridon_HighHeal,
		StatRatings:      StatRatingsWeights_Testing(),
		StatRequirements: StatRequirementsHitExpertise_None(),
		ReforgeRules:     ReforgeRules_tank,
		EnchantChoice:    EnchantChoice_ForSpec(spec, goal),
		GemChoice:        GemChoice_ForSpec(spec, goal),
		SetBonus:         SetBonus_Empty(),
		Professions: ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile: files.GearFileProtDps,
	}
}
