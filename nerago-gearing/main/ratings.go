package main

import (
	"cmp"
	"encoding/json"
	"maps"
	"math"
	"os"
	"paladin_gearing_go/files"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/model"
	"paladin_gearing_go/model/requirements"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/build"
	"paladin_gearing_go/solver/stathighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"slices"
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
	// goal := stats.UpgradeGoal_Mitigation

	fight := stats.Fight_Horridon_LowHeal
	spec := stats.Spec_PaladinProt
	startGear := files.GearFileProtMitigationWithSet
	modelEquipOnly := model.Model_PallyProtMitigation_WithSet()
	goal := stats.OptimiseGoal_Mitigation

	// fight := stats.Fight_Horridon_HighHeal
	// spec := stats.Spec_PaladinProtDps
	// startGear := files.GearFileProtDps
	// modelEquipOnly := model.Model_PallyProtDps()
	// goal := stats.UpgradeGoal_Dps

	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(startGear), &modelEquipOnly, printer)

	// var statAdd uint32 = 50
	// var statAdd uint32 = 200
	var statAdd int32 = 400
	// var statAdd uint32 = 600

	var incrementBaseHaste int32 = 0
	var decrementBaseExpertise int32 = 0
	baseStat := initialBonusStatMap(printer, items.FullItemSet_FromMap(currentEquip), incrementBaseHaste, decrementBaseExpertise, statAdd)

	statCheckList := []stats.StatType{
		stats.Stat_Strength, stats.Stat_Stamina, stats.Stat_Crit, stats.Stat_Haste,
		stats.Stat_Expertise, stats.Stat_Mastery, stats.Stat_Dodge, stats.Stat_Parry,
	}

	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(len(statCheckList) + 1)
	defer tracker.Stop()

	csv := util.CSVOutputByColumn{}
	csv.InitRows(len(simulate.SimTypeList) + 1)

	simBase := simulate.WowSim_Execute_SpecifyAll(simSpeed, spec, goal, fight, modelEquipOnly.Professions, &currentEquip, &baseStat, tracker.MakeNested())
	csv.AddString("base")
	simResultAddToCSV(simBase, &csv)
	csv.FinishColumn()

	for _, statCheck := range statCheckList {
		bonusStat := maps.Clone(baseStat)
		bonusStat[statCheck] += statAdd
		simResult := simulate.WowSim_Execute_SpecifyAll(simSpeed, spec, goal, fight, modelEquipOnly.Professions, &currentEquip, &bonusStat, tracker.MakeNested())

		csv.AddString(statCheck.Name())
		simResultAddToCSV(simResult, &csv)
		csv.FinishColumn()
	}

	csv.Write(printer)
}

func simResultAddToCSV(simResult simulate.SimData, csv *util.CSVOutputByColumn) {
	for _, simType := range simulate.SimTypeList {
		num := simResult.Get(simType)
		if simType == simulate.Sim_DEATH {
			num *= 100
		}
		csv.AddFloat64(num, 2)
	}
}

