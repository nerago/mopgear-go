package stathighs

import (
	"math"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
// c_RankLargeWeight = 50.0
// c_RankLargeScore  = 500.0
// c_RankLargeRank   = 10000.0
)

type RankingStatWeightProcess3 struct {
	printer *util.PrintRecorder

	targetRatios simulate.SimData
	dataAll      []rankEntry3

	input *utilhighs.InputBuilder

	scaleStats    map[stats.StatType]float64
	weightColumns map[stats.StatType]utilhighs.ColumnIndex
	pairLinks     util.MapMap[int, int, rankPair3]
}

type rankEntry3 struct {
	data *WeightInput

	simScore   float64
	targetRank int

	scoreColumn utilhighs.ColumnIndex
	rankColumn  utilhighs.ColumnIndex
	// rankDiffColumn    utilhighs.ColumnIndex
	rankDiffAbsColumn utilhighs.ColumnIndex
}

type rankPair3 struct {
	indexOne, indexTwo            int
	isGreaterScore, isGreaterRank utilhighs.ColumnIndex
	sequenceDiff                  utilhighs.ColumnIndex
}

func (ranker *RankingStatWeightProcess3) Init(printer *util.PrintRecorder) {
	ranker.printer = printer
}

func (ranker *RankingStatWeightProcess3) SupplyData(inputData []WeightInput) {
	ranker.scaleStats = chooseStatScaling(inputData, ranker.printer)
	ranker.dataAll = util.MapSliceAsNew(inputData, func(input *WeightInput) rankEntry3 {
		return rankEntry3{
			data:              input,
			simScore:          -1,
			targetRank:        -1,
			scoreColumn:       -1,
			rankColumn:        -1,
			rankDiffAbsColumn: -1,
			// rankDiffColumn:    -1,
		}
	})
}

func (ranker *RankingStatWeightProcess3) SetTargetRatios(targetRatios simulate.SimData) {
	ranker.targetRatios = targetRatios
}

func (ranker *RankingStatWeightProcess3) Run() (WeightResult, WeightResult, WeightResult) {
	ranker.input = new(utilhighs.InputBuilder)
	ranker.input.Minimise = true

	// FIRST ROUND: minimal data, dumb initial values
	minimalData := ranker.dataAll[0:200]
	ranker.prepareRankings(minimalData)
	ranker.createWeightColumns()
	ranker.makeDataListEntryColumns(minimalData)
	ranker.setupDumbInitialSolution(minimalData)
	solution1, log := ranker.input.RunHighs()
	ranker.printer.AppendOther(log)
	weights1 := ranker.extractAndReportSolution(solution1)

	// SECOND ROUND, minimal data, add extra conditions, copy initial from previous
	ranker.makeDataListPairRules(minimalData)
	ranker.setupInitialSolutionFromPrevious(solution1)
	solution2, log := ranker.input.RunHighs()
	ranker.printer.AppendOther(log)
	weights2 := ranker.extractAndReportSolution(solution2)

	// THIRD ROUND, full data, copy just weights from previous
	// data change means column ids won't line up
	ranker.input = new(utilhighs.InputBuilder)
	ranker.input.Minimise = true
	fullData := ranker.dataAll
	ranker.prepareRankings(fullData)
	ranker.createWeightColumns()
	ranker.makeDataListEntryColumns(fullData)
	ranker.makeDataListPairRules(fullData)
	ranker.setupInitialSolutionFromPreviousWeightOnly(solution2)
	solution3, log := ranker.input.RunHighs()
	ranker.printer.AppendOther(log)
	weights3 := ranker.extractAndReportSolution(solution3)

	return weights1, weights2, weights3
}

