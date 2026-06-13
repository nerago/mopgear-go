package stathighs

import (
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"strconv"
	"time"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	// c_RankLargeWeight = 50.0
	// c_RankLargeScore  = 500.0
	c_RankLimitScore    = 200.0
	c_RankInitialSample = 15
	c_RankAddSample     = 5

// c_RankLargeRank   = 10000.0
)

type RankingStatWeightProcess4 struct {
	printer *util.PrintRecorder

	targetRatios simulate.SimData
	dataAll      []WeightInput
}

type rankInternalRun4 struct {
	process *RankingStatWeightProcess4

	input *utilhighs.InputBuilder

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
	run.scaleStats = chooseStatScaling(inputData, run.process.printer)
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

func (process *RankingStatWeightProcess4) SetTargetRatios(targetRatios simulate.SimData) {
	process.targetRatios = targetRatios
}

func (process *RankingStatWeightProcess4) Run(doRound2 bool) []WeightResult {
	weightResultList := make([]WeightResult, 0)

	// FIRST ROUND: minimal data, dumb initial values
	process.printer.Println("RankingStatWeightProcess4 FIRST ROUND")
	run1 := rankInternalRun4_create(process)
	run1.input.TimeLimitSeconds = 3600
	run1.supplyData(takeDataSample(process.dataAll, c_RankInitialSample))
	run1.prepareRankings()
	run1.createWeightColumns()
	run1.makeDataListEntryColumns()
	run1.makeDataListPairRules()
	run1.setupInitialSolutionDumb2()
	weights1, solution1 := run1.run()
	weights1.ApplyIfValue(func(w WeightResult) { weightResultList = append(weightResultList, w) })

	// noinitial, sample25, accuracy = 85.238976, Duration = 5m40.3326511s
	// dumb2, sample25, accuracy = 85.227590, Duration = 4m0.7662002s
	// dumb1, sample25, accuracy = 85.225498 Duration = 4m32.3105105s

	latestRun := run1
	latestSolution := solution1

	times := make(map[int]time.Duration)
	for size := c_RankInitialSample + c_RankAddSample; size <= len(process.dataAll); size += c_RankAddSample {
		startTime := time.Now()

		dataSample := takeDataSample(process.dataAll, size)
		process.printer.Println("RankingStatWeightProcess4 SECOND ROUND " + strconv.Itoa(size))
		run2 := rankInternalRun4_create(process)
		run2.input.TimeLimitSeconds = 8000
		run2.supplyData(dataSample)
		run2.prepareRankings()
		run2.createWeightColumns()
		run2.makeDataListEntryColumns()
		run2.makeDataListPairRules()
		run2.setupInitialSolutionFromPrevious(latestRun, latestSolution)
		weights2, solution2 := run2.run()
		weights2.ApplyIfValue(func(w WeightResult) { weightResultList = append(weightResultList, w) })

		timeTaken := time.Since(startTime)
		times[size] = timeTaken

		if solution2.HasSolution() {
			latestRun = run2
			latestSolution = solution2
		}

		if solution2.Status == highs.ModelStatusTimeLimit {
			break
		}
	}

	for size, duration := range times {
		process.printer.Printf("%4d %s\n", size, duration)
	}

	return weightResultList
}

func (process *RankingStatWeightProcess4) RunUsingExternalStart(initialWeight WeightResult) util.Optional[WeightResult] {
	run2 := rankInternalRun4_create(process)
	run2.input.TimeLimitSeconds = 2500
	run2.supplyData(process.dataAll)
	run2.prepareRankings()
	run2.createWeightColumns()
	run2.makeDataListEntryColumns()
	run2.makeDataListPairRules()
	run2.setupInitialSolutionFromExternal2(initialWeight)
	weights2, _ := run2.run()
	return weights2
}

func takeDataSample(slice []WeightInput, size int) []WeightInput {
	if len(slice) < size {
		return slice
	} else {
		return slice[0:size]
	}
}

func rankInternalRun4_create(process *RankingStatWeightProcess4) *rankInternalRun4 {
	run := new(rankInternalRun4)
	run.process = process
	run.input = new(utilhighs.InputBuilder)
	run.input.Minimise = true
	// run.input.Mip_disallow_restart = true
	run.input.Mip_lp_solver = "ipx"
	return run
}

func (run *rankInternalRun4) run() (util.Optional[WeightResult], *highs.Solution) {
	solution, log := run.input.RunHighs()
	run.process.printer.AppendOther(log)
	if solution.HasSolution() {
		weights := run.extractAndReportSolution(solution)
		return util.Optional_OfValue(weights), solution
	} else {
		return util.Optional_Empty[WeightResult](), solution
	}
}

func (run *rankInternalRun4) createWeightColumns() {
	lo := -c_RankLargeWeight
	hi := c_RankLargeWeight
	strengthMin := 0.01

	sumWeights := utilhighs.ConstraintRowBuild{Debug: "sumWeights"}
	run.weightColumns = make(map[stats.StatType]utilhighs.ColumnIndex)
	for _, statType := range G_RequiredStats {
		var colWeight utilhighs.ColumnIndex
		if statType == stats.Stat_Strength {
			colWeight = run.input.CreateColumnGeneral(highs.Continuous, strengthMin, hi, utilhighs.DebugString{Text: "WEIGHT " + statType.Name()})
		} else {
			colWeight = run.input.CreateColumnGeneral(highs.Continuous, lo, hi, utilhighs.DebugString{Text: "WEIGHT " + statType.Name()})
		}
		run.weightColumns[statType] = colWeight
		sumWeights.Add(colWeight, 1)
	}

	sumWeights.Finish(run.input, 1.0, hi) // force positive and non-zero result
}

func (run *rankInternalRun4) prepareRankings() {
	// reset values
	for i := range run.runData {
		run.runData[i].simScore = 0
		run.runData[i].targetRank = 0
	}

	// score each sim
	for _, simType := range G_RequiredSims {
		for entry, simDetailRankHiLo := range util.CalculateRankingRanges(simType.IsHighGood(), run.runData, func(x *rankEntry4) float64 { return x.data.SimResult.Get(simType) }) {
			entry.simScore += float64(simDetailRankHiLo.Mid()) * run.process.targetRatios.Get(simType)
		}
	}

	// rank combined sims
	for entry, simRankHiLo := range util.CalculateRankingRanges(true, run.runData, func(x *rankEntry4) float64 { return x.simScore }) {
		entry.targetRank = simRankHiLo.Lo
	}
}

func (run *rankInternalRun4) makeDataListEntryColumns() {
	maxRank := float64(len(run.runData) - 1)

	primes := util.PrimesSmall(int64(len(run.runData)))
	sumRanksByPrime := utilhighs.ConstraintRowBuild{Debug: "sumRanksByPrime"}
	targetSum := 0.0

	for i := range run.runData {
		entry := &run.runData[i]
		primeMultiplier := float64(primes[i])

		run.makeEntryColumnRefs(entry, maxRank)

		sumRanksByPrime.Add(entry.rankColumn, primeMultiplier)
		targetSum += float64(entry.targetRank) * primeMultiplier
	}

	// sumRanksByPrimeDiff := run.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugText("sumRanksByPrimeDiff"))
	// sumRanksByPrime.Add(sumRanksByPrimeDiff, 1)
	// sumRanksByPrime.Finish(run.input, targetSum, targetSum)

	// sumRanksByPrimeDiffAbs := run.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugText("sumRanksByPrimeDiffAbs"))
	// utilhighs.AbsoluteValue(run.input, sumRanksByPrimeDiff, sumRanksByPrimeDiffAbs)

	// so with this we roughly score on correct ordering.
	// individual ranks will have their own factors which could complicate things but probably not many ways to get exact sum
}

func (run *rankInternalRun4) makeEntryColumnRefs(entry *rankEntry4, maxRank float64) {
	// these scores are meaningless in themselves, at least in value terms
	// however their increasing sequence should correlate to combinedSimRankScore
	// which is what we'll optimise for
	rankStr := strconv.FormatInt(int64(entry.targetRank), 10)
	entry.scoreColumn = run.input.CreateColumnGeneral(highs.Continuous, -c_RankLimitScore, c_RankLimitScore, utilhighs.DebugText("score-"+rankStr))

	scoreRow := utilhighs.ConstraintRowBuild{Debug: "scoreRow"}
	for _, statType := range G_RequiredStats {
		weightColumn := run.weightColumns[statType]
		statValue := entry.data.TotalStat.GetFloat(statType)
		statScale := run.scaleStats[statType]

		scoreRow.Add(weightColumn, statValue*statScale)
	}
	scoreRow.Add(entry.scoreColumn, -1)
	scoreRow.Finish(run.input, 0, 0)

	entry.rankColumn = run.input.CreateColumnGeneral(highs.Integer, 0, maxRank, utilhighs.DebugText("derivedRank-"+rankStr))

	rankDiff := run.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugText("rankDiff"))
	utilhighs.AbsoluteValueFromDiffOneToConst(run.input, entry.rankColumn, 1, float64(entry.targetRank), rankDiff, "rankDiff")
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
	// so we could totally do a boolean thing where scoreA>scoreB then implies rankA>rankB
	// would need all possible pairs connected, but would then force solver to make a full integer order

	oneIsGreater := run.input.CreateColumnBool(utilhighs.DebugText("oneIsGreater"))
	// twoIsGreater := run.input.CreateColumnBool(utilhighs.DebugText("twoIsGreater"))
	// utilhighs.ColumnIsGreaterOrEqualColumn(run.input, two.scoreColumn, one.scoreColumn, oneIsGreater, c_RankLargeScore)
	// utilhighs.ColumnIsGreaterOrEqualColumn(run.input, one.scoreColumn, two.scoreColumn, twoIsGreater, c_RankLargeScore)

	oneGreaterRow := utilhighs.ConstraintRowBuild{Debug: "oneGreaterRow"}
	oneGreaterRow.Add(two.scoreColumn, -1)
	oneGreaterRow.Add(one.scoreColumn, 1)
	oneGreaterRow.Add(oneIsGreater, -c_RankLargeScore)
	oneGreaterRow.Finish(run.input, -c_RankLargeScore, 0)

	// twoGreaterRow := utilhighs.ConstraintRowBuild{Debug: "twoGreaterRow"}
	// twoGreaterRow.Add(one.scoreColumn, -1)
	// twoGreaterRow.Add(two.scoreColumn, 1)
	// twoGreaterRow.Add(twoIsGreater, -c_RankLargeScore)
	// twoGreaterRow.Finish(run.input, -c_RankLargeScore, 0)

	// utilhighs.ConstraintNot(run.input, oneIsGreater, twoIsGreater)

	// -range <= two - one - w*range + n*range <= range
	// if two>one: -range <= positive - w*range + n*range <= range   :::   -range <= positive - w*range + n*range   :::  w=01 n=01
	//                                                                     positive - w*range + n*range <= range    :::  w=01 n=0
	// mixRow := utilhighs.ConstraintRowBuild{Debug: ""}
	// mixRow.Add(one.scoreColumn, -1)
	// mixRow.Add(two.scoreColumn, 1)
	// mixRow.Add(twoIsGreater, -c_RankLargeScore)
	// mixRow.Add(oneIsGreater, c_RankLargeScore)
	// mixRow.Finish(run.input, -c_RankLargeScore, c_RankLargeScore)

	run.pairLinks.Put(indexOne, indexTwo, &rankPair4{
		entryOne:     one,
		entryTwo:     two,
		oneIsGreater: oneIsGreater,
		// twoIsGreater: twoIsGreater,
	})
}

