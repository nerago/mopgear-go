package stathighs

import (
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_rank1_scaleTarget = 1.0
	c_Rank1_LargeWeight = 50.0
	c_Rank1_LargeScore  = 500.0
	c_Rank1_LargeRank   = 10000.0
)

type RankingStatWeightProcess struct {
	printer *util.PrintRecorder

	targetRatios  stats.SimData
	requiredStats []stats.StatType
	requiredSims  []stats.SimType
	data          []rankEntry

	build *utilhighs.LinearBuilder

	// linearEquationDiff int
	// linearInclude      int

	scaleStats    map[stats.StatType]float64
	weightColumns map[stats.StatType]utilhighs.ColumnIndex

	// minimumIncludeRate float64
	// includeColumns     []utilhighs.ColumnIndex
	// includeCountRow    utilhighs.ConstraintRowBuild

	RANKMODE int
}

type rankEntry struct {
	data *WeightInput

	simRanks         map[stats.SimType]int
	combinedSimScore float64
	targetRank       int

	scoreColumn utilhighs.ColumnIndex
	rankColumn  utilhighs.ColumnIndex
}

func (ranker *RankingStatWeightProcess) Init(printer *util.PrintRecorder) {
	ranker.printer = printer
}

func (ranker *RankingStatWeightProcess) SupplyData(inputData []WeightInput) {
	ranker.scaleStats = chooseStatScaling(inputData, c_rank1_scaleTarget, ranker.printer)
	ranker.data = util.MapSliceAsNew(inputData, func(input *WeightInput) rankEntry {
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

// func (ranker *RankStatWeightProcess) SetMinimumIncludeRate(percent float64) {
// 	ranker.minimumIncludeRate = percent
// }

func (ranker *RankingStatWeightProcess) Run(stopwatch *util.Stopwatch, timeout int) WeightResult {
	ranker.build = new(utilhighs.LinearBuilder)
	ranker.build.Minimise = true
	if ranker.RANKMODE == 0 || ranker.RANKMODE == 1 || ranker.RANKMODE == 2 {
		ranker.build.Solver = utilhighs.Solver_LP_USE_GPU
	} else {
		ranker.build.Solver = utilhighs.Solver_MIP_Interior
	}
	ranker.build.TimeLimitSeconds = timeout
	if ranker.RANKMODE != 3 {
		ranker.build.DisablePreSolve = true
	}

	ranker.createWeightColumns()
	ranker.prepareRankings()
	ranker.processData()

	// ranker.includeCountRow.Finish(ranker.input, float64(len(ranker.inputData))*ranker.minimumIncludeRate, utilhighs.C_PlusInf)

	solution := ranker.build.RunHighs(ranker.printer, stopwatch)

	return ranker.extractAndReportSolution(solution)
}

func (ranker *RankingStatWeightProcess) createWeightColumns() {
	lo := -c_Rank1_LargeWeight
	hi := c_Rank1_LargeWeight

	sumWeights := utilhighs.ConstraintRow{}
	ranker.weightColumns = make(map[stats.StatType]utilhighs.ColumnIndex)
	for _, statType := range ranker.requiredStats {
		colDetailWeight := ranker.build.CreateColumnGeneral(highs.Continuous, lo, hi, utilhighs.DebugString{Text: "WEIGHT " + statType.Name()})
		ranker.weightColumns[statType] = colDetailWeight
		sumWeights.Add(colDetailWeight, 1)
	}

	if ranker.RANKMODE == 0 || ranker.RANKMODE == 3 {
		// TODO just doesn't seem right
		sumWeights.Build(ranker.build, 1.0, utilhighs.C_PlusInf) // force positive and non-zero result
	}

	// this assumes that a positive strength will work for the scoring system, should be true in most situations
	// if not then i'd be concerned more about garbage input data
	// setStrength := utilhighs.ConstraintRowBuild{}
	// setStrength.Add(ranker.weightColumns[stats.Stat_Strength], 1)
	// setStrength.Finish(ranker.input, 1.0, 1.0)
}

func (ranker *RankingStatWeightProcess) prepareRankings() {
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
	case 3:
		sumRanks := utilhighs.ConstraintRow{}
		for entry := range util.ForPointer(ranker.data) {
			ranker.processDataEntryPlusRankCompareToExpected(entry)
			sumRanks.Add(entry.rankColumn, 1)
		}
		// QUERY does this add accuracy or speed, or both?
		// check 10, accuracy with disabled: 87.1993%, Duration = 2m5.2630056s
		// check 10, accuracy with enabled: 87.1993%, Duration = 1m48.1206008s
		expectedSum := len(ranker.data) * (len(ranker.data) - 1) / 2
		sumRanks.Build(ranker.build, float64(expectedSum), float64(expectedSum))

		eachCheckCount := 2
		for a := 0; a < len(ranker.data); a++ {
			for b := a + 1; b < min(a+eachCheckCount, len(ranker.data)); b++ {
				ranker.processEntrySequencePairToDerivedRank(&ranker.data[a], &ranker.data[b])
			}
		}
		// check 10: 87.1993%, Duration = 1m48.1206008s
		// check 4: 88.0098%, Duration = 1m16.4640117s
		// check 3: 88.0145%, Duration = 49.6121775s
		// check 2: 88.0154%, Duration = 28.5001548s
		// check 1: 86.2758%, Duration = 27.303175s

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
	entry.scoreColumn = ranker.build.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugText("score"))

	scoreRow := utilhighs.ConstraintRow{}
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
	entry.scoreColumn = ranker.build.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugText("score-"+rankStr))

	scoreRow := utilhighs.ConstraintRow{}
	for _, statType := range ranker.requiredStats {
		weightColumn := ranker.weightColumns[statType]
		statValue := entry.data.TotalStat.GetFloat(statType)
		statScale := ranker.scaleStats[statType]

		scoreRow.Add(weightColumn, statValue*statScale)
	}
	scoreRow.Add(entry.scoreColumn, -1)
	scoreRow.Build(ranker.build, 0, 0)

	// TODO give it an initial solution to ranks

	entry.rankColumn = ranker.build.CreateColumnGeneral(highs.Integer, 0, float64(len(ranker.data)-1), utilhighs.DebugText("derivedRank-"+rankStr))
	rankDiffColumn := ranker.build.CreateColumnGeneral(highs.Integer, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugText("rankDiff-"+rankStr))
	rankDiffAbsColumn := ranker.build.CreateColumnWithOutput(highs.Integer, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugText("rankDiffAbs-"+rankStr))

	targetRank := float64(entry.targetRank)
	rankDiff := utilhighs.ConstraintRow{}
	rankDiff.Add(entry.rankColumn, 1)
	rankDiff.Add(rankDiffColumn, -1)
	rankDiff.Build(ranker.build, targetRank, targetRank)
	ranker.build.AbsoluteValue(rankDiffColumn, rankDiffAbsColumn)

	ranker.build.SetInitialSolutionValue(entry.rankColumn, float64(entry.targetRank))
}

// one thing to try might be just to say that stats->score should calculate a rank directly. then can do a direct compare to targetRank
func (ranker *RankingStatWeightProcess) processDataEntryForceScoreToRank(entry *rankEntry) {
	offsetColumn := ranker.build.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugText("offset"))
	offsetAbs := ranker.build.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugText("offset"))
	ranker.build.AbsoluteValue(offsetColumn, offsetAbs)

	scoreRow := utilhighs.ConstraintRow{}
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
	isGreater := ranker.build.CreateColumnBool(utilhighs.DebugText("isGreater"))

	ranker.build.ColumnIsGreaterOrEqualColumn(one.scoreColumn, two.scoreColumn, isGreater, c_Rank1_LargeScore, 0.0001)

	ranker.build.ColumnIsGreaterOrEqualColumn(one.rankColumn, two.rankColumn, isGreater, c_Rank1_LargeRank, 1.0)
}

