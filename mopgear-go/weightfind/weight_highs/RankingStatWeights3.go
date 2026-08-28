package weight_highs

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"strconv"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/tools"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_highs"
	"github.com/nerago/mopgear-go/weightfind/simrank"
	"github.com/nerago/mopgear-go/weightfind/util_weight"
	"github.com/nerago/mopgear-go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_rank3_scaleTarget         = 10.0
	c_rank3_initial_data_sample = 12
	c_rank3_target_total_weight = 1.0

	c_Rank3_LargeWeight = 10.0
	c_Rank3_LargeScore  = 500.0
	c_Rank3_LargeRank   = 10000.0
)

type RankingStatWeightProcess3 struct {
	printer *util.PrintRecorder
	timeout int

	targetRatios    weight_types.SimPriorityBasic
	requiredStats   []stats.StatType
	requiredSims    []stats.SimType
	dataAllOriginal []*rankEntry3
	dataSample      []*rankEntry3
	SCALE1          bool
	ALGO            int

	build *util_highs.LinearBuilder

	scaleStats    stats.StatTypeMap[float64]
	weightColumns stats.StatTypeMap[util_highs.ColumnIndex]
	pairLinks     util_collection.MapMap[int, int, rankPair3]
}

type rankEntry3 struct {
	weight_types.RankStatWeightsCommon

	InitialStatScore float64

	ScoreColumn       util_highs.ColumnIndex
	RankColumn        util_highs.ColumnIndex
	RankDiffAbsColumn util_highs.ColumnIndex
}

type rankPair3 struct {
	indexOne, indexTwo            int
	isGreaterScore, isGreaterRank util_highs.ColumnIndex
	isSequenceDiff                util_highs.ColumnIndex
}

func (ranker *RankingStatWeightProcess3) Init(printer *util.PrintRecorder, timeout int) {
	ranker.printer = printer
	ranker.timeout = timeout
}

func (ranker *RankingStatWeightProcess3) SupplyData(inputData []weight_types.WeightInput) {
	if ranker.SCALE1 {
		ranker.scaleStats = stats.StatTypeMap[float64]{}
		for _, statType := range stats.StatType_List {
			ranker.scaleStats.Put(statType, 1)
		}
	} else {
		ranker.scaleStats = util_weight.ChooseStatScalingBasic(inputData, c_rank3_scaleTarget, false, ranker.printer)
	}
	ranker.dataAllOriginal = util_collection.MapSliceAsNew(inputData, func(input *weight_types.WeightInput) *rankEntry3 {
		return &rankEntry3{
			RankStatWeightsCommon: weight_types.RankStatWeightsCommon{
				Data:       input,
				SimScore:   -1,
				TargetRank: -1,
			},
			ScoreColumn:       -1,
			RankColumn:        -1,
			RankDiffAbsColumn: -1,
			// rankDiffColumn:    -1,
		}
	})
}

func (ranker *RankingStatWeightProcess3) SetRequiredStats(requiredStats []stats.StatType) {
	ranker.requiredStats = requiredStats
}

func (ranker *RankingStatWeightProcess3) SetTargetRatios(targetRatios weight_types.SimPriorityBasic) {
	ranker.targetRatios = targetRatios
	ranker.requiredSims = targetRatios.SimTypes()
}

func (ranker *RankingStatWeightProcess3) makeBuilder() {
	ranker.build = new(util_highs.LinearBuilder)
	ranker.build.Minimise = true
	ranker.build.TimeLimitSeconds = ranker.timeout
	if ranker.ALGO >= 2 {
		ranker.build.Solver = util_highs.Solver_LP_USE_GPU
	} else {
		ranker.build.Solver = util_highs.Solver_MIP_Interior
	}
}

