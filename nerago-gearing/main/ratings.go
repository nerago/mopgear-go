package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"math/rand"
	"os"
	"paladin_gearing_go/files"
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/gear_model/model_factory"
	"paladin_gearing_go/items"
	"paladin_gearing_go/loaders"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/weightfind"
	"paladin_gearing_go/weightfind/util_weight"
	"paladin_gearing_go/weightfind/weight_highs"
	"paladin_gearing_go/weightfind/weight_highs/fitting1"
	"paladin_gearing_go/weightfind/weight_highs/fitting2"
	"paladin_gearing_go/weightfind/weight_highs/fitting3"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"
	"strconv"
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
	modelEquipOnly := model_factory.Model_PallyProtMitigation_WithSet()
	targetRatio := modelEquipOnly.SimPriority
	goal := stats.OptimiseGoal_Mitigation

	// fight := stats.Fight_Horridon_HighHeal
	// spec := stats.Spec_PaladinProtDps
	// startGear := files.GearFileProtDps
	// modelEquipOnly := model.Model_PallyProtDps()
	// goal := stats.UpgradeGoal_Dps

	currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(startGear), &modelEquipOnly, setup.MissingEnchant_Panic, printer)
	itemSet := items.FullItemSet_FromMap(currentEquip)

	inputData, simBase := generateRatingsInputFromArtificialStatOverrides_ForBasic(itemSet, printer, simSpeed, modelEquipOnly.SimSpeedUp, modelEquipOnly.StatsForWeighting, spec, goal, fight, modelEquipOnly.Professions, util.TrackProgress_Start())

	process := weight_highs.BasicStatWeightProcess{}
	process.Init(printer)
	process.SetRequiredStats(model_factory.StatsForWeighting_strengthTank)
	process.SetTargetRatios(targetRatio)
	process.SetBaseline(simBase)
	for _, data := range inputData {
		process.AddSimData(data.IncrementStat, uint32(data.IncrementValue), data.SimResult)
	}
	process.Run()
}

type basicStatInput struct {
	IncrementStat  stats.StatType
	IncrementValue int32
	SimResult      stats.SimData
}