func (ranker *RankingStatWeightProcess3) createWeightColumns() {
	lo := -c_RankLargeWeight
	hi := c_RankLargeWeight

	sumWeights := utilhighs.ConstraintRowBuild{Debug: "sumWeights"}
	ranker.weightColumns = make(map[stats.StatType]utilhighs.ColumnIndex)
	for _, statType := range G_RequiredStats {
		colWeight := ranker.input.CreateColumnGeneral(highs.Continuous, lo, hi, utilhighs.DebugString{Text: "WEIGHT " + statType.Name()})
		ranker.weightColumns[statType] = colWeight
		sumWeights.Add(colWeight, 1)
	}

	sumWeights.Finish(ranker.input, 1.0, utilhighs.C_PlusInf) // force positive and non-zero result
}

func (ranker *RankingStatWeightProcess3) prepareRankings(data []rankEntry3) {
	// reset values
	for i := range data {
		data[i].simScore = 0
		data[i].targetRank = 0
	}

	// score each sim
	for _, simType := range G_RequiredSims {
		for entry, simDetailRank := range util.CalculateRanking(simType.IsHighGood(), data, func(x *rankEntry3) float64 { return x.data.SimResult.Get(simType) }) {
			entry.simScore += float64(simDetailRank) * ranker.targetRatios.Get(simType)
		}
	}

	// rank combined sims
	for entry, simRank := range util.CalculateRanking(true, data, func(x *rankEntry3) float64 { return x.simScore }) {
		entry.targetRank = simRank
	}
}

func (ranker *RankingStatWeightProcess3) makeDataListEntryColumns(data []rankEntry3) {
	maxRank := float64(len(data) - 1)

	sumRanks := utilhighs.ConstraintRowBuild{Debug: "sumRanks"}
	for entry := range util.ForPointer(data) {
		ranker.makeEntryColumns(entry, maxRank)
		sumRanks.Add(entry.rankColumn, 1)
	}

	expectedSum := float64(len(data)) * float64(len(data)-1) / 2.0
	allowedSumRange := 0.0 // a bit of flex for rounding, as well as just to support easier solving
	sumRanks.Finish(ranker.input, expectedSum-allowedSumRange, expectedSum+allowedSumRange)
}

func (ranker *RankingStatWeightProcess3) makeDataListPairRules(data []rankEntry3) {
	eachCheckCount := 2
	for a := 0; a < len(data); a++ {
		for b := a + 1; b < min(a+eachCheckCount, len(data)); b++ {
			ranker.makeEntryPairSequenceConstraints(&data[a], &data[b], a, b)
		}
	}

	// TODO try out ranges of ranks, each times their own 2power/prime, and desirable sum, which would imply exact ordering
}

func (ranker *RankingStatWeightProcess3) makeEntryColumns(entry *rankEntry3, maxRank float64) {
	// these scores are meaningless in themselves, at least in value terms
	// however their increasing sequence should correlate to combinedSimRankScore
	// which is what we'll optimise for
	rankStr := strconv.FormatInt(int64(entry.targetRank), 10)
	entry.scoreColumn = ranker.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugText("score-"+rankStr))

	scoreRow := utilhighs.ConstraintRowBuild{Debug: "scoreRow"}
	for _, statType := range G_RequiredStats {
		weightColumn := ranker.weightColumns[statType]
		statValue := entry.data.TotalStat.GetFloat(statType)
		statScale := ranker.scaleStats[statType]

		scoreRow.Add(weightColumn, statValue*statScale)
	}
	scoreRow.Add(entry.scoreColumn, -1)
	scoreRow.Finish(ranker.input, 0, 0)

	entry.rankColumn = ranker.input.CreateColumnGeneral(highs.Integer, 0, maxRank, utilhighs.DebugText("derivedRank-"+rankStr))
	entry.rankDiffAbsColumn = ranker.input.CreateColumnWithOutput(highs.Integer, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugText("rankDiffAbs-"+rankStr))

	targetRank := float64(entry.targetRank)
	utilhighs.AbsoluteValueFromDiffOneToConst(ranker.input, entry.rankColumn, 1, targetRank, entry.rankDiffAbsColumn, "")

	// entry.rankDiffColumn = ranker.input.CreateColumnGeneral(highs.Integer, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugText("rankDiff-"+rankStr))
	// rankDiffColumn:= ranker.input.CreateColumnGeneral(highs.Integer, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugText("rankDiff-"+rankStr))
	// rankDiff := utilhighs.ConstraintRowBuild{Debug: "rankDiff"}
	// rankDiff.Add(entry.rankColumn, 1)
	// // rankDiff.Add(entry.rankDiffColumn, -1)
	// rankDiff.Add(rankDiffColumn, -1)
	// rankDiff.Finish(ranker.input, targetRank, targetRank)
}

