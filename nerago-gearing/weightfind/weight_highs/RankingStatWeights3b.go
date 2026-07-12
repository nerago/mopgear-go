package weight_highs

import (
	"cmp"
	"fmt"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_highs"
	"slices"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_rank3b_scaleTarget       = 10.0
	c_rank3b_min_total_weight  = 0.01
	c_rank3b_targetTotalWeight = 5.0
	c_Rank3b_largeWeight       = 10.0
)

type RankingStatWeightProcess3b struct {
	printer        *util.PrintRecorder
	timeoutSeconds int

	targetRatios    stats.SimData
	requiredStats   []stats.StatType
	requiredSims    []stats.SimType
	dataAllOriginal []rankEntry3b
	dataSample      []rankEntry3b
	ALGO            int
	TOTALWEIGHT     int

	build *util_highs.LinearBuilder

	weightColumns map[stats.StatType]util_highs.ColumnIndex
	pairLinks     util.MapMapDiagonal[int, rankPair3b]
}

type rankEntry3b struct {
	data *WeightInput

	initialStatScore float64
	simScore         float64
	targetRank       int
	scoreColumn      util_highs.ColumnIndex
}

type rankPair3b struct {
	indexOne, indexTwo int
	scoreSlackColumn   util_highs.ColumnIndex
}

func (ranker *RankingStatWeightProcess3b) Init(printer *util.PrintRecorder, timeoutSeconds int) {
	ranker.printer = printer
	ranker.timeoutSeconds = timeoutSeconds
}

func (ranker *RankingStatWeightProcess3b) SupplyData(inputData []WeightInput) {
	ranker.dataAllOriginal = util.MapSliceAsNew(inputData, func(input *WeightInput) rankEntry3b {
		return rankEntry3b{
			data:        input,
			simScore:    -1,
			targetRank:  -1,
			scoreColumn: -1,
		}
	})
}

func (ranker *RankingStatWeightProcess3b) SetRequiredStats(requiredStats []stats.StatType) {
	ranker.requiredStats = requiredStats
}

func (ranker *RankingStatWeightProcess3b) SetTargetRatios(targetRatios stats.SimData) {
	ranker.targetRatios = targetRatios
	ranker.requiredSims = targetRatios.NonZeroTypes()
}

func (ranker *RankingStatWeightProcess3b) newBuilder() {
	ranker.build = new(util_highs.LinearBuilder)
	ranker.build.Minimise = true
	ranker.build.TimeLimitSeconds = ranker.timeoutSeconds
	ranker.build.Solver = util_highs.Solver_Force_IPX

	// IPX Duration 1m1.4189406s		92.046912%
	// simplex Duration 3m45.3631097s	92.047609%
	// hipdlp stopped at 10m, stuck
	// hipo fails, reverts to IPX		92.047145%
}

func (ranker *RankingStatWeightProcess3b) RunMultiRound(stopwatch *util.Stopwatch) *util_async.FutureCancellable[WeightResult] {

	// FIRST ROUND: minimal data, no initial values
	ranker.dataSample = takeDataSample_Start(ranker.dataAllOriginal, 12)
	ranker.newBuilder()
	ranker.prepareRankings()
	ranker.createWeightColumns()
	ranker.doAlgos()
	ranker.setupDumbInitialSolution()

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
		}
		return ranker.build.RunHighsFuture(stopwatch)
	})

	return util_async.FutureCancellable_MapValue(solution2Future, func(linearResult2 util_highs.LinearResult) (WeightResult, bool) {
		solution2 := linearResult2.GetSolutionAndSaveLog(ranker.printer)
		return ranker.extractAndReportSolution(solution2), true
	})
}

