package weight_highs

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/simrank"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_Rank4ScaleTarget    = 1.0
	c_Rank4LargeWeight    = 50.0
	c_Rank4WeightTotalSum = 10.0
	c_Rank4LimitScore     = 200.0
	c_Rank4LargeScore     = 500.0
)

type RankingStatWeightProcess4 struct {
	printer  *util.PrintRecorder
	MULTIPLY int

	targetRatios  weight_types.SimPriorityBasic
	requiredStats []stats.StatType
	requiredSims  []stats.SimType
	dataAll       []weight_types.WeightInput
}

type rankInternalRun4 struct {
	process *RankingStatWeightProcess4

	build *util_highs.LinearBuilder

	runData    []*rankEntry4
	scaleStats map[stats.StatType]float64

	weightColumns map[stats.StatType]util_highs.ColumnIndex
	pairLinks     util_collection.MapMapDiagonal[int, *rankPair4]
}

type rankEntry4 struct {
	weight_types.RankStatWeightsCommon

	InitialStatScore float64

	ScoreColumn util_highs.ColumnIndex
	RankColumn  util_highs.ColumnIndex
}

type rankPair4 struct {
	entryOne     *rankEntry4
	oneIsGreater util_highs.ColumnIndex

	entryTwo *rankEntry4
	// twoIsGreater utilhighs.ColumnIndex
}

func (process *RankingStatWeightProcess4) Init(printer *util.PrintRecorder) {
	process.printer = printer
}

func (process *RankingStatWeightProcess4) SupplyData(inputData []weight_types.WeightInput) {
	process.dataAll = inputData
}

func (run *rankInternalRun4) supplyData(inputData []weight_types.WeightInput) {
	run.scaleStats = chooseStatScalingBasic(inputData, c_Rank4ScaleTarget, false, run.process.printer)
	run.runData = util_collection.MapSliceAsNew(inputData, func(input *weight_types.WeightInput) *rankEntry4 {
		return &rankEntry4{
			RankStatWeightsCommon: weight_types.RankStatWeightsCommon{
				Data:       input,
				SimScore:   0,
				TargetRank: 0,
			},
			ScoreColumn: -1,
			RankColumn:  -1,
		}
	})
}

func (process *RankingStatWeightProcess4) SetRequiredStats(requiredStats []stats.StatType) {
	process.requiredStats = requiredStats
}

func (process *RankingStatWeightProcess4) SetTargetRatios(targetRatios weight_types.SimPriorityBasic) {
	process.targetRatios = targetRatios
	process.requiredSims = targetRatios.SimTypes()
}

func (process *RankingStatWeightProcess4) RunUsingExternalStart(initialWeight weight_types.Weight1Basic, stopwatch *util.Stopwatch, timeout int) util_collection.Optional[weight_types.Weight1Basic] {
	run2 := rankInternalRun4_create(process)
	run2.build.TimeLimitSeconds = timeout
	run2.supplyData(process.dataAll)
	run2.prepareRankings()
	run2.createWeightColumns()
	run2.makeDataListEntryColumns()
	run2.makeDataListPairRules()
	run2.setupInitialSolutionFromExternal2(initialWeight)
	weights2, _ := run2.run(stopwatch)
	return weights2
}

func takeDataSample_Start[T any](slice []T, size int) []T {
	if len(slice) < size {
		return slice
	} else {
		return slice[0:size]
	}
}

func takeDataSample_Random[T any](slice []T, size int) []T {
	if len(slice) < size {
		return slice
	} else {
		copy := slices.Clone(slice)
		util_collection.Shuffle(copy)
		return copy[0:size]
	}
}

func rankInternalRun4_create(process *RankingStatWeightProcess4) *rankInternalRun4 {
	run := new(rankInternalRun4)
	run.process = process
	run.build = new(util_highs.LinearBuilder)
	run.build.Minimise = true
	run.build.Solver = util_highs.Solver_MIP_Interior
	return run
}

func (run *rankInternalRun4) run(stopwatch *util.Stopwatch) (util_collection.Optional[weight_types.Weight1Basic], *highs.Solution) {
	solutionFuture := run.build.RunHighsFuture(stopwatch)
	linearResult := solutionFuture.WaitForResultOrPanic()
	solution := linearResult.GetSolutionAndSaveLog(run.process.printer)
	if solution.HasSolution() {
		weights := run.extractAndReportSolution(solution)
		return util_collection.Optional_OfValue(weights), solution
	} else {
		return util_collection.Optional_Empty[weight_types.Weight1Basic](), solution
	}
}

