package stathighs

import (
	"cmp"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"slices"
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

	scoreColumn utilhighs.ColumnIndex
	rankColumn  utilhighs.ColumnIndex
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
		colDetailWeight := ranker.input.CreateColumnGeneral(highs.Continuous, lo, hi, utilhighs.DebugString{Text: "WEIGHT " + statType.Name()})
		ranker.weightColumns[statType] = colDetailWeight
		sumWeights.Add(colDetailWeight, 1)
	}

	sumWeights.Finish(ranker.input, 1.0, utilhighs.C_PlusInf) // force positive and non-zero result
}

func (ranker *RankingStatWeightProcess3) prepareRankings() {
	// collect individual values of each sim type
	rankedSims := util.MapSlice[simulate.SimType, float64]{}
	for entry := range util.ForPointer(ranker.data) {
		for _, simType := range G_RequiredSims {
			rankedSims.Add(simType, entry.data.SimResult.Get(simType))
		}
	}

	// sort the values
	rankedSims.MapInternalSlicesAll(func(simType simulate.SimType, inner []float64) []float64 {
		if simType.IsHighGood() {
			// ascending, so that later indexes are better and worth more rank
			slices.Sort(inner)
		} else {
			// decending, later entries are smaller numerically, but worth more in index
			slices.SortFunc(inner, func(a, b float64) int { return cmp.Compare(b, a) })
		}
		return inner
	})

	// set ranks
	for entry := range util.ForPointer(ranker.data) {
		for simType, rankedValues := range rankedSims.SeqGroupsInternalSlice() {
			queryValue := entry.data.SimResult.Get(simType)

			rank := slices.Index(rankedValues, queryValue)
			if rank == -1 {
				panic("missing value")
			}
			entry.simScore += float64(rank) * ranker.targetRatios.Get(simType)
		}
	}

	// sort the data in ascending rank order
	slices.SortFunc(ranker.data, func(a, b rankEntry3) int { return cmp.Compare(a.simScore, b.simScore) })
	for i := range ranker.data {
		ranker.data[i].targetRank = i
	}
}

func (ranker *RankingStatWeightProcess3) processData() {
	sumRanks := utilhighs.ConstraintRowBuild{}
	for entry := range util.ForPointer(ranker.data) {
		ranker.processDataEntryPlusRankCompareToExpected(entry)
		sumRanks.Add(entry.rankColumn, 1)
	}

	expectedSum := float64(len(ranker.data)) * float64(len(ranker.data) - 1) / 2.0
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

	// TODO give it an initial solution to ranks

	entry.rankColumn = ranker.input.CreateColumnGeneral(highs.Integer, 0, float64(len(ranker.data)-1), utilhighs.DebugText("derivedRank-"+rankStr))
	rankDiffColumn := ranker.input.CreateColumnGeneral(highs.Integer, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugText("rankDiff-"+rankStr))
	rankDiffAbsColumn := ranker.input.CreateColumnWithOutput(highs.Integer, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugText("rankDiffAbs-"+rankStr))

	targetRank := float64(entry.targetRank)
	rankDiff := utilhighs.ConstraintRowBuild{}
	rankDiff.Add(entry.rankColumn, 1)
	rankDiff.Add(rankDiffColumn, -1)
	rankDiff.Finish(ranker.input, targetRank, targetRank)
	utilhighs.AbsoluteValue(ranker.input, rankDiffColumn, rankDiffAbsColumn)

	// ranker.input.SetInitialSolutionValue(entry.rankColumn, float64(entry.targetRank))
}

// // parameters don't imply order
func (ranker *RankingStatWeightProcess3) processEntrySequencePairToDerivedRank(one *rankEntry3, two *rankEntry3) {
	// so we could totally do a boolean thing where scoreA>scoreB then implies rankA>rankB
	// would need all possible pairs connected, but would then force solver to make a full integer order
	isGreater := ranker.input.CreateColumnBool(utilhighs.DebugText("isGreater"))
	// ranker.input.SetInitialSolutionValue(isGreater, 1)

	utilhighs.ColumnIsGreaterOrEqualColumn(ranker.input, one.scoreColumn, two.scoreColumn, isGreater, c_RankLargeScore)

	utilhighs.ColumnIsGreaterOrEqualColumn(ranker.input, one.rankColumn, two.rankColumn, isGreater, c_RankLargeRank)
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
