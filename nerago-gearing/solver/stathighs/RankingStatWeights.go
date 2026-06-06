package stathighs

import (
	"cmp"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"slices"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_RankLargeWeight = 50.0
	// c_RankHighDiff         = 1000.0
	// c_RankOutputPerInclude = -0.1
)

type RankingStatWeightProcess struct {
	printer *util.PrintRecorder

	targetRatios simulate.SimData
	data         []rankEntry

	input *utilhighs.InputBuilder

	// linearEquationDiff int
	// linearInclude      int

	scaleStats    map[stats.StatType]float64
	weightColumns map[stats.StatType]utilhighs.ColumnIndex

	// minimumIncludeRate float64
	// includeColumns     []utilhighs.ColumnIndex
	// includeCountRow    utilhighs.ConstraintRowBuild
}

type rankEntry struct {
	data *WeightInput

	simRanks             map[simulate.SimType]int
	combinedSimRankScore float64

	scoreColumn utilhighs.ColumnIndex
}

func (ranker *RankingStatWeightProcess) Init(printer *util.PrintRecorder) {
	ranker.printer = printer
}

func (ranker *RankingStatWeightProcess) SupplyData(inputData []WeightInput) {
	ranker.scaleStats = chooseStatScaling(inputData, ranker.printer)
	ranker.data = util.MapSliceAsNew(inputData, func(input *WeightInput) rankEntry {
		return rankEntry{data: input}
	})
}

func (ranker *RankingStatWeightProcess) SetTargetRatios(targetRatios simulate.SimData) {
	ranker.targetRatios = targetRatios
}

// func (ranker *RankStatWeightProcess) SetMinimumIncludeRate(percent float64) {
// 	ranker.minimumIncludeRate = percent
// }

func (ranker *RankingStatWeightProcess) Run() map[stats.StatType]float64 {
	ranker.input = new(utilhighs.InputBuilder)
	ranker.input.Minimise = true
	ranker.input.Solver = "ipx"
	ranker.input.DisablePreSolve = true
	// ranker.input.Solver = "pdlp"
	// ranker.input.Solver = "hipdlp"

	ranker.createWeightColumns()
	ranker.prepareRankings()
	ranker.processData()

	// ranker.includeCountRow.Finish(ranker.input, float64(len(ranker.inputData))*ranker.minimumIncludeRate, utilhighs.C_PlusInf)

	solution, log := ranker.input.RunHighs()
	ranker.printer.AppendOther(log)

	return ranker.extractAndReportSolution(solution)
}

func (ranker *RankingStatWeightProcess) createWeightColumns() {
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

	// this assumes that a positive strength will work for the scoring system, should be true in most situations
	// if not then i'd be concerned more about garbage input data
	// setStrength := utilhighs.ConstraintRowBuild{}
	// setStrength.Add(ranker.weightColumns[stats.Stat_Strength], 1)
	// setStrength.Finish(ranker.input, 1.0, 1.0)
}

func (ranker *RankingStatWeightProcess) prepareRankings() {
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
		entry.simRanks = make(map[simulate.SimType]int)
		for simType, rankedValues := range rankedSims.SeqGroupsInternalSlice() {
			queryValue := entry.data.SimResult.Get(simType)

			rank := slices.Index(rankedValues, queryValue)
			entry.simRanks[simType] = rank

			entry.combinedSimRankScore += float64(rank) * ranker.targetRatios.Get(simType)
		}
	}

	// sort the data in ascending rank order
	slices.SortFunc(ranker.data, func(a, b rankEntry) int { return cmp.Compare(a.combinedSimRankScore, b.combinedSimRankScore) })
}

func (ranker *RankingStatWeightProcess) processData() {
	for entry := range util.ForPointer(ranker.data) {
		ranker.processDataEntry(entry)
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

	// several adjacent
	eachCheckCount := 20
	for a := 0; a < len(ranker.data); a++ {
		for b := a + 1; b < min(a+eachCheckCount, len(ranker.data)); b++ {
			ranker.processEntrySequencePair(&ranker.data[a], &ranker.data[b])
		}
	}
}

func (ranker *RankingStatWeightProcess) processDataEntry(entry *rankEntry) {
	// these scores are meaningless in themselves, at least in value terms
	// however their increasing sequence should correlate to combinedSimRankScore
	// which is what we'll optimise for
	entry.scoreColumn = ranker.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugText("score"))

	scoreRow := utilhighs.ConstraintRowBuild{}
	for _, statType := range G_RequiredStats {
		weightColumn := ranker.weightColumns[statType]
		statValue := entry.data.TotalStat.GetFloat(statType)
		statScale := ranker.scaleStats[statType]

		scoreRow.Add(weightColumn, statValue*statScale)
	}

	scoreRow.Add(entry.scoreColumn, -1)
	scoreRow.Finish(ranker.input, 0, 0)
}

// we want to optimise for higher.score > lower.score
func (ranker *RankingStatWeightProcess) processEntrySequencePair(lower *rankEntry, higher *rankEntry) {
	offsetColumn := ranker.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugText("offset"))

	// if lower <= higher then it will trivially pass the >= 0 check. offset will be free but under minimise pressure so effectively zero
	// if lower > higher then it will initially fail the >= 0 check, and need an extra boost from offset to get over the line
	compareRow := utilhighs.ConstraintRowBuild{}
	compareRow.Add(lower.scoreColumn, -1)
	compareRow.Add(higher.scoreColumn, 1)
	compareRow.Add(offsetColumn, 1)
	compareRow.Finish(ranker.input, 0, utilhighs.C_PlusInf)
}

func (ranker *RankingStatWeightProcess) extractAndReportSolution(solution *highs.Solution) map[stats.StatType]float64 {
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

	ranker.reportRankingOfInputs(statWeightResult)

	return statWeightResult
}

func (ranker *RankingStatWeightProcess) reportRankingOfInputs(statWeightResult map[stats.StatType]float64) {
	for i, entry := range ranker.data {
		ranker.printer.Printf("%4d %8f %8f\n", i, entry.combinedSimRankScore, calcStatScore(entry.data, statWeightResult))
	}
}

func calcStatScore(input *WeightInput, statWeights map[stats.StatType]float64) float64 {
	total := 0.0
	for statType, weightValue := range statWeights {
		total += input.TotalStat.GetFloat(statType) * weightValue
	}
	return total
}
