package main

import (
	"cmp"
	"encoding/json"
	"maps"
	"math/rand/v2"
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
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"paladin_gearing_go/weightfind"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

func simResultAddToCSV(simResult stats.SimData, csv *util.CSVOutputByColumn) {
	for _, simType := range stats.SimTypeList {
		num := simResult.Get(simType)
		if simType == stats.Sim_DEATH {
			num *= 100
		}
		csv.AddFloat64(num, 2)
	}
}

func testBasicStatsGeneral(printer *util.PrintRecorder) {
	// simSpeed := simulate.RunSize_QuickDirty
	simSpeed := simulate.RunSize_Common

	// fight := stats.Fight_Animus
	// spec := stats.Spec_PaladinProtMitigation
	// startGear := files.GearFileProtMitigationNoSet
	// modelEquipOnly := model.Model_PallyProtMitigation_NoSet()
	// goal := stats.UpgradeGoal_Mitigation

	fight := stats.Fight_Horridon_LowHeal
	spec := stats.Spec_PaladinProt
	startGear := files.GearFileProtMitigationWithSet
	modelEquipOnly := model.Model_PallyProtMitigation_WithSet()
	targetRatio := modelEquipOnly.SimRatioWeighting
	goal := stats.OptimiseGoal_Mitigation

	// fight := stats.Fight_Horridon_HighHeal
	// spec := stats.Spec_PaladinProtDps
	// startGear := files.GearFileProtDps
	// modelEquipOnly := model.Model_PallyProtDps()
	// goal := stats.UpgradeGoal_Dps

	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(startGear), &modelEquipOnly, setup.MissingEnchant_Panic, printer)
	itemSet := items.FullItemSet_FromMap(currentEquip)

	inputData, simBase := generateRatingsInputFromArtificalStatOverrides_ForBasic(itemSet, printer, simSpeed, modelEquipOnly.SimSpeedUp, modelEquipOnly.StatsForWeighting, spec, goal, fight, modelEquipOnly.Professions)

	process := stathighs.BasicStatWeightProcess{}
	process.Init(printer)
	process.SetRequiredStats(model.StatsForWeighting_strengthTank)
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
	gearMitiNoSet := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(files.GearFileProtMitigationNoSet), &modelMitiNoSet, setup.MissingEnchant_Panic, printer)
	itemSetMitiNoSet := items.FullItemSet_FromMap(gearMitiNoSet)

	modelDps := model.Model_PallyProtDps()
	gearDps := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(files.GearFileProtDps), &modelDps, setup.MissingEnchant_Panic, printer)
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

type basicStatInput struct {
	IncrementStat  stats.StatType
	IncrementValue int32
	SimResult      stats.SimData
}

func generateRatingsInputFromArtificalStatOverrides_ForBasic(currentItemSet items.FullItemSet, printer *util.PrintRecorder, simSpeed simulate.WowSim_RunSize, speedUp int, requiredStats []stats.StatType, spec stats.SpecType, goal stats.OptimiseGoal, fight stats.WowSim_Fight, profession model.ProfessionInfo) ([]basicStatInput, stats.SimData) {
	var incrementValue int32 = 250

	initialBaseStats := weightfind.InitialBonusStatMap_fixRanges(printer, currentItemSet, incrementValue)
	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(len(requiredStats) + 1)
	defer tracker.SetDone()

	simBase := simulate.WowSim_Execute_SpecifyAll(simSpeed, speedUp, spec, goal, fight, profession, currentItemSet.Items(), nil, tracker.NewChild())

	inputList := channel_op.Map_SliceToSlice(len(requiredStats), requiredStats, func(incStat *stats.StatType) basicStatInput {
		innerPrint := util.PrintRecorder_HoldAll()

		bonusStat := maps.Clone(initialBaseStats)
		str := util.StringBuild2{}
		str.WriteString("STATS SCENARIO ")
		bonusStat[*incStat] += incrementValue
		str.WriteString(incStat.Name())
		str.WriteRune('=')
		str.WriteInt32(bonusStat[*incStat])
		str.WriteRune(' ')

		simResult := simulate.WowSim_Execute_SpecifyAll(simSpeed, speedUp, spec, goal, fight, profession, currentItemSet.Items(), &bonusStat, tracker.NewChild())

		str.WriteString("   --> ")
		simResult.CompactStringGeneralBuilder(&str)
		innerPrint.PrintlnFromBuild(str)

		printer.AppendOther(innerPrint)

		return basicStatInput{
			IncrementStat:  *incStat,
			IncrementValue: incrementValue,
			SimResult:      simResult,
		}
	})
	return inputList, simBase
}

