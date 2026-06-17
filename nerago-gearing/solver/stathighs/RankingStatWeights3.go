package stathighs

import (
	"cmp"
	"math"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"slices"
	"strconv"
	"time"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
// c_RankLargeWeight = 50.0
// c_RankLargeScore  = 500.0
// c_RankLargeRank   = 10000.0
)

type RankingStatWeightProcess3 struct {
	printer *util.PrintRecorder

	targetRatios    simulate.SimData
	dataAllOriginal []rankEntry3
	dataSample      []rankEntry3

	build *utilhighs.LinearBuilder

	scaleStats    map[stats.StatType]float64
	weightColumns map[stats.StatType]utilhighs.ColumnIndex
	pairLinks     util.MapMap[int, int, rankPair3]
}

type rankEntry3 struct {
	data *WeightInput

	initialStatScore float64
	simScore         float64
	targetRank       int
	primeMultiplier  int64

	scoreColumn       utilhighs.ColumnIndex
	rankColumn        utilhighs.ColumnIndex
	rankDiffAbsColumn utilhighs.ColumnIndex
}

type rankPair3 struct {
	indexOne, indexTwo            int
	isGreaterScore, isGreaterRank utilhighs.ColumnIndex
	isSequenceDiff                utilhighs.ColumnIndex
}

func (ranker *RankingStatWeightProcess3) Init(printer *util.PrintRecorder) {
	ranker.printer = printer
}

func (ranker *RankingStatWeightProcess3) SupplyData(inputData []WeightInput) {
	ranker.scaleStats = chooseStatScaling(inputData, ranker.printer)
	ranker.dataAllOriginal = util.MapSliceAsNew(inputData, func(input *WeightInput) rankEntry3 {
		return rankEntry3{
			data:              input,
			simScore:          -1,
			targetRank:        -1,
			scoreColumn:       -1,
			rankColumn:        -1,
			rankDiffAbsColumn: -1,
			// rankDiffColumn:    -1,
		}
	})
}

func (ranker *RankingStatWeightProcess3) SetTargetRatios(targetRatios simulate.SimData) {
	ranker.targetRatios = targetRatios
}

func (ranker *RankingStatWeightProcess3) Run(doRound3 bool) []WeightResult {
	weightResultList := make([]WeightResult, 0)

	ranker.build = new(utilhighs.LinearBuilder)
	ranker.build.Minimise = true

	// FIRST ROUND: minimal data, dumb initial values
	ranker.printer.Println("RankingStatWeightProcess3 FIRST ROUND")
	ranker.dataSample = takeDataSample_Start(ranker.dataAllOriginal, 100)
	ranker.prepareRankings()
	ranker.createWeightColumns()
	ranker.makeDataListEntryColumns()
	ranker.setupDumbInitialSolution()
	solution1, log := ranker.build.RunHighs()
	ranker.printer.AppendOther(log)
	weights := ranker.extractAndReportSolution(solution1)
	weightResultList = append(weightResultList, weights)

	// SECOND ROUND, minimal data, add extra conditions, copy initial from previous
	ranker.printer.Println("RankingStatWeightProcess3 SECOND ROUND")
	ranker.makeDataListPairRules()
	ranker.setupInitialSolutionFromPrevious(solution1)
	solution2, log := ranker.build.RunHighs()
	ranker.printer.AppendOther(log)
	weights = ranker.extractAndReportSolution(solution2)
	weightResultList = append(weightResultList, weights)

	// THIRD ROUND, full data, copy just weights from previous
	// data change means column ids won't line up
	// var weights3 WeightResult
	// if doRound3 {
	// 	ranker.printer.Println("RankingStatWeightProcess3 THIRD ROUND")
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

	latestSolution := solution2
	times := make(map[int]time.Duration)
	for size := 150; size < 400; size += 25 {
		startTime := time.Now()

		ranker.dataSample = takeDataSample_Start(ranker.dataAllOriginal, size)
		ranker.printer.Println("RankingStatWeightProcess3 THIRD ROUND " + strconv.Itoa(size))
		ranker.build = new(utilhighs.LinearBuilder)
		ranker.build.Minimise = true
		ranker.build.TimeLimitSeconds = 2000
		ranker.prepareRankings()
		ranker.createWeightColumns()
		ranker.makeDataListEntryColumns()
		ranker.makeDataListPairRules()
		ranker.setupInitialSolutionFromPreviousWeightOnly(latestSolution)
		solution3, log := ranker.build.RunHighs()
		ranker.printer.AppendOther(log)
		weights = ranker.extractAndReportSolution(solution3)
		weightResultList = append(weightResultList, weights)
		latestSolution = solution3

		timeTaken := time.Since(startTime)
		times[size] = timeTaken

		if solution3.Status == highs.ModelStatusTimeLimit {
			break
		}
	}

	for size, duration := range times {
		ranker.printer.Printf("%4d %s\n", size, duration)
	}

	return weightResultList
}