func (ranker *RankingStatWeightProcess3) Run() (*util_async.FutureCancellable[weight_types.WeightResult1], error) {
	// FIRST ROUND: minimal data, no initial values
	ranker.dataSample = util_collection.SliceSampleRandom(ranker.dataAllOriginal, c_rank3_initial_data_sample)
	ranker.prepare()
	ranker.setupDumbInitialSolution()

	stopwatch := util.StopwatchMakeStopped()
	solution1Future := ranker.build.RunHighsFuture(stopwatch)

	return util_async.FutureCancellable_MapValue(solution1Future, func(linearResult1 util_highs.LinearResult) weight_types.WeightResult1 {
		solution1, err1 := linearResult1.GetSolutionAndSaveLog(ranker.printer)
		if err1 != nil {
			return weight_types.WeightResult1MakeError(stopwatch.Elapsed(), err1)
		}
		if !solution1.HasSolution() {
			return weight_types.WeightResult1MakeError(stopwatch.Elapsed(), fmt.Errorf("first stage solution %v", solution1.Status))
		}

		// FULL RUN
		ranker.dataSample = ranker.dataAllOriginal
		ranker.prepare()
		ranker.setupInitialSolutionFromPreviousWeightOnly(solution1)

		solution2Future := ranker.build.RunHighsFuture(stopwatch)
		if linearResult2, hasResult2 := solution2Future.WaitForResult(); hasResult2 {
			if solution2, err2 := linearResult2.GetSolutionAndSaveLog(ranker.printer); err2 == nil {
				weight := ranker.extractAndReportSolution(solution2)
				return weight_types.WeightResult1Make(&weight, stopwatch.Elapsed(), solution2.Status)
			} else {
				return weight_types.WeightResult1MakeError(stopwatch.Elapsed(), err2)
			}
		} else {
			return weight_types.WeightResult1MakeError(stopwatch.Elapsed(), fmt.Errorf("no solution"))
		}
	})
}

func (ranker *RankingStatWeightProcess3) RunUsingExternalStart(initialWeight weight_types.Weight1Basic) (*util_async.FutureCancellable[weight_types.WeightResult1], error) {
	ranker.dataSample = ranker.dataAllOriginal
	ranker.prepare()
	ranker.setupInitialSolutionFromExternal2(initialWeight)

	stopwatch := util.StopwatchMakeStopped()
	solutionFuture := ranker.build.RunHighsFuture(stopwatch)
	return util_async.FutureCancellable_MapValue(solutionFuture, func(linearResult util_highs.LinearResult) weight_types.WeightResult1 {
		if solution, err := linearResult.GetSolutionAndSaveLog(ranker.printer); err == nil {
			weight := ranker.extractAndReportSolution(solution)
			return weight_types.WeightResult1Make(&weight, stopwatch.Elapsed(), solution.Status)
		} else {
			return weight_types.WeightResult1MakeError(stopwatch.Elapsed(), err)
		}
	})
}

func (ranker *RankingStatWeightProcess3) prepare() {
	ranker.makeBuilder()
	ranker.prepareRankings()
	ranker.createWeightColumns()
	ranker.doAlgos()
}

func (ranker *RankingStatWeightProcess3) createWeightColumns() {
	lo := -c_Rank3_LargeWeight
	hi := c_Rank3_LargeWeight

	sumWeights := util_highs.ConstraintRow{Debug: "sumWeights"}
	for _, statType := range ranker.requiredStats {
		colWeight := ranker.build.CreateColumnGeneral(highs.Continuous, lo, hi, util_highs.DebugString{Text: "WEIGHT " + statType.Name()})
		ranker.weightColumns.Put(statType, colWeight)
		sumWeights.Add(colWeight, 1)
	}

	sumWeights.Build(ranker.build, c_rank3_target_total_weight, c_rank3_target_total_weight)
}

func (ranker *RankingStatWeightProcess3) prepareRankings() {
	// reset values
	for i := range ranker.dataSample {
		ranker.dataSample[i].SimScore = 0
		ranker.dataSample[i].TargetRank = 0
	}
	simrank.RankingWeightsPrepareBasicRankings(ranker.requiredSims, &ranker.targetRatios, ranker.dataSample)
}

func (ranker *RankingStatWeightProcess3) doAlgos() {
	switch ranker.ALGO {
	case 0:
		ranker.makeDataListEntryColumns()
		for baseIndex := range ranker.dataSample {
			for compareTo := baseIndex + 1; compareTo < len(ranker.dataSample); compareTo++ {
				ranker.makeEntryPairSequenceConstraints(ranker.dataSample[baseIndex], ranker.dataSample[compareTo], baseIndex, compareTo)
			}
		}
	case 1:
		ranker.makeDataListEntryColumns()
		for baseIndex := range ranker.dataSample {
			for compareTo := baseIndex + 1; compareTo < len(ranker.dataSample); compareTo++ {
				ranker.makeEntryPairSequenceConstraintsRequireEqualMode(ranker.dataSample[baseIndex], ranker.dataSample[compareTo], baseIndex, compareTo, 1)
			}
		}
	default:
		panic("dunno")
	}
}

