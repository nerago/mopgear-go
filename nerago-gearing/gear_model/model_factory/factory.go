package model_factory

import (
	"paladin_gearing_go/files"
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/gear_model/ratings_old"
	"paladin_gearing_go/gear_model/requirements"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/weightfind/weight_types"
)

// ////////// standard model builders
func Model_PallyProtMitigation_WithSet() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_Mitigation
	weight := tools.StatRatingsWeights_ReadFile(files.WeightMitiWithSetFile, false, true, false)
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
		SetBonus:          gear_model.SetBonus_Named("Plate of the Lightning Emperor", "Plate of Winged Triumph"),
		SetBonusRequired: []gear_model.ActiveSetCountsRequired{
			gear_model.ActiveSetCountsRequiredMake(gear_model.ActiveSet_Named("Plate of the Lightning Emperor"), 2, gear_model.ActiveSet_Named("Plate of Winged Triumph"), 2),
		},
		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile: files.GearFileProtMitigationWithSet,
	}
}

func Model_PallyProtHeal() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_HalfMitiHeal
	weight := tools.StatRatingsWeights_ReadFile(files.WeightHealFile, false, true, false)
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
		SetBonus:          gear_model.SetBonus_Named("Plate of Winged Triumph"),
		SetBonusRequired: []gear_model.ActiveSetCountsRequired{
			gear_model.ActiveSetCountsRequiredMake(gear_model.ActiveSet_Named("Plate of Winged Triumph"), 4),
		},
		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile: files.GearFileProtHeal,
	}
}

func Model_PallyProtMitigation_NoSet() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_Mitigation
	weight := tools.StatRatingsWeights_ReadFile(files.WeightMitiNoSetFile, false, true, false)
	return gear_model.SpecModel{
		Spec:                 spec,
		Goal:                 goal,
		SimulateAs:           stats.Fight_Juggernaut_NoExternalHeal,
		SimPriority:          SimPriority_generalMiti,
		StatWeights:          weight,
		StatRequirements:     requirements.StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:         gear_model.ReforgeRules_tank,
		StatsForWeighting:    StatsForWeighting_strengthTank,
		EnchantChoice:        gear_model.EnchantChoice_ForSpec(spec, goal),
		GemChoice:            gear_model.GemChoice_ForSpec(spec, goal),
		SetBonus:             gear_model.SetBonus_Empty(),
		FixedWeightsSetBonus: gear_model.ActiveSetCountsRequiredMake_Pointer(gear_model.ActiveSet_Named("Plate of Winged Triumph"), 2),
		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile: files.GearFileProtMitigationNoSet,
	}
}

func Model_PallyProtCompromise() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_HalfMitiDps
	weight := tools.StatRatingsWeights_ReadFile(files.WeightCompromiseFile, false, true, false)
	return gear_model.SpecModel{
		Spec:                 spec,
		Goal:                 goal,
		SimulateAs:           stats.Fight_Juggernaut_HighHeal,
		SimPriority:          SimPriority_compromise,
		StatWeights:          weight,
		StatRequirements:     requirements.StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:         gear_model.ReforgeRules_tank,
		StatsForWeighting:    StatsForWeighting_strengthTank,
		EnchantChoice:        gear_model.EnchantChoice_ForSpec(spec, goal),
		GemChoice:            gear_model.GemChoice_ForSpec(spec, goal),
		SetBonus:             gear_model.SetBonus_Empty(),
		FixedWeightsSetBonus: gear_model.ActiveSetCountsRequiredMake_Pointer(gear_model.ActiveSet_Named("Plate of Winged Triumph"), 2),
		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile: files.GearFileProtCompromise,
	}
}

func Model_PallyProtDps() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_Dps
	weight := tools.StatRatingsWeights_ReadFile(files.WeightDpsFile, false, true, false)
	return gear_model.SpecModel{
		Spec:                 spec,
		Goal:                 goal,
		SimulateAs:           stats.Fight_Horridon_HighHeal,
		SimPriority:          SimPriority_dps,
		StatWeights:          weight,
		StatRequirements:     requirements.StatRequirementsHitExpertise_ProtFlexibleParry(),
		ReforgeRules:         gear_model.ReforgeRules_tank,
		StatsForWeighting:    StatsForWeighting_strengthTank,
		EnchantChoice:        gear_model.EnchantChoice_ForSpec(spec, goal),
		GemChoice:            gear_model.GemChoice_ForSpec(spec, goal),
		FixedWeightsSetBonus: gear_model.ActiveSetCountsRequiredMake_Pointer(gear_model.ActiveSet_Named("Plate of Winged Triumph"), 0),
		SetBonus:             gear_model.SetBonus_Empty(),
		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile: files.GearFileProtDps,
	}
}

