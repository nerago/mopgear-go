package main

import (
	"cmp"
	"encoding/json/v2"
	"fmt"
	"maps"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/nerago/mopgear-go/gear_model/model_factory"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind"
	"github.com/nerago/mopgear-go/weightfind/weight_highs"
	"github.com/nerago/mopgear-go/weightfind/weight_highs/fitting1"
	"github.com/nerago/mopgear-go/weightfind/weight_highs/fitting2"
	"github.com/nerago/mopgear-go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

//goland:noinspection GoBoolExpressions
func statWeights_CompareAlgorithms(printer *util.PrintRecorder) {
	//targetRatio := model_factory.SimPriority_survival
	targetRatio := model_factory.SimPriority_mitigation
	//targetRatio := gear_model.SimPriority_heal
	requiredStats := model_factory.StatsForWeighting_strengthTank
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
	//weight_types.WeightInputWriteFile(inputDataGrid, "tempdata/sim-stats-compare-grid.json")
	//weight_types.WeightInputWriteFile(inputDataRandomUnsafe, "tempdata/sim-stats-compare-rand-unsafe.json")
	//weight_types.WeightInputWriteFile(inputDataRandomSafe, "tempdata/sim-stats-compare-rand-safe.json")
	//weight_types.WeightInputWriteFile(inputDataRandom, "tempdata/sim-stats-compare-rand.json")
	//writeWeightBasicInputsToFile(inputDataBasic, basicSimBase, "tempdata/sim-stats-compare-basic.json")

	inputDataBasic, basicSimBase := readWeightBasicInputsFile("tempdata/sim-stats-compare-basic.json")
	//inputDataGrid := readWeightInputFile("tempdata/weightfind-sim-grid-Prot-Mitigation-NoSet.json")
	//inputDataRandom := readWeightInputFile("tempdata/weightfind-sim-real-Prot-Mitigation-NoSet.json")
	inputDataGrid := weight_types.WeightInputReadFile("tempdata/weightfind-sim-grid-Prot-Mitigation.json")
	inputDataRandom := weight_types.WeightInputReadFile("tempdata/weightfind-sim-real-Prot-Mitigation.json")
	//inputDataGrid := readWeightInputFile("tempdata/sim-stats-compare-grid.json")
	//inputDataRandom := readWeightInputFile("tempdata/sim-stats-compare-rand.json")
	mixedInputDataFull := slices.Concat(inputDataGrid, inputDataRandom)

	//sampleSize := 50
	//inputDataGrid = util_collection.SliceSampleRandom(inputDataGrid, sampleSize)
	//inputDataRandom = util_collection.SliceSampleRandom(inputDataRandom, sampleSize)
	//mixedInputData := util_collection.SliceSampleRandom(mixedInputDataFull, sampleSize)
	//inputDataGrid = util_collection.SliceSampleRandom(inputDataGrid, sampleSize, 1234)
	//inputDataRandom = util_collection.SliceSampleRandom(inputDataRandom, sampleSize, 1234)
	//mixedInputData := util_collection.SliceSampleRandom(mixedInputDataFull, sampleSize, 1234)
	mixedInputData := mixedInputDataFull

	// weight value 87.7342
	weightsMidRange := weight_types.Weight1Basic_Make()
	weightsMidRange.Put(stats.Stat_Strength, 1.0000)
	weightsMidRange.Put(stats.Stat_Stamina, 1.2309)
	weightsMidRange.Put(stats.Stat_Crit, 0.1167)
	weightsMidRange.Put(stats.Stat_Haste, 0.3614)
	weightsMidRange.Put(stats.Stat_Expertise, 0.0054)
	weightsMidRange.Put(stats.Stat_Mastery, 0.5866)
	weightsMidRange.Put(stats.Stat_Dodge, 0.0824)
	weightsMidRange.Put(stats.Stat_Parry, 0.0532)

	cancel := util_async.CancelSignal_Make()
	util_async.CancelOnKeyPress(cancel)

	taskMap := make(map[string]func() weight_types.WeightResult)
	addTask := func(label string, run func() weight_types.WeightResult) {
		if taskMap[label] != nil {
			panic("duplicate")
		}
		taskMap[label] = run
	}

	reportOnTweakedVersions := false
	standardTimeout := 600
	shortTimeout := 400

	runBasic := true
	runFormulaVariants := false // best is about 87%, moderate time
	runFitting1 := false        // slow, low 90%
	runFitting2 := false

	runGrid1Original := true
	runGrid1Variants := false // todo return
	runGrid1VariantsFewer := false
	runGrid1C := true
	runGrid2 := true

	runRankingOlder := true
	runRanking3aPreferred := false // broken
	runRanking3aVariants := false  // broken
	runRanking3bVariants := false
	runRanking3bPreferred := true
	runRanking3c := true
	runRanking4 := true  // still a little slow, midrange 94% etc
	runRanking5 := false // excellent but slow

	runSearches := false
	runSearch0 := false

	runRankingSep := true
	runFormula2 := true

	//runBasic := true
	//runFormulaVariants := true // best is about 87%, moderate time
	//runFitting1 := true        // slow, low 90%
	//runFitting2 := true
	//
	//runGrid1Original := true
	//runGrid1Variants := false // todo return
	//runGrid1VariantsFewer := false
	//runGrid1C := true
	//runGrid2 := true
	//
	//runRankingOlder := true
	//runRanking3aPreferred := false // broken
	//runRanking3aVariants := false  // broken
	//runRanking3bVariants := true
	//runRanking3bPreferred := true
	//runRanking3c := true
	//runRanking4 := true  // still a little slow, midrange 94% etc
	//runRanking5 := false // excellent but slow
	//
	//runSearches := true
	//runSearch0 := true
	//
	//runRankingSep := true
	//runFormula2 := true

	addTask("itemLevel", func() weight_types.WeightResult {
		weight := weight_types.Weight1Basic_Make()
		weight.Put(requiredStats[0], 1)
		return weight_types.WeightResult{
			Weight:    &weight,
			SolveTime: 0,
			Status:    highs.ModelStatusOptimal,
		}
	})

	if runBasic {
		addTask("basic", func() weight_types.WeightResult {
			basic := weight_highs.BasicStatWeightProcess{}
			basic.Init(printer)
			basic.SetRequiredStats(requiredStats)
			basic.SetTargetRatios(targetRatio)
			basic.SetBaseline(basicSimBase)
			for _, data := range inputDataBasic {
				basic.AddSimData(data.IncrementStat, uint32(data.IncrementValue), data.SimResult)
			}
			weightFuture := basic.Run()
			util_async.ChainCancel(cancel, weightFuture)
			return weightFuture.WaitForResultOrNilValue()
		})

		addTask("form", func() weight_types.WeightResult {
			comp := weight_highs.FormulaStatWeightProcess{}
			comp.Init(printer)
			comp.SetRequiredStats(requiredStats)
			comp.SetTargetRatios(targetRatio)
			comp.SetMinimumIncludeRate(1)
			comp.SupplyData(slices.Clone(inputDataRandom))
			weightFuture := comp.Run(standardTimeout)
			util_async.ChainCancel(cancel, weightFuture)
			return weightFuture.WaitForResultOrNilValue()
		})
	}

	if runFormulaVariants {
		for BLEND := range 6 {
			label := fmt.Sprintf("form-blend%d-inc70", BLEND)
			addTask(label, func() weight_types.WeightResult {
				comp := weight_highs.FormulaStatWeightProcess{}
				comp.BLEND = BLEND
				comp.Init(printer)
				comp.SetRequiredStats(requiredStats)
				comp.SetTargetRatios(targetRatio)
				comp.SetMinimumIncludeRate(0.7)
				comp.SupplyData(slices.Clone(inputDataRandom))
				futureResult := comp.Run(standardTimeout)
				util_async.ChainCancel(cancel, futureResult)
				return futureResult.WaitForResultOrNilValue()
			})
			label = fmt.Sprintf("form-blend%d-inc100", BLEND)
			addTask(label, func() weight_types.WeightResult {
				comp := weight_highs.FormulaStatWeightProcess{}
				comp.BLEND = BLEND
				comp.Init(printer)
				comp.SetRequiredStats(requiredStats)
				comp.SetTargetRatios(targetRatio)
				comp.SetMinimumIncludeRate(1)
				comp.SupplyData(slices.Clone(inputDataRandom))
				futureResult := comp.Run(standardTimeout)
				util_async.ChainCancel(cancel, futureResult)
				return futureResult.WaitForResultOrNilValue()
			})
		}
	}

	if runFitting1 {
		addTask("fitting", func() weight_types.WeightResult {
			fitting := fitting1.FittingEachStatWeightProcess{}
			fitting.Init(printer, shortTimeout)
			fitting.SetRequiredStats(requiredStats, requiredSims)
			fitting.SetOnlyComputeSingleSegmentEach(true)
			fitting.SupplyData(inputDataRandom)
			return fitting.Run(cancel)
		})
	}
	if runFitting2 {
		addTask("fitting2", func() weight_types.WeightResult {
			fitting := fitting2.FittingEachStatWeightProcess2{}
			fitting.Init(3, printer, shortTimeout)
			fitting.SetRequiredStats(requiredStats, requiredSims)
			fitting.SetTargetRatios(targetRatio)
			fitting.SupplyData(inputDataRandom)
			return fitting.Run(cancel)
		})
	}

	if runGrid1Original {
		addTask("grid1", func() weight_types.WeightResult {
			grid1 := weight_highs.GridStatWeightProcess{}
			grid1.Init(printer, standardTimeout)
			grid1.SetRequiredStats(requiredStats)
			grid1.SetTargetRatios(targetRatio)
			grid1.SupplyData(slices.Clone(inputDataGrid))
			weightFuture := grid1.Run()
			util_async.ChainCancel(cancel, weightFuture)
			return weightFuture.WaitForResultOrNilValue()
		})
	}

	if runGrid1Variants {
		addTask("grid1-1", func() weight_types.WeightResult {
			grid1 := weight_highs.GridStatWeightProcess{}
			grid1.CHECKRANGE = 1
			grid1.Init(printer, standardTimeout)
			grid1.SetRequiredStats(requiredStats)
			grid1.SetTargetRatios(targetRatio)
			grid1.SupplyData(slices.Clone(inputDataGrid))
			weightFuture := grid1.Run()
			util_async.ChainCancel(cancel, weightFuture)
			return weightFuture.WaitForResultOrNilValue()
		})

		for SCALEMODE := range 5 {
			for ROUNDMODE := range 3 {
				for OUTLIER := range 5 {
					for CALCMODE := range 4 {
						for RATIO := range 4 {
							label := fmt.Sprintf("grid1b-outlier%d-scale%d-round%d-calc%d-rat%d", OUTLIER, SCALEMODE, ROUNDMODE, CALCMODE, RATIO)
							addTask(label, func() weight_types.WeightResult {
								grid1 := weight_highs.GridStatWeightProcess1B{}
								grid1.SCALEMODE = SCALEMODE
								grid1.ROUNDMODE = ROUNDMODE
								grid1.OUTLIER = OUTLIER
								grid1.CALCMODE = CALCMODE
								grid1.RATIO = RATIO
								grid1.Init(printer, shortTimeout)
								grid1.SetRequiredStats(requiredStats)
								grid1.SetTargetRatios(targetRatio)
								grid1.SupplyData(slices.Clone(inputDataGrid))
								weightFuture := grid1.Run()
								util_async.ChainCancel(cancel, weightFuture)
								return weightFuture.WaitForResultOrNilValue()
							})
						}
					}
				}
			}
		}
	}
	if runGrid1VariantsFewer {
		for OUTLIER := range 5 {
			label := fmt.Sprintf("grid1b-outlier%d-scale1-round2-calc2", OUTLIER)
			addTask(label, func() weight_types.WeightResult {
				grid1 := weight_highs.GridStatWeightProcess1B{}
				grid1.OUTLIER = OUTLIER
				grid1.SCALEMODE = 1
				grid1.ROUNDMODE = 2
				grid1.CALCMODE = 2
				grid1.Init(printer, shortTimeout)
				grid1.SetRequiredStats(requiredStats)
				grid1.SetTargetRatios(targetRatio)
				grid1.SupplyData(slices.Clone(inputDataGrid))
				weightFuture := grid1.Run()
				util_async.ChainCancel(cancel, weightFuture)
				return weightFuture.WaitForResultOrNilValue()
			})
		}
	}

	if runGrid1C {
		for RESCALE := range 3 {
			for ROUNDMODE := range 3 {
				label := fmt.Sprintf("grid1c-round%d-rescale%d", ROUNDMODE, RESCALE)
				addTask(label, func() weight_types.WeightResult {
					grid1 := weight_highs.GridStatWeightProcess1C{}
					grid1.ROUNDMODE = ROUNDMODE
					grid1.RESCALE = RESCALE
					grid1.Init(printer, shortTimeout)
					grid1.SetRequiredStats(requiredStats)
					grid1.SetTargetRatios(targetRatio)
					grid1.SupplyData(slices.Clone(inputDataGrid))
					weightFuture := grid1.Run()
					util_async.ChainCancel(cancel, weightFuture)
					return weightFuture.WaitForResultOrNilValue()
				})
			}
		}
	}

	if runGrid2 {
		addTask("grid2-1", func() weight_types.WeightResult {
			grid2 := weight_highs.GridStatWeightProcess2{}
			grid2.IncludeDiffs1 = true
			grid2.Init(printer, shortTimeout)
			grid2.SetRequiredStats(requiredStats)
			grid2.SetTargetRatios(targetRatio)
			grid2.SupplyData(slices.Clone(inputDataGrid))
			weightFuture := grid2.Run()
			util_async.ChainCancel(cancel, weightFuture)
			return weightFuture.WaitForResultOrNilValue()
		})
		addTask("grid2-2", func() weight_types.WeightResult {
			grid2 := weight_highs.GridStatWeightProcess2{}
			grid2.IncludeDiffs2 = true
			grid2.Init(printer, shortTimeout)
			grid2.SetRequiredStats(requiredStats)
			grid2.SetTargetRatios(targetRatio)
			grid2.SupplyData(slices.Clone(inputDataGrid))
			weightFuture := grid2.Run()
			util_async.ChainCancel(cancel, weightFuture)
			return weightFuture.WaitForResultOrNilValue()
		})
		addTask("grid2-12", func() weight_types.WeightResult {
			grid2 := weight_highs.GridStatWeightProcess2{}
			grid2.IncludeDiffs1 = true
			grid2.IncludeDiffs2 = true
			grid2.Init(printer, shortTimeout)
			grid2.SetRequiredStats(requiredStats)
			grid2.SetTargetRatios(targetRatio)
			grid2.SupplyData(slices.Clone(inputDataGrid))
			weightFuture := grid2.Run()
			util_async.ChainCancel(cancel, weightFuture)
			return weightFuture.WaitForResultOrNilValue()
		})
		addTask("grid2-123", func() weight_types.WeightResult {
			grid2 := weight_highs.GridStatWeightProcess2{}
			grid2.IncludeDiffs1 = true
			grid2.IncludeDiffs2 = true
			grid2.IncludeDiffs3 = true
			grid2.Init(printer, shortTimeout)
			grid2.SetRequiredStats(requiredStats)
			grid2.SetTargetRatios(targetRatio)
			grid2.SupplyData(slices.Clone(inputDataGrid))
			weightFuture := grid2.Run()
			util_async.ChainCancel(cancel, weightFuture)
			return weightFuture.WaitForResultOrNilValue()
		})
	}

	if runRankingOlder {
		for RANKMODE := range 3 {
			for WEIGHTSUM := range 3 {
				label := fmt.Sprintf("rankorig-%d-%d", RANKMODE, WEIGHTSUM)
				addTask(label, func() weight_types.WeightResult {
					ranking := weight_highs.RankingStatWeightProcess{}
					ranking.RANKMODE = RANKMODE
					ranking.WEIGHTSUM = WEIGHTSUM
					ranking.Init(printer)
					ranking.SetRequiredStats(requiredStats)
					ranking.SetTargetRatios(targetRatio)
					ranking.SupplyData(slices.Clone(mixedInputData))
					weightFuture := ranking.Run(standardTimeout)
					util_async.ChainCancel(cancel, weightFuture)
					return weightFuture.WaitForResultOrNilValue()
				})
			}
		}

	}

	if runRanking3aVariants {
		for ALGO := range 2 {
			label := fmt.Sprintf("ranking3a-scale_stat-algo%d", ALGO)
			addTask(label, func() weight_types.WeightResult {
				ranking := weight_highs.RankingStatWeightProcess3{}
				ranking.ALGO = ALGO
				ranking.SCALE1 = false
				ranking.Init(printer, shortTimeout)
				ranking.SetRequiredStats(requiredStats)
				ranking.SetTargetRatios(targetRatio)
				ranking.SupplyData(slices.Clone(mixedInputData))
				weightFuture := ranking.Run()
				util_async.ChainCancel(cancel, weightFuture)
				return weightFuture.WaitForResultOrNilValue()
			})
			label = fmt.Sprintf("ranking3a-scale1-algo%d", ALGO)
			addTask(label, func() weight_types.WeightResult {
				ranking := weight_highs.RankingStatWeightProcess3{}
				ranking.ALGO = ALGO
				ranking.SCALE1 = true
				ranking.Init(printer, shortTimeout)
				ranking.SetRequiredStats(requiredStats)
				ranking.SetTargetRatios(targetRatio)
				ranking.SupplyData(slices.Clone(mixedInputData))
				weightFuture := ranking.Run()
				util_async.ChainCancel(cancel, weightFuture)
				return weightFuture.WaitForResultOrNilValue()
			})
		}
	}

	if runRanking3aPreferred {
		label := fmt.Sprintf("ranking3a-false-1")
		addTask(label, func() weight_types.WeightResult {
			ranking := weight_highs.RankingStatWeightProcess3{}
			ranking.ALGO = 1
			ranking.SCALE1 = false
			ranking.Init(printer, standardTimeout)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(mixedInputData))
			weightFuture := ranking.RunUsingExternalStart(weightsMidRange)
			util_async.ChainCancel(cancel, weightFuture)
			return weightFuture.WaitForResultOrNilValue()
		})
		label = fmt.Sprintf("ranking3a-true-1")
		addTask(label, func() weight_types.WeightResult {
			ranking := weight_highs.RankingStatWeightProcess3{}
			ranking.ALGO = 1
			ranking.SCALE1 = true
			ranking.Init(printer, standardTimeout)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(mixedInputData))
			weightFuture := ranking.RunUsingExternalStart(weightsMidRange)
			util_async.ChainCancel(cancel, weightFuture)
			return weightFuture.WaitForResultOrNilValue()
		})
	}

	if runRanking3bVariants {
		for TOTALWEIGHT := range 3 {
			for ALGO := range 2 {
				label := fmt.Sprintf("ranking3b-%d-%d", TOTALWEIGHT, ALGO)
				addTask(label, func() weight_types.WeightResult {
					ranking := weight_highs.RankingStatWeightProcess3b{}
					ranking.TOTALWEIGHT = TOTALWEIGHT
					ranking.ALGO = ALGO
					ranking.Init(printer, shortTimeout)
					ranking.SetRequiredStats(requiredStats)
					ranking.SetTargetRatios(targetRatio)
					ranking.SupplyData(slices.Clone(mixedInputData))
					weightFuture := ranking.RunSinglePassFromExternal(weightsMidRange)
					util_async.ChainCancel(cancel, weightFuture)
					return weightFuture.WaitForResultOrNilValue()
				})
			}
		}
	}
	if runRanking3bPreferred {
		label := fmt.Sprintf("ranking3b-pref")
		addTask(label, func() weight_types.WeightResult {
			ranking := weight_highs.RankingStatWeightProcess3b{}
			ranking.TOTALWEIGHT = 0
			ranking.ALGO = 1
			ranking.Init(printer, shortTimeout)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(mixedInputData))
			weightFuture := ranking.RunSinglePassFromExternal(weightsMidRange)
			util_async.ChainCancel(cancel, weightFuture)
			return weightFuture.WaitForResultOrNilValue()
		})
	}

	if runRanking3c {
		addTask("ranking3c", func() weight_types.WeightResult {
			ranking := weight_highs.RankingStatWeightProcess3c{}
			ranking.Init(printer, shortTimeout)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(mixedInputData))
			weightFuture := ranking.RunMultiRound()
			util_async.ChainCancel(cancel, weightFuture)
			return weightFuture.WaitForResultOrNilValue()
		})
	}

	if runRanking4 {
		for MULTIPLY := range 3 {
			label := fmt.Sprintf("ranking4-%d", MULTIPLY)
			addTask(label, func() weight_types.WeightResult {
				ranking := weight_highs.RankingStatWeightProcess4{}
				ranking.MULTIPLY = MULTIPLY
				ranking.Init(printer)
				ranking.SetRequiredStats(requiredStats)
				ranking.SetTargetRatios(targetRatio)
				ranking.SupplyData(slices.Clone(inputDataRandom))
				weightFuture := ranking.RunUsingExternalStart(weightsMidRange, standardTimeout)
				util_async.ChainCancel(cancel, weightFuture)
				return weightFuture.WaitForResultOrNilValue()
			})
		}
	}

	if runRanking5 {
		addTask("ranking5-0", func() weight_types.WeightResult {
			ranking := weight_highs.RankingStatWeightProcess5{}
			ranking.WEIGHTSUM = 0
			ranking.Init(printer)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(inputDataRandom))
			futureWeight := ranking.Run(standardTimeout, 60) // note restricted data sample
			return futureWeight.WaitForResultOrNilValue()
		})
		addTask("ranking5-1", func() weight_types.WeightResult {
			ranking := weight_highs.RankingStatWeightProcess5{}
			ranking.WEIGHTSUM = 1
			ranking.Init(printer)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(inputDataRandom))
			futureWeight := ranking.Run(standardTimeout, 60) // note restricted data sample
			return futureWeight.WaitForResultOrNilValue()
		})
	}

	if runSearch0 {
		addTask("search0-accF", func() weight_types.WeightResult {
			search := weightfind.WeightSearcher0{}
			search.AccuracyStatistical = false
			search.Init(requiredStats, targetRatio, printer)
			search.SupplyData(mixedInputData)
			return search.Run(cancel)
		})
		addTask("search0-accT", func() weight_types.WeightResult {
			search := weightfind.WeightSearcher0{}
			search.AccuracyStatistical = true
			search.Init(requiredStats, targetRatio, printer)
			search.SupplyData(mixedInputData)
			return search.Run(cancel)
		})
	}
	if runSearches {
		addTask("search1-accF", func() weight_types.WeightResult {
			search := weightfind.WeightSearcher1{}
			search.AccuracyStatistical = false
			search.Init(requiredStats, targetRatio, printer)
			search.SupplyData(mixedInputData)
			return search.Run(cancel)
		})
		addTask("search2-accF", func() weight_types.WeightResult {
			innerCancel := util_async.CancelSignal_Make()
			util_async.CancelAfterTimeout(innerCancel, time.Duration(shortTimeout)*time.Second, printer)
			util_async.ChainCancel(cancel, innerCancel)
			search := weightfind.WeightSearcher2{}
			search.AccuracyStatistical = false
			search.Init(requiredStats, targetRatio, printer)
			search.SupplyData(mixedInputData)
			search.SetRanges(-1.0, 10.0)
			return search.Run(innerCancel)
		})
		addTask("search3-accF", func() weight_types.WeightResult {
			search := weightfind.WeightSearcher3{}
			search.AccuracyStatistical = false
			search.Init(requiredStats, targetRatio)
			search.SupplyData(mixedInputData)
			search.SetRanges(-1.0, 10.0)
			return search.Run(cancel)
		})

		addTask("search1-accT", func() weight_types.WeightResult {
			search := weightfind.WeightSearcher1{}
			search.AccuracyStatistical = true
			search.Init(requiredStats, targetRatio, printer)
			search.SupplyData(mixedInputData)
			return search.Run(cancel)
		})
		addTask("search2-accT", func() weight_types.WeightResult {
			innerCancel := util_async.CancelSignal_Make()
			util_async.CancelAfterTimeout(innerCancel, time.Duration(shortTimeout)*time.Second, printer)
			util_async.ChainCancel(cancel, innerCancel)
			search := weightfind.WeightSearcher2{}
			search.AccuracyStatistical = true
			search.Init(requiredStats, targetRatio, printer)
			search.SupplyData(mixedInputData)
			search.SetRanges(-1.0, 10.0)
			return search.Run(innerCancel)
		})
		addTask("search3-accT", func() weight_types.WeightResult {
			search := weightfind.WeightSearcher3{}
			search.AccuracyStatistical = true
			search.Init(requiredStats, targetRatio)
			search.SupplyData(mixedInputData)
			search.SetRanges(-1.0, 10.0)
			return search.Run(cancel)
		})
	}

	if runRankingSep {
		label := "ranking-sep"
		addTask(label, func() weight_types.WeightResult {
			comp := weight_highs.RankingSeparatedWeights{}
			comp.Init(printer, standardTimeout)
			comp.SetRequiredStats(requiredStats, requiredSims)
			comp.SetTargetRatios(targetRatio)
			comp.SupplyData(mixedInputData)
			return comp.Run().WaitForResultOrPanic()
		})
	}
	if runFormula2 {
		label := "formula2"
		addTask(label, func() weight_types.WeightResult {
			comp := weight_highs.FormulaStatWeightProcess2{}
			// comp.BLEND if SetMinimumIncludeRate < 1.0
			comp.Init(printer)
			comp.SetRequiredStats(requiredStats)
			comp.SetTargetRatios(targetRatio)
			comp.SetMinimumIncludeRate(1.0)
			comp.SupplyData(mixedInputData)
			return comp.Run(standardTimeout).WaitForResultOrPanic()
		})
	}

	taskKeys := slices.Collect(maps.Keys(taskMap))
	util_collection.Shuffle(taskKeys)

	outputByAlgorithm := util_collection.MapConcurrent[string, weight_types.WeightResult]{}
	util_async.ForEach_Slice_Cancellable(5, taskKeys, cancel, func(taskLabel *string) {
		task := taskMap[*taskLabel]

		printer.Println("################# " + *taskLabel)
		weightResult := task()
		printer.Println("///////////////// " + *taskLabel)

		outputByAlgorithm.Put(*taskLabel, weightResult)
	})

	reportByAlgorithm := make(map[string]algorithmReport)
	for label, weightResult := range outputByAlgorithm.SeqWithKeys_ThreadSafeCopy() {
		accuracy := weightfind.EvaluateAccuracy(weightResult.Weight, requiredSims, &targetRatio, mixedInputDataFull)
		accuracyStat := weightfind.EvaluateAccuracyStatistical(weightResult.Weight, requiredSims, &targetRatio, mixedInputDataFull)
		weight1 := weightResult.AsWeight1()
		reportByAlgorithm[label] = algorithmReport{
			weight1,
			accuracy,
			accuracyStat,
			weightResult,
		}
	}

	resultOrder := slices.Collect(maps.Keys(reportByAlgorithm))

	slices.SortFunc(resultOrder, func(a, b string) int {
		return cmp.Compare(reportByAlgorithm[a].statAccuracy, reportByAlgorithm[b].statAccuracy)
	})
	compareReport(requiredStats, resultOrder, reportByAlgorithm, requiredSims, targetRatio, mixedInputDataFull, reportOnTweakedVersions, printer)

	slices.SortFunc(resultOrder, func(a, b string) int {
		return cmp.Compare(reportByAlgorithm[a].initialAccuracy, reportByAlgorithm[b].initialAccuracy)
	})
	compareReport(requiredStats, resultOrder, reportByAlgorithm, requiredSims, targetRatio, mixedInputDataFull, reportOnTweakedVersions, printer)
}

