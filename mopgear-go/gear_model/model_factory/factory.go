package model_factory

import (
	"github.com/nerago/mopgear-go/cmd/mygear"
	"github.com/nerago/mopgear-go/files"
	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/gear_model/bonus_set"
	"github.com/nerago/mopgear-go/gear_model/requirements"
	"github.com/nerago/mopgear-go/stats"
)

var bonusMake = bonus_set.ItemCountsRequiredMake

const (
	setNameT15Prot      = "Plate of the Lightning Emperor"
	setNameT16Prot      = "Plate of Winged Triumph"
	setNameT15Ret       = "Battlegear of the Lightning Emperor"
	setNameT16Ret       = "Battlegear of Winged Triumph"
	eternalFlameT16Prot = "Plate of Winged Triumph - Eternal Flame Full"
)

// ////////// standard model builders
func Model_PallyProtSurvival() gear_model.SpecModel {
	return gear_model.SpecModel{
		Spec: stats.Spec_PaladinProt,
		ModelItems: gear_model.ModelItems{
			GearFile:           files.GearFileProtSurvival,
			BlockSpecificItems: mygear.TrinketsStrengthMeleeOnly,
			ReforgeRules:       gear_model.ReforgeRules_tank,
			Professions:        gear_model.ProfessionInfo{IsBlacksmith: true, IsEngineer: true},
		},
		ModelSolve: gear_model.ModelSolve{
			WeightFile:        files.WeightProtSurvival,
			SimPriority:       SimPriority_survival,
			StatsForWeighting: StatsForWeighting_strengthTank,
			StatRequirements:  requirements.StatRequirementsHitExpertise_ProtFlexibleParry(),
		},
		ModelSimulate: gear_model.ModelSimulate{
			Goal:       stats.OptimiseGoal_Mitigation,
			SimulateAs: stats.Fight_Juggernaut_SelfWordGlory,
		},
		ModelBonus: gear_model.ModelBonus{
			BonusEnabled: bonus_set.SpecSetsEnableNamed(setNameT16Prot),
			BonusRequiredSolve: bonus_set.ItemCountsRequiredOptionsMake(
				bonus_set.CountMode_Minimum,
				bonusMake(setNameT16Prot, 2),
			),
			BonusRequiredWeight: new(bonusMake(setNameT15Prot, 0, setNameT16Prot, 2)),
		},
	}
}

func Model_PallyProtHeal() gear_model.SpecModel {
	return gear_model.SpecModel{
		Spec: stats.Spec_PaladinProt,
		ModelItems: gear_model.ModelItems{
			GearFile:           files.GearFileProtHeal,
			ReforgeRules:       gear_model.ReforgeRules_tank,
			Professions:        gear_model.ProfessionInfo{IsBlacksmith: true, IsEngineer: true},
			BlockSpecificItems: mygear.TrinketsStrengthMeleeOnly,
		},
		ModelSolve: gear_model.ModelSolve{
			WeightFile:        files.WeightProtHeal,
			SimPriority:       SimPriority_heal,
			StatsForWeighting: StatsForWeighting_strengthTank,
			StatRequirements:  requirements.StatRequirementsHitExpertise_ProtFlexibleParry(),
		},
		ModelSimulate: gear_model.ModelSimulate{
			Goal:       stats.OptimiseGoal_HalfMitiHeal,
			SimulateAs: stats.Fight_Juggernaut_OffHealer,
		},
		ModelBonus: gear_model.ModelBonus{
			BonusEnabled: bonus_set.SpecSetsEnableNamed(eternalFlameT16Prot),
			BonusRequiredSolve: bonus_set.ItemCountsRequiredOptionsMake(
				bonus_set.CountMode_AllowPlusOne,
				bonusMake(eternalFlameT16Prot, 4),
			),
			BonusRequiredWeight: new(bonusMake(eternalFlameT16Prot, 4)),
		},
	}
}

func Model_PallyProtMitigation() gear_model.SpecModel {
	return gear_model.SpecModel{
		Spec: stats.Spec_PaladinProt,
		ModelItems: gear_model.ModelItems{
			GearFile:           files.GearFileProtMitigation,
			ReforgeRules:       gear_model.ReforgeRules_tank,
			Professions:        gear_model.ProfessionInfo{IsBlacksmith: true, IsEngineer: true},
			BlockSpecificItems: mygear.TrinketsStrengthMeleeOnly,
		},
		ModelSolve: gear_model.ModelSolve{
			WeightFile:        files.WeightProtMitigation,
			SimPriority:       SimPriority_mitigation,
			StatsForWeighting: StatsForWeighting_strengthTank,
			StatRequirements:  requirements.StatRequirementsHitExpertise_ProtFlexibleParry(),
		},
		ModelSimulate: gear_model.ModelSimulate{
			Goal:       stats.OptimiseGoal_Mitigation,
			SimulateAs: stats.Fight_Juggernaut_NoExternalHeal,
		},
		ModelBonus: gear_model.ModelBonus{
			BonusEnabled: bonus_set.SpecSetsEnableNamed(setNameT16Prot),
			BonusRequiredSolve: bonus_set.ItemCountsRequiredOptionsMake(
				bonus_set.CountMode_AllowPlusOne, // no real justification for any restriction
				bonusMake(setNameT16Prot, 0),
				bonusMake(setNameT16Prot, 2),
			),
			BonusRequiredWeight: new(bonusMake(setNameT16Prot, 2, setNameT15Prot, 0)),
		},
	}
}

