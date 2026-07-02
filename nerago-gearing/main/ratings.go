package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"maps"
	"math/rand"
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
	"paladin_gearing_go/util/util_rank"
	"paladin_gearing_go/weightfind"
	"slices"
	"strconv"
	"strings"
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

	inputData, simBase := generateRatingsInputFromArtificialStatOverrides_ForBasic(itemSet, printer, simSpeed, modelEquipOnly.SimSpeedUp, modelEquipOnly.StatsForWeighting, spec, goal, fight, modelEquipOnly.Professions, util.TrackProgress_Start())

	process := stathighs.BasicStatWeightProcess{}
	process.Init(printer)
	process.SetRequiredStats(model.StatsForWeighting_strengthTank)
	process.SetTargetRatios(targetRatio)
	process.SetBaseline(simBase)
	for _, data := range inputData {
		process.AddSimData(data.IncrementStat, uint32(data.IncrementValue), data.SimResult)
	}
	process.Run(nil)
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

func generateRatingsInputFromArtificialStatOverrides_ForBasic(currentItemSet items.FullItemSet, printer *util.PrintRecorder, simSpeed simulate.WowSim_RunSize, speedUp int, requiredStats []stats.StatType, spec stats.SpecType, goal stats.OptimiseGoal, fight stats.WowSim_Fight, profession model.ProfessionInfo, tracker *util.TrackProgress) ([]basicStatInput, stats.SimData) {
	var incrementValue int32 = 250

	initialBaseStats := weightfind.InitialBonusStatMap_fixRanges(printer, currentItemSet, incrementValue)
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
	weights := comp.Run(nil, 3000)
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

	startWeight := stathighs.WeightResult_Make()
	startWeight.Put(stats.Stat_Strength, 1.0000)
	startWeight.Put(stats.Stat_Stamina, 1.2309)
	startWeight.Put(stats.Stat_Crit, 0.1167)
	startWeight.Put(stats.Stat_Haste, 0.3614)
	startWeight.Put(stats.Stat_Expertise, 0.0054)
	startWeight.Put(stats.Stat_Mastery, 0.5866)
	startWeight.Put(stats.Stat_Dodge, 0.0824)
	startWeight.Put(stats.Stat_Parry, 0.0532)

	ranking := stathighs.RankingStatWeightProcess5{}
	ranking.Init(printer)
	ranking.SetRequiredStats(model.StatsForWeighting_strengthTank)
	ranking.SetTargetRatios(targetRatio)
	ranking.SupplyData(filteredInput)
	ranking.SupplyInitialWeights(startWeight)
	weightsList := ranking.Run(nil, 3000)
	for _, weight := range weightsList {
		tools.WritePawnString(weight, printer)
		printer.Printf("accuracy = %f\n", weightfind.EvaluateAccuracy(weight, mixedInputData, targetRatio))
	}
}

func statWeightsGridIntoRanking(printer *util.PrintRecorder) {
	targetRatio := model.SimRatio_generalMiti
	requiredStats := model.StatsForWeighting_strengthTank

	inputDataGrid := readWeightInputFile("sim-stats-compare-grid.json")
	inputDataRandom := readWeightInputFile("sim-stats-compare-rand.json")
	mixedInputData := slices.Concat(inputDataRandom, inputDataGrid)

	var weights1 stathighs.WeightResult
	if false {
		grid := stathighs.GridStatWeightProcess{}
		grid.Init(printer, 3000)
		grid.SetRequiredStats(requiredStats)
		grid.SetTargetRatios(targetRatio)
		grid.SupplyData(inputDataGrid)
		weights1 = grid.Run(nil)
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

		// example weight value 87.7342
		//weights1 = stathighs.WeightResult{
		//	stats.Stat_Strength:  1.000000,
		//	stats.Stat_Stamina:   1.2309,
		//	stats.Stat_Crit:      0.1167,
		//	stats.Stat_Haste:     0.3614,
		//	stats.Stat_Expertise: 0.0054,
		//	stats.Stat_Mastery:   0.5866,
		//	stats.Stat_Dodge:     0.0824,
		//	stats.Stat_Parry:     0.0532,
		//}

		// better than it thinks is optimal?
		weights1 = stathighs.WeightResult_Make()
		weights1.Put(stats.Stat_Strength, 1.000000)
		weights1.Put(stats.Stat_Stamina, 1.0877848527)
		weights1.Put(stats.Stat_Crit, 2.4071360469)
		weights1.Put(stats.Stat_Haste, 0.8057261904)
		weights1.Put(stats.Stat_Expertise, 3.4897523628)
		weights1.Put(stats.Stat_Mastery, 4.9061745410)
		weights1.Put(stats.Stat_Dodge, 2.3181400067)
		weights1.Put(stats.Stat_Parry, 3.2064183845)
	}

	//mixedInputData = takeDataSample_Random_Seed(mixedInputData, 20, 1234)

	ranking := stathighs.RankingStatWeightProcess3b{}
	// 0 is all combinations, 1 is just adjacent
	ranking.ALGO = 0
	ranking.SCALE1 = true
	// ranking := stathighs.RankingStatWeightProcess4{}
	ranking.Init(printer, 300)
	ranking.SetRequiredStats(requiredStats)
	ranking.SetTargetRatios(targetRatio)
	//ranking.SupplyData(mixedInputData)
	ranking.SupplyData(mixedInputData)
	weights2 := ranking.RunSinglePassFromExternal(weights1, nil)
	//weights2 := ranking.Run(&weights1)
	//weights2 := ranking.Run(nil)

	tools.WritePawnString(weights1, printer)
	printer.Printf("accuracy_initial = %f\n", weightfind.EvaluateAccuracy(weights1, mixedInputData, targetRatio))

	tools.WritePawnString(weights2, printer)
	printer.Printf("accuracy_algo = %f\n", weightfind.EvaluateAccuracy(weights2, mixedInputData, targetRatio))

	weights3, _ := weightfind.WeightTweaker(weights2, requiredStats, targetRatio, mixedInputData, util.PrintRecorder_HoldAll())

	tools.WritePawnString(weights3, printer)
	printer.Printf("accuracy_tweak = %f\n", weightfind.EvaluateAccuracy(weights3, mixedInputData, targetRatio))

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
	fitting.Init(printer, stats.Stat_Haste, stats.Sim_DPS, 3000)
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
	fitting.Init(printer, 3000)
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
	weights := process.Run(nil)
	tools.WritePawnString(weights, printer)
}

