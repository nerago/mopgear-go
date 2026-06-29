package stathighs

import (
	"cmp"
	"math"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"slices"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_rank3_scaleTarget         = 10.0
	c_rank3_initial_data_sample = 12
	c_rank3_min_total_weight    = 1.0
	c_rank3_time_limit          = 5000

	c_Rank3_LargeWeight = 10.0
	c_Rank3_LargeScore  = 500.0
	c_Rank3_LargeRank   = 10000.0
)

type RankingStatWeightProcess3 struct {
	printer *util.PrintRecorder

	targetRatios    stats.SimData
	requiredStats   []stats.StatType
	requiredSims    []stats.SimType
	dataAllOriginal []rankEntry3
	dataSample      []rankEntry3
	SCALE1          bool
	ALGO            int

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
	//primeMultiplier  int64

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
	if ranker.SCALE1 {
		ranker.scaleStats = make(map[stats.StatType]float64)
		for _, statType := range stats.StatType_List {
			ranker.scaleStats[statType] = 1
		}
	} else {
		ranker.scaleStats = chooseStatScaling(inputData, c_rank3_scaleTarget, ranker.printer)
	}
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

func (ranker *RankingStatWeightProcess3) SetRequiredStats(requiredStats []stats.StatType) {
	ranker.requiredStats = requiredStats
}

func (ranker *RankingStatWeightProcess3) SetTargetRatios(targetRatios stats.SimData) {
	ranker.targetRatios = targetRatios
	ranker.requiredSims = targetRatios.NonZeroTypes()
}

func (ranker *RankingStatWeightProcess3) makeBuilder() {
	ranker.build = new(utilhighs.LinearBuilder)
	ranker.build.Minimise = true
	ranker.build.TimeLimitSeconds = c_rank3_time_limit
	if ranker.ALGO >= 2 {
		ranker.build.Solver = utilhighs.Solver_LP_USE_GPU
	} else {
		ranker.build.Solver = utilhighs.Solver_MIP_Interior
	}
}

func (ranker *RankingStatWeightProcess3) Run(stopwatch *util.Stopwatch) WeightResult {

	// FIRST ROUND: minimal data, no initial values
	ranker.dataSample = takeDataSample_Start(ranker.dataAllOriginal, c_rank3_initial_data_sample)
	ranker.makeBuilder()
	ranker.prepareRankings()
	ranker.createWeightColumns()
	ranker.doAlgos()
	//ranker.setupDumbInitialSolution()
	solution1 := ranker.build.RunHighs(ranker.printer, stopwatch)
	_ = ranker.extractAndReportSolution(solution1)

	// FULL RUN
	ranker.dataSample = ranker.dataAllOriginal
	ranker.makeBuilder()
	ranker.prepareRankings()
	ranker.createWeightColumns()
	ranker.doAlgos()
	ranker.setupInitialSolutionFromPreviousWeightOnly(solution1)
	solution2 := ranker.build.RunHighs(ranker.printer, stopwatch)
	weights2 := ranker.extractAndReportSolution(solution2)

	return weights2
}

//func (ranker *RankingStatWeightProcess3) RunUsingExternalStart(initialWeight WeightResult) util.Optional[WeightResult] {
//	ranker.build = new(utilhighs.LinearBuilder)
//	ranker.build.Minimise = true
//	ranker.build.TimeLimitSeconds = 5000
//	ranker.build.Solver = utilhighs.Solver_MIP_Interior
//
//	ranker.printer.Println("RankingStatWeightProcess3 FIRST ROUND")
//	ranker.dataSample = ranker.dataAllOriginal
//	ranker.prepareRankings()
//	ranker.createWeightColumns()
//	ranker.makeDataListEntryColumns()
//	ranker.makeDataListPairRules()
//	ranker.setupInitialSolutionFromExternal2(initialWeight)
//	solution := ranker.build.RunHighs(ranker.printer)
//	if solution.HasSolution() {
//		weights := ranker.extractAndReportSolution(solution)
//		return util.Optional_OfValue(weights)
//	} else {
//		return util.Optional_Empty[WeightResult]()
//	}
//}

func (ranker *RankingStatWeightProcess3) createWeightColumns() {
	lo := -c_Rank3_LargeWeight
	hi := c_Rank3_LargeWeight

	sumWeights := utilhighs.ConstraintRow{Debug: "sumWeights"}
	ranker.weightColumns = make(map[stats.StatType]utilhighs.ColumnIndex)
	for _, statType := range ranker.requiredStats {
		colWeight := ranker.build.CreateColumnGeneral(highs.Continuous, lo, hi, utilhighs.DebugString{Text: "WEIGHT " + statType.Name()})
		ranker.weightColumns[statType] = colWeight
		sumWeights.Add(colWeight, 1)
	}

	sumWeights.Build(ranker.build, c_rank3_min_total_weight, utilhighs.C_PlusInf)
}

func (ranker *RankingStatWeightProcess3) prepareRankings() {
	// reset values
	for i := range ranker.dataSample {
		ranker.dataSample[i].simScore = 0
		ranker.dataSample[i].targetRank = 0
	}

	// score each sim
	for _, simType := range ranker.requiredSims {
		for entry, simDetailRank := range util.CalculateRanking(simType.IsHighGood(), ranker.dataSample, func(x *rankEntry3) float64 { return x.data.SimResult.Get(simType) }) {
			entry.simScore += float64(simDetailRank) * ranker.targetRatios.Get(simType)
		}
	}

	// TODO ranking ranges
	// TODO alternately deny duplicates, either on simScore, or full detail

	// rank combined sims
	for entry, simRank := range util.CalculateRanking(true, ranker.dataSample, func(x *rankEntry3) float64 { return x.simScore }) {
		entry.targetRank = simRank
	}
}

func (ranker *RankingStatWeightProcess3) doAlgos() {
	switch ranker.ALGO {
	case 0:
		ranker.makeDataListEntryColumns()
		for baseIndex := range ranker.dataSample {
			for compareTo := baseIndex + 1; compareTo < len(ranker.dataSample); compareTo++ {
				ranker.makeEntryPairSequenceConstraints(&ranker.dataSample[baseIndex], &ranker.dataSample[compareTo], baseIndex, compareTo)
			}
		}
	case 1:
		ranker.makeDataListEntryColumns()
		for baseIndex := range ranker.dataSample {
			for compareTo := baseIndex + 1; compareTo < len(ranker.dataSample); compareTo++ {
				ranker.makeEntryPairSequenceConstraintsRequireEqualMode(&ranker.dataSample[baseIndex], &ranker.dataSample[compareTo], baseIndex, compareTo, 1)
			}
		}
	default:
		panic("dunno")
	}
}

func (ranker *RankingStatWeightProcess3) makeDataListEntryColumns() {
	maxRank := float64(len(ranker.dataSample) - 1)

	sumRanks := utilhighs.ConstraintRow{Debug: "sumRanks"}
	for entry := range util.ForPointer(ranker.dataSample) {
		debugStr := strconv.FormatInt(int64(entry.targetRank), 10)
		ranker.makeScoreColumn(entry, debugStr)
		ranker.makeRankColumnAndDiff(entry, maxRank, debugStr)
		sumRanks.Add(entry.rankColumn, 1)
	}

	expectedSum := float64(len(ranker.dataSample)) * float64(len(ranker.dataSample)-1) / 2.0
	sumRanks.Build(ranker.build, expectedSum, expectedSum)
}

func (ranker *RankingStatWeightProcess3) makeDataListEntryColumnsNoMIP() {
	for entry := range util.ForPointer(ranker.dataSample) {
		debugStr := strconv.FormatInt(int64(entry.targetRank), 10)
		ranker.makeScoreColumn(entry, debugStr)
	}
}

func (ranker *RankingStatWeightProcess3) makeRankColumnAndDiff(entry *rankEntry3, maxRank float64, debugStr string) {
	entry.rankColumn = ranker.build.CreateColumnGeneral(highs.Integer, 0, maxRank, utilhighs.DebugText("derivedRank-"+debugStr))
	entry.rankDiffAbsColumn = ranker.build.CreateColumnWithOutput(highs.Integer, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugText("rankDiffAbs-"+debugStr))

	targetRank := float64(entry.targetRank)
	ranker.build.AbsoluteValueFromDiffOneToConst(entry.rankColumn, 1, targetRank, entry.rankDiffAbsColumn, "diffRankToTarget")
}

func (ranker *RankingStatWeightProcess3) makeScoreColumn(entry *rankEntry3, debugStr string) {
	entry.scoreColumn = ranker.build.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugText("score-"+debugStr))

	scoreRow := utilhighs.ConstraintRow{Debug: "scoreRow-" + debugStr}
	for _, statType := range ranker.requiredStats {
		weightColumn := ranker.weightColumns[statType]
		statValue := entry.data.TotalStat.GetFloat(statType)
		statScale := ranker.scaleStats[statType]

		scoreRow.Add(weightColumn, statValue*statScale)
	}
	scoreRow.Add(entry.scoreColumn, -1)
	scoreRow.Build(ranker.build, 0, 0)
}

