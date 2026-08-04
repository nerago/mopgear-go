package weight_highs

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/simrank"
	"paladin_gearing_go/weightfind/util_weight"
	"paladin_gearing_go/weightfind/weight_types"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_rank1_scaleTarget     = 1.0
	c_Rank1_LargeWeight     = 50.0
	c_Rank1_LargeScore      = 500.0
	c_Rank1_LargeRank       = 10000.0
	c_Rank1_TargetWeightSum = 10.0
)

type RankingStatWeightProcess struct {
	printer *util.PrintRecorder

	targetRatios  weight_types.SimPriorityBasic
	requiredStats []stats.StatType
	requiredSims  []stats.SimType
	data          []*rankEntry1

	build *util_highs.LinearBuilder

	scaleStats    util_collection.EnumMap[stats.StatType, float64]
	weightColumns map[stats.StatType]util_highs.ColumnIndex

	RANKMODE  int
	WEIGHTSUM int
}

type rankEntry1 struct {
	weight_types.RankStatWeightsCommon

	ScoreColumn util_highs.ColumnIndex
	RankColumn  util_highs.ColumnIndex
}

func (ranker *RankingStatWeightProcess) Init(printer *util.PrintRecorder) {
	ranker.printer = printer
}

func (ranker *RankingStatWeightProcess) SupplyData(inputData []weight_types.WeightInput) {
	ranker.scaleStats = util_weight.ChooseStatScalingBasic(inputData, c_rank1_scaleTarget, false, ranker.printer)
	ranker.data = util_collection.MapSliceAsNew(inputData, func(input *weight_types.WeightInput) *rankEntry1 {
		return &rankEntry1{
			RankStatWeightsCommon: weight_types.RankStatWeightsCommon{
				Data:       input,
				SimScore:   -1,
				TargetRank: -1,
			},
		}
	})
}

func (ranker *RankingStatWeightProcess) SetRequiredStats(requiredStats []stats.StatType) {
	ranker.requiredStats = requiredStats
}

func (ranker *RankingStatWeightProcess) SetTargetRatios(targetRatios weight_types.SimPriorityBasic) {
	ranker.targetRatios = targetRatios
	ranker.requiredSims = targetRatios.SimTypes()
}

func (ranker *RankingStatWeightProcess) Run(timeout int) *util_async.FutureCancellable[weight_types.WeightResult] {
	ranker.build = new(util_highs.LinearBuilder)
	ranker.build.Minimise = true
	ranker.build.Solver = util_highs.Solver_Force_IPX
	ranker.build.TimeLimitSeconds = timeout

	ranker.createWeightColumns()
	ranker.prepareRankings()
	ranker.processData()

	stopwatch := util.StopwatchMakeStopped()
	solutionFuture := ranker.build.RunHighsFuture(stopwatch)
	return util_async.FutureCancellable_MapValue(solutionFuture, func(linearResult util_highs.LinearResult) (weight_types.WeightResult, bool) {
		solution := linearResult.GetSolutionAndSaveLog(ranker.printer)
		weight := ranker.extractAndReportSolution(solution)
		return weight_types.WeightResult{Weight: &weight, SolveTime: stopwatch.Elapsed(), Status: solution.Status}, true
	})
}

func (ranker *RankingStatWeightProcess) createWeightColumns() {
	lo := -c_Rank1_LargeWeight
	hi := c_Rank1_LargeWeight

	sumWeights := util_highs.ConstraintRow{}
	ranker.weightColumns = make(map[stats.StatType]util_highs.ColumnIndex)
	for _, statType := range ranker.requiredStats {
		colDetailWeight := ranker.build.CreateColumnGeneral(highs.Continuous, lo, hi, util_highs.DebugString{Text: "WEIGHT " + statType.Name()})
		ranker.weightColumns[statType] = colDetailWeight
		sumWeights.Add(colDetailWeight, 1)
	}

	if ranker.WEIGHTSUM == 0 {
		sumWeights.Build(ranker.build, c_Rank1_TargetWeightSum, c_Rank1_TargetWeightSum)
	} else {
		sumWeights.Build(ranker.build, 0.001, util_highs.InfPos()) // force positive and non-zero result
	}
}