type algorithmReport struct {
	//weight3         *weight_types.Weight3ExtendedRanged
	//weight2         *weight_types.Weight2Extended
	weight1         *weight_types.Weight1Basic
	initialAccuracy float64
	statAccuracy    float64
	weightResult    weight_types.WeightResult
}

func compareReport(requiredStats []stats.StatType, resultOrder []string, reportByAlgorithm map[string]algorithmReport, requiredSims []stats.SimType, targetRatio weight_types.SimPriorityBasic, mixedInputDataFull []weight_types.WeightInput, reportOnTweakedVersions bool, printer *util.PrintRecorder) {
	tab := util.TabulateOutput{}
	tab.SetColumnSpacing(2)
	tab.AddColumnHeader("algo", false)
	for _, stat := range requiredStats {
		tab.AddColumnHeader(stat.Name(), true)
	}
	tab.AddColumnHeader("accX", false)
	tab.AddColumnHeader("accX_stat", false)
	tab.AddColumnHeader("acc1", false)
	tab.AddColumnHeader("acc1_stat", false)
	tab.AddColumnHeader("time", false)
	tab.AddColumnHeader("status", false)

	for _, label := range resultOrder {
		report := reportByAlgorithm[label]
		row := make([]string, 0, tab.ColumnCount())
		row = append(row, label)

		accuracy := 0.0
		accuracyStat := 0.0
		if report.weight1 != nil {
			for _, stat := range requiredStats {
				value := report.weight1.Get(stat)
				row = append(row, strconv.FormatFloat(value, 'f', 4, 64))
			}
			accuracy = weightfind.EvaluateAccuracy(report.weight1, requiredSims, &targetRatio, mixedInputDataFull)
			accuracyStat = weightfind.EvaluateAccuracyStatistical(report.weight1, requiredSims, &targetRatio, mixedInputDataFull)
		} else {
			for range requiredStats {
				row = append(row, "")
			}
		}

		row = append(row, strconv.FormatFloat(report.initialAccuracy, 'f', 4, 64))
		row = append(row, strconv.FormatFloat(report.statAccuracy, 'f', 4, 64))
		row = append(row, strconv.FormatFloat(accuracy, 'f', 4, 64))
		row = append(row, strconv.FormatFloat(accuracyStat, 'f', 4, 64))
		row = append(row, report.weightResult.SolveTime.String())
		row = append(row, report.weightResult.Status.String())
		tab.AddRow(row)

		//if reportOnTweakedVersions {
		//	weightTweak, _ := weightfind.WeightTweakerWithLogging(*report.weight1, requiredStats, &targetRatio, mixedInputDataFull, util.PrintRecorder_Nop())
		//	accuracyTweak := weightfind.EvaluateAccuracy(&weightTweak, requiredSims, &targetRatio, mixedInputDataFull)
		//	accuracyTweakStat := weightfind.EvaluateAccuracyStatistical(&weightTweak, requiredSims, &targetRatio, mixedInputDataFull)
		//	row = make([]string, 0)
		//	row = append(row, label)
		//	for _, stat := range requiredStats {
		//		value := weightTweak.Get(stat)
		//		row = append(row, strconv.FormatFloat(value, 'f', 4, 64))
		//	}
		//	row = append(row, strconv.FormatFloat(accuracyTweak, 'f', 4, 64))
		//	row = append(row, strconv.FormatFloat(accuracyTweakStat, 'f', 4, 64))
		//	row = append(row, report.weightResult.SolveTime.String())
		//	row = append(row, report.weightResult.Status.String())
		//	tab.AddRow(row)
		//}
	}
	tab.Write(printer)
}

type basicStatInput struct {
	IncrementStat  stats.StatType
	IncrementValue int32
	SimResult      stats.SimData
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