func (run *rankInternalRun4) makeRankDerivation(mainIndex int, mainEntry *rankEntry4) {
	sumCompareFlags := utilhighs.ConstraintRowBuild{Debug: "rankDerive " + strconv.FormatInt(int64(mainIndex), 10)}

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
	sumCompareFlags.Finish(run.input, rowSumTarget, rowSumTarget)
}

func (run *rankInternalRun4) extractAndReportSolution(solution *highs.Solution) WeightResult {
	run.input.DebugPrintColumns(solution, run.process.printer)

	run.process.printer.Println("WEIGHTS")

	statWeightResult := WeightResult_Make()
	for _, statType := range G_RequiredStats {
		weightColumn := run.weightColumns[statType]
		statScale := run.scaleStats[statType]

		modelWeight := solution.ColValues[weightColumn]
		usableWeight := modelWeight / statScale

		statWeightResult.Put(statType, usableWeight)
	}

	divideBy := statWeightResult.Get(stats.Stat_Strength)
	for _, statType := range G_RequiredStats {
		value := statWeightResult.Get(statType) / divideBy
		statWeightResult.Put(statType, value)
		run.process.printer.Printf("%10s %f\n", statType.Name(), value)
	}

	return statWeightResult
}

func (run *rankInternalRun4) setupInitialSolutionDumb1() {
	for statType, colWeight := range run.weightColumns {
		if statType == stats.Stat_Strength {
			run.input.SetInitialSolutionValue(colWeight, 1)
		} else {
			run.input.SetInitialSolutionValue(colWeight, 0)
		}
	}

	statScale := run.scaleStats[stats.Stat_Strength]
	for entry := range util.ForPointer(run.runData) {
		entry.initialStatScore = entry.data.TotalStat.GetFloat(stats.Stat_Strength) * statScale
	}

	run.setupRemainingInitialSolution()
	run.input.ValidateInitialSolutionState()
}