func (ranker *RankingStatWeightProcess) prepareRankings() {
	simrank.RankingWeightsPrepareBasicRankings(ranker.requiredSims, &ranker.targetRatios, ranker.data)
}

func (ranker *RankingStatWeightProcess) processData() {
	switch ranker.RANKMODE {
	case 0:
		for _, entry := range ranker.data {
			ranker.processDataEntryOriginal(entry)
		}

		// several adjacent
		eachCheckCount := 50
		for a := 0; a < len(ranker.data); a++ {
			for b := a + 1; b < min(a+eachCheckCount, len(ranker.data)); b++ {
				ranker.processEntrySequencePairOriginal(ranker.data[a], ranker.data[b])
			}
		}
	case 1, 2:
		for _, entry := range ranker.data {
			ranker.processDataEntryForceScoreToRank(entry)
		}
	}

	// just adjacent samples
	// for i := 0; i < len(ranker.data)-1; i++ {
	// 	ranker.processEntrySequencePair(&ranker.data[i], &ranker.data[i+1])
	// }

	// all pairs of samples
	// for a := 0; a < len(ranker.data); a++ {
	// 	for b := a + 1; b < len(ranker.data); b++ {
	// 		ranker.processEntrySequencePair(&ranker.data[a], &ranker.data[b])
	// 	}
	// }

	// problem is we don't really penalise greater divergence specifically
}

func (ranker *RankingStatWeightProcess) processDataEntryOriginal(entry *rankEntry1) {
	// these scores are meaningless in themselves, at least in value terms
	// however their increasing sequence should correlate to combinedSimRankScore
	// which is what we'll optimise for
	entry.ScoreColumn = ranker.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugText("score"))

	scoreRow := util_highs.ConstraintRow{}
	for _, statType := range ranker.requiredStats {
		weightColumn := ranker.weightColumns[statType]
		statValue := entry.Data.TotalStat.GetFloat(statType)
		statScale := ranker.scaleStats.GetOrPanic(statType)

		scoreRow.Add(weightColumn, statValue*statScale)
	}

	scoreRow.Add(entry.ScoreColumn, -1)
	scoreRow.Build(ranker.build, 0, 0)
}

func (ranker *RankingStatWeightProcess) processDataEntryPlusRankCompareToExpected(entry *rankEntry1) {
	// these scores are meaningless in themselves, at least in value terms
	// however their increasing sequence should correlate to combinedSimRankScore
	// which is what we'll optimise for
	rankStr := strconv.FormatInt(int64(entry.TargetRank), 10)
	entry.ScoreColumn = ranker.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugText("score-"+rankStr))

	scoreRow := util_highs.ConstraintRow{}
	for _, statType := range ranker.requiredStats {
		weightColumn := ranker.weightColumns[statType]
		statValue := entry.Data.TotalStat.GetFloat(statType)
		statScale := ranker.scaleStats.GetOrPanic(statType)

		scoreRow.Add(weightColumn, statValue*statScale)
	}
	scoreRow.Add(entry.ScoreColumn, -1)
	scoreRow.Build(ranker.build, 0, 0)

	// TODO give it an initial solution to ranks

	entry.RankColumn = ranker.build.CreateColumnGeneral(highs.Integer, 0, float64(len(ranker.data)-1), util_highs.DebugText("derivedRank-"+rankStr))
	rankDiffColumn := ranker.build.CreateColumnGeneral(highs.Integer, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugText("rankDiff-"+rankStr))
	rankDiffAbsColumn := ranker.build.CreateColumnWithOutput(highs.Integer, 0, util_highs.InfPos(), 1, util_highs.DebugText("rankDiffAbs-"+rankStr))

	targetRank := float64(entry.TargetRank)
	rankDiff := util_highs.ConstraintRow{}
	rankDiff.Add(entry.RankColumn, 1)
	rankDiff.Add(rankDiffColumn, -1)
	rankDiff.Build(ranker.build, targetRank, targetRank)
	ranker.build.AbsoluteValue(rankDiffColumn, rankDiffAbsColumn)

	ranker.build.SetInitialSolutionValue(entry.RankColumn, float64(entry.TargetRank))
}

