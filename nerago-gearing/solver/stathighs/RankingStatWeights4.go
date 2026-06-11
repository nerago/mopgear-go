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
	// pairLinks     util.MapMap[int, int, rankPair3]

	// primeRankSumRow utilhighs.ConstraintRowBuild
}

type rankEntry4 struct {
	data *WeightInput

	initialStatScore float64
	simScore         float64
	targetRank       int

	scoreColumn utilhighs.ColumnIndex
	rankColumn  utilhighs.ColumnIndex
	// rankDiffColumn    utilhighs.ColumnIndex
	// rankDiffAbsColumn utilhighs.ColumnIndex
}

type rankPair4 struct {
	indexOne, indexTwo            int
	isGreaterScore, isGreaterRank utilhighs.ColumnIndex
	sequenceDiff                  utilhighs.ColumnIndex
}

func (rwp *RankingStatWeightProcess4) Init(printer *util.PrintRecorder) {
	rwp.printer = printer
}

func (rwp *RankingStatWeightProcess4) SupplyData(inputData []WeightInput) {
	rwp.dataAll = inputData
}

func (rr *rankInternalRun4) SupplyData(inputData []WeightInput) {
	rr.scaleStats = chooseStatScaling(inputData, rr.process.printer)
	rr.runData = util.MapSliceAsNew(inputData, func(input *WeightInput) rankEntry4 {
		return rankEntry4{
			data:        input,
			simScore:    -1,
			targetRank:  -1,
			scoreColumn: -1,
			rankColumn:  -1,
			// rankDiffAbsColumn: -1,
		}
	})
}

func (rwp *RankingStatWeightProcess4) SetTargetRatios(targetRatios simulate.SimData) {
	rwp.targetRatios = targetRatios
}

func (rwp *RankingStatWeightProcess4) Run(doRound3 bool) []WeightResult {
	weightResultList := make([]WeightResult, 0)

	// FIRST ROUND: minimal data, dumb initial values
	rwp.printer.Println("RankingStatWeightProcess4 FIRST ROUND")
	minimalData := rwp.dataAll[0:100]
	run1 := rankInternalRun4_create(rwp)
	run1.SupplyData(minimalData)
	run1.prepareRankings()
	run1.createWeightColumns()
	run1.makeDataListEntryColumns()
	run1.setupDumbInitialSolution()
	weights, solution1 := run1.run()
	weights.ApplyIfValue(func(w WeightResult) { weightResultList = append(weightResultList, w) })

	// SECOND ROUND, minimal data, add extra conditions, copy initial from previous
	rwp.printer.Println("RankingStatWeightProcess4 SECOND ROUND")
	run2 := rankInternalRun4_create(rwp)
	run2.SupplyData(minimalData)
	run2.prepareRankings()
	run2.createWeightColumns()
	run2.makeDataListEntryColumns()
	run2.makeDataListPairRules() // extra beyond first round
	run2.setupInitialSolutionFromPrevious(solution1)
	weights, solution2 := run2.run()
	weights.ApplyIfValue(func(w WeightResult) { weightResultList = append(weightResultList, w) })

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

	times := make(map[int]time.Duration)
	for size := 150; size < 500; size += 50 {
		startTime := time.Now()

		dataSample := rwp.dataAll[0:size]
		rwp.printer.Println("RankingStatWeightProcess4 THIRD ROUND " + strconv.Itoa(size))
		run3 := rankInternalRun4_create(rwp)
		run3.input.TimeLimitSeconds = 3600
		run2.SupplyData(dataSample)
		run3.prepareRankings()
		run3.createWeightColumns()
		run3.makeDataListEntryColumns()
		run3.makeDataListPairRules()
		run3.setupInitialSolutionFromPreviousWeightOnly(solution2)
		weights, _ := run2.run()
		weights.ApplyIfValue(func(w WeightResult) { weightResultList = append(weightResultList, w) })

		timeTaken := time.Since(startTime)
		times[size] = timeTaken
	}

	for size, duration := range times {
		rwp.printer.Printf("%4d %s\n", size, duration)
	}

	return weightResultList
}

func rankInternalRun4_create(process *RankingStatWeightProcess4) *rankInternalRun4 {
	run := new(rankInternalRun4)
	run.process = process
	run.input = new(utilhighs.InputBuilder)
	run.input.Minimise = true
	return run
}