func (ranker *RankingStatWeightProcess3b) RunSinglePassFromExternal(initial WeightResult, stopwatch *util.Stopwatch) *util_async.FutureCancellable[WeightResult] {
	// FULL RUN
	ranker.dataSample = ranker.dataAllOriginal
	ranker.newBuilder()
	ranker.prepareRankings()
	ranker.createWeightColumns()
	ranker.doAlgos()
	ranker.setupInitialSolutionFromExternal(initial)
	solutionFuture := ranker.build.RunHighsFuture(stopwatch)

	return util_async.FutureCancellable_MapValue(solutionFuture, func(linearResult2 util_highs.LinearResult) (WeightResult, bool) {
		solution2 := linearResult2.GetSolutionAndSaveLog(ranker.printer)
		return ranker.extractAndReportSolution(solution2), true
	})
}

func (ranker *RankingStatWeightProcess3b) createWeightColumns() {
	lo := -c_Rank3b_largeWeight
	hi := c_Rank3b_largeWeight

	sumWeights := util_highs.ConstraintRow{Debug: "sumWeights"}
	ranker.weightColumns = make(map[stats.StatType]util_highs.ColumnIndex)
	for _, statType := range ranker.requiredStats {
		colWeight := ranker.build.CreateColumnGeneral(highs.Continuous, lo, hi, util_highs.DebugString{Text: "WEIGHT " + statType.Name()})
		ranker.weightColumns[statType] = colWeight
		sumWeights.Add(colWeight, 1)
	}

	if ranker.TOTALWEIGHT == 0 {
		sumWeights.Build(ranker.build, c_rank3b_min_total_weight, util_highs.C_PlusInf)
	} else if ranker.TOTALWEIGHT == 1 {
		maxWeight := c_Rank3b_largeWeight * float64(len(ranker.requiredStats))
		sumWeightCol := ranker.build.CreateColumnWithOutput(highs.Continuous, c_rank3b_min_total_weight, maxWeight, 1, util_highs.DebugText("sumWeightCol"))
		sumWeights.Add(sumWeightCol, -1)
		sumWeights.Build(ranker.build, 0, 0)
	} else {
		sumWeights.Build(ranker.build, c_rank3b_targetTotalWeight, c_rank3b_targetTotalWeight)
	}
}

func (ranker *RankingStatWeightProcess3b) prepareRankings() {
	// reset values
	for i := range ranker.dataSample {
		ranker.dataSample[i].simScore = 0
		ranker.dataSample[i].targetRank = 0
	}

	// score each sim
	for _, simType := range ranker.requiredSims {
		for entry, simDetailRank := range util.CalculateRanking(simType.IsHighGood(), ranker.dataSample, func(x *rankEntry3b) float64 { return x.data.SimResult.Get(simType) }) {
			entry.simScore += float64(simDetailRank) * ranker.targetRatios.Get(simType)
		}
	}

	// TODO ranking ranges
	// TODO alternately deny duplicates, either on simScore, or full detail

	// rank combined sims
	for entry, simRank := range util.CalculateRanking(true, ranker.dataSample, func(x *rankEntry3b) float64 { return x.simScore }) {
		entry.targetRank = simRank
	}

	slices.SortFunc(ranker.dataSample, func(a, b rankEntry3b) int { return cmp.Compare(a.targetRank, b.targetRank) })
}

func (ranker *RankingStatWeightProcess3b) doAlgos() {
	if ranker.ALGO == 0 {
		ranker.makeDataListEntryColumns()
		for baseIndex := range ranker.dataSample {
			for compareTo := baseIndex + 1; compareTo < len(ranker.dataSample); compareTo++ {
				ranker.makeEntryPairCheckScoreOrderMatchesTargetOrderWithSlackVar(&ranker.dataSample[baseIndex], &ranker.dataSample[compareTo], baseIndex, compareTo)
			}
		}
	} else {
		ranker.makeDataListEntryColumns()
		for baseIndex := 0; baseIndex < len(ranker.dataSample)-1; baseIndex++ {
			ranker.makeEntryPairCheckScoreOrderMatchesTargetOrderWithSlackVar(&ranker.dataSample[baseIndex], &ranker.dataSample[baseIndex+1], baseIndex, baseIndex+1)
		}
	}
}