func statWeightsGrid(printer *util.PrintRecorder) {
	// inputData, targetRatio := generateRatingsInputFromArtificalStatOverrides(printer)
	// writeWeightInputsToFile(inputData, "sim-stats-input-grid.json" )

	inputDataFull := readWeightInputFile("sim-stats-compare-grid.json")
	//inputData := takeDataSample_Random(inputDataFull, 50)
	inputData := inputDataFull

	targetRatio := model.SimRatio_generalMiti
	requiredStats := model.StatsForWeighting_strengthTank

	for ROUNDMODE := range 3 {
		for RESCALE := range 3 {
			process := stathighs.GridStatWeightProcess1C{}
			process.ROUNDMODE = ROUNDMODE
			process.RESCALE = RESCALE
			process.Init(printer, 3000)
			process.SetRequiredStats(requiredStats)
			process.SetTargetRatios(targetRatio)
			process.SupplyData(inputData)
			weights := process.Run(nil)
			tools.WritePawnString(weights, printer)
			acc := weightfind.EvaluateAccuracy(weights, inputDataFull, targetRatio)

			printer.Printf("accuracy = %f\n", acc)
			printer.Printf("############## DONE %d %d\n", ROUNDMODE, RESCALE)

			//weights2, _ := weightfind.WeightTweaker(weights, requiredStats, targetRatio, inputDataFull, printer)
			//acc2 := weightfind.EvaluateAccuracy(weights2, inputDataFull, targetRatio)
			//printer.Printf("accuracy_tweak = %f\n", acc2)
		}
	}
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
	err = os.WriteFile(filename, bytes, 0666)
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
	err = os.WriteFile(filename, bytes, 0666)
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

//goland:noinspection GoBoolExpressions
func statWeights_CompareAlgorithms(printer *util.PrintRecorder) {
	targetRatio := model.SimRatio_generalMiti
	requiredStats := model.StatsForWeighting_strengthTank

	//simSpeed := simulate.RunSize_Common
	//gearFile := files.GearFileProtMitigationNoSet
	//gearModel := model.Model_PallyProtMitigation_NoSet()
	//currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(gearFile), &gearModel, setup.MissingEnchant_Fix, printer)
	//currentItemSet := items.FullItemSet_FromMap(currentEquip)

	//trackProcess := util.TrackProgress_Start()
	//trackProcess.RunOuterTracking(4)
	//inputDataBasic, basicSimBase := generateRatingsInputFromArtificialStatOverrides_ForBasic(currentItemSet, printer, simSpeed, 1, requiredStats, gearModel.Spec, gearModel.Goal, gearModel.SimulateAs, gearModel.Professions, trackProcess.NewChild())
	//inputDataGrid := weightfind.SimulateSteppedStatChangesForGrid(currentItemSet, printer, simSpeed, 1, requiredStats, gearModel.Spec, gearModel.Goal, gearModel.SimulateAs, gearModel.Professions, trackProcess.NewChild())
	//inputDataRandomUnsafe := weightfind.SimulateRealRandomSets(gearFile, substituteItemsMiti, &gearModel, 256, simSpeed, false, printer, trackProcess.NewChild())
	//inputDataRandomSafe := weightfind.SimulateRealRandomSets(gearFile, substituteItemsMiti, &gearModel, 256, simSpeed, true, printer, trackProcess.NewChild())
	//inputDataRandom := slices.Concat(inputDataRandomUnsafe, inputDataRandomSafe)
	//trackProcess.SetDone()

	//writeWeightInputsToFile(inputDataGrid, "sim-stats-compare-grid.json")
	//writeWeightInputsToFile(inputDataRandomUnsafe, "sim-stats-compare-rand-unsafe.json")
	//writeWeightInputsToFile(inputDataRandomSafe, "sim-stats-compare-rand-safe.json")
	//writeWeightInputsToFile(inputDataRandom, "sim-stats-compare-rand.json")
	//writeWeightBasicInputsToFile(inputDataBasic, basicSimBase, "sim-stats-compare-basic.json")

	inputDataBasic, basicSimBase := readWeightBasicInputsFile("sim-stats-compare-basic.json")
	inputDataGrid := readWeightInputFile("sim-stats-compare-grid.json")
	inputDataRandom := readWeightInputFile("sim-stats-compare-rand.json")
	mixedInputDataFull := slices.Concat(inputDataGrid, inputDataRandom)

	//sampleSize := 50
	//inputDataGrid = takeDataSample_Random(inputDataGrid, sampleSize*2)
	//inputDataRandom = takeDataSample_Random(inputDataRandom, sampleSize)
	//mixedInputData := takeDataSample_Random(mixedInputDataFull, sampleSize)
	//inputDataGrid = takeDataSample_Random_Seed(inputDataGrid, sampleSize, 1234)
	//inputDataRandom = takeDataSample_Random_Seed(inputDataRandom, sampleSize, 1234)
	//mixedInputData := takeDataSample_Random_Seed(mixedInputDataFull, sampleSize, 1234)
	mixedInputData := mixedInputDataFull

	//weightsNearOptimal := stathighs.WeightResult_Make()
	//weightsNearOptimal.Put(stats.Stat_Strength, 1.000000)
	//weightsNearOptimal.Put(stats.Stat_Stamina, 1.0877848527)
	//weightsNearOptimal.Put(stats.Stat_Crit, 2.4071360469)
	//weightsNearOptimal.Put(stats.Stat_Haste, 0.8057261904)
	//weightsNearOptimal.Put(stats.Stat_Expertise, 3.4897523628)
	//weightsNearOptimal.Put(stats.Stat_Mastery, 4.9061745410)
	//weightsNearOptimal.Put(stats.Stat_Dodge, 2.3181400067)
	//weightsNearOptimal.Put(stats.Stat_Parry, 3.2064183845)

	// weight value 87.7342
	weightsMidRange := stathighs.WeightResult_Make()
	weightsMidRange.Put(stats.Stat_Strength, 1.0000)
	weightsMidRange.Put(stats.Stat_Stamina, 1.2309)
	weightsMidRange.Put(stats.Stat_Crit, 0.1167)
	weightsMidRange.Put(stats.Stat_Haste, 0.3614)
	weightsMidRange.Put(stats.Stat_Expertise, 0.0054)
	weightsMidRange.Put(stats.Stat_Mastery, 0.5866)
	weightsMidRange.Put(stats.Stat_Dodge, 0.0824)
	weightsMidRange.Put(stats.Stat_Parry, 0.0532)

	resultsByAlgorithm := make(map[string]stathighs.WeightResult)
	timesByAlgorithm := make(map[string]time.Duration)

	tasks := make([]func(), 0)

	reportOnTweakedVersions := false
	standardTimeout := 6000
	shortTimeout := 1000

	runBasic := true
	runFitting := true

	runGrid1Original := true
	runGrid1Variants := true
	runGrid1VariantsFewer := true
	runGrid1C := true
	runGrid2 := true
	runSelGrid := true

	runRankingOlder := true
	runRanking3aPreferred := true
	runRanking3aVariants := true
	runRanking3bVariants := true
	runRanking3bPreferred := true
	runRanking4 := true
	runRanking5 := true

	if runBasic {
		tasks = append(tasks, func() {
			printer.Println("################# BASIC ###################")
			stopwatch := util.StopwatchMakeStopped()
			basic := stathighs.BasicStatWeightProcess{}
			basic.Init(printer)
			basic.SetRequiredStats(requiredStats)
			basic.SetTargetRatios(targetRatio)
			basic.SetBaseline(basicSimBase)
			for _, data := range inputDataBasic {
				basic.AddSimData(data.IncrementStat, uint32(data.IncrementValue), data.SimResult)
			}
			resultsByAlgorithm["basic"] = basic.Run(stopwatch)
			timesByAlgorithm["basic"] = stopwatch.Elapsed()
			printer.Println("///////////////// BASIC /////////////////")
		})

		tasks = append(tasks, func() {
			printer.Println("################# FORMULA ###################")
			stopwatch := util.StopwatchMakeStopped()
			comp := stathighs.FormulaStatWeightProcess{}
			comp.Init(printer)
			comp.SetRequiredStats(requiredStats)
			comp.SetTargetRatios(targetRatio)
			comp.SetMinimumIncludeRate(1)
			comp.SupplyData(slices.Clone(inputDataRandom))
			resultsByAlgorithm["form"] = comp.Run(stopwatch, standardTimeout)
			timesByAlgorithm["form"] = stopwatch.Elapsed()
			printer.Println("///////////////// FORMULA /////////////////")
		})
	}

	if runFitting {
		tasks = append(tasks, func() {
			printer.Println("################# FITTING ###################")
			stopwatch := util.StopwatchMakeStopped()
			fitting := stathighs.FittingEachStatWeightProcess{}
			fitting.Init(printer, shortTimeout)
			fitting.SetRequiredStats(requiredStats)
			fitting.SetTargetRatios(targetRatio)
			fitting.SetLazyMode(true)
			fitting.SupplyDataFromStandard(inputDataRandom)
			resultsByAlgorithm["fitting"] = fitting.Run(stopwatch)
			timesByAlgorithm["fitting"] = stopwatch.Elapsed()
			printer.Println("///////////////// FITTING /////////////////")
		})
	}

	if runGrid1Original {
		tasks = append(tasks, func() {
			printer.Println("################# GRID1 ###################")
			stopwatch := util.StopwatchMakeStopped()
			grid1 := stathighs.GridStatWeightProcess{}
			grid1.Init(printer, standardTimeout)
			grid1.SetRequiredStats(requiredStats)
			grid1.SetTargetRatios(targetRatio)
			grid1.SupplyData(slices.Clone(inputDataGrid))
			resultsByAlgorithm["grid1"] = grid1.Run(stopwatch)
			timesByAlgorithm["grid1"] = stopwatch.Elapsed()
			printer.Println("///////////////// GRID1 /////////////////")
		})
	}

	if runGrid1Variants {
		tasks = append(tasks, func() {
			printer.Println("################# GRID1 ###################")
			stopwatch := util.StopwatchMakeStopped()
			grid1 := stathighs.GridStatWeightProcess{}
			grid1.CHECKRANGE = 1
			grid1.Init(printer, standardTimeout)
			grid1.SetRequiredStats(requiredStats)
			grid1.SetTargetRatios(targetRatio)
			grid1.SupplyData(slices.Clone(inputDataGrid))
			resultsByAlgorithm["grid1-1"] = grid1.Run(stopwatch)
			timesByAlgorithm["grid1-1"] = stopwatch.Elapsed()
			printer.Println("///////////////// GRID1 /////////////////")
		})

		for SCALEMODE := range 6 {
			for ROUNDMODE := range 3 {
				for OUTLIER := range 5 {
					tasks = append(tasks, func() {
						printer.Println("################# GRID1B ###################")
						stopwatch := util.StopwatchMakeStopped()
						grid1 := stathighs.GridStatWeightProcess1B{}
						grid1.SCALEMODE = SCALEMODE
						grid1.ROUNDMODE = ROUNDMODE
						grid1.OUTLIER = OUTLIER
						grid1.Init(printer, shortTimeout)
						grid1.SetRequiredStats(requiredStats)
						grid1.SetTargetRatios(targetRatio)
						grid1.SupplyData(slices.Clone(inputDataGrid))
						label := fmt.Sprintf("grid1b-outlier%d-scale%d-round%d", OUTLIER, SCALEMODE, ROUNDMODE)
						resultsByAlgorithm[label] = grid1.Run(stopwatch)
						timesByAlgorithm[label] = stopwatch.Elapsed()
						printer.Println("///////////////// GRID1B /////////////////")
					})
				}
			}
		}
	}
	if runGrid1VariantsFewer {
		for SCALEMODE := range 6 {
			printer.Println("################# GRID1B ###################")
			stopwatch := util.StopwatchMakeStopped()
			grid1 := stathighs.GridStatWeightProcess1B{}
			grid1.SCALEMODE = SCALEMODE
			grid1.ROUNDMODE = 2
			grid1.OUTLIER = 3
			grid1.Init(printer, shortTimeout)
			grid1.SetRequiredStats(requiredStats)
			grid1.SetTargetRatios(targetRatio)
			grid1.SupplyData(slices.Clone(inputDataGrid))
			label := fmt.Sprintf("grid1b-outlier%d-scale%d-round%d", 3, SCALEMODE, 2)
			resultsByAlgorithm[label] = grid1.Run(stopwatch)
			timesByAlgorithm[label] = stopwatch.Elapsed()
			printer.Println("///////////////// GRID1B /////////////////")
		}
	}

	if runGrid1C {
		for RESCALE := range 3 {
			for ROUNDMODE := range 3 {
				tasks = append(tasks, func() {
					printer.Println("################# GRID1C ###################")
					stopwatch := util.StopwatchMakeStopped()
					grid1 := stathighs.GridStatWeightProcess1C{}
					grid1.ROUNDMODE = ROUNDMODE
					grid1.RESCALE = RESCALE
					grid1.Init(printer, shortTimeout)
					grid1.SetRequiredStats(requiredStats)
					grid1.SetTargetRatios(targetRatio)
					grid1.SupplyData(slices.Clone(inputDataGrid))
					label := fmt.Sprintf("grid1c-round%d-rescale%d", ROUNDMODE, RESCALE)
					resultsByAlgorithm[label] = grid1.Run(stopwatch)
					timesByAlgorithm[label] = stopwatch.Elapsed()
					printer.Println("///////////////// GRID1C /////////////////")
				})
			}
		}
	}

	if runGrid2 {
		tasks = append(tasks, func() {
			printer.Println("################# GRID2-1 ###################")
			stopwatch := util.StopwatchMakeStopped()
			grid2 := stathighs.GridStatWeightProcess2{}
			grid2.DIFFINCLUDE = 1
			grid2.Init(printer, standardTimeout)
			grid2.SetRequiredStats(requiredStats)
			grid2.SetTargetRatios(targetRatio)
			grid2.SupplyData(slices.Clone(inputDataGrid))
			resultsByAlgorithm["grid2-1"] = grid2.Run(stopwatch)
			timesByAlgorithm["grid2-1"] = stopwatch.Elapsed()
			printer.Println("///////////////// GRID2-1 /////////////////")
		})
		tasks = append(tasks, func() {
			printer.Println("################# GRID2-2 ###################")
			stopwatch := util.StopwatchMakeStopped()
			grid2 := stathighs.GridStatWeightProcess2{}
			grid2.DIFFINCLUDE = 2
			grid2.Init(printer, standardTimeout)
			grid2.SetRequiredStats(requiredStats)
			grid2.SetTargetRatios(targetRatio)
			grid2.SupplyData(slices.Clone(inputDataGrid))
			resultsByAlgorithm["grid2-2"] = grid2.Run(stopwatch)
			timesByAlgorithm["grid2-2"] = stopwatch.Elapsed()
			printer.Println("///////////////// GRID2-2 /////////////////")
		})
		tasks = append(tasks, func() {
			printer.Println("################# GRID2-12 ###################")
			stopwatch := util.StopwatchMakeStopped()
			grid2 := stathighs.GridStatWeightProcess2{}
			grid2.DIFFINCLUDE = 12
			grid2.Init(printer, standardTimeout)
			grid2.SetRequiredStats(requiredStats)
			grid2.SetTargetRatios(targetRatio)
			grid2.SupplyData(slices.Clone(inputDataGrid))
			resultsByAlgorithm["grid2-12"] = grid2.Run(stopwatch)
			timesByAlgorithm["grid2-12"] = stopwatch.Elapsed()
			printer.Println("///////////////// GRID2-12 /////////////////")
		})

		tasks = append(tasks, func() {
			printer.Println("################# GRID2-1001 ###################")
			stopwatch := util.StopwatchMakeStopped()
			grid2 := stathighs.GridStatWeightProcess2{}
			grid2.DIFFINCLUDE = 1001
			grid2.Init(printer, standardTimeout)
			grid2.SetRequiredStats(requiredStats)
			grid2.SetTargetRatios(targetRatio)
			grid2.SupplyData(slices.Clone(inputDataGrid))
			resultsByAlgorithm["grid2-1001"] = grid2.Run(stopwatch)
			timesByAlgorithm["grid2-1001"] = stopwatch.Elapsed()
			printer.Println("///////////////// GRID2-1001 /////////////////")
		})

		tasks = append(tasks, func() {
			printer.Println("################# GRID2-1002 ###################")
			stopwatch := util.StopwatchMakeStopped()
			grid2 := stathighs.GridStatWeightProcess2{}
			grid2.DIFFINCLUDE = 1002
			grid2.Init(printer, standardTimeout)
			grid2.SetRequiredStats(requiredStats)
			grid2.SetTargetRatios(targetRatio)
			grid2.SupplyData(slices.Clone(inputDataGrid))
			resultsByAlgorithm["grid2-1002"] = grid2.Run(stopwatch)
			timesByAlgorithm["grid2-1002"] = stopwatch.Elapsed()
			printer.Println("///////////////// GRID2-1002 /////////////////")
		})

		tasks = append(tasks, func() {
			printer.Println("################# GRID2-1012 ###################")
			stopwatch := util.StopwatchMakeStopped()
			grid2 := stathighs.GridStatWeightProcess2{}
			grid2.DIFFINCLUDE = 1012
			grid2.Init(printer, standardTimeout)
			grid2.SetRequiredStats(requiredStats)
			grid2.SetTargetRatios(targetRatio)
			grid2.SupplyData(slices.Clone(inputDataGrid))
			resultsByAlgorithm["grid2-1012"] = grid2.Run(stopwatch)
			timesByAlgorithm["grid2-1012"] = stopwatch.Elapsed()
			printer.Println("///////////////// GRID2-1012 /////////////////")
		})
	}

	if runRankingOlder {
		tasks = append(tasks, func() {
			printer.Println("################# RANKING0 ###################")
			stopwatch := util.StopwatchMakeStopped()
			ranking := stathighs.RankingStatWeightProcess{}
			ranking.Init(printer)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(mixedInputData))
			ranking.RANKMODE = 0
			resultsByAlgorithm["ranking0"] = ranking.Run(stopwatch, standardTimeout)
			timesByAlgorithm["ranking0"] = stopwatch.Elapsed()
			printer.Println("///////////////// RANKING0 /////////////////")
		})

		tasks = append(tasks, func() {
			printer.Println("################# RANKING1 ###################")
			stopwatch := util.StopwatchMakeStopped()
			ranking := stathighs.RankingStatWeightProcess{}
			ranking.Init(printer)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(mixedInputData))
			ranking.RANKMODE = 1
			resultsByAlgorithm["ranking1"] = ranking.Run(stopwatch, standardTimeout)
			timesByAlgorithm["ranking1"] = stopwatch.Elapsed()
			printer.Println("///////////////// RANKING1 /////////////////")
		})

		tasks = append(tasks, func() {
			printer.Println("################# RANKING2 ###################")
			stopwatch := util.StopwatchMakeStopped()
			ranking := stathighs.RankingStatWeightProcess{}
			ranking.Init(printer)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(mixedInputData))
			ranking.RANKMODE = 2
			resultsByAlgorithm["ranking2"] = ranking.Run(stopwatch, standardTimeout)
			timesByAlgorithm["ranking2"] = stopwatch.Elapsed()
			printer.Println("///////////////// RANKING2 /////////////////")
		})
	}

	if runRanking3aVariants {
		for ALGO := range 2 {
			tasks = append(tasks, func() {
				printer.Println("################# RANKING3a ###################")
				stopwatch := util.StopwatchMakeStopped()
				ranking := stathighs.RankingStatWeightProcess3{}
				ranking.ALGO = ALGO
				ranking.SCALE1 = false
				ranking.Init(printer, shortTimeout)
				ranking.SetRequiredStats(requiredStats)
				ranking.SetTargetRatios(targetRatio)
				ranking.SupplyData(slices.Clone(mixedInputData))
				weight := ranking.Run(stopwatch)
				label := fmt.Sprintf("ranking3a-scale_stat-algo%d", ALGO)
				timesByAlgorithm[label] = stopwatch.Elapsed()
				resultsByAlgorithm[label] = weight
				printer.Println("///////////////// RANKING3a /////////////////")
			})
			tasks = append(tasks, func() {
				printer.Println("################# RANKING3a ###################")
				stopwatch := util.StopwatchMakeStopped()
				ranking := stathighs.RankingStatWeightProcess3{}
				ranking.ALGO = ALGO
				ranking.SCALE1 = true
				ranking.Init(printer, shortTimeout)
				ranking.SetRequiredStats(requiredStats)
				ranking.SetTargetRatios(targetRatio)
				ranking.SupplyData(slices.Clone(mixedInputData))
				weight := ranking.Run(stopwatch)
				label := fmt.Sprintf("ranking3a-scale1-algo%d", ALGO)
				timesByAlgorithm[label] = stopwatch.Elapsed()
				resultsByAlgorithm[label] = weight
				printer.Println("///////////////// RANKING3a /////////////////")
			})
		}
	}

	if runRanking3aPreferred {
		tasks = append(tasks, func() {
			printer.Println("################# RANKING3a-false-1 ###################")
			stopwatch := util.StopwatchMakeStopped()
			ranking := stathighs.RankingStatWeightProcess3{}
			ranking.ALGO = 1
			ranking.SCALE1 = false
			ranking.Init(printer, standardTimeout)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(mixedInputData))
			weight := ranking.RunUsingExternalStart(weightsMidRange, stopwatch)
			label := fmt.Sprintf("ranking3a-false-1")
			timesByAlgorithm[label] = stopwatch.Elapsed()
			resultsByAlgorithm[label] = weight
			printer.Println("///////////////// RANKING3a-false-1 /////////////////")
		})
		tasks = append(tasks, func() {
			printer.Println("################# RANKING3a-true-1 ###################")
			stopwatch := util.StopwatchMakeStopped()
			ranking := stathighs.RankingStatWeightProcess3{}
			ranking.ALGO = 1
			ranking.SCALE1 = true
			ranking.Init(printer, standardTimeout)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(mixedInputData))
			weight := ranking.RunUsingExternalStart(weightsMidRange, stopwatch)
			label := fmt.Sprintf("ranking3a-true-1")
			timesByAlgorithm[label] = stopwatch.Elapsed()
			resultsByAlgorithm[label] = weight
			printer.Println("///////////////// RANKING3a-true-1 /////////////////")
		})
	}

	if runRanking3bVariants {
		for FINAL := range 3 {
			tasks = append(tasks, func() {
				printer.Println("################# RANKING3b ###################")
				stopwatch := util.StopwatchMakeStopped()
				ranking := stathighs.RankingStatWeightProcess3b{}
				ranking.SCALE1 = false
				ranking.FINAL = FINAL
				ranking.Init(printer, shortTimeout)
				ranking.SetRequiredStats(requiredStats)
				ranking.SetTargetRatios(targetRatio)
				ranking.SupplyData(slices.Clone(mixedInputData))
				weight := ranking.RunSinglePassFromExternal(weightsMidRange, stopwatch)
				label := fmt.Sprintf("ranking3b-scale_full-%d", FINAL)
				timesByAlgorithm[label] = stopwatch.Elapsed()
				resultsByAlgorithm[label] = weight
				printer.Println("///////////////// RANKING3b /////////////////")
			})
			tasks = append(tasks, func() {
				printer.Println("################# RANKING3b ###################")
				stopwatch := util.StopwatchMakeStopped()
				ranking := stathighs.RankingStatWeightProcess3b{}
				ranking.SCALE1 = true
				ranking.FINAL = FINAL
				ranking.Init(printer, shortTimeout)
				ranking.SetRequiredStats(requiredStats)
				ranking.SetTargetRatios(targetRatio)
				ranking.SupplyData(slices.Clone(mixedInputData))
				weight := ranking.RunSinglePassFromExternal(weightsMidRange, stopwatch)
				label := fmt.Sprintf("ranking3b-scale1-%d", FINAL)
				timesByAlgorithm[label] = stopwatch.Elapsed()
				resultsByAlgorithm[label] = weight
				printer.Println("///////////////// RANKING3b /////////////////")
			})
		}
	}
	if runRanking3bPreferred {
		tasks = append(tasks, func() {
			printer.Println("################# RANKING3b ###################")
			stopwatch := util.StopwatchMakeStopped()
			ranking := stathighs.RankingStatWeightProcess3b{}
			ranking.SCALE1 = false
			ranking.FINAL = 0
			ranking.Init(printer, shortTimeout)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(mixedInputData))
			weight := ranking.RunSinglePassFromExternal(weightsMidRange, stopwatch)
			label := fmt.Sprintf("ranking3b-scale_full-%d", 0)
			timesByAlgorithm[label] = stopwatch.Elapsed()
			resultsByAlgorithm[label] = weight
			printer.Println("///////////////// RANKING3b /////////////////")
		})
	}

	if runRanking4 {
		tasks = append(tasks, func() {
			printer.Println("################# RANKING4 ###################")
			stopwatch := util.StopwatchMakeStopped()
			ranking := stathighs.RankingStatWeightProcess4{}
			ranking.Init(printer)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(inputDataRandom))

			//best := util_rank.BestCollector1[stathighs.WeightResult]{}
			//weightList := ranking.RunUsingExternalStart(weightsMidRange, stopwatch)
			//for i, weight := range weightList {
			//	resultsByAlgorithm["ranking4-"+strconv.Itoa(i)] = weight
			//	best.Offer(&weight, weightfind.EvaluateAccuracy(weight, mixedInputDataFull, targetRatio))
			//}
			weight := ranking.RunUsingExternalStart(weightsMidRange, stopwatch, standardTimeout)
			timesByAlgorithm["ranking4"] = stopwatch.Elapsed()
			resultsByAlgorithm["ranking4"] = weight.GetOrPanic()
			//resultsByAlgorithm["ranking4"] = best.GetBestOrPanic()
			printer.Println("///////////////// RANKING4 /////////////////")
		})
	}

	if runRanking5 {
		tasks = append(tasks, func() {
			printer.Println("################# RANKING5 ###################")
			stopwatch := util.StopwatchMakeStopped()
			ranking := stathighs.RankingStatWeightProcess5{}
			ranking.Init(printer)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(inputDataRandom))
			weightList := ranking.Run(stopwatch, standardTimeout)
			best := util_rank.BestCollector1[stathighs.WeightResult]{}
			for i, weight := range weightList {
				resultsByAlgorithm["ranking5-"+strconv.Itoa(i)] = weight
				best.Offer(&weight, weightfind.EvaluateAccuracy(weight, mixedInputDataFull, targetRatio))
			}
			timesByAlgorithm["ranking5"] = stopwatch.Elapsed()
			resultsByAlgorithm["ranking5"] = best.GetBestOrPanic()
			printer.Println("///////////////// RANKING5 /////////////////")
		})
	}

	if runSelGrid {
		tasks = append(tasks, func() {
			printer.Println("################# SELECTIVE GRID ###################")
			stopwatch := util.StopwatchMakeStopped()
			selgrid := stathighs.SelectiveGridStatWeightProcess{}
			selgrid.Init(printer, standardTimeout)
			selgrid.SetRequiredStats(requiredStats)
			selgrid.SetTargetRatios(targetRatio)
			selgrid.SupplyData(inputDataGrid)
			resultsByAlgorithm["selgrid"] = selgrid.Run(stopwatch)
			timesByAlgorithm["selgrid"] = stopwatch.Elapsed()
			printer.Println("///////////////// SELECTIVE GRID /////////////////")
		})
	}

	channel_op.ForEach_Slice(10, tasks, func(f *func()) {
		(*f)()
	})

	printer.Println("################# FINAL RESULT ###################")
	tab := util.TabulateOutput{}
	tab.SetColumnSpacing(2)
	tab.AddColumnHeader("algo", false)
	for _, stat := range requiredStats {
		tab.AddColumnHeader(stat.Name(), true)
	}
	tab.AddColumnHeader("accuracy", false)
	tab.AddColumnHeader("accuracy_tweaked", false)
	tab.AddColumnHeader("accuracy_old", false)
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
		accuracyOld := weightfind.EvaluateAccuracyOriginal(weight, mixedInputDataFull, targetRatio)
		row = append(row, strconv.FormatFloat(accuracy, 'f', 4, 64))
		row = append(row, "")
		row = append(row, strconv.FormatFloat(accuracyOld, 'f', 4, 64))
		row = append(row, timesByAlgorithm[label].String())
		tab.AddRow(row)

		if reportOnTweakedVersions {
			weightTweak, _ := weightfind.WeightTweaker(weight, requiredStats, targetRatio, mixedInputDataFull, util.PrintRecorder_HoldAll())
			accuracyTweak := weightfind.EvaluateAccuracy(weightTweak, mixedInputDataFull, targetRatio)
			accuracyOld = weightfind.EvaluateAccuracyOriginal(weightTweak, mixedInputDataFull, targetRatio)
			row = make([]string, 0)
			row = append(row, label)
			for _, stat := range requiredStats {
				value := weightTweak.Get(stat)
				row = append(row, strconv.FormatFloat(value, 'f', 4, 64))
			}
			row = append(row, "")
			row = append(row, strconv.FormatFloat(accuracyTweak, 'f', 4, 64))
			row = append(row, strconv.FormatFloat(accuracyOld, 'f', 4, 64))
			row = append(row, timesByAlgorithm[label].String())
			tab.AddRow(row)
		}
	}
	tab.Write(printer)
}