func (rr *rankInternalRun4) run() (util.Optional[WeightResult], *highs.Solution) {
	solution, log := rr.input.RunHighs()
	rr.process.printer.AppendOther(log)
	if solution.HasSolution() {
		weights := rr.extractAndReportSolution(solution)
		return util.Optional_OfValue(weights), solution
	} else {
		return util.Optional_Empty[WeightResult](), solution
	}
}

func (rr *rankInternalRun4) createWeightColumns() {
	lo := -c_RankLargeWeight
	hi := c_RankLargeWeight

	sumWeights := utilhighs.ConstraintRowBuild{Debug: "sumWeights"}
	rr.weightColumns = make(map[stats.StatType]utilhighs.ColumnIndex)
	for _, statType := range G_RequiredStats {
		colWeight := rr.input.CreateColumnGeneral(highs.Continuous, lo, hi, utilhighs.DebugString{Text: "WEIGHT " + statType.Name()})
		rr.weightColumns[statType] = colWeight
		sumWeights.Add(colWeight, 1)
	}

	sumWeights.Finish(rr.input, 1.0, hi) // force positive and non-zero result
}

func (rr *rankInternalRun4) prepareRankings() {
	// reset values
	for i := range rr.runData {
		rr.runData[i].simScore = 0
		rr.runData[i].targetRank = 0
	}

	// score each sim
	for _, simType := range G_RequiredSims {
		for entry, simDetailRank := range util.CalculateRanking(simType.IsHighGood(), rr.runData, func(x *rankEntry4) float64 { return x.data.SimResult.Get(simType) }) {
			entry.simScore += float64(simDetailRank) * rr.process.targetRatios.Get(simType)
		}
	}

	// rank combined sims
	for entry, simRank := range util.CalculateRanking(true, rr.runData, func(x *rankEntry4) float64 { return x.simScore }) {
		entry.targetRank = simRank
	}
}

func (rr *rankInternalRun4) makeDataListEntryColumns() {
	maxRank := float64(len(rr.runData) - 1)

	primes := util.PrimesSmall(int64(len(rr.runData)))
	sumRanksByPrime := utilhighs.ConstraintRowBuild{Debug: "sumRanksByPrime"}
	targetSum := 0.0

	for i := range rr.runData {
		entry := &rr.runData[i]
		primeMultiplier := float64(primes[i])

		rr.makeEntryColumnRefs(entry, maxRank)

		sumRanksByPrime.Add(entry.rankColumn, primeMultiplier)
		targetSum += float64(entry.targetRank) * primeMultiplier
	}

	sumRanksByPrimeDiff := rr.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugText("sumRanksByPrimeDiff"))
	sumRanksByPrime.Add(sumRanksByPrimeDiff, 1)
	sumRanksByPrime.Finish(rr.input, targetSum, targetSum)

	sumRanksByPrimeDiffAbs := rr.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugText("sumRanksByPrimeDiffAbs"))
	utilhighs.AbsoluteValue(rr.input, sumRanksByPrimeDiff, sumRanksByPrimeDiffAbs)

	// so with this we roughly score on correct ordering.
	// individual ranks will have their own factors which could complicate things but probably not many ways to get exact sum
}

func (rr *rankInternalRun4) makeEntryColumnRefs(entry *rankEntry4, maxRank float64) {
	// these scores are meaningless in themselves, at least in value terms
	// however their increasing sequence should correlate to combinedSimRankScore
	// which is what we'll optimise for
	rankStr := strconv.FormatInt(int64(entry.targetRank), 10)
	entry.scoreColumn = rr.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugText("score-"+rankStr))

	scoreRow := utilhighs.ConstraintRowBuild{Debug: "scoreRow"}
	for _, statType := range G_RequiredStats {
		weightColumn := rr.weightColumns[statType]
		statValue := entry.data.TotalStat.GetFloat(statType)
		statScale := rr.scaleStats[statType]

		scoreRow.Add(weightColumn, statValue*statScale)
	}
	scoreRow.Add(entry.scoreColumn, -1)
	scoreRow.Finish(rr.input, 0, 0)

	entry.rankColumn = rr.input.CreateColumnGeneral(highs.Integer, 0, maxRank, utilhighs.DebugText("derivedRank-"+rankStr))
	// entry.rankDiffAbsColumn = rr.input.CreateColumnWithOutput(highs.Integer, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugText("rankDiffAbs-"+rankStr))

	// targetRank := float64(entry.targetRank)
	// utilhighs.AbsoluteValueFromDiffOneToConst(rr.input, entry.rankColumn, 1, targetRank, entry.rankDiffAbsColumn, "")

	// entry.rankDiffColumn = ranker.input.CreateColumnGeneral(highs.Integer, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugText("rankDiff-"+rankStr))
	// rankDiffColumn:= ranker.input.CreateColumnGeneral(highs.Integer, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugText("rankDiff-"+rankStr))
	// rankDiff := utilhighs.ConstraintRowBuild{Debug: "rankDiff"}
	// rankDiff.Add(entry.rankColumn, 1)
	// // rankDiff.Add(entry.rankDiffColumn, -1)
	// rankDiff.Add(rankDiffColumn, -1)
	// rankDiff.Finish(ranker.input, targetRank, targetRank)
}