func generateRatingsInputFromRealRandomSetsT5(printer *util.PrintRecorder) ([]stathighs.WeightInput, stats.SimData) {
	makeSetCount := 2000
	simSize := simulate.RunSize_Common

	model := model.Model_PallyProtMitigation_NoSet()
	targetRatio := model.SimRatioWeighting

	_, itemOptions := allT5stuff(&model, files.GearFileProtMitigationNoSet, printer)

	// setList := build.SolverBuildRandom_MakeN_FullAndValidate(&itemOptions, &model, makeSetCount, printer, 14000)
	setList := build.SolverBuildRandom_MakeN_FullAndValidate(&itemOptions, &model, makeSetCount, printer, 0)

	track := util.TrackProgress_Start()
	track.RunOuterTracking(len(setList))
	defer track.SetDone()

	weightInputs := channel_op.Map_SliceToSlice(6, setList, func(itemSet *items.FullItemSet) stathighs.WeightInput {
		simResult := simulate.WowSim_Execute_UseModel(simSize, &model, itemSet.Items(), nil, track.NewChild())
		return stathighs.WeightInput{TotalStat: *itemSet.Total(), SimResult: simResult}
	})

	writeWeightInputsToFile(weightInputs, "sim-stats-input-data2.json")

	return weightInputs, targetRatio
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

	comp := stathighs.FormulaStatWeightProcess{}

	comp.Init(printer)
	comp.SetRequiredStats(model.StatsForWeighting_strengthTank)
	comp.SetTargetRatios(model.SimRatio_generalMiti)
	comp.SetMinimumIncludeRate(0.7)
	comp.SupplyData(filteredInput)
	weights := comp.Run()
	tools.WritePawnString(weights, printer)
}

func statWeightsRanking(printer *util.PrintRecorder) {
	// weightInputs, targetRatio := generateRatingsInputFromRealRandomSets(printer)
	targetRatio := model.SimRatio_generalMiti

	inputDataGrid := readWeightInputFile("sim-stats-compare-grid.json")
	inputDataRandom := readWeightInputFile("sim-stats-compare-rand.json")
	// mixedInputData := slices.Concat(inputDataGrid, inputDataRandom)
	mixedInputData := slices.Concat(inputDataRandom, inputDataGrid)

	filteredInput := mixedInputData
	printer.Printf("filteredInput size %d\n", len(filteredInput))

	startWeight := stathighs.WeightResult{
		stats.Stat_Strength:  1.000000,
		stats.Stat_Stamina:   0.639290,
		stats.Stat_Crit:      0.671679,
		stats.Stat_Haste:     1.186655,
		stats.Stat_Expertise: 0.858334,
		stats.Stat_Mastery:   2.530013,
		stats.Stat_Dodge:     0.671679,
		stats.Stat_Parry:     1.186655,
	}

	ranking := stathighs.RankingStatWeightProcess5{}
	ranking.Init(printer)
	ranking.SetRequiredStats(model.StatsForWeighting_strengthTank)
	ranking.SetTargetRatios(targetRatio)
	ranking.SupplyData(filteredInput)
	ranking.SupplyInitialWeights(startWeight)
	weightsList := ranking.Run()
	for _, weight := range weightsList {
		tools.WritePawnString(weight, printer)
		printer.Printf("accuracy = %f\n", weightfind.EvaluateAccuracy(weight, mixedInputData, targetRatio))
	}
}

