package weight_highs

import (
	"fmt"
	"strconv"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/tools"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_highs"
	"github.com/nerago/mopgear-go/weightfind/util_weight"
	"github.com/nerago/mopgear-go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_rank30_targetTotalWeight = 1.0
	c_rank30_targetTotalRatio  = 1.0
	c_rank30_maxWeight         = 1.0
	c_rank30_maxRatio          = 1.0
	c_rank30_maxRatioChange    = 0.15
	c_rank30_ratioOutputScale  = 100.0
	c_rank30_rankOutputScale   = 1.0
)

// RankingWeightsRatio30: based on RankingStatWeightProcess3b
type RankingWeightsRatio30 struct {
	printer        *util.PrintRecorder
	timeoutSeconds int

	targetRatios  weight_types.SimPriorityBasic
	requiredStats []stats.StatType
	requiredSims  []stats.SimType
	simScale      stats.SimTypeMap[util_weight.ScaleAndOffset]
	dataSample    []rankEntry30
	AllPairs      bool

	build *util_highs.LinearBuilder

	weightColumns stats.StatTypeMap[util_highs.ColumnIndex]
	ratioColumns  stats.SimTypeMap[util_highs.ColumnIndex]
	pairLinks     util_collection.MapMapDiagonal[int, rankPair30]
}

type rankEntry30 struct {
	data *weight_types.WeightInput

	simScoreColumn  util_highs.ColumnIndex
	statScoreColumn util_highs.ColumnIndex
}

type rankPair30 struct {
	indexOne, indexTwo int
	scoreSlackColumn   util_highs.ColumnIndex
}

func (ranker *RankingWeightsRatio30) Init(printer *util.PrintRecorder, timeoutSeconds int) {
	ranker.printer = printer
	ranker.timeoutSeconds = timeoutSeconds
}

func (ranker *RankingWeightsRatio30) SupplyData(inputData []weight_types.WeightInput) {
	ranker.dataSample = util_collection.MapSliceAsNew(inputData, func(input *weight_types.WeightInput) rankEntry30 {
		return rankEntry30{
			data:            input,
			simScoreColumn:  -1,
			statScoreColumn: -1,
		}
	})
	ranker.simScale = util_weight.ChooseSimUnfriendlyUnitScaleAndOffset(inputData, ranker.requiredSims)
}

func (ranker *RankingWeightsRatio30) SetRequiredStats(requiredStats []stats.StatType) {
	ranker.requiredStats = requiredStats
}

func (ranker *RankingWeightsRatio30) SetTargetRatios(targetRatios weight_types.SimPriorityBasic) {
	ranker.targetRatios = targetRatios
	ranker.requiredSims = targetRatios.SimTypes()
}

func (ranker *RankingWeightsRatio30) newBuilder() {
	ranker.build = new(util_highs.LinearBuilder)
	ranker.build.Minimise = true
	ranker.build.TimeLimitSeconds = ranker.timeoutSeconds
	ranker.build.Solver = util_highs.Solver_Force_Simplex
	// others can be faster but often fails, then eventually end up in simplex anyway
}

func (ranker *RankingWeightsRatio30) RunSinglePassRaw() *util_async.FutureCancellable[weight_types.WeightResult] {
	// FULL RUN
	ranker.newBuilder()
	ranker.createWeightColumns()
	ranker.createRatioColumns()
	ranker.makeRatioOutputSlacks()
	ranker.doAlgos()

	stopwatch := util.StopwatchMakeStopped()
	solutionFuture := ranker.build.RunHighsFuture(stopwatch)

	return util_async.FutureCancellable_MapValue(solutionFuture, func(linearResult2 util_highs.LinearResult) (weight_types.WeightResult, bool) {
		solution := linearResult2.GetSolutionAndSaveLog(ranker.printer)
		weight := ranker.extractAndReportSolution(solution)
		return weight_types.WeightResult{Weight: &weight, SolveTime: stopwatch.Elapsed(), Status: solution.Status}, true
	})
}