func (rr *rankInternalRun4) makeDataListPairRules() {
	eachCheckCount := 2
	for a := 0; a < len(rr.runData); a++ {
		for b := a + 1; b < min(a+eachCheckCount, len(rr.runData)); b++ {
			rr.makeEntryPairSequenceConstraints(&rr.runData[a], &rr.runData[b], a, b)
		}
	}

	// TODO some sort of cross multiply stuff
}

// parameters don't imply order
func (rr *rankInternalRun4) makeEntryPairSequenceConstraints(one *rankEntry4, two *rankEntry4, indexOne, indexTwo int) {
	// so we could totally do a boolean thing where scoreA>scoreB then implies rankA>rankB
	// would need all possible pairs connected, but would then force solver to make a full integer order
	isGreaterScore := rr.input.CreateColumnBool(utilhighs.DebugText("isGreaterScore"))
	isGreaterRank := rr.input.CreateColumnBool(utilhighs.DebugText("isGreaterRank"))
	sequenceDiff := rr.input.CreateColumnWithOutput(highs.Integer, 0, 1, 1, utilhighs.DebugText("sequenceDiff"))

	var (
		checkHighColumn utilhighs.ColumnIndex = two.scoreColumn
	)
	isGreaterEqual := utilhighs.ConstraintRowBuild{Debug: "isGreaterEqual"}
	isGreaterEqual.Add(one.scoreColumn, -1)
	isGreaterEqual.Add(checkHighColumn, 1)
	isGreaterEqual.Add(isGreaterScore, -c_RankLargeScore)
	isGreaterEqual.Finish(rr.input, -c_RankLargeScore, 0)
	{
		var (
			checkHighColumn utilhighs.ColumnIndex = two.rankColumn
		)
		isGreaterEqual := utilhighs.ConstraintRowBuild{Debug: "isGreaterEqual"}
		isGreaterEqual.Add(one.rankColumn, -1)
		isGreaterEqual.Add(checkHighColumn, 1)
		isGreaterEqual.Add(isGreaterRank, -c_RankLargeRank)
		isGreaterEqual.Finish(rr.input, -c_RankLargeRank, 0)
	}

	negative := utilhighs.ConstraintRowBuild{Debug: "Xor"}
	negative.Add(isGreaterRank, 1)
	negative.Add(isGreaterScore, -1)
	negative.Add(sequenceDiff, 1)
	negative.Finish(rr.input, 0, 2)
	positive := utilhighs.ConstraintRowBuild{Debug: "Xor"}
	positive.Add(isGreaterRank, 1)
	positive.Add(isGreaterScore, -1)
	positive.Add(sequenceDiff, -1)
	positive.Finish(rr.input, -2, 0)

	// rr.pairLinks.Put(indexOne, indexTwo, rankPair3{
	// 	indexOne:       indexOne,
	// 	indexTwo:       indexTwo,
	// 	isGreaterScore: isGreaterScore,
	// 	isGreaterRank:  isGreaterRank,
	// 	sequenceDiff:   sequenceDiff,
	// })
}