func generateRatingsInputFromArtificialStatOverrides_ForBasic(currentItemSet items.FullItemSet, printer *util.PrintRecorder, simSpeed simulate.WowSim_RunSize, speedUp int, requiredStats []stats.StatType, spec stats.SpecType, goal stats.OptimiseGoal, fight stats.WowSim_Fight, profession gear_model.ProfessionInfo, tracker *util.TrackProgress) ([]basicStatInput, stats.SimData) {
	var incrementValue int32 = 250

	initialBaseStats := weightfind.InitialBonusStatMap_fixRanges(printer, currentItemSet, incrementValue,
		weight_types.FixStatsRangeMode_ExpertiseAlways|weight_types.FixStatsRangeMode_HasteHigherOnly, false)
	tracker.RunOuterTracking(len(requiredStats) + 1)
	defer tracker.SetDone()

	simBase := simulate.WowSim_Execute_SpecifyAll(simSpeed, speedUp, spec, goal, fight, profession, currentItemSet.Items(), nil, tracker.NewChild())

	inputList := util_async.Map_SliceToSlice(len(requiredStats), requiredStats, func(incStat *stats.StatType) basicStatInput {
		innerPrint := util.PrintRecorder_HoldAll()

		bonusStat := initialBaseStats.Clone()
		str := util.StringBuild2{}
		str.WriteString("STATS SCENARIO ")
		bonusStat.Put(*incStat, bonusStat.GetOrPanic(*incStat)+incrementValue)
		str.WriteString(incStat.Name())
		str.WriteRune('=')
		str.WriteInt32(bonusStat.GetOrPanic(*incStat))
		str.WriteRune(' ')

		simResult := simulate.WowSim_Execute_SpecifyAll(simSpeed, speedUp, spec, goal, fight, profession, currentItemSet.Items(), bonusStat, tracker.NewChild())

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

func statWeightsFormula(printer *util.PrintRecorder) {
	// weightInputs, targetRatio := generateRatingsInputFromRealRandomSets(printer)

	weightInputs := readWeightInputFile("sim-stats-compare-rand.json")

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

	comp := weight_highs.FormulaStatWeightProcess2{}

	comp.Init(printer)
	comp.SetRequiredStats(model_factory.StatsForWeighting_strengthTank)
	comp.SetTargetRatios(model_factory.SimPriority_generalMiti)
	comp.SetMinimumIncludeRate(1.0)
	comp.SupplyData(filteredInput)
	weightResult := comp.Run(3000).WaitForResultOrPanic()
	//weights2 := weightResult.Weight
	weights1 := weightResult.AsWeight1()
	if weights1 != nil {
		tools.WritePawnString(*weights1, printer)
	} else {
		printer.Println("MISSING WEIGHT")
	}
}

func statWeightsRanking(printer *util.PrintRecorder) {
	//weightInputs := readWeightInputFile("sim-stats-compare-rand.json")
	weightInputs1 := readWeightInputFile("tempdata\\weightfind-sim-real-Prot-Heal.json")
	weightInputs2 := readWeightInputFile("tempdata\\weightfind-sim-grid-Prot-Heal.json")
	weightInputs := slices.Concat(weightInputs1, weightInputs2)

	statList := model_factory.StatsForWeighting_strengthTank
	ratio := model_factory.SimPriority_heal

	//comp := weight_highs.RankingSeparatedWeights{}
	//comp.Init(printer, 3000)
	//comp.SetRequiredStats(model_factory.StatsForWeighting_strengthTank, model_factory.SimPriority_generalMiti.SimTypes())
	//comp.SetTargetRatios(model_factory.SimPriority_generalMiti)
	//comp.SupplyData(weightInputs)
	//weightResult := comp.Run().WaitForResultOrPanic()
	//weights1 := weightResult.AsWeight1()
	//if weights1 != nil {
	//	tools.WritePawnString(*weights1, printer)
	//} else {
	//	printer.Println("MISSING WEIGHT")
	//}

	ranking := weight_highs.RankingStatWeightProcess{}
	ranking.RANKMODE = 0
	ranking.WEIGHTSUM = 1 // 0 or 1
	ranking.Init(printer)
	ranking.SetRequiredStats(statList)
	ranking.SetTargetRatios(ratio)
	ranking.SupplyData(weightInputs)
	weightResult := ranking.Run(1000).WaitForResultOrPanic()
	weights1 := weightResult.AsWeight1()
	if weights1 != nil {
		tools.WritePawnString(*weights1, printer)
		acc := weightfind.EvaluateAccuracy(weights1, ratio.SimTypes(), &ratio, weightInputs)
		printer.Printf("acc %f\n", acc)
	} else {
		printer.Println("MISSING WEIGHT")
	}
}

func statWeightsRanking3b(printer *util.PrintRecorder) {
	//weightInputs := readWeightInputFile("sim-stats-compare-rand.json")
	//weightInputs1 := readWeightInputFile("tempdata\\weightfind-sim-real-Prot-Heal.json")
	//weightInputs2 := readWeightInputFile("tempdata\\weightfind-sim-grid-Prot-Heal.json")
	weightInputs1 := readWeightInputFile("tempdata\\weightfind-sim-real-Prot-Mitigation-NoSet.json")
	weightInputs2 := readWeightInputFile("tempdata\\weightfind-sim-grid-Prot-Mitigation-NoSet.json")
	weightInputs := slices.Concat(weightInputs1, weightInputs2)

	statList := model_factory.StatsForWeighting_strengthTank
	//ratio := model_factory.SimPriority_heal
	ratio := model_factory.SimPriority_generalMiti

	ranking := weight_highs.RankingStatWeightProcess3c{}
	//ranking := weight_highs.RankingStatWeightProcess3b{}
	//ranking.TOTALWEIGHT = 2
	//ranking.ALGO = 0
	ranking.Init(printer, 1000)
	ranking.SetRequiredStats(statList)
	ranking.SetTargetRatios(ratio)
	ranking.SupplyData(weightInputs)
	//weightsFuture = ranking.RunSinglePassFromExternal(bestWeightsSoFar.weight)
	weightsFuture := ranking.RunMultiRound()
	//weightsFuture := ranking.RunSinglePassRaw() // 8m35.3713307s Objective value 2.4803613132e+08
	weightResult := weightsFuture.WaitForResultOrPanic()
	weights1 := weightResult.AsWeight1()
	if weights1 != nil {
		tools.WritePawnString(*weights1, printer)
		acc := weightfind.EvaluateAccuracy(weights1, ratio.SimTypes(), &ratio, weightInputs)
		printer.Printf("acc %f\n", acc)
	} else {
		printer.Println("MISSING WEIGHT")
	}
}

func statWeightsCustom(printer *util.PrintRecorder) {
	// weightInputs, targetRatio := generateRatingsInputFromRealRandomSets(printer)
	targetRatio := model_factory.SimPriority_generalMiti
	weightStats := model_factory.StatsForWeighting_strengthTank

	//inputDataGrid := readWeightInputFile("sim-stats-compare-grid.json")
	//inputDataRandom := readWeightInputFile("sim-stats-compare-rand.json")
	inputDataGrid := readWeightInputFile("tempdata\\weightfind-sim-grid-Prot-Mitigation-NoSet.json")
	inputDataRandom := readWeightInputFile("tempdata\\weightfind-sim-real-Prot-Mitigation-NoSet.json")
	// mixedInputData := slices.Concat(inputDataGrid, inputDataRandom)
	mixedInputData := slices.Concat(inputDataRandom, inputDataGrid)

	filteredInput := mixedInputData
	printer.Printf("mixedInputData size %d\n", len(filteredInput))

	//search := weightfind.WeightSearcher0{}
	//search.Init(weightStats, targetRatio, util.PrintRecorder_Nop())
	//search.Init(weightStats, targetRatio, printer)
	//search.SupplyData(mixedInputData)

	search := weightfind.WeightSearcher2{}
	//search.Init(weightStats, targetRatio, printer)
	search.Init(weightStats, targetRatio, nil)
	search.SupplyData(inputDataGrid)
	search.SetRanges(-1.0, 10.0)

	weightResult := search.Run(util_async.CancelSignal_Make())
	printer.Printf("time = %s\n", weightResult.SolveTime)
	weight := *weightResult.AsWeight1()
	tools.WritePawnString(weight, printer)
	printer.Printf("accuracy = %f\n", weightfind.EvaluateAccuracy(&weight, targetRatio.SimTypes(), &targetRatio, mixedInputData))

	//( Pawn: v1: "Gearing Weights": Class=Paladin,Strength=1.0000000000,Stamina=1.6065881006,CritRating=0.6369231133,HasteRating=1.5962452471,ExpertiseRating=-0.0001959652,MasteryRating=1.6330798479,DodgeRating=0.9962463273,ParryRating=0.6430861217, )
	//accuracy = 92.632887 92.633057(updated)

	//( Pawn: v1: "Gearing Weights": Class=Paladin,Strength=5.5997314453,Stamina=8.9848632812,CritRating=3.3398437500,HasteRating=8.7237148285,ExpertiseRating=0.0181579590,MasteryRating=8.9848632812,DodgeRating=5.6239013672,ParryRating=3.4168853760, )
	//accuracy = 92.702569

	// WeightSearcher0:
	// using nice version was accuracy = 92.464871 //Duration = 15m58.0662216s
	// using first inline version accuracy = 92.461819 //Duration = 13m34.2183064s
	// with one pointer level removed, after accuracy fix = 92.525906 Duration = 15m10.0442006s
	// fully blocky accuracy = 92.568800 //Duration = 26m57.7175777s. some background may have happenned
	// inlined2, back to full pointers accuracy = 92.441813 //Duration = 13m10.6024466s
	// all still had unnecessary internal printers

	//( Pawn: v1: "Gearing Weights": Class=Paladin,Strength=5.3900000000,Stamina=7.4731250000,CritRating=2.6908057851,
	//HasteRating=7.4200000000,ExpertiseRating=-0.5654897163,MasteryRating=8.0470000000,DodgeRating=4.8125000000,ParryRating=3.2600000000, )
	//accuracy = 92.459276
	//Duration = 11m6.0417638s
	// Search2
	// basically working: accuracy = 92.120870 Duration = 6.2703352s

	// from a full run:
	//search2                          1.0000    1.4739    0.1710      1.4739    0.1710    1.4739    1.0000    0.6448  92.0990                     92.3286       19.5663546s
	//ranking5                         1.0000    1.2018    0.4893      1.6454   -0.1401    1.4705    0.9258    0.6931  92.2846                     92.4398       41m40.039692s
	//ranking0                         1.0000    1.4935    0.3678      1.7556    0.1588    1.4678    0.8076    0.6879  92.3865                     92.3479       40.1894627s
	//search1                          1.0000    1.3952    0.5016      1.3766   -0.1044    1.4976    0.8998    0.6141  92.4257                     92.6758       14m29.8025657s
	//search0                          1.0000    1.4969    0.4926      1.4637    0.0014    1.4948    1.0003    0.5206  92.6007                     92.7165       8m44.6815893s
	// might have still had search2 using opSearch1, given 2 was broken

	// using searcher1
	//EvaluateAccuracyNoRangeAntiInline baseline accuracy = 92.353482 Duration = 49.3194057s 48.3001556s
	//EvaluateAccuracyNoRangeInlined1 2m0.3090194s 2m13.7903418s
	//these version are also the most blocky, array flat, input copied into it too
	//change to WeightInput reference is pointer. 1m47.45968s 1m44.2405868s
	//both as pointers Duration = 1m12.6290301s 1m19.442192s
	//just outer pointer, copy input 1m14.7939587s 1m18.7979703s
	//use accuracyInfoX, back to both pointers 1m13.6677116s 1m20.990594s
	//sortSimType funcs Duration = 49.0108431s
	//sortSimType2 1m21.7422665s 1m25.5339404s
	//sortSimType3 1m22.3571741s
	//back to sortSimType  53.4168611s

	//EvaluateAccuracyRanged0  0.923655 Duration = 47.67764s  0.923655 Duration = 49.843148s
	//EvaluateAccuracyWithRange2  accuracy = 0.923655 Duration = 39.271139s 0.923655 Duration = 40.1025633s
	//EvaluateAccuracyWithRangePartialRefactor3 0.923655 Duration = 43.9915364s time = 44.0702381s
	//EvaluateAccuracyFullRangeInlined4

	//latest accuracy = 92.595079 Duration = 39.9764841s
	//cached 91.739231 8.0985595s

	//accuracy = 84.631348 Duration = 4m4.8212136s
	// fixed accuracy = 92.137316	Duration = 5m34.6080965s

	// search2 accuracy = 92.488437 Duration = 5m4.0083467s ( Pawn: v1: "Gearing Weights": Class=Paladin,Strength=1.0000000000,Stamina=1.2485875706,CritRating=-0.0021186441,HasteRating=1.2485875706,ExpertiseRating=0.0056497175,MasteryRating=1.0000000000,DodgeRating=0.3785310734,ParryRating=0.3785310734, )
	// accuracy = 92.488437( Pawn: v1: "Gearing Weights": Class=Paladin,Strength=5.5312500000,Stamina=6.9062500000,CritRating=-0.0117187500,HasteRating=6.9062500000,ExpertiseRating=0.0312500000,MasteryRating=5.5312500000,DodgeRating=2.0937500000,ParryRating=2.0937500000, )

}

func statWeightsGridIntoRanking(printer *util.PrintRecorder) {
	targetRatio := model_factory.SimPriority_generalMiti
	requiredStats := model_factory.StatsForWeighting_strengthTank
	simTypes := targetRatio.SimTypes()

	inputDataGrid := readWeightInputFile("sim-stats-compare-grid.json")
	inputDataRandom := readWeightInputFile("sim-stats-compare-rand.json")
	mixedInputData := slices.Concat(inputDataRandom, inputDataGrid)

	var weights1 weight_types.Weight1Basic
	if false {
		grid := weight_highs.GridStatWeightProcess{}
		grid.Init(printer, 3000)
		grid.SetRequiredStats(requiredStats)
		grid.SetTargetRatios(targetRatio)
		grid.SupplyData(inputDataGrid)
		weightResult := grid.Run().WaitForResultOrPanic()
		weights1 = *weightResult.AsWeight1()
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

		// better than it thinks is optimal?
		weights1 = weight_types.Weight1Basic_Make()
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

	ranking := weight_highs.RankingStatWeightProcess3b{}
	// 0 is all combinations, 1 is just adjacent
	ranking.ALGO = 0
	// ranking := stathighs.RankingStatWeightProcess4{}
	ranking.Init(printer, 300)
	ranking.SetRequiredStats(requiredStats)
	ranking.SetTargetRatios(targetRatio)
	//ranking.SupplyData(mixedInputData)
	ranking.SupplyData(mixedInputData)
	weightsResult2 := ranking.RunSinglePassFromExternal(weights1).WaitForResultOrPanic()
	weights2 := *weightsResult2.AsWeight1()

	tools.WritePawnString(weights1, printer)
	printer.Printf("accuracy_initial = %f\n", weightfind.EvaluateAccuracy(&weights1, simTypes, &targetRatio, mixedInputData))

	tools.WritePawnString(weights2, printer)
	printer.Printf("accuracy_algo = %f\n", weightfind.EvaluateAccuracy(&weights2, simTypes, &targetRatio, mixedInputData))

	weights3, _ := weightfind.WeightTweakerWithLogging(weights2, requiredStats, &targetRatio, mixedInputData, util.PrintRecorder_Nop())

	tools.WritePawnString(weights3, printer)
	printer.Printf("accuracy_tweak = %f\n", weightfind.EvaluateAccuracy(&weights3, simTypes, &targetRatio, mixedInputData))

	// ( Pawn: v1: "Protection WoWSims Weights": Class=Paladin,Strength=1.0000000000,Stamina=0.4805050000,CritRating=0.6462260000,HasteRating=0.8598560000,ExpertiseRating=0.6679750000,MasteryRating=1.9405810000,DodgeRating=0.6518220000,ParryRating=0.6243300000, )
	// accuracy1 = 92.635522
	// ( Pawn: v1: "Protection WoWSims Weights": Class=Paladin,Strength=1.0000000000,Stamina=-0.0896998019,CritRating=0.3760289134,HasteRating=0.4969834753,ExpertiseRating=0.3863096443,MasteryRating=1.1063898778,DodgeRating=0.3787533903,ParryRating=0.3581785849, )
	// accuracy2 = 91.501292
	// Duration = 2h2m1.7439799s

}

func statWeightsFitting(printer *util.PrintRecorder) {
	// generateRatingsInputFromRealRandomSets(printer)

	bytes, err := os.ReadFile("sim-stats-compare-rand.json")
	// bytes, err := os.ReadFile("sim-stats-input-data.json")
	if err != nil {
		panic(err)
	}
	var weightInputs []weight_types.WeightInput
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

	sampleDataPreScale := util_collection.MapSliceAsNew(weightInputs, func(input *weight_types.WeightInput) util_weight.FittingSample {
		return util_weight.FittingSample{
			StatValue: input.TotalStat.GetFloat(stats.Stat_Haste),
			SimResult: input.SimResult.Get(stats.Sim_DPS),
		}
	})

	statMin := util_collection.FindMinFunc(sampleDataPreScale, func(s util_weight.FittingSample) float64 { return s.StatValue })
	statMax := util_collection.FindMaxFunc(sampleDataPreScale, func(s util_weight.FittingSample) float64 { return s.StatValue })
	simMax := util_collection.FindMaxFunc(sampleDataPreScale, func(s util_weight.FittingSample) float64 { return s.SimResult })
	sampleData := util_collection.MapSliceAsNew(sampleDataPreScale, func(sample *util_weight.FittingSample) util_weight.FittingSample {
		return util_weight.FittingSample{
			StatValue: sample.StatValue / statMax,
			SimResult: sample.SimResult / simMax,
		}
	})

	fitting := fitting1.FittingSingleStatSegmentsProcess{}
	fitting.Init(printer, 3000)
	fitting.SupplyData(sampleData)

	weightMap := fitting.Run(util_async.CancelSignal_Make())
	printer.Printf("weightMap size %d\n", len(weightMap))
	weightList := slices.SortedFunc(maps.Values(weightMap), func(a, b fitting1.FittingSingleStatResult) int { return cmp.Compare(a.Minimum, b.Minimum) })

	tab := util.TabulateOutput{}
	tab.SetColumnSpacing(1)
	tab.AddColumnHeader("minF", true)
	tab.AddColumnHeader("maxF", false)
	tab.AddColumnHeader("min", true)
	tab.AddColumnHeader("max", false)
	tab.AddColumnHeader("m", true)
	tab.AddColumnHeader("c", false)
	tab.AddColumnHeader("used", true)
	tab.AddColumnHeader("used%", false)
	tab.AddColumnHeader("total%", false)
	tab.AddColumnHeader("sequence", false)
	for _, oneWeight := range weightList {
		tab.AddRow([]string{
			strconv.FormatFloat(oneWeight.Minimum, 'f', 6, 64),
			strconv.FormatFloat(oneWeight.Maximum, 'f', 6, 64),
			strconv.FormatUint(uint64(math.Round(oneWeight.Minimum*statMax)), 10),
			strconv.FormatUint(uint64(math.Round(oneWeight.Maximum*statMax)), 10),
			strconv.FormatFloat(oneWeight.LineSlope, 'f', 8, 64),
			strconv.FormatFloat(oneWeight.LineOffset, 'f', 5, 64),
			strconv.FormatUint(uint64(oneWeight.IncludeCount), 10),
			strconv.FormatFloat(oneWeight.IncludePercentOfStageInput*100, 'f', 1, 64),
			strconv.FormatFloat(float64(oneWeight.IncludeCount)/float64(len(sampleData))*100, 'f', 1, 64),
			strconv.FormatUint(uint64(oneWeight.BuiltSequence), 10),
		})
	}
	tab.Write(printer)

	printer.Printf("stat,target,")
	for _, oneWeight := range weightList {
		printer.Printf("weight%d,", oneWeight.BuiltSequence)
	}
	printer.Println0()

	skip := uint32(250)
	startVal := float64((uint32(statMin) / skip) * skip)
	for stat := startVal; stat < statMax; stat += float64(skip) {
		sampleDataPreScale = append(sampleDataPreScale, util_weight.FittingSample{
			StatValue: stat,
			SimResult: 0,
		})
	}
	slices.SortFunc(sampleDataPreScale, func(a, b util_weight.FittingSample) int {
		return cmp.Compare(a.StatValue, b.StatValue)
	})
	for _, sample := range sampleDataPreScale {
		printer.Printf("%.0f,", sample.StatValue)
		if sample.SimResult != 0 {
			printer.Printf("%.0f,", sample.SimResult)
		} else {
			printer.Printf(",")
		}
		for _, oneWeight := range weightList {
			statValue := sample.StatValue / statMax
			guessSim := statValue*oneWeight.LineSlope + oneWeight.LineOffset
			if statValue >= oneWeight.Minimum && statValue <= oneWeight.Maximum {
				printer.Printf("%.0f,", guessSim*simMax)
			} else {
				printer.Printf(",")
			}
		}
		printer.Println0()
	}

	//for _, sample := range sampleDataPreScale {
	//	printer.Printf("%.0f,%.0f,", sample.StatValue, sample.SimResult)
	//	for _, oneWeight := range weightList {
	//		statValue := sample.StatValue / statMax
	//		guessSim := statValue*oneWeight.LineSlope + oneWeight.LineOffset
	//		printer.Printf("%.0f,", guessSim*simMax)
	//	}
	//	printer.Println0()
	//}
	//for _, sample := range sampleDataPreScale {
	//	printer.Printf("%.0f,%.6f,", sample.StatValue, sample.SimResult/simMax)
	//	for _, oneWeight := range weightList {
	//		statValue := sample.StatValue / statMax
	//		effective := statValue*oneWeight.LineSlope + oneWeight.LineOffset
	//		printer.Printf("%.6f,", effective)
	//	}
	//	printer.Println0()
	//}
}

func statWeightsFitting2(printer *util.PrintRecorder) {
	// generateRatingsInputFromRealRandomSets(printer)

	//bytes, err := os.ReadFile("tempdata/weightfind-sim-real-Prot-Heal.json")
	//bytes, err := os.ReadFile("tempdata/weightfind-sim-grid-Prot-Heal.json")
	//bytes, err := os.ReadFile("sim-stats-compare-rand.json")
	// bytes, err := os.ReadFile("sim-stats-input-data.json")
	//weightInputs := readWeightInputFile("tempdata/weightfind-sim-real-Prot-Heal.json")
	weightInputs := readWeightInputFile("sim-stats-compare-rand.json")

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

	//weightInputs = weightInputs[0:400]

	sampleDataPreScale := util_collection.MapSliceAsNew(weightInputs, func(input *weight_types.WeightInput) util_weight.FittingSample {
		return util_weight.FittingSample{
			//StatValue: input.TotalStat.GetFloat(stats.Stat_Haste),
			StatValue: input.TotalStat.GetFloat(stats.Stat_Dodge),
			SimResult: input.SimResult.Get(stats.Sim_DEATH),
		}
	})

	statMin := util_collection.FindMinFunc(sampleDataPreScale, func(s util_weight.FittingSample) float64 { return s.StatValue })
	statMax := util_collection.FindMaxFunc(sampleDataPreScale, func(s util_weight.FittingSample) float64 { return s.StatValue })
	simMax := util_collection.FindMaxFunc(sampleDataPreScale, func(s util_weight.FittingSample) float64 { return s.SimResult })
	sampleData := util_collection.MapSliceAsNew(sampleDataPreScale, func(sample *util_weight.FittingSample) util_weight.FittingSample {
		return util_weight.FittingSample{
			StatValue: sample.StatValue / statMax,
			SimResult: sample.SimResult / simMax,
		}
	})
	scaleStat := 1.0 / statMax

	fitting := fitting2.SingleSegmented2{}
	fitting.Init(3, scaleStat, printer, 5000)
	fitting.SupplyData(sampleData)

	weightMapCancel := fitting.Run()
	weightMap := weightMapCancel.WaitForResultOrPanic()
	printer.Printf("weightMap size %d\n", len(weightMap.Segments))
	weightList := weightMap.Segments
	slices.SortFunc(weightList, func(a, b fitting2.InitialSegment) int {
		return cmp.Compare(a.StatRange.Minimum, b.StatRange.Minimum)
	})

	fittingCsvDataReport(printer, weightList, statMin, statMax, sampleDataPreScale, simMax)

	fittingTableReport(printer, weightList, statMax, sampleData)

	//for _, sample := range sampleDataPreScale {
	//	printer.Printf("%.0f,%.0f,", sample.StatValue, sample.SimResult)
	//	for _, oneWeight := range weightList {
	//		statValue := sample.StatValue / statMax
	//		guessSim := statValue*oneWeight.LineSlope + oneWeight.LineOffset
	//		printer.Printf("%.0f,", guessSim*simMax)
	//	}
	//	printer.Println0()
	//}
	//for _, sample := range sampleDataPreScale {
	//	printer.Printf("%.0f,%.6f,", sample.StatValue, sample.SimResult/simMax)
	//	for _, oneWeight := range weightList {
	//		statValue := sample.StatValue / statMax
	//		effective := statValue*oneWeight.LineSlope + oneWeight.LineOffset
	//		printer.Printf("%.6f,", effective)
	//	}
	//	printer.Println0()
	//}
}

func statWeightsFitting2each(printer *util.PrintRecorder) {
	weightInputs := readWeightInputFile("tempdata/weightfind-sim-real-Prot-Heal.json")

	task := func(simType stats.SimType, statType stats.StatType) {
		sampleDataPreScale := util_collection.MapSliceAsNew(weightInputs, func(input *weight_types.WeightInput) util_weight.FittingSample {
			return util_weight.FittingSample{
				StatValue: input.TotalStat.GetFloat(statType),
				SimResult: input.SimResult.Get(simType),
			}
		})

		statMax := util_collection.FindMaxFunc(sampleDataPreScale, func(s util_weight.FittingSample) float64 { return s.StatValue })
		simMax := util_collection.FindMaxFunc(sampleDataPreScale, func(s util_weight.FittingSample) float64 { return s.SimResult })
		sampleData := util_collection.MapSliceAsNew(sampleDataPreScale, func(sample *util_weight.FittingSample) util_weight.FittingSample {
			return util_weight.FittingSample{
				StatValue: sample.StatValue / statMax,
				SimResult: sample.SimResult / simMax,
			}
		})
		scaleStat := 1.0 / statMax

		fitting := fitting2.SingleSegmented2{}
		innerPrint := util.PrintRecorder_Nop()
		fitting.Init(3, scaleStat, innerPrint, 1000)
		fitting.SupplyData(sampleData)

		weightMapCancel := fitting.Run()
		weightMap := weightMapCancel.WaitForResultOrPanic()
		weightList := weightMap.Segments
		slices.SortFunc(weightList, func(a, b fitting2.InitialSegment) int {
			return cmp.Compare(a.StatRange.Minimum, b.StatRange.Minimum)
		})

		fittingTableReport(printer, weightList, statMax, sampleData)
	}

	simTypes := model_factory.SimPriority_heal.SimTypes()
	util_async.ForEach_Slice(8, simTypes, func(sim *stats.SimType) {
		for _, statType := range model_factory.StatsForWeighting_strengthTank {
			task(*sim, statType)
		}
	})
}

func statWeightsFitting2eachProper(printer *util.PrintRecorder) {
	//bytes, err := os.ReadFile("")
	weightInputs := readWeightInputFile("sim-stats-compare-rand.json")
	//weightInputs := readWeightInputFile("tempdata/weightfind-sim-real-Prot-Heal.json")
	weightInputs = weightInputs[0:10]
	simTypes := model_factory.SimPriority_heal.SimTypes()
	statTypes := model_factory.StatsForWeighting_strengthTank
	targetRatio := model_factory.SimPriority_heal

	fitting := fitting2.FittingEachStatWeightProcess2{}
	fitting.Init(3, printer, 1000)
	fitting.SetRequiredStats(statTypes, simTypes)
	fitting.SetTargetRatios(targetRatio)
	fitting.SupplyData(weightInputs)
	weightResult := fitting.Run(util_async.CancelSignal_Make())
	weight3 := weightResult.Weight.(*weight_types.Weight3ExtendedRanged)

	tools.WriteWeightString(weight3, printer)
	printer.Printf("weight3 isempty = %v\n", weight3.IsEmpty())
	printer.Printf("weight2 isempty = %v\n", weight3.ConvertToWeight2().IsEmpty())
	printer.Printf("weight1 isempty = %v\n", weight3.ConvertToWeight2().ConvertToWeight1().IsEmpty())

	acc3 := weightfind.EvaluateAccuracy(weight3, simTypes, &targetRatio, weightInputs)
	acc2 := weightfind.EvaluateAccuracy(weight3.ConvertToWeight2(), simTypes, &targetRatio, weightInputs)
	acc1 := weightfind.EvaluateAccuracy(weight3.ConvertToWeight2().ConvertToWeight1(), simTypes, &targetRatio, weightInputs)
	printer.Printf("weight3 acc = %v\n", acc3)
	printer.Printf("weight2 acc = %v\n", acc2)
	printer.Printf("weight1 acc = %v\n", acc1)

}

func statWeightsFitting3eachProper(printer *util.PrintRecorder) {
	weightInputs := readWeightInputFile("tempdata/weightfind-sim-real-Prot-Mitigation-NoSet.json")
	//weightInputs = weightInputs[0:30]
	simTypes := model_factory.SimPriority_heal.SimTypes()
	statTypes := model_factory.StatsForWeighting_strengthTank
	targetRatio := model_factory.SimPriority_generalMiti

	fitting := fitting3.FittingEachStatWeightProcess3{}
	fitting.Init(3, printer, 1000)
	fitting.SetRequiredStats(statTypes, simTypes)
	fitting.SetTargetRatios(targetRatio)
	fitting.SupplyData(weightInputs)
	weightResult := fitting.Run(util_async.CancelSignal_Make())
	weight3 := weightResult.Weight.(*weight_types.Weight3ExtendedRanged)

	tools.WriteWeightString(weight3, printer)
	printer.Printf("weight3 isempty = %v\n", weight3.IsEmpty())
	printer.Printf("weight2 isempty = %v\n", weight3.ConvertToWeight2().IsEmpty())
	printer.Printf("weight1 isempty = %v\n", weight3.ConvertToWeight2().ConvertToWeight1().IsEmpty())

	acc3 := weightfind.EvaluateAccuracy(weight3, simTypes, &targetRatio, weightInputs)
	acc2 := weightfind.EvaluateAccuracy(weight3.ConvertToWeight2(), simTypes, &targetRatio, weightInputs)
	acc1 := weightfind.EvaluateAccuracy(weight3.ConvertToWeight2().ConvertToWeight1(), simTypes, &targetRatio, weightInputs)
	printer.Printf("weight3 acc = %v\n", acc3)
	printer.Printf("weight2 acc = %v\n", acc2)
	printer.Printf("weight1 acc = %v\n", acc1)

}

func statWeightsFitting1eachProper(printer *util.PrintRecorder) {
	//bytes, err := os.ReadFile("")
	//weightInputs := readWeightInputFile("sim-stats-compare-rand.json")
	//weightInputs := readWeightInputFile("tempdata/weightfind-sim-real-Prot-Heal.json")
	weightInputs := readWeightInputFile("tempdata/weightfind-sim-real-Prot-Mitigation-WithSet.json")
	//weightInputs = weightInputs[0:10]
	statTypes := model_factory.StatsForWeighting_strengthTank
	targetRatio := model_factory.SimPriority_withSet
	simTypes := targetRatio.SimTypes()

	fitting := fitting1.FittingEachStatWeightProcess{}
	fitting.Init(printer, 1000)
	fitting.SetRequiredStats(statTypes, simTypes)
	fitting.SetOnlyComputeSingleSegmentEach(true)
	fitting.SupplyData(weightInputs)
	weightResult := fitting.Run(util_async.CancelSignal_Make())
	weight3 := weightResult.Weight.(*weight_types.Weight3ExtendedRanged)

	tools.WriteWeightString(weight3, printer)
	printer.Printf("weight3 isempty = %v\n", weight3.IsEmpty())
	printer.Printf("weight2 isempty = %v\n", weight3.ConvertToWeight2().IsEmpty())
	printer.Printf("weight1 isempty = %v\n", weight3.ConvertToWeight2().ConvertToWeight1().IsEmpty())

	acc3 := weightfind.EvaluateAccuracy(weight3, simTypes, &targetRatio, weightInputs)
	acc2 := weightfind.EvaluateAccuracy(weight3.ConvertToWeight2(), simTypes, &targetRatio, weightInputs)
	acc1 := weightfind.EvaluateAccuracy(weight3.ConvertToWeight2().ConvertToWeight1(), simTypes, &targetRatio, weightInputs)
	printer.Printf("weight3 acc = %v\n", acc3)
	printer.Printf("weight2 acc = %v\n", acc2)
	printer.Printf("weight1 acc = %v\n", acc1)

}

func fittingTableReport(printer *util.PrintRecorder, weightList []fitting2.InitialSegment, statMax float64, sampleData []util_weight.FittingSample) {
	tab := util.TabulateOutput{}
	tab.SetColumnSpacing(1)
	tab.AddColumnHeader("minF", true)
	tab.AddColumnHeader("maxF", false)
	tab.AddColumnHeader("min", true)
	tab.AddColumnHeader("max", false)
	tab.AddColumnHeader("m", true)
	tab.AddColumnHeader("c", false)
	tab.AddColumnHeader("used", true)
	tab.AddColumnHeader("used%", false)
	tab.AddColumnHeader("total%", false)
	tab.AddColumnHeader("sequence", false)
	for _, oneWeight := range weightList {
		tab.AddRow([]string{
			strconv.FormatFloat(oneWeight.StatRange.Minimum, 'f', 6, 64),
			strconv.FormatFloat(oneWeight.StatRange.Maximum, 'f', 6, 64),
			strconv.FormatUint(uint64(math.Round(oneWeight.StatRange.Minimum*statMax)), 10),
			strconv.FormatUint(uint64(math.Round(oneWeight.StatRange.Maximum*statMax)), 10),
			strconv.FormatFloat(oneWeight.LineSlope, 'f', 8, 64),
			strconv.FormatFloat(oneWeight.LineOffset, 'f', 5, 64),
			strconv.FormatUint(uint64(oneWeight.IncludeCount), 10),
			"", //strconv.FormatFloat(oneWeight.IncludePercentOfStageInput*100, 'f', 1, 64),
			strconv.FormatFloat(float64(oneWeight.IncludeCount)/float64(len(sampleData))*100, 'f', 1, 64),
			"", //strconv.FormatUint(uint64(oneWeight.BuiltSequence), 10),
		})
	}
	tab.Write(printer)
}

func statWeightsFitting3(printer *util.PrintRecorder) {
	//weightInputs := readWeightInputFile("tempdata/weightfind-sim-real-Prot-Heal.json")
	weightInputs := readWeightInputFile("sim-stats-compare-rand.json")
	checkSimType := stats.Sim_TMI
	checkStatType := stats.Stat_Dodge

	printer.Printf("Initial weight input size = %d\n", len(weightInputs))

	//weightInputs = weightInputs[0:400]

	scaleStats := util_weight.ChooseStatScalingBasic(weightInputs, 1.0, true, printer)
	scaleStat := scaleStats.GetOrPanic(checkStatType)
	scaleSims := util_weight.ChooseSimDetailUnitScaleAndOffset(weightInputs, []stats.SimType{checkSimType})
	scaleSim := scaleSims.GetOrPanic(checkSimType)

	sampleData := util_collection.MapSliceAsNew(weightInputs, func(input *weight_types.WeightInput) util_weight.FittingSample3 {
		detail := input.SimResult.GetDetailed2(checkSimType)
		avg := input.SimResult.Get(checkSimType)
		sim := util_weight.MakeFittingDetail(avg, detail, scaleSim)
		return util_weight.FittingSample3{
			StatValue: input.TotalStat.GetFloat(stats.Stat_Dodge),
			SimResult: sim,
		}
	})

	//statMin := util_collection.FindMinFunc(weightInputs, func(s weight_types.WeightInput) float64 { return s.TotalStat.GetFloat(checkStatType) })
	//statMax := util_collection.FindMaxFunc(weightInputs, func(s weight_types.WeightInput) float64 { return s.TotalStat.GetFloat(checkStatType) })
	//simMax := util_collection.FindMaxFunc(weightInputs, func(s weight_types.WeightInput) float64 { return s.SimResult })

	fitting := fitting3.FittingSingleSegmented3{}
	fitting.Init(3, scaleStat, printer, 5000)
	fitting.SupplyData(sampleData)

	weightMapCancel := fitting.Run()
	weightMap := weightMapCancel.WaitForResultOrPanic()
	printer.Printf("weightMap size %d\n", len(weightMap.Segments))
	weightList := weightMap.Segments
	slices.SortFunc(weightList, func(a, b fitting2.InitialSegment) int {
		return cmp.Compare(a.StatRange.Minimum, b.StatRange.Minimum)
	})

	//fittingCsvDataReport(printer, weightList, statMin, statMax, sampleDataPreScale, simMax)
	//
	//fittingTableReport(printer, weightList, statMax, sampleData)

	//for _, sample := range sampleDataPreScale {
	//	printer.Printf("%.0f,%.0f,", sample.StatValue, sample.SimResult)
	//	for _, oneWeight := range weightList {
	//		statValue := sample.StatValue / statMax
	//		guessSim := statValue*oneWeight.LineSlope + oneWeight.LineOffset
	//		printer.Printf("%.0f,", guessSim*simMax)
	//	}
	//	printer.Println0()
	//}
	//for _, sample := range sampleDataPreScale {
	//	printer.Printf("%.0f,%.6f,", sample.StatValue, sample.SimResult/simMax)
	//	for _, oneWeight := range weightList {
	//		statValue := sample.StatValue / statMax
	//		effective := statValue*oneWeight.LineSlope + oneWeight.LineOffset
	//		printer.Printf("%.6f,", effective)
	//	}
	//	printer.Println0()
	//}
}

func fittingCsvDataReport(printer *util.PrintRecorder, weightList []fitting2.InitialSegment, statMin float64, statMax float64, sampleDataPreScale []util_weight.FittingSample, simMax float64) {
	printer.Printf("stat,target,")
	for i := range weightList {
		printer.Printf("weight%d,", i)
	}
	printer.Println0()

	skip := uint32(250)
	startVal := float64((uint32(statMin) / skip) * skip)
	for stat := startVal; stat < statMax; stat += float64(skip) {
		sampleDataPreScale = append(sampleDataPreScale, util_weight.FittingSample{
			StatValue: stat,
			SimResult: 0,
		})
	}
	slices.SortFunc(sampleDataPreScale, func(a, b util_weight.FittingSample) int {
		return cmp.Compare(a.StatValue, b.StatValue)
	})
	for _, sample := range sampleDataPreScale {
		printer.Printf("%.0f,", sample.StatValue)
		if sample.SimResult != 0 {
			printer.Printf("%e,", sample.SimResult)
		} else {
			printer.Printf(",")
		}
		for _, oneWeight := range weightList {
			statValue := sample.StatValue / statMax
			guessSim := statValue*oneWeight.LineSlope + oneWeight.LineOffset
			if statValue >= oneWeight.StatRange.Minimum && statValue <= oneWeight.StatRange.Maximum {
				printer.Printf("%e,", guessSim*simMax)
			} else {
				printer.Printf(",")
			}
		}
		printer.Println0()
	}
}

func statWeightsFitting1a(printer *util.PrintRecorder) {
	bytes, err := os.ReadFile("sim-stats-input-data2.json")
	// bytes, err := os.ReadFile("sim-stats-input-data.json")
	if err != nil {
		panic(err)
	}
	var weightInputs []weight_types.WeightInput
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

	fitting := fitting1.FittingEachStatWeightProcess{}
	fitting.Init(printer, 3000)
	fitting.SupplyData(weightInputs)

	weightResult := fitting.Run(util_async.CancelSignal_Make())
	weight3 := weightResult.Weight.(*weight_types.Weight3ExtendedRanged)
	for entry := range weight3.StatWeights.SeqKey1Key2ValueSeqEntries() {
		weightSeq := entry.ValueSeq

		printer.Printf("################### %s %s ###################\n", entry.Key1.Name(), entry.Key2.Name())

		tab := util.TabulateOutput{}
		tab.SetColumnSpacing(1)
		tab.AddColumnHeader("min", true)
		tab.AddColumnHeader("max", false)
		tab.AddColumnHeader("m", true)
		tab.AddColumnHeader("c", false)
		tab.AddColumnHeader("quality", true)
		for oneWeight := range weightSeq {
			tab.AddRow([]string{
				strconv.FormatUint(uint64(oneWeight.StatRange.Minimum), 10),
				strconv.FormatUint(uint64(oneWeight.StatRange.Maximum), 10),
				strconv.FormatFloat(oneWeight.RatingWeight, 'f', 6, 64),
				strconv.FormatFloat(oneWeight.RatingOffset, 'f', 1, 64),
				strconv.FormatFloat(oneWeight.EstimationQuality, 'f', 1, 64),
			})
		}
		tab.Write(printer)
	}
}

func statWeightsGrid1Orig(printer *util.PrintRecorder) {
	// inputData, targetRatio := generateRatingsInputFromArtificalStatOverrides(printer)
	// writeWeightInputsToFile(inputData, "sim-stats-input-grid.json" )

	inputDataGrid := readWeightInputFile("tempdata/weightfind-sim-grid-Prot-Mitigation-NoSet.json")
	inputDataRandom := readWeightInputFile("tempdata/weightfind-sim-real-Prot-Mitigation-NoSet.json")
	//inputData := takeDataSample_Random(inputDataFull, 30)
	inputData := inputDataGrid

	targetRatio := model_factory.SimPriority_generalMiti
	//targetRatio := model_factory.SimPriority_dps
	requiredStats := model_factory.StatsForWeighting_strengthTank
	simTypes := targetRatio.SimTypes()

	runOne := func(inc1 int, label string) {
		process := weight_highs.GridStatWeightProcess{}
		process.CHECKRANGE = inc1

		//process.Init(printer, 2000)
		process.Init(util.PrintRecorder_Nop(), 1500)
		process.SetRequiredStats(requiredStats)
		process.SetTargetRatios(targetRatio)
		process.SupplyData(inputData)
		weightsResult := process.Run().WaitForResultOrNilValue()
		weights1 := *weightsResult.AsWeight1()
		tools.WritePawnString(weights1, printer)

		acc := weightfind.EvaluateAccuracy(&weights1, simTypes, &targetRatio, inputDataGrid)
		acc2 := weightfind.EvaluateAccuracy(&weights1, simTypes, &targetRatio, inputDataRandom)
		acc3 := weightfind.EvaluateAccuracy(&weights1, simTypes, &targetRatio, slices.Concat(inputDataGrid, inputDataRandom))
		acc4 := weightfind.EvaluateAccuracyStatistical(&weights1, simTypes, &targetRatio, slices.Concat(inputDataGrid, inputDataRandom))
		printer.Printf("accuracy %s: grid data = %f, rand data = %f, data mix = %f, stat mix = %f\n", label, acc, acc2, acc3, acc4)
	}

	//runOne(true, true, true, 1, 3, "select")

	type optParam struct {
		inc1  int
		label string
	}
	optList := make([]optParam, 0)
	for a := range 2 {
		label := fmt.Sprintf("GRID1 %d", a)
		optList = append(optList, optParam{a, label})
	}

	util_async.ForEach_Slice(5, optList, func(o *optParam) {
		runOne(o.inc1, o.label)
		printer.Println(o.label)
	})

	//for simHigh := range 5 {
	//	printer.Printf("version %d %d\n", 1, simHigh)
	//	runOne(false, true, false, 1, simHigh)
	//}

	// so for old method @1, worse is multiply=2, best is multiply=1
	//  simHigh only really makes much difference once we're in multiply=1
	//  anything that isn't version 0 is fine, I like 3/4

	// so for new method @1, [1 1], [1 2] by far the best

	// old method @2, yay slow

	//for ROUNDMODE := range 3 {
	//	for RESCALE := range 3 {
	//		process := weight_highs.GridStatWeightProcess1C{}
	//		process.ROUNDMODE = ROUNDMODE
	//		process.RESCALE = RESCALE
	//		process.Init(printer, 3000)
	//		process.SetRequiredStats(requiredStats)
	//		process.SetTargetRatios(targetRatio)
	//		process.SupplyData(inputData)
	//		weights := process.Run(nil).WaitForResultOrPanic()
	//		tools.WritePawnString(weights, printer)
	//		acc := weightfind.EvaluateAccuracyRanged(weights, simTypes, targetRatio, inputDataFull)
	//
	//		printer.Printf("accuracy = %f\n", acc)
	//		printer.Printf("############## DONE %d %d\n", ROUNDMODE, RESCALE)
	//
	//		//weights2, _ := weightfind.WeightTweaker(weights, requiredStats, targetRatio, inputDataFull, printer)
	//		//acc2 := weightfind.EvaluateAccuracy(weights2, inputDataFull, targetRatio)
	//		//printer.Printf("accuracy_tweak = %f\n", acc2)
	//	}
	//}
}

func statWeightsGrid2(printer *util.PrintRecorder) {
	// inputData, targetRatio := generateRatingsInputFromArtificalStatOverrides(printer)
	// writeWeightInputsToFile(inputData, "sim-stats-input-grid.json" )

	inputDataGrid := readWeightInputFile("tempdata/weightfind-sim-grid-Prot-Mitigation-NoSet.json")
	inputDataRandom := readWeightInputFile("tempdata/weightfind-sim-real-Prot-Mitigation-NoSet.json")
	//inputData := takeDataSample_Random(inputDataFull, 30)
	inputData := inputDataGrid

	//targetRatio := gear_model.SimRatio_generalMiti
	targetRatio := model_factory.SimPriority_dps
	requiredStats := model_factory.StatsForWeighting_strengthTank
	simTypes := targetRatio.SimTypes()

	runOne := func(inc1, inc2, inc3 bool, label string) {
		process := weight_highs.GridStatWeightProcess2{}
		process.IncludeDiffs1 = inc1
		process.IncludeDiffs2 = inc2
		process.IncludeDiffs3 = inc3

		//process.Init(printer, 2000)
		process.Init(util.PrintRecorder_Nop(), 500)
		process.SetRequiredStats(requiredStats)
		process.SetTargetRatios(targetRatio)
		process.SupplyData(inputData)
		weightsResult := process.Run().WaitForResultOrNilValue()
		weights1 := *weightsResult.AsWeight1()
		tools.WritePawnString(weights1, printer)

		acc := weightfind.EvaluateAccuracy(&weights1, simTypes, &targetRatio, inputDataGrid)
		acc2 := weightfind.EvaluateAccuracy(&weights1, simTypes, &targetRatio, inputDataRandom)
		acc3 := weightfind.EvaluateAccuracy(&weights1, simTypes, &targetRatio, slices.Concat(inputDataGrid, inputDataRandom))
		acc4 := weightfind.EvaluateAccuracyStatistical(&weights1, simTypes, &targetRatio, slices.Concat(inputDataGrid, inputDataRandom))
		printer.Printf("accuracy %s: grid data = %f, rand data = %f, data mix = %f, stat mix = %f\n", label, acc, acc2, acc3, acc4)
	}

	type optParam struct {
		inc1, inc2, inc3 bool
		label            string
	}
	optList := make([]optParam, 0)
	for a := range 2 {
		for b := range 2 {
			for c := range 2 {
				if a == 0 && b == 0 && c == 0 {
					continue
				}
				label := fmt.Sprintf("GRID2 %d %d %d", a, b, c)
				optList = append(optList, optParam{a == 1, b == 1, c == 1, label})
			}
		}
	}

	util_async.ForEach_Slice(5, optList, func(o *optParam) {
		runOne(o.inc1, o.inc2, o.inc3, o.label)
		printer.Println(o.label)
	})
}

func statWeightsGrid1b(printer *util.PrintRecorder) {
	// inputData, targetRatio := generateRatingsInputFromArtificalStatOverrides(printer)
	// writeWeightInputsToFile(inputData, "sim-stats-input-grid.json" )

	inputDataGrid := readWeightInputFile("tempdata/weightfind-sim-grid-Prot-Mitigation-NoSet.json")
	inputDataRandom := readWeightInputFile("tempdata/weightfind-sim-real-Prot-Mitigation-NoSet.json")
	//inputData := takeDataSample_Random(inputDataFull, 30)
	inputData := inputDataGrid

	targetRatio := model_factory.SimPriority_generalMiti
	//targetRatio := model_factory.SimPriority_dps
	requiredStats := model_factory.StatsForWeighting_strengthTank
	simTypes := targetRatio.SimTypes()

	runOne := func(a, b, c, d int, label string) {
		grid := weight_highs.GridStatWeightProcess1B{}
		grid.OUTLIER = a
		grid.SCALEMODE = b
		grid.ROUNDMODE = c
		grid.CALCMODE = d
		grid.Init(printer, 1000)
		grid.SetTargetRatios(targetRatio)
		grid.SetRequiredStats(requiredStats)
		grid.SupplyData(inputData)
		weightsFuture := grid.Run()
		weightResult := weightsFuture.WaitForResultOrPanic()
		weights1 := *weightResult.AsWeight1()

		acc := weightfind.EvaluateAccuracy(&weights1, simTypes, &targetRatio, inputDataGrid)
		acc2 := weightfind.EvaluateAccuracy(&weights1, simTypes, &targetRatio, inputDataRandom)
		acc3 := weightfind.EvaluateAccuracy(&weights1, simTypes, &targetRatio, slices.Concat(inputDataGrid, inputDataRandom))
		acc4 := weightfind.EvaluateAccuracyStatistical(&weights1, simTypes, &targetRatio, slices.Concat(inputDataGrid, inputDataRandom))
		printer.Printf("accuracy %s: grid data = %f, rand data = %f, data mix = %f, stat mix = %f\n", label, acc, acc2, acc3, acc4)
	}

	runOne(0, 1, 2, 2, "select")

	//type optParam struct {
	//	a, b, c, d int
	//	label      string
	//}
	//optList := make([]optParam, 0)
	//for a := range 5 {
	//	for b := range 5 {
	//		for c := range 3 {
	//			for d := range 4 {
	//				label := fmt.Sprintf("GRID1 %d %d %d %d", a, b, c, d)
	//				optList = append(optList, optParam{a, b, c, d, label})
	//			}
	//		}
	//	}
	//}
	//
	//util_async.ForEach_Slice(5, optList, func(o *optParam) {
	//	runOne(o.a, o.b, o.c, o.d, o.label)
	//	printer.Println(o.label)
	//})
}

func writeWeightInputsToFile(weightInputs []weight_types.WeightInput, filename string) {
	bytes, err := json.Marshal(weightInputs)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(filename, bytes, 0666)
	if err != nil {
		panic(err)
	}
}

func readWeightInputFile(filename string) []weight_types.WeightInput {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var weightInputs []weight_types.WeightInput
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

func takeDataSample_Start(slice []weight_types.WeightInput, size int) []weight_types.WeightInput {
	if len(slice) < size {
		return slice
	} else {
		return slice[0:size]
	}
}

func takeDataSample_Random(slice []weight_types.WeightInput, size int) []weight_types.WeightInput {
	if len(slice) < size {
		return slice
	} else {
		copy := slices.Clone(slice)
		rand.Shuffle(len(copy), func(a, b int) { copy[a], copy[b] = copy[b], copy[a] })
		return copy[0:size]
	}
}

func takeDataSample_Random_Seed(slice []weight_types.WeightInput, size int, seed int64) []weight_types.WeightInput {
	if len(slice) < size {
		return slice
	} else {
		rng := rand.New(rand.NewSource(seed))

		copy := slices.Clone(slice)
		rng.Shuffle(len(copy), func(a, b int) { copy[a], copy[b] = copy[b], copy[a] })
		return copy[0:size]
	}
}