func (ranker *RankingStatWeightProcess3b) makeDataListEntryColumns() {
	for entry := range util.ForPointer(ranker.dataSample) {
		debugStr := strconv.FormatInt(int64(entry.targetRank), 10)
		ranker.makeScoreColumn(entry, debugStr)
	}
}

func (ranker *RankingStatWeightProcess3b) makeScoreColumn(entry *rankEntry3b, debugStr string) {
	entry.scoreColumn = ranker.build.CreateColumnGeneral(highs.Continuous, util_highs.C_MinusInf, util_highs.C_PlusInf, util_highs.DebugText("score-"+debugStr))

	scoreRow := util_highs.ConstraintRow{Debug: "scoreRow-" + debugStr}
	for _, statType := range ranker.requiredStats {
		weightColumn := ranker.weightColumns[statType]
		statValue := entry.data.TotalStat.GetFloat(statType)
		statScale := c_rank3b_scaleTarget

		scoreRow.Add(weightColumn, statValue*statScale)
	}
	scoreRow.Add(entry.scoreColumn, -1)
	scoreRow.Build(ranker.build, 0, 0)
}

func (ranker *RankingStatWeightProcess3b) makeEntryPairCheckScoreOrderMatchesTargetOrderWithSlackVar(one *rankEntry3b, two *rankEntry3b, indexOne, indexTwo int) {
	debug := fmt.Sprintf("-%d-%d", indexOne, indexTwo)
	slack := ranker.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.C_PlusInf, 1, util_highs.DebugText("slack"+debug))

	if one.targetRank > two.targetRank {
		row := util_highs.ConstraintRow{Debug: "row" + debug}
		row.Add(one.scoreColumn, 1)
		row.Add(two.scoreColumn, -1)
		row.Add(slack, 1)
		row.Build(ranker.build, 0, util_highs.C_PlusInf)
	} else if two.targetRank > one.targetRank {
		row := util_highs.ConstraintRow{Debug: "row" + debug}
		row.Add(two.scoreColumn, 1)
		row.Add(one.scoreColumn, -1)
		row.Add(slack, 1)
		row.Build(ranker.build, 0, util_highs.C_PlusInf)
	} else {
		// TODO overlapping ranks may be permitted
		panic("unexpected equal ranks")
	}

	ranker.pairLinks.Put(indexOne, indexTwo, rankPair3b{
		indexOne:         indexOne,
		indexTwo:         indexTwo,
		scoreSlackColumn: slack,
	})
}

// we do a complete ranking using all weights set equal
func (ranker *RankingStatWeightProcess3b) setupDumbInitialSolution() {
	internalWeights := WeightResult_Make()
	for _, statType := range ranker.requiredStats {
		internalWeights.Put(statType, 1)
	}
	ranker.setupInitialFromInternalWeights(internalWeights)
}

func (ranker *RankingStatWeightProcess3b) setupInitialSolutionFromPreviousSolutionWeights(solution *highs.Solution) {
	internalWeights := WeightResult_Make()
	for statType, colWeight := range ranker.weightColumns {
		weight := solution.ColValues[colWeight]
		internalWeights.Put(statType, weight)
	}
	ranker.setupInitialFromInternalWeights(internalWeights)
}

func (ranker *RankingStatWeightProcess3b) setupInitialSolutionFromExternal(weights WeightResult) {
	internalWeights := WeightResult_Make()
	for _, statType := range ranker.requiredStats {
		value := weights.Get(statType)
		internalWeights.Put(statType, value)
	}
	ranker.setupInitialFromInternalWeights(internalWeights)
}