func statWeightsGridIntoRanking(printer *util.PrintRecorder) {
	targetRatio := model.SimRatio_generalMiti

	inputDataGrid := readWeightInputFile("sim-stats-compare-grid.json")
	inputDataRandom := readWeightInputFile("sim-stats-compare-rand.json")
	mixedInputData := slices.Concat(inputDataRandom, inputDataGrid)

	var weights1 stathighs.WeightResult
	if false {
		grid := stathighs.GridStatWeightProcess{}
		grid.Init(printer)
		grid.SetRequiredStats(model.StatsForWeighting_strengthTank)
		grid.SetTargetRatios(targetRatio)
		grid.SupplyData(inputDataGrid)
		weights1 = grid.Run()
		tools.WritePawnString(weights1, printer)
	} else {
		//  Pawn: v1: "Protection WoWSims Weights": Class=Paladin,Strength=1.0000000000,Stamina=0.4804976439,CritRating=0.6462171056,HasteRating=0.8598561605,ExpertiseRating=0.6679862341,MasteryRating=1.9405533853,DodgeRating=0.6518112608,ParryRating=0.6243298125, )
		// FINAL WEIGHTS
		//        str 1.000000
		//       stam 0.480498
		//       crit 0.646220
		//      haste 0.859856
		//     expert 0.667986
		//     master 1.940554
		//      dodge 0.651811
		//      parry 0.624330

		weights1 = stathighs.WeightResult{
			stats.Stat_Strength:  1.000000,
			stats.Stat_Stamina:   0.480505,
			stats.Stat_Crit:      0.646226,
			stats.Stat_Haste:     0.859856,
			stats.Stat_Expertise: 0.667975,
			stats.Stat_Mastery:   1.940581,
			stats.Stat_Dodge:     0.651822,
			stats.Stat_Parry:     0.624330,
		}
	}

	ranking := stathighs.RankingStatWeightProcess3{}
	// ranking := stathighs.RankingStatWeightProcess4{}
	ranking.Init(printer)
	ranking.SetRequiredStats(model.StatsForWeighting_strengthTank)
	ranking.SetTargetRatios(targetRatio)
	ranking.SupplyData(mixedInputData)
	weights2 := ranking.RunUsingExternalStart(weights1).GetOrPanic()

	tools.WritePawnString(weights1, printer)
	printer.Printf("accuracy1 = %f\n", weightfind.EvaluateAccuracy(weights1, mixedInputData, targetRatio))

	tools.WritePawnString(weights2, printer)
	printer.Printf("accuracy2 = %f\n", weightfind.EvaluateAccuracy(weights2, mixedInputData, targetRatio))

	// ( Pawn: v1: "Protection WoWSims Weights": Class=Paladin,Strength=1.0000000000,Stamina=0.4805050000,CritRating=0.6462260000,HasteRating=0.8598560000,ExpertiseRating=0.6679750000,MasteryRating=1.9405810000,DodgeRating=0.6518220000,ParryRating=0.6243300000, )
	// accuracy1 = 92.635522
	// ( Pawn: v1: "Protection WoWSims Weights": Class=Paladin,Strength=1.0000000000,Stamina=-0.0896998019,CritRating=0.3760289134,HasteRating=0.4969834753,ExpertiseRating=0.3863096443,MasteryRating=1.1063898778,DodgeRating=0.3787533903,ParryRating=0.3581785849, )
	// accuracy2 = 91.501292
	// Duration = 2h2m1.7439799s

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

	// for _, entry := range weightInputs {
	// 	if hasteInDiscontinuityRange(entry.TotalStat.GetUInt(stats.Stat_Haste)) {
	// 		printer.Println("haste in discontinuity range")
	// 	}
	// }

	printer.Printf("Initial weight input size = %d\n", len(weightInputs))

	// fitting := stathighs.FittingSingleStatWeightProcess{}
	// fitting.Init(printer)
	// fitting.SetMinimumIncludeRate(0.2)
	// fitting.SupplyDataFromStandard(weightInputs[0:32], stats.Stat_Crit, simulate.Result_DPS)
	// oneWeight := fitting.Run()
	// printer.Printf("%f %f %f %f %f\n", oneWeight.LineSlope, oneWeight.LineOffset, oneWeight.Minimum, oneWeight.Maximum, oneWeight.IncludePercent)
	// tools.WritePawnString(weights, printer)

	fitting := stathighs.FittingSingleStatSegmentsProcess{}
	// fitting.Init(printer, stats.Stat_Crit, simulate.Result_DPS)
	fitting.Init(printer, stats.Stat_Haste, stats.Sim_DPS)
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
	// tools.WritePawnString(weights, printer)
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

	// for _, entry := range weightInputs {
	// 	if hasteInDiscontinuityRange(entry.TotalStat.GetUInt(stats.Stat_Haste)) {
	// 		printer.Println("haste in discontinuity range")
	// 	}
	// }

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
	process.SetRequiredStats(model.StatsForWeighting_strengthTank)
	process.SetTargetRatios(model.SimRatio_generalMiti)
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
	tools.WritePawnString(weights, printer)
}