// we do a complete ranking just on the strength stat
func (rr *rankInternalRun4) setupDumbInitialSolution() {
	for statType, colWeight := range rr.weightColumns {
		if statType == stats.Stat_Strength {
			rr.input.SetInitialSolutionValue(colWeight, 1)
		} else {
			rr.input.SetInitialSolutionValue(colWeight, 0)
		}
	}

	statScale := rr.scaleStats[stats.Stat_Strength]
	for entry, dumbRank := range util.CalculateRanking(true, rr.runData, func(x *rankEntry4) float64 { return x.data.TotalStat.GetFloat(stats.Stat_Strength) }) {
		thisValue := entry.data.TotalStat.GetFloat(stats.Stat_Strength)
		scaledValue := thisValue * statScale
		rr.input.SetInitialSolutionValue(entry.scoreColumn, scaledValue)
		rr.input.SetInitialSolutionValue(entry.rankColumn, float64(dumbRank))
		// diff := float64(dumbRank) - float64(entry.targetRank)
		// ranker.input.SetInitialSolutionValue(entry.rankDiffColumn, diff)
		// rr.input.SetInitialSolutionValue(entry.rankDiffAbsColumn, math.Abs(diff))
	}

	// TODO check if we even use pairs in dumb
	// for pair := range rr.pairLinks.SeqValues() {
	// 	one, two := &data[pair.indexOne], &data[pair.indexTwo]
	// 	scoreOne, scoreTwo := rwp.input.GetInitialSolutionValue(one.scoreColumn), rwp.input.GetInitialSolutionValue(two.scoreColumn)
	// 	rankOne, rankTwo := rwp.input.GetInitialSolutionValue(one.rankColumn), rwp.input.GetInitialSolutionValue(two.rankColumn)
	// 	if scoreTwo >= scoreOne {
	// 		rwp.input.SetInitialSolutionValue(pair.isGreaterScore, 1)
	// 		if rankTwo >= rankOne {
	// 			rwp.input.SetInitialSolutionValue(pair.isGreaterRank, 1)
	// 			rwp.input.SetInitialSolutionValue(pair.sequenceDiff, 0)
	// 		} else {
	// 			rwp.input.SetInitialSolutionValue(pair.isGreaterRank, 0)
	// 			rwp.input.SetInitialSolutionValue(pair.sequenceDiff, 1)
	// 		}
	// 	} else {
	// 		rwp.input.SetInitialSolutionValue(pair.isGreaterScore, 0)
	// 		if rankTwo >= rankOne {
	// 			rwp.input.SetInitialSolutionValue(pair.isGreaterRank, 1)
	// 			rwp.input.SetInitialSolutionValue(pair.sequenceDiff, 1)
	// 		} else {
	// 			rwp.input.SetInitialSolutionValue(pair.isGreaterRank, 0)
	// 			rwp.input.SetInitialSolutionValue(pair.sequenceDiff, 0)
	// 		}
	// 	}
	// }

	rr.input.ValidateInitialSolutionState()
}

func (rr *rankInternalRun4) setupInitialSolutionFromPrevious(solution *highs.Solution) {
	rr.input.ClearInitialSolutionValue()
	for i := range solution.ColValues {
		rr.input.SetInitialSolutionValue(utilhighs.ColumnIndex(i), solution.ColValues[i])
	}
	// ranker.input.ValidateInitialSolutionState()
}

// data []rankEntry3, weights map[stats.StatType]float64
func (rr *rankInternalRun4) setupInitialSolutionFromPreviousWeightOnly(solution *highs.Solution) {
	internalWeights := WeightResult_Make()
	for statType, colWeight := range rr.weightColumns {
		weight := solution.ColValues[colWeight]
		rr.input.SetInitialSolutionValue(colWeight, weight)
		internalWeights.Put(statType, weight)
	}

	for entry := range util.ForPointer(rr.runData) {
		entry.initialStatScore = internalWeights.CalcStatScoreScaled(entry.data, rr.scaleStats)
	}

	for entry, calcRank := range util.CalculateRanking(true, rr.runData, func(x *rankEntry4) float64 { return x.initialStatScore }) {
		rr.input.SetInitialSolutionValue(entry.scoreColumn, entry.initialStatScore)
		rr.input.SetInitialSolutionValue(entry.rankColumn, float64(calcRank))
		// diff := float64(calcRank) - float64(entry.targetRank)
		// rr.input.SetInitialSolutionValue(entry.rankDiffAbsColumn, math.Abs(diff))
	}
}

func (rr *rankInternalRun4) extractAndReportSolution(solution *highs.Solution) WeightResult {
	rr.input.DebugPrintColumns(solution, rr.process.printer)

	rr.process.printer.Println("WEIGHTS")

	statWeightResult := WeightResult_Make()
	for _, statType := range G_RequiredStats {
		weightColumn := rr.weightColumns[statType]
		statScale := rr.scaleStats[statType]

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

// func (ranker *RankingStatWeightProcess4) reportRankingOfInputs(statWeightResult map[stats.StatType]float64) {
// 	ranker.printer.Println("INPUT CHECK (index, combinedSimRank, calcStat)")
// 	for i, entry := range ranker.data {
// 		ranker.printer.Printf("%4d %8f %8f\n", i, entry.combinedSimRankScore, calcStatScore(entry.data, statWeightResult))
// 	}
// }
