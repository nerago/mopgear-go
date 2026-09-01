package updateProc

import (
	"fmt"
	"slices"
	"time"

	"github.com/bartolsthoorn/gohighs/highs"
	"github.com/nerago/mopgear-go/files"
	"github.com/nerago/mopgear-go/tools"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/weightfind"
	"github.com/nerago/mopgear-go/weightfind/weight_highs"
	"github.com/nerago/mopgear-go/weightfind/weight_highs/fitting3"
	"github.com/nerago/mopgear-go/weightfind/weight_highs/fitting4"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type solves struct {
	cancel      util_async.CancelSignal
	tracker     *util.TrackProgress
	timeoutEach int
	printer     *util.PrintRecorder

	input  *updateInputs
	output *choiceOutput

	taskPool *util_async.NestedTaskPoolChild
}

type solveTask func()

func (sol *solves) startSolvers() {
	// FORMULA2 WEIGHTS - MIP
	sol.solveFormulaWeight()

	// FITTING - Slow MIP
	sol.solveFittingWeight()
	sol.solveFittingFast()

	// GRID WEIGHTS - GPU*2 - later for less contention
	sol.solveGridWeights()

	sol.solveGrid2Weights()

	// SEARCH weights - Non-Highs
	for searchMode := range 3 {
		sol.solveSearchWeights(searchMode)
	}

	// RANKING WEIGHTS - simplex*2, IPX*2
	sol.solveRankingWeight()
}

func (sol *solves) addTask(task solveTask) {
	sol.taskPool.Go(func() (err error) {
		defer func() {
			anyMessage := recover()
			err = util.ErrorFromAny(anyMessage)
		}()
		sol.printer.Println("SOLVE TASK start")
		task()
		sol.printer.Println("SOLVE TASK end")
		return nil
	})
}

func (sol *solves) solveGridWeights() {
	for gridOutlierSetting := range 2 {
		sol.addTask(func() {
			gridData := sol.input.dataGrid
			if c_useSamplingGrid1 {
				gridData = util_collection.SliceSampleFromStart(gridData, c_dataSampleGrid)
			}

			choiceName := fmt.Sprintf("GRID1-%d", gridOutlierSetting)

			grid := weight_highs.GridStatWeightProcess1B{}
			grid.OUTLIER = gridOutlierSetting
			grid.SCALEMODE = 1
			grid.ROUNDMODE = 2
			grid.CALCMODE = 2
			grid.Init(sol.printer, sol.timeoutEach)
			grid.SetTargetRatios(sol.input.targetRatio)
			grid.SetRequiredStats(sol.input.statTypes)
			grid.SupplyData(gridData)
			weightsFuture, err := grid.Run()
			sol.handleFuture1OrError(choiceName, weightsFuture, err)
		})
	}
}

func (sol *solves) solveGrid2Weights() {
	for groups := range 4 {
		sol.addTask(func() {
			choiceName := fmt.Sprintf("GRID2-%d", groups)

			grid2 := weight_highs.GridStatWeightProcess2{}
			switch groups {
			case 0:
				grid2.IncludeDiffs1 = true
			case 1:
				grid2.IncludeDiffs2 = true
			case 2:
				grid2.IncludeDiffs1 = true
				grid2.IncludeDiffs2 = true
			default:
				grid2.IncludeDiffs1 = true
				grid2.IncludeDiffs2 = true
				grid2.IncludeDiffs3 = true
			}
			grid2.Init(sol.printer, sol.timeoutEach)
			grid2.SetRequiredStats(sol.input.statTypes)
			grid2.SetTargetRatios(sol.input.targetRatio)
			grid2.SupplyData(slices.Clone(sol.input.dataGrid))
			weightsFuture, err := grid2.Run()
			sol.handleFuture2OrError(choiceName, weightsFuture, err)
		})
	}
}

func (sol *solves) solveRankingWeight() {
	rankData := sol.input.dataAll

	sol.addTask(func() {
		ranking := weight_highs.RankingStatWeightProcess3c{}
		ranking.Init(sol.printer, sol.timeoutEach)
		ranking.SetRequiredStats(sol.input.statTypes)
		ranking.SetTargetRatios(sol.input.targetRatio)
		ranking.SupplyData(rankData)
		weightsFuture, err := ranking.RunMultiRound()
		sol.handleFuture1OrError("RANK3C", weightsFuture, err)
	})

	sol.addTask(func() {
		ranking := weight_highs.RankingStatWeightProcess3b{}
		ranking.TOTALWEIGHT = 2
		ranking.ALGO = 0
		ranking.Init(sol.printer, sol.timeoutEach)
		ranking.SetRequiredStats(sol.input.statTypes)
		ranking.SetTargetRatios(sol.input.targetRatio)
		ranking.SupplyData(rankData)
		var weightsFuture *util_async.FutureCancellable[weight_types.WeightResult1]
		var err error
		if bestWeightsSoFar, hasBest := sol.output.bestWeightChoice1(); hasBest {
			weightsFuture, err = ranking.RunSinglePassFromExternal(bestWeightsSoFar.weight1)
		} else {
			weightsFuture, err = ranking.RunMultiRound()
		}
		sol.handleFuture1OrError("RANK3-2-0", weightsFuture, err)
	})

	sol.addTask(func() {
		ranking := weight_highs.RankingStatWeightProcess3b{}
		ranking.TOTALWEIGHT = 2
		ranking.ALGO = 1
		ranking.Init(sol.printer, sol.timeoutEach)
		ranking.SetRequiredStats(sol.input.statTypes)
		ranking.SetTargetRatios(sol.input.targetRatio)
		ranking.SupplyData(rankData)
		var weightsFuture *util_async.FutureCancellable[weight_types.WeightResult1]
		var err error
		if bestWeightsSoFar, hasBest := sol.output.bestWeightChoice1(); hasBest {
			weightsFuture, err = ranking.RunSinglePassFromExternal(bestWeightsSoFar.weight1)
		} else {
			weightsFuture, err = ranking.RunMultiRound()
		}
		sol.handleFuture1OrError("RANK3-2-1", weightsFuture, err)
	})

	sol.addTask(func() {
		ranking := weight_highs.RankingStatWeightProcess{}
		ranking.RANKMODE = 0
		ranking.WEIGHTSUM = 0
		ranking.Init(sol.printer)
		ranking.SetRequiredStats(sol.input.statTypes)
		ranking.SetTargetRatios(sol.input.targetRatio)
		ranking.SupplyData(rankData)
		weightsFuture, err := ranking.Run(sol.timeoutEach)
		sol.handleFuture1OrError("RANK1-0", weightsFuture, err)
	})

	sol.addTask(func() {
		ranking := weight_highs.RankingStatWeightProcess{}
		ranking.RANKMODE = 0
		ranking.WEIGHTSUM = 1
		ranking.Init(sol.printer)
		ranking.SetRequiredStats(sol.input.statTypes)
		ranking.SetTargetRatios(sol.input.targetRatio)
		ranking.SupplyData(rankData)
		weightsFuture, err := ranking.Run(sol.timeoutEach)
		sol.handleFuture1OrError("RANK1-1", weightsFuture, err)
	})

	sol.addTask(func() {
		if c_useSamplingRank4 {
			rankData = util_collection.SliceSampleRandom(rankData, c_dataSampleFitRank)
		}
		ranking := weight_highs.RankingStatWeightProcess4{}
		ranking.MULTIPLY = 0
		ranking.Init(sol.printer)
		ranking.SetRequiredStats(sol.input.statTypes)
		ranking.SetTargetRatios(sol.input.targetRatio)
		ranking.SupplyData(rankData)
		if bestWeightSoFar, hasBest := sol.output.bestWeightChoice1(); hasBest {
			existing1 := bestWeightSoFar.weight1
			weightsFuture, err := ranking.RunUsingExternalStart(existing1, sol.timeoutEach)
			sol.handleFuture1OrError("RANK4", weightsFuture, err)
		}
	})
}

func (sol *solves) solveFormulaWeight() {
	sol.addTask(func() {
		comp := weight_highs.FormulaStatWeightProcess2{}
		comp.Init(sol.printer)
		comp.SetRequiredStats(sol.input.statTypes)
		comp.SetTargetRatios(sol.input.targetRatio)
		comp.SetMinimumIncludeRate(1.0)
		comp.SupplyData(sol.input.dataAll)
		weights2Future, err := comp.Run(sol.timeoutEach)
		sol.handleFuture2OrError("FORM2", weights2Future, err)
	})

	sol.addTask(func() {
		compB := weight_highs.FormulaStatWeightProcess2{}
		compB.BLEND = 3
		compB.Init(sol.printer)
		compB.SetRequiredStats(sol.input.statTypes)
		compB.SetTargetRatios(sol.input.targetRatio)
		compB.SetMinimumIncludeRate(0.7)
		if c_useSamplingFormMIP {
			compB.SupplyData(util_collection.SliceSampleRandom(sol.input.dataAll, c_dataSampleFitRank))
		} else {
			compB.SupplyData(sol.input.dataAll)
		}
		weights2FutureB, err := compB.Run(sol.timeoutEach)
		sol.handleFuture2OrError("FORM2-70", weights2FutureB, err)
	})
}

func (sol *solves) solveFittingWeight() {
	fitTimeout := sol.timeoutEach / 8

	sol.addTask(func() {
		fit1 := fitting3.FittingEachStatWeightProcess3{}
		fit1.Init(4, sol.printer, fitTimeout)
		fit1.SetRequiredStats(sol.input.statTypes, sol.input.simTypes)
		fit1.SetTargetRatios(sol.input.targetRatio)
		fit1.SupplyData(sol.input.dataFit)
		res1 := fit1.Run(sol.cancel, sol.tracker)
		sol.output.evaluateWeightResult3("FITTING3-dataFit", &res1)
	})

	sol.addTask(func() {
		fit2 := fitting3.FittingEachStatWeightProcess3{}
		fit2.Init(3, sol.printer, fitTimeout)
		fit2.SetRequiredStats(sol.input.statTypes, sol.input.simTypes)
		fit2.SetTargetRatios(sol.input.targetRatio)
		fit2.SupplyData(slices.Concat(sol.input.dataFit, sol.input.dataGrid))
		res2 := fit2.Run(sol.cancel, sol.tracker)
		sol.output.evaluateWeightResult3("FITTING3-dataGridFit", &res2)
	})

	sol.addTask(func() {
		fit3 := fitting3.FittingEachStatWeightProcess3{}
		fit3.Init(3, sol.printer, fitTimeout)
		fit3.SetRequiredStats(sol.input.statTypes, sol.input.simTypes)
		fit3.SetTargetRatios(sol.input.targetRatio)
		fit3.SupplyData(sol.input.dataAll)
		res3 := fit3.Run(sol.cancel, sol.tracker)
		sol.output.evaluateWeightResult3("FITTING3-dataAll", &res3)
	})
}

func (sol *solves) solveFittingFast() {
	for segments := 2; segments <= 4; segments++ {
		sol.addTask(func() {
			fit4data := fitting4.FittingEachStatWeightProcess4{}
			fit4data.SegmentOnData = true
			fit4data.Init(segments, sol.printer, sol.timeoutEach)
			fit4data.SetRequiredStats(sol.input.statTypes, sol.input.simTypes)
			fit4data.SetTargetRatios(sol.input.targetRatio)
			fit4data.SupplyData(sol.input.dataAll)
			weights3data := fit4data.Run(sol.cancel)
			label := fmt.Sprintf("FITTING4-data-%d", segments)
			sol.output.evaluateWeightResult3(label, &weights3data)
		})

		sol.addTask(func() {
			fit4stat := fitting4.FittingEachStatWeightProcess4{}
			fit4stat.SegmentOnData = false
			fit4stat.Init(segments, sol.printer, sol.timeoutEach)
			fit4stat.SetRequiredStats(sol.input.statTypes, sol.input.simTypes)
			fit4stat.SetTargetRatios(sol.input.targetRatio)
			fit4stat.SupplyData(sol.input.dataAll)
			weights3stat := fit4stat.Run(sol.cancel)
			label := fmt.Sprintf("FITTING4-stat-%d", segments)
			sol.output.evaluateWeightResult3(label, &weights3stat)
		})
	}
}

func (sol *solves) makeCanceller() (*util_async.CancelSignalBasic, *time.Timer) {
	innerCancel := util_async.CancelSignal_Make()
	timer := util_async.CancelAfterTimeout(innerCancel, time.Second*time.Duration(sol.timeoutEach), sol.printer)
	_ = util_async.ChainCancel(sol.cancel, innerCancel)
	return innerCancel, timer
}

func (sol *solves) solveSearchWeights(searchMode int) {
	sol.addTask(func() {
		innerCancel, timer := sol.makeCanceller()
		defer timer.Stop()

		search := weightfind.WeightSearcher2{}
		search.AccuracyStatistical = false
		search.Init(sol.input.statTypes, sol.input.targetRatio, sol.printer)
		search.SupplyData(sol.input.dataAll)
		search.SetRanges(-1.0, 10.0)
		weightResult := search.Run(innerCancel)
		sol.output.evaluateWeightResult1("SEARCH2", &weightResult)
	})

	sol.addTask(func() {
		innerCancel, timer := sol.makeCanceller()
		defer timer.Stop()

		search := weightfind.WeightSearcher3{}
		search.AccuracyStatistical = true
		search.Init(sol.input.statTypes, sol.input.targetRatio)
		search.SupplyData(sol.input.dataAll)
		search.SetRanges(-1.0, 10.0)
		weightResult := search.Run(innerCancel)
		sol.output.evaluateWeightResult1("SEARCH3", &weightResult)
	})

	sol.addTask(func() {
		innerCancel, timer := sol.makeCanceller()
		defer timer.Stop()

		sw := util.StopwatchMakeStarted()
		search := weightfind.WeightSearcherExtended1{}
		search.Init(sol.input.statTypes, sol.input.targetRatio)
		search.SupplyData(sol.input.dataAll)
		search.SetRanges(-1.0, 5.0)
		weight2 := new(search.Run(innerCancel))

		weightResult := weight_types.WeightResult2Make(weight2, sw.Elapsed(), highs.ModelStatusOptimal)
		sol.output.evaluateWeightResult2("SEARCH-EX1", &weightResult)
	})
}

func (sol *solves) handleFuture1OrError(choiceName string, weightsFuture *util_async.FutureCancellable[weight_types.WeightResult1], err error) {
	if err != nil {
		sol.output.handleWeightError(choiceName, err)
		return
	}

	err = util_async.ChainCancel(sol.cancel, weightsFuture)
	if err != nil {
		sol.output.handleWeightError(choiceName, err)
		return
	}

	sol.output.evaluateWeightResultFuture1(choiceName, weightsFuture)
}

func (sol *solves) handleFuture2OrError(choiceName string, weightsFuture *util_async.FutureCancellable[weight_types.WeightResult2], err error) {
	if err != nil {
		sol.output.handleWeightError(choiceName, err)
		return
	}

	err = util_async.ChainCancel(sol.cancel, weightsFuture)
	if err != nil {
		sol.output.handleWeightError(choiceName, err)
		return
	}

	sol.output.evaluateWeightResultFuture2(choiceName, weightsFuture)
}

func loadOldWeights(param *SpecParam, out *choiceOutput) {
	oldWeightBlock, _, oldWeightExists := tools.PawnWeightReadFile(param.WeightFile1)
	if oldWeightExists {
		oldWeight := weight_types.Weight1Basic_FromBlock(oldWeightBlock)
		out.evaluateWeight1("OLD", &oldWeight)
	}

	if param.WeightFile2 == "" {
		param.WeightFile2 = files.ToWeight2(param.WeightFile1)
	}
	if weight2, weight2Found := tools.ReadWeight2File(param.WeightFile2); weight2Found {
		out.evaluateWeight2("OLD2", weight2)
	}

	if param.WeightFile3 == "" {
		param.WeightFile3 = files.ToWeight3(param.WeightFile1)
	}
	if weight3, weight3Found := tools.ReadWeight3File(param.WeightFile3); weight3Found {
		out.evaluateWeight3("OLD3", weight3)
	}
}