func statWeightsGrid(printer *util.PrintRecorder) {
	// inputData, targetRatio := generateRatingsInputFromArtificalStatOverrides(printer)
	// writeWeightInputsToFile(inputData, "sim-stats-input-grid.json" )

	inputDataFull := readWeightInputFile("sim-stats-compare-grid.json")
	inputData := takeDataSample_Random(inputDataFull, 50)
	// inputData := inputDataFull

	targetRatio := model.SimRatio_generalMiti
	requiredStats := model.StatsForWeighting_strengthTank

	// process := stathighs.GridStatWeightProcess{}
	process := stathighs.SelectiveGridStatWeightProcess{}
	process.Init(printer)
	process.SetRequiredStats(requiredStats)
	process.SetTargetRatios(targetRatio)
	process.SupplyData(inputData)
	weights := process.Run()
	tools.WritePawnString(weights, printer)
	acc := weightfind.EvaluateAccuracy(weights, inputDataFull, targetRatio)
	printer.Printf("accuracy = %f\n", acc)

	weights2 := weightfind.WeightTweaker(weights, requiredStats, targetRatio, inputDataFull, printer)
	acc2 := weightfind.EvaluateAccuracy(weights2, inputDataFull, targetRatio)
	printer.Printf("accuracy_tweak = %f\n", acc2)
}

func parseSimStats(str string) stats.SimData {
	result := stats.SimData{}
	parts := strings.Split(str, " ")
	for i, simType := range stats.SimTypeList {
		value, err := strconv.ParseFloat(parts[i], 64)
		if err != nil {
			panic(err)
		}
		result.Set(simType, value)
	}
	return result
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

type basicWeightsTestDataFormat struct {
	InputDataBasic []basicStatInput
	BasicSimBase   stats.SimData
}

func writeWeightBasicInputsToFile(inputDataBasic []basicStatInput, basicSimBase stats.SimData, filename string) {
	bytes, err := json.Marshal(&basicWeightsTestDataFormat{inputDataBasic, basicSimBase})
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(filename, bytes, 0)
	if err != nil {
		panic(err)
	}
}

func readWeightBasicInputsFile(filename string) ([]basicStatInput, stats.SimData) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var weightInputs basicWeightsTestDataFormat
	err = json.Unmarshal(bytes, &weightInputs)
	if err != nil {
		panic(err)
	}
	return weightInputs.InputDataBasic, weightInputs.BasicSimBase
}