func (ranker *RankingStatWeightProcess3) RunUsingExternalStart(initialWeight WeightResult) util.Optional[WeightResult] {
	ranker.build = new(utilhighs.LinearBuilder)
	ranker.build.Minimise = true
	ranker.build.TimeLimitSeconds = 5000

	ranker.printer.Println("RankingStatWeightProcess3 FIRST ROUND")
	ranker.dataSample = ranker.dataAllOriginal
	ranker.prepareRankings()
	ranker.createWeightColumns()
	ranker.makeDataListEntryColumns()
	ranker.makeDataListPairRules()
	ranker.setupInitialSolutionFromExternal2(initialWeight)
	solution, log := ranker.build.RunHighs()
	ranker.printer.AppendOther(log)
	if solution.HasSolution() {
		weights := ranker.extractAndReportSolution(solution)
		return util.Optional_OfValue(weights)
	} else {
		return util.Optional_Empty[WeightResult]()
	}
}

func (ranker *RankingStatWeightProcess3) createWeightColumns() {
	lo := -c_RankLargeWeight
	hi := c_RankLargeWeight

	sumWeights := utilhighs.ConstraintRow{Debug: "sumWeights"}
	ranker.weightColumns = make(map[stats.StatType]utilhighs.ColumnIndex)
	for _, statType := range G_RequiredStats {
		colWeight := ranker.build.CreateColumnGeneral(highs.Continuous, lo, hi, utilhighs.DebugString{Text: "WEIGHT " + statType.Name()})
		ranker.weightColumns[statType] = colWeight
		sumWeights.Add(colWeight, 1)
	}

	sumWeights.Build(ranker.build, 0.0001, utilhighs.C_PlusInf) // force positive and non-zero result. would use 1 but stat scaling confuses things
}

func (ranker *RankingStatWeightProcess3) prepareRankings() {
	// reset values
	for i := range ranker.dataSample {
		ranker.dataSample[i].simScore = 0
		ranker.dataSample[i].targetRank = 0
	}

	// score each sim
	for _, simType := range G_RequiredSims {
		for entry, simDetailRank := range util.CalculateRanking(simType.IsHighGood(), ranker.dataSample, func(x *rankEntry3) float64 { return x.data.SimResult.Get(simType) }) {
			entry.simScore += float64(simDetailRank) * ranker.targetRatios.Get(simType)
		}
	}

	// rank combined sims
	for entry, simRank := range util.CalculateRanking(true, ranker.dataSample, func(x *rankEntry3) float64 { return x.simScore }) {
		entry.targetRank = simRank
	}
}