func (run *rankInternalRun4) runFuture(stopwatch *util.Stopwatch) *util_async.FutureCancellable[weight_types.Weight1Basic] {
	solutionFuture := run.build.RunHighsFuture(stopwatch)
	return util_async.FutureCancellable_MapValue(solutionFuture, func(linearResult util_highs.LinearResult) (weight_types.Weight1Basic, bool) {
		solution := linearResult.GetSolutionAndSaveLog(run.process.printer)
		return run.extractAndReportSolution(solution), true
	})
}

func (run *rankInternalRun4) createWeightColumns() {
	lo := -c_Rank4LargeWeight
	hi := c_Rank4LargeWeight
	strengthMin := 0.01

	sumWeights := util_highs.ConstraintRow{Debug: "sumWeights"}
	run.weightColumns = make(map[stats.StatType]util_highs.ColumnIndex)
	for _, statType := range run.process.requiredStats {
		var colWeight util_highs.ColumnIndex
		if statType == stats.Stat_Strength {
			colWeight = run.build.CreateColumnGeneral(highs.Continuous, strengthMin, hi, util_highs.DebugString{Text: "WEIGHT " + statType.Name()})
		} else {
			colWeight = run.build.CreateColumnGeneral(highs.Continuous, lo, hi, util_highs.DebugString{Text: "WEIGHT " + statType.Name()})
		}
		run.weightColumns[statType] = colWeight
		sumWeights.Add(colWeight, 1)
	}

	sumWeights.Build(run.build, c_Rank4WeightTotalSum, c_Rank4WeightTotalSum)
}

func (run *rankInternalRun4) prepareRankings() {
	simrank.RankingWeightsPrepareUsingMidRange(run.process.requiredSims, &run.process.targetRatios, run.runData)
}

func (run *rankInternalRun4) makeDataListEntryColumns() {
	maxRank := float64(len(run.runData) - 1)

	primes := util.PrimesSmall(int64(len(run.runData)))
	sumRanksByPrime := util_highs.ConstraintRow{Debug: "sumRanksByPrime"}
	targetSum := 0.0

	for i := range run.runData {
		entry := run.runData[i]
		primeMultiplier := float64(primes[i])

		run.makeEntryColumnRefs(entry, maxRank)

		sumRanksByPrime.Add(entry.RankColumn, primeMultiplier)
		targetSum += float64(entry.TargetRank) * primeMultiplier
	}
}

func (run *rankInternalRun4) makeEntryColumnRefs(entry *rankEntry4, maxRank float64) {
	rankStr := strconv.FormatInt(int64(entry.TargetRank), 10)
	entry.ScoreColumn = run.build.CreateColumnGeneral(highs.Continuous, -c_Rank4LimitScore, c_Rank4LimitScore, util_highs.DebugText("score-"+rankStr))

	scoreRow := util_highs.ConstraintRow{Debug: "scoreRow"}
	for _, statType := range run.process.requiredStats {
		weightColumn := run.weightColumns[statType]
		statValue := entry.Data.TotalStat.GetFloat(statType)
		statScale := run.scaleStats[statType]

		scoreRow.Add(weightColumn, statValue*statScale)
	}
	scoreRow.Add(entry.ScoreColumn, -1)
	scoreRow.Build(run.build, 0, 0)

	entry.RankColumn = run.build.CreateColumnGeneral(highs.Integer, 0, maxRank, util_highs.DebugText("derivedRank-"+rankStr))

	rankDiff := run.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.C_PlusInf, 1, util_highs.DebugText("rankDiff"))
	run.build.AbsoluteValueFromDiffOneToConst(entry.RankColumn, 1, float64(entry.TargetRank), rankDiff, "rankDiff")
}

// what if i have all pairs of possible greater thans
// rank is the sum of those
func (run *rankInternalRun4) makeDataListPairRules() {
	for a := 0; a < len(run.runData); a++ {
		for b := a + 1; b < len(run.runData); b++ {
			run.makeEntryPairScoreChecks(run.runData[a], run.runData[b], a, b)
		}
	}

	if run.pairLinks.Size() != len(run.runData)*(len(run.runData)-1)/2 {
		panic("expected exact size")
	}

	for i := 0; i < len(run.runData); i++ {
		run.makeRankDerivation(i, run.runData[i])
	}
}