func (ranker *RankingStatWeightProcess3) makeDataListEntryColumns() {
	maxRank := float64(len(ranker.dataSample) - 1)

	sumRanks := util_highs.ConstraintRow{Debug: "sumRanks"}
	for _, entry := range ranker.dataSample {
		debugStr := strconv.FormatInt(int64(entry.TargetRank), 10)
		ranker.makeScoreColumn(entry, debugStr)
		ranker.makeRankColumnAndDiff(entry, maxRank, debugStr)
		sumRanks.Add(entry.RankColumn, 1)
	}

	expectedSum := float64(len(ranker.dataSample)) * float64(len(ranker.dataSample)-1) / 2.0
	sumRanks.Build(ranker.build, expectedSum, expectedSum)
}

func (ranker *RankingStatWeightProcess3) makeDataListEntryColumnsNoMIP() {
	for _, entry := range ranker.dataSample {
		debugStr := strconv.FormatInt(int64(entry.TargetRank), 10)
		ranker.makeScoreColumn(entry, debugStr)
	}
}

func (ranker *RankingStatWeightProcess3) makeRankColumnAndDiff(entry *rankEntry3, maxRank float64, debugStr string) {
	entry.RankColumn = ranker.build.CreateColumnGeneral(highs.Integer, 0, maxRank, util_highs.DebugText("derivedRank-"+debugStr))
	entry.RankDiffAbsColumn = ranker.build.CreateColumnWithOutput(highs.Integer, 0, util_highs.InfPos(), 1, util_highs.DebugText("rankDiffAbs-"+debugStr))

	targetRank := float64(entry.TargetRank)
	ranker.build.AbsoluteValueFromDiffOneToConst(entry.RankColumn, 1, targetRank, entry.RankDiffAbsColumn, "diffRankToTarget")
}

func (ranker *RankingStatWeightProcess3) makeScoreColumn(entry *rankEntry3, debugStr string) {
	entry.ScoreColumn = ranker.build.CreateColumnGeneral(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), util_highs.DebugText("score-"+debugStr))

	scoreRow := util_highs.ConstraintRow{Debug: "scoreRow-" + debugStr}
	for _, statType := range ranker.requiredStats {
		weightColumn := ranker.weightColumns.GetOrPanic(statType)
		statValue := entry.Data.TotalStat.GetFloat(statType)
		statScale := ranker.scaleStats.GetOrPanic(statType)

		scoreRow.Add(weightColumn, statValue*statScale)
	}
	scoreRow.Add(entry.ScoreColumn, -1)
	scoreRow.Build(ranker.build, 0, 0)
}

func (ranker *RankingStatWeightProcess3) makeEntryPairSequenceConstraints(one *rankEntry3, two *rankEntry3, indexOne, indexTwo int) {
	// so we could totally do a boolean thing where scoreA>scoreB then implies rankA>rankB
	// would need all possible pairs connected, but would then force solver to make a full integer order
	isGreaterScore := ranker.build.CreateColumnBool(util_highs.DebugText("isGreaterScore"))
	isGreaterRank := ranker.build.CreateColumnBool(util_highs.DebugText("isGreaterRank"))
	isSequenceDiff := ranker.build.CreateColumnBoolWithOutput(1, util_highs.DebugText("sequenceDiff"))

	ranker.build.ColumnIsGreaterOrEqualColumn(one.ScoreColumn, two.ScoreColumn, isGreaterScore, c_Rank3_LargeScore, 0.0001)
	ranker.build.ColumnIsGreaterOrEqualColumn(one.RankColumn, two.RankColumn, isGreaterRank, c_Rank3_LargeRank, 1.0)
	ranker.build.IsXor(isGreaterRank, isGreaterScore, isSequenceDiff)

	ranker.pairLinks.Put(indexOne, indexTwo, rankPair3{
		indexOne:       indexOne,
		indexTwo:       indexTwo,
		isGreaterScore: isGreaterScore,
		isGreaterRank:  isGreaterRank,
		isSequenceDiff: isSequenceDiff,
	})
}