// one thing to try might be just to say that stats->score should calculate a rank directly. then can do a direct compare to targetRank
func (ranker *RankingStatWeightProcess) processDataEntryForceScoreToRank(entry *rankEntry1) {
	offsetColumn := ranker.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugText("offset"))
	offsetAbs := ranker.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), 1, util_highs.DebugText("offset"))
	ranker.build.AbsoluteValue(offsetColumn, offsetAbs)

	scoreRow := util_highs.ConstraintRow{}
	for _, statType := range ranker.requiredStats {
		weightColumn := ranker.weightColumns[statType]
		statValue := entry.Data.TotalStat.GetFloat(statType)
		statScale := ranker.scaleStats.GetOrPanic(statType)

		scoreRow.Add(weightColumn, statValue*statScale)
	}

	scoreRow.Add(offsetColumn, 1)

	switch ranker.RANKMODE {
	case 1:
		targetNum := float64(entry.TargetRank)
		scoreRow.Build(ranker.build, targetNum, targetNum)
	case 2:
		targetNum := entry.SimScore
		scoreRow.Build(ranker.build, targetNum, targetNum)
	}
}

// parameters don't imply order
func (ranker *RankingStatWeightProcess) processEntrySequencePairToDerivedRank(one *rankEntry1, two *rankEntry1) {
	// so we could totally do a boolean thing where scoreA>scoreB then implies rankA>rankB
	// would need all possible pairs connected, but would then force solver to make a full integer order
	isGreater := ranker.build.CreateColumnBool(util_highs.DebugText("isGreater"))

	ranker.build.ColumnIsGreaterOrEqualColumn(one.ScoreColumn, two.ScoreColumn, isGreater, c_Rank1_LargeScore, 0.0001)

	ranker.build.ColumnIsGreaterOrEqualColumn(one.RankColumn, two.RankColumn, isGreater, c_Rank1_LargeRank, 1.0)
}

// we want to optimise for higher.score > lower.score
func (ranker *RankingStatWeightProcess) processEntrySequencePairOriginal(lower *rankEntry1, higher *rankEntry1) {
	offsetColumn := ranker.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), 1, util_highs.DebugText("offset"))

	// if lower <= higher then it will trivially pass the >= 0 check. offset will be free but under minimise pressure so effectively zero
	// if lower > higher then it will initially fail the >= 0 check, and need an extra boost from offset to get over the line
	compareRow := util_highs.ConstraintRow{}
	compareRow.Add(lower.ScoreColumn, -1)
	compareRow.Add(higher.ScoreColumn, 1)
	compareRow.Add(offsetColumn, 1)
	compareRow.Build(ranker.build, 0, util_highs.InfPos())
}

func (ranker *RankingStatWeightProcess) extractAndReportSolution(solution *highs.Solution) weight_types.Weight1Basic {
	ranker.build.DebugPrintColumns(solution, ranker.printer)

	ranker.printer.Println("WEIGHTS")

	weight := weight_types.Weight1Basic_Make()
	for _, statType := range ranker.requiredStats {
		weightColumn := ranker.weightColumns[statType]
		statScale := ranker.scaleStats.GetOrPanic(statType)

		modelWeight := solution.ColValues[weightColumn]
		usableWeight := modelWeight * statScale

		weight.Put(statType, usableWeight)
	}

	baseStat := ranker.requiredStats[0]
	divideBy := weight.Get(baseStat)
	for _, statType := range ranker.requiredStats {
		weight.Put(statType, weight.Get(statType)/divideBy)
	}

	ranker.reportRankingOfInputs(weight)

	return weight
}

func (ranker *RankingStatWeightProcess) reportRankingOfInputs(weight weight_types.Weight1Basic) {
	ranker.printer.Println("INPUT CHECK (index, combinedSimRank, calcStat)")
	for i, entry := range ranker.data {
		ranker.printer.Printf("%4d %8f %8f\n", i, entry.SimScore, weight.CalcStatScore(&entry.Data.TotalStat))
	}
}