func testBasicStatsGeneral(printer *util.PrintRecorder) {
	// simSpeed := simulate.RunSize_QuickDirty
	simSpeed := simulate.RunSize_SlowAccurate

	// fight := stats.Fight_Animus
	// spec := stats.Spec_PaladinProtMitigation
	// startGear := files.GearFileProtMitigationNoSet
	// modelEquipOnly := model.Model_PallyProtMitigation_NoSet()
	// goal := stats.UpgradeGoal_Mitigation

	fight := stats.Fight_Horridon_LowHeal
	spec := stats.Spec_PaladinProt
	startGear := files.GearFileProtMitigationWithSet
	modelEquipOnly := model.Model_PallyProtMitigation_WithSet()
	targetRatio := stathighs.NewStatWeights_generalMiti
	goal := stats.OptimiseGoal_Mitigation

	// fight := stats.Fight_Horridon_HighHeal
	// spec := stats.Spec_PaladinProtDps
	// startGear := files.GearFileProtDps
	// modelEquipOnly := model.Model_PallyProtDps()
	// goal := stats.UpgradeGoal_Dps

	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(startGear), &modelEquipOnly, printer)
	itemSet := items.FullItemSet_FromMap(currentEquip)

	inputData, simBase := generateRatingsInputFromArtificalStatOverrides_ForBasic(itemSet, printer, simSpeed, spec, goal, fight, modelEquipOnly.Professions)

	process := stathighs.BasicStatWeightProcess{}
	process.Init(printer)
	process.SetTargetRatios(targetRatio)
	process.SetBaseline(simBase)
	for _, data := range inputData {
		process.AddSimData(data.IncrementStat, uint32(data.IncrementValue), data.SimResult)
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

func generateRatingsInputFromArtificalStatOverrides(printer *util.PrintRecorder) ([]stathighs.WeightInput, simulate.SimData) {
	// simSpeed := simulate.RunSize_TestOnly
	simSpeed := simulate.RunSize_QuickDirty
	// simSpeed := simulate.RunSize_SlowAccurate

	// fight := stats.Fight_Animus
	// spec := stats.Spec_PaladinProtMitigation
	// startGear := files.GearFileProtMitigationNoSet
	// modelEquipOnly := model.Model_PallyProtMitigation_NoSet()
	// targetRatio := stathighs.NewStatWeights_generalMiti
	// goal := stats.UpgradeGoal_Mitigation

	// fight := stats.Fight_Horridon_LowHeal
	// spec := stats.Spec_PaladinProtMitigation
	// startGear := files.GearFileProtMitigationSet
	// modelEquipOnly := model.Model_PallyProtMitigation_WithSet()
	// targetRatio := stathighs.NewStatWeights_radenWeight
	// goal := stats.UpgradeGoal_Mitigation

	fight := stats.Fight_Animus
	spec := stats.Spec_PaladinProt
	startGear := files.GearFileProtCompromise
	modelEquipOnly := model.Model_PallyProtCompromise()
	targetRatio := stathighs.NewStatWeights_animusWeight
	goal := stats.OptimiseGoal_HalfMitiDps

	// fight := stats.Fight_Horridon_HighHeal
	// spec := stats.Spec_PaladinProtDps
	// startGear := files.GearFileProtDps
	// modelEquipOnly := model.Model_PallyProtDps()
	// targetRatio := stathighs.NewStatWeights_dpsWeight
	// goal := stats.UpgradeGoal_Dps

	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(startGear), &modelEquipOnly, printer)
	currentItemSet := items.FullItemSet_FromMap(currentEquip)
	inputList := generateRatingsInputFromArtificalStatOverrides_ForGrid(currentItemSet, printer, simSpeed, spec, goal, fight, modelEquipOnly.Professions)

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

func generateRatingsInputFromArtificalStatOverrides_ForGrid(currentItemSet items.FullItemSet, printer *util.PrintRecorder, simSpeed simulate.WowSim_RunSize, spec stats.SpecType, goal stats.OptimiseGoal, fight stats.WowSim_Fight, profession model.ProfessionInfo) []stathighs.WeightInput {
	var incrementMin int32 = 0
	var incrementMax int32 = 500
	var incrementStep int32 = 250

	initialBaseStats := initialBonusStatMap_fixRanges(printer, currentItemSet, incrementMax)

	statCheckList := stathighs.G_RequiredStats
	type incrementStat struct {
		stat  stats.StatType
		value int32
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

		bonusStat := maps.Clone(initialBaseStats)
		str := util.StringBuild2{}
		str.WriteString("STATS SCENARIO ")
		for _, inc := range *increments {
			bonusStat[inc.stat] += inc.value

			str.WriteString(inc.stat.Name())
			str.WriteRune('=')
			str.WriteInt32(bonusStat[inc.stat])
			str.WriteRune(' ')
		}

		simResult := simulate.WowSim_Execute_SpecifyAll(simSpeed, spec, goal, fight, profession, currentItemSet.Items(), &bonusStat, tracker.MakeNested())

		resultChannel <- stathighs.WeightInput{
			TotalStat: addBonusStats(currentItemSet.Total(), bonusStat),
			SimResult: simResult,
		}

		innerPrint.PrintlnFromBuild(str)
		innerPrint.Println("   --> " + simResult.CompactStringGeneral())

		printer.AppendOther(innerPrint)
	})
	return inputList
}

type basicStatInput struct {
	IncrementStat  stats.StatType
	IncrementValue int32
	SimResult      simulate.SimData
}

func generateRatingsInputFromArtificalStatOverrides_ForBasic(currentItemSet items.FullItemSet, printer *util.PrintRecorder, simSpeed simulate.WowSim_RunSize, spec stats.SpecType, goal stats.OptimiseGoal, fight stats.WowSim_Fight, profession model.ProfessionInfo) ([]basicStatInput, simulate.SimData) {
	var incrementValue int32 = 250

	initialBaseStats := initialBonusStatMap_fixRanges(printer, currentItemSet, incrementValue)

	statCheckList := stathighs.G_RequiredStats

	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(len(statCheckList) + 1)
	defer tracker.Stop()

	simBase := simulate.WowSim_Execute_SpecifyAll(simSpeed, spec, goal, fight, profession, currentItemSet.Items(), nil, tracker.MakeNested())

	inputList := channel_op.Map_SliceToSlice(len(statCheckList), statCheckList, func(incStat *stats.StatType, resultChannel chan<- basicStatInput) {
		innerPrint := util.PrintRecorder_HoldAll()

		bonusStat := maps.Clone(initialBaseStats)
		str := util.StringBuild2{}
		str.WriteString("STATS SCENARIO ")
		bonusStat[*incStat] += incrementValue
		str.WriteString(incStat.Name())
		str.WriteRune('=')
		str.WriteInt32(bonusStat[*incStat])
		str.WriteRune(' ')

		simResult := simulate.WowSim_Execute_SpecifyAll(simSpeed, spec, goal, fight, profession, currentItemSet.Items(), &bonusStat, tracker.MakeNested())

		resultChannel <- basicStatInput{
			IncrementStat:  *incStat,
			IncrementValue: incrementValue,
			SimResult:      simResult,
		}

		innerPrint.PrintlnFromBuild(str)
		innerPrint.Println("   --> " + simResult.CompactStringGeneral())

		printer.AppendOther(innerPrint)
	})
	return inputList, simBase
}

