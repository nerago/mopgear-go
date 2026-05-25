package main

import (
	"encoding/json"
	"os"
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/model"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/build"
	"paladin_gearing_go/solver/stathighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"strconv"
	"strings"
)

// old code for spreadsheet weights
func forSpreadsheetGenerateRatingsDataFromSims(printer *util.PrintRecorder) {
	// simSpeed := simulate.RunSize_QuickDirty
	simSpeed := simulate.RunSize_SlowAccurate

	// fight := stats.Fight_Animus
	// spec := stats.Spec_PaladinProtMitigation
	// startGear := files.GearFileProtMitigationNoSet
	// modelEquipOnly := model.Model_PallyProtMitigation_NoSet()

	fight := stats.Fight_Horridon_LowHeal
	spec := stats.Spec_PaladinProtMitigation
	startGear := files.GearFileProtMitigationSet
	modelEquipOnly := model.Model_PallyProtMitigation_WithSet()

	// fight := stats.Fight_Horridon_HighHeal
	// spec := stats.Spec_PaladinProtDps
	// startGear := files.GearFileProtDps
	// modelEquipOnly := model.Model_PallyProtDps()

	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(startGear), &modelEquipOnly, printer)

	baseStat := stats.StatBlock_of(stats.Stat_Haste, 500) // for miti sets avoid breakpoint
	// baseStat := stats.StatBlock_of(stats.Stat_Haste, 0)
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
	csv.InitRows(len(simulate.SimResultTypeList) + 1)

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

func forBasicStatsGenerateRatingsDataFromSims(printer *util.PrintRecorder) {
	// simSpeed := simulate.RunSize_QuickDirty
	simSpeed := simulate.RunSize_SlowAccurate

	// fight := stats.Fight_Animus
	// spec := stats.Spec_PaladinProtMitigation
	// startGear := files.GearFileProtMitigationNoSet
	// modelEquipOnly := model.Model_PallyProtMitigation_NoSet()

	fight := stats.Fight_Horridon_LowHeal
	spec := stats.Spec_PaladinProtMitigation
	startGear := files.GearFileProtMitigationSet
	modelEquipOnly := model.Model_PallyProtMitigation_WithSet()
	targetRatio := stathighs.NewStatWeights_generalMiti

	// fight := stats.Fight_Horridon_HighHeal
	// spec := stats.Spec_PaladinProtDps
	// startGear := files.GearFileProtDps
	// modelEquipOnly := model.Model_PallyProtDps()

	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(startGear), &modelEquipOnly, printer)

	baseStat := stats.StatBlock_of(stats.Stat_Haste, 500) // for miti sets avoid breakpoint
	// baseStat := stats.StatBlock_of(stats.Stat_Haste, 0)
	// var statAdd uint32 = 50
	// var statAdd uint32 = 200
	var statAdd uint32 = 400
	// var statAdd uint32 = 600
	statCheckList := []stats.StatType{
		stats.Stat_Strength, stats.Stat_Stamina, stats.Stat_Crit, stats.Stat_Haste,
		stats.Stat_Expertise, stats.Stat_Mastery, stats.Stat_Dodge, stats.Stat_Parry,
	}

	process := stathighs.BasicStatWeightProcess{}
	process.Init(printer)
	process.SetTargetRatios(targetRatio)

	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(len(statCheckList) + 1)
	defer tracker.Stop()

	simBase := simulate.WowSim_Execute_SelectFight(simSpeed, spec, fight, &currentEquip, modelEquipOnly.Professions, &baseStat, tracker.MakeNested())
	process.SetBaseline(simBase)

	for _, statCheck := range statCheckList {
		bonusStat := baseStat
		bonusStat[statCheck] += statAdd
		simResult := simulate.WowSim_Execute_SelectFight(simSpeed, spec, fight, &currentEquip, modelEquipOnly.Professions, &bonusStat, tracker.MakeNested())
		process.AddSimData(statCheck, statAdd, simResult)
	}

	process.Run()
}

// oldish code, may sometimes want to mix basic ratings??
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