func (run *rankInternalRun4) setupInitialSolutionDumb2() {
	initialDumbValue := 0.2

	dumbWeights := WeightResult_Make()
	for statType, colWeight := range run.weightColumns {
		run.input.SetInitialSolutionValue(colWeight, initialDumbValue)
		dumbWeights.Put(statType, initialDumbValue)
	}

	for entry := range util.ForPointer(run.runData) {
		entry.initialStatScore = dumbWeights.CalcStatScoreScaled(entry.data, run.scaleStats)
	}

	run.setupRemainingInitialSolution()
	run.input.ValidateInitialSolutionState()
}

func (run *rankInternalRun4) setupInitialSolutionFromPrevious(previous *rankInternalRun4, solution *highs.Solution) {
	internalWeights := WeightResult_Make()
	for statType, oldWeightCol := range previous.weightColumns {
		newWeightCol := run.weightColumns[statType]
		oldWeightValue := solution.ColValues[oldWeightCol]
		run.input.SetInitialSolutionValue(newWeightCol, oldWeightValue)
		internalWeights.Put(statType, oldWeightValue)
	}

	for entry := range util.ForPointer(run.runData) {
		entry.initialStatScore = internalWeights.CalcStatScoreScaled(entry.data, run.scaleStats)
	}

	run.setupRemainingInitialSolution()
	run.input.ValidateInitialSolutionState()
}