func statWeightsGrid_updateAll(printer *util.PrintRecorder) {
	//simSpeed := simulate.RunSize_TestOnly
	//simSpeed := simulate.RunSize_QuickDirty
	// simSpeed := simulate.RunSize_QuickDirty/10
	simSpeed := simulate.RunSize_Common

	weightfind.StatWeights_updateAll(simSpeed, printer, []weightfind.WeightOptions{
		{
			Label:           "Prot-Mitigation-NoSet",
			WeightFileOut:   files.WeightMitiNoSetFile,
			GearFile:        files.GearFileProtMitigationNoSet,
			Model:           model.Model_PallyProtMitigation_NoSet(),
			SubstituteItems: substituteItemsMiti,
		},
		{
			Label:           "Prot-Mitigation-WithSet",
			WeightFileOut:   files.WeightMitiWithSetFile,
			GearFile:        files.GearFileProtMitigationWithSet,
			Model:           model.Model_PallyProtMitigation_WithSet(),
			SubstituteItems: substituteItemsMiti,
		},
		{
			Label:           "Prot-Damage",
			WeightFileOut:   files.WeightDpsFile,
			GearFile:        files.GearFileProtDps,
			Model:           model.Model_PallyProtDps(),
			SubstituteItems: substituteItemsDps,
		},
		{
			Label:           "Prot-Compromise",
			WeightFileOut:   files.WeightCompromiseFile,
			GearFile:        files.GearFileProtCompromise,
			Model:           model.Model_PallyProtCompromise(),
			SubstituteItems: util.RemoveDuplicatesComparable(slices.Concat(substituteItemsDps, substituteItemsMiti)),
		},
		{
			Label:           "Prot-Heal",
			WeightFileOut:   files.WeightHealFile,
			GearFile:        files.GearFileProtHeal,
			Model:           model.Model_PallyProtHeal(),
			SubstituteItems: util.RemoveDuplicatesComparable(slices.Concat(substituteItemsDps, substituteItemsMiti)),
		},
		//{
		//  Label:           "Ret",
		//	WeightFileOut:   files.WeightRetFile,
		//	GearFile:        files.GearFileRet,
		//	Model:           model.Model_PallyRet(),
		//	SubstituteItems: substituteItemsRet,
		//},
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

func takeDataSample_Random_Seed(slice []stathighs.WeightInput, size int, seed int64) []stathighs.WeightInput {
	if len(slice) < size {
		return slice
	} else {
		rng := rand.New(rand.NewSource(seed))

		copy := slices.Clone(slice)
		rng.Shuffle(len(copy), func(a, b int) { copy[a], copy[b] = copy[b], copy[a] })
		return copy[0:size]
	}
}
