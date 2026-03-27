package model

import (
	"paladin_gearing_go/files"
	. "paladin_gearing_go/model/ratings"
	. "paladin_gearing_go/model/requirements"
	. "paladin_gearing_go/stats"
)

// ////////// standard model builders
func Model_PallyProtMitigation() Model {
	weightMiti := StatRatingsWeights_ReadFile(files.WeightMitiFile, false, true, false)
	weightDps := StatRatingsWeights_ReadFile(files.WeightDpsFile, false, true, false)
	weight := StatRatingsWeights_Mix(weightMiti, 95, weightDps, 34)
	return Model{
		Spec:             Spec_PaladinProtMitigation,
		StatRatings:      weight,
		StatRequirements: StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:     ReforgeRules_tank,
		EnchantChoice:    EnchantChoice_ForSpec(Spec_PaladinProtMitigation),
		GemChoice:        GemChoice_ForSpec(Spec_PaladinProtMitigation),
		SetBonus:         SetBonus_Named("Plate of the Lightning Emperor Prot Mitigation"),
		IsBlacksmith:     true,
		IsEngineer:       true,
	}
}

func Model_PallyProtDps() Model {
	weightMiti := StatRatingsWeights_ReadFile(files.WeightMitiFile, false, true, false)
	weightDps := StatRatingsWeights_ReadFile(files.WeightDpsFile, false, true, false)
	weight := StatRatingsWeights_Mix(weightMiti, 32, weightDps, 146)
	return Model{
		Spec:             Spec_PaladinProtDps,
		StatRatings:      weight,
		StatRequirements: StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:     ReforgeRules_tank,
		EnchantChoice:    EnchantChoice_ForSpec(Spec_PaladinProtDps),
		GemChoice:        GemChoice_ForSpec(Spec_PaladinProtDps),
		SetBonus:         SetBonus_Named("Plate of the Lightning Emperor Prot Damage"),
		IsBlacksmith:     true,
		IsEngineer:       true,
	}
}

func Model_PallyRet() Model {
	weight := StatRatingsWeights_ReadFile(files.WeightRetFile, false, false, false)
	return Model{
		Spec:             Spec_PaladinRet,
		StatRatings:      weight,
		StatRequirements: StatRequirementsHitExpertise_RetWideCap(),
		ReforgeRules:     ReforgeRules_melee,
		EnchantChoice:    EnchantChoice_ForSpec(Spec_PaladinRet),
		GemChoice:        GemChoice_ForSpec(Spec_PaladinRet),
		SetBonus:         SetBonus_ForSpec(Spec_PaladinRet),
		IsBlacksmith:     true,
		IsEngineer:       true,
	}
}

func Model_Testing() Model {
	weight := StatRatingsWeights_ReadFile(files.WeightDpsFile, false, true, false)
	return Model{
		Spec:             Spec_PaladinProtDps,
		StatRatings:      weight,
		StatRequirements: StatRequirementsHitExpertise_None(),
		ReforgeRules:     ReforgeRules_tank,
		EnchantChoice:    EnchantChoice_ForSpec(Spec_PaladinProtDps),
		GemChoice:        GemChoice_ForSpec(Spec_PaladinProtDps),
		SetBonus:         SetBonus_Empty(),
		IsBlacksmith:     false,
		IsEngineer:       false,
	}
}
