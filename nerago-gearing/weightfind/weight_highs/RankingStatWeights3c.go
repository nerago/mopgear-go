package weight_highs

import (
	"fmt"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/tools"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/simrank"
	"paladin_gearing_go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_rank3c_targetTotalWeight = 1.0
	c_rank3c_largeWeight       = 10.0
)

type RankingStatWeightProcess3c struct {
	printer        *util.PrintRecorder
	timeoutSeconds int

	targetRatios    weight_types.SimPriorityBasic
	requiredStats   []stats.StatType
	requiredSims    []stats.SimType
	dataAllOriginal []*rankEntry3c
	dataSample      []*rankEntry3c

	build *util_highs.LinearBuilder

	weightColumns map[stats.StatType]util_highs.ColumnIndex
	pairLinks     util_collection.MapMapDiagonal[int, rankPair3c]
}

type rankEntry3c struct {
	data        *weight_types.WeightInput
	simScore    float64
	targetRange *util_collection.HiLoInt
	scoreColumn util_highs.ColumnIndex
}

func (re *rankEntry3c) GetSimData() *stats.SimData {
	return &re.data.SimResult
}

func (re *rankEntry3c) GetSimScore() float64 {
	return re.simScore
}

func (re *rankEntry3c) IncrementSimScore(add float64) {
	re.simScore += add
}

func (re *rankEntry3c) ResetSimScore() {
	re.simScore = 0
}

func (re *rankEntry3c) SetSimRankRange(targetRange *util_collection.HiLoInt) {
	re.targetRange = targetRange
}

func (re *rankEntry3c) GetSimRankRange() *util_collection.HiLoInt {
	return re.targetRange
}

type rankPair3c struct {
	indexOne, indexTwo int
	scoreSlackColumn   util_highs.ColumnIndex
}

func (ranker *RankingStatWeightProcess3c) Init(printer *util.PrintRecorder, timeoutSeconds int) {
	ranker.printer = printer
	ranker.timeoutSeconds = timeoutSeconds
}

func (ranker *RankingStatWeightProcess3c) SupplyData(inputData []weight_types.WeightInput) {
	ranker.dataAllOriginal = util_collection.MapSliceAsNew(inputData, func(input *weight_types.WeightInput) *rankEntry3c {
		return &rankEntry3c{
			data:        input,
			scoreColumn: -1,
		}
	})
}

func (ranker *RankingStatWeightProcess3c) SetRequiredStats(requiredStats []stats.StatType) {
	ranker.requiredStats = requiredStats
}

func (ranker *RankingStatWeightProcess3c) SetTargetRatios(targetRatios weight_types.SimPriorityBasic) {
	ranker.targetRatios = targetRatios
	ranker.requiredSims = targetRatios.SimTypes()
}

func (ranker *RankingStatWeightProcess3c) newBuilder() {
	ranker.build = new(util_highs.LinearBuilder)
	ranker.build.Minimise = true
	ranker.build.TimeLimitSeconds = ranker.timeoutSeconds
	ranker.build.Solver = util_highs.Solver_Force_IPX
}

func (ranker *RankingStatWeightProcess3c) RunMultiRound() *util_async.FutureCancellable[weight_types.WeightResult] {

	// FIRST ROUND: minimal data, no initial values
	ranker.dataSample = takeDataSample_Start(ranker.dataAllOriginal, 64)
	ranker.newBuilder()
	ranker.prepareRankings()
	ranker.createWeightColumns()
	ranker.doAlgos()
	ranker.setupDumbInitialSolution()

	stopwatch := util.StopwatchMakeStopped()
	solution1Future := ranker.build.RunHighsFuture(stopwatch)

	solution2Future := util_async.FutureCancellable_MapToFuture(solution1Future, func(linearResult1 util_highs.LinearResult) *util_async.FutureCancellable[util_highs.LinearResult] {
		solution1 := linearResult1.GetSolutionAndSaveLog(ranker.printer)
		_ = ranker.extractAndReportSolution(solution1)

		// FULL RUN
		ranker.dataSample = ranker.dataAllOriginal
		ranker.newBuilder()
		ranker.prepareRankings()
		ranker.createWeightColumns()
		ranker.doAlgos()
		if solution1.HasSolution() {
			ranker.setupInitialSolutionFromPreviousSolutionWeights(solution1)
		} else {
			return nil
		}
		return ranker.build.RunHighsFuture(stopwatch)
	})

	return util_async.FutureCancellable_MapValue(solution2Future, func(linearResult2 util_highs.LinearResult) (weight_types.WeightResult, bool) {
		solution2 := linearResult2.GetSolutionAndSaveLog(ranker.printer)
		weight := ranker.extractAndReportSolution(solution2)
		return weight_types.WeightResult{Weight: &weight, SolveTime: stopwatch.Elapsed(), Status: solution2.Status}, true
	})
}

