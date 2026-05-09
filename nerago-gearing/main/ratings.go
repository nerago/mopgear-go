package main

import (
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/model"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

func generateRatingsDataFromSims(printer *util.PrintRecorder) {
	// simSpeed := simulate.RunSize_QuickDirty
	simSpeed := simulate.RunSize_SlowAccurate

	// fight := simulate.Fight_Animus
	// spec := stats.Spec_PaladinProtMitigation
	// startGear := files.GearFileProtMitigationNoSet
	// modelEquipOnly := model.Model_PallyProtMitigation_NoSet()

	// fight := simulate.Fight_Horridon_LowHeal
	// spec := stats.Spec_PaladinProtMitigation
	// startGear := files.GearFileProtMitigationSet
	// modelEquipOnly := model.Model_PallyProtMitigation_WithSet()

	fight := stats.Fight_Horridon_HighHeal
	spec := stats.Spec_PaladinProtDps
	startGear := files.GearFileProtDps
	modelEquipOnly := model.Model_PallyProtDps()

	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(startGear), &modelEquipOnly, printer)

	// baseStat := stats.StatBlock_of(stats.Stat_Haste, 500) // for miti sets avoid breakpoint
	baseStat := stats.StatBlock_of(stats.Stat_Haste, 0)
	// var statAdd uint32 = 50
	// var statAdd uint32 = 200
	var statAdd uint32 = 400
	// var statAdd uint32 = 600
	statCheckList := []stats.StatType{
		stats.Stat_Strength, stats.Stat_Stamina, stats.Stat_Crit, stats.Stat_Haste,
		stats.Stat_Expertise, stats.Stat_Mastery, stats.Stat_Dodge, stats.Stat_Parry,
	}

	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(len(statCheckList) + 1)
	defer tracker.Stop()

	csv := util.CSVOutputByColumn{}

	simBase := simulate.WowSim_Execute_SelectFight(simSpeed, spec, fight, &currentEquip, modelEquipOnly.Professions, &baseStat, tracker.MakeNested())
	csv.AddString("base")
	simResultAddToCSV(simBase, &csv)
	csv.FinishColumn()

	for _, statCheck := range statCheckList {
		bonusStat := baseStat
		bonusStat[statCheck] += statAdd
		simResult := simulate.WowSim_Execute_SelectFight(simSpeed, spec, fight, &currentEquip, modelEquipOnly.Professions, &bonusStat, tracker.MakeNested())

		csv.AddString(statCheck.Name())
		simResultAddToCSV(simResult, &csv)
		csv.FinishColumn()
	}

	csv.Write(printer)
}

func simResultAddToCSV(simResult simulate.SimResultStats, csv *util.CSVOutputByColumn) {
	for _, simType := range simulate.SimResultTypeList {
		num := simResult.Get(simType)
		if simType == simulate.Result_DEATH {
			num *= 100
		}
		csv.AddFloat64(num, 2)
	}
}

func relativeRatingsCompromise(printer *util.PrintRecorder) {
	modelMitiNoSet := model.Model_PallyProtMitigation_NoSet()
	gearMitiNoSet := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(files.GearFileProtMitigationNoSet), &modelMitiNoSet, printer)
	itemSetMitiNoSet := items.FullItemSet_FromMap(gearMitiNoSet)

	modelDps := model.Model_PallyProtDps()
	gearDps := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(files.GearFileProtDps), &modelDps, printer)
	itemSetDps := items.FullItemSet_FromMap(gearDps)

	var targetCombined = 10000000000.0
	targetRatio := 0.7

	rateA1 := modelMitiNoSet.CalcRatingFull(&itemSetMitiNoSet)
	rateA2 := modelDps.CalcRatingFull(&itemSetMitiNoSet)
	printer.Printf("A %f %f\n", rateA1, rateA2)
	multA1 := targetCombined * targetRatio / rateA1
	multA2 := targetCombined * (1 - targetRatio) / rateA2
	printer.Printf("* %f %f\n", multA1, multA2)
	printer.Printf("? %f %f\n", multA1*rateA1/targetCombined, multA2*rateA2/targetCombined)

	rateB1 := modelMitiNoSet.CalcRatingFull(&itemSetDps)
	rateB2 := modelDps.CalcRatingFull(&itemSetDps)
	printer.Printf("B %f %f\n", rateB1, rateB2)
	multB1 := targetCombined * targetRatio / rateB1
	multB2 := targetCombined * (1 - targetRatio) / rateB2
	printer.Printf("* %f %f\n", multB1, multB2)
	printer.Printf("? %f %f\n", multB1*rateB1/targetCombined, multB2*rateB2/(targetCombined))
}
