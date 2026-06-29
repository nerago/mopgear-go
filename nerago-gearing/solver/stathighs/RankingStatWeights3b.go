package stathighs

import (
	"fmt"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_rank3b_scaleTarget         = 10.0
	c_rank3b_initial_data_sample = 12
	c_rank3b_min_total_weight    = 1.0
	c_rank3b_time_limit          = 5000

	c_Rank3b_LargeWeight = 10.0
	c_Rank3b_LargeScore  = 500.0
	c_Rank3b_LargeRank   = 10000.0
)

type RankingStatWeightProcess3b struct {
	printer *util.PrintRecorder

	targetRatios    stats.SimData
	requiredStats   []stats.StatType
	requiredSims    []stats.SimType
	dataAllOriginal []rankEntry3b
	dataSample      []rankEntry3b
	SCALE1          bool
	ALGO            int

	build *utilhighs.LinearBuilder

	scaleStats    map[stats.StatType]float64
	weightColumns map[stats.StatType]utilhighs.ColumnIndex
	pairLinks     util.MapMapDiagonal[int, rankPair3b]
}

type rankEntry3b struct {
	data *WeightInput

	initialStatScore float64
	simScore         float64
	targetRank       int
	scoreColumn      utilhighs.ColumnIndex
}

type rankPair3b struct {
	indexOne, indexTwo int
	scoreSlackColumn   utilhighs.ColumnIndex
}

func (ranker *RankingStatWeightProcess3b) Init(printer *util.PrintRecorder) {
	ranker.printer = printer
}

func (ranker *RankingStatWeightProcess3b) SupplyData(inputData []WeightInput) {
	if ranker.SCALE1 {
		ranker.scaleStats = make(map[stats.StatType]float64)
		for _, statType := range stats.StatType_List {
			ranker.scaleStats[statType] = c_rank3b_scaleTarget
		}
	} else {
		ranker.scaleStats = chooseStatScaling(inputData, c_rank3b_scaleTarget, ranker.printer)
	}
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
	ranker.build = new(utilhighs.LinearBuilder)
	ranker.build.Minimise = true
	ranker.build.TimeLimitSeconds = c_rank3b_time_limit
	ranker.build.Solver = utilhighs.Solver_LP_NO_GPU
}

func (ranker *RankingStatWeightProcess3b) Run(optionalInitial *WeightResult, stopwatch *util.Stopwatch) WeightResult {

	// FIRST ROUND: minimal data, no initial values
	ranker.dataSample = takeDataSample_Start(ranker.dataAllOriginal, 12)
	ranker.newBuilder()
	ranker.prepareRankings()
	ranker.createWeightColumns()
	ranker.doAlgos()
	if optionalInitial != nil {
		ranker.setupInitialSolutionFromExternal(*optionalInitial)
	} else {
		ranker.setupDumbInitialSolution()
	}
	solution1 := ranker.build.RunHighs(ranker.printer, stopwatch)
	_ = ranker.extractAndReportSolution(solution1)

	// FULL RUN
	ranker.dataSample = ranker.dataAllOriginal
	ranker.newBuilder()
	ranker.prepareRankings()
	ranker.createWeightColumns()
	ranker.doAlgos()
	ranker.setupInitialSolutionFromPreviousWeightOnly(solution1)
	solution2 := ranker.build.RunHighs(ranker.printer, stopwatch)
	weights2 := ranker.extractAndReportSolution(solution2)

	return weights2
}

func (ranker *RankingStatWeightProcess3b) RunSinglePassFromExternal(initial WeightResult, stopwatch *util.Stopwatch) WeightResult {
	// FULL RUN
	ranker.dataSample = ranker.dataAllOriginal
	ranker.newBuilder()
	ranker.prepareRankings()
	ranker.createWeightColumns()
	ranker.doAlgos()
	ranker.setupInitialSolutionFromExternal(initial)
	solution2 := ranker.build.RunHighs(ranker.printer, stopwatch)
	weights2 := ranker.extractAndReportSolution(solution2)

	return weights2
}

func (ranker *RankingStatWeightProcess3b) createWeightColumns() {
	lo := -c_Rank3b_LargeWeight
	hi := c_Rank3b_LargeWeight

	sumWeights := utilhighs.ConstraintRow{Debug: "sumWeights"}
	ranker.weightColumns = make(map[stats.StatType]utilhighs.ColumnIndex)
	for _, statType := range ranker.requiredStats {
		colWeight := ranker.build.CreateColumnGeneral(highs.Continuous, lo, hi, utilhighs.DebugString{Text: "WEIGHT " + statType.Name()})
		ranker.weightColumns[statType] = colWeight
		sumWeights.Add(colWeight, 1)
	}

	sumWeights.Build(ranker.build, c_rank3b_min_total_weight, utilhighs.C_PlusInf)
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

func (ranker *RankingStatWeightProcess3b) makeEntryPairCheckScoreOrderMatchesTargetOrderWithSlackVar(one *rankEntry3b, two *rankEntry3b, indexOne, indexTwo int) {
	debug := fmt.Sprintf("-%d-%d", indexOne, indexTwo)
	slack := ranker.build.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, 1, utilhighs.DebugText("slack"+debug))

	if one.targetRank > two.targetRank {
		row := utilhighs.ConstraintRow{Debug: "row" + debug}
		row.Add(one.scoreColumn, 1)
		row.Add(two.scoreColumn, -1)
		row.Add(slack, 1)
		row.Build(ranker.build, 0, utilhighs.C_PlusInf)
	} else if two.targetRank > one.targetRank {
		row := utilhighs.ConstraintRow{Debug: "row" + debug}
		row.Add(two.scoreColumn, 1)
		row.Add(one.scoreColumn, -1)
		row.Add(slack, 1)
		row.Build(ranker.build, 0, utilhighs.C_PlusInf)
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
	for statType, colWeight := range ranker.weightColumns {
		ranker.build.SetInitialSolutionValue(colWeight, 1)
		internalWeights.Put(statType, 1)
	}

	ranker.setupInitialRemainingVariables(internalWeights)

	ranker.setupInitialPairsDetail()

	ranker.build.ValidateInitialSolutionState()
}

// data []rankEntry3b, weights map[stats.StatType]float64
func (ranker *RankingStatWeightProcess3b) setupInitialSolutionFromPreviousWeightOnly(solution *highs.Solution) {
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

		ranker.setupInitialRemainingVariables(internalWeights)
	}

	ranker.build.ValidateInitialSolutionState()
}

func (ranker *RankingStatWeightProcess3b) setupInitialSolutionFromExternal(weights WeightResult) {
	internalWeights := WeightResult_Make()
	for statType, colWeight := range ranker.weightColumns {
		basicValue := weights.Get(statType)
		//scale := ranker.scaleStats[statType]
		//scaledValue := basicValue * scale
		ranker.build.SetInitialSolutionValue(colWeight, basicValue)
		internalWeights.Put(statType, basicValue)
	}

	ranker.setupInitialRemainingVariables(internalWeights)

	ranker.build.ValidateInitialSolutionState()
}

func (ranker *RankingStatWeightProcess3b) setupInitialRemainingVariables(internalWeights WeightResult) {
	for entry := range util.ForPointer(ranker.dataSample) {
		entry.initialStatScore = internalWeights.CalcStatScoreScaled(entry.data, ranker.scaleStats)
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
		statScale := ranker.scaleStats[statType]

		modelWeight := solution.ColValues[weightColumn]
		usableWeight := modelWeight * statScale

		statWeightResult.Put(statType, usableWeight)
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