func (ranker *RankingWeightsRatio30) createWeightColumns() {
	lo := -c_rank30_maxWeight
	hi := c_rank30_maxWeight

	//sumWeights := util_highs.ConstraintRow{Debug: "sumWeights"}
	for _, statType := range ranker.requiredStats {
		colWeight := ranker.build.CreateColumnGeneral(highs.Continuous, lo, hi, util_highs.DebugString{Text: "WEIGHT " + statType.Name()})
		ranker.weightColumns.Put(statType, colWeight)
		//sumWeights.Add(colWeight, 1)
	}

	//sumWeights.Build(ranker.build, c_rank30_targetTotalWeight, c_rank30_targetTotalWeight)
}

func (ranker *RankingWeightsRatio30) createRatioColumns() {
	lo := 0.0
	hi := c_rank30_maxRatio

	sumRatio := util_highs.ConstraintRow{Debug: "sumRatio"}
	for _, simType := range ranker.requiredSims {
		colRatio := ranker.build.CreateColumnGeneral(highs.Continuous, lo, hi, util_highs.DebugString{Text: "RATIO " + simType.Name()})
		ranker.ratioColumns.Put(simType, colRatio)
		sumRatio.Add(colRatio, 1)
	}

	sumRatio.Build(ranker.build, c_rank30_targetTotalRatio, c_rank30_targetTotalRatio)
}

func (ranker *RankingWeightsRatio30) makeRatioOutputSlacks() {
	for _, simType := range ranker.requiredSims {
		colRatio := ranker.ratioColumns.GetOrPanic(simType)
		targetRatio := ranker.targetRatios.GetOrPanic(simType)
		ratioSlack := ranker.build.CreateColumnWithOutput(highs.Continuous,
			//0, c_rank30_maxRatioChange,
			0, util_highs.InfPos(),
			c_rank30_ratioOutputScale,
			util_highs.DebugString{Text: "RATIO SLACK " + simType.Name()})
		ranker.build.AbsoluteValueFromDiffOneToConst(
			colRatio, 1,
			targetRatio,
			ratioSlack, "RATIO CHANGE "+simType.Name())
	}
}

func (ranker *RankingWeightsRatio30) doAlgos() {
	if ranker.AllPairs {
		ranker.makeDataListEntryColumns()
		for baseIndex := range ranker.dataSample {
			for compareTo := baseIndex + 1; compareTo < len(ranker.dataSample); compareTo++ {
				ranker.makeEntryPairCheckScoreOrderMatchesTargetOrderWithSlackVar(&ranker.dataSample[baseIndex], &ranker.dataSample[compareTo], baseIndex, compareTo)
			}
		}
	} else {
		ranker.makeDataListEntryColumns()
		for baseIndex := 0; baseIndex < len(ranker.dataSample)-1; baseIndex++ {
			ranker.makeEntryPairCheckScoreOrderMatchesTargetOrderWithSlackVar(&ranker.dataSample[baseIndex], &ranker.dataSample[baseIndex+1], baseIndex, baseIndex+1)
		}
	}
}

func (ranker *RankingWeightsRatio30) makeDataListEntryColumns() {
	for index := range ranker.dataSample {
		entry := &ranker.dataSample[index]
		debugStr := strconv.FormatInt(int64(index), 10)
		ranker.makeStatScoreColumn(entry, debugStr)
		ranker.makeSimScoreColumn(entry, debugStr)
	}
}

func (ranker *RankingWeightsRatio30) makeStatScoreColumn(entry *rankEntry30, debugStr string) {
	entry.statScoreColumn = ranker.build.CreateColumnGeneral(highs.Continuous, 0, util_highs.InfPos(), util_highs.DebugText("statScore-"+debugStr))
	scoreRow := util_highs.ConstraintRow{Debug: "scoreRow-" + debugStr}
	for _, statType := range ranker.requiredStats {
		weightColumn := ranker.weightColumns.GetOrPanic(statType)
		statValue := entry.data.TotalStat.GetFloat(statType)
		scoreRow.Add(weightColumn, statValue)
	}
	scoreRow.Add(entry.statScoreColumn, -1)
	scoreRow.Build(ranker.build, 0, 0)
}

