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
	modelEquipOnly := gear_model.Model_PallyProtMitigation_WithSet()
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
	process.SetRequiredStats(gear_model.StatsForWeighting_strengthTank)
	process.SetTargetRatios(targetRatio)
	process.SetBaseline(simBase)
	for _, data := range inputData {
		process.AddSimData(data.IncrementStat, uint32(data.IncrementValue), data.SimResult)
	}
	process.Run(nil)
}

// oldish code, may sometimes want to mix basic ratings??
func relativeRatingsCompromise(printer *util.PrintRecorder) {
	modelMitiNoSet := gear_model.Model_PallyProtMitigation_NoSet()
	gearMitiNoSet := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(files.GearFileProtMitigationNoSet), &modelMitiNoSet, setup.MissingEnchant_Panic, printer)
	itemSetMitiNoSet := items.FullItemSet_FromMap(gearMitiNoSet)

	modelDps := gear_model.Model_PallyProtDps()
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

func generateRatingsInputFromArtificialStatOverrides_ForBasic(currentItemSet items.FullItemSet, printer *util.PrintRecorder, simSpeed simulate.WowSim_RunSize, speedUp int, requiredStats []stats.StatType, spec stats.SpecType, goal stats.OptimiseGoal, fight stats.WowSim_Fight, profession gear_model.ProfessionInfo, tracker *util.TrackProgress) ([]basicStatInput, stats.SimData) {
	var incrementValue int32 = 250

	initialBaseStats := weightfind.InitialBonusStatMap_fixRanges(printer, currentItemSet, incrementValue)
	tracker.RunOuterTracking(len(requiredStats) + 1)
	defer tracker.SetDone()

	simBase := simulate.WowSim_Execute_SpecifyAll(simSpeed, speedUp, spec, goal, fight, profession, currentItemSet.Items(), nil, tracker.NewChild())

	inputList := util_async.Map_SliceToSlice(len(requiredStats), requiredStats, func(incStat *stats.StatType) basicStatInput {
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
	comp.SetRequiredStats(gear_model.StatsForWeighting_strengthTank)
	comp.SetTargetRatios(gear_model.SimPriority_generalMiti)
	comp.SetMinimumIncludeRate(1.0)
	comp.SupplyData(filteredInput)
	weights2 := comp.Run(nil, 3000).WaitForResultOrPanic()
	weights1 := weights2.ConvertToWeight1()
	tools.WritePawnString(weights1, printer)
}

func statWeightsRanking(printer *util.PrintRecorder) {
	weightInputs := readWeightInputFile("sim-stats-compare-rand.json")

	comp := weight_highs.RankingSeparatedWeights{}

	comp.Init(printer, 3000)
	comp.SetRequiredStats(gear_model.StatsForWeighting_strengthTank, gear_model.SimPriority_generalMiti.SimTypes())
	comp.SetTargetRatios(gear_model.SimPriority_generalMiti)
	comp.SupplyData(weightInputs)
	weights2 := comp.Run(nil).WaitForResultOrPanic()
	weights1 := weights2.ConvertToWeight1()
	tools.WritePawnString(weights1, printer)
}

func statWeightsCustom(printer *util.PrintRecorder) {
	// weightInputs, targetRatio := generateRatingsInputFromRealRandomSets(printer)
	targetRatio := gear_model.SimPriority_generalMiti
	weightStats := gear_model.StatsForWeighting_strengthTank

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

	//var interest = []float64{ // accuracy = 92.633057
	//	5.39,
	//	8.659509862234,
	//	3.433015580687,
	//	8.603761881869,
	//	-0.001056252428,
	//	8.802300380181,
	//	5.369767704147,
	//	3.466234195963}
	//interestWR := stathighs.WeightResult_Of(interest, weightStats)
	//printer.Printf("interest = %f\n", weightfind.EvaluateAccuracyRanged(interestWR, targetRatio, mixedInputData))

	search := weightfind.WeightSearcher2{}
	//search.Init(weightStats, targetRatio, printer)
	search.Init(weightStats, targetRatio, nil)
	search.SupplyData(inputDataGrid)
	search.SetRanges(-1.0, 10.0)

	sw := util.StopwatchMakeStarted()
	weight := search.Run(util_async.CancelSignal_Make())
	printer.Printf("time = %s\n", sw.Elapsed().String())
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
	targetRatio := gear_model.SimPriority_generalMiti
	requiredStats := gear_model.StatsForWeighting_strengthTank
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
		weights1 = grid.Run(nil).WaitForResultOrPanic()
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
		weights1 = weight_types.Weight1Basic_Make(targetRatio)
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
	weights2 := ranking.RunSinglePassFromExternal(weights1, nil).WaitForResultOrPanic()
	//weights2 := ranking.Run(&weights1)
	//weights2 := ranking.Run(nil)

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

	simTypes := gear_model.SimPriority_heal.SimTypes()
	util_async.ForEach_Slice(8, simTypes, func(sim *stats.SimType) {
		for _, statType := range gear_model.StatsForWeighting_strengthTank {
			task(*sim, statType)
		}
	})
}

func statWeightsFitting2eachProper(printer *util.PrintRecorder) {
	//bytes, err := os.ReadFile("")
	weightInputs := readWeightInputFile("sim-stats-compare-rand.json")
	//weightInputs := readWeightInputFile("tempdata/weightfind-sim-real-Prot-Heal.json")
	weightInputs = weightInputs[0:30]
	simTypes := gear_model.SimPriority_heal.SimTypes()
	statTypes := gear_model.StatsForWeighting_strengthTank
	targetRatio := gear_model.SimPriority_heal

	fitting := fitting2.FittingEachStatWeightProcess2{}
	fitting.Init(3, printer, 1000)
	fitting.SetRequiredStats(statTypes, simTypes)
	fitting.SetTargetRatios(targetRatio)
	fitting.SupplyData(weightInputs)
	weight3 := fitting.Run(util.StopwatchMakeStopped(), util_async.CancelSignal_Make())

	tools.WriteWeight3String(weight3, printer)
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
		sim := util_weight.FittingSimDetail{
			Average:   scaleSim.Apply(input.SimResult.Get(checkSimType)),
			Min:       scaleSim.Apply(detail.Min),
			Max:       scaleSim.Apply(detail.Max),
			StdDev:    scaleSim.Scale * detail.StdDev,
			HasDetail: true,
		}
		sim.FlipMinMaxAsNeeded()
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

	weight3 := fitting.Run(util.StopwatchMakeStopped(), util_async.CancelSignal_Make())
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

func statWeightsGrid(printer *util.PrintRecorder) {
	// inputData, targetRatio := generateRatingsInputFromArtificalStatOverrides(printer)
	// writeWeightInputsToFile(inputData, "sim-stats-input-grid.json" )

	inputDataFull := readWeightInputFile("tempdata/weightfind-sim-grid-Prot-Damage.json")
	inputDataRandom := readWeightInputFile("tempdata/weightfind-sim-real-Prot-Damage.json")
	//inputData := takeDataSample_Random(inputDataFull, 30)
	inputData := inputDataFull

	//targetRatio := gear_model.SimRatio_generalMiti
	targetRatio := gear_model.SimPriority_dps
	requiredStats := gear_model.StatsForWeighting_strengthTank
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
		weights2 := process.Run(nil).WaitForResultOrNilValue()
		weights1 := weights2.ConvertToWeight1()
		tools.WritePawnString(weights1, printer)

		acc := weightfind.EvaluateAccuracy(&weights1, simTypes, &targetRatio, inputDataFull)
		acc2 := weightfind.EvaluateAccuracy(&weights1, simTypes, &targetRatio, inputDataRandom)
		acc3 := weightfind.EvaluateAccuracy(&weights1, simTypes, &targetRatio, slices.Concat(inputDataFull, inputDataRandom))
		acc4 := weightfind.EvaluateAccuracyStatistical(&weights1, simTypes, &targetRatio, slices.Concat(inputDataFull, inputDataRandom))
		printer.Printf("accuracy %s: grid data = %f, rand data = %f, data mix = %f, stat mix = %f\n", label, acc, acc2, acc3, acc4)
	}

	//runOne(true, true, true, 1, 3, "select")

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

//goland:noinspection GoBoolExpressions
func statWeights_CompareAlgorithms() {
	printer := util.PrintRecorder_CreateLogFileNamed(files.LogOutputPath, "statWeights_CompareAlgorithms")

	//targetRatio := gear_model.SimRatio_generalMiti
	targetRatio := gear_model.SimPriority_heal
	requiredStats := gear_model.StatsForWeighting_strengthTank
	requiredSims := targetRatio.SimTypes()

	//simSpeed := simulate.RunSize_Common
	//gearFile := files.GearFileProtMitigationWithSet
	//gearModel := gear_model.Model_PallyProtMitigation_WithSet()
	//currentEquip := setup.OptionsSetup_ExactEquippedOnly(loaders.GearFileReader_Read(gearFile), &gearModel, setup.MissingEnchant_Fix, printer)
	//currentItemSet := items.FullItemSet_FromMap(currentEquip)
	//
	//trackProcess := util.TrackProgress_Start()
	//trackProcess.RunOuterTracking(4)
	//inputDataBasic, basicSimBase := generateRatingsInputFromArtificialStatOverrides_ForBasic(currentItemSet, printer, simSpeed, 1, requiredStats, gearModel.Spec, gearModel.Goal, gearModel.SimulateAs, gearModel.Professions, trackProcess.NewChild())
	//inputDataGrid := weightfind.SimulateSteppedStatChangesForGrid(currentItemSet, printer, simSpeed, 1, requiredStats, gearModel.Spec, gearModel.Goal, gearModel.SimulateAs, gearModel.Professions, trackProcess.NewChild())
	//inputDataRandomUnsafe := weightfind.SimulateRealRandomSets(gearFile, substituteItemsMiti, &gearModel, 256, simSpeed, false, printer, trackProcess.NewChild())
	//inputDataRandomSafe := weightfind.SimulateRealRandomSets(gearFile, substituteItemsMiti, &gearModel, 256, simSpeed, true, printer, trackProcess.NewChild())
	//inputDataRandom := slices.Concat(inputDataRandomUnsafe, inputDataRandomSafe)
	//trackProcess.SetDone()
	//
	//writeWeightInputsToFile(inputDataGrid, "sim-stats-compare-grid.json")
	//writeWeightInputsToFile(inputDataRandomUnsafe, "sim-stats-compare-rand-unsafe.json")
	//writeWeightInputsToFile(inputDataRandomSafe, "sim-stats-compare-rand-safe.json")
	//writeWeightInputsToFile(inputDataRandom, "sim-stats-compare-rand.json")
	//writeWeightBasicInputsToFile(inputDataBasic, basicSimBase, "sim-stats-compare-basic.json")

	inputDataBasic, basicSimBase := readWeightBasicInputsFile("sim-stats-compare-basic.json")
	inputDataGrid := readWeightInputFile("tempdata/weightfind-sim-grid-Prot-Heal.json")
	inputDataRandom := readWeightInputFile("tempdata/weightfind-sim-real-Prot-Heal.json")
	//inputDataGrid := readWeightInputFile("sim-stats-compare-grid.json")
	//inputDataRandom := readWeightInputFile("sim-stats-compare-rand.json")
	mixedInputDataFull := slices.Concat(inputDataGrid, inputDataRandom)

	sampleSize := 50
	inputDataGrid = takeDataSample_Random(inputDataGrid, sampleSize)
	inputDataRandom = takeDataSample_Random(inputDataRandom, sampleSize)
	mixedInputData := takeDataSample_Random(mixedInputDataFull, sampleSize)
	//inputDataGrid = takeDataSample_Random_Seed(inputDataGrid, sampleSize, 1234)
	//inputDataRandom = takeDataSample_Random_Seed(inputDataRandom, sampleSize, 1234)
	//mixedInputData := takeDataSample_Random_Seed(mixedInputDataFull, sampleSize, 1234)
	//mixedInputData := mixedInputDataFull

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
	weightsMidRange := weight_types.Weight1Basic_Make(targetRatio)
	weightsMidRange.Put(stats.Stat_Strength, 1.0000)
	weightsMidRange.Put(stats.Stat_Stamina, 1.2309)
	weightsMidRange.Put(stats.Stat_Crit, 0.1167)
	weightsMidRange.Put(stats.Stat_Haste, 0.3614)
	weightsMidRange.Put(stats.Stat_Expertise, 0.0054)
	weightsMidRange.Put(stats.Stat_Mastery, 0.5866)
	weightsMidRange.Put(stats.Stat_Dodge, 0.0824)
	weightsMidRange.Put(stats.Stat_Parry, 0.0532)

	resultsByAlgorithm := util_collection.MapConcurrent[string, weight_types.Weight1Basic]{}
	resultsByAlgorithm2 := util_collection.MapConcurrent[string, weight_types.Weight2Extended]{}
	resultsByAlgorithm3 := util_collection.MapConcurrent[string, weight_types.Weight3ExtendedRanged]{}
	timesByAlgorithm := util_collection.MapConcurrent[string, time.Duration]{}

	cancel := util_async.CancelSignal_Make()
	tasks := make([]func(), 0)

	reportOnTweakedVersions := false
	standardTimeout := 1000
	shortTimeout := 500

	runBasic := false
	runFormulaVariants := false // best is about 87%, moderate time
	runFitting1 := false        // slow, low 90%
	runFitting2 := true

	runGrid1Original := true
	runGrid1Variants := false
	runGrid1VariantsFewer := true
	runGrid1C := false
	runGrid2 := true

	runRankingOlder := true
	runRanking3aPreferred := false // broken
	runRanking3aVariants := false  // broken
	runRanking3bVariants := false
	runRanking3bPreferred := true
	runRanking4 := false // still a little slow, midrange 94% etc
	runRanking5 := false // excellent but slow

	runSearches := true
	runSearch0 := false

	runRankingSep := true
	runFormula2 := true

	if runBasic {
		tasks = append(tasks, func() {
			printer.Println("################# BASIC ###################")
			stopwatch := util.StopwatchMakeStopped()
			basic := weight_highs.BasicStatWeightProcess{}
			basic.Init(printer)
			basic.SetRequiredStats(requiredStats)
			basic.SetTargetRatios(targetRatio)
			basic.SetBaseline(basicSimBase)
			for _, data := range inputDataBasic {
				basic.AddSimData(data.IncrementStat, uint32(data.IncrementValue), data.SimResult)
			}
			weightFuture := basic.Run(stopwatch)
			util_async.ChainCancel(cancel, weightFuture)
			resultsByAlgorithm.Put("basic", weightFuture.WaitForResultOrNilValue())
			timesByAlgorithm.Put("basic", stopwatch.Elapsed())
			printer.Println("///////////////// BASIC /////////////////")
		})

		tasks = append(tasks, func() {
			printer.Println("################# FORMULA ###################")
			stopwatch := util.StopwatchMakeStopped()
			comp := weight_highs.FormulaStatWeightProcess{}
			comp.Init(printer)
			comp.SetRequiredStats(requiredStats)
			comp.SetTargetRatios(targetRatio)
			comp.SetMinimumIncludeRate(1)
			comp.SupplyData(slices.Clone(inputDataRandom))
			weightFuture := comp.Run(stopwatch, standardTimeout)
			util_async.ChainCancel(cancel, weightFuture)
			resultsByAlgorithm2.Put("form", weightFuture.WaitForResultOrNilValue())
			timesByAlgorithm.Put("form", stopwatch.Elapsed())
			printer.Println("///////////////// FORMULA /////////////////")
		})
	}

	if runFormulaVariants {
		for BLEND := range 6 {
			tasks = append(tasks, func() {
				printer.Println("################# FORMULA ###################")
				stopwatch := util.StopwatchMakeStopped()
				comp := weight_highs.FormulaStatWeightProcess{}
				comp.BLEND = BLEND
				comp.Init(printer)
				comp.SetRequiredStats(requiredStats)
				comp.SetTargetRatios(targetRatio)
				comp.SetMinimumIncludeRate(0.7)
				comp.SupplyData(slices.Clone(inputDataRandom))
				label := fmt.Sprintf("form-blend%d-inc70", BLEND)
				futureResult := comp.Run(stopwatch, standardTimeout)

				util_async.ChainCancel(cancel, futureResult)
				futureResult.WaitForResultThenRun(func(result weight_types.Weight2Extended) {
					resultsByAlgorithm2.Put(label, result)
					timesByAlgorithm.Put(label, stopwatch.Elapsed())
				}, func() {
					resultsByAlgorithm2.Put(label, weight_types.Weight2Extended{})
				})
				printer.Println("///////////////// FORMULA /////////////////")
			})
			tasks = append(tasks, func() {
				printer.Println("################# FORMULA ###################")
				stopwatch := util.StopwatchMakeStopped()
				comp := weight_highs.FormulaStatWeightProcess{}
				comp.BLEND = BLEND
				comp.Init(printer)
				comp.SetRequiredStats(requiredStats)
				comp.SetTargetRatios(targetRatio)
				comp.SetMinimumIncludeRate(1)
				comp.SupplyData(slices.Clone(inputDataRandom))
				label := fmt.Sprintf("form-blend%d-inc100", BLEND)
				futureResult := comp.Run(stopwatch, standardTimeout)

				util_async.ChainCancel(cancel, futureResult)
				futureResult.WaitForResultThenRun(func(result weight_types.Weight2Extended) {
					resultsByAlgorithm2.Put(label, result)
					timesByAlgorithm.Put(label, stopwatch.Elapsed())
				}, func() {
					resultsByAlgorithm.Put(label, weight_types.Weight1Basic{})
				})
				printer.Println("///////////////// FORMULA /////////////////")
			})
		}
	}

	if runFitting1 {
		tasks = append(tasks, func() {
			printer.Println("################# FITTING ###################")
			stopwatch := util.StopwatchMakeStopped()
			fitting := fitting1.FittingEachStatWeightProcess{}
			fitting.Init(printer, shortTimeout)
			fitting.SetRequiredStats(requiredStats, requiredSims)
			fitting.SetOnlyComputeSingleSegmentEach(true)
			fitting.SupplyData(inputDataRandom)
			resultsByAlgorithm3.Put("fitting", fitting.Run(stopwatch, cancel))
			timesByAlgorithm.Put("fitting", stopwatch.Elapsed())
			printer.Println("///////////////// FITTING /////////////////")
		})
	}
	if runFitting2 {
		tasks = append(tasks, func() {
			printer.Println("################# FITTING2 ###################")
			stopwatch := util.StopwatchMakeStopped()
			fitting := fitting2.FittingEachStatWeightProcess2{}
			fitting.Init(3, printer, shortTimeout)
			fitting.SetRequiredStats(requiredStats, requiredSims)
			fitting.SetTargetRatios(targetRatio)
			fitting.SupplyData(inputDataRandom)
			weight3 := fitting.Run(stopwatch, cancel)
			resultsByAlgorithm3.Put("fitting2", weight3)
			timesByAlgorithm.Put("fitting2", stopwatch.Elapsed())
			printer.Println("///////////////// FITTING2 /////////////////")
		})
	}

	if runGrid1Original {
		tasks = append(tasks, func() {
			printer.Println("################# GRID1 ###################")
			stopwatch := util.StopwatchMakeStopped()
			grid1 := weight_highs.GridStatWeightProcess{}
			grid1.Init(printer, standardTimeout)
			grid1.SetRequiredStats(requiredStats)
			grid1.SetTargetRatios(targetRatio)
			grid1.SupplyData(slices.Clone(inputDataGrid))
			weightFuture := grid1.Run(stopwatch)
			util_async.ChainCancel(cancel, weightFuture)
			resultsByAlgorithm.Put("grid1", weightFuture.WaitForResultOrNilValue())
			timesByAlgorithm.Put("grid1", stopwatch.Elapsed())
			printer.Println("///////////////// GRID1 /////////////////")
		})
	}

	if runGrid1Variants {
		tasks = append(tasks, func() {
			printer.Println("################# GRID1 ###################")
			stopwatch := util.StopwatchMakeStopped()
			grid1 := weight_highs.GridStatWeightProcess{}
			grid1.CHECKRANGE = 1
			grid1.Init(printer, standardTimeout)
			grid1.SetRequiredStats(requiredStats)
			grid1.SetTargetRatios(targetRatio)
			grid1.SupplyData(slices.Clone(inputDataGrid))
			weightFuture := grid1.Run(stopwatch)
			util_async.ChainCancel(cancel, weightFuture)
			resultsByAlgorithm.Put("grid1-1", weightFuture.WaitForResultOrNilValue())
			timesByAlgorithm.Put("grid1-1", stopwatch.Elapsed())
			printer.Println("///////////////// GRID1 /////////////////")
		})

		for SCALEMODE := range 5 {
			for ROUNDMODE := range 3 {
				for OUTLIER := range 5 {
					for CALCMODE := range 4 {
						tasks = append(tasks, func() {
							label := fmt.Sprintf("grid1b-outlier%d-scale%d-round%d-calc%d", OUTLIER, SCALEMODE, ROUNDMODE, CALCMODE)
							printer.Println("################# " + label + " ###################")
							stopwatch := util.StopwatchMakeStopped()
							grid1 := weight_highs.GridStatWeightProcess1B{}
							grid1.SCALEMODE = SCALEMODE
							grid1.ROUNDMODE = ROUNDMODE
							grid1.OUTLIER = OUTLIER
							grid1.CALCMODE = CALCMODE
							grid1.Init(printer, shortTimeout)
							grid1.SetRequiredStats(requiredStats)
							grid1.SetTargetRatios(targetRatio)
							grid1.SupplyData(slices.Clone(inputDataGrid))
							weightFuture := grid1.Run(stopwatch)
							util_async.ChainCancel(cancel, weightFuture)
							resultsByAlgorithm.Put(label, weightFuture.WaitForResultOrNilValue())
							timesByAlgorithm.Put(label, stopwatch.Elapsed())
							printer.Println("///////////////// " + label + " /////////////////")
						})
					}
				}
			}
		}
	}
	if runGrid1VariantsFewer {
		for OUTLIER := range 5 {
			label := fmt.Sprintf("grid1b-outlier%d-scale1-round2-calc2", OUTLIER)
			printer.Println("################# " + label + " ###################")
			stopwatch := util.StopwatchMakeStopped()
			grid1 := weight_highs.GridStatWeightProcess1B{}
			grid1.OUTLIER = OUTLIER
			grid1.SCALEMODE = 1
			grid1.ROUNDMODE = 2
			grid1.CALCMODE = 2
			grid1.Init(printer, shortTimeout)
			grid1.SetRequiredStats(requiredStats)
			grid1.SetTargetRatios(targetRatio)
			grid1.SupplyData(slices.Clone(inputDataGrid))
			weightFuture := grid1.Run(stopwatch)
			util_async.ChainCancel(cancel, weightFuture)
			resultsByAlgorithm.Put(label, weightFuture.WaitForResultOrNilValue())
			timesByAlgorithm.Put(label, stopwatch.Elapsed())
			printer.Println("///////////////// " + label + " /////////////////")
		}
	}

	if runGrid1C {
		for RESCALE := range 3 {
			for ROUNDMODE := range 3 {
				tasks = append(tasks, func() {
					label := fmt.Sprintf("grid1c-round%d-rescale%d", ROUNDMODE, RESCALE)
					printer.Println("################# " + label + " ###################")
					stopwatch := util.StopwatchMakeStopped()
					grid1 := weight_highs.GridStatWeightProcess1C{}
					grid1.ROUNDMODE = ROUNDMODE
					grid1.RESCALE = RESCALE
					grid1.Init(printer, shortTimeout)
					grid1.SetRequiredStats(requiredStats)
					grid1.SetTargetRatios(targetRatio)
					grid1.SupplyData(slices.Clone(inputDataGrid))
					weightFuture := grid1.Run(stopwatch)
					util_async.ChainCancel(cancel, weightFuture)
					resultsByAlgorithm.Put(label, weightFuture.WaitForResultOrNilValue())
					timesByAlgorithm.Put(label, stopwatch.Elapsed())
					printer.Println("///////////////// " + label + " /////////////////")
				})
			}
		}
	}

	if runGrid2 {
		tasks = append(tasks, func() {
			label := "grid2-1"
			printer.Println("################# " + label + " ###################")
			stopwatch := util.StopwatchMakeStopped()
			grid2 := weight_highs.GridStatWeightProcess2{}
			grid2.IncludeDiffs1 = true
			grid2.Init(printer, shortTimeout)
			grid2.SetRequiredStats(requiredStats)
			grid2.SetTargetRatios(targetRatio)
			grid2.SupplyData(slices.Clone(inputDataGrid))
			weightFuture := grid2.Run(stopwatch)
			util_async.ChainCancel(cancel, weightFuture)
			resultsByAlgorithm2.Put(label, weightFuture.WaitForResultOrNilValue())
			timesByAlgorithm.Put(label, stopwatch.Elapsed())
			printer.Println("///////////////// " + label + " /////////////////")
		})
		tasks = append(tasks, func() {
			label := "grid2-2"
			printer.Println("################# " + label + " ###################")
			stopwatch := util.StopwatchMakeStopped()
			grid2 := weight_highs.GridStatWeightProcess2{}
			grid2.IncludeDiffs2 = true
			grid2.Init(printer, shortTimeout)
			grid2.SetRequiredStats(requiredStats)
			grid2.SetTargetRatios(targetRatio)
			grid2.SupplyData(slices.Clone(inputDataGrid))
			weightFuture := grid2.Run(stopwatch)
			util_async.ChainCancel(cancel, weightFuture)
			resultsByAlgorithm2.Put(label, weightFuture.WaitForResultOrNilValue())
			timesByAlgorithm.Put(label, stopwatch.Elapsed())
			printer.Println("///////////////// " + label + " /////////////////")
		})
		tasks = append(tasks, func() {
			label := "grid2-12"
			printer.Println("################# " + label + " ###################")
			stopwatch := util.StopwatchMakeStopped()
			grid2 := weight_highs.GridStatWeightProcess2{}
			grid2.IncludeDiffs1 = true
			grid2.IncludeDiffs2 = true
			grid2.Init(printer, shortTimeout)
			grid2.SetRequiredStats(requiredStats)
			grid2.SetTargetRatios(targetRatio)
			grid2.SupplyData(slices.Clone(inputDataGrid))
			weightFuture := grid2.Run(stopwatch)
			util_async.ChainCancel(cancel, weightFuture)
			resultsByAlgorithm2.Put(label, weightFuture.WaitForResultOrNilValue())
			timesByAlgorithm.Put(label, stopwatch.Elapsed())
			printer.Println("///////////////// " + label + " /////////////////")
		})
		tasks = append(tasks, func() {
			label := "grid2-123"
			printer.Println("################# " + label + " ###################")
			stopwatch := util.StopwatchMakeStopped()
			grid2 := weight_highs.GridStatWeightProcess2{}
			grid2.IncludeDiffs1 = true
			grid2.IncludeDiffs2 = true
			grid2.IncludeDiffs3 = true
			grid2.Init(printer, shortTimeout)
			grid2.SetRequiredStats(requiredStats)
			grid2.SetTargetRatios(targetRatio)
			grid2.SupplyData(slices.Clone(inputDataGrid))
			weightFuture := grid2.Run(stopwatch)
			util_async.ChainCancel(cancel, weightFuture)
			resultsByAlgorithm2.Put(label, weightFuture.WaitForResultOrNilValue())
			timesByAlgorithm.Put(label, stopwatch.Elapsed())
			printer.Println("///////////////// " + label + " /////////////////")
		})
	}

	if runRankingOlder {
		for RANKMODE := range 3 {
			for WEIGHTSUM := range 3 {
				tasks = append(tasks, func() {
					label := fmt.Sprintf("rankorig-%d-%d", RANKMODE, WEIGHTSUM)
					printer.Println("################# RANKING0 ###################")
					stopwatch := util.StopwatchMakeStopped()
					ranking := weight_highs.RankingStatWeightProcess{}
					ranking.RANKMODE = RANKMODE
					ranking.WEIGHTSUM = WEIGHTSUM
					ranking.Init(printer)
					ranking.SetRequiredStats(requiredStats)
					ranking.SetTargetRatios(targetRatio)
					ranking.SupplyData(slices.Clone(mixedInputData))
					weightFuture := ranking.Run(stopwatch, standardTimeout)
					util_async.ChainCancel(cancel, weightFuture)
					resultsByAlgorithm.Put(label, weightFuture.WaitForResultOrNilValue())
					timesByAlgorithm.Put(label, stopwatch.Elapsed())
					printer.Println("///////////////// RANKING0 /////////////////")
				})
			}
		}

	}

	if runRanking3aVariants {
		for ALGO := range 2 {
			tasks = append(tasks, func() {
				label := fmt.Sprintf("ranking3a-scale_stat-algo%d", ALGO)
				printer.Println("################# " + label + " ###################")
				stopwatch := util.StopwatchMakeStopped()
				ranking := weight_highs.RankingStatWeightProcess3{}
				ranking.ALGO = ALGO
				ranking.SCALE1 = false
				ranking.Init(printer, shortTimeout)
				ranking.SetRequiredStats(requiredStats)
				ranking.SetTargetRatios(targetRatio)
				ranking.SupplyData(slices.Clone(mixedInputData))
				weightFuture := ranking.Run(stopwatch)
				util_async.ChainCancel(cancel, weightFuture)
				weight := weightFuture.WaitForResultOrNilValue()
				timesByAlgorithm.Put(label, stopwatch.Elapsed())
				resultsByAlgorithm.Put(label, weight)
				printer.Println("///////////////// " + label + " /////////////////")
			})
			tasks = append(tasks, func() {
				label := fmt.Sprintf("ranking3a-scale1-algo%d", ALGO)
				printer.Println("################# " + label + " ###################")
				stopwatch := util.StopwatchMakeStopped()
				ranking := weight_highs.RankingStatWeightProcess3{}
				ranking.ALGO = ALGO
				ranking.SCALE1 = true
				ranking.Init(printer, shortTimeout)
				ranking.SetRequiredStats(requiredStats)
				ranking.SetTargetRatios(targetRatio)
				ranking.SupplyData(slices.Clone(mixedInputData))
				weightFuture := ranking.Run(stopwatch)
				util_async.ChainCancel(cancel, weightFuture)
				weight := weightFuture.WaitForResultOrNilValue()
				timesByAlgorithm.Put(label, stopwatch.Elapsed())
				resultsByAlgorithm.Put(label, weight)
				printer.Println("///////////////// " + label + " /////////////////")
			})
		}
	}

	if runRanking3aPreferred {
		tasks = append(tasks, func() {
			label := fmt.Sprintf("ranking3a-false-1")
			printer.Println("################# " + label + " ###################")
			stopwatch := util.StopwatchMakeStopped()
			ranking := weight_highs.RankingStatWeightProcess3{}
			ranking.ALGO = 1
			ranking.SCALE1 = false
			ranking.Init(printer, standardTimeout)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(mixedInputData))
			weightFuture := ranking.RunUsingExternalStart(weightsMidRange, stopwatch)
			util_async.ChainCancel(cancel, weightFuture)
			weight := weightFuture.WaitForResultOrNilValue()
			timesByAlgorithm.Put(label, stopwatch.Elapsed())
			resultsByAlgorithm.Put(label, weight)
			printer.Println("///////////////// " + label + " /////////////////")
		})
		tasks = append(tasks, func() {
			label := fmt.Sprintf("ranking3a-true-1")
			printer.Println("################# " + label + " ###################")
			stopwatch := util.StopwatchMakeStopped()
			ranking := weight_highs.RankingStatWeightProcess3{}
			ranking.ALGO = 1
			ranking.SCALE1 = true
			ranking.Init(printer, standardTimeout)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(mixedInputData))
			weightFuture := ranking.RunUsingExternalStart(weightsMidRange, stopwatch)
			util_async.ChainCancel(cancel, weightFuture)
			weight := weightFuture.WaitForResultOrNilValue()
			timesByAlgorithm.Put(label, stopwatch.Elapsed())
			resultsByAlgorithm.Put(label, weight)
			printer.Println("///////////////// " + label + " /////////////////")
		})
	}

	if runRanking3bVariants {
		for TOTALWEIGHT := range 3 {
			for ALGO := range 2 {
				tasks = append(tasks, func() {
					label := fmt.Sprintf("ranking3b-%d-%d", TOTALWEIGHT, ALGO)
					printer.Println("################# " + label + " ###################")
					stopwatch := util.StopwatchMakeStopped()
					ranking := weight_highs.RankingStatWeightProcess3b{}
					ranking.TOTALWEIGHT = TOTALWEIGHT
					ranking.ALGO = ALGO
					ranking.Init(printer, shortTimeout)
					ranking.SetRequiredStats(requiredStats)
					ranking.SetTargetRatios(targetRatio)
					ranking.SupplyData(slices.Clone(mixedInputData))
					weightFuture := ranking.RunSinglePassFromExternal(weightsMidRange, stopwatch)
					util_async.ChainCancel(cancel, weightFuture)
					weight := weightFuture.WaitForResultOrNilValue()
					timesByAlgorithm.Put(label, stopwatch.Elapsed())
					resultsByAlgorithm.Put(label, weight)
					printer.Println("///////////////// " + label + " /////////////////")
				})
			}
		}
	}
	if runRanking3bPreferred {
		tasks = append(tasks, func() {
			label := fmt.Sprintf("ranking3b-pref")
			printer.Println("################# " + label + " ###################")
			stopwatch := util.StopwatchMakeStopped()
			ranking := weight_highs.RankingStatWeightProcess3b{}
			ranking.TOTALWEIGHT = 0
			ranking.ALGO = 1
			ranking.Init(printer, shortTimeout)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(mixedInputData))
			weightFuture := ranking.RunSinglePassFromExternal(weightsMidRange, stopwatch)
			util_async.ChainCancel(cancel, weightFuture)
			weight := weightFuture.WaitForResultOrNilValue()
			timesByAlgorithm.Put(label, stopwatch.Elapsed())
			resultsByAlgorithm.Put(label, weight)
			printer.Println("///////////////// " + label + " /////////////////")
		})
	}

	if runRanking4 {
		for MULTIPLY := range 3 {
			tasks = append(tasks, func() {
				label := fmt.Sprintf("ranking4-%d", MULTIPLY)
				printer.Println("################# RANKING4 ###################")
				stopwatch := util.StopwatchMakeStopped()
				ranking := weight_highs.RankingStatWeightProcess4{}
				ranking.MULTIPLY = MULTIPLY
				ranking.Init(printer)
				ranking.SetRequiredStats(requiredStats)
				ranking.SetTargetRatios(targetRatio)
				ranking.SupplyData(slices.Clone(inputDataRandom))

				weight := ranking.RunUsingExternalStart(weightsMidRange, stopwatch, standardTimeout)
				timesByAlgorithm.Put(label, stopwatch.Elapsed())
				resultsByAlgorithm.Put(label, weight.GetOrPanic())
				printer.Println("///////////////// RANKING4 /////////////////")
			})
		}
	}

	if runRanking5 {
		tasks = append(tasks, func() {
			printer.Println("################# RANKING5 ###################")
			stopwatch := util.StopwatchMakeStopped()
			ranking := weight_highs.RankingStatWeightProcess5{}
			ranking.WEIGHTSUM = 0
			ranking.Init(printer)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(inputDataRandom))
			futureWeight := ranking.Run(stopwatch, standardTimeout)
			weight, _ := futureWeight.WaitForResult()
			timesByAlgorithm.Put("ranking5-0", stopwatch.Elapsed())
			resultsByAlgorithm.Put("ranking5-0", weight)
			printer.Println("///////////////// RANKING5 /////////////////")
		})
		tasks = append(tasks, func() {
			printer.Println("################# RANKING5 ###################")
			stopwatch := util.StopwatchMakeStopped()
			ranking := weight_highs.RankingStatWeightProcess5{}
			ranking.WEIGHTSUM = 1
			ranking.Init(printer)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(inputDataRandom))
			futureWeight := ranking.Run(stopwatch, standardTimeout)
			weight, _ := futureWeight.WaitForResult()
			timesByAlgorithm.Put("ranking5-1", stopwatch.Elapsed())
			resultsByAlgorithm.Put("ranking5-1", weight)
			printer.Println("///////////////// RANKING5 /////////////////")
		})
	}

	if runSearch0 {
		tasks = append(tasks, func() {
			printer.Println("################# SEARCH0 ###################")
			stopwatch := util.StopwatchMakeStarted()
			search := weightfind.WeightSearcher0{}
			search.AccuracyStatistical = false
			search.Init(requiredStats, targetRatio, printer)
			search.SupplyData(mixedInputData)
			resultsByAlgorithm.Put("search0-accF", search.Run(cancel))
			timesByAlgorithm.Put("search0-accF", stopwatch.Elapsed())
			printer.Println("///////////////// SEARCH0 /////////////////")
		})
		tasks = append(tasks, func() {
			printer.Println("################# SEARCH0 ###################")
			stopwatch := util.StopwatchMakeStarted()
			search := weightfind.WeightSearcher0{}
			search.AccuracyStatistical = true
			search.Init(requiredStats, targetRatio, printer)
			search.SupplyData(mixedInputData)
			resultsByAlgorithm.Put("search0-accT", search.Run(cancel))
			timesByAlgorithm.Put("search0-accT", stopwatch.Elapsed())
			printer.Println("///////////////// SEARCH0 /////////////////")
		})
	}
	if runSearches {
		tasks = append(tasks, func() {
			printer.Println("################# SEARCH1 ###################")
			stopwatch := util.StopwatchMakeStarted()
			search := weightfind.WeightSearcher1{}
			search.AccuracyStatistical = false
			search.Init(requiredStats, targetRatio, printer)
			search.SupplyData(mixedInputData)
			resultsByAlgorithm.Put("search1-accF", search.Run(cancel))
			timesByAlgorithm.Put("search1-accF", stopwatch.Elapsed())
			printer.Println("///////////////// SEARCH1 /////////////////")
		})
		tasks = append(tasks, func() {
			printer.Println("################# SEARCH2 ###################")
			stopwatch := util.StopwatchMakeStarted()
			search := weightfind.WeightSearcher2{}
			search.AccuracyStatistical = false
			search.Init(requiredStats, targetRatio, printer)
			search.SupplyData(mixedInputData)
			search.SetRanges(-1.0, 10.0)
			resultsByAlgorithm.Put("search2-accF", search.Run(cancel))
			timesByAlgorithm.Put("search2-accF", stopwatch.Elapsed())
			printer.Println("///////////////// SEARCH2 /////////////////")
		})
		tasks = append(tasks, func() {
			printer.Println("################# SEARCH3 ###################")
			stopwatch := util.StopwatchMakeStarted()
			search := weightfind.WeightSearcher3{}
			search.AccuracyStatistical = false
			search.Init(requiredStats, targetRatio)
			search.SupplyData(mixedInputData)
			search.SetRanges(-1.0, 10.0)
			resultsByAlgorithm.Put("search3-accF", search.Run(cancel))
			timesByAlgorithm.Put("search3-accF", stopwatch.Elapsed())
			printer.Println("///////////////// SEARCH3 /////////////////")
		})

		tasks = append(tasks, func() {
			printer.Println("################# SEARCH1 ###################")
			stopwatch := util.StopwatchMakeStarted()
			search := weightfind.WeightSearcher1{}
			search.AccuracyStatistical = true
			search.Init(requiredStats, targetRatio, printer)
			search.SupplyData(mixedInputData)
			resultsByAlgorithm.Put("search1-accT", search.Run(cancel))
			timesByAlgorithm.Put("search1-accT", stopwatch.Elapsed())
			printer.Println("///////////////// SEARCH1 /////////////////")
		})
		tasks = append(tasks, func() {
			printer.Println("################# SEARCH2 ###################")
			stopwatch := util.StopwatchMakeStarted()
			search := weightfind.WeightSearcher2{}
			search.AccuracyStatistical = true
			search.Init(requiredStats, targetRatio, printer)
			search.SupplyData(mixedInputData)
			search.SetRanges(-1.0, 10.0)
			resultsByAlgorithm.Put("search2-accT", search.Run(cancel))
			timesByAlgorithm.Put("search2-accT", stopwatch.Elapsed())
			printer.Println("///////////////// SEARCH2 /////////////////")
		})
		tasks = append(tasks, func() {
			printer.Println("################# SEARCH3 ###################")
			stopwatch := util.StopwatchMakeStarted()
			search := weightfind.WeightSearcher3{}
			search.AccuracyStatistical = true
			search.Init(requiredStats, targetRatio)
			search.SupplyData(mixedInputData)
			search.SetRanges(-1.0, 10.0)
			resultsByAlgorithm.Put("search3-accT", search.Run(cancel))
			timesByAlgorithm.Put("search3-accT", stopwatch.Elapsed())
			printer.Println("///////////////// SEARCH3 /////////////////")
		})
	}

	if runRankingSep {
		label := "ranking-sep"
		printer.Println("################# " + label + " ###################")
		stopwatch := util.StopwatchMakeStopped()
		comp := weight_highs.RankingSeparatedWeights{}
		comp.Init(printer, standardTimeout)
		comp.SetRequiredStats(requiredStats, requiredSims)
		comp.SetTargetRatios(targetRatio)
		comp.SupplyData(mixedInputData)
		weights2 := comp.Run(stopwatch).WaitForResultOrPanic()
		resultsByAlgorithm2.Put(label, weights2)
		timesByAlgorithm.Put(label, stopwatch.Elapsed())
		printer.Println("///////////////// " + label + " /////////////////")
	}
	if runFormula2 {
		label := "formula2"
		printer.Println("################# " + label + " ###################")
		stopwatch := util.StopwatchMakeStopped()
		comp := weight_highs.FormulaStatWeightProcess2{}
		// comp.BLEND if SetMinimumIncludeRate < 1.0
		comp.Init(printer)
		comp.SetRequiredStats(requiredStats)
		comp.SetTargetRatios(targetRatio)
		comp.SetMinimumIncludeRate(1.0)
		comp.SupplyData(mixedInputData)
		weights2 := comp.Run(stopwatch, standardTimeout).WaitForResultOrPanic()
		resultsByAlgorithm2.Put(label, weights2)
		timesByAlgorithm.Put(label, stopwatch.Elapsed())
		printer.Println("///////////////// " + label + " /////////////////")
	}

	util_collection.Shuffle(tasks)

	util_async.CancelOnKeyPress(cancel)
	util_async.ForEach_Slice_Cancellable(5, tasks, cancel, func(f *func()) {
		(*f)()
	})

	type algorithmReport struct {
		weight3         *weight_types.Weight3ExtendedRanged
		weight2         *weight_types.Weight2Extended
		weight1         *weight_types.Weight1Basic
		initialAccuracy float64
	}
	reportByAlgorithm := make(map[string]algorithmReport)
	for label, weight3 := range resultsByAlgorithm3.SeqWithKeys_StaleInefficient() {
		weight2 := weight3.ConvertToWeight2()
		weight1 := weight2.ConvertToWeight1()
		reportByAlgorithm[label] = algorithmReport{
			weight3:         &weight3,
			weight2:         weight2,
			weight1:         &weight1,
			initialAccuracy: weightfind.EvaluateAccuracy(&weight3, requiredSims, &targetRatio, mixedInputDataFull),
		}
	}
	for label, weight2 := range resultsByAlgorithm2.SeqWithKeys_StaleInefficient() {
		weight1 := weight2.ConvertToWeight1()
		reportByAlgorithm[label] = algorithmReport{
			weight2:         &weight2,
			weight1:         &weight1,
			initialAccuracy: weightfind.EvaluateAccuracy(&weight2, requiredSims, &targetRatio, mixedInputDataFull),
		}
	}
	for label, weight1 := range resultsByAlgorithm.SeqWithKeys_StaleInefficient() {
		reportByAlgorithm[label] = algorithmReport{
			weight1:         &weight1,
			initialAccuracy: weightfind.EvaluateAccuracy(&weight1, requiredSims, &targetRatio, mixedInputDataFull),
		}
	}

	printer.Println("################# FINAL RESULT ###################")
	tab := util.TabulateOutput{}
	tab.SetColumnSpacing(2)
	tab.AddColumnHeader("algo", false)
	for _, stat := range requiredStats {
		tab.AddColumnHeader(stat.Name(), true)
	}
	tab.AddColumnHeader("accuracy", false)
	tab.AddColumnHeader("accuracy_tweaked", false)
	tab.AddColumnHeader("accuracy_stat", false)
	tab.AddColumnHeader("time", false)

	resultOrder := slices.Collect(maps.Keys(reportByAlgorithm))
	slices.SortFunc(resultOrder, func(a, b string) int {
		return cmp.Compare(reportByAlgorithm[a].initialAccuracy, reportByAlgorithm[b].initialAccuracy)
	})

	for _, label := range resultOrder {
		report := reportByAlgorithm[label]
		row := make([]string, 0)
		row = append(row, label)
		for _, stat := range requiredStats {
			value := report.weight1.Get(stat)
			row = append(row, strconv.FormatFloat(value, 'f', 4, 64))
		}
		accuracy := report.initialAccuracy
		accuracyStat := weightfind.EvaluateAccuracyStatistical(report.weight1, requiredSims, &targetRatio, mixedInputDataFull)
		row = append(row, strconv.FormatFloat(accuracy, 'f', 4, 64))
		row = append(row, "")
		row = append(row, strconv.FormatFloat(accuracyStat, 'f', 4, 64))
		row = append(row, timesByAlgorithm.GetOrNil(label).String())
		tab.AddRow(row)

		if reportOnTweakedVersions {
			weightTweak, _ := weightfind.WeightTweakerWithLogging(*report.weight1, requiredStats, &targetRatio, mixedInputDataFull, util.PrintRecorder_Nop())
			accuracyTweak := weightfind.EvaluateAccuracy(&weightTweak, requiredSims, &targetRatio, mixedInputDataFull)
			accuracyTweakStat := weightfind.EvaluateAccuracyStatistical(&weightTweak, requiredSims, &targetRatio, mixedInputDataFull)
			row = make([]string, 0)
			row = append(row, label)
			for _, stat := range requiredStats {
				value := weightTweak.Get(stat)
				row = append(row, strconv.FormatFloat(value, 'f', 4, 64))
			}
			row = append(row, "")
			row = append(row, strconv.FormatFloat(accuracyTweak, 'f', 4, 64))
			row = append(row, strconv.FormatFloat(accuracyTweakStat, 'f', 4, 64))
			row = append(row, timesByAlgorithm.GetOrNil(label).String())
			tab.AddRow(row)
		}
	}
	tab.Write(printer)
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
