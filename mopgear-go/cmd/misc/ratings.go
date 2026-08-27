package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"os"
	"slices"
	"strconv"

	"github.com/nerago/mopgear-go/files"
	"github.com/nerago/mopgear-go/gear_model"
	"github.com/nerago/mopgear-go/gear_model/model_factory"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/loaders"
	"github.com/nerago/mopgear-go/setup"
	"github.com/nerago/mopgear-go/simulate"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/tools"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind"
	"github.com/nerago/mopgear-go/weightfind/util_weight"
	"github.com/nerago/mopgear-go/weightfind/weight_highs"
	"github.com/nerago/mopgear-go/weightfind/weight_highs/fitting1"
	"github.com/nerago/mopgear-go/weightfind/weight_highs/fitting2"
	"github.com/nerago/mopgear-go/weightfind/weight_highs/fitting3"
	"github.com/nerago/mopgear-go/weightfind/weight_highs/fitting4"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
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
	startGear := files.GearFileProtSurvival
	modelEquipOnly := model_factory.Model_PallyProtSurvival()
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

	simBase := simulate.ExecuteSpecifyAll(simSpeed, speedUp, spec, goal, fight, profession, currentItemSet.Items(), nil, tracker.NewChild())

	inputList := util_async.Map_SliceToSlice(len(requiredStats), requiredStats, func(incStat *stats.StatType) basicStatInput {
		innerPrint := util.PrintRecorder_HoldAll()

		bonusStat := initialBaseStats.Clone()
		str := util.StringBuild2{}
		str.WriteString("STATS SCENARIO ")
		bonusStat.Compute(*incStat, func(v int32) int32 { return v + incrementValue })
		str.WriteString(incStat.Name())
		str.WriteRune('=')
		str.WriteInt32(bonusStat.GetOrPanic(*incStat))
		str.WriteRune(' ')

		simResult := simulate.ExecuteSpecifyAll(simSpeed, speedUp, spec, goal, fight, profession, currentItemSet.Items(), bonusStat, tracker.NewChild())

		str.WriteString("   --> ")
		simResult.CompactStringGeneralAppend(&str)
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

	weightInputs := weight_types.WeightInputReadFile("tempdata/sim-stats-compare-rand.json")

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
	comp.SetTargetRatios(model_factory.SimPriority_mitigation)
	comp.SetMinimumIncludeRate(1.0)
	comp.SupplyData(filteredInput)
	weightResult := comp.Run(3000).WaitForResultOrPanic()
	//weights2 := weightResult.Weight
	weights1 := weightResult.AsWeight1(weightInputs)
	if weights1 != nil {
		tools.WritePawnString(*weights1, printer)
	} else {
		printer.Println("MISSING WEIGHT")
	}
}

func statWeightsRanking(printer *util.PrintRecorder) {
	//weightInputs := readWeightInputFile("tempdata/sim-stats-compare-rand.json")
	weightInputs1 := weight_types.WeightInputReadFile("tempdata\\weightfind-sim-real-Prot-Heal.json")
	weightInputs2 := weight_types.WeightInputReadFile("tempdata\\weightfind-sim-grid-Prot-Heal.json")
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
	weights1 := weightResult.AsWeight1(weightInputs)
	if weights1 != nil {
		tools.WritePawnString(*weights1, printer)
		acc := weightfind.EvaluateAccuracyBasic(weights1, ratio.SimTypes(), &ratio, weightInputs)
		printer.Printf("acc %f\n", acc)
	} else {
		printer.Println("MISSING WEIGHT")
	}
}

func statWeightsRanking3b(printer *util.PrintRecorder) {
	//weightInputs := readWeightInputFile("tempdata/sim-stats-compare-rand.json")
	//weightInputs1 := readWeightInputFile("tempdata\\weightfind-sim-real-Prot-Heal.json")
	//weightInputs2 := readWeightInputFile("tempdata\\weightfind-sim-grid-Prot-Heal.json")
	weightInputs1 := weight_types.WeightInputReadFile("tempdata\\weightfind-sim-real-Prot-Mitigation.json")
	weightInputs2 := weight_types.WeightInputReadFile("tempdata\\weightfind-sim-grid-Prot-Mitigation.json")
	weightInputs := slices.Concat(weightInputs1, weightInputs2)

	statList := model_factory.StatsForWeighting_strengthTank
	//ratio := model_factory.SimPriority_heal
	ratio := model_factory.SimPriority_mitigation

	weightsMidRange := weight_types.Weight1Basic_Make()
	weightsMidRange.Put(stats.Stat_Strength, 1.0000)
	weightsMidRange.Put(stats.Stat_Stamina, 1.2309)
	weightsMidRange.Put(stats.Stat_Crit, 0.1167)
	weightsMidRange.Put(stats.Stat_Haste, 0.3614)
	weightsMidRange.Put(stats.Stat_Expertise, 0.0054)
	weightsMidRange.Put(stats.Stat_Mastery, 0.5866)
	weightsMidRange.Put(stats.Stat_Dodge, 0.0824)
	weightsMidRange.Put(stats.Stat_Parry, 0.0532)

	//ranking := weight_highs.RankingStatWeightProcess3c{}
	ranking := weight_highs.RankingStatWeightProcess3b{}
	ranking.TOTALWEIGHT = 2
	// 0 49.3953ms
	// 1 80.5979ms
	// 2 46.1839ms
	ranking.ALGO = 0
	// 1^
	// 0
	// simplex; 31177     9.2792909882e+08 Pr: 0(0) 379.7s; acc 90.397887; Duration = 6m20.218095s
	// hipo; 138   9.27929099e+08   9.27929099e+08   1.05e-05   1.84e-05  1.88e-14   658.9;acc 90.397887;Duration = 11m1.8882727s
	// ipx: 16m40.5983198s timeout
	ranking.Init(printer, 1000)
	ranking.SetRequiredStats(statList)
	ranking.SetTargetRatios(ratio)
	ranking.SupplyData(weightInputs)
	//weightsFuture = ranking.RunSinglePassFromExternal(bestWeightsSoFar.weight)
	weightsFuture := ranking.RunMultiRound()
	//weightsFuture := ranking.RunSinglePassFromExternal(weightsMidRange)
	//weightsFuture := ranking.RunSinglePassRaw() // 8m35.3713307s Objective value 2.4803613132e+08
	weightResult := weightsFuture.WaitForResultOrPanic()
	weights1 := weightResult.AsWeight1(weightInputs)
	if weights1 != nil {
		tools.WritePawnString(*weights1, printer)
		acc := weightfind.EvaluateAccuracyBasic(weights1, ratio.SimTypes(), &ratio, weightInputs)
		printer.Printf("acc %f\n", acc)
	} else {
		printer.Println("MISSING WEIGHT")
	}
}

func statWeightsRanking30(printer *util.PrintRecorder) {
	weightInputs1 := weight_types.WeightInputReadFile("tempdata\\weightfind-sim-real-Prot-Mitigation.json")
	weightInputs2 := weight_types.WeightInputReadFile("tempdata\\weightfind-sim-grid-Prot-Mitigation.json")
	weightInputs := slices.Concat(weightInputs1, weightInputs2)
	//weightInputs = util_collection.SliceSampleRandom(weightInputs, 256)
	//weightInputs = util_collection.SliceSampleRandom(weightInputs, 32)
	weightInputs = util_collection.SliceSampleRandom(weightInputs, 64)

	statList := model_factory.StatsForWeighting_strengthTank
	ratio := model_factory.SimPriority_mitigation

	ranking := weight_highs.RankingWeightsRatio30{}
	//ranking.AllPairs = true
	ranking.RandPairs = true
	ranking.RandPairCount = 3
	//ranking.UseMipCompare = false
	ranking.UseMipCompare = true
	ranking.Init(printer, 1000)
	ranking.SetRequiredStats(statList)
	ranking.SetTargetRatios(ratio)
	ranking.SupplyData(weightInputs)
	weightsFuture := ranking.RunSinglePassRaw()
	weightResult := weightsFuture.WaitForResultOrPanic()
	weights1 := weightResult.AsWeight1(weightInputs)
	newRatio := weightResult.NewRatio
	if weights1 != nil {
		tools.WritePawnString(*weights1, printer)
		acc := weightfind.EvaluateAccuracyBasic(weights1, ratio.SimTypes(), &ratio, weightInputs)
		printer.Printf("acc %f\n", acc)
		accSt := weightfind.EvaluateAccuracyStatisticalExtended(weights1, ratio.SimTypes(), &ratio, weightInputs)
		printer.Printf("acc st %f\n", accSt)

		newAcc := weightfind.EvaluateAccuracyBasic(weights1, newRatio.SimTypes(), newRatio, weightInputs)
		printer.Printf("new acc %f\n", newAcc)
		newAccSt := weightfind.EvaluateAccuracyStatisticalExtended(weights1, newRatio.SimTypes(), newRatio, weightInputs)
		printer.Printf("new acc st %f\n", newAccSt)
	} else {
		printer.Println("MISSING WEIGHT")
	}
}

func statWeightsSearchRatio(printer *util.PrintRecorder) {
	weightInputs1 := weight_types.WeightInputReadFile("tempdata\\weightfind-sim-real-Prot-Survival.json")
	weightInputs2 := weight_types.WeightInputReadFile("tempdata\\weightfind-sim-grid-Prot-Survival.json")
	weightInputs := slices.Concat(weightInputs1, weightInputs2)
	//weightInputs = util_collection.SliceSampleRandom(weightInputs, 64)

	statList := model_factory.StatsForWeighting_strengthTank
	ratio := model_factory.SimPriority_survival

	cancel := util_async.CancelSignal_Make()
	util_async.CancelOnKeyPress(cancel)

	statRange := weight_types.StatRangeFloat{Minimum: -1.0, Maximum: 1.0}

	simRangeMap := stats.SimTypeMap[weight_types.StatRangeFloat]{}
	simRangeMap.Put(stats.Sim_DPS, weight_types.StatRangeFloat{Minimum: 0.00, Maximum: 0.15})
	simRangeMap.Put(stats.Sim_DEATH, weight_types.StatRangeFloat{Minimum: 0.20, Maximum: 0.60})
	simRangeMap.Put(stats.Sim_TMI, weight_types.StatRangeFloat{Minimum: 0.00, Maximum: 0.25})
	simRangeMap.Put(stats.Sim_DTPS, weight_types.StatRangeFloat{Minimum: 0.20, Maximum: 0.50})

	ranking := weightfind.WeightSearcherRatio1{}
	ranking.AccuracyStatistical = true
	ranking.Init(statList, ratio.SimTypes())
	ranking.SupplyData(weightInputs)
	ranking.SetStatSimRanges(statRange, simRangeMap)
	weightResult := ranking.Run(cancel)
	weights1 := weightResult.AsWeight1(weightInputs)
	newRatio := weightResult.NewRatio
	if weights1 != nil {
		tools.WritePawnString(*weights1, printer)
		acc := weightfind.EvaluateAccuracyBasic(weights1, ratio.SimTypes(), &ratio, weightInputs)
		printer.Printf("acc %f\n", acc)
		accSt := weightfind.EvaluateAccuracyStatisticalExtended(weights1, ratio.SimTypes(), &ratio, weightInputs)
		printer.Printf("acc st %f\n", accSt)

		printer.Println(newRatio.String())
		newAcc := weightfind.EvaluateAccuracyBasic(weights1, newRatio.SimTypes(), newRatio, weightInputs)
		printer.Printf("new acc %f\n", newAcc)
		newAccSt := weightfind.EvaluateAccuracyStatisticalExtended(weights1, newRatio.SimTypes(), newRatio, weightInputs)
		printer.Printf("new acc st %f\n", newAccSt)
	} else {
		printer.Println("MISSING WEIGHT")
	}
}

func statWeightsSearch(printer *util.PrintRecorder) {
	// weightInputs, targetRatio := generateRatingsInputFromRealRandomSets(printer)
	targetRatio := model_factory.SimPriority_mitigation
	weightStats := model_factory.StatsForWeighting_strengthTank

	//inputDataGrid := readWeightInputFile("tempdata/sim-stats-compare-grid.json")
	//inputDataRandom := readWeightInputFile("tempdata/sim-stats-compare-rand.json")
	inputDataGrid := weight_types.WeightInputReadFile("tempdata\\weightfind-sim-grid-Prot-Mitigation.json")
	inputDataRandom := weight_types.WeightInputReadFile("tempdata\\weightfind-sim-real-Prot-Mitigation.json")
	// mixedInputData := slices.Concat(inputDataGrid, inputDataRandom)
	mixedInputData := slices.Concat(inputDataRandom, inputDataGrid)

	filteredInput := mixedInputData
	printer.Printf("mixedInputData size %d\n", len(filteredInput))

	//search := weightfind.WeightSearcher0{}
	//search.Init(weightStats, targetRatio, util.PrintRecorder_Nop())
	//search.Init(weightStats, targetRatio, printer)
	//search.SupplyData(mixedInputData)

	//search := weightfind.WeightSearcher2{}
	search := weightfind.WeightSearcher3{}
	search.AccuracyStatistical = true
	//search.Init(weightStats, targetRatio, printer)
	search.Init(weightStats, targetRatio)
	search.SupplyData(mixedInputData)
	search.SetRanges(-1.0, 10.0)

	weightResult := search.Run(util_async.CancelSignal_Make())
	printer.Printf("time = %s\n", weightResult.SolveTime)
	weight := *weightResult.AsWeight1(mixedInputData)
	tools.WritePawnString(weight, printer)
	printer.Printf("accuracy = %f\n", weightfind.EvaluateAccuracyBasic(&weight, targetRatio.SimTypes(), &targetRatio, mixedInputData))
	printer.Printf("accuracy stat = %f\n", weightfind.EvaluateAccuracyStatisticalExtended(&weight, targetRatio.SimTypes(), &targetRatio, mixedInputData))

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

func statWeightsSearchExtended(printer *util.PrintRecorder) {
	// weightInputs, targetRatio := generateRatingsInputFromRealRandomSets(printer)
	targetRatio := model_factory.SimPriority_mitigation
	weightStats := model_factory.StatsForWeighting_strengthTank

	//inputDataGrid := readWeightInputFile("tempdata/sim-stats-compare-grid.json")
	//inputDataRandom := readWeightInputFile("tempdata/sim-stats-compare-rand.json")
	inputDataGrid := weight_types.WeightInputReadFile("tempdata\\weightfind-sim-grid-Prot-Mitigation.json")
	inputDataRandom := weight_types.WeightInputReadFile("tempdata\\weightfind-sim-real-Prot-Mitigation.json")
	// mixedInputData := slices.Concat(inputDataGrid, inputDataRandom)
	mixedInputData := slices.Concat(inputDataRandom, inputDataGrid)

	filteredInput := mixedInputData
	printer.Printf("mixedInputData size %d\n", len(filteredInput))

	//search := weightfind.WeightSearcher0{}
	//search.Init(weightStats, targetRatio, util.PrintRecorder_Nop())
	//search.Init(weightStats, targetRatio, printer)
	//search.SupplyData(mixedInputData)

	//search := weightfind.WeightSearcher2{}

	cancel := util_async.CancelSignal_Make()
	util_async.CancelOnKeyPress(cancel)

	search := weightfind.WeightSearcherExtended1{}
	search.Init(weightStats, targetRatio)
	search.SupplyData(mixedInputData)
	search.SetRanges(-1.0, 1.0)

	weight2 := search.Run(cancel)
	//printer.Printf("time = %s\n", weightResult.SolveTime)
	weight1 := weight2.ConvertToWeight1()
	tools.WritePawnString(*weight1, printer)
	printer.Println(tools.FormatWeight2String(&weight2))
	printer.Printf("accuracy1 = %f\n", weightfind.EvaluateAccuracyBasic(weight1, targetRatio.SimTypes(), &targetRatio, mixedInputData))
	//printer.Printf("accuracy1 stat = %f\n", weightfind.EvaluateAccuracyStatistical(weight1, targetRatio.SimTypes(), &targetRatio, mixedInputData))
	printer.Printf("accuracy1 stat ex = %f\n", weightfind.EvaluateAccuracyStatisticalExtended(weight1, targetRatio.SimTypes(), &targetRatio, mixedInputData))
	printer.Printf("accuracy2 = %f\n", weightfind.EvaluateAccuracyBasic(&weight2, targetRatio.SimTypes(), &targetRatio, mixedInputData))
	//printer.Printf("accuracy2 stat = %f\n", weightfind.EvaluateAccuracyStatistical(&weight2, targetRatio.SimTypes(), &targetRatio, mixedInputData))
	printer.Printf("accuracy2 stat ex = %f\n", weightfind.EvaluateAccuracyStatisticalExtended(&weight2, targetRatio.SimTypes(), &targetRatio, mixedInputData))

	prep0 := weightfind.EvaluateAccuracyPrepared{}
	prep0.Init(mixedInputData, &targetRatio, false, false)
	printer.Printf("prep accuracy1 stat = %f\n", prep0.EvaluateWeight1(weight1))
	printer.Printf("prep accuracy2 stat = %f\n", prep0.EvaluateWeight2(&weight2))

	//prep1 := weightfind.EvaluateAccuracyPrepared{}
	//prep1.Init(mixedInputData, &targetRatio, true, false)
	//printer.Printf("prep accuracy1 stat = %f\n", prep1.EvaluateWeight1(weight1))
	//printer.Printf("prep accuracy2 stat = %f\n", prep1.EvaluateWeight2(&weight2))

	prep2 := weightfind.EvaluateAccuracyPrepared{}
	prep2.Init(mixedInputData, &targetRatio, true, true)
	printer.Printf("prep accuracy1 stat ex = %f\n", prep2.EvaluateWeight1(weight1))
	printer.Printf("prep accuracy2 stat ex = %f\n", prep2.EvaluateWeight2(&weight2))
}

func compareAccuracy(printer *util.PrintRecorder) {
	targetRatio := model_factory.SimPriority_mitigation
	//weightStats := model_factory.StatsForWeighting_strengthTank

	inputDataGrid := weight_types.WeightInputReadFile("tempdata\\weightfind-sim-grid-Prot-Mitigation.json")
	inputDataRandom := weight_types.WeightInputReadFile("tempdata\\weightfind-sim-real-Prot-Mitigation.json")
	testData := slices.Concat(inputDataRandom, inputDataGrid)
	//testData = testData[0:5]
	//testData = testData[102:107]

	weight1 := weight_types.Weight1Basic_Make()
	weight1.Put(stats.Stat_Strength, 1.0000)
	weight1.Put(stats.Stat_Stamina, 1.2309)
	weight1.Put(stats.Stat_Crit, 0.1167)
	weight1.Put(stats.Stat_Haste, 0.3614)
	weight1.Put(stats.Stat_Expertise, 0.0054)
	weight1.Put(stats.Stat_Mastery, 0.5866)
	weight1.Put(stats.Stat_Dodge, 0.0824)
	weight1.Put(stats.Stat_Parry, 0.0532)

	//printer.Printf("accuracy1 = %f\n", weightfind.EvaluateAccuracy(&weight1, targetRatio.SimTypes(), &targetRatio, testData))
	//printer.Printf("accuracy1 stat = %f\n", weightfind.EvaluateAccuracyStatistical(&weight1, targetRatio.SimTypes(), &targetRatio, testData))
	printer.Printf("accuracy1 stat ex = %f\n", weightfind.EvaluateAccuracyStatisticalExtended(&weight1, targetRatio.SimTypes(), &targetRatio, testData))

	//prep0 := weightfind.EvaluateAccuracyPrepared{}
	//prep0.Init(testData, &targetRatio, false, false)
	//printer.Printf("prep accuracy1 = %f\n", prep0.EvaluateWeight1(&weight1))
	//
	//prep1 := weightfind.EvaluateAccuracyPrepared{}
	//prep1.Init(testData, &targetRatio, true, false)
	//printer.Printf("prep accuracy1 stat = %f\n", prep1.EvaluateWeight1(&weight1))

	prep2 := weightfind.EvaluateAccuracyPrepared{}
	prep2.Init(testData, &targetRatio, true, true)
	printer.Printf("prep accuracy1 stat ex = %f\n", prep2.EvaluateWeight1(&weight1))
}

func statWeightsGridIntoRanking(printer *util.PrintRecorder) {
	targetRatio := model_factory.SimPriority_mitigation
	requiredStats := model_factory.StatsForWeighting_strengthTank
	simTypes := targetRatio.SimTypes()

	inputDataGrid := weight_types.WeightInputReadFile("tempdata/sim-stats-compare-grid.json")
	inputDataRandom := weight_types.WeightInputReadFile("tempdata/sim-stats-compare-rand.json")
	mixedInputData := slices.Concat(inputDataRandom, inputDataGrid)

	var weights1 weight_types.Weight1Basic
	if false {
		grid := weight_highs.GridStatWeightProcess{}
		grid.Init(printer, 3000)
		grid.SetRequiredStats(requiredStats)
		grid.SetTargetRatios(targetRatio)
		grid.SupplyData(inputDataGrid)
		weightResult := grid.Run().WaitForResultOrPanic()
		weights1 = *weightResult.AsWeight1(mixedInputData)
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
	weights2 := *weightsResult2.AsWeight1(mixedInputData)

	tools.WritePawnString(weights1, printer)
	printer.Printf("accuracy_initial = %f\n", weightfind.EvaluateAccuracyBasic(&weights1, simTypes, &targetRatio, mixedInputData))

	tools.WritePawnString(weights2, printer)
	printer.Printf("accuracy_algo = %f\n", weightfind.EvaluateAccuracyBasic(&weights2, simTypes, &targetRatio, mixedInputData))

	weights3, _ := weightfind.WeightTweakerWithLogging(weights2, requiredStats, &targetRatio, mixedInputData, util.PrintRecorder_Nop())

	tools.WritePawnString(weights3, printer)
	printer.Printf("accuracy_tweak = %f\n", weightfind.EvaluateAccuracyBasic(&weights3, simTypes, &targetRatio, mixedInputData))

	// ( Pawn: v1: "Protection WoWSims Weights": Class=Paladin,Strength=1.0000000000,Stamina=0.4805050000,CritRating=0.6462260000,HasteRating=0.8598560000,ExpertiseRating=0.6679750000,MasteryRating=1.9405810000,DodgeRating=0.6518220000,ParryRating=0.6243300000, )
	// accuracy1 = 92.635522
	// ( Pawn: v1: "Protection WoWSims Weights": Class=Paladin,Strength=1.0000000000,Stamina=-0.0896998019,CritRating=0.3760289134,HasteRating=0.4969834753,ExpertiseRating=0.3863096443,MasteryRating=1.1063898778,DodgeRating=0.3787533903,ParryRating=0.3581785849, )
	// accuracy2 = 91.501292
	// Duration = 2h2m1.7439799s

}

func statWeightsFitting(printer *util.PrintRecorder) {
	// generateRatingsInputFromRealRandomSets(printer)

	bytes, err := os.ReadFile("tempdata/sim-stats-compare-rand.json")
	// bytes, err := os.ReadFile("tempdata/sim-stats-input-data.json")
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
	//bytes, err := os.ReadFile("tempdata/sim-stats-compare-rand.json")
	// bytes, err := os.ReadFile("tempdata/sim-stats-input-data.json")
	//weightInputs := readWeightInputFile("tempdata/weightfind-sim-real-Prot-Heal.json")
	weightInputs := weight_types.WeightInputReadFile("tempdata/sim-stats-compare-rand.json")

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
	weightInputs := weight_types.WeightInputReadFile("tempdata/weightfind-sim-real-Prot-Heal.json")

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
	weightInputs := weight_types.WeightInputReadFile("tempdata/sim-stats-compare-rand.json")
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
	weight3 := weightResult.Weight

	tools.WriteWeightString(weight3, printer)
	printer.Printf("weight3 isempty = %v\n", weight3.IsEmpty())
	printer.Printf("weight2 isempty = %v\n", weight3.ConvertToWeight2(weightInputs).IsEmpty())
	printer.Printf("weight1 isempty = %v\n", weight3.ConvertToWeight2(weightInputs).ConvertToWeight1().IsEmpty())

	acc3 := weightfind.EvaluateAccuracyBasic(weight3, simTypes, &targetRatio, weightInputs)
	acc2 := weightfind.EvaluateAccuracyBasic(weight3.ConvertToWeight2(weightInputs), simTypes, &targetRatio, weightInputs)
	acc1 := weightfind.EvaluateAccuracyBasic(weight3.ConvertToWeight2(weightInputs).ConvertToWeight1(), simTypes, &targetRatio, weightInputs)
	printer.Printf("weight3 acc = %v\n", acc3)
	printer.Printf("weight2 acc = %v\n", acc2)
	printer.Printf("weight1 acc = %v\n", acc1)

}

func statWeightsFitting3eachProper(printer *util.PrintRecorder) {
	weightInputs := weight_types.WeightInputReadFile("tempdata/weightfind-sim-real-Prot-Mitigation-NoSet.json")
	//weightInputs = weightInputs[0:30]
	simTypes := model_factory.SimPriority_heal.SimTypes()
	statTypes := model_factory.StatsForWeighting_strengthTank
	targetRatio := model_factory.SimPriority_mitigation

	notes := make([]string, 0)

	for segmentCount := 2; segmentCount <= 10; segmentCount++ {
		//fitting := fitting3.FittingEachStatWeightProcess3{}
		fitting := fitting4.FittingEachStatWeightProcess4{}
		fitting.Init(segmentCount, printer, 1000)
		fitting.SetRequiredStats(statTypes, simTypes)
		fitting.SetTargetRatios(targetRatio)
		fitting.SupplyData(weightInputs)
		weightResult := fitting.Run(util_async.CancelSignal_Make())
		weight3 := weightResult.Weight

		tools.WriteWeightString(weight3, printer)
		printer.Printf("weight3 isempty = %v\n", weight3.IsEmpty())
		printer.Printf("weight2 isempty = %v\n", weight3.ConvertToWeight2(weightInputs).IsEmpty())
		printer.Printf("weight1 isempty = %v\n", weight3.ConvertToWeight2(weightInputs).ConvertToWeight1().IsEmpty())

		acc3 := weightfind.EvaluateAccuracyBasic(weight3, simTypes, &targetRatio, weightInputs)
		acc2 := weightfind.EvaluateAccuracyBasic(weight3.ConvertToWeight2(weightInputs), simTypes, &targetRatio, weightInputs)
		acc1 := weightfind.EvaluateAccuracyBasic(weight3.ConvertToWeight2(weightInputs).ConvertToWeight1(), simTypes, &targetRatio, weightInputs)
		printer.Printf("weight3 acc = %v\n", acc3)
		printer.Printf("weight2 acc = %v\n", acc2)
		printer.Printf("weight1 acc = %v\n", acc1)

		note := fmt.Sprintf("fit segments %d acc3=%f acc1=%f", segmentCount, acc3, acc1)
		notes = append(notes, note)
	}

	for _, note := range notes {
		printer.Println(note)
	}

	// STAT SPLIT
	//fit segments 2 acc3=84.740787 acc1=83.895873
	//fit segments 3 acc3=86.428130 acc1=85.875156
	//fit segments 4 acc3=83.264534 acc1=80.220693
	//fit segments 5 acc3=83.770047 acc1=78.272867
	//fit segments 6 acc3=75.410591 acc1=77.115319
	//fit segments 7 acc3=81.844019 acc1=84.167117
	//fit segments 8 acc3=80.955232 acc1=76.562621
	//fit segments 9 acc3=74.724893 acc1=76.934029
	//fit segments 10 acc3=79.255472 acc1=66.088123

	// DATA SPLIT
	//fit segments 2 acc3=86.641152 acc1=86.757321
	//fit segments 3 acc3=84.574122 acc1=86.507599
	//fit segments 4 acc3=85.427313 acc1=82.234468
	//fit segments 5 acc3=81.302359 acc1=82.638161
	//fit segments 6 acc3=81.527522 acc1=85.675103
	//fit segments 7 acc3=80.728965 acc1=81.849814
	//fit segments 8 acc3=81.356442 acc1=85.007064
	//fit segments 9 acc3=77.016810 acc1=86.579066
	//fit segments 10 acc3=77.603448 acc1=85.311144

}

func statWeightsFitting1eachProper(printer *util.PrintRecorder) {
	//bytes, err := os.ReadFile("")
	//weightInputs := readWeightInputFile("tempdata/sim-stats-compare-rand.json")
	//weightInputs := readWeightInputFile("tempdata/weightfind-sim-real-Prot-Heal.json")
	weightInputs := weight_types.WeightInputReadFile("tempdata/weightfind-sim-real-Prot-Mitigation-WithSet.json")
	//weightInputs = weightInputs[0:10]
	statTypes := model_factory.StatsForWeighting_strengthTank
	targetRatio := model_factory.SimPriority_survival
	simTypes := targetRatio.SimTypes()

	fitting := fitting1.FittingEachStatWeightProcess{}
	fitting.Init(printer, 1000)
	fitting.SetRequiredStats(statTypes, simTypes)
	fitting.SetOnlyComputeSingleSegmentEach(true)
	fitting.SupplyData(weightInputs)
	weightResult := fitting.Run(util_async.CancelSignal_Make())
	weight3 := weightResult.Weight

	tools.WriteWeightString(weight3, printer)
	printer.Printf("weight3 isempty = %v\n", weight3.IsEmpty())
	printer.Printf("weight2 isempty = %v\n", weight3.ConvertToWeight2(weightInputs).IsEmpty())
	printer.Printf("weight1 isempty = %v\n", weight3.ConvertToWeight2(weightInputs).ConvertToWeight1().IsEmpty())

	acc3 := weightfind.EvaluateAccuracyBasic(weight3, simTypes, &targetRatio, weightInputs)
	acc2 := weightfind.EvaluateAccuracyBasic(weight3.ConvertToWeight2(weightInputs), simTypes, &targetRatio, weightInputs)
	acc1 := weightfind.EvaluateAccuracyBasic(weight3.ConvertToWeight2(weightInputs).ConvertToWeight1(), simTypes, &targetRatio, weightInputs)
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
	weightInputs := weight_types.WeightInputReadFile("tempdata/sim-stats-compare-rand.json")
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
	bytes, err := os.ReadFile("tempdata/sim-stats-input-data2.json")
	// bytes, err := os.ReadFile("tempdata/sim-stats-input-data.json")
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
	weight3 := weightResult.Weight
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

//goland:noinspection ALL
func statWeightsGrid1Orig(printer *util.PrintRecorder) {
	// inputData, targetRatio := generateRatingsInputFromArtificalStatOverrides(printer)
	// weight_types.WeightInputWriteFile(inputData, "tempdata/sim-stats-input-grid.json" )

	inputDataGrid := weight_types.WeightInputReadFile("tempdata/weightfind-sim-grid-Prot-Mitigation-NoSet.json")
	inputDataRandom := weight_types.WeightInputReadFile("tempdata/weightfind-sim-real-Prot-Mitigation-NoSet.json")
	//inputData := SliceSampleRandom(inputDataFull, 30)
	inputData := inputDataGrid

	targetRatio := model_factory.SimPriority_mitigation
	//targetRatio := model_factory.SimPriority_dps
	requiredStats := model_factory.StatsForWeighting_strengthTank
	simTypes := targetRatio.SimTypes()

	runOne := func(a int, b int, c int, d int, e int, label string) {
		//process := weight_highs.GridStatWeightProcess{}
		//process.CHECKRANGE = inc1

		process := weight_highs.GridStatWeightProcess1B{}
		process.OUTLIER = a
		process.SCALEMODE = b
		process.ROUNDMODE = c
		process.CALCMODE = d
		process.RATIO = e

		//process.Init(printer, 2000)
		process.Init(util.PrintRecorder_Nop(), 90)
		process.SetRequiredStats(requiredStats)
		process.SetTargetRatios(targetRatio)
		process.SupplyData(inputData)
		weightsResult := process.Run().WaitForResultOrNilValue()
		weights1 := *weightsResult.AsWeight1(inputData)
		tools.WritePawnString(weights1, printer)

		acc := weightfind.EvaluateAccuracyBasic(&weights1, simTypes, &targetRatio, inputDataGrid)
		acc2 := weightfind.EvaluateAccuracyBasic(&weights1, simTypes, &targetRatio, inputDataRandom)
		acc3 := weightfind.EvaluateAccuracyBasic(&weights1, simTypes, &targetRatio, slices.Concat(inputDataGrid, inputDataRandom))
		acc4 := weightfind.EvaluateAccuracyStatisticalExtended(&weights1, simTypes, &targetRatio, slices.Concat(inputDataGrid, inputDataRandom))
		printer.Printf("accuracy %s: grid data = %f, rand data = %f, data mix = %f, stat mix = %f\n", label, acc, acc2, acc3, acc4)
	}

	//runOne(true, true, true, 1, 3, "select")

	type optParam struct {
		a     int
		b     int
		c     int
		d     int
		e     int
		label string
	}
	optList := make([]optParam, 0)
	for a := range 5 {
		for b := range 4 {
			for c := range 3 {
				for d := range 4 {
					for e := range 4 {
						label := fmt.Sprintf("GRID1 %d %d %d %d %d", a, b, c, d, e)
						optList = append(optList, optParam{a, b, c, d, e, label})
					}
				}
			}
		}
	}
	//for a := range 2 {
	//	label := fmt.Sprintf("GRID1 %d", a)
	//	optList = append(optList, optParam{a, label})
	//}

	util_async.ForEach_Slice(5, optList, func(o *optParam) {
		runOne(o.a, o.b, o.c, o.d, o.e, o.label)
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
	// weight_types.WeightInputWriteFile(inputData, "tempdata/sim-stats-input-grid.json" )

	inputDataGrid := weight_types.WeightInputReadFile("tempdata/weightfind-sim-grid-Prot-Mitigation-NoSet.json")
	inputDataRandom := weight_types.WeightInputReadFile("tempdata/weightfind-sim-real-Prot-Mitigation-NoSet.json")
	//inputData := SliceSampleRandom(inputDataFull, 30)
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
		weights1 := *weightsResult.AsWeight1(inputData)
		tools.WritePawnString(weights1, printer)

		acc := weightfind.EvaluateAccuracyBasic(&weights1, simTypes, &targetRatio, inputDataGrid)
		acc2 := weightfind.EvaluateAccuracyBasic(&weights1, simTypes, &targetRatio, inputDataRandom)
		acc3 := weightfind.EvaluateAccuracyBasic(&weights1, simTypes, &targetRatio, slices.Concat(inputDataGrid, inputDataRandom))
		acc4 := weightfind.EvaluateAccuracyStatisticalExtended(&weights1, simTypes, &targetRatio, slices.Concat(inputDataGrid, inputDataRandom))
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
	// weight_types.WeightInputWriteFile(inputData, "tempdata/sim-stats-input-grid.json" )

	inputDataGrid := weight_types.WeightInputReadFile("tempdata/weightfind-sim-grid-Prot-Mitigation-NoSet.json")
	inputDataRandom := weight_types.WeightInputReadFile("tempdata/weightfind-sim-real-Prot-Mitigation-NoSet.json")
	//inputData := SliceSampleRandom(inputDataFull, 30)
	inputData := inputDataGrid

	targetRatio := model_factory.SimPriority_mitigation
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
		weights1 := *weightResult.AsWeight1(inputData)

		acc := weightfind.EvaluateAccuracyBasic(&weights1, simTypes, &targetRatio, inputDataGrid)
		acc2 := weightfind.EvaluateAccuracyBasic(&weights1, simTypes, &targetRatio, inputDataRandom)
		acc3 := weightfind.EvaluateAccuracyBasic(&weights1, simTypes, &targetRatio, slices.Concat(inputDataGrid, inputDataRandom))
		acc4 := weightfind.EvaluateAccuracyStatisticalExtended(&weights1, simTypes, &targetRatio, slices.Concat(inputDataGrid, inputDataRandom))
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
