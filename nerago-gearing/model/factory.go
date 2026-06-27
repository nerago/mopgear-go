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
	spec := Spec_PaladinProt
	goal := OptimiseGoal_Mitigation
	weight := StatRatingsWeights_ReadFile(files.WeightMitiWithSetFile, false, true, false)
	return Model{
		Spec:              spec,
		Goal:              goal,
		SimulateAs:        Fight_Juggernaut_NoExternalHeal,
		SimRatioWeighting: SimRatio_malkrokWeight,
		StatRatings:       weight,
		StatRequirements:  StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:      ReforgeRules_tank,
		StatsForWeighting: StatsForWeighting_strengthTank,
		EnchantChoice:     EnchantChoice_ForSpec(spec, goal),
		GemChoice:         GemChoice_ForSpec(spec, goal),
		SetBonus:          SetBonus_Named("Plate of the Lightning Emperor", "Plate of Winged Triumph"),
		SetBonusRequired: []ActiveSetCountsRequired{
			ActiveSetCountsRequiredMake(ActiveSet_Named("Plate of the Lightning Emperor"), 4),
			ActiveSetCountsRequiredMake(ActiveSet_Named("Plate of the Lightning Emperor"), 2, ActiveSet_Named("Plate of Winged Triumph"), 2),
		},
		Professions: ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile: files.GearFileProtMitigationWithSet,
	}
}

func Model_PallyProtHeal() Model {
	spec := Spec_PaladinProt
	goal := OptimiseGoal_HalfMitiHeal
	weight := StatRatingsWeights_ReadFile(files.WeightHealFile, false, true, false) // TODO new file
	return Model{
		Spec:              spec,
		Goal:              goal,
		SimulateAs:        Fight_Juggernaut_OffHealer,
		SimRatioWeighting: SimRatio_healWeight,
		StatRatings:       weight,
		StatRequirements:  StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:      ReforgeRules_tank,
		StatsForWeighting: StatsForWeighting_strengthTank,
		EnchantChoice:     EnchantChoice_ForSpec(spec, goal),
		GemChoice:         GemChoice_ForSpec(spec, goal),
		SetBonus:          SetBonus_Named("Plate of Winged Triumph"),
		SetBonusRequired: []ActiveSetCountsRequired{
			ActiveSetCountsRequiredMake(ActiveSet_Named("Plate of Winged Triumph"), 4),
		},
		Professions: ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile: files.GearFileProtMitigationWithSet, // TODO new file
	}
}

func Model_PallyProtMitigation_NoSet() Model {
	spec := Spec_PaladinProt
	goal := OptimiseGoal_Mitigation
	weight := StatRatingsWeights_ReadFile(files.WeightMitiNoSetFile, false, true, false)
	return Model{
		Spec:              spec,
		Goal:              goal,
		SimulateAs:        Fight_Juggernaut_NoExternalHeal,
		SimRatioWeighting: SimRatio_generalMiti,
		StatRatings:       weight,
		StatRequirements:  StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:      ReforgeRules_tank,
		StatsForWeighting: StatsForWeighting_strengthTank,
		EnchantChoice:     EnchantChoice_ForSpec(spec, goal),
		GemChoice:         GemChoice_ForSpec(spec, goal),
		SetBonus:          SetBonus_Empty(),
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
		Spec:              spec,
		Goal:              goal,
		SimulateAs:        Fight_Juggernaut_HighHeal,
		SimRatioWeighting: SimRatio_animusWeight,
		StatRatings:       weight,
		StatRequirements:  StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:      ReforgeRules_tank,
		StatsForWeighting: StatsForWeighting_strengthTank,
		EnchantChoice:     EnchantChoice_ForSpec(spec, goal),
		GemChoice:         GemChoice_ForSpec(spec, goal),
		SetBonus:          SetBonus_Empty(),
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
		Spec:              spec,
		Goal:              goal,
		SimulateAs:        Fight_Horridon_HighHeal,
		SimRatioWeighting: SimRatio_dpsWeight,
		StatRatings:       weight,
		StatRequirements:  StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:      ReforgeRules_tank,
		StatsForWeighting: StatsForWeighting_strengthTank,
		EnchantChoice:     EnchantChoice_ForSpec(spec, goal),
		GemChoice:         GemChoice_ForSpec(spec, goal),
		// SetBonus:         SetBonus_Named("Plate of the Lightning Emperor Prot Damage"),
		SetBonus: SetBonus_Empty(),
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
		Spec:              spec,
		Goal:              goal,
		SimulateAs:        Fight_Horridon_HighHeal,
		SimRatioWeighting: SimRatio_retWeight,
		SimSpeedUp:        8,
		StatRatings:       weight,
		StatRequirements:  StatRequirementsHitExpertise_RetWideCap(),
		StatsForWeighting: StatsForWeighting_strengthMelee,
		ReforgeRules:      ReforgeRules_melee,
		EnchantChoice:     EnchantChoice_ForSpec(spec, goal),
		GemChoice:         GemChoice_ForSpec(spec, goal),
		SetBonus:          SetBonus_Named("Battlegear of the Lightning Emperor", "Battlegear of Winged Triumph"),
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
		Spec:              spec,
		Goal:              goal,
		SimulateAs:        Fight_Horridon_HighHeal,
		StatRatings:       StatRatingsWeights_Testing(),
		StatRequirements:  StatRequirementsHitExpertise_None(),
		StatsForWeighting: StatsForWeighting_strengthTank,
		ReforgeRules:      ReforgeRules_tank,
		EnchantChoice:     EnchantChoice_ForSpec(spec, goal),
		GemChoice:         GemChoice_ForSpec(spec, goal),
		SetBonus:          SetBonus_Empty(),
		Professions: ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile: files.GearFileProtDps,
	}
}

// now about noset - tortos, horridon, ironqon, jikun, durumu
var SimRatio_generalMiti = stats.SimData{
	DPS:   0.2,
	DEATH: 0.2,
	TMI:   0.2,
	DTPS:  0.4,
}

// for withset - malkrok
var SimRatio_malkrokWeight = stats.SimData{
	DPS:   0.10,
	DEATH: 0.3,
	TMI:   0.05,
	DTPS:  0.55,
}

// for compromise set - animus
var SimRatio_animusWeight = stats.SimData{
	DPS:   0.5,
	DEATH: 0.1,
	TMI:   0.3,
	DTPS:  0.1,
}

// for dps set
var SimRatio_dpsWeight = stats.SimData{
	DPS:   0.90,
	DEATH: 0.03,
	TMI:   0.03,
	DTPS:  0.04,
}

// for heal set
var SimRatio_healWeight = stats.SimData{
	DPS:   0.05,
	DEATH: 0.1,
	TMI:   0.1,
	DTPS:  0.3,
	HPS:   0.45,
}

// for ret set
var SimRatio_retWeight = stats.SimData{
	DPS: 1,
}

var StatsForWeighting_strengthTank = []stats.StatType{
	stats.Stat_Strength,
	stats.Stat_Stamina,
	stats.Stat_Crit,
	stats.Stat_Haste,
	stats.Stat_Expertise,
	stats.Stat_Mastery,
	stats.Stat_Dodge,
	stats.Stat_Parry,
}

var StatsForWeighting_strengthMelee = []stats.StatType{
	stats.Stat_Strength,
	stats.Stat_Crit,
	stats.Stat_Haste,
	stats.Stat_Expertise,
	stats.Stat_Mastery,
}