func Model_PallyProtBalanced() gear_model.SpecModel {
	return gear_model.SpecModel{
		Spec: stats.Spec_PaladinProt,
		ModelItems: gear_model.ModelItems{
			GearFile:           files.GearFileProtBalanced,
			ReforgeRules:       gear_model.ReforgeRules_tank,
			Professions:        gear_model.ProfessionInfo{IsBlacksmith: true, IsEngineer: true},
			BlockSpecificItems: mygear.TrinketsStrengthMeleeOnly,
		},
		ModelSolve: gear_model.ModelSolve{
			WeightFile:        files.WeightProtBalanced,
			SimPriority:       SimPriority_balanced,
			StatsForWeighting: StatsForWeighting_strengthTank,
			StatRequirements:  requirements.StatRequirementsHitExpertise_ProtFlexibleParry(),
		},
		ModelSimulate: gear_model.ModelSimulate{
			Goal:       stats.OptimiseGoal_HalfMitiDps,
			SimulateAs: stats.Fight_Juggernaut_HighHeal,
		},
		ModelBonus: gear_model.ModelBonus{
			BonusEnabled: bonus_set.SpecSetsEnableNamed(setNameT16Prot),
			BonusRequiredSolve: bonus_set.ItemCountsRequiredOptionsMake(
				bonus_set.CountMode_Exact,
				bonusMake(setNameT16Prot, 0),
				bonusMake(setNameT16Prot, 2)),
			BonusRequiredWeight: new(bonusMake(setNameT16Prot, 0)),
		},
	}
}

func Model_PallyProtDamage() gear_model.SpecModel {
	return gear_model.SpecModel{
		Spec: stats.Spec_PaladinProt,
		ModelItems: gear_model.ModelItems{
			GearFile:           files.GearFileProtDamage,
			ReforgeRules:       gear_model.ReforgeRules_tank,
			Professions:        gear_model.ProfessionInfo{IsBlacksmith: true, IsEngineer: true},
			BlockSpecificItems: mygear.TrinketsStrengthMeleeOnly,
		},
		ModelSolve: gear_model.ModelSolve{
			WeightFile:        files.WeightProtDamage,
			SimPriority:       SimPriority_dps,
			StatsForWeighting: StatsForWeighting_strengthTank,
			StatRequirements:  requirements.StatRequirementsHitExpertise_ProtFlexibleParry(),
		},
		ModelSimulate: gear_model.ModelSimulate{
			Goal:       stats.OptimiseGoal_Dps,
			SimulateAs: stats.Fight_Horridon_HighHeal,
		},
		ModelBonus: gear_model.ModelBonus{
			BonusEnabled:        bonus_set.SpecSetsEnableNone(),
			BonusRequiredSolve:  bonus_set.ItemCountsRequiredOptionsAny(),
			BonusRequiredWeight: nil,
		},
	}
}

func Model_PallyRet() gear_model.SpecModel {
	priority := SimPriority_ret
	return gear_model.SpecModel{
		Spec: stats.Spec_PaladinRet,
		ModelItems: gear_model.ModelItems{
			GearFile:           files.GearFileRet,
			ReforgeRules:       gear_model.ReforgeRules_melee,
			Professions:        gear_model.ProfessionInfo{IsBlacksmith: true, IsEngineer: true},
			BlockSpecificItems: mygear.TrinketsStrengthTankOnly,
		},
		ModelSolve: gear_model.ModelSolve{
			WeightFile:        files.WeightRet,
			SimPriority:       priority,
			StatsForWeighting: StatsForWeighting_strengthMelee,
			StatRequirements:  requirements.StatRequirementsHitExpertise_RetWideCap(),
		},
		ModelSimulate: gear_model.ModelSimulate{
			Goal:       stats.OptimiseGoal_Dps,
			SimulateAs: stats.Fight_Horridon_HighHeal,
			SimSpeedUp: 4,
		},
		ModelBonus: gear_model.ModelBonus{
			BonusEnabled: bonus_set.SpecSetsEnableNamed(setNameT16Ret),
			BonusRequiredSolve: bonus_set.ItemCountsRequiredOptionsMake(
				bonus_set.CountMode_Minimum,
				bonusMake(setNameT16Ret, 2),
			),
			BonusRequiredWeight: new(bonusMake(setNameT15Ret, 0, setNameT16Ret, 2)),
		},
	}
}