func (ranker *RankingStatWeightProcess3) makeEntryPairSequenceConstraints(one *rankEntry3, two *rankEntry3, indexOne, indexTwo int) {
	// so we could totally do a boolean thing where scoreA>scoreB then implies rankA>rankB
	// would need all possible pairs connected, but would then force solver to make a full integer order
	isGreaterScore := ranker.build.CreateColumnBool(utilhighs.DebugText("isGreaterScore"))
	isGreaterRank := ranker.build.CreateColumnBool(utilhighs.DebugText("isGreaterRank"))
	isSequenceDiff := ranker.build.CreateColumnBoolWithOutput(1, utilhighs.DebugText("sequenceDiff"))

	ranker.build.ColumnIsGreaterOrEqualColumn(one.scoreColumn, two.scoreColumn, isGreaterScore, c_Rank3_LargeScore, 0.0001)
	ranker.build.ColumnIsGreaterOrEqualColumn(one.rankColumn, two.rankColumn, isGreaterRank, c_Rank3_LargeRank, 1.0)
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
	isGreaterScore := ranker.build.CreateColumnBool(utilhighs.DebugText("isGreaterScore"))
	isGreaterRank := ranker.build.CreateColumnBool(utilhighs.DebugText("isGreaterRank"))
	isSequenceDiff := ranker.build.CreateColumnBoolWithOutput(scaleDiffOutput, utilhighs.DebugText("sequenceDiff"))

	ranker.build.ColumnIsGreaterOrEqualColumn(one.scoreColumn, two.scoreColumn, isGreaterScore, c_Rank3_LargeScore, 0.0001)
	ranker.build.ColumnIsGreaterOrEqualColumn(one.rankColumn, two.rankColumn, isGreaterRank, c_Rank3_LargeRank, 1.0)

	rowEqual := utilhighs.ConstraintRow{Debug: "rowEqual"}
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

// we do a complete ranking just on the strength stat
func (ranker *RankingStatWeightProcess3) setupDumbInitialSolution() {
	internalWeights := WeightResult_Make()
	for statType, colWeight := range ranker.weightColumns {
		ranker.build.SetInitialSolutionValue(colWeight, 1)
		internalWeights.Put(statType, 1)
	}

	ranker.setupInitialRemainingVariables(internalWeights)

	// TODO check if we even use pairs in dumb
	ranker.setupInitialPairsDetail()

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
		internalWeights.Put(statType, weight)
	}

	if !internalWeights.IsEmpty() {
		for statType, colWeight := range ranker.weightColumns {
			weight := internalWeights.Get(statType)
			ranker.build.SetInitialSolutionValue(colWeight, weight)
		}
	}

	//ranker.setupInitialRemainingVariables(internalWeights)

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

	ranker.setupInitialRemainingVariables(internalWeights)

	ranker.build.ValidateInitialSolutionState()
}

