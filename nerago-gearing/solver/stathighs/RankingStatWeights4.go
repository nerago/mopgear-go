package stathighs

import (
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
	c_RankLimitScore = 200.0

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
	indexOne     int
	oneIsGreater utilhighs.ColumnIndex

	indexTwo     int
	twoIsGreater utilhighs.ColumnIndex
}

func (process *RankingStatWeightProcess4) Init(printer *util.PrintRecorder) {
	process.printer = printer
}

func (process *RankingStatWeightProcess4) SupplyData(inputData []WeightInput) {
	process.dataAll = inputData
}

func (run *rankInternalRun4) SupplyData(inputData []WeightInput) {
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

func (process *RankingStatWeightProcess4) Run(doRound3 bool) []WeightResult {
	weightResultList := make([]WeightResult, 0)

	// FIRST ROUND: minimal data, dumb initial values
	process.printer.Println("RankingStatWeightProcess4 FIRST ROUND")
	minimalData := takeDataSample(process.dataAll, 40)
	run1 := rankInternalRun4_create(process)
	run1.SupplyData(minimalData)
	run1.prepareRankings()
	run1.createWeightColumns()
	run1.makeDataListEntryColumns()
	run1.makeDataListPairRules()
	// run1.setupDumbInitialSolution()
	weights, _ := run1.run()
	weights.ApplyIfValue(func(w WeightResult) { weightResultList = append(weightResultList, w) })

	// SECOND ROUND, minimal data, add extra conditions, copy initial from previous
	// rwp.printer.Println("RankingStatWeightProcess4 SECOND ROUND")
	// run2 := rankInternalRun4_create(rwp)
	// run2.SupplyData(minimalData)
	// run2.prepareRankings()
	// run2.createWeightColumns()
	// run2.makeDataListEntryColumns()
	// run2.makeDataListPairRules() // extra beyond first round
	// // run2.setupInitialSolutionFromPrevious(solution1)
	// weights, solution2 := run2.run()
	// weights.ApplyIfValue(func(w WeightResult) { weightResultList = append(weightResultList, w) })

	// THIRD ROUND, full data, copy just weights from previous
	// data change means column ids won't line up
	// var weights3 WeightResult
	// if doRound3 {
	// 	ranker.printer.Println("RankingStatWeightProcess4 THIRD ROUND")
	// 	ranker.input = new(utilhighs.InputBuilder)
	// 	ranker.input.Minimise = true
	// 	ranker.input.TimeLimitSeconds = 3600
	// 	fullData := ranker.dataAll
	// 	ranker.prepareRankings(fullData)
	// 	ranker.createWeightColumns()
	// 	ranker.makeDataListEntryColumns(fullData)
	// 	ranker.makeDataListPairRules(fullData)
	// 	ranker.setupInitialSolutionFromPreviousWeightOnly(solution2, fullData)
	// 	solution3, log := ranker.input.RunHighs()
	// 	ranker.printer.AppendOther(log)
	// 	weights = ranker.extractAndReportSolution(solution3)
	// 	weightResultList = append(weightResultList, weights)
	// }

	// times := make(map[int]time.Duration)
	// for size := 150; size < 500; size += 50 {
	// 	startTime := time.Now()

	// 	dataSample := rwp.dataAll[0:size]
	// 	rwp.printer.Println("RankingStatWeightProcess4 THIRD ROUND " + strconv.Itoa(size))
	// 	run3 := rankInternalRun4_create(rwp)
	// 	run3.input.TimeLimitSeconds = 3600
	// 	run2.SupplyData(dataSample)
	// 	run3.prepareRankings()
	// 	run3.createWeightColumns()
	// 	run3.makeDataListEntryColumns()
	// 	run3.makeDataListPairRules()
	// 	run3.setupInitialSolutionFromPreviousWeightOnly(solution2)
	// 	weights, _ := run2.run()
	// 	weights.ApplyIfValue(func(w WeightResult) { weightResultList = append(weightResultList, w) })

	// 	timeTaken := time.Since(startTime)
	// 	times[size] = timeTaken
	// }

	// for size, duration := range times {
	// 	rwp.printer.Printf("%4d %s\n", size, duration)
	// }

	return weightResultList
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
		for entry, simDetailRank := range util.CalculateRanking(simType.IsHighGood(), run.runData, func(x *rankEntry4) float64 { return x.data.SimResult.Get(simType) }) {
			entry.simScore += float64(simDetailRank) * run.process.targetRatios.Get(simType)
		}
	}

	// rank combined sims
	for entry, simRank := range util.CalculateRanking(true, run.runData, func(x *rankEntry4) float64 { return x.simScore }) {
		entry.targetRank = simRank
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

	sumRanksByPrimeDiff := run.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugText("sumRanksByPrimeDiff"))
	sumRanksByPrime.Add(sumRanksByPrimeDiff, 1)
	sumRanksByPrime.Finish(run.input, targetSum, targetSum)

	sumRanksByPrimeDiffAbs := run.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugText("sumRanksByPrimeDiffAbs"))
	utilhighs.AbsoluteValue(run.input, sumRanksByPrimeDiff, sumRanksByPrimeDiffAbs)

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

	for a := 0; a < len(run.runData); a++ {
		run.makeRankDerivation(a)
	}
}

