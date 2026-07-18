package gear_model

import (
	"paladin_gearing_go/files"
	. "paladin_gearing_go/gear_model/ratings_old"
	. "paladin_gearing_go/gear_model/requirements"
	"paladin_gearing_go/stats"
	. "paladin_gearing_go/stats"
)

// ////////// standard model builders
func Model_PallyProtMitigation_WithSet() SpecModel {
	spec := Spec_PaladinProt
	goal := OptimiseGoal_Mitigation
	weight := StatRatingsWeights_ReadFile(files.WeightMitiWithSetFile, false, true, false)
	return SpecModel{
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
			ActiveSetCountsRequiredMake(ActiveSet_Named("Plate of the Lightning Emperor"), 2, ActiveSet_Named("Plate of Winged Triumph"), 2),
		},
		Professions: ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile: files.GearFileProtMitigationWithSet,
	}
}

func Model_PallyProtHeal() SpecModel {
	spec := Spec_PaladinProt
	goal := OptimiseGoal_HalfMitiHeal
	weight := StatRatingsWeights_ReadFile(files.WeightHealFile, false, true, false)
	return SpecModel{
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
		ReferenceGearFile: files.GearFileProtHeal,
	}
}

func Model_PallyProtMitigation_NoSet() SpecModel {
	spec := Spec_PaladinProt
	goal := OptimiseGoal_Mitigation
	weight := StatRatingsWeights_ReadFile(files.WeightMitiNoSetFile, false, true, false)
	return SpecModel{
		Spec:                 spec,
		Goal:                 goal,
		SimulateAs:           Fight_Juggernaut_NoExternalHeal,
		SimRatioWeighting:    SimRatio_generalMiti,
		StatRatings:          weight,
		StatRequirements:     StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:         ReforgeRules_tank,
		StatsForWeighting:    StatsForWeighting_strengthTank,
		EnchantChoice:        EnchantChoice_ForSpec(spec, goal),
		GemChoice:            GemChoice_ForSpec(spec, goal),
		SetBonus:             SetBonus_Empty(),
		FixedWeightsSetBonus: ActiveSetCountsRequiredMake_Pointer(ActiveSet_Named("Plate of Winged Triumph"), 2),
		Professions: ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile: files.GearFileProtMitigationNoSet,
	}
}

func Model_PallyProtCompromise() SpecModel {
	spec := Spec_PaladinProt
	goal := OptimiseGoal_HalfMitiDps
	weight := StatRatingsWeights_ReadFile(files.WeightCompromiseFile, false, true, false)
	return SpecModel{
		Spec:                 spec,
		Goal:                 goal,
		SimulateAs:           Fight_Juggernaut_HighHeal,
		SimRatioWeighting:    SimRatio_compromiseWeight,
		StatRatings:          weight,
		StatRequirements:     StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:         ReforgeRules_tank,
		StatsForWeighting:    StatsForWeighting_strengthTank,
		EnchantChoice:        EnchantChoice_ForSpec(spec, goal),
		GemChoice:            GemChoice_ForSpec(spec, goal),
		SetBonus:             SetBonus_Empty(),
		FixedWeightsSetBonus: ActiveSetCountsRequiredMake_Pointer(ActiveSet_Named("Plate of Winged Triumph"), 2),
		Professions: ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile: files.GearFileProtCompromise,
	}
}

func Model_PallyProtDps() SpecModel {
	spec := Spec_PaladinProt
	goal := OptimiseGoal_Dps
	weight := StatRatingsWeights_ReadFile(files.WeightDpsFile, false, true, false)
	return SpecModel{
		Spec:                 spec,
		Goal:                 goal,
		SimulateAs:           Fight_Horridon_HighHeal,
		SimRatioWeighting:    SimRatio_dpsWeight,
		StatRatings:          weight,
		StatRequirements:     StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:         ReforgeRules_tank,
		StatsForWeighting:    StatsForWeighting_strengthTank,
		EnchantChoice:        EnchantChoice_ForSpec(spec, goal),
		GemChoice:            GemChoice_ForSpec(spec, goal),
		FixedWeightsSetBonus: ActiveSetCountsRequiredMake_Pointer(ActiveSet_Named("Plate of Winged Triumph"), 0),
		SetBonus:             SetBonus_Empty(),
		Professions: ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile: files.GearFileProtDps,
	}
}

func Model_PallyRet() SpecModel {
	spec := Spec_PaladinRet
	goal := OptimiseGoal_Dps
	weight := StatRatingsWeights_ReadFile(files.WeightRetFile, false, false, false)
	return SpecModel{
		Spec:                 spec,
		Goal:                 goal,
		SimulateAs:           Fight_Horridon_HighHeal,
		SimRatioWeighting:    SimRatio_retWeight,
		SimSpeedUp:           8,
		StatRatings:          weight,
		StatRequirements:     StatRequirementsHitExpertise_RetWideCap(),
		StatsForWeighting:    StatsForWeighting_strengthMelee,
		ReforgeRules:         ReforgeRules_melee,
		EnchantChoice:        EnchantChoice_ForSpec(spec, goal),
		GemChoice:            GemChoice_ForSpec(spec, goal),
		SetBonus:             SetBonus_Named("Battlegear of the Lightning Emperor", "Battlegear of Winged Triumph"),
		FixedWeightsSetBonus: ActiveSetCountsRequiredMake_Pointer(ActiveSet_Named("Battlegear of the Lightning Emperor"), 2, ActiveSet_Named("Battlegear of Winged Triumph"), 0),
		Professions: ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile: files.GearFileRet,
	}
}

func Model_Testing() SpecModel {
	spec := Spec_PaladinProt
	goal := OptimiseGoal_Dps
	return SpecModel{
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

// for noset - juggernaut, shamans, siegecrafter
var SimRatio_generalMiti = stats.SimData_Make(
	Sim_DPS, 0.1,
	Sim_DEATH, 0.2,
	Sim_TMI, 0.3,
	Sim_DTPS, 0.4,
)

// for withset - malkrok, thok, nazgrim
var SimRatio_malkrokWeight = stats.SimData_Make(
	Sim_DPS, 0.05,
	Sim_DEATH, 0.3,
	Sim_TMI, 0.05,
	Sim_DTPS, 0.60,
)

// for compromise set - spoils, galakras, paragons, etc
var SimRatio_compromiseWeight = stats.SimData_Make(
	Sim_DPS, 0.40,
	Sim_DEATH, 0.10,
	Sim_TMI, 0.35,
	Sim_DTPS, 0.15,
)

// for dps set
var SimRatio_dpsWeight = stats.SimData_Make(
	Sim_DPS, 0.95,
	Sim_DEATH, 0.01,
	Sim_TMI, 0.03,
	Sim_DTPS, 0.01,
)

// for heal set, garrosh, immerseus
var SimRatio_healWeight = stats.SimData_Make(
	Sim_DEATH, 0.15,
	Sim_TMI, 0.15,
	Sim_DTPS, 0.4,
	Sim_HPS, 0.3,
)

// for ret set
var SimRatio_retWeight = stats.SimData_Make(
	Sim_DPS, 1,
)

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