func (ranker *RankingStatWeightProcess3) setupInitialRemainingVariables(internalWeights WeightResult) {
	for entry := range util.ForPointer(ranker.dataSample) {
		entry.initialStatScore = internalWeights.CalcStatScoreScaled(entry.data, ranker.scaleStats)
	}

	for entry, calcRank := range util.CalculateRanking(true, ranker.dataSample, func(x *rankEntry3) float64 { return x.initialStatScore }) {
		ranker.build.SetInitialSolutionValue(entry.scoreColumn, entry.initialStatScore)
		if entry.rankColumn != -1 {
			ranker.build.SetInitialSolutionValue(entry.rankColumn, float64(calcRank))
		}
		if entry.rankDiffAbsColumn != -1 {
			diff := float64(calcRank) - float64(entry.targetRank)
			ranker.build.SetInitialSolutionValue(entry.rankDiffAbsColumn, math.Abs(diff))
		}
	}
}

func (ranker *RankingStatWeightProcess3) setupInitialPairsDetail() {
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
}

func (ranker *RankingStatWeightProcess3) extractAndReportSolution(solution *highs.Solution) WeightResult {
	ranker.build.DebugPrintColumns(solution, ranker.printer)

	ranker.printer.Println("WEIGHTS")

	statWeightResult := WeightResult_Make()
	for _, statType := range ranker.requiredStats {
		weightColumn := ranker.weightColumns[statType]
		statScale := ranker.scaleStats[statType]

		modelWeight := solution.ColValues[weightColumn]
		// TODO changed to multiply following analysis on other algorithms, not checked here
		usableWeight := modelWeight * statScale

		statWeightResult.Put(statType, usableWeight)
	}

	divideBy := statWeightResult.Get(stats.Stat_Strength)
	for _, statType := range ranker.requiredStats {
		value := statWeightResult.Get(statType) / divideBy
		statWeightResult.Put(statType, value)
		ranker.printer.Printf("%10s %f\n", statType.Name(), value)
	}

	ranker.reportRankingOfInputs(statWeightResult, solution)

	return statWeightResult
}

func (ranker *RankingStatWeightProcess3) reportRankingOfInputs(statWeightResult WeightResult, solution *highs.Solution) {
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