func generateRatingsInputFromRealRandomSetsT5(printer *util.PrintRecorder) ([]stathighs.WeightInput, simulate.SimData) {
	makeSetCount := 2000
	simSize := simulate.RunSize_Medium

	targetRatio := stathighs.NewStatWeights_generalMiti
	model := model.Model_PallyProtMitigation_NoSet()

	_, itemOptions := allT5stuff(&model, files.GearFileProtMitigationNoSet, printer)

	// setList := build.SolverBuildRandom_MakeN_FullAndValidate(&itemOptions, &model, makeSetCount, printer, 14000)
	setList := build.SolverBuildRandom_MakeN_FullAndValidate(&itemOptions, &model, makeSetCount, printer, 0)

	track := util.TrackProgress_Start()
	track.RunOuterTracking(len(setList))
	defer track.Stop()

	weightInputs := channel_op.Map_SliceToSlice(6, setList, func(itemSet *items.FullItemSet, weightInputs chan<- stathighs.WeightInput) {
		simResult := simulate.WowSim_Execute_UseModel(simSize, &model, itemSet.Items(), nil, track.MakeNested())
		weightInputs <- stathighs.WeightInput{TotalStat: *itemSet.Total(), SimResult: simResult}
	})

	writeWeightInputsToFile(weightInputs, "sim-stats-input-data2.json")

	return weightInputs, targetRatio
}

func generateRatingsInputFromRealRandomSetsGeneral(gearFile string, substituteItems []items.ItemId, model *model.Model, makeSetCount int, simSize simulate.WowSim_RunSize, doFixRanges bool) []stathighs.WeightInput {
	itemOptions := setup.OptionsSetup_FromGearFile(gearFile, model, setup.MissingEnchant_Panic, printer)
	for _, itemId := range substituteItems {
		opts, example := setup.OptionsSetup_Single_FromIdOnlyUseAllDefaults(itemId, 2, model, printer)
		for _, slotEquip := range example.SlotItem().ToSlotEquipOptions() {
			if itemOptions.Has(slotEquip) {
				itemOptions.AddSeveralOptionsSpecific(slotEquip, opts)
			}
		}
	}
	itemOptions.RemoveDuplicates()

	setList := build.SolverBuildRandom_MakeN_FullAndValidate(&itemOptions, model, makeSetCount, printer, 0)

	track := util.TrackProgress_Start()
	track.RunOuterTracking(len(setList))
	defer track.Stop()

	weightInputs := channel_op.Map_SliceToSlice(6, setList, func(itemSet *items.FullItemSet, weightInputs chan<- stathighs.WeightInput) {
		var bonusStats *map[stats.StatType]int32 = nil
		if doFixRanges {
			bonusFix := initialBonusStatMap_fixRanges(printer, *itemSet, 0)
			bonusStats = &bonusFix
		}

		simResult := simulate.WowSim_Execute_UseModel(simSize, model, itemSet.Items(), bonusStats, track.MakeNested())
		weightInputs <- stathighs.WeightInput{TotalStat: *itemSet.Total(), SimResult: simResult}
	})

	return weightInputs
}

func statWeightsComplex(printer *util.PrintRecorder) {
	// weightInputs, targetRatio := generateRatingsInputFromRealRandomSets(printer)

	weightInputs := readWeightInputFile("sim-stats-input-data2.json")

	// between := func(w *stathighs.WeightInput, stat stats.StatType, lo, hi uint32) bool {
	// 	value := w.TotalStat.Get(stat)
	// 	return lo <= value && value <= hi
	// }
	// filteredInput := util.FilterSliceAsNew(weightInputs, func(w *stathighs.WeightInput) bool {
	// 	return between(w, stats.Stat_Haste, 3187, 8675) &&
	// 		between(w, stats.Stat_Expertise, 2805, 4950) &&
	// 		between(w, stats.Stat_Mastery, 5767, 11787) &&
	// 		between(w, stats.Stat_Dodge, 4271, 8342) &&
	// 		between(w, stats.Stat_Crit, 0, 3399)
	// })
	// filteredInput := util.FilterSliceAsNew(weightInputs, func(w *stathighs.WeightInput) bool {
	// 	return between(w, stats.Stat_Haste, 0, 8200)
	// })

	filteredInput := weightInputs
	printer.Printf("filteredInput size %d\n", len(filteredInput))

	comp := stathighs.ComplexStatWeightProcess{}

	comp.Init(printer)
	comp.SetTargetRatios(stathighs.NewStatWeights_generalMiti)
	comp.SetMinimumIncludeRate(0.7)
	comp.SupplyData(filteredInput)
	weights := comp.Run()
	writePawnString(weights, printer)
}