func (ranker *RankingStatWeightProcess3) makeDataListEntryColumns() {
	maxRank := float64(len(ranker.dataSample) - 1)

	sumRanks := utilhighs.ConstraintRow{Debug: "sumRanks"}
	for entry := range util.ForPointer(ranker.dataSample) {
		ranker.makeEntryColumns(entry, maxRank)
		sumRanks.Add(entry.rankColumn, 1)
	}

	expectedSum := float64(len(ranker.dataSample)) * float64(len(ranker.dataSample)-1) / 2.0
	allowedSumRange := 0.0 // a bit of flex for rounding, as well as just to support easier solving
	sumRanks.Build(ranker.build, expectedSum-allowedSumRange, expectedSum+allowedSumRange)
}

func (ranker *RankingStatWeightProcess3) makeDataListPairRules() {
	for baseIndex := range ranker.dataSample {
		for compareTo := baseIndex + 1; compareTo < len(ranker.dataSample); compareTo++ {
			ranker.makeEntryPairSequenceConstraints(&ranker.dataSample[baseIndex], &ranker.dataSample[compareTo], baseIndex, compareTo, 1)
		}
	}

	// for baseIndex := range data {
	// 	for jump := 1; jump <= len(data); jump *= 2 {
	// 		compareTo := baseIndex + jump
	// 		if compareTo < len(data) {
	// 			ranker.printer.Printf("compare pairs %3d %3d\n", baseIndex, compareTo)
	// 			ranker.makeEntryPairSequenceConstraints(&data[baseIndex], &data[compareTo], baseIndex, compareTo, jump)
	// 		}
	// 	}
	// }
}

func (ranker *RankingStatWeightProcess3) makeEntryColumns(entry *rankEntry3, maxRank float64) {
	// these scores are meaningless in themselves, at least in value terms
	// however their increasing sequence should correlate to combinedSimRankScore
	// which is what we'll optimise for
	rankStr := strconv.FormatInt(int64(entry.targetRank), 10)
	entry.scoreColumn = ranker.build.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugText("score-"+rankStr))

	scoreRow := utilhighs.ConstraintRow{Debug: "scoreRow"}
	for _, statType := range G_RequiredStats {
		weightColumn := ranker.weightColumns[statType]
		statValue := entry.data.TotalStat.GetFloat(statType)
		statScale := ranker.scaleStats[statType]

		scoreRow.Add(weightColumn, statValue*statScale)
	}
	scoreRow.Add(entry.scoreColumn, -1)
	scoreRow.Build(ranker.build, 0, 0)

	entry.rankColumn = ranker.build.CreateColumnGeneral(highs.Integer, 0, maxRank, utilhighs.DebugText("derivedRank-"+rankStr))
	entry.rankDiffAbsColumn = ranker.build.CreateColumnWithOutput(highs.Integer, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugText("rankDiffAbs-"+rankStr))

	targetRank := float64(entry.targetRank)
	ranker.build.AbsoluteValueFromDiffOneToConst(entry.rankColumn, 1, targetRank, entry.rankDiffAbsColumn, "")
}

func (ranker *RankingStatWeightProcess3) makeEntryPairSequenceConstraints(one *rankEntry3, two *rankEntry3, indexOne, indexTwo int, scaleDiffOutput int) {
	// so we could totally do a boolean thing where scoreA>scoreB then implies rankA>rankB
	// would need all possible pairs connected, but would then force solver to make a full integer order
	isGreaterScore := ranker.build.CreateColumnBool(utilhighs.DebugText("isGreaterScore"))
	isGreaterRank := ranker.build.CreateColumnBool(utilhighs.DebugText("isGreaterRank"))
	isSequenceDiff := ranker.build.CreateColumnBoolWithOutput(float64(scaleDiffOutput), utilhighs.DebugText("sequenceDiff"))

	ranker.build.ColumnIsGreaterOrEqualColumn(one.scoreColumn, two.scoreColumn, isGreaterScore, c_RankLargeScore, 0.0001)
	ranker.build.ColumnIsGreaterOrEqualColumn(one.rankColumn, two.rankColumn, isGreaterRank, c_RankLargeRank, 1.0)
	ranker.build.IsXor(isGreaterRank, isGreaterScore, isSequenceDiff)

	ranker.pairLinks.Put(indexOne, indexTwo, rankPair3{
		indexOne:       indexOne,
		indexTwo:       indexTwo,
		isGreaterScore: isGreaterScore,
		isGreaterRank:  isGreaterRank,
		isSequenceDiff: isSequenceDiff,
	})
}

