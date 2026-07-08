package stathighs

import (
	"cmp"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
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
	printer   *util.PrintRecorder
	WEIGHTSUM int

	targetRatios  stats.SimData
	requiredStats []stats.StatType
	requiredSims  []stats.SimType
	dataAll       []WeightInput
}

type rankInternalRun4 struct {
	process *RankingStatWeightProcess4

	build *utilhighs.LinearBuilder

	runData    []rankEntry4
	scaleStats map[stats.StatType]float64

	weightColumns map[stats.StatType]utilhighs.ColumnIndex
	pairLinks     util.MapMapDiagonal[int, *rankPair4]
}

type rankEntry4 struct {
	data *WeightInput

	initialStatScore float64
	simScore         float64
	targetRank       int

	scoreColumn utilhighs.ColumnIndex
	rankColumn  utilhighs.ColumnIndex
}

type rankPair4 struct {
	entryOne     *rankEntry4
	oneIsGreater utilhighs.ColumnIndex

	entryTwo *rankEntry4
	// twoIsGreater utilhighs.ColumnIndex
}

func (process *RankingStatWeightProcess4) Init(printer *util.PrintRecorder) {
	process.printer = printer
}

func (process *RankingStatWeightProcess4) SupplyData(inputData []WeightInput) {
	process.dataAll = inputData
}

func (run *rankInternalRun4) supplyData(inputData []WeightInput) {
	run.scaleStats = chooseStatScaling(inputData, c_Rank4ScaleTarget, run.process.printer)
	run.runData = util.MapSliceAsNew(inputData, func(input *WeightInput) rankEntry4 {
		return rankEntry4{
			data:        input,
			simScore:    -1,
			targetRank:  -1,
			scoreColumn: -1,
			rankColumn:  -1,
		}
	})
}

func (process *RankingStatWeightProcess4) SetRequiredStats(requiredStats []stats.StatType) {
	process.requiredStats = requiredStats
}

func (process *RankingStatWeightProcess4) SetTargetRatios(targetRatios stats.SimData) {
	process.targetRatios = targetRatios
	process.requiredSims = targetRatios.NonZeroTypes()
}

//func (process *RankingStatWeightProcess4) Run(overallStopwatch *util.Stopwatch, timeout int) []WeightResult {
//	weightResultList := make([]WeightResult, 0)
//
//	// FIRST ROUND: minimal data, dumb initial values
//	process.printer.Println("RankingStatWeightProcess4 FIRST ROUND")
//	run1 := rankInternalRun4_create(process)
//	run1.build.TimeLimitSeconds = timeout / 4
//	run1.supplyData(takeDataSample_Random(process.dataAll, c_Rank4InitialSample))
//	run1.prepareRankings()
//	run1.createWeightColumns()
//	run1.makeDataListEntryColumns()
//	run1.makeDataListPairRules()
//	run1.setupInitialSolutionDumb2()
//	weights1, solution1 := run1.run(overallStopwatch)
//	weights1.ApplyIfValue(func(w WeightResult) { weightResultList = append(weightResultList, w) })
//
//	// noinitial, sample25, accuracy = 85.238976, Duration = 5m40.3326511s
//	// dumb2, sample25, accuracy = 85.227590, Duration = 4m0.7662002s
//	// dumb1, sample25, accuracy = 85.225498 Duration = 4m32.3105105s
//
//	latestRun := run1
//	latestSolution := solution1
//
//	times := make(map[int]time.Duration)
//	for size := c_Rank4InitialSample + c_Rank4AddSample; size <= len(process.dataAll); size += c_Rank4AddSample {
//		roundTimer := util.StopwatchMakeStopped()
//
//		dataSample := takeDataSample_Random(process.dataAll, size)
//		process.printer.Println("RankingStatWeightProcess4 SECOND ROUND " + strconv.Itoa(size))
//		run2 := rankInternalRun4_create(process)
//		run2.build.TimeLimitSeconds = timeout / 4
//		run2.supplyData(dataSample)
//		run2.prepareRankings()
//		run2.createWeightColumns()
//		run2.makeDataListEntryColumns()
//		run2.makeDataListPairRules()
//		run2.setupInitialSolutionFromPrevious(latestRun, latestSolution)
//		weights2, solution2 := run2.run(roundTimer)
//		weights2.ApplyIfValue(func(w WeightResult) { weightResultList = append(weightResultList, w) })
//
//		times[size] = roundTimer.Elapsed()
//		timeout -= int(roundTimer.Elapsed())
//		overallStopwatch.AddElapsedFrom(roundTimer)
//
//		if solution2.HasSolution() {
//			latestRun = run2
//			latestSolution = solution2
//		}
//
//		if solution2.Status == highs.ModelStatusTimeLimit || timeout < 0 {
//			break
//		}
//	}
//
//	for size, duration := range times {
//		process.printer.Printf("%4d %s\n", size, duration)
//	}
//
//	return weightResultList
//}