func Model_PallyRet() gear_model.SpecModel {
	spec := stats.Spec_PaladinRet
	goal := stats.OptimiseGoal_Dps
	weight := tools.StatRatingsWeights_ReadFile(files.WeightRetFile, false, false, false)
	return gear_model.SpecModel{
		Spec:                 spec,
		Goal:                 goal,
		SimulateAs:           stats.Fight_Horridon_HighHeal,
		SimPriority:          SimPriority_ret,
		SimSpeedUp:           8,
		StatWeights:          weight,
		StatRequirements:     requirements.StatRequirementsHitExpertise_RetWideCap(),
		StatsForWeighting:    StatsForWeighting_strengthMelee,
		ReforgeRules:         gear_model.ReforgeRules_melee,
		EnchantChoice:        gear_model.EnchantChoice_ForSpec(spec, goal),
		GemChoice:            gear_model.GemChoice_ForSpec(spec, goal),
		SetBonus:             gear_model.SetBonus_Named("Battlegear of the Lightning Emperor", "Battlegear of Winged Triumph"),
		FixedWeightsSetBonus: gear_model.ActiveSetCountsRequiredMake_Pointer(gear_model.ActiveSet_Named("Battlegear of the Lightning Emperor"), 2, gear_model.ActiveSet_Named("Battlegear of Winged Triumph"), 0),
		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile: files.GearFileRet,
	}
}

func Model_Testing() gear_model.SpecModel {
	spec := stats.Spec_PaladinProt
	goal := stats.OptimiseGoal_Dps
	return gear_model.SpecModel{
		Spec:              spec,
		Goal:              goal,
		SimulateAs:        stats.Fight_Horridon_HighHeal,
		StatWeights:       ratings_old.StatRatingsWeights_Testing(),
		StatRequirements:  requirements.StatRequirementsHitExpertise_None(),
		StatsForWeighting: StatsForWeighting_strengthTank,
		ReforgeRules:      gear_model.ReforgeRules_tank,
		EnchantChoice:     gear_model.EnchantChoice_ForSpec(spec, goal),
		GemChoice:         gear_model.GemChoice_ForSpec(spec, goal),
		SetBonus:          gear_model.SetBonus_Empty(),
		Professions: gear_model.ProfessionInfo{
			IsBlacksmith: true,
			IsEngineer:   true,
		},
		ReferenceGearFile: files.GearFileProtDps,
	}
}

// for noset - juggernaut, shamans, siegecrafter
var SimPriority_generalMiti = weight_types.SimPriorityBasic_Make(
	stats.Sim_DPS, 0.2,
	stats.Sim_DEATH, 0.1,
	stats.Sim_TMI, 0.3,
	stats.Sim_DTPS, 0.4,
)

// for withset - malkrok, thok, nazgrim
var SimPriority_withSet = weight_types.SimPriorityBasic_Make(
	stats.Sim_DPS, 0.01,
	stats.Sim_DEATH, 0.35,
	stats.Sim_TMI, 0.09,
	stats.Sim_DTPS, 0.55,
)

// for compromise set - spoils, galakras, paragons, etc
var SimPriority_compromise = weight_types.SimPriorityBasic_Make(
	stats.Sim_DPS, 0.40,
	stats.Sim_DEATH, 0.15,
	stats.Sim_TMI, 0.30,
	stats.Sim_DTPS, 0.15,
)

// for dps set
var SimPriority_dps = weight_types.SimPriorityBasic_Make(
	stats.Sim_DPS, 0.95,
	stats.Sim_DEATH, 0.01,
	stats.Sim_TMI, 0.03,
	stats.Sim_DTPS, 0.01,
)

// for heal set, garrosh, immerseus
var SimPriority_heal = weight_types.SimPriorityBasic_Make(
	stats.Sim_DPS, 0.10,
	stats.Sim_HPS, 0.20,
	stats.Sim_DEATH, 0.10,
	stats.Sim_TMI, 0.15,
	stats.Sim_DTPS, 0.45,
)

// for ret set
var SimPriority_ret = weight_types.SimPriorityBasic_Make(
	stats.Sim_DPS, 1,
)

var StatsForWeighting_strengthTank = []stats.StatType{
	stats.Stat_Strength,
	stats.Stat_Stamina,
	stats.Stat_Haste,
	stats.Stat_Mastery,
	stats.Stat_Crit,
	stats.Stat_Dodge,
	stats.Stat_Parry,
	stats.Stat_Expertise,
}

var StatsForWeighting_strengthMelee = []stats.StatType{
	stats.Stat_Strength,
	stats.Stat_Haste,
	stats.Stat_Mastery,
	stats.Stat_Crit,
	stats.Stat_Expertise,
}