func (ranker *RankingWeightsRatio30) makeSimScoreColumn(entry *rankEntry30, debugStr string) {
	// score each sim
	//for _, simType := range ranker.requiredSims {
	//	for entry, simDetailRank := range util_collection.CalculateRanking(simType.IsHighGood(), ranker.dataSample, func(x *rankEntry30) float64 { return x.data.SimResult.Get(simType) }) {
	//		entry.simScore += float64(simDetailRank) * ranker.targetRatios.GetOrPanic(simType)
	//	}
	//}
	//
	//// rank combined sims
	//for entry, simRank := range util_collection.CalculateRanking(true, ranker.dataSample, func(x *rankEntry30) float64 { return x.simScore }) {
	//	entry.targetRank = simRank
	//}

	entry.simScoreColumn = ranker.build.CreateColumnGeneral(highs.Continuous, 0, util_highs.InfPos(), util_highs.DebugText("simScore-"+debugStr))
	simRow := util_highs.ConstraintRow{Debug: "simRow-" + debugStr}
	for _, simType := range ranker.requiredSims {
		ratioColumn := ranker.ratioColumns.GetOrPanic(simType)
		simValue := entry.data.SimResult.Get(simType)
		scale := ranker.simScale.GetOrPanic(simType)
		scaledSimValue := scale.Apply(simValue)
		simRow.Add(ratioColumn, scaledSimValue)
	}
	simRow.Add(entry.statScoreColumn, -1)
	simRow.Build(ranker.build, 0, 0)
}

func (ranker *RankingWeightsRatio30) makeEntryPairCheckScoreOrderMatchesTargetOrderWithSlackVar(one *rankEntry30, two *rankEntry30, indexOne, indexTwo int) {
	debug := fmt.Sprintf(" %d %d", indexOne, indexTwo)
	slack := ranker.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), c_rank30_rankOutputScale, util_highs.DebugText("slack"+debug))

	// opposite signs on the two differences, going for:
	// simDiff=(simOne-simTwo), statDiff=(statOne-statTwo)
	// slack = abs(simDiff-statDiff) = abs((simOne-simTwo)-(statOne-statTwo))
	//                               = abs(simOne-simTwo-statOne+statTwo)
	ranker.build.AbsoluteValueFromSumSeveralThenDiffToConst(
		[]util_highs.ColumnIndex{one.simScoreColumn, two.simScoreColumn, one.statScoreColumn, two.statScoreColumn},
		[]float64{1, -1, -1, 1},
		0,
		slack, "diff scores "+debug)

	ranker.pairLinks.Put(indexOne, indexTwo, rankPair30{
		indexOne:         indexOne,
		indexTwo:         indexTwo,
		scoreSlackColumn: slack,
	})
}

func (ranker *RankingWeightsRatio30) extractAndReportSolution(solution *highs.Solution) weight_types.Weight1Basic {
	ranker.build.DebugPrintColumns(solution, ranker.printer)

	ranker.printer.Println("WEIGHTS")
	weight := weight_types.Weight1Basic_Make()
	for _, statType := range ranker.requiredStats {
		weightColumn := ranker.weightColumns.GetOrPanic(statType)
		modelWeight := solution.ColValues[weightColumn]
		weight.Put(statType, modelWeight)
	}
	weight.NormalizeForBase(ranker.requiredStats)
	tools.WriteWeightString(&weight, ranker.printer)

	ranker.printer.Println("RATIOS")
	for _, simType := range ranker.requiredSims {
		ratioColumn := ranker.ratioColumns.GetOrPanic(simType)
		modelRatio := solution.ColValues[ratioColumn]
		ranker.printer.Printf(" %s = %.8f\n", simType.Name(), modelRatio)
	}

	return weight
}
