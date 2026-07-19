package weight_highs

import (
	"cmp"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"
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

	targetRatios  stats.SimData
	requiredStats []stats.StatType
	requiredSims  []stats.SimType
	data          []rankEntry

	build *util_highs.LinearBuilder

	scaleStats    map[stats.StatType]float64
	weightColumns map[stats.StatType]util_highs.ColumnIndex

	RANKMODE  int
	WEIGHTSUM int
}

type rankEntry struct {
	data *weight_types.WeightInput

	simRanks         map[stats.SimType]int
	combinedSimScore float64
	targetRank       int

	scoreColumn util_highs.ColumnIndex
	rankColumn  util_highs.ColumnIndex
}

func (ranker *RankingStatWeightProcess) Init(printer *util.PrintRecorder) {
	ranker.printer = printer
}

func (ranker *RankingStatWeightProcess) SupplyData(inputData []weight_types.WeightInput) {
	ranker.scaleStats = chooseStatScaling(inputData, c_rank1_scaleTarget, false, ranker.printer)
	ranker.data = util.MapSliceAsNew(inputData, func(input *weight_types.WeightInput) rankEntry {
		return rankEntry{data: input, simRanks: make(map[stats.SimType]int)}
	})
}

func (ranker *RankingStatWeightProcess) SetRequiredStats(requiredStats []stats.StatType) {
	ranker.requiredStats = requiredStats
}

func (ranker *RankingStatWeightProcess) SetTargetRatios(targetRatios stats.SimData) {
	ranker.targetRatios = targetRatios
	ranker.requiredSims = targetRatios.NonZeroTypes()
}