func (run *rankInternalRun4) setupInitialSolutionFromExternal(weights stats.StatBlock) {
	internalWeights := WeightResult_Make()
	for statType, colWeight := range run.weightColumns {
		basicValue := weights.GetFloat(statType) / 1000.0 // NOTE reverse scale as used in StatRatingsWeights
		// scale := run.scaleStats[statType]
		scaledValue := basicValue 
		run.input.SetInitialSolutionValue(colWeight, scaledValue)
		internalWeights.Put(statType, scaledValue)
	}

	for entry := range util.ForPointer(run.runData) {
		entry.initialStatScore = internalWeights.CalcStatScoreScaled(entry.data, run.scaleStats)
		run.input.SetInitialSolutionValue(entry.scoreColumn, entry.initialStatScore)
	}

	run.setupRemainingInitialSolution()
	run.input.ValidateInitialSolutionState()
}

func (run *rankInternalRun4) setupInitialSolutionFromExternal2(weights WeightResult) {
	internalWeights := WeightResult_Make()
	for statType, colWeight := range run.weightColumns {
		basicValue := weights.Get(statType)
		// scale := run.scaleStats[statType]
		scaledValue := basicValue 
		run.input.SetInitialSolutionValue(colWeight, scaledValue)
		internalWeights.Put(statType, scaledValue)
	}

	for entry := range util.ForPointer(run.runData) {
		entry.initialStatScore = internalWeights.CalcStatScoreScaled(entry.data, run.scaleStats)
		run.input.SetInitialSolutionValue(entry.scoreColumn, entry.initialStatScore)
	}

	run.setupRemainingInitialSolution()
	run.input.ValidateInitialSolutionState()
}

func (run *rankInternalRun4) setupRemainingInitialSolution() {
	// for entry, rankHiLo := range util.CalculateRankingRanges(true, run.runData, func(x *rankEntry4) float64 { return x.initialStatScore }) {
	// run.input.SetInitialSolutionValue(entry.rankColumn, float64(rankHiLo.Lo))
	// }

	for pair := range run.pairLinks.SeqValues() {
		entryOne := pair.entryOne
		entryTwo := pair.entryTwo
		if entryOne.initialStatScore > entryTwo.initialStatScore {
			run.input.SetInitialSolutionValue(pair.oneIsGreater, 1)
			// run.input.SetInitialSolutionValue(pair.twoIsGreater, 0)
		} else if entryOne.initialStatScore < entryTwo.initialStatScore {
			run.input.SetInitialSolutionValue(pair.oneIsGreater, 0)
			// run.input.SetInitialSolutionValue(pair.twoIsGreater, 1)
		} else {
			// run.input.SetInitialSolutionValue(pair.oneIsGreater, 0)
			// run.input.SetInitialSolutionValue(pair.twoIsGreater, 0)
		}
	}
}
