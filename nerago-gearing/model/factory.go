package model

import (
	"paladin_gearing_go/files"
	. "paladin_gearing_go/model/ratings"
	. "paladin_gearing_go/model/requirements"
	. "paladin_gearing_go/stats"
)

// ////////// standard model builders
func Model_PallyProtMitigation_WithSet() Model {
	spec := Spec_PaladinProtMitigation
	weight := StatRatingsWeights_ReadFile(files.WeightMitiWithSetFile, false, true, false)
	return Model{
		Spec:        spec,
		SimulateAs:  Fight_Horridon_LowHeal, // TODO really a raden set
		StatRatings: weight,
		StatRequirements: StatRequirementsHitExpertise_ProtFlexibleParry_PlusAdditional(
			&StatAndValue{StatType: Stat_Haste, Value: 13500}),
		ReforgeRules:  ReforgeRules_tank,
		EnchantChoice: EnchantChoice_ForSpec(spec),
		GemChoice:     GemChoice_ForSpec(spec),
		SetBonus:      SetBonus_Named("Plate of the Lightning Emperor"),
		Professions: ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
	}
}

func Model_PallyProtMitigation_NoSet() Model {
	spec := Spec_PaladinProtMitigation
	weight := StatRatingsWeights_ReadFile(files.WeightMitiNoSetFile, false, true, false)
	return Model{
		Spec:             spec,
		SimulateAs:       Fight_Animus,
		StatRatings:      weight,
		StatRequirements: StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:     ReforgeRules_tank,
		EnchantChoice:    EnchantChoice_ForSpec(spec),
		GemChoice:        GemChoice_ForSpec(spec),
		SetBonus:         SetBonus_Empty(),
		Professions: ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
	}
}

func Model_PallyProtCompromise() Model {
	spec := Spec_PaladinProtCompromise
	weight := StatRatingsWeights_ReadFile(files.GearFileProtCompromise, false, true, false)
	return Model{
		Spec:             spec,
		SimulateAs:       Fight_Animus,
		StatRatings:      weight,
		StatRequirements: StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:     ReforgeRules_tank,
		EnchantChoice:    EnchantChoice_ForSpec(spec),
		GemChoice:        GemChoice_ForSpec(spec),
		SetBonus:         SetBonus_Named("Plate of the Lightning Emperor Prot Damage"),
		Professions: ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
	}
}

func Model_PallyProtDps() Model {
	spec := Spec_PaladinProtDps
	weight := StatRatingsWeights_ReadFile(files.WeightDpsFile, false, true, false)
	return Model{
		Spec:             spec,
		SimulateAs:       Fight_Horridon_HighHeal,
		StatRatings:      weight,
		StatRequirements: StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:     ReforgeRules_tank,
		EnchantChoice:    EnchantChoice_ForSpec(spec),
		GemChoice:        GemChoice_ForSpec(spec),
		SetBonus:         SetBonus_Named("Plate of the Lightning Emperor Prot Damage"),
		Professions: ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
	}
}

func Model_PallyRet() Model {
	weight := StatRatingsWeights_ReadFile(files.WeightRetFile, false, false, false)
	return Model{
		Spec:             Spec_PaladinRet,
		SimulateAs:       Fight_Horridon_HighHeal,
		StatRatings:      weight,
		StatRequirements: StatRequirementsHitExpertise_RetWideCap(),
		ReforgeRules:     ReforgeRules_melee,
		EnchantChoice:    EnchantChoice_ForSpec(Spec_PaladinRet),
		GemChoice:        GemChoice_ForSpec(Spec_PaladinRet),
		SetBonus:         SetBonus_ForSpec(Spec_PaladinRet),
		Professions: ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
	}
}

func Model_Testing() Model {
	return Model{
		Spec:             Spec_PaladinProtDps,
		SimulateAs:       Fight_Horridon_HighHeal,
		StatRatings:      StatRatingsWeights_Testing(),
		StatRequirements: StatRequirementsHitExpertise_None(),
		ReforgeRules:     ReforgeRules_tank,
		EnchantChoice:    EnchantChoice_ForSpec(Spec_PaladinProtDps),
		GemChoice:        GemChoice_ForSpec(Spec_PaladinProtDps),
		SetBonus:         SetBonus_Empty(),
		Professions: ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
	}
}