func (run *rankInternalRun4) makeEntryPairScoreChecks(one *rankEntry4, two *rankEntry4, indexOne, indexTwo int) {
	oneIsGreater := run.build.CreateColumnBool(util_highs.DebugText("oneIsGreater"))

	oneGreaterRow := util_highs.ConstraintRow{Debug: "oneGreaterRow"}
	oneGreaterRow.Add(two.ScoreColumn, -1)
	oneGreaterRow.Add(one.ScoreColumn, 1)
	oneGreaterRow.Add(oneIsGreater, -c_Rank4LargeScore)
	oneGreaterRow.Build(run.build, -c_Rank4LargeScore, 0)

	run.pairLinks.Put(indexOne, indexTwo, &rankPair4{
		entryOne:     one,
		entryTwo:     two,
		oneIsGreater: oneIsGreater,
	})
}

func (run *rankInternalRun4) makeRankDerivation(mainIndex int, mainEntry *rankEntry4) {
	sumCompareFlags := util_highs.ConstraintRow{Debug: "rankDerive " + strconv.FormatInt(int64(mainIndex), 10)}

	// x+y = 1 (NOT logic)
	// y = 1-x
	// formula                a + b + c + d - rank = 0
	// substitute a NOT:  a + b + (1-c) + d - rank = 0
	//                    a + b + 1 - c + d - rank = 0
	//                        a + b - c + d - rank = -1
	rowSumTarget := 0.0
	for _, pair := range run.pairLinks.SeqInnerWithKeyValue(mainIndex) {
		if pair.entryOne == mainEntry {
			sumCompareFlags.Add(pair.oneIsGreater, 1)
		} else {
			// sumCompareFlags.Add(pair.twoIsGreater, 1)
			sumCompareFlags.Add(pair.oneIsGreater, -1)
			rowSumTarget--
		}
	}

	sumCompareFlags.Add(mainEntry.RankColumn, -1)
	sumCompareFlags.Build(run.build, rowSumTarget, rowSumTarget)
}

func (run *rankInternalRun4) extractAndReportSolution(solution *highs.Solution) weight_types.Weight1Basic {
	run.build.DebugPrintColumns(solution, run.process.printer)

	run.process.printer.Println("WEIGHTS")

	statWeightResult := weight_types.Weight1Basic_Make(run.process.targetRatios)
	for _, statType := range run.process.requiredStats {
		weightColumn := run.weightColumns[statType]
		statScale := run.scaleStats[statType]

		modelWeight := solution.ColValues[weightColumn]
		usableWeight := modelWeight
		if run.process.MULTIPLY == 1 {
			usableWeight = modelWeight * statScale
		} else if run.process.MULTIPLY == 2 {
			usableWeight = modelWeight / statScale
		}

		statWeightResult.Put(statType, usableWeight)
	}

	baseStat := run.process.requiredStats[0]
	divideBy := statWeightResult.Get(baseStat)
	for _, statType := range run.process.requiredStats {
		value := statWeightResult.Get(statType) / divideBy
		statWeightResult.Put(statType, value)
		run.process.printer.Printf("%10s %f\n", statType.Name(), value)
	}

	return statWeightResult
}

func (run *rankInternalRun4) setupInitialSolutionFromExternal2(weights weight_types.Weight1Basic) {
	internalWeights := weights.ScaleForTotalSum(c_Rank4WeightTotalSum)
	for statType, colWeight := range run.weightColumns {
		basicValue := internalWeights.Get(statType)
		run.build.SetInitialSolutionValue(colWeight, basicValue)
	}

	for _, entry := range run.runData {
		entry.InitialStatScore = internalWeights.CalcStatScoreScaled(entry.Data, run.scaleStats)
		run.build.SetInitialSolutionValue(entry.ScoreColumn, entry.InitialStatScore)
	}

	run.setupRemainingInitialSolution()
	run.build.ValidateInitialSolutionState()
}

func (run *rankInternalRun4) setupRemainingInitialSolution() {
	// for entry, rankHiLo := range util.CalculateRankingRanges(true, run.runData, func(x *rankEntry4) float64 { return x.initialStatScore }) {
	// run.input.SetInitialSolutionValue(entry.rankColumn, float64(rankHiLo.Lo))
	// }

	for pair := range run.pairLinks.SeqValues() {
		entryOne := pair.entryOne
		entryTwo := pair.entryTwo
		if entryOne.InitialStatScore > entryTwo.InitialStatScore {
			run.build.SetInitialSolutionValue(pair.oneIsGreater, 1)
			// run.input.SetInitialSolutionValue(pair.twoIsGreater, 0)
		} else if entryOne.InitialStatScore < entryTwo.InitialStatScore {
			run.build.SetInitialSolutionValue(pair.oneIsGreater, 0)
			// run.input.SetInitialSolutionValue(pair.twoIsGreater, 1)
		} else {
			// run.input.SetInitialSolutionValue(pair.oneIsGreater, 0)
			// run.input.SetInitialSolutionValue(pair.twoIsGreater, 0)
		}
	}
}
