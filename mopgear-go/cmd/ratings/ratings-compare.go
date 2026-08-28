package main

import (
	"cmp"
	"encoding/json/v2"
	"fmt"
	"maps"
	"os"
	"runtime/debug"
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
	"github.com/nerago/mopgear-go/weightfind/weight_highs/fitting3"
	"github.com/nerago/mopgear-go/weightfind/weight_highs/fitting4"
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

	// weight value 85.7928 - 89.6152
	weightsExample := weight_types.Weight1Basic_Make()
	weightsExample.Put(stats.Stat_Strength, 1.0000)
	weightsExample.Put(stats.Stat_Stamina, 0.8306046635)
	weightsExample.Put(stats.Stat_Crit, 1.4247130669)
	weightsExample.Put(stats.Stat_Haste, 0.3729476205)
	weightsExample.Put(stats.Stat_Expertise, 0.9410219636)
	weightsExample.Put(stats.Stat_Mastery, 0.3211543408)
	weightsExample.Put(stats.Stat_Dodge, 0.3759897596)
	weightsExample.Put(stats.Stat_Parry, 1.3238476723)

	cancel := util_async.CancelSignal_Make()
	util_async.CancelOnKeyPress(cancel)

	taskMap := make(map[string]func() weight_types.IWeightResult)
	addTask1 := func(label string, run func() weight_types.WeightResult1) {
		if taskMap[label] != nil {
			panic("duplicate")
		}
		taskMap[label] = func() weight_types.IWeightResult {
			return new(run())
		}
	}
	addTask2 := func(label string, run func() weight_types.WeightResult2) {
		if taskMap[label] != nil {
			panic("duplicate")
		}
		taskMap[label] = func() weight_types.IWeightResult {
			return new(run())
		}
	}
	addTask3 := func(label string, run func() weight_types.WeightResult3) {
		if taskMap[label] != nil {
			panic("duplicate")
		}
		taskMap[label] = func() weight_types.IWeightResult {
			return new(run())
		}
	}

	reportOnTweakedVersions := false
	//shortTimeout := 20
	shortTimeout := 1800
	standardTimeout := shortTimeout
	//
	//runBasic := true
	//runFormulaVariants := true // best is about 87%, moderate time
	//runFitting1 := true        // slow, low 90%
	//runFitting2 := true
	//runFitting34 := true
	//
	//runGrid1Original := true
	//runGrid1Variants := false
	//runGrid1VariantsFewer := false
	//runGrid1C := false
	//runGrid2 := true
	//
	//runRankingOlder := false
	//runRanking3aPreferred := false // broken
	//runRanking3aVariants := false  // broken
	//runRanking3bVariants := false
	//runRanking3bPreferred := true
	//runRanking3c := false
	//runRanking4 := false // still a little slow, midrange 94% etc
	//runRanking5 := false // excellent but slow
	//
	//runSearches := false
	//runSearch0 := false
	//
	//runRankingSep := true
	//runFormula2 := true

	runBasic := true
	runFormulaVariants := false // best is about 87%, moderate time
	runFitting1 := true         // slow, low 90%
	runFitting2 := true
	runFitting34 := true

	runGrid1Original := true
	runGrid1Variants := true
	runGrid1VariantsFewer := false
	runGrid1C := true
	runGrid2 := true

	runRankingOlder := true
	runRanking3aPreferred := true // broken
	runRanking3aVariants := false // broken
	runRanking3bVariants := true
	runRanking3bPreferred := true
	runRanking3c := true
	runRanking4 := true // still a little slow, midrange 94% etc
	runRanking5 := true // excellent but slow

	runSearches := true
	runSearch0 := true

	runRankingSep := true
	runFormula2 := true

	addTask1("itemLevel", func() weight_types.WeightResult1 {
		weight := weight_types.Weight1Basic_Make()
		weight.Put(requiredStats[0], 1)
		return weight_types.WeightResult1Make(
			&weight,
			0,
			highs.ModelStatusOptimal,
		)
	})

	if runBasic {
		addTask1("basic", func() weight_types.WeightResult1 {
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

		addTask2("form", func() weight_types.WeightResult2 {
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
			addTask2(label, func() weight_types.WeightResult2 {
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
			addTask2(label, func() weight_types.WeightResult2 {
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
		addTask3("fitting", func() weight_types.WeightResult3 {
			fitting := fitting1.FittingEachStatWeightProcess{}
			fitting.Init(printer, shortTimeout)
			fitting.SetRequiredStats(requiredStats, requiredSims)
			fitting.SetOnlyComputeSingleSegmentEach(true)
			fitting.SupplyData(inputDataRandom)
			return fitting.Run(cancel)
		})
	}
	if runFitting2 {
		addTask3("fitting2", func() weight_types.WeightResult3 {
			fitting := fitting2.FittingEachStatWeightProcess2{}
			fitting.Init(3, printer, shortTimeout)
			fitting.SetRequiredStats(requiredStats, requiredSims)
			fitting.SetTargetRatios(targetRatio)
			fitting.SupplyData(inputDataRandom)
			return fitting.Run(cancel)
		})
	}
	if runFitting34 {
		addTask3("fitting3", func() weight_types.WeightResult3 {
			fitting := fitting3.FittingEachStatWeightProcess3{}
			fitting.Init(3, printer, shortTimeout)
			fitting.SetRequiredStats(requiredStats, requiredSims)
			fitting.SetTargetRatios(targetRatio)
			fitting.SupplyData(inputDataRandom)
			return fitting.Run(cancel, util.TrackProgress_Nop())
		})
		addTask3("fitting4", func() weight_types.WeightResult3 {
			fitting := fitting4.FittingEachStatWeightProcess4{}
			fitting.Init(3, printer, shortTimeout)
			fitting.SetRequiredStats(requiredStats, requiredSims)
			fitting.SetTargetRatios(targetRatio)
			fitting.SupplyData(inputDataRandom)
			return fitting.Run(cancel)
		})
	}

	if runGrid1Original {
		addTask1("grid1", func() weight_types.WeightResult1 {
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
		addTask1("grid1-1", func() weight_types.WeightResult1 {
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
							addTask1(label, func() weight_types.WeightResult1 {
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
			addTask1(label, func() weight_types.WeightResult1 {
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
				addTask1(label, func() weight_types.WeightResult1 {
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
		addTask2("grid2-1", func() weight_types.WeightResult2 {
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
		addTask2("grid2-2", func() weight_types.WeightResult2 {
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
		addTask2("grid2-12", func() weight_types.WeightResult2 {
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
		addTask2("grid2-123", func() weight_types.WeightResult2 {
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
				addTask1(label, func() weight_types.WeightResult1 {
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
			addTask1(label, func() weight_types.WeightResult1 {
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
			addTask1(label, func() weight_types.WeightResult1 {
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
		addTask1(label, func() weight_types.WeightResult1 {
			ranking := weight_highs.RankingStatWeightProcess3{}
			ranking.ALGO = 1
			ranking.SCALE1 = false
			ranking.Init(printer, standardTimeout)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(mixedInputData))
			weightFuture := ranking.RunUsingExternalStart(weightsExample)
			util_async.ChainCancel(cancel, weightFuture)
			return weightFuture.WaitForResultOrNilValue()
		})
		label = fmt.Sprintf("ranking3a-true-1")
		addTask1(label, func() weight_types.WeightResult1 {
			ranking := weight_highs.RankingStatWeightProcess3{}
			ranking.ALGO = 1
			ranking.SCALE1 = true
			ranking.Init(printer, standardTimeout)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(mixedInputData))
			weightFuture := ranking.RunUsingExternalStart(weightsExample)
			util_async.ChainCancel(cancel, weightFuture)
			return weightFuture.WaitForResultOrNilValue()
		})
	}

	if runRanking3bVariants {
		for TOTALWEIGHT := range 3 {
			for ALGO := range 2 {
				label := fmt.Sprintf("ranking3b-%d-%d", TOTALWEIGHT, ALGO)
				addTask1(label, func() weight_types.WeightResult1 {
					ranking := weight_highs.RankingStatWeightProcess3b{}
					ranking.TOTALWEIGHT = TOTALWEIGHT
					ranking.ALGO = ALGO
					ranking.Init(printer, shortTimeout)
					ranking.SetRequiredStats(requiredStats)
					ranking.SetTargetRatios(targetRatio)
					ranking.SupplyData(slices.Clone(mixedInputData))
					weightFuture := ranking.RunSinglePassFromExternal(weightsExample)
					util_async.ChainCancel(cancel, weightFuture)
					return weightFuture.WaitForResultOrNilValue()
				})
			}
		}
	}
	if runRanking3bPreferred {
		label := fmt.Sprintf("ranking3b-pref")
		addTask1(label, func() weight_types.WeightResult1 {
			ranking := weight_highs.RankingStatWeightProcess3b{}
			ranking.TOTALWEIGHT = 0
			ranking.ALGO = 1
			ranking.Init(printer, shortTimeout)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(mixedInputData))
			weightFuture := ranking.RunSinglePassFromExternal(weightsExample)
			util_async.ChainCancel(cancel, weightFuture)
			return weightFuture.WaitForResultOrNilValue()
		})
	}

	if runRanking3c {
		addTask1("ranking3c", func() weight_types.WeightResult1 {
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
			addTask1(label, func() weight_types.WeightResult1 {
				ranking := weight_highs.RankingStatWeightProcess4{}
				ranking.MULTIPLY = MULTIPLY
				ranking.Init(printer)
				ranking.SetRequiredStats(requiredStats)
				ranking.SetTargetRatios(targetRatio)
				ranking.SupplyData(slices.Clone(inputDataRandom))
				weightFuture := ranking.RunUsingExternalStart(weightsExample, standardTimeout)
				util_async.ChainCancel(cancel, weightFuture)
				return weightFuture.WaitForResultOrNilValue()
			})
		}
	}

	if runRanking5 {
		addTask1("ranking5-0-1", func() weight_types.WeightResult1 {
			ranking := weight_highs.RankingStatWeightProcess5{}
			ranking.WEIGHTSUM = 0
			ranking.SIMRANK = 1
			ranking.Init(printer)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(inputDataRandom))
			futureWeight := ranking.Run(standardTimeout, 60) // note restricted data sample
			return futureWeight.WaitForResultOrNilValue()
		})
		addTask1("ranking5-1-1", func() weight_types.WeightResult1 {
			ranking := weight_highs.RankingStatWeightProcess5{}
			ranking.WEIGHTSUM = 1
			ranking.SIMRANK = 1
			ranking.Init(printer)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(inputDataRandom))
			futureWeight := ranking.Run(standardTimeout, 60) // note restricted data sample
			return futureWeight.WaitForResultOrNilValue()
		})
		addTask1("ranking5-0-2", func() weight_types.WeightResult1 {
			ranking := weight_highs.RankingStatWeightProcess5{}
			ranking.WEIGHTSUM = 0
			ranking.SIMRANK = 2
			ranking.Init(printer)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(inputDataRandom))
			futureWeight := ranking.Run(standardTimeout, 60) // note restricted data sample
			return futureWeight.WaitForResultOrNilValue()
		})
		addTask1("ranking5-1-2", func() weight_types.WeightResult1 {
			ranking := weight_highs.RankingStatWeightProcess5{}
			ranking.WEIGHTSUM = 1
			ranking.SIMRANK = 2
			ranking.Init(printer)
			ranking.SetRequiredStats(requiredStats)
			ranking.SetTargetRatios(targetRatio)
			ranking.SupplyData(slices.Clone(inputDataRandom))
			futureWeight := ranking.Run(standardTimeout, 60) // note restricted data sample
			return futureWeight.WaitForResultOrNilValue()
		})
	}

	if runSearch0 {
		addTask1("search0-accF", func() weight_types.WeightResult1 {
			search := weightfind.WeightSearcher0{}
			search.AccuracyStatistical = false
			search.Init(requiredStats, targetRatio, printer)
			search.SupplyData(mixedInputData)
			return search.Run(cancel)
		})
		addTask1("search0-accT", func() weight_types.WeightResult1 {
			search := weightfind.WeightSearcher0{}
			search.AccuracyStatistical = true
			search.Init(requiredStats, targetRatio, printer)
			search.SupplyData(mixedInputData)
			return search.Run(cancel)
		})
	}
	if runSearches {
		addTask1("search1-accF", func() weight_types.WeightResult1 {
			search := weightfind.WeightSearcher1{}
			search.AccuracyStatistical = false
			search.Init(requiredStats, targetRatio, printer)
			search.SupplyData(mixedInputData)
			return search.Run(cancel)
		})
		addTask1("search2-accF", func() weight_types.WeightResult1 {
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
		addTask1("search3-accF", func() weight_types.WeightResult1 {
			search := weightfind.WeightSearcher3{}
			search.AccuracyStatistical = false
			search.Init(requiredStats, targetRatio)
			search.SupplyData(mixedInputData)
			search.SetRanges(-1.0, 10.0)
			return search.Run(cancel)
		})

		addTask1("search1-accT", func() weight_types.WeightResult1 {
			search := weightfind.WeightSearcher1{}
			search.AccuracyStatistical = true
			search.Init(requiredStats, targetRatio, printer)
			search.SupplyData(mixedInputData)
			return search.Run(cancel)
		})
		addTask1("search2-accT", func() weight_types.WeightResult1 {
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
		addTask1("search3-accT", func() weight_types.WeightResult1 {
			search := weightfind.WeightSearcher3{}
			search.AccuracyStatistical = true
			search.Init(requiredStats, targetRatio)
			search.SupplyData(mixedInputData)
			search.SetRanges(-1.0, 10.0)
			return search.Run(cancel)
		})
	}

	if runRankingSep {
		addTask2("ranking-sep", func() weight_types.WeightResult2 {
			comp := weight_highs.RankingSeparatedWeights{}
			comp.Init(printer, standardTimeout)
			comp.SetRequiredStats(requiredStats, requiredSims)
			comp.SetTargetRatios(targetRatio)
			comp.SupplyData(mixedInputData)
			return comp.Run().WaitForResultOrPanic()
		})
		addTask2("ranking-sep2", func() weight_types.WeightResult2 {
			comp := weight_highs.RankingSeparatedWeights2{}
			comp.Init(printer, standardTimeout)
			comp.SetRequiredStats(requiredStats, requiredSims)
			comp.SetTargetRatios(targetRatio)
			comp.SupplyData(mixedInputData)
			sw := util.StopwatchMakeStopped()
			weight2 := comp.Run(sw).WaitForResultOrPanic()
			return weight_types.WeightResult2Make(&weight2, sw.Elapsed(), highs.ModelStatusUnknown)
		})
	}
	if runFormula2 {
		addTask2("formula2", func() weight_types.WeightResult2 {
			comp := weight_highs.FormulaStatWeightProcess2{}
			// comp.BLEND if SetMinimumIncludeRate < 1.0
			comp.Init(printer)
			comp.SetRequiredStats(requiredStats)
			comp.SetTargetRatios(targetRatio)
			comp.SetMinimumIncludeRate(1.0)
			comp.SupplyData(mixedInputData)
			return comp.Run(standardTimeout).WaitForResultOrPanic()
		})
		for BLEND := range 6 {
			label := fmt.Sprintf("formula2-%d", BLEND)
			addTask2(label, func() weight_types.WeightResult2 {
				comp := weight_highs.FormulaStatWeightProcess2{}
				comp.BLEND = BLEND
				comp.Init(printer)
				comp.SetRequiredStats(requiredStats)
				comp.SetTargetRatios(targetRatio)
				comp.SetMinimumIncludeRate(0.7)
				comp.SupplyData(mixedInputData)
				return comp.Run(standardTimeout).WaitForResultOrPanic()
			})
		}
	}

	taskKeys := slices.Collect(maps.Keys(taskMap))
	util_collection.Shuffle(taskKeys)

	outputByAlgorithm := util_collection.MapConcurrent[string, weight_types.IWeightResult]{}
	util_async.ForEach_Slice_Cancellable(5, taskKeys, cancel, func(taskLabel *string) {
		defer func() {
			if x := recover(); x != nil {
				printer.Println("!!!!!!!!!!!!!!!! " + *taskLabel)
				debug.PrintStack()
				outputByAlgorithm.Put(*taskLabel, new(weight_types.WeightResult1Make(nil, 0, highs.ModelStatusSolveError)))
			}
		}()

		task := taskMap[*taskLabel]

		printer.Println("################# " + *taskLabel)
		weightResult := task()
		printer.Println("///////////////// " + *taskLabel)

		outputByAlgorithm.Put(*taskLabel, weightResult)
	})

	reportForWeight1(&outputByAlgorithm, mixedInputDataFull, targetRatio, requiredSims, requiredStats, reportOnTweakedVersions, printer)
	reportForWeight2(&outputByAlgorithm, mixedInputDataFull, targetRatio, requiredSims, requiredStats, printer)
	reportForWeight3(&outputByAlgorithm, mixedInputDataFull, targetRatio, requiredSims, requiredStats, printer)
}

func reportForWeight1(outputByAlgorithm *util_collection.MapConcurrent[string, weight_types.IWeightResult], mixedInputDataFull []weight_types.WeightInput, priorityBasic weight_types.SimPriorityBasic, requiredSims []stats.SimType, requiredStats []stats.StatType, reportOnTweakedVersions bool, printer *util.PrintRecorder) {
	reportByAlgorithm := make(map[string]algorithmReport1)
	for label, weightResult := range outputByAlgorithm.SeqWithKeys_ThreadSafeCopy() {
		weight1 := weightResult.AsWeight1(mixedInputDataFull)
		accuracy := weightfind.EvaluateAccuracyBasic(weight1, requiredSims, &priorityBasic, mixedInputDataFull)
		accuracyStat := weightfind.EvaluateAccuracyStatisticalExtended(weight1, requiredSims, &priorityBasic, mixedInputDataFull)
		reportByAlgorithm[label] = algorithmReport1{
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
	printer.Println("<<<<<<< WEIGHT1 BY stat accuracy >>>>>>>")
	compareReport1(requiredStats, resultOrder, reportByAlgorithm, requiredSims, priorityBasic, mixedInputDataFull, reportOnTweakedVersions, printer)
	printer.Println0()
	printer.Println0()

	slices.SortFunc(resultOrder, func(a, b string) int {
		return cmp.Compare(reportByAlgorithm[a].basicAccuracy, reportByAlgorithm[b].basicAccuracy)
	})
	printer.Println("<<<<<<< WEIGHT1 BY basic accuracy >>>>>>>")
	compareReport1(requiredStats, resultOrder, reportByAlgorithm, requiredSims, priorityBasic, mixedInputDataFull, reportOnTweakedVersions, printer)
	printer.Println0()
	printer.Println0()
}

type algorithmReport1 struct {
	weight1       *weight_types.Weight1Basic
	basicAccuracy float64
	statAccuracy  float64
	weightResult  weight_types.IWeightResult
}

func compareReport1(requiredStats []stats.StatType, resultOrder []string, reportByAlgorithm map[string]algorithmReport1, requiredSims []stats.SimType, targetRatio weight_types.SimPriorityBasic, mixedInputDataFull []weight_types.WeightInput, reportOnTweakedVersions bool, printer *util.PrintRecorder) {
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
			accuracy = weightfind.EvaluateAccuracyBasic(report.weight1, requiredSims, &targetRatio, mixedInputDataFull)
			accuracyStat = weightfind.EvaluateAccuracyStatisticalExtended(report.weight1, requiredSims, &targetRatio, mixedInputDataFull)
		} else {
			for range requiredStats {
				row = append(row, "")
			}
		}

		row = append(row, strconv.FormatFloat(report.basicAccuracy, 'f', 4, 64))
		row = append(row, strconv.FormatFloat(report.statAccuracy, 'f', 4, 64))
		row = append(row, strconv.FormatFloat(accuracy, 'f', 4, 64))
		row = append(row, strconv.FormatFloat(accuracyStat, 'f', 4, 64))
		row = append(row, report.weightResult.GetSolveTime().String())
		row = append(row, report.weightResult.GetStatus().String())
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

type algorithmReport2 struct {
	weight2       *weight_types.Weight2Extended
	basicAccuracy float64
	statAccuracy  float64
	weightResult  weight_types.IWeightResult
}
type algorithmReport3 struct {
	weight3       *weight_types.Weight3ExtendedRanged
	basicAccuracy float64
	statAccuracy  float64
	weightResult  weight_types.IWeightResult
}

func reportForWeight2(outputByAlgorithm *util_collection.MapConcurrent[string, weight_types.IWeightResult], mixedInputDataFull []weight_types.WeightInput, priorityBasic weight_types.SimPriorityBasic, requiredSims []stats.SimType, requiredStats []stats.StatType, printer *util.PrintRecorder) {
	reportByAlgorithm := make(map[string]algorithmReport2)
	for label, weightResult := range outputByAlgorithm.SeqWithKeys_ThreadSafeCopy() {
		weight2 := weightResult.AsWeight2(mixedInputDataFull)
		if weight2 != nil {
			accuracy := weightfind.EvaluateAccuracyBasic(weight2, requiredSims, &priorityBasic, mixedInputDataFull)
			accuracyStat := weightfind.EvaluateAccuracyStatisticalExtended(weight2, requiredSims, &priorityBasic, mixedInputDataFull)
			reportByAlgorithm[label] = algorithmReport2{weight2, accuracy, accuracyStat, weightResult}
		}
	}

	resultOrder := slices.Collect(maps.Keys(reportByAlgorithm))
	slices.SortFunc(resultOrder, func(a, b string) int {
		return cmp.Compare(
			max(reportByAlgorithm[a].basicAccuracy, reportByAlgorithm[a].statAccuracy),
			max(reportByAlgorithm[b].basicAccuracy, reportByAlgorithm[b].statAccuracy),
		)
	})

	printer.Println("<<<<<<< WEIGHT2 BY best accuracy >>>>>>>")
	compareReport2(reportByAlgorithm, resultOrder, printer)
	printer.Println0()
	printer.Println0()
}

func compareReport2(reportByAlgorithm map[string]algorithmReport2, reportOrder []string, printer *util.PrintRecorder) {
	tab := util.TabulateOutput{}
	tab.SetColumnSpacing(2)
	tab.AddColumnHeader("algo", false)
	tab.AddColumnHeader("acc_basic", false)
	tab.AddColumnHeader("acc_stat", false)
	tab.AddColumnHeader("time", false)
	tab.AddColumnHeader("status", false)

	for _, label := range reportOrder {
		report := reportByAlgorithm[label]
		row := make([]string, 0, tab.ColumnCount())
		row = append(row, label)
		row = append(row, strconv.FormatFloat(report.basicAccuracy, 'f', 4, 64))
		row = append(row, strconv.FormatFloat(report.statAccuracy, 'f', 4, 64))
		row = append(row, report.weightResult.GetSolveTime().String())
		row = append(row, report.weightResult.GetStatus().String())
		tab.AddRow(row)
	}
	tab.Write(printer)
}
func reportForWeight3(outputByAlgorithm *util_collection.MapConcurrent[string, weight_types.IWeightResult], mixedInputDataFull []weight_types.WeightInput, priorityBasic weight_types.SimPriorityBasic, requiredSims []stats.SimType, requiredStats []stats.StatType, printer *util.PrintRecorder) {
	reportByAlgorithm := make(map[string]algorithmReport3)
	for label, weightResult := range outputByAlgorithm.SeqWithKeys_ThreadSafeCopy() {
		weight3 := weightResult.AsWeight3(mixedInputDataFull)
		if weight3 != nil {
			accuracy := weightfind.EvaluateAccuracyBasic(weight3, requiredSims, &priorityBasic, mixedInputDataFull)
			accuracyStat := weightfind.EvaluateAccuracyStatisticalExtended(weight3, requiredSims, &priorityBasic, mixedInputDataFull)
			reportByAlgorithm[label] = algorithmReport3{weight3, accuracy, accuracyStat, weightResult}
		}
	}

	resultOrder := slices.Collect(maps.Keys(reportByAlgorithm))
	slices.SortFunc(resultOrder, func(a, b string) int {
		return cmp.Compare(
			max(reportByAlgorithm[a].basicAccuracy, reportByAlgorithm[a].statAccuracy),
			max(reportByAlgorithm[b].basicAccuracy, reportByAlgorithm[b].statAccuracy),
		)
	})

	printer.Println("<<<<<<< WEIGHT3 BY best accuracy >>>>>>>")
	compareReport3(reportByAlgorithm, resultOrder, printer)
	printer.Println0()
	printer.Println0()
}

func compareReport3(reportByAlgorithm map[string]algorithmReport3, reportOrder []string, printer *util.PrintRecorder) {
	tab := util.TabulateOutput{}
	tab.SetColumnSpacing(2)
	tab.AddColumnHeader("algo", false)
	tab.AddColumnHeader("acc_basic", false)
	tab.AddColumnHeader("acc_stat", false)
	tab.AddColumnHeader("time", false)
	tab.AddColumnHeader("status", false)

	for _, label := range reportOrder {
		report := reportByAlgorithm[label]
		row := make([]string, 0, tab.ColumnCount())
		row = append(row, label)
		row = append(row, strconv.FormatFloat(report.basicAccuracy, 'f', 4, 64))
		row = append(row, strconv.FormatFloat(report.statAccuracy, 'f', 4, 64))
		row = append(row, report.weightResult.GetSolveTime().String())
		row = append(row, report.weightResult.GetStatus().String())
		tab.AddRow(row)
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
