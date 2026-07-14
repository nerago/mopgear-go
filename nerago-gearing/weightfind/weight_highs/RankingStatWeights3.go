package weight_highs

import (
	"cmp"
	"math"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"
	"strconv"

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

	targetRatios    stats.SimData
	requiredStats   []stats.StatType
	requiredSims    []stats.SimType
	dataAllOriginal []rankEntry3
	dataSample      []rankEntry3
	SCALE1          bool
	ALGO            int

	build *util_highs.LinearBuilder

	scaleStats    map[stats.StatType]float64
	weightColumns map[stats.StatType]util_highs.ColumnIndex
	pairLinks     util.MapMap[int, int, rankPair3]
}

type rankEntry3 struct {
	data *weight_types.WeightInput

	initialStatScore float64
	simScore         float64
	targetRank       int
	//primeMultiplier  int64

	scoreColumn       util_highs.ColumnIndex
	rankColumn        util_highs.ColumnIndex
	rankDiffAbsColumn util_highs.ColumnIndex
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
		ranker.scaleStats = make(map[stats.StatType]float64)
		for _, statType := range stats.StatType_List {
			ranker.scaleStats[statType] = 1
		}
	} else {
		ranker.scaleStats = chooseStatScaling(inputData, c_rank3_scaleTarget, false, ranker.printer)
	}
	ranker.dataAllOriginal = util.MapSliceAsNew(inputData, func(input *weight_types.WeightInput) rankEntry3 {
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
	ranker.build = new(util_highs.LinearBuilder)
	ranker.build.Minimise = true
	ranker.build.TimeLimitSeconds = ranker.timeout
	if ranker.ALGO >= 2 {
		ranker.build.Solver = util_highs.Solver_LP_USE_GPU
	} else {
		ranker.build.Solver = util_highs.Solver_MIP_Interior
	}
}

func (ranker *RankingStatWeightProcess3) Run(stopwatch *util.Stopwatch) *util_async.FutureCancellable[weight_types.WeightResult] {
	// FIRST ROUND: minimal data, no initial values
	ranker.dataSample = takeDataSample_Start(ranker.dataAllOriginal, c_rank3_initial_data_sample)
	ranker.prepare()
	ranker.setupDumbInitialSolution()

	solution1Future := ranker.build.RunHighsFuture(stopwatch)

	solution2Future := util_async.FutureCancellable_MapToFuture(solution1Future, func(linearResult1 util_highs.LinearResult) *util_async.FutureCancellable[util_highs.LinearResult] {
		solution1 := linearResult1.GetSolutionAndSaveLog(ranker.printer)
		_ = ranker.extractAndReportSolution(solution1)

		// FULL RUN
		ranker.dataSample = ranker.dataAllOriginal
		ranker.prepare()
		if solution1.HasSolution() {
			ranker.setupInitialSolutionFromPreviousWeightOnly(solution1)
		}
		return ranker.build.RunHighsFuture(stopwatch)
	})

	return util_async.FutureCancellable_MapValue(solution2Future, func(linearResult2 util_highs.LinearResult) (weight_types.WeightResult, bool) {
		solution2 := linearResult2.GetSolutionAndSaveLog(ranker.printer)
		return ranker.extractAndReportSolution(solution2), true
	})
}

func (ranker *RankingStatWeightProcess3) RunUsingExternalStart(initialWeight weight_types.WeightResult, stopwatch *util.Stopwatch) *util_async.FutureCancellable[weight_types.WeightResult] {
	ranker.dataSample = ranker.dataAllOriginal
	ranker.prepare()
	ranker.setupInitialSolutionFromExternal2(initialWeight)
	solutionFuture := ranker.build.RunHighsFuture(stopwatch)
	return util_async.FutureCancellable_MapValue(solutionFuture, func(linearResult util_highs.LinearResult) (weight_types.WeightResult, bool) {
		solution := linearResult.GetSolutionAndSaveLog(ranker.printer)
		return ranker.extractAndReportSolution(solution), true
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
	ranker.weightColumns = make(map[stats.StatType]util_highs.ColumnIndex)
	for _, statType := range ranker.requiredStats {
		colWeight := ranker.build.CreateColumnGeneral(highs.Continuous, lo, hi, util_highs.DebugString{Text: "WEIGHT " + statType.Name()})
		ranker.weightColumns[statType] = colWeight
		sumWeights.Add(colWeight, 1)
	}

	sumWeights.Build(ranker.build, c_rank3_target_total_weight, c_rank3_target_total_weight)
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

	slices.SortFunc(ranker.dataSample, func(a, b rankEntry3) int { return cmp.Compare(a.targetRank, b.targetRank) })
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

	sumRanks := util_highs.ConstraintRow{Debug: "sumRanks"}
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
	entry.rankColumn = ranker.build.CreateColumnGeneral(highs.Integer, 0, maxRank, util_highs.DebugText("derivedRank-"+debugStr))
	entry.rankDiffAbsColumn = ranker.build.CreateColumnWithOutput(highs.Integer, 0, util_highs.C_PlusInf, 1, util_highs.DebugText("rankDiffAbs-"+debugStr))

	targetRank := float64(entry.targetRank)
	ranker.build.AbsoluteValueFromDiffOneToConst(entry.rankColumn, 1, targetRank, entry.rankDiffAbsColumn, "diffRankToTarget")
}

func (ranker *RankingStatWeightProcess3) makeScoreColumn(entry *rankEntry3, debugStr string) {
	entry.scoreColumn = ranker.build.CreateColumnGeneral(highs.Continuous, util_highs.C_MinusInf, util_highs.C_PlusInf, util_highs.DebugText("score-"+debugStr))

	scoreRow := util_highs.ConstraintRow{Debug: "scoreRow-" + debugStr}
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
	isGreaterScore := ranker.build.CreateColumnBool(util_highs.DebugText("isGreaterScore"))
	isGreaterRank := ranker.build.CreateColumnBool(util_highs.DebugText("isGreaterRank"))
	isSequenceDiff := ranker.build.CreateColumnBoolWithOutput(1, util_highs.DebugText("sequenceDiff"))

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
	isGreaterScore := ranker.build.CreateColumnBool(util_highs.DebugText("isGreaterScore"))
	isGreaterRank := ranker.build.CreateColumnBool(util_highs.DebugText("isGreaterRank"))
	isSequenceDiff := ranker.build.CreateColumnBoolWithOutput(scaleDiffOutput, util_highs.DebugText("sequenceDiff"))

	ranker.build.ColumnIsGreaterOrEqualColumn(one.scoreColumn, two.scoreColumn, isGreaterScore, c_Rank3_LargeScore, 0.0001)
	ranker.build.ColumnIsGreaterOrEqualColumn(one.rankColumn, two.rankColumn, isGreaterRank, c_Rank3_LargeRank, 1.0)

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
	internalWeights := weight_types.WeightResult_Make()
	for statType := range ranker.weightColumns {
		internalWeights.Put(statType, 1)
	}
	internalWeights = internalWeights.ScaleForTotalSum(c_rank3_target_total_weight)

	ranker.setupFromInternalWeights(internalWeights)
}

func (ranker *RankingStatWeightProcess3) setupInitialSolutionFromPreviousWeightOnly(solution *highs.Solution) {
	internalWeights := weight_types.WeightResult_Make()
	for statType, colWeight := range ranker.weightColumns {
		weight := solution.ColValues[colWeight]
		internalWeights.Put(statType, weight)
	}

	ranker.setupFromInternalWeights(internalWeights)
}

func (ranker *RankingStatWeightProcess3) setupInitialSolutionFromExternal2(weights weight_types.WeightResult) {
	internalWeights := weights.ScaleForTotalSum(c_rank3_target_total_weight)
	ranker.setupFromInternalWeights(internalWeights)
}

func (ranker *RankingStatWeightProcess3) setupFromInternalWeights(internalWeights weight_types.WeightResult) {
	if !internalWeights.IsEmpty() {
		for statType, colWeight := range ranker.weightColumns {
			weight := internalWeights.Get(statType)
			ranker.build.SetInitialSolutionValue(colWeight, weight)
		}

		ranker.setupInitialRemainingVariables(internalWeights)
	}

	ranker.build.ValidateInitialSolutionState()
}

func (ranker *RankingStatWeightProcess3) setupInitialRemainingVariables(internalWeights weight_types.WeightResult) {
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

func (ranker *RankingStatWeightProcess3) extractAndReportSolution(solution *highs.Solution) weight_types.WeightResult {
	ranker.build.DebugPrintColumns(solution, ranker.printer)

	ranker.printer.Println("WEIGHTS")

	statWeightResult := weight_types.WeightResult_Make()
	for _, statType := range ranker.requiredStats {
		weightColumn := ranker.weightColumns[statType]
		statScale := ranker.scaleStats[statType]

		modelWeight := solution.ColValues[weightColumn]
		usableWeight := modelWeight / statScale

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

func (ranker *RankingStatWeightProcess3) reportRankingOfInputs(statWeightResult weight_types.WeightResult, solution *highs.Solution) {
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