// we want to optimise for higher.score > lower.score
func (ranker *RankingStatWeightProcess) processEntrySequencePairOriginal(lower *rankEntry, higher *rankEntry) {
	offsetColumn := ranker.build.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugText("offset"))

	// if lower <= higher then it will trivially pass the >= 0 check. offset will be free but under minimise pressure so effectively zero
	// if lower > higher then it will initially fail the >= 0 check, and need an extra boost from offset to get over the line
	compareRow := utilhighs.ConstraintRow{}
	compareRow.Add(lower.scoreColumn, -1)
	compareRow.Add(higher.scoreColumn, 1)
	compareRow.Add(offsetColumn, 1)
	compareRow.Build(ranker.build, 0, utilhighs.C_PlusInf)
}

func (ranker *RankingStatWeightProcess) extractAndReportSolution(solution *highs.Solution) WeightResult {
	ranker.build.DebugPrintColumns(solution, ranker.printer)

	ranker.printer.Println("WEIGHTS")

	statWeightResult := WeightResult_Make()
	for _, statType := range ranker.requiredStats {
		weightColumn := ranker.weightColumns[statType]
		statScale := ranker.scaleStats[statType]

		modelWeight := solution.ColValues[weightColumn]
		// TODO changed to multiply following analysis on other algorithms, not checked here
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

func (ranker *RankingStatWeightProcess) reportRankingOfInputs(statWeightResult WeightResult) {
	ranker.printer.Println("INPUT CHECK (index, combinedSimRank, calcStat)")
	for i, entry := range ranker.data {
		ranker.printer.Printf("%4d %8f %8f\n", i, entry.combinedSimScore, statWeightResult.CalcStatScore(entry.data))
	}
}