func (process *RankingStatWeightProcess4) RunUsingExternalStart(initialWeight WeightResult, stopwatch *util.Stopwatch, timeout int) util.Optional[WeightResult] {
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
		util.Shuffle(copy)
		return copy[0:size]
	}
}

func rankInternalRun4_create(process *RankingStatWeightProcess4) *rankInternalRun4 {
	run := new(rankInternalRun4)
	run.process = process
	run.build = new(utilhighs.LinearBuilder)
	run.build.Minimise = true
	run.build.Solver = utilhighs.Solver_MIP_Interior
	return run
}

func (run *rankInternalRun4) run(stopwatch *util.Stopwatch) (util.Optional[WeightResult], *highs.Solution) {
	solutionFuture := run.build.RunHighsFuture(stopwatch)
	linearResult := solutionFuture.WaitForResultOrPanic()
	solution := linearResult.GetSolutionAndSaveLog(run.process.printer)
	if solution.HasSolution() {
		weights := run.extractAndReportSolution(solution)
		return util.Optional_OfValue(weights), solution
	} else {
		return util.Optional_Empty[WeightResult](), solution
	}
}

func (run *rankInternalRun4) runFuture(stopwatch *util.Stopwatch) *channel_op.FutureCancellable[WeightResult] {
	solutionFuture := run.build.RunHighsFuture(stopwatch)
	return channel_op.FutureCancellable_MapValue(solutionFuture, func(linearResult utilhighs.LinearResult) (WeightResult, bool) {
		solution := linearResult.GetSolutionAndSaveLog(run.process.printer)
		return run.extractAndReportSolution(solution), true
	})
}

func (run *rankInternalRun4) createWeightColumns() {
	lo := -c_Rank4LargeWeight
	hi := c_Rank4LargeWeight
	strengthMin := 0.01

	sumWeights := utilhighs.ConstraintRow{Debug: "sumWeights"}
	run.weightColumns = make(map[stats.StatType]utilhighs.ColumnIndex)
	for _, statType := range run.process.requiredStats {
		var colWeight utilhighs.ColumnIndex
		if statType == stats.Stat_Strength {
			colWeight = run.build.CreateColumnGeneral(highs.Continuous, strengthMin, hi, utilhighs.DebugString{Text: "WEIGHT " + statType.Name()})
		} else {
			colWeight = run.build.CreateColumnGeneral(highs.Continuous, lo, hi, utilhighs.DebugString{Text: "WEIGHT " + statType.Name()})
		}
		run.weightColumns[statType] = colWeight
		sumWeights.Add(colWeight, 1)
	}

	if run.process.WEIGHTSUM == 0 {
		sumWeights.Build(run.build, 1.0, hi) // force positive and non-zero result
	} else {
		sumWeights.Build(run.build, c_Rank4WeightTotalSum, c_Rank4WeightTotalSum)
	}
}

func (run *rankInternalRun4) prepareRankings() {
	// reset values
	for i := range run.runData {
		run.runData[i].simScore = 0
		run.runData[i].targetRank = 0
	}

	// score each sim
	for _, simType := range run.process.requiredSims {
		for entry, simDetailRankHiLo := range util.CalculateRankingRanges(simType.IsHighGood(), run.runData, func(x *rankEntry4) float64 { return x.data.SimResult.Get(simType) }) {
			entry.simScore += float64(simDetailRankHiLo.Mid()) * run.process.targetRatios.Get(simType)
		}
	}

	// rank combined sims
	for entry, simRankHiLo := range util.CalculateRankingRanges(true, run.runData, func(x *rankEntry4) float64 { return x.simScore }) {
		entry.targetRank = simRankHiLo.Lo
	}

	slices.SortFunc(run.runData, func(a, b rankEntry4) int { return cmp.Compare(a.targetRank, b.targetRank) })
}