func (ranker *RankingStatWeightProcess3) makeEntryPairSequenceConstraintsRequireEqualMode(one *rankEntry3, two *rankEntry3, indexOne, indexTwo int, scaleDiffOutput float64) {
	// so we could totally do a boolean thing where scoreA>scoreB then implies rankA>rankB
	// would need all possible pairs connected, but would then force solver to make a full integer order
	isGreaterScore := ranker.build.CreateColumnBool(util_highs.DebugText("isGreaterScore"))
	isGreaterRank := ranker.build.CreateColumnBool(util_highs.DebugText("isGreaterRank"))
	isSequenceDiff := ranker.build.CreateColumnBoolWithOutput(scaleDiffOutput, util_highs.DebugText("sequenceDiff"))

	ranker.build.ColumnIsGreaterOrEqualColumn(one.ScoreColumn, two.ScoreColumn, isGreaterScore, c_Rank3_LargeScore, 0.0001)
	ranker.build.ColumnIsGreaterOrEqualColumn(one.RankColumn, two.RankColumn, isGreaterRank, c_Rank3_LargeRank, 1.0)

	rowEqual := util_highs.ConstraintRow{Debug: "rowEqual"}
	rowEqual.Add(isGreaterScore, 1)
	rowEqual.Add(isGreaterRank, -1)
	rowEqual.Build(ranker.build, 0, 0)

	ranker.pairLinks.Put(indexOne, indexTwo, rankPair3{
		indexOne:       indexOne,
		indexTwo:       indexTwo,
		isGreaterScore: isGreaterScore,
		isGreaterRank:  isGreaterRank,
		isSequenceDiff: isSequenceDiff,
	})
}

func (ranker *RankingStatWeightProcess3) setupDumbInitialSolution() {
	internalWeights := weight_types.Weight1Basic_Make()
	for statType := range ranker.weightColumns.SeqKey() {
		internalWeights.Put(statType, 1)
	}
	internalWeights = internalWeights.ScaleForTotalSum(c_rank3_target_total_weight)

	ranker.setupFromInternalWeights(internalWeights)
}

func (ranker *RankingStatWeightProcess3) setupInitialSolutionFromPreviousWeightOnly(solution *highs.Solution) {
	internalWeights := weight_types.Weight1Basic_Make()
	for statType, colWeight := range ranker.weightColumns.SeqKeyValue() {
		weight := solution.ColValues[colWeight]
		internalWeights.Put(statType, weight)
	}

	ranker.setupFromInternalWeights(internalWeights)
}

func (ranker *RankingStatWeightProcess3) setupInitialSolutionFromExternal2(weights weight_types.Weight1Basic) {
	internalWeights := weights.ScaleForTotalSum(c_rank3_target_total_weight)
	ranker.setupFromInternalWeights(internalWeights)
}

func (ranker *RankingStatWeightProcess3) setupFromInternalWeights(internalWeights weight_types.Weight1Basic) {
	if !internalWeights.IsEmpty() {
		for statType, colWeight := range ranker.weightColumns.SeqKeyValue() {
			weight := internalWeights.Get(statType)
			ranker.build.SetInitialSolutionValue(colWeight, weight)
		}

		ranker.setupInitialRemainingVariables(internalWeights)
	}

	ranker.build.ValidateInitialSolutionState()
}

func (ranker *RankingStatWeightProcess3) setupInitialRemainingVariables(internalWeights weight_types.Weight1Basic) {
	for _, entry := range ranker.dataSample {
		entry.InitialStatScore = internalWeights.CalcStatScoreScaled(entry.Data, ranker.scaleStats)
	}

	for entryPointer, calcRank := range util_collection.CalculateRanking(true, ranker.dataSample, func(x **rankEntry3) float64 { return (*x).InitialStatScore }) {
		entry := *entryPointer
		ranker.build.SetInitialSolutionValue(entry.ScoreColumn, entry.InitialStatScore)
		if entry.RankColumn != -1 {
			ranker.build.SetInitialSolutionValue(entry.RankColumn, float64(calcRank))
		}
		if entry.RankDiffAbsColumn != -1 {
			diff := float64(calcRank) - float64(entry.TargetRank)
			ranker.build.SetInitialSolutionValue(entry.RankDiffAbsColumn, math.Abs(diff))
		}
	}
}