func statWeightsRanking(printer *util.PrintRecorder) {
	// weightInputs, targetRatio := generateRatingsInputFromRealRandomSets(printer)

	weightInputs := readWeightInputFile("sim-stats-input-data2.json")

	filteredInput := weightInputs
	printer.Printf("filteredInput size %d\n", len(filteredInput))

	ranking := stathighs.RankingStatWeightProcess{}

	ranking.Init(printer)
	ranking.SetTargetRatios(stathighs.NewStatWeights_generalMiti)
	// ranking.SetMinimumIncludeRate(0.7)
	ranking.SupplyData(filteredInput)
	weights := ranking.Run()
	writePawnString(weights, printer)
}

func statWeightsFitting(printer *util.PrintRecorder) {
	// generateRatingsInputFromRealRandomSets(printer)

	bytes, err := os.ReadFile("sim-stats-input-data2.json")
	// bytes, err := os.ReadFile("sim-stats-input-data.json")
	if err != nil {
		panic(err)
	}
	var weightInputs []stathighs.WeightInput
	err = json.Unmarshal(bytes, &weightInputs)
	if err != nil {
		panic(err)
	}

	for _, entry := range weightInputs {
		if hasteInDiscontinuityRange(entry.TotalStat.GetUInt(stats.Stat_Haste)) {
			printer.Println("haste in discontinuity range")
		}
	}

	printer.Printf("Initial weight input size = %d\n", len(weightInputs))

	// fitting := stathighs.FittingSingleStatWeightProcess{}
	// fitting.Init(printer)
	// fitting.SetMinimumIncludeRate(0.2)
	// fitting.SupplyDataFromStandard(weightInputs[0:32], stats.Stat_Crit, simulate.Result_DPS)
	// oneWeight := fitting.Run()
	// printer.Printf("%f %f %f %f %f\n", oneWeight.LineSlope, oneWeight.LineOffset, oneWeight.Minimum, oneWeight.Maximum, oneWeight.IncludePercent)
	// writePawnString(weights, printer)

	fitting := stathighs.FittingSingleStatSegmentsProcess{}
	// fitting.Init(printer, stats.Stat_Crit, simulate.Result_DPS)
	fitting.Init(printer, stats.Stat_Haste, simulate.Sim_DPS)
	fitting.SupplyDataFromStandard(weightInputs)

	weightMap := fitting.Run()
	printer.Printf("weightMap size %d\n", len(weightMap))
	weightList := slices.SortedFunc(maps.Values(weightMap), func(a, b stathighs.FittingSingleStatResult) int { return cmp.Compare(a.Minimum, b.Minimum) })

	tab := util.TabulateOutput{}
	tab.SetColumnSpacing(1)
	tab.AddColumnHeader("min", true)
	tab.AddColumnHeader("max", false)
	tab.AddColumnHeader("m", true)
	tab.AddColumnHeader("c", false)
	tab.AddColumnHeader("used", true)
	tab.AddColumnHeader("used%", false)
	for _, oneWeight := range weightList {
		tab.AddRow([]string{
			strconv.FormatUint(uint64(oneWeight.Minimum), 10),
			strconv.FormatUint(uint64(oneWeight.Maximum), 10),
			strconv.FormatFloat(oneWeight.LineSlope, 'f', 6, 64),
			strconv.FormatFloat(oneWeight.LineOffset, 'f', 1, 64),
			strconv.FormatUint(uint64(oneWeight.IncludeCount), 10),
			strconv.FormatFloat(oneWeight.IncludePercent*100, 'f', 1, 64),
		})
		// printer.Printf("%f %f %d %d %f\n", oneWeight.LineSlope, oneWeight.LineOffset, oneWeight.Minimum, oneWeight.Maximum, oneWeight.IncludePercent)
	}
	tab.Write(printer)
	// writePawnString(weights, printer)
}