func generateRatingsInputFromArtificalStatOverrides(printer *util.PrintRecorder) ([]stathighs.WeightInput, simulate.SimResultStats) {
	simSpeed := simulate.RunSize_TestOnly
	// simSpeed := simulate.RunSize_QuickDirty
	// simSpeed := simulate.RunSize_SlowAccurate

	// fight := stats.Fight_Animus
	// spec := stats.Spec_PaladinProtMitigation
	// startGear := files.GearFileProtMitigationNoSet
	// modelEquipOnly := model.Model_PallyProtMitigation_NoSet()
	// targetRatio := stathighs.NewStatWeights_generalMiti

	// fight := stats.Fight_Horridon_LowHeal
	// spec := stats.Spec_PaladinProtMitigation
	// startGear := files.GearFileProtMitigationSet
	// modelEquipOnly := model.Model_PallyProtMitigation_WithSet()
	// targetRatio := stathighs.NewStatWeights_radenWeight

	fight := stats.Fight_Animus
	spec := stats.Spec_PaladinProtCompromise
	startGear := files.GearFileProtCompromise
	modelEquipOnly := model.Model_PallyProtCompromise()
	targetRatio := stathighs.NewStatWeights_animusWeight

	// fight := stats.Fight_Horridon_HighHeal
	// spec := stats.Spec_PaladinProtDps
	// startGear := files.GearFileProtDps
	// modelEquipOnly := model.Model_PallyProtDps()
	// targetRatio := stathighs.NewStatWeights_dpsWeight

	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(startGear), &modelEquipOnly, printer)
	currentItemSet := items.FullItemSet_FromMap(currentEquip)
	currentHaste := currentItemSet.Total().Get(stats.Stat_Haste)
	var incrementBaseHaste uint32 = 1900
	printer.Printf("Current gear haste %d\n", currentHaste)
	printer.Printf("Simulated minimum gear haste %d\n", currentHaste+incrementBaseHaste)
	if currentHaste+incrementBaseHaste < 14000 {
		panic("haste in discontinuity range")
	}

	initialBaseStats := stats.StatBlock{}
	initialBaseStats[stats.Stat_Haste] += incrementBaseHaste

	var incrementMin uint32 = 0
	var incrementMax uint32 = 500
	var incrementStep uint32 = 250

	statCheckList := stathighs.G_RequiredStats
	type incrementStat struct {
		stat  stats.StatType
		value uint32
	}

	incrementOptions := make([][]incrementStat, 0)
	for _, stat := range statCheckList {
		optionArray := make([]incrementStat, 0)
		for value := incrementMin; value < incrementMax; value += incrementStep {
			entry := incrementStat{stat, value}
			optionArray = append(optionArray, entry)
		}
		incrementOptions = append(incrementOptions, optionArray)
	}

	incrementPermutations := util.PermuteAll_Slice(incrementOptions)

	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(len(incrementPermutations))
	defer tracker.Stop()

	inputList := channel_op.Map_SliceToSlice(6, incrementPermutations, func(increments *[]incrementStat, resultChannel chan<- stathighs.WeightInput) {
		innerPrint := util.PrintRecorder_HoldAll()

		bonusStat := initialBaseStats
		str := util.StringBuild2{}
		str.WriteString("STATS SCENARIO ")
		for _, inc := range *increments {
			bonusStat[inc.stat] += inc.value
			str.WriteString(inc.stat.Name())
			str.WriteRune('=')
			str.WriteUint32(bonusStat[inc.stat])
			str.WriteRune(' ')
		}

		simResult := simulate.WowSim_Execute_SelectFight(simSpeed, spec, fight, &currentEquip, modelEquipOnly.Professions, &bonusStat, tracker.MakeNested())

		resultChannel <- stathighs.WeightInput{
			TotalStat: bonusStat,
			SimResult: simResult,
		}

		innerPrint.PrintlnFromBuild(str)
		innerPrint.Println("   --> " + simResult.CompactStringGeneral())

		printer.AppendOther(innerPrint)
	})

	// bytes, err := json.Marshal(inputList)
	// if err != nil {
	// 	panic(err)
	// }
	// err = os.WriteFile("sim-stats-grid-data.json", bytes, 0)
	// if err != nil {
	// 	panic(err)
	// }

	return inputList, targetRatio
}

func generateRatingsInputFromRealRandomSets(printer *util.PrintRecorder) ([]stathighs.WeightInput, simulate.SimResultStats) {
	makeSetCount := 256
	simSize := simulate.RunSize_Medium

	targetRatio := stathighs.NewStatWeights_generalMiti
	model := model.Model_PallyProtMitigation_NoSet()
	itemOptions := setup.OptionsSetup_FromGearFile(files.GearFileProtMitigationNoSet, &model, setup.MissingEnchant_Panic, printer)
	for _, itemId := range substituteItemsMiti {
		if !itemOptions.IncludesItemId(itemId) {
			opts, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, 2, &model, printer)
			for _, slotEquip := range example.SlotItem().ToSlotEquipOptions() {
				if itemOptions.Has(slotEquip) {
					itemOptions.AddSeveralOptionsSpecific(slotEquip, opts)
				}
			}
		}
	}
	itemOptions.RemoveItemIdFromAll(95141)

	setList := build.SolverBuildRandom_MakeN_FullAndValidate(&itemOptions, &model, makeSetCount, printer, 14000)

	track := util.TrackProgress_Start()
	track.RunOuterTracking(len(setList))
	defer track.Stop()

	weightInputs := channel_op.Map_SliceToSlice(6, setList, func(itemSet *items.FullItemSet, weightInputs chan<- stathighs.WeightInput) {
		simResult := simulate.WowSim_Execute_SelectFight(simSize, model.Spec, model.SimulateAs, itemSet.Items(), model.Professions, nil, track.MakeNested())
		weightInputs <- stathighs.WeightInput{TotalStat: *itemSet.Total(), SimResult: simResult}
	})

	bytes, err := json.Marshal(weightInputs)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("sim-stats-input-data.json", bytes, 0)
	if err != nil {
		panic(err)
	}

	return weightInputs, targetRatio
}

