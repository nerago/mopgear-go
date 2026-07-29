package gear_model

import (
	"paladin_gearing_go/files"
	. "paladin_gearing_go/gear_model/ratings_old"
	. "paladin_gearing_go/gear_model/requirements"
	"paladin_gearing_go/stats"
	. "paladin_gearing_go/stats"
	"paladin_gearing_go/weightfind/weight_types"
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
		SimPriority:       SimPriority_withSet,
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
		SimPriority:       SimPriority_heal,
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
		SimPriority:          SimPriority_generalMiti,
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
		SimPriority:          SimPriority_compromise,
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
		SimPriority:          SimPriority_dps,
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
		SimPriority:          SimPriority_ret,
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
var SimPriority_generalMiti = weight_types.SimPriorityBasic_Make(
	Sim_DPS, 0.2,
	Sim_DEATH, 0.1,
	Sim_TMI, 0.3,
	Sim_DTPS, 0.4,
)

// for withset - malkrok, thok, nazgrim
var SimPriority_withSet = weight_types.SimPriorityBasic_Make(
	Sim_DPS, 0.01,
	Sim_DEATH, 0.35,
	Sim_TMI, 0.09,
	Sim_DTPS, 0.55,
)

// for compromise set - spoils, galakras, paragons, etc
var SimPriority_compromise = weight_types.SimPriorityBasic_Make(
	Sim_DPS, 0.40,
	Sim_DEATH, 0.15,
	Sim_TMI, 0.30,
	Sim_DTPS, 0.15,
)

// for dps set
var SimPriority_dps = weight_types.SimPriorityBasic_Make(
	Sim_DPS, 0.95,
	Sim_DEATH, 0.01,
	Sim_TMI, 0.03,
	Sim_DTPS, 0.01,
)

// for heal set, garrosh, immerseus
var SimPriority_heal = weight_types.SimPriorityBasic_Make(
	Sim_DPS, 0.10,
	Sim_HPS, 0.20,
	Sim_DEATH, 0.10,
	Sim_TMI, 0.15,
	Sim_DTPS, 0.45,
)

// for ret set
var SimPriority_ret = weight_types.SimPriorityBasic_Make(
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