func statWeightsFitting2(printer *util.PrintRecorder) {
	bytes, err := os.ReadFile("sim-stats-input-data2.json")
	// bytes, err := os.ReadFile("sim-stats-input-data.json")
	if err != nil {
		panic(err)
	}
	var weightInputs []stathighs.WeightInput
	err = json.Unmarshal(bytes, &weightInputs)
	if err != nil {
		panic(err)
	}

	for _, entry := range weightInputs {
		if hasteInDiscontinuityRange(entry.TotalStat.GetUInt(stats.Stat_Haste)) {
			printer.Println("haste in discontinuity range")
		}
	}

	printer.Printf("Initial weight input size = %d\n", len(weightInputs))

	fitting := stathighs.FittingEachStatWeightProcess{}
	fitting.Init(printer)
	fitting.SupplyDataFromStandard(weightInputs)

	weightMapMapMap := fitting.RunDetailedResults()
	for entry := range weightMapMapMap.SeqWithKeys() {
		weightMap := entry.Value

		printer.Printf("################### %s %s ###################\n", entry.Key1.Name(), entry.Key2.Name())
		printer.Printf("weightMap size %d\n", len(weightMap))
		weightList := slices.SortedFunc(maps.Values(weightMap), func(a, b stathighs.FittingSingleStatResult) int { return cmp.Compare(a.Minimum, b.Minimum) })

		tab := util.TabulateOutput{}
		tab.SetColumnSpacing(1)
		tab.AddColumnHeader("min", true)
		tab.AddColumnHeader("max", false)
		tab.AddColumnHeader("m", true)
		tab.AddColumnHeader("c", false)
		tab.AddColumnHeader("used", true)
		tab.AddColumnHeader("used%", false)
		for _, oneWeight := range weightList {
			tab.AddRow([]string{
				strconv.FormatUint(uint64(oneWeight.Minimum), 10),
				strconv.FormatUint(uint64(oneWeight.Maximum), 10),
				strconv.FormatFloat(oneWeight.LineSlope, 'f', 6, 64),
				strconv.FormatFloat(oneWeight.LineOffset, 'f', 1, 64),
				strconv.FormatUint(uint64(oneWeight.IncludeCount), 10),
				strconv.FormatFloat(oneWeight.IncludePercent*100, 'f', 1, 64),
			})
			// printer.Printf("%f %f %d %d %f\n", oneWeight.LineSlope, oneWeight.LineOffset, oneWeight.Minimum, oneWeight.Maximum, oneWeight.IncludePercent)
		}
		tab.Write(printer)
	}
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
	// inputData, targetRatio := generateRatingsInputFromArtificalStatOverrides(printer)
	// writeWeightInputsToFile(inputData, "sim-stats-input-grid.json" )

	inputData := readWeightInputFile("sim-stats-input-grid.json")
	targetRatio := stathighs.NewStatWeights_animusWeight

	process := stathighs.GridStatWeightProcess2{}
	process.Init(printer)
	process.SetTargetRatios(targetRatio)
	process.SupplyData(inputData)
	weights := process.Run()
	writePawnString(weights, printer)
}

func writePawnString(weights map[stats.StatType]float64, printer *util.PrintRecorder) string {
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
	return str.String()
}

func parseSimStats(str string) simulate.SimData {
	result := simulate.SimData{}
	parts := strings.Split(str, " ")
	for i, simType := range simulate.SimTypeList {
		value, err := strconv.ParseFloat(parts[i], 64)
		if err != nil {
			panic(err)
		}
		result.Set(simType, value)
	}
	return result
}

const c_hasteDiscontinuityStart = 10500
const c_hasteDiscontinuityEnd = 14000

func hasteInDiscontinuityRange(value uint32) bool {
	return value >= c_hasteDiscontinuityStart && value <= c_hasteDiscontinuityEnd
}

func checkBadHasteRange(printer *util.PrintRecorder, currentHaste uint32, incrementBaseHaste int32, plannedIncrementTestRange int32) bool {
	printer.Printf("Current gear haste %d\n", currentHaste)
	min := int32(currentHaste) + incrementBaseHaste
	max := int32(currentHaste) + incrementBaseHaste + plannedIncrementTestRange
	printer.Printf("Planned simulated gear haste %d-%d\n", min, max)

	return max > c_hasteDiscontinuityStart && min < c_hasteDiscontinuityEnd
}

func checkBadExpertRange(printer *util.PrintRecorder, current uint32, decrementBase int32, plannedIncrementTestRange int32) bool {
	printer.Printf("Current gear expertise %d\n", current)
	min := int32(current) - decrementBase
	max := int32(current) - decrementBase + plannedIncrementTestRange
	printer.Printf("Planned simulated gear expertise %d-%d\n", min, max)
	return max > int32(requirements.TARGET_RATING_TANK)
}

func fixBadHasteRange(printer *util.PrintRecorder, currentHaste uint32, plannedIncrementTestRange int32) int32 {
	printer.Printf("Current gear haste %d\n", currentHaste)
	min := int32(currentHaste)
	max := int32(currentHaste) + plannedIncrementTestRange
	printer.Printf("Planned simulated gear haste %d-%d\n", min, max)

	var fix int32
	if max > c_hasteDiscontinuityStart {
		fix = c_hasteDiscontinuityStart - max
	} else if min < c_hasteDiscontinuityEnd {
		fix = c_hasteDiscontinuityEnd - min
	}

	if fix != 0 {
		printer.Printf("Corrected simulated gear haste %d-%d\n", min+fix, max+fix)
	}
	return fix
}