func (ranker *RankingStatWeightProcess3b) setupInitialFromInternalWeights(internalWeights WeightResult) {
	if !internalWeights.IsEmpty() {
		internalWeights = internalWeights.ScaleForTotalSum(c_rank3b_targetTotalWeight)

		for statType, colWeight := range ranker.weightColumns {
			weight := internalWeights.Get(statType)
			ranker.build.SetInitialSolutionValue(colWeight, weight)
		}

		ranker.setupInitialRemainingVariables(internalWeights)
		ranker.setupInitialPairsDetail()
	}

	ranker.build.ValidateInitialSolutionState()
}

func (ranker *RankingStatWeightProcess3b) setupInitialRemainingVariables(internalWeights WeightResult) {
	for entry := range util.ForPointer(ranker.dataSample) {
		entry.initialStatScore = internalWeights.CalcStatScore(entry.data) * c_rank3b_scaleTarget
		ranker.build.SetInitialSolutionValue(entry.scoreColumn, entry.initialStatScore)
	}
}

func (ranker *RankingStatWeightProcess3b) setupInitialPairsDetail() {
	for pair := range ranker.pairLinks.SeqValues() {
		one, two := &ranker.dataSample[pair.indexOne], &ranker.dataSample[pair.indexTwo]

		var slack float64 = 0
		if one.targetRank > two.targetRank {
			if one.initialStatScore > two.initialStatScore {
				slack = 0
			} else {
				slack = two.initialStatScore - one.initialStatScore
			}
		} else if two.targetRank > one.targetRank {
			if two.initialStatScore > one.initialStatScore {
				slack = 0
			} else {
				slack = one.initialStatScore - two.initialStatScore
			}
		} else {
			panic("unexpected equal ranks")
		}

		ranker.build.SetInitialSolutionValue(pair.scoreSlackColumn, slack)
	}
}

/*
		 ACTUAL RUNTIME:
		 score = sum(stat1*scale1*weight1)
		 score = sum(statB*scaleB*weightB + stat1*scale1*weight1 + stat2*scale2*weight2)

		 WHAT WE WISH IT WERE
		 pretendScore = sum(statB*usableWeightB + stat1*usableWeight1 + stat2*usableWeight2)
		 finalWeight1 = usableWeight1/usableWeightB

		 IS THERE an interesting ratio between score and pretendScore, common divisor that comes out of all the weights
		 that could mean that messing with scales is wrong?
		 no, if we're doing it right the score==pretendScore
		 the only factor in the finalWeight will be rescaling by strength/etc

	     THEN WE CAN EQUATE THEM
	     sum(statB*scaleB*weightB + stat1*scale1*weight1 + stat2*scale2*weight2) = sum(statB*usableWeightB + stat1*usableWeight1 + stat2*usableWeight2)
	     and be sure that usableWeightB=scaleB*weightB, usableWeight1=scale1*weight1
	     finalWeight1=(scale1*weight1)/(scaleB*weightB)

						   worked example
						   statB= 110 stat1=50 stat2=180
						   scale= 0.1 0.2 0.05
					       scaledStats= 11 10 9
						   runtimeWeights=0.16 0.69 0.2
					       runtimeScore=10.46
				           usableWeight=runtimeWeight*scale
				           usableWeight=0.016 0.138 0.01
			               finalWeights=1 8.625 0.625
*/
func (ranker *RankingStatWeightProcess3b) extractAndReportSolution(solution *highs.Solution) WeightResult {
	ranker.build.DebugPrintColumns(solution, ranker.printer)

	ranker.printer.Println("WEIGHTS")

	statWeightResult := WeightResult_Make()
	for _, statType := range ranker.requiredStats {
		weightColumn := ranker.weightColumns[statType]

		modelWeight := solution.ColValues[weightColumn]
		statWeightResult.Put(statType, modelWeight)
	}

	baseStat := ranker.requiredStats[0]
	divideBy := statWeightResult.Get(baseStat)
	for _, statType := range ranker.requiredStats {
		value := statWeightResult.Get(statType) / divideBy
		statWeightResult.Put(statType, value)
		ranker.printer.Printf("%10s %f\n", statType.Name(), value)
	}

	return statWeightResult
}