// parameters don't imply order
func (ranker *RankingStatWeightProcess3) makeEntryPairSequenceConstraints(one *rankEntry3, two *rankEntry3, indexOne, indexTwo int) {
	// so we could totally do a boolean thing where scoreA>scoreB then implies rankA>rankB
	// would need all possible pairs connected, but would then force solver to make a full integer order
	isGreaterScore := ranker.input.CreateColumnBool(utilhighs.DebugText("isGreaterScore"))
	isGreaterRank := ranker.input.CreateColumnBool(utilhighs.DebugText("isGreaterRank"))
	sequenceDiff := ranker.input.CreateColumnWithOutput(highs.Integer, 0, 1, 1, utilhighs.DebugText("sequenceDiff"))

	utilhighs.ColumnIsGreaterOrEqualColumn(ranker.input, one.scoreColumn, two.scoreColumn, isGreaterScore, c_RankLargeScore)
	utilhighs.ColumnIsGreaterOrEqualColumn(ranker.input, one.rankColumn, two.rankColumn, isGreaterRank, c_RankLargeRank)
	utilhighs.Xor(ranker.input, isGreaterRank, isGreaterScore, sequenceDiff)

	ranker.pairLinks.Put(indexOne, indexTwo, rankPair3{
		indexOne:       indexOne,
		indexTwo:       indexTwo,
		isGreaterScore: isGreaterScore,
		isGreaterRank:  isGreaterRank,
		sequenceDiff:   sequenceDiff,
	})
}

// we do a complete ranking just on the strength stat
func (ranker *RankingStatWeightProcess3) setupDumbInitialSolution(data []rankEntry3) {
	for statType, colWeight := range ranker.weightColumns {
		if statType == stats.Stat_Strength {
			ranker.input.SetInitialSolutionValue(colWeight, 1)
		} else {
			ranker.input.SetInitialSolutionValue(colWeight, 0)
		}
	}

	statScale := ranker.scaleStats[stats.Stat_Strength]
	for entry, dumbRank := range util.CalculateRanking(true, data, func(x *rankEntry3) float64 { return x.data.TotalStat.GetFloat(stats.Stat_Strength) }) {
		thisValue := entry.data.TotalStat.GetFloat(stats.Stat_Strength)
		scaledValue := thisValue * statScale
		ranker.input.SetInitialSolutionValue(entry.scoreColumn, scaledValue)
		ranker.input.SetInitialSolutionValue(entry.rankColumn, float64(dumbRank))
		diff := float64(dumbRank) - float64(entry.targetRank)
		// ranker.input.SetInitialSolutionValue(entry.rankDiffColumn, diff)
		ranker.input.SetInitialSolutionValue(entry.rankDiffAbsColumn, math.Abs(diff))
	}

	// TODO check if we even use pairs in dumb
	for pair := range ranker.pairLinks.SeqValues() {
		one, two := &data[pair.indexOne], &data[pair.indexTwo]
		scoreOne, scoreTwo := ranker.input.GetInitialSolutionValue(one.scoreColumn), ranker.input.GetInitialSolutionValue(two.scoreColumn)
		rankOne, rankTwo := ranker.input.GetInitialSolutionValue(one.rankColumn), ranker.input.GetInitialSolutionValue(two.rankColumn)
		if scoreTwo >= scoreOne {
			ranker.input.SetInitialSolutionValue(pair.isGreaterScore, 1)
			if rankTwo >= rankOne {
				ranker.input.SetInitialSolutionValue(pair.isGreaterRank, 1)
				ranker.input.SetInitialSolutionValue(pair.sequenceDiff, 0)
			} else {
				ranker.input.SetInitialSolutionValue(pair.isGreaterRank, 0)
				ranker.input.SetInitialSolutionValue(pair.sequenceDiff, 1)
			}
		} else {
			ranker.input.SetInitialSolutionValue(pair.isGreaterScore, 0)
			if rankTwo >= rankOne {
				ranker.input.SetInitialSolutionValue(pair.isGreaterRank, 1)
				ranker.input.SetInitialSolutionValue(pair.sequenceDiff, 1)
			} else {
				ranker.input.SetInitialSolutionValue(pair.isGreaterRank, 0)
				ranker.input.SetInitialSolutionValue(pair.sequenceDiff, 0)
			}
		}
	}

	ranker.input.ValidateInitialSolutionState()
}