// we do a complete ranking just on the strength stat
func (ranker *RankingStatWeightProcess3) setupDumbInitialSolution() {
	for statType, colWeight := range ranker.weightColumns {
		if statType == stats.Stat_Strength {
			ranker.build.SetInitialSolutionValue(colWeight, 1)
		} else {
			ranker.build.SetInitialSolutionValue(colWeight, 0)
		}
	}

	statScale := ranker.scaleStats[stats.Stat_Strength]
	for entry, dumbRank := range util.CalculateRanking(true, ranker.dataSample, func(x *rankEntry3) float64 { return x.data.TotalStat.GetFloat(stats.Stat_Strength) }) {
		thisValue := entry.data.TotalStat.GetFloat(stats.Stat_Strength)
		scaledValue := thisValue * statScale
		ranker.build.SetInitialSolutionValue(entry.scoreColumn, scaledValue)
		ranker.build.SetInitialSolutionValue(entry.rankColumn, float64(dumbRank))
		diff := float64(dumbRank) - float64(entry.targetRank)
		// ranker.input.SetInitialSolutionValue(entry.rankDiffColumn, diff)
		ranker.build.SetInitialSolutionValue(entry.rankDiffAbsColumn, math.Abs(diff))
	}

	// TODO check if we even use pairs in dumb
	for pair := range ranker.pairLinks.SeqValues() {
		one, two := &ranker.dataSample[pair.indexOne], &ranker.dataSample[pair.indexTwo]
		scoreOne, scoreTwo := ranker.build.GetInitialSolutionValue(one.scoreColumn), ranker.build.GetInitialSolutionValue(two.scoreColumn)
		rankOne, rankTwo := ranker.build.GetInitialSolutionValue(one.rankColumn), ranker.build.GetInitialSolutionValue(two.rankColumn)
		if scoreTwo >= scoreOne {
			ranker.build.SetInitialSolutionValue(pair.isGreaterScore, 1)
			if rankTwo >= rankOne {
				ranker.build.SetInitialSolutionValue(pair.isGreaterRank, 1)
				ranker.build.SetInitialSolutionValue(pair.isSequenceDiff, 0)
			} else {
				ranker.build.SetInitialSolutionValue(pair.isGreaterRank, 0)
				ranker.build.SetInitialSolutionValue(pair.isSequenceDiff, 1)
			}
		} else {
			ranker.build.SetInitialSolutionValue(pair.isGreaterScore, 0)
			if rankTwo >= rankOne {
				ranker.build.SetInitialSolutionValue(pair.isGreaterRank, 1)
				ranker.build.SetInitialSolutionValue(pair.isSequenceDiff, 1)
			} else {
				ranker.build.SetInitialSolutionValue(pair.isGreaterRank, 0)
				ranker.build.SetInitialSolutionValue(pair.isSequenceDiff, 0)
			}
		}
	}

	ranker.build.ValidateInitialSolutionState()
}

func (ranker *RankingStatWeightProcess3) setupInitialSolutionFromPrevious(solution *highs.Solution) {
	ranker.build.ClearInitialSolutionValue()
	for i := range solution.ColValues {
		ranker.build.SetInitialSolutionValue(utilhighs.ColumnIndex(i), solution.ColValues[i])
	}

	ranker.build.ValidateInitialSolutionState()
}