func statWeightsFromHighAndSim(printer *util.PrintRecorder) {
	weightInputs, targetRatio := generateRatingsInputFromRealRandomSets(printer)

	// bytes, err := os.ReadFile("sim-stats-input-data.json")
	// if err != nil {
	// 	panic(err)
	// }
	// var weightInputs []stathighs.WeightInput
	// err = json.Unmarshal(bytes, &weightInputs)
	// if err != nil {
	// 	panic(err)
	// }

	// for _, entry := range weightInputs {
	// 	if entry.TotalStat.Get(stats.Stat_Haste) < 14000 {
	// 		panic("haste in discontinuity range")
	// 	}
	// }

	weights := stathighs.CalcComplexStatWeights(weightInputs, targetRatio, printer)
	writePawnString(weights, printer)
}

func statWeightsBasic(printer *util.PrintRecorder) {
	process := stathighs.BasicStatWeightProcess{}
	process.Init(printer)
	process.SetTargetRatios(stathighs.NewStatWeights_generalMiti)
	process.SetBaseline(parseSimStats("254619.21 1604831.48 27870.13 39389.66 56.82 14.23"))
	process.AddSimData(stats.Stat_Strength, +600, parseSimStats("256235.27 1614633.03 27573.09 39660.8 56.16 12.89"))
	process.AddSimData(stats.Stat_Stamina, +600, parseSimStats("254474.09 1603914.71 27941.9 39360.88 55.72 13.62"))
	process.AddSimData(stats.Stat_Crit, +600, parseSimStats("257106.61 1620383.56 27870.13 39389.66 56.82 14.23"))
	process.AddSimData(stats.Stat_Haste, +600, parseSimStats("256815.91 1619941.66 27591.27 39782.48 55.98 12.51"))
	process.AddSimData(stats.Stat_Expertise, +600, parseSimStats("256349.38 1615203.97 27893.43 40077.8 56.72 13.16"))
	process.AddSimData(stats.Stat_Mastery, +600, parseSimStats("254483.79 1603982.91 27230.32 39355.29 55.17 12.03"))
	process.AddSimData(stats.Stat_Dodge, +600, parseSimStats("254623.68 1604870.45 27649.33 39384.54 56.34 13.46"))
	process.AddSimData(stats.Stat_Parry, +600, parseSimStats("254649.78 1605018.37 27660.8 39408.39 56.36 13.45"))
	weights := process.Run()
	writePawnString(weights, printer)
}

func statWeightsGrid(printer *util.PrintRecorder) {
	inputData, targetRatio := generateRatingsInputFromArtificalStatOverrides(printer)
	process := stathighs.GridStatWeightProcess{}
	process.Init(printer)
	process.SetTargetRatios(targetRatio)
	process.SupplyData(inputData)
	weights := process.Run()
	writePawnString(weights, printer)
}

func writePawnString(weights map[stats.StatType]float64, printer *util.PrintRecorder) {
	str := util.StringBuild2{}
	str.WriteString("( Pawn: v1: \"Protection WoWSims Weights\": Class=Paladin,Strength=")
	str.WriteFloat64(weights[stats.Stat_Strength], 10)
	str.WriteString(",Stamina=")
	str.WriteFloat64(weights[stats.Stat_Stamina], 10)
	str.WriteString(",CritRating=")
	str.WriteFloat64(weights[stats.Stat_Crit], 10)
	str.WriteString(",HasteRating=")
	str.WriteFloat64(weights[stats.Stat_Haste], 10)
	str.WriteString(",ExpertiseRating=")
	str.WriteFloat64(weights[stats.Stat_Expertise], 10)
	str.WriteString(",MasteryRating=")
	str.WriteFloat64(weights[stats.Stat_Mastery], 10)
	str.WriteString(",DodgeRating=")
	str.WriteFloat64(weights[stats.Stat_Dodge], 10)
	str.WriteString(",ParryRating=")
	str.WriteFloat64(weights[stats.Stat_Parry], 10)
	str.WriteString(", )")
	printer.PrintlnFromBuild(str)
}

func parseSimStats(str string) simulate.SimResultStats {
	result := simulate.SimResultStats{}
	parts := strings.Split(str, " ")
	for i, simType := range simulate.SimResultTypeList {
		value, err := strconv.ParseFloat(parts[i], 64)
		if err != nil {
			panic(err)
		}
		result.Set(simType, value)
	}
	return result
}