func fixBadExpertRange(printer *util.PrintRecorder, currentExpert uint32, plannedIncrementTestRange int32) int32 {
	printer.Printf("Current gear expertise %d\n", currentExpert)
	min := int32(currentExpert)
	max := int32(currentExpert) + plannedIncrementTestRange
	printer.Printf("Planned simulated gear expertise %d-%d\n", min, max)

	var fix int32
	if max >= int32(requirements.TARGET_RATING_TANK) {
		fix = int32(requirements.TARGET_RATING_TANK) - max
	}

	if fix != 0 {
		printer.Printf("Corrected simulated gear expertise %d-%d\n", min+fix, max+fix)
	}
	return fix
}

func initialBonusStatMap_fixRanges(printer *util.PrintRecorder, currentItemSet items.FullItemSet, plannedIncrementTestRange int32) map[stats.StatType]int32 {
	incrementBaseHaste := fixBadHasteRange(printer, currentItemSet.Total().GetUInt(stats.Stat_Haste), plannedIncrementTestRange)
	incrementBaseExpertise := fixBadExpertRange(printer, currentItemSet.Total().Expertise(), plannedIncrementTestRange)
	initialBaseStats := make(map[stats.StatType]int32)
	initialBaseStats[stats.Stat_Haste] += incrementBaseHaste
	initialBaseStats[stats.Stat_Expertise] += incrementBaseExpertise
	return initialBaseStats
}

func initialBonusStatMap(printer *util.PrintRecorder, currentItemSet items.FullItemSet, incrementBaseHaste int32, decrementBaseExpertise int32, incrementMax int32) map[stats.StatType]int32 {
	if checkBadHasteRange(printer, currentItemSet.Total().GetUInt(stats.Stat_Haste), incrementBaseHaste, incrementMax) {
		panic("haste in discontinuity range")
	}
	if checkBadExpertRange(printer, currentItemSet.Total().Expertise(), decrementBaseExpertise, incrementMax) {
		panic("simulate will overcap expertise")
	}
	initialBaseStats := make(map[stats.StatType]int32)
	initialBaseStats[stats.Stat_Haste] += incrementBaseHaste
	initialBaseStats[stats.Stat_Expertise] -= decrementBaseExpertise
	return initialBaseStats
}

func addBonusStats(base *stats.StatBlock, bonusStat map[stats.StatType]int32) stats.StatBlock {
	resultBlock := *base
	for stat, add := range bonusStat {
		value := int64(resultBlock[stat]) + int64(add)
		if value < 0 || value > math.MaxUint32 {
			panic("out of range")
		}
		resultBlock[stat] = uint32(value)
	}
	return resultBlock
}

func statWeightsGrid_updateAll(printer *util.PrintRecorder) {
	// simSpeed := simulate.RunSize_TestOnly
	simSpeed := simulate.RunSize_QuickDirty
	// simSpeed := simulate.RunSize_Medium

	weightFileOut := files.WeightMitiNoSetFile
	gearFile := files.GearFileProtMitigationNoSet
	gearModel := model.Model_PallyProtMitigation_NoSet()
	ratios := stathighs.NewStatWeights_generalMiti
	statWeightsGrid_updateOne(gearModel, gearFile, ratios, weightFileOut, printer, simSpeed)

	weightFileOut = files.WeightMitiWithSetFile
	gearFile = files.GearFileProtMitigationWithSet
	gearModel = model.Model_PallyProtMitigation_WithSet()
	ratios = stathighs.NewStatWeights_radenWeight
	statWeightsGrid_updateOne(gearModel, gearFile, ratios, weightFileOut, printer, simSpeed)

	weightFileOut = files.WeightDpsFile
	gearFile = files.GearFileProtDps
	gearModel = model.Model_PallyProtDps()
	ratios = stathighs.NewStatWeights_dpsWeight
	statWeightsGrid_updateOne(gearModel, gearFile, ratios, weightFileOut, printer, simSpeed)

	weightFileOut = files.WeightCompromiseFile
	gearFile = files.GearFileProtCompromise
	gearModel = model.Model_PallyProtCompromise()
	ratios = stathighs.NewStatWeights_animusWeight
	statWeightsGrid_updateOne(gearModel, gearFile, ratios, weightFileOut, printer, simSpeed)
}

func statWeightsGrid_updateOne(gearModel model.Model, gearFile string, ratios simulate.SimData, weightFileOut string, printer *util.PrintRecorder, simSpeed simulate.WowSim_RunSize) {
	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(gearFile), &gearModel, printer)
	currentItemSet := items.FullItemSet_FromMap(currentEquip)

	// SIMULATE STAT CHANGES
	inputData := generateRatingsInputFromArtificalStatOverrides_ForGrid(currentItemSet, printer, simSpeed, gearModel.Spec, gearModel.Goal, gearModel.SimulateAs, gearModel.Professions)

	// SOLVE FOR STAT WEIGHTS
	process := stathighs.GridStatWeightProcess{}
	process.Init(printer)
	process.SetTargetRatios(ratios)
	process.SupplyData(inputData)
	weights := process.Run()
	pawn := writePawnString(weights, printer)

	// OVERWRITE WEIGHT FILE
	writeFile(weightFileOut, pawn)
}