// data []rankEntry3, weights map[stats.StatType]float64
func (ranker *RankingStatWeightProcess3) setupInitialSolutionFromPreviousWeightOnly(solution *highs.Solution) {
	internalWeights := WeightResult_Make()
	for statType, colWeight := range ranker.weightColumns {
		weight := solution.ColValues[colWeight]
		ranker.build.SetInitialSolutionValue(colWeight, weight)
		internalWeights.Put(statType, weight)
	}

	for entry := range util.ForPointer(ranker.dataSample) {
		entry.initialStatScore = internalWeights.CalcStatScoreScaled(entry.data, ranker.scaleStats)
	}

	for entry, calcRank := range util.CalculateRanking(true, ranker.dataSample, func(x *rankEntry3) float64 { return x.initialStatScore }) {
		ranker.build.SetInitialSolutionValue(entry.scoreColumn, entry.initialStatScore)
		ranker.build.SetInitialSolutionValue(entry.rankColumn, float64(calcRank))
		diff := float64(calcRank) - float64(entry.targetRank)
		ranker.build.SetInitialSolutionValue(entry.rankDiffAbsColumn, math.Abs(diff))
	}

	ranker.build.ValidateInitialSolutionState()
}

func (ranker *RankingStatWeightProcess3) setupInitialSolutionFromExternal2(weights WeightResult) {
	internalWeights := WeightResult_Make()
	for statType, colWeight := range ranker.weightColumns {
		basicValue := weights.Get(statType)
		scale := ranker.scaleStats[statType]
		scaledValue := basicValue * scale
		ranker.build.SetInitialSolutionValue(colWeight, scaledValue)
		internalWeights.Put(statType, scaledValue)
	}

	for entry := range util.ForPointer(ranker.dataSample) {
		entry.initialStatScore = internalWeights.CalcStatScoreScaled(entry.data, ranker.scaleStats)
	}

	for entry, calcRank := range util.CalculateRanking(true, ranker.dataSample, func(x *rankEntry3) float64 { return x.initialStatScore }) {
		ranker.build.SetInitialSolutionValue(entry.scoreColumn, entry.initialStatScore)
		ranker.build.SetInitialSolutionValue(entry.rankColumn, float64(calcRank))
		diff := float64(calcRank) - float64(entry.targetRank)
		ranker.build.SetInitialSolutionValue(entry.rankDiffAbsColumn, math.Abs(diff))
	}

	ranker.build.ValidateInitialSolutionState()
}

func (ranker *RankingStatWeightProcess3) extractAndReportSolution(solution *highs.Solution) WeightResult {
	ranker.build.DebugPrintColumns(solution, ranker.printer)

	ranker.printer.Println("WEIGHTS")

	statWeightResult := WeightResult_Make()
	for _, statType := range G_RequiredStats {
		weightColumn := ranker.weightColumns[statType]
		statScale := ranker.scaleStats[statType]

		modelWeight := solution.ColValues[weightColumn]
		usableWeight := modelWeight / statScale

		statWeightResult.Put(statType, usableWeight)
	}

	divideBy := statWeightResult.Get(stats.Stat_Strength)
	for _, statType := range G_RequiredStats {
		statWeightResult.Put(statType, statWeightResult.Get(statType)/divideBy)
	}

	ranker.reportRankingOfInputs(statWeightResult, solution)

	return statWeightResult
}

func (ranker *RankingStatWeightProcess3) reportRankingOfInputs(statWeightResult WeightResult, solution *highs.Solution) {
	ranker.printer.Println("INPUT CHECK")

	type entryCheck struct {
		simScore   float64
		targetRank int
		calcScore  float64
		colScore   float64
		colRank    float64
	}

	check := util.MapSliceAsNew(ranker.dataSample, func(x *rankEntry3) entryCheck {
		return entryCheck{
			x.simScore,
			x.targetRank,
			statWeightResult.CalcStatScore(x.data),
			solution.ColValues[x.scoreColumn],
			solution.ColValues[x.rankColumn],
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