func (ranker *RankingStatWeightProcess3c) createWeightColumns() {
	lo := -c_rank3c_largeWeight
	hi := c_rank3c_largeWeight

	sumWeights := util_highs.ConstraintRow{Debug: "sumWeights"}
	ranker.weightColumns = make(map[stats.StatType]util_highs.ColumnIndex)
	for _, statType := range ranker.requiredStats {
		colWeight := ranker.build.CreateColumnGeneral(highs.Continuous, lo, hi, util_highs.DebugString{Text: "WEIGHT " + statType.Name()})
		ranker.weightColumns[statType] = colWeight
		sumWeights.Add(colWeight, 1)
	}

	sumWeights.Build(ranker.build, c_rank3c_targetTotalWeight, c_rank3c_targetTotalWeight)
}

func (ranker *RankingStatWeightProcess3c) prepareRankings() {
	simrank.RankSimsStatisticalForRanged(ranker.requiredSims, ranker.dataSample, &ranker.targetRatios)
}

func (ranker *RankingStatWeightProcess3c) doAlgos() {
	ranker.makeDataListEntryColumns()

	ranker.pairLinks.Clear()
	for baseIndex := 0; baseIndex < len(ranker.dataSample)-1; baseIndex++ {
		ranker.makeEntryPairCheckScoreOrderMatchesTargetOrderWithSlackVar(ranker.dataSample[baseIndex], ranker.dataSample[baseIndex+1], baseIndex, baseIndex+1)
	}
}

func (ranker *RankingStatWeightProcess3c) makeDataListEntryColumns() {
	for _, entry := range ranker.dataSample {
		ranker.makeScoreColumn(entry)
	}
}

func (ranker *RankingStatWeightProcess3c) makeScoreColumn(entry *rankEntry3c) {
	entry.scoreColumn = ranker.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), nil)

	scoreRow := util_highs.ConstraintRow{Debug: "scoreRow"}
	for _, statType := range ranker.requiredStats {
		weightColumn := ranker.weightColumns[statType]
		statValue := entry.data.TotalStat.GetFloat(statType)
		scoreRow.Add(weightColumn, statValue)
	}
	scoreRow.Add(entry.scoreColumn, -1)
	scoreRow.Build(ranker.build, 0, 0)
}

func (ranker *RankingStatWeightProcess3c) makeEntryPairCheckScoreOrderMatchesTargetOrderWithSlackVar(one *rankEntry3c, two *rankEntry3c, indexOne, indexTwo int) {
	if one.targetRange.Overlap(*two.targetRange) {
		return
	}

	debug := fmt.Sprintf("-%d-%d", indexOne, indexTwo)
	slack := ranker.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.InfPos(), 1, util_highs.DebugText("slack"+debug))

	if one.targetRange.Lo > two.targetRange.Hi {
		// one is entirely above two
		row := util_highs.ConstraintRow{Debug: "row" + debug}
		row.Add(one.scoreColumn, 1)
		row.Add(two.scoreColumn, -1)
		row.Add(slack, 1)
		row.Build(ranker.build, 0, util_highs.InfPos())
	} else if two.targetRange.Lo > one.targetRange.Hi {
		// two is entirely above one
		row := util_highs.ConstraintRow{Debug: "row" + debug}
		row.Add(two.scoreColumn, 1)
		row.Add(one.scoreColumn, -1)
		row.Add(slack, 1)
		row.Build(ranker.build, 0, util_highs.InfPos())
	} else {
		panic("unexpected equal ranks, should already be checked in overlap")
	}

	ranker.pairLinks.Put(indexOne, indexTwo, rankPair3c{
		indexOne:         indexOne,
		indexTwo:         indexTwo,
		scoreSlackColumn: slack,
	})
}