func (ranker *RankingStatWeightProcess3) setupInitialSolutionFromPrevious(solution *highs.Solution) {
	ranker.input.ClearInitialSolutionValue()
	for i := range solution.ColValues {
		ranker.input.SetInitialSolutionValue(utilhighs.ColumnIndex(i), solution.ColValues[i])
	}
	ranker.input.ValidateInitialSolutionState()
}

// data []rankEntry3, weights map[stats.StatType]float64
func (ranker *RankingStatWeightProcess3) setupInitialSolutionFromPreviousWeightOnly(solution *highs.Solution) {
	internalWeights := make(map[stats.StatType]float64)
	for statType, colWeight := range ranker.weightColumns {
		weight := solution.ColValues[colWeight]
		ranker.input.SetInitialSolutionValue(colWeight, weight)
		internalWeights[statType] = weight
	}

	// statScale := ranker.scaleStats[stats.Stat_Strength]
	// for entry, dumbRank := range util.CalculateRanking(true, data, func(x *rankEntry3) float64 { return x.data.TotalStat.GetFloat(stats.Stat_Strength) }) {
	// 	thisValue := entry.data.TotalStat.GetFloat(stats.Stat_Strength)
	// 	scaledValue := thisValue * statScale
	// 	ranker.input.SetInitialSolutionValue(entry.scoreColumn, scaledValue)
	// 	ranker.input.SetInitialSolutionValue(entry.rankColumn, float64(dumbRank))
	// 	diff := float64(dumbRank) - float64(entry.targetRank)
	// 	// ranker.input.SetInitialSolutionValue(entry.rankDiffColumn, diff)
	// 	ranker.input.SetInitialSolutionValue(entry.rankDiffAbsColumn, math.Abs(diff))
	// }
}

func (ranker *RankingStatWeightProcess3) extractAndReportSolution(solution *highs.Solution) WeightResult {
	ranker.input.DebugPrintColumns(solution, ranker.printer)

	ranker.printer.Println("WEIGHTS")

	statWeightResult := WeightResult_Make()
	for _, statType := range G_RequiredStats {
		weightColumn := ranker.weightColumns[statType]
		statScale := ranker.scaleStats[statType]

		modelWeight := solution.ColValues[weightColumn]
		usableWeight := modelWeight / statScale

		statWeightResult.Put(statType, usableWeight)
	}

	divideBy := statWeightResult.Get(stats.Stat_Strength)
	for _, statType := range G_RequiredStats {
		statWeightResult.Put(statType, statWeightResult.Get(statType)/divideBy)
	}

	// ranker.reportRankingOfInputs(statWeightResult)

	return statWeightResult
}

// func (ranker *RankingStatWeightProcess3) reportRankingOfInputs(statWeightResult map[stats.StatType]float64) {
// 	ranker.printer.Println("INPUT CHECK (index, combinedSimRank, calcStat)")
// 	for i, entry := range ranker.data {
// 		ranker.printer.Printf("%4d %8f %8f\n", i, entry.combinedSimRankScore, calcStatScore(entry.data, statWeightResult))
// 	}
// }