func writeWeightInputsToFile(weightInputs []stathighs.WeightInput, filename string) {
	bytes, err := json.Marshal(weightInputs)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(filename, bytes, 0)
	if err != nil {
		panic(err)
	}
}

func readWeightInputFile(filename string) []stathighs.WeightInput {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var weightInputs []stathighs.WeightInput
	err = json.Unmarshal(bytes, &weightInputs)
	if err != nil {
		panic(err)
	}
	return weightInputs
}

func statWeights_CompareAlgorithms(printer *util.PrintRecorder) {
	// simSpeed := simulate.RunSize_Medium
	simSpeed := simulate.RunSize_QuickDirty
	// simSpeed := simulate.RunSize_TestOnly
	// makeSetCount := 400
	// makeSetCount := 40

	gearFile := files.GearFileProtMitigationNoSet
	gearModel := model.Model_PallyProtMitigation_NoSet()
	targetRatio := stathighs.NewStatWeights_generalMiti

	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(gearFile), &gearModel, printer)
	currentItemSet := items.FullItemSet_FromMap(currentEquip)

	inputDataBasic, basicSimBase := generateRatingsInputFromArtificalStatOverrides_ForBasic(currentItemSet, printer, simSpeed, gearModel.Spec, gearModel.Goal, gearModel.SimulateAs, gearModel.Professions)
	// inputDataGrid := generateRatingsInputFromArtificalStatOverrides_ForGrid(currentItemSet, printer, simSpeed, gearModel.Spec, gearModel.Goal, gearModel.SimulateAs, gearModel.Professions)
	// inputDataRandom := generateRatingsInputFromRealRandomSetsGeneral(gearFile, substituteItemsMiti, &gearModel, 400, simSpeed, true)
	// writeWeightInputsToFile(inputDataGrid, "sim-stats-compare-grid.json")
	// writeWeightInputsToFile(inputDataRandom, "sim-stats-compare-rand.json")
	inputDataGrid := readWeightInputFile("sim-stats-compare-grid.json")
	inputDataRandom := readWeightInputFile("sim-stats-compare-rand.json")
	mixedInputData := slices.Concat(inputDataGrid, inputDataRandom)

	resultsByAlgorithm := make(map[string]map[stats.StatType]float64)

	printer.Println("################# BASIC ###################")
	{
		basic := stathighs.BasicStatWeightProcess{}
		basic.Init(printer)
		basic.SetTargetRatios(targetRatio)
		basic.SetBaseline(basicSimBase)
		for _, data := range inputDataBasic {
			basic.AddSimData(data.IncrementStat, uint32(data.IncrementValue), data.SimResult)
		}
		resultsByAlgorithm["basic"] = basic.Run()
	}

	printer.Println("################# COMPLEX ###################")
	{
		comp := stathighs.ComplexStatWeightProcess{}
		comp.Init(printer)
		comp.SetTargetRatios(targetRatio)
		comp.SetMinimumIncludeRate(1)
		comp.SupplyData(inputDataRandom)
		resultsByAlgorithm["complex"] = comp.Run()
	}

	// printer.Println("################# FITTING ###################")
	// {
	// 	fitting := stathighs.FittingEachStatWeightProcess{}
	// 	fitting.Init(printer)
	// 	fitting.SetTargetRatios(targetRatio)
	// 	fitting.SetLazyMode(true)
	// 	fitting.SupplyDataFromStandard(inputDataRandom)
	// 	resultsByAlgorithm["fitting"] = fitting.Run()
	// }

	// printer.Println("################# GRID1 ###################")
	// {
	// 	grid1 := stathighs.GridStatWeightProcess{}
	// 	grid1.Init(printer)
	// 	grid1.SetTargetRatios(targetRatio)
	// 	grid1.SupplyData(inputDataGrid)
	// 	resultsByAlgorithm["grid1"] = grid1.Run()
	// }

	// printer.Println("################# GRID2 ###################")
	// {
	// 	grid2 := stathighs.GridStatWeightProcess2{}
	// 	grid2.Init(printer)
	// 	grid2.SetTargetRatios(targetRatio)
	// 	grid2.SupplyData(inputDataGrid)
	// 	resultsByAlgorithm["grid2"] = grid2.Run()
	// }

	printer.Println("################# RANKING0 ###################")
	{
		ranking := stathighs.RankingStatWeightProcess{}
		ranking.Init(printer)
		ranking.SetTargetRatios(targetRatio)
		// ranking.SupplyData(inputDataGrid)
		ranking.SupplyData(mixedInputData)
		ranking.RANKMODE = 0

		resultsByAlgorithm["ranking0"] = ranking.Run()
	}

	printer.Println("################# RANKING1 ###################")
	{
		ranking := stathighs.RankingStatWeightProcess{}
		ranking.Init(printer)
		ranking.SetTargetRatios(targetRatio)
		// ranking.SupplyData(inputDataGrid)
		ranking.SupplyData(mixedInputData)
		ranking.RANKMODE = 1

		resultsByAlgorithm["ranking1"] = ranking.Run()
	}
	printer.Println("################# RANKING2 ###################")
	{
		ranking := stathighs.RankingStatWeightProcess{}
		ranking.Init(printer)
		ranking.SetTargetRatios(targetRatio)
		// ranking.SupplyData(inputDataGrid)
		ranking.SupplyData(mixedInputData)
		ranking.RANKMODE = 2

		resultsByAlgorithm["ranking2"] = ranking.Run()
	}

	// printer.Println("################# SELECTIVE GRID ###################")
	// {
	// 	selgrid := stathighs.SelectiveGridStatWeightProcess{}
	// 	selgrid.Init(printer)
	// 	selgrid.SetTargetRatios(targetRatio)
	// 	selgrid.SupplyData(inputDataGrid)
	// 	resultsByAlgorithm["selgrid"] = selgrid.Run()
	// }

	printer.Println("################# FINAL RESULT ###################")
	tab := util.TabulateOutput{}
	tab.SetColumnSpacing(2)
	tab.AddColumnHeader("algo", false)
	for _, stat := range stathighs.G_RequiredStats {
		tab.AddColumnHeader(stat.Name(), true)
	}
	tab.AddColumnHeader("accuracy", false)
	for label, resultMap := range resultsByAlgorithm {
		row := make([]string, 0)
		row = append(row, label)
		for _, stat := range stathighs.G_RequiredStats {
			value := resultMap[stat]
			row = append(row, strconv.FormatFloat(value, 'f', 4, 64))
		}
		accuracy := evaluateAccuracy(resultMap, mixedInputData, targetRatio)
		row = append(row, strconv.FormatFloat(accuracy, 'f', 4, 64))
		tab.AddRow(row)
	}
	tab.Write(printer)
}