// we do a complete ranking using all weights set equal
func (ranker *RankingStatWeightProcess3c) setupDumbInitialSolution() {
	internalWeights := weight_types.Weight1Basic_Make()
	for _, statType := range ranker.requiredStats {
		internalWeights.Put(statType, 1)
	}
	ranker.setupInitialFromInternalWeights(internalWeights)
}

func (ranker *RankingStatWeightProcess3c) setupInitialSolutionFromPreviousSolutionWeights(solution *highs.Solution) {
	internalWeights := weight_types.Weight1Basic_Make()
	for statType, colWeight := range ranker.weightColumns {
		weight := solution.ColValues[colWeight]
		internalWeights.Put(statType, weight)
	}
	ranker.setupInitialFromInternalWeights(internalWeights)
}

func (ranker *RankingStatWeightProcess3c) setupInitialSolutionFromExternal(weights weight_types.Weight1Basic) {
	internalWeights := weight_types.Weight1Basic_Make()
	for _, statType := range ranker.requiredStats {
		value := weights.Get(statType)
		internalWeights.Put(statType, value)
	}
	ranker.setupInitialFromInternalWeights(internalWeights)
}

func (ranker *RankingStatWeightProcess3c) setupInitialFromInternalWeights(internalWeights weight_types.Weight1Basic) {
	ranker.build.ClearInitialSolutionValue()
	if !internalWeights.IsEmpty() {
		internalWeights = internalWeights.ScaleForTotalSum(c_rank3c_targetTotalWeight)

		for statType, colWeight := range ranker.weightColumns {
			weight := internalWeights.Get(statType)
			ranker.build.SetInitialSolutionValue(colWeight, weight)
		}

		ranker.setupInitialRemainingVariables(internalWeights)
		ranker.setupInitialPairsDetail()
	}

	ranker.build.ValidateInitialSolutionState()
}

func (ranker *RankingStatWeightProcess3c) setupInitialRemainingVariables(internalWeights weight_types.Weight1Basic) {
	for _, entry := range ranker.dataSample {
		initialStatScore := internalWeights.CalcStatScore(&entry.data.TotalStat)
		ranker.build.SetInitialSolutionValue(entry.scoreColumn, initialStatScore)
	}
}

func (ranker *RankingStatWeightProcess3c) setupInitialPairsDetail() {
	for pair := range ranker.pairLinks.SeqValues() {
		one, two := ranker.dataSample[pair.indexOne], ranker.dataSample[pair.indexTwo]

		scoreOne := ranker.build.GetInitialSolutionValue(one.scoreColumn)
		scoreTwo := ranker.build.GetInitialSolutionValue(two.scoreColumn)

		var slack float64
		if one.targetRange.Overlap(*two.targetRange) {
			slack = 0
		} else if one.targetRange.Lo > two.targetRange.Hi {
			// one is entirely above two
			if scoreOne > scoreTwo {
				slack = 0
			} else {
				slack = scoreTwo - scoreOne
			}
		} else if two.targetRange.Lo > one.targetRange.Hi {
			// two is entirely above one
			if scoreTwo > scoreOne {
				slack = 0
			} else {
				slack = scoreOne - scoreTwo
			}
		} else {
			panic("unexpected equal ranks")
		}

		ranker.build.SetInitialSolutionValue(pair.scoreSlackColumn, slack)
	}
}

func (ranker *RankingStatWeightProcess3c) extractAndReportSolution(solution *highs.Solution) weight_types.Weight1Basic {
	ranker.build.DebugPrintColumns(solution, ranker.printer)

	ranker.printer.Println("WEIGHTS")

	weight := weight_types.Weight1Basic_Make()
	for _, statType := range ranker.requiredStats {
		weightColumn := ranker.weightColumns[statType]

		modelWeight := solution.ColValues[weightColumn]
		weight.Put(statType, modelWeight)
	}

	weight.NormalizeForBase(ranker.requiredStats)
	tools.WriteWeightString(&weight, ranker.printer)
	return weight
}