func statWeights_CompareAlgorithms(printer *util.PrintRecorder) {
	// simSpeed := simulate.RunSize_Medium
	// simSpeed := simulate.RunSize_QuickDirty
	// simSpeed := simulate.RunSize_TestOnly
	// makeSetCount := 400
	// makeSetCount := 40

	// gearFile := files.GearFileProtMitigationNoSet
	// gearModel := model.Model_PallyProtMitigation_NoSet()
	targetRatio := model.SimRatio_generalMiti
	requiredStats := model.StatsForWeighting_strengthTank

	// currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(gearFile), &gearModel, printer)
	// currentItemSet := items.FullItemSet_FromMap(currentEquip)

	// inputDataBasic, basicSimBase := generateRatingsInputFromArtificalStatOverrides_ForBasic(currentItemSet, printer, simSpeed, gearModel.Spec, gearModel.Goal, gearModel.SimulateAs, gearModel.Professions)
	// inputDataGrid := generateRatingsInputFromArtificalStatOverrides_ForGrid(currentItemSet, printer, simSpeed, gearModel.Spec, gearModel.Goal, gearModel.SimulateAs, gearModel.Professions)
	// inputDataRandom := generateRatingsInputFromRealRandomSetsGeneral(gearFile, substituteItemsMiti, &gearModel, 400, simSpeed, true)
	// writeWeightInputsToFile(inputDataGrid, "sim-stats-compare-grid.json")
	// writeWeightInputsToFile(inputDataRandom, "sim-stats-compare-rand.json")
	// writeWeightBasicInputsToFile(inputDataBasic, basicSimBase, "sim-stats-compare-basic.json")
	// inputDataBasic, basicSimBase := readWeightBasicInputsFile("sim-stats-compare-basic.json")
	inputDataGrid := readWeightInputFile("sim-stats-compare-grid.json")
	inputDataRandom := readWeightInputFile("sim-stats-compare-rand.json")
	mixedInputDataFull := slices.Concat(inputDataGrid, inputDataRandom)

	sampleSize := 80
	inputDataGrid = takeDataSample_Random(inputDataGrid, sampleSize)
	inputDataRandom = takeDataSample_Random(inputDataRandom, sampleSize)
	mixedInputData := takeDataSample_Random(mixedInputDataFull, sampleSize)
	// mixedInputData := mixedInputDataFull

	resultsByAlgorithm := make(map[string]stathighs.WeightResult)
	timesByAlgorithm := make(map[string]time.Duration)

	wg := sync.WaitGroup{}

	// wg.Go(func() {
	// 	printer.Println("################# BASIC ###################")
	// 	start := time.Now()
	// 	basic := stathighs.BasicStatWeightProcess{}
	// 	basic.Init(printer)
	// 	basic.SetRequiredStats(requiredStats)
	// 	basic.SetTargetRatios(targetRatio)
	// 	basic.SetBaseline(basicSimBase)
	// 	for _, data := range inputDataBasic {
	// 		basic.AddSimData(data.IncrementStat, uint32(data.IncrementValue), data.SimResult)
	// 	}
	// 	resultsByAlgorithm["basic"] = basic.Run()
	// 	timesByAlgorithm["basic"] = time.Since(start)
	// 	printer.Println("///////////////// BASIC /////////////////")
	// })

	// wg.Go(func() {
	// 	printer.Println("################# FORMULA ###################")
	// 	start := time.Now()
	// 	comp := stathighs.FormulaStatWeightProcess{}
	// 	comp.Init(printer)
	// 	comp.SetRequiredStats(requiredStats)
	// 	comp.SetTargetRatios(targetRatio)
	// 	comp.SetMinimumIncludeRate(1)
	// 	comp.SupplyData(slices.Clone(inputDataRandom))
	// 	resultsByAlgorithm["form"] = comp.Run()
	// 	timesByAlgorithm["form"] = time.Since(start)
	// 	printer.Println("///////////////// FORMULA /////////////////")
	// })

	// wg.Go(func() {
	// 	printer.Println("################# FITTING ###################")
	// 	start := time.Now()
	// 	fitting := stathighs.FittingEachStatWeightProcess{}
	// 	fitting.Init(printer)
	//  fitting.SetRequiredStats(requiredStats)
	// 	fitting.SetTargetRatios(targetRatio)
	// 	fitting.SetLazyMode(true)
	// 	fitting.SupplyDataFromStandard(inputDataRandom)
	// 	resultsByAlgorithm["fitting"] = fitting.Run()
	// 	timesByAlgorithm["fitting"] = time.Since(start)
	// 	printer.Println("///////////////// FITTING /////////////////")
	// })

	wg.Go(func() {
		printer.Println("################# GRID1 ###################")
		start := time.Now()
		grid1 := stathighs.GridStatWeightProcess{}
		grid1.Init(printer)
		grid1.SetRequiredStats(requiredStats)
		grid1.SetTargetRatios(targetRatio)
		grid1.SupplyData(slices.Clone(inputDataGrid))
		resultsByAlgorithm["grid1"] = grid1.Run()
		timesByAlgorithm["grid1"] = time.Since(start)
		printer.Println("///////////////// GRID1 /////////////////")
	})

	wg.Go(func() {
		printer.Println("################# GRID1B ###################")
		start := time.Now()
		grid1 := stathighs.GridStatWeightProcess1B{}
		grid1.Init(printer)
		grid1.SetRequiredStats(requiredStats)
		grid1.SetTargetRatios(targetRatio)
		grid1.SupplyData(slices.Clone(inputDataGrid))
		resultsByAlgorithm["grid1b"] = grid1.Run()
		timesByAlgorithm["grid1b"] = time.Since(start)
		printer.Println("///////////////// GRID1B /////////////////")
	})

	// wg.Go(func() {
	// 	printer.Println("################# GRID2-1 ###################")
	// 	start := time.Now()
	// 	grid2 := stathighs.GridStatWeightProcess2{}
	// 	grid2.DIFFINCLUDE = 1
	// 	grid2.Init(printer)
	// 	grid2.SetRequiredStats(requiredStats)
	// 	grid2.SetTargetRatios(targetRatio)
	// 	grid2.SupplyData(slices.Clone(inputDataGrid))
	// 	resultsByAlgorithm["grid2-1"] = grid2.Run()
	// 	timesByAlgorithm["grid2-1"] = time.Since(start)
	// 	printer.Println("///////////////// GRID2-1 /////////////////")
	// })
	// wg.Go(func() {
	// 	printer.Println("################# GRID2-2 ###################")
	// 	start := time.Now()
	// 	grid2 := stathighs.GridStatWeightProcess2{}
	// 	grid2.DIFFINCLUDE = 2
	// 	grid2.Init(printer)
	//  grid2.SetRequiredStats(requiredStats)
	// 	grid2.SetTargetRatios(targetRatio)
	// 	grid2.SupplyData(slices.Clone(inputDataGrid))
	// 	resultsByAlgorithm["grid2-2"] = grid2.Run()
	// 	timesByAlgorithm["grid2-2"] = time.Since(start)
	// 	printer.Println("///////////////// GRID2-2 /////////////////")
	// })
	// wg.Go(func() {
	// 	printer.Println("################# GRID2-12 ###################")
	// 	start := time.Now()
	// 	grid2 := stathighs.GridStatWeightProcess2{}
	// 	grid2.DIFFINCLUDE = 12
	// 	grid2.Init(printer)
	//  grid2.SetRequiredStats(requiredStats)
	// 	grid2.SetTargetRatios(targetRatio)
	// 	grid2.SupplyData(slices.Clone(inputDataGrid))
	// 	resultsByAlgorithm["grid2-12"] = grid2.Run()
	// 	timesByAlgorithm["grid2-12"] = time.Since(start)
	// 	printer.Println("///////////////// GRID2-12 /////////////////")
	// })

	// wg.Go(func() {
	// 	printer.Println("################# GRID2-1001 ###################")
	// 	start := time.Now()
	// 	grid2 := stathighs.GridStatWeightProcess2{}
	// 	grid2.DIFFINCLUDE = 1001
	// 	grid2.Init(printer)
	// 	grid2.SetRequiredStats(requiredStats)
	// 	grid2.SetTargetRatios(targetRatio)
	// 	grid2.SupplyData(slices.Clone(inputDataGrid))
	// 	resultsByAlgorithm["grid2-1001"] = grid2.Run()
	// 	timesByAlgorithm["grid2-1001"] = time.Since(start)
	// 	printer.Println("///////////////// GRID2-1001 /////////////////")
	// })

	// wg.Go(func() {
	// 	printer.Println("################# GRID2-1002 ###################")
	// 	start := time.Now()
	// 	grid2 := stathighs.GridStatWeightProcess2{}
	// 	grid2.DIFFINCLUDE = 1002
	// 	grid2.Init(printer)
	// 	grid2.SetRequiredStats(requiredStats)
	// 	grid2.SetTargetRatios(targetRatio)
	// 	grid2.SupplyData(slices.Clone(inputDataGrid))
	// 	resultsByAlgorithm["grid2-1002"] = grid2.Run()
	// 	timesByAlgorithm["grid2-1002"] = time.Since(start)
	// 	printer.Println("///////////////// GRID2-1002 /////////////////")
	// })

	// wg.Go(func() {
	// 	printer.Println("################# GRID2-1012 ###################")
	// 	start := time.Now()
	// 	grid2 := stathighs.GridStatWeightProcess2{}
	// 	grid2.DIFFINCLUDE = 1012
	// 	grid2.Init(printer)
	// 	grid2.SetRequiredStats(requiredStats)
	// 	grid2.SetTargetRatios(targetRatio)
	// 	grid2.SupplyData(slices.Clone(inputDataGrid))
	// 	resultsByAlgorithm["grid2-1012"] = grid2.Run()
	// 	timesByAlgorithm["grid2-1012"] = time.Since(start)
	// 	printer.Println("///////////////// GRID2-1012 /////////////////")
	// })

	// wg.Go(func() {
	// 	printer.Println("################# RANKING0 ###################")
	// 	start := time.Now()
	// 	ranking := stathighs.RankingStatWeightProcess{}
	// 	ranking.Init(printer)
	//  ranking.SetRequiredStats(requiredStats)
	// 	ranking.SetTargetRatios(targetRatio)
	// 	ranking.SupplyData(slices.Clone(mixedInputData))
	// 	ranking.RANKMODE = 0
	// 	resultsByAlgorithm["ranking0"] = ranking.Run()
	// 	timesByAlgorithm["ranking0"] = time.Since(start)
	// 	printer.Println("///////////////// RANKING0 /////////////////")
	// })

	// wg.Go(func() {
	// 	printer.Println("################# RANKING1 ###################")
	// 	start := time.Now()
	// 	ranking := stathighs.RankingStatWeightProcess{}
	// 	ranking.Init(printer)
	//  ranking.SetRequiredStats(requiredStats)
	// 	ranking.SetTargetRatios(targetRatio)
	// 	ranking.SupplyData(slices.Clone(mixedInputData))
	// 	ranking.RANKMODE = 1
	// 	resultsByAlgorithm["ranking1"] = ranking.Run()
	// 	timesByAlgorithm["ranking1"] = time.Since(start)
	// 	printer.Println("///////////////// RANKING1 /////////////////")
	// })

	// wg.Go(func() {
	// 	printer.Println("################# RANKING2 ###################")
	// 	start := time.Now()
	// 	ranking := stathighs.RankingStatWeightProcess{}
	// 	ranking.Init(printer)
	//  ranking.SetRequiredStats(requiredStats)
	// 	ranking.SetTargetRatios(targetRatio)
	// 	ranking.SupplyData(slices.Clone(mixedInputData))
	// 	ranking.RANKMODE = 2
	// 	resultsByAlgorithm["ranking2"] = ranking.Run()
	// 	timesByAlgorithm["ranking2"] = time.Since(start)
	// 	printer.Println("///////////////// RANKING2 /////////////////")
	// })

	// wg.Go(func() {
	// 	printer.Println("################# RANKING3 ###################")
	// 	start := time.Now()
	// 	ranking := stathighs.RankingStatWeightProcess3{}
	// 	ranking.Init(printer)
	//  ranking.SetRequiredStats(requiredStats)
	// 	ranking.SetTargetRatios(targetRatio)
	// 	ranking.SupplyData(slices.Clone(mixedInputData))
	// 	weightList := ranking.Run(false)
	// 	best := util_rank.BestCollector1[stathighs.WeightResult]{}
	// 	for i, weight := range weightList {
	// 		resultsByAlgorithm["ranking3-"+strconv.Itoa(i)] = weight
	// 		best.Offer(&weight, weightfind.EvaluateAccuracy(weight, mixedInputDataFull, targetRatio))
	// 	}
	// 	timesByAlgorithm["ranking3"] = time.Since(start)
	// 	resultsByAlgorithm["ranking3"] = best.GetBestOrPanic()
	// 	printer.Println("///////////////// RANKING3 /////////////////")
	// })

	// wg.Go(func() {
	// 	printer.Println("################# RANKING4 ###################")
	// 	start := time.Now()
	// 	ranking := stathighs.RankingStatWeightProcess4{}
	// 	ranking.Init(printer)
	//  ranking.SetRequiredStats(requiredStats)
	// 	ranking.SetTargetRatios(targetRatio)
	// 	ranking.SupplyData(slices.Clone(inputDataRandom))

	// 	best := util_rank.BestCollector1[stathighs.WeightResult]{}
	// 	weightList := ranking.Run(false)
	// 	for i, weight := range weightList {
	// 		resultsByAlgorithm["ranking4-"+strconv.Itoa(i)] = weight
	// 		best.Offer(&weight, weightfind.EvaluateAccuracy(weight, mixedInputDataFull, targetRatio))
	// 	}
	// 	timesByAlgorithm["ranking4"] = time.Since(start)
	// 	resultsByAlgorithm["ranking4"] = best.GetBestOrPanic()
	// 	printer.Println("///////////////// RANKING4 /////////////////")
	// })

	// wg.Go(func() {
	// 	printer.Println("################# RANKING5 ###################")
	// 	start := time.Now()
	// 	ranking := stathighs.RankingStatWeightProcess5{}
	// 	ranking.Init(printer)
	//  ranking.SetRequiredStats(requiredStats)
	// 	ranking.SetTargetRatios(targetRatio)
	// 	ranking.SupplyData(slices.Clone(inputDataRandom))
	// 	weightList := ranking.Run()
	// 	best := util_rank.BestCollector1[stathighs.WeightResult]{}
	// 	for i, weight := range weightList {
	// 		resultsByAlgorithm["ranking5-"+strconv.Itoa(i)] = weight
	// 		best.Offer(&weight, weightfind.EvaluateAccuracy(weight, mixedInputDataFull, targetRatio))
	// 	}
	// 	timesByAlgorithm["ranking5"] = time.Since(start)
	// 	resultsByAlgorithm["ranking5"] = best.GetBestOrPanic()
	// 	printer.Println("///////////////// RANKING5 /////////////////")
	// })

	// TODO what about ranking by each simtype, then combine. simlar to fitting?

	// wg.Go(func() {
	// 	printer.Println("################# SELECTIVE GRID ###################")
	// 	start := time.Now()
	// 	selgrid := stathighs.SelectiveGridStatWeightProcess{}
	// 	selgrid.Init(printer)
	//  selgrid.SetRequiredStats(requiredStats)
	// 	selgrid.SetTargetRatios(targetRatio)
	// 	selgrid.SupplyData(inputDataGrid)
	// 	resultsByAlgorithm["selgrid"] = selgrid.Run()
	// 	timesByAlgorithm["selgrid"] = time.Since(start)
	// 	printer.Println("///////////////// SELECTIVE GRID /////////////////")
	// })

	wg.Wait()

	printer.Println("################# FINAL RESULT ###################")
	tab := util.TabulateOutput{}
	tab.SetColumnSpacing(2)
	tab.AddColumnHeader("algo", false)
	for _, stat := range requiredStats {
		tab.AddColumnHeader(stat.Name(), true)
	}
	tab.AddColumnHeader("accuracy", false)
	tab.AddColumnHeader("accuracy_tweaked", false)
	tab.AddColumnHeader("time", false)

	resultOrder := slices.SortedFunc(maps.Keys(resultsByAlgorithm), func(a, b string) int {
		return cmp.Compare(
			weightfind.EvaluateAccuracy(resultsByAlgorithm[a], mixedInputDataFull, targetRatio),
			weightfind.EvaluateAccuracy(resultsByAlgorithm[b], mixedInputDataFull, targetRatio),
		)
	})

	for _, label := range resultOrder {
		weight := resultsByAlgorithm[label]
		row := make([]string, 0)
		row = append(row, label)
		for _, stat := range requiredStats {
			value := weight.Get(stat)
			row = append(row, strconv.FormatFloat(value, 'f', 4, 64))
		}
		accuracy := weightfind.EvaluateAccuracy(weight, mixedInputDataFull, targetRatio)
		row = append(row, strconv.FormatFloat(accuracy, 'f', 4, 64))
		row = append(row, "")
		row = append(row, timesByAlgorithm[label].String())
		tab.AddRow(row)

		weightTweak := weightfind.WeightTweaker(weight, requiredStats, targetRatio, mixedInputData, util.PrintRecorder_HoldAll())
		accuracyTweak := weightfind.EvaluateAccuracy(weightTweak, mixedInputDataFull, targetRatio)
		row = make([]string, 0)
		row = append(row, label)
		for _, stat := range requiredStats {
			value := weightTweak.Get(stat)
			row = append(row, strconv.FormatFloat(value, 'f', 4, 64))
		}
		row = append(row, "")
		row = append(row, strconv.FormatFloat(accuracyTweak, 'f', 4, 64))
		row = append(row, timesByAlgorithm[label].String())
		tab.AddRow(row)
	}
	tab.Write(printer)
}