//func (ranker *RankingStatWeightProcess3) setupInitialPairsDetail() {
//	for pair := range ranker.pairLinks.SeqValues() {
//		one, two := &ranker.dataSample[pair.indexOne], &ranker.dataSample[pair.indexTwo]
//		scoreOne, scoreTwo := ranker.build.GetInitialSolutionValue(one.scoreColumn), ranker.build.GetInitialSolutionValue(two.scoreColumn)
//		rankOne, rankTwo := ranker.build.GetInitialSolutionValue(one.rankColumn), ranker.build.GetInitialSolutionValue(two.rankColumn)
//		if scoreTwo >= scoreOne {
//			ranker.build.SetInitialSolutionValue(pair.isGreaterScore, 1)
//			if rankTwo >= rankOne {
//				ranker.build.SetInitialSolutionValue(pair.isGreaterRank, 1)
//				ranker.build.SetInitialSolutionValue(pair.isSequenceDiff, 0)
//			} else {
//				ranker.build.SetInitialSolutionValue(pair.isGreaterRank, 0)
//				ranker.build.SetInitialSolutionValue(pair.isSequenceDiff, 1)
//			}
//		} else {
//			ranker.build.SetInitialSolutionValue(pair.isGreaterScore, 0)
//			if rankTwo >= rankOne {
//				ranker.build.SetInitialSolutionValue(pair.isGreaterRank, 1)
//				ranker.build.SetInitialSolutionValue(pair.isSequenceDiff, 1)
//			} else {
//				ranker.build.SetInitialSolutionValue(pair.isGreaterRank, 0)
//				ranker.build.SetInitialSolutionValue(pair.isSequenceDiff, 0)
//			}
//		}
//	}
//}

func (ranker *RankingStatWeightProcess3) extractAndReportSolution(solution *highs.Solution) weight_types.Weight1Basic {
	ranker.build.DebugPrintColumns(solution, ranker.printer)

	ranker.printer.Println("WEIGHTS")

	weight := weight_types.Weight1Basic_Make()
	for _, statType := range ranker.requiredStats {
		weightColumn := ranker.weightColumns.GetOrPanic(statType)
		statScale := ranker.scaleStats.GetOrPanic(statType)

		modelWeight := solution.ColValues[weightColumn]
		usableWeight := modelWeight / statScale

		weight.Put(statType, usableWeight)

	}

	weight.NormalizeForBase(ranker.requiredStats)
	tools.WriteWeightString(&weight, ranker.printer)

	ranker.reportRankingOfInputs(weight, solution)

	return weight
}

func (ranker *RankingStatWeightProcess3) reportRankingOfInputs(weight weight_types.Weight1Basic, solution *highs.Solution) {
	if ranker.ALGO != 0 {
		return
	}

	ranker.printer.Println("INPUT CHECK")

	type entryCheck struct {
		simScore   float64
		targetRank int
		calcScore  float64
		colScore   float64
		colRank    float64
	}

	check := util_collection.MapSliceAsNew(ranker.dataSample, func(x **rankEntry3) entryCheck {
		e := *x
		return entryCheck{
			e.SimScore,
			e.TargetRank,
			weight.CalcStatScore(&e.Data.TotalStat),
			solution.ColValues[e.ScoreColumn],
			solution.ColValues[e.RankColumn],
		}
	})

	slices.SortFunc(check, func(a, b entryCheck) int { return cmp.Compare(a.targetRank, b.targetRank) })
	for i, entry := range check {
		ranker.printer.Printf("%4d %8f %4d %8.1f %12.10f %4.1f\n", i,
			entry.simScore, entry.targetRank,
			entry.calcScore, entry.colScore, entry.colRank)
	}
	ranker.printer.Println0()
	ranker.printer.Println0()

	slices.SortFunc(check, func(a, b entryCheck) int { return cmp.Compare(a.calcScore, b.calcScore) })
	for i, entry := range check {
		ranker.printer.Printf("%4d %8f %4d %8.1f %12.10f %4.1f\n", i,
			entry.simScore, entry.targetRank,
			entry.calcScore, entry.colScore, entry.colRank)
	}
	ranker.printer.Println0()
	ranker.printer.Println0()

	slices.SortFunc(check, func(a, b entryCheck) int { return cmp.Compare(a.colScore, b.colScore) })
	for i, entry := range check {
		ranker.printer.Printf("%4d %8f %4d %8.1f %12.10f %4.1f\n", i,
			entry.simScore, entry.targetRank,
			entry.calcScore, entry.colScore, entry.colRank)
	}
	ranker.printer.Println0()
	ranker.printer.Println0()
}