func (run *rankInternalRun4) makeDataListEntryColumns() {
	maxRank := float64(len(run.runData) - 1)

	primes := util.PrimesSmall(int64(len(run.runData)))
	sumRanksByPrime := utilhighs.ConstraintRow{Debug: "sumRanksByPrime"}
	targetSum := 0.0

	for i := range run.runData {
		entry := &run.runData[i]
		primeMultiplier := float64(primes[i])

		run.makeEntryColumnRefs(entry, maxRank)

		sumRanksByPrime.Add(entry.rankColumn, primeMultiplier)
		targetSum += float64(entry.targetRank) * primeMultiplier
	}
}

func (run *rankInternalRun4) makeEntryColumnRefs(entry *rankEntry4, maxRank float64) {
	rankStr := strconv.FormatInt(int64(entry.targetRank), 10)
	entry.scoreColumn = run.build.CreateColumnGeneral(highs.Continuous, -c_Rank4LimitScore, c_Rank4LimitScore, utilhighs.DebugText("score-"+rankStr))

	scoreRow := utilhighs.ConstraintRow{Debug: "scoreRow"}
	for _, statType := range run.process.requiredStats {
		weightColumn := run.weightColumns[statType]
		statValue := entry.data.TotalStat.GetFloat(statType)
		statScale := run.scaleStats[statType]

		scoreRow.Add(weightColumn, statValue*statScale)
	}
	scoreRow.Add(entry.scoreColumn, -1)
	scoreRow.Build(run.build, 0, 0)

	entry.rankColumn = run.build.CreateColumnGeneral(highs.Integer, 0, maxRank, utilhighs.DebugText("derivedRank-"+rankStr))

	rankDiff := run.build.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugText("rankDiff"))
	run.build.AbsoluteValueFromDiffOneToConst(entry.rankColumn, 1, float64(entry.targetRank), rankDiff, "rankDiff")
}

// what if i have all pairs of possible greater thans
// rank is the sum of those
func (run *rankInternalRun4) makeDataListPairRules() {
	for a := 0; a < len(run.runData); a++ {
		for b := a + 1; b < len(run.runData); b++ {
			run.makeEntryPairScoreChecks(&run.runData[a], &run.runData[b], a, b)
		}
	}

	if run.pairLinks.Size() != len(run.runData)*(len(run.runData)-1)/2 {
		panic("expected exact size")
	}

	for i := 0; i < len(run.runData); i++ {
		run.makeRankDerivation(i, &run.runData[i])
	}
}

func (run *rankInternalRun4) makeEntryPairScoreChecks(one *rankEntry4, two *rankEntry4, indexOne, indexTwo int) {
	oneIsGreater := run.build.CreateColumnBool(utilhighs.DebugText("oneIsGreater"))

	oneGreaterRow := utilhighs.ConstraintRow{Debug: "oneGreaterRow"}
	oneGreaterRow.Add(two.scoreColumn, -1)
	oneGreaterRow.Add(one.scoreColumn, 1)
	oneGreaterRow.Add(oneIsGreater, -c_Rank4LargeScore)
	oneGreaterRow.Build(run.build, -c_Rank4LargeScore, 0)

	run.pairLinks.Put(indexOne, indexTwo, &rankPair4{
		entryOne:     one,
		entryTwo:     two,
		oneIsGreater: oneIsGreater,
	})
}

func (run *rankInternalRun4) makeRankDerivation(mainIndex int, mainEntry *rankEntry4) {
	sumCompareFlags := utilhighs.ConstraintRow{Debug: "rankDerive " + strconv.FormatInt(int64(mainIndex), 10)}

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

	sumCompareFlags.Add(mainEntry.rankColumn, -1)
	sumCompareFlags.Build(run.build, rowSumTarget, rowSumTarget)
}

