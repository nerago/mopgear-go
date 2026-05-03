package model

import (
	"paladin_gearing_go/files"
	. "paladin_gearing_go/model/ratings"
	. "paladin_gearing_go/model/requirements"
	"paladin_gearing_go/stats"
	. "paladin_gearing_go/stats"
)

// ////////// standard model builders
func Model_PallyProtMitigation_WithSet() Model {
	spec := Spec_PaladinProtMitigation
	weight := StatRatingsWeights_ReadFile(files.WeightMitiWithSetFile, false, true, false)
	return Model{
		Spec:             spec,
		SimulateAs:       stats.Fight_Horridon_LowHeal,
		StatRatings:      weight,
		StatRequirements: StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:     ReforgeRules_tank,
		EnchantChoice:    EnchantChoice_ForSpec(spec),
		GemChoice:        GemChoice_ForSpec(spec),
		SetBonus:         SetBonus_Named("Plate of the Lightning Emperor"),
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
		SimulateAs:       stats.Fight_Animus,
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
	weightMiti := StatRatingsWeights_ReadFile(files.WeightMitiNoSetFile, false, true, false)
	weightDps := StatRatingsWeights_ReadFile(files.WeightDpsFile, false, true, false)
	weight := StatRatingsWeights_Mix(weightMiti, 62, weightDps, 51)
	return Model{
		Spec:             spec,
		SimulateAs:       stats.Fight_Animus,
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
		SimulateAs:       stats.Fight_Horridon_HighHeal,
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
		SimulateAs:       stats.Fight_Horridon_HighHeal,
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
		SimulateAs:       stats.Fight_Horridon_HighHeal,
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