func evaluateAccuracy(statWeights map[stats.StatType]float64, inputData []stathighs.WeightInput, simRatios simulate.SimData) float64 {
	// Make ranked sim data
	rankedSims := util.MapSlice[simulate.SimType, float64]{}
	for _, sample := range inputData {
		for _, simType := range stathighs.G_RequiredSims {
			rankedSims.Add(simType, sample.SimResult.Get(simType))
		}
	}
	rankedSims.MapInternalSlicesAll(func(simType simulate.SimType, inner []float64) []float64 {
		if simType.IsHighGood() {
			// ascending, so that later indexes are better and worth more rank
			slices.Sort(inner)
		} else {
			// decending, later entries are smaller numerically, but worth more in index
			slices.SortFunc(inner, func(a, b float64) int { return cmp.Compare(b, a) })
		}
		return inner
	})

	// make structures
	type accuracyInfo struct {
		input *stathighs.WeightInput

		simRanks             map[simulate.SimType]int
		combinedSimRankScore float64

		statScore float64
		statRank  int
	}
	accuracyData := util.MapSliceAsNew(inputData, func(input *stathighs.WeightInput) accuracyInfo {
		return accuracyInfo{
			input:     input,
			statScore: calcStatScore(input, statWeights),
		}
	})

	// rank the accuracy structures by stat score
	slices.SortFunc(accuracyData, func(a, b accuracyInfo) int { return cmp.Compare(a.statScore, b.statScore) })

	// set rank values
	for statRankIndex := range accuracyData {
		info := &accuracyData[statRankIndex]
		info.statRank = statRankIndex // save the stat ranks from current slice order

		info.simRanks = make(map[simulate.SimType]int)
		for _, simType := range stathighs.G_RequiredSims {
			rankedValues := rankedSims.GetInternalSlice(simType)
			queryValue := info.input.SimResult.Get(simType)

			rank := slices.Index(rankedValues, queryValue)
			if rank == -1 {
				panic("missing value")
			}
			info.simRanks[simType] = rank

			info.combinedSimRankScore += float64(rank) * simRatios.Get(simType)
		}
	}

	// rank the accuracy structures by sim score
	slices.SortFunc(accuracyData, func(a, b accuracyInfo) int { return cmp.Compare(a.combinedSimRankScore, b.combinedSimRankScore) })

	// compute average difference between stat rank and sim rank
	totalComparePercents := 0.0
	for combinedSimRankIndex := range accuracyData {
		info := &accuracyData[combinedSimRankIndex]

		// 100% if ranks are equal, 90% if average 10% difference, etc
		diff := util.AbsIntDiff(combinedSimRankIndex, info.statRank)
		diffAsPercent := float64(diff) / float64(len(accuracyData))
		percentScore := 100.0 - (diffAsPercent * 100.0)
		totalComparePercents += percentScore
	}

	return totalComparePercents / float64(len(accuracyData))
}

func calcStatScore(input *stathighs.WeightInput, statWeights map[stats.StatType]float64) float64 {
	total := 0.0
	for statType, weightValue := range statWeights {
		total += input.TotalStat.GetFloat(statType) * weightValue
	}
	return total
}