func (ranker *RankingStatWeightProcess) Run(stopwatch *util.Stopwatch, timeout int) *util_async.FutureCancellable[weight_types.Weight1Basic] {
	ranker.build = new(util_highs.LinearBuilder)
	ranker.build.Minimise = true
	ranker.build.Solver = util_highs.Solver_LP_USE_GPU
	ranker.build.DisablePreSolve = true
	ranker.build.TimeLimitSeconds = timeout

	ranker.createWeightColumns()
	ranker.prepareRankings()
	ranker.processData()

	solutionFuture := ranker.build.RunHighsFuture(stopwatch)
	return util_async.FutureCancellable_MapValue(solutionFuture, func(linearResult util_highs.LinearResult) (weight_types.Weight1Basic, bool) {
		solution := linearResult.GetSolutionAndSaveLog(ranker.printer)
		return ranker.extractAndReportSolution(solution), true
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
		sumWeights.Build(ranker.build, 0.001, util_highs.C_PlusInf) // force positive and non-zero result
	}
}

func (ranker *RankingStatWeightProcess) prepareRankings() {
	// TODO clean this up into a more standard method

	// score each sim
	for _, simType := range ranker.requiredSims {
		for entry, simDetailRank := range util.CalculateRanking(simType.IsHighGood(), ranker.data, func(x *rankEntry) float64 { return x.data.SimResult.Get(simType) }) {
			entry.simRanks[simType] = simDetailRank
			entry.combinedSimScore += float64(simDetailRank) * ranker.targetRatios.Get(simType)
		}
	}

	// rank combined sims
	for entry, simRank := range util.CalculateRanking(true, ranker.data, func(x *rankEntry) float64 { return x.combinedSimScore }) {
		entry.targetRank = simRank
	}

	slices.SortFunc(ranker.data, func(a, b rankEntry) int { return cmp.Compare(a.targetRank, b.targetRank) })
}

func (ranker *RankingStatWeightProcess) processData() {
	switch ranker.RANKMODE {
	case 0:
		for entry := range util.ForPointer(ranker.data) {
			ranker.processDataEntryOriginal(entry)
		}

		// several adjacent
		eachCheckCount := 50
		for a := 0; a < len(ranker.data); a++ {
			for b := a + 1; b < min(a+eachCheckCount, len(ranker.data)); b++ {
				ranker.processEntrySequencePairOriginal(&ranker.data[a], &ranker.data[b])
			}
		}
	case 1, 2:
		for entry := range util.ForPointer(ranker.data) {
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

func (ranker *RankingStatWeightProcess) processDataEntryOriginal(entry *rankEntry) {
	// these scores are meaningless in themselves, at least in value terms
	// however their increasing sequence should correlate to combinedSimRankScore
	// which is what we'll optimise for
	entry.scoreColumn = ranker.build.CreateColumnGeneral(highs.Continuous, util_highs.C_MinusInf, util_highs.C_PlusInf, util_highs.DebugText("score"))

	scoreRow := util_highs.ConstraintRow{}
	for _, statType := range ranker.requiredStats {
		weightColumn := ranker.weightColumns[statType]
		statValue := entry.data.TotalStat.GetFloat(statType)
		statScale := ranker.scaleStats[statType]

		scoreRow.Add(weightColumn, statValue*statScale)
	}

	scoreRow.Add(entry.scoreColumn, -1)
	scoreRow.Build(ranker.build, 0, 0)
}

func (ranker *RankingStatWeightProcess) processDataEntryPlusRankCompareToExpected(entry *rankEntry) {
	// these scores are meaningless in themselves, at least in value terms
	// however their increasing sequence should correlate to combinedSimRankScore
	// which is what we'll optimise for
	rankStr := strconv.FormatInt(int64(entry.targetRank), 10)
	entry.scoreColumn = ranker.build.CreateColumnGeneral(highs.Continuous, util_highs.C_MinusInf, util_highs.C_PlusInf, util_highs.DebugText("score-"+rankStr))

	scoreRow := util_highs.ConstraintRow{}
	for _, statType := range ranker.requiredStats {
		weightColumn := ranker.weightColumns[statType]
		statValue := entry.data.TotalStat.GetFloat(statType)
		statScale := ranker.scaleStats[statType]

		scoreRow.Add(weightColumn, statValue*statScale)
	}
	scoreRow.Add(entry.scoreColumn, -1)
	scoreRow.Build(ranker.build, 0, 0)

	// TODO give it an initial solution to ranks

	entry.rankColumn = ranker.build.CreateColumnGeneral(highs.Integer, 0, float64(len(ranker.data)-1), util_highs.DebugText("derivedRank-"+rankStr))
	rankDiffColumn := ranker.build.CreateColumnGeneral(highs.Integer, util_highs.C_MinusInf, util_highs.C_PlusInf, util_highs.DebugText("rankDiff-"+rankStr))
	rankDiffAbsColumn := ranker.build.CreateColumnWithOutput(highs.Integer, 0, util_highs.C_PlusInf, 1, util_highs.DebugText("rankDiffAbs-"+rankStr))

	targetRank := float64(entry.targetRank)
	rankDiff := util_highs.ConstraintRow{}
	rankDiff.Add(entry.rankColumn, 1)
	rankDiff.Add(rankDiffColumn, -1)
	rankDiff.Build(ranker.build, targetRank, targetRank)
	ranker.build.AbsoluteValue(rankDiffColumn, rankDiffAbsColumn)

	ranker.build.SetInitialSolutionValue(entry.rankColumn, float64(entry.targetRank))
}

// one thing to try might be just to say that stats->score should calculate a rank directly. then can do a direct compare to targetRank
func (ranker *RankingStatWeightProcess) processDataEntryForceScoreToRank(entry *rankEntry) {
	offsetColumn := ranker.build.CreateColumnGeneral(highs.Continuous, util_highs.C_MinusInf, util_highs.C_PlusInf, util_highs.DebugText("offset"))
	offsetAbs := ranker.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.C_PlusInf, 1, util_highs.DebugText("offset"))
	ranker.build.AbsoluteValue(offsetColumn, offsetAbs)

	scoreRow := util_highs.ConstraintRow{}
	for _, statType := range ranker.requiredStats {
		weightColumn := ranker.weightColumns[statType]
		statValue := entry.data.TotalStat.GetFloat(statType)
		statScale := ranker.scaleStats[statType]

		scoreRow.Add(weightColumn, statValue*statScale)
	}

	scoreRow.Add(offsetColumn, 1)

	switch ranker.RANKMODE {
	case 1:
		targetNum := float64(entry.targetRank)
		scoreRow.Build(ranker.build, targetNum, targetNum)
	case 2:
		targetNum := entry.combinedSimScore
		scoreRow.Build(ranker.build, targetNum, targetNum)
	}
}

// parameters don't imply order
func (ranker *RankingStatWeightProcess) processEntrySequencePairToDerivedRank(one *rankEntry, two *rankEntry) {
	// so we could totally do a boolean thing where scoreA>scoreB then implies rankA>rankB
	// would need all possible pairs connected, but would then force solver to make a full integer order
	isGreater := ranker.build.CreateColumnBool(util_highs.DebugText("isGreater"))

	ranker.build.ColumnIsGreaterOrEqualColumn(one.scoreColumn, two.scoreColumn, isGreater, c_Rank1_LargeScore, 0.0001)

	ranker.build.ColumnIsGreaterOrEqualColumn(one.rankColumn, two.rankColumn, isGreater, c_Rank1_LargeRank, 1.0)
}

// we want to optimise for higher.score > lower.score
func (ranker *RankingStatWeightProcess) processEntrySequencePairOriginal(lower *rankEntry, higher *rankEntry) {
	offsetColumn := ranker.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.C_PlusInf, 1, util_highs.DebugText("offset"))

	// if lower <= higher then it will trivially pass the >= 0 check. offset will be free but under minimise pressure so effectively zero
	// if lower > higher then it will initially fail the >= 0 check, and need an extra boost from offset to get over the line
	compareRow := util_highs.ConstraintRow{}
	compareRow.Add(lower.scoreColumn, -1)
	compareRow.Add(higher.scoreColumn, 1)
	compareRow.Add(offsetColumn, 1)
	compareRow.Build(ranker.build, 0, util_highs.C_PlusInf)
}

func (ranker *RankingStatWeightProcess) extractAndReportSolution(solution *highs.Solution) weight_types.Weight1Basic {
	ranker.build.DebugPrintColumns(solution, ranker.printer)

	ranker.printer.Println("WEIGHTS")

	statWeightResult := weight_types.Weight1Basic_Make()
	for _, statType := range ranker.requiredStats {
		weightColumn := ranker.weightColumns[statType]
		statScale := ranker.scaleStats[statType]

		modelWeight := solution.ColValues[weightColumn]
		usableWeight := modelWeight * statScale

		statWeightResult.Put(statType, usableWeight)
	}

	divideBy := statWeightResult.Get(stats.Stat_Strength)
	for _, statType := range ranker.requiredStats {
		statWeightResult.Put(statType, statWeightResult.Get(statType)/divideBy)
	}

	ranker.reportRankingOfInputs(statWeightResult)

	return statWeightResult
}

func (ranker *RankingStatWeightProcess) reportRankingOfInputs(statWeightResult weight_types.Weight1Basic) {
	ranker.printer.Println("INPUT CHECK (index, combinedSimRank, calcStat)")
	for i, entry := range ranker.data {
		ranker.printer.Printf("%4d %8f %8f\n", i, entry.combinedSimScore, statWeightResult.CalcStatScore(entry.data))
	}
}