func (run *rankInternalRun4) makeEntryPairScoreChecks(one *rankEntry4, two *rankEntry4, indexOne, indexTwo int) {
	// so we could totally do a boolean thing where scoreA>scoreB then implies rankA>rankB
	// would need all possible pairs connected, but would then force solver to make a full integer order

	oneIsGreater := run.input.CreateColumnBool(utilhighs.DebugText("oneIsGreater"))
	twoIsGreater := run.input.CreateColumnBool(utilhighs.DebugText("twoIsGreater"))
	// utilhighs.ColumnIsGreaterOrEqualColumn(run.input, two.scoreColumn, one.scoreColumn, twoIsGreater, c_RankLargeScore)
	utilhighs.ColumnIsGreaterOrEqualColumn(run.input, one.scoreColumn, two.scoreColumn, twoIsGreater, c_RankLargeScore)
	utilhighs.ConstraintNot(run.input, twoIsGreater, oneIsGreater)

	run.pairLinks.Put(indexOne, indexTwo, &rankPair4{
		indexOne:     indexOne, // first index is always the "threshholdLow"
		indexTwo:     indexTwo,
		oneIsGreater: oneIsGreater,
		twoIsGreater: twoIsGreater,
	})
}

func (run *rankInternalRun4) makeRankDerivation(mainIndex int) {
	sumCompareFlags := utilhighs.ConstraintRowBuild{Debug: "rankDerive " + strconv.FormatInt(int64(mainIndex), 10)}
	// technically the not is 1-x, could just include negative var, and bump target number
	for _, pair := range run.pairLinks.SeqInnerWithKeyValue(mainIndex) {
		if pair.indexOne == mainIndex {
			sumCompareFlags.Add(pair.oneIsGreater, 1)
		} else {
			sumCompareFlags.Add(pair.twoIsGreater, 1)
		}
	}

	

	// rankFudge := run.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugText("rankFudge"))
	// sumCompareFlags.Add(rankFudge, 1)
	// rankFudgeAbs := run.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugText("rankFudgeAbs"))
	// utilhighs.AbsoluteValue(run.input, rankFudge, rankFudgeAbs)

	entry := &run.runData[mainIndex]
	sumCompareFlags.Add(entry.rankColumn, -1)
	sumCompareFlags.Finish(run.input, 0, 0)
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
		statWeightResult.Put(statType, statWeightResult.Get(statType)/divideBy)
	}

	return statWeightResult
}

func (run *rankInternalRun4) setupDumbInitialSolution() {
	for statType, colWeight := range run.weightColumns {
		if statType == stats.Stat_Strength {
			run.input.SetInitialSolutionValue(colWeight, 1)
		} else {
			run.input.SetInitialSolutionValue(colWeight, 0)
		}
	}

	statScale := run.scaleStats[stats.Stat_Strength]
	for entry, dumbRank := range util.CalculateRanking(true, run.runData, func(x *rankEntry4) float64 { return x.data.TotalStat.GetFloat(stats.Stat_Strength) }) {
		thisValue := entry.data.TotalStat.GetFloat(stats.Stat_Strength)
		scaledValue := thisValue * statScale
		run.input.SetInitialSolutionValue(entry.scoreColumn, scaledValue)
		run.input.SetInitialSolutionValue(entry.rankColumn, float64(dumbRank))
	}

	for pair := range run.pairLinks.SeqValues() {
		entryOne := &run.runData[pair.indexOne]
		entryTwo := &run.runData[pair.indexTwo]
		if entryOne.data.TotalStat.GetFloat(stats.Stat_Strength) > entryTwo.data.TotalStat.GetFloat(stats.Stat_Strength) {
			run.input.SetInitialSolutionValue(pair.oneIsGreater, 1)
			run.input.SetInitialSolutionValue(pair.twoIsGreater, 0)
		} else {
			run.input.SetInitialSolutionValue(pair.oneIsGreater, 0)
			run.input.SetInitialSolutionValue(pair.twoIsGreater, 1)
		}
	}

	run.input.ValidateInitialSolutionState()
}