func statWeightsGrid_updateAll(printer *util.PrintRecorder) {
	// simSpeed := simulate.RunSize_TestOnly
	simSpeed := simulate.RunSize_QuickDirty
	// simSpeed := simulate.RunSize_QuickDirty/10
	// simSpeed := simulate.RunSize_Medium

	weightfind.StatWeights_updateAll(simSpeed, printer, []weightfind.WeightOptions{
		// {
		// 	WeightFileOut:   files.WeightMitiNoSetFile,
		// 	GearFile:        files.GearFileProtMitigationNoSet,
		// 	Model:           model.Model_PallyProtMitigation_NoSet(),
		// 	SubstituteItems: substituteItemsMiti,
		// },
		// {
		// 	WeightFileOut:   files.WeightMitiWithSetFile,
		// 	GearFile:        files.GearFileProtMitigationWithSet,
		// 	Model:           model.Model_PallyProtMitigation_WithSet(),
		// 	SubstituteItems: substituteItemsMiti,
		// },
		// {
		// 	WeightFileOut:   files.WeightDpsFile,
		// 	GearFile:        files.GearFileProtDps,
		// 	Model:           model.Model_PallyProtDps(),
		// 	SubstituteItems: substituteItemsDps,
		// },
		// {
		// 	WeightFileOut:   files.WeightCompromiseFile,
		// 	GearFile:        files.GearFileProtCompromise,
		// 	Model:           model.Model_PallyProtCompromise(),
		// 	SubstituteItems: util.RemoveDuplicatesComparable(slices.Concat(substituteItemsDps, substituteItemsMiti)),
		// },
		// {
		// 	WeightFileOut:   files.WeightHealFile,
		// 	GearFile:        files.GearFileProtHeal,
		// 	Model:           model.Model_PallyProtHeal(),
		// 	SubstituteItems: util.RemoveDuplicatesComparable(slices.Concat(substituteItemsDps, substituteItemsMiti)),
		// },
		{
			WeightFileOut:   files.WeightRetFile,
			GearFile:        files.GearFileRet,
			Model:           model.Model_PallyRet(),
			SubstituteItems: substituteItemsRet,
		},
	})
}

func takeDataSample_Start(slice []stathighs.WeightInput, size int) []stathighs.WeightInput {
	if len(slice) < size {
		return slice
	} else {
		return slice[0:size]
	}
}

func takeDataSample_Random(slice []stathighs.WeightInput, size int) []stathighs.WeightInput {
	if len(slice) < size {
		return slice
	} else {
		copy := slices.Clone(slice)
		rand.Shuffle(len(copy), func(a, b int) { copy[a], copy[b] = copy[b], copy[a] })
		return copy[0:size]
	}
}
