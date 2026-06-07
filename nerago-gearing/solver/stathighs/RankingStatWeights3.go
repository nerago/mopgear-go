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
	data         []rankEntry3

	input *utilhighs.InputBuilder

	scaleStats    map[stats.StatType]float64
	weightColumns map[stats.StatType]utilhighs.ColumnIndex
}

type rankEntry3 struct {
	data *WeightInput

	simScore   float64
	targetRank int

	scoreColumn       utilhighs.ColumnIndex
	rankColumn        utilhighs.ColumnIndex
	rankDiffColumn    utilhighs.ColumnIndex
	rankDiffAbsColumn utilhighs.ColumnIndex
}

func (ranker *RankingStatWeightProcess3) Init(printer *util.PrintRecorder) {
	ranker.printer = printer
}

func (ranker *RankingStatWeightProcess3) SupplyData(inputData []WeightInput) {
	ranker.scaleStats = chooseStatScaling(inputData, ranker.printer)
	ranker.data = util.MapSliceAsNew(inputData, func(input *WeightInput) rankEntry3 {
		return rankEntry3{data: input}
	})
}

func (ranker *RankingStatWeightProcess3) SetTargetRatios(targetRatios simulate.SimData) {
	ranker.targetRatios = targetRatios
}

func (ranker *RankingStatWeightProcess3) Run() map[stats.StatType]float64 {
	ranker.input = new(utilhighs.InputBuilder)
	ranker.input.Minimise = true

	ranker.createWeightColumns()
	ranker.prepareRankings()
	ranker.processData()

	ranker.setupDumbInitialSolution()

	solution, log := ranker.input.RunHighs()
	ranker.printer.AppendOther(log)

	return ranker.extractAndReportSolution(solution)
}

func (ranker *RankingStatWeightProcess3) createWeightColumns() {
	lo := -c_RankLargeWeight
	hi := c_RankLargeWeight

	sumWeights := utilhighs.ConstraintRowBuild{}
	ranker.weightColumns = make(map[stats.StatType]utilhighs.ColumnIndex)
	for _, statType := range G_RequiredStats {
		colWeight := ranker.input.CreateColumnGeneral(highs.Continuous, lo, hi, utilhighs.DebugString{Text: "WEIGHT " + statType.Name()})
		ranker.weightColumns[statType] = colWeight
		sumWeights.Add(colWeight, 1)
	}

	sumWeights.Finish(ranker.input, 1.0, utilhighs.C_PlusInf) // force positive and non-zero result
}

func (ranker *RankingStatWeightProcess3) prepareRankings() {
	// score each sim
	for _, simType := range G_RequiredSims {
		for entry, simDetailRank := range util.CalculateRanking(simType.IsHighGood(), ranker.data, func(x *rankEntry3) float64 { return x.data.SimResult.Get(simType) }) {
			entry.simScore += float64(simDetailRank) * ranker.targetRatios.Get(simType)
		}
	}

	// rank combined sims
	for entry, simRank := range util.CalculateRanking(true, ranker.data, func(x *rankEntry3) float64 { return x.simScore }) {
		entry.targetRank = simRank
	}
}

func (ranker *RankingStatWeightProcess3) processData() {
	sumRanks := utilhighs.ConstraintRowBuild{}
	for entry := range util.ForPointer(ranker.data) {
		ranker.processDataEntryPlusRankCompareToExpected(entry)
		sumRanks.Add(entry.rankColumn, 1)
	}

	expectedSum := float64(len(ranker.data)) * float64(len(ranker.data)-1) / 2.0
	allowedSumRange := 0.0 // a bit of flex for rounding, as well as just to support easier solving
	sumRanks.Finish(ranker.input, expectedSum-allowedSumRange, expectedSum+allowedSumRange)

	eachCheckCount := 2
	for a := 0; a < len(ranker.data); a++ {
		for b := a + 1; b < min(a+eachCheckCount, len(ranker.data)); b++ {
			ranker.processEntrySequencePairToDerivedRank(&ranker.data[a], &ranker.data[b])
		}
	}
}