func (run *rankInternalRun4) extractAndReportSolution(solution *highs.Solution) WeightResult {
	run.build.DebugPrintColumns(solution, run.process.printer)

	run.process.printer.Println("WEIGHTS")

	statWeightResult := WeightResult_Make()
	for _, statType := range run.process.requiredStats {
		weightColumn := run.weightColumns[statType]
		statScale := run.scaleStats[statType]

		modelWeight := solution.ColValues[weightColumn]
		// TODO changed to multiply following analysis on other algorithms, not checked here
		usableWeight := modelWeight * statScale

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

func (run *rankInternalRun4) setupInitialSolutionDumb1() {
	baseStat := run.process.requiredStats[0]
	for statType, colWeight := range run.weightColumns {
		if statType == baseStat {
			run.build.SetInitialSolutionValue(colWeight, 1)
		} else {
			run.build.SetInitialSolutionValue(colWeight, 0)
		}
	}

	statScale := run.scaleStats[baseStat]
	for entry := range util.ForPointer(run.runData) {
		entry.initialStatScore = entry.data.TotalStat.GetFloat(baseStat) * statScale
	}

	run.setupRemainingInitialSolution()
	run.build.ValidateInitialSolutionState()
}

func (run *rankInternalRun4) setupInitialSolutionDumb2() {
	initialDumbValue := 0.2

	dumbWeights := WeightResult_Make()
	for statType, colWeight := range run.weightColumns {
		run.build.SetInitialSolutionValue(colWeight, initialDumbValue)
		dumbWeights.Put(statType, initialDumbValue)
	}

	for entry := range util.ForPointer(run.runData) {
		entry.initialStatScore = dumbWeights.CalcStatScoreScaled(entry.data, run.scaleStats)
	}

	run.setupRemainingInitialSolution()
	run.build.ValidateInitialSolutionState()
}

func (run *rankInternalRun4) setupInitialSolutionFromPrevious(previous *rankInternalRun4, solution *highs.Solution) {
	internalWeights := WeightResult_Make()
	for statType, oldWeightCol := range previous.weightColumns {
		newWeightCol := run.weightColumns[statType]
		oldWeightValue := solution.ColValues[oldWeightCol]
		run.build.SetInitialSolutionValue(newWeightCol, oldWeightValue)
		internalWeights.Put(statType, oldWeightValue)
	}

	for entry := range util.ForPointer(run.runData) {
		entry.initialStatScore = internalWeights.CalcStatScoreScaled(entry.data, run.scaleStats)
	}

	run.setupRemainingInitialSolution()
	run.build.ValidateInitialSolutionState()
}

func (run *rankInternalRun4) setupInitialSolutionFromExternal(weights stats.StatBlock) {
	internalWeights := WeightResult_Make()
	for statType, colWeight := range run.weightColumns {
		basicValue := weights.GetFloat(statType)
		// scale := run.scaleStats[statType]
		scaledValue := basicValue
		run.build.SetInitialSolutionValue(colWeight, scaledValue)
		internalWeights.Put(statType, scaledValue)
	}

	for entry := range util.ForPointer(run.runData) {
		entry.initialStatScore = internalWeights.CalcStatScoreScaled(entry.data, run.scaleStats)
		run.build.SetInitialSolutionValue(entry.scoreColumn, entry.initialStatScore)
	}

	run.setupRemainingInitialSolution()
	run.build.ValidateInitialSolutionState()
}

func (run *rankInternalRun4) setupInitialSolutionFromExternal2(weights WeightResult) {
	internalWeights := WeightResult_Make()
	for statType, colWeight := range run.weightColumns {
		basicValue := weights.Get(statType)
		// scale := run.scaleStats[statType]
		scaledValue := basicValue
		run.build.SetInitialSolutionValue(colWeight, scaledValue)
		internalWeights.Put(statType, scaledValue)
	}

	for entry := range util.ForPointer(run.runData) {
		entry.initialStatScore = internalWeights.CalcStatScoreScaled(entry.data, run.scaleStats)
		run.build.SetInitialSolutionValue(entry.scoreColumn, entry.initialStatScore)
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
		if entryOne.initialStatScore > entryTwo.initialStatScore {
			run.build.SetInitialSolutionValue(pair.oneIsGreater, 1)
			// run.input.SetInitialSolutionValue(pair.twoIsGreater, 0)
		} else if entryOne.initialStatScore < entryTwo.initialStatScore {
			run.build.SetInitialSolutionValue(pair.oneIsGreater, 0)
			// run.input.SetInitialSolutionValue(pair.twoIsGreater, 1)
		} else {
			// run.input.SetInitialSolutionValue(pair.oneIsGreater, 0)
			// run.input.SetInitialSolutionValue(pair.twoIsGreater, 0)
		}
	}
}