func (ranker *RankingStatWeightProcess3) processDataEntryPlusRankCompareToExpected(entry *rankEntry3) {
	// these scores are meaningless in themselves, at least in value terms
	// however their increasing sequence should correlate to combinedSimRankScore
	// which is what we'll optimise for
	rankStr := strconv.FormatInt(int64(entry.targetRank), 10)
	entry.scoreColumn = ranker.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugText("score-"+rankStr))

	scoreRow := utilhighs.ConstraintRowBuild{}
	for _, statType := range G_RequiredStats {
		weightColumn := ranker.weightColumns[statType]
		statValue := entry.data.TotalStat.GetFloat(statType)
		statScale := ranker.scaleStats[statType]

		scoreRow.Add(weightColumn, statValue*statScale)
	}
	scoreRow.Add(entry.scoreColumn, -1)
	scoreRow.Finish(ranker.input, 0, 0)

	entry.rankColumn = ranker.input.CreateColumnGeneral(highs.Integer, 0, float64(len(ranker.data)-1), utilhighs.DebugText("derivedRank-"+rankStr))
	entry.rankDiffColumn = ranker.input.CreateColumnGeneral(highs.Integer, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugText("rankDiff-"+rankStr))
	entry.rankDiffAbsColumn = ranker.input.CreateColumnWithOutput(highs.Integer, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugText("rankDiffAbs-"+rankStr))

	targetRank := float64(entry.targetRank)
	rankDiff := utilhighs.ConstraintRowBuild{}
	rankDiff.Add(entry.rankColumn, 1)
	rankDiff.Add(entry.rankDiffColumn, -1)
	rankDiff.Finish(ranker.input, targetRank, targetRank)
	utilhighs.AbsoluteValue(ranker.input, entry.rankDiffColumn, entry.rankDiffAbsColumn)

}

// // parameters don't imply order
func (ranker *RankingStatWeightProcess3) processEntrySequencePairToDerivedRank(one *rankEntry3, two *rankEntry3) {
	// so we could totally do a boolean thing where scoreA>scoreB then implies rankA>rankB
	// would need all possible pairs connected, but would then force solver to make a full integer order
	isGreater := ranker.input.CreateColumnBool(utilhighs.DebugText("isGreater"))

	// dumb initial solution
	if two.data.TotalStat.GetFloat(stats.Stat_Strength) >= one.data.TotalStat.GetFloat(stats.Stat_Strength) {
		ranker.input.SetInitialSolutionValue(isGreater, 1)
	} else {
		ranker.input.SetInitialSolutionValue(isGreater, 0)
	}

	utilhighs.ColumnIsGreaterOrEqualColumn(ranker.input, one.scoreColumn, two.scoreColumn, isGreater, c_RankLargeScore)

	utilhighs.ColumnIsGreaterOrEqualColumn(ranker.input, one.rankColumn, two.rankColumn, isGreater, c_RankLargeRank)
}

// we do a complete ranking just on the strength stat
func (ranker *RankingStatWeightProcess3) setupDumbInitialSolution() {
	for statType, colWeight := range ranker.weightColumns {
		if statType == stats.Stat_Strength {
			ranker.input.SetInitialSolutionValue(colWeight, 1)
		} else {
			ranker.input.SetInitialSolutionValue(colWeight, 0)
		}
	}

	statScale := ranker.scaleStats[stats.Stat_Strength]
	for entry, dumbRank := range util.CalculateRanking(true, ranker.data, func(x *rankEntry3) float64 { return x.data.TotalStat.GetFloat(stats.Stat_Strength) }) {
		thisValue := entry.data.TotalStat.GetFloat(stats.Stat_Strength)
		scaledValue := thisValue * statScale
		ranker.input.SetInitialSolutionValue(entry.scoreColumn, scaledValue)
		ranker.input.SetInitialSolutionValue(entry.rankColumn, float64(dumbRank))
		diff := float64(dumbRank) - float64(entry.targetRank)
		ranker.input.SetInitialSolutionValue(entry.rankDiffColumn, diff)
		ranker.input.SetInitialSolutionValue(entry.rankDiffAbsColumn, math.Abs(diff))
	}
}

func (ranker *RankingStatWeightProcess3) extractAndReportSolution(solution *highs.Solution) map[stats.StatType]float64 {
	ranker.input.DebugPrintColumns(solution, ranker.printer)

	ranker.printer.Println("WEIGHTS")

	statWeightResult := make(map[stats.StatType]float64)
	for _, statType := range G_RequiredStats {
		weightColumn := ranker.weightColumns[statType]
		statScale := ranker.scaleStats[statType]

		modelWeight := solution.ColValues[weightColumn]
		usableWeight := modelWeight / statScale

		statWeightResult[statType] = usableWeight
	}

	divideBy := statWeightResult[stats.Stat_Strength]
	for _, statType := range G_RequiredStats {
		statWeightResult[statType] /= divideBy
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
