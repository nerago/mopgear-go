package stathighs

import (
	"cmp"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"slices"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
// c_RankLargeWeight = 50.0
// c_RankLargeScore  = 500.0
// c_RankLimitScore    = 200.0
// c_RankInitialSample = 15
// c_RankAddSample     = 5

// c_RankLargeRank   = 10000.0
)

type RankingStatWeightProcess5 struct {
	printer *util.PrintRecorder

	targetRatios simulate.SimData
	dataAll      []WeightInput
}

type rankInternalRun5 struct {
	process *RankingStatWeightProcess5

	input *utilhighs.InputBuilder

	runData    []rankEntry5
	scaleStats map[stats.StatType]float64

	weightColumns map[stats.StatType]utilhighs.ColumnIndex
	pairLinks     util.MapMapDiagonal[int, *rankPair5]
}

type rankEntry5 struct {
	data *WeightInput

	// initialStatScore float64
	simScore   float64
	targetRank int

	scoreColumn      utilhighs.ColumnIndex
	scoreSlack       utilhighs.ColumnIndex
	scoreSlackOutput utilhighs.ColumnIndex
}

type rankPair5 struct {
	entryOne *rankEntry5
	entryTwo *rankEntry5
	// twoIsGreater utilhighs.ColumnIndex
}

func (process *RankingStatWeightProcess5) Init(printer *util.PrintRecorder) {
	process.printer = printer
}

func (process *RankingStatWeightProcess5) SupplyData(inputData []WeightInput) {
	process.dataAll = inputData
}

func (process *RankingStatWeightProcess5) SetTargetRatios(targetRatios simulate.SimData) {
	process.targetRatios = targetRatios
}

func (process *RankingStatWeightProcess5) Run(doRound2 bool) []WeightResult {
	weightResultList := make([]WeightResult, 0)

	// FIRST ROUND: minimal data, dumb initial values
	process.printer.Println("RankingStatWeightProcess5 FIRST ROUND")
	run1 := rankInternalRun5_create(process)
	run1.input.TimeLimitSeconds = 3600
	// run1.supplyData(takeDataSample(process.dataAll, c_RankInitialSample))
	run1.supplyData(process.dataAll)
	run1.prepareRankings()
	run1.createWeightColumns()
	run1.makeDataListEntryColumns()
	run1.makeDataListPairRules()
	weights1, _ := run1.run()
	weights1.ApplyIfValue(func(w WeightResult) { weightResultList = append(weightResultList, w) })

	// latestRun := run1
	// latestSolution := solution1

	// times := make(map[int]time.Duration)
	// for size := c_RankInitialSample + c_RankAddSample; size <= len(process.dataAll); size += c_RankAddSample {
	// 	startTime := time.Now()

	// 	dataSample := takeDataSample(process.dataAll, size)
	// 	process.printer.Println("RankingStatWeightProcess5 SECOND ROUND " + strconv.Itoa(size))
	// 	run2 := rankInternalRun5_create(process)
	// 	run2.input.TimeLimitSeconds = 8000
	// 	run2.supplyData(dataSample)
	// 	run2.prepareRankings()
	// 	run2.createWeightColumns()
	// 	run2.makeDataListEntryColumns()
	// 	run2.makeDataListPairRules()
	// 	run2.setupInitialSolutionFromPrevious(latestRun, latestSolution)
	// 	weights2, solution2 := run2.run()
	// 	weights2.ApplyIfValue(func(w WeightResult) { weightResultList = append(weightResultList, w) })

	// 	timeTaken := time.Since(startTime)
	// 	times[size] = timeTaken

	// 	if solution2.HasSolution() {
	// 		latestRun = run2
	// 		latestSolution = solution2
	// 	}

	// 	if solution2.Status == highs.ModelStatusTimeLimit {
	// 		break
	// 	}
	// }

	// for size, duration := range times {
	// 	process.printer.Printf("%4d %s\n", size, duration)
	// }

	return weightResultList
}

// func (process *RankingStatWeightProcess5) RunUsingExternalStart(initialWeight WeightResult) util.Optional[WeightResult] {
// 	run2 := rankInternalRun5_create(process)
// 	run2.input.TimeLimitSeconds = 2500
// 	run2.supplyData(process.dataAll)
// 	run2.prepareRankings()
// 	run2.createWeightColumns()
// 	run2.makeDataListEntryColumns()
// 	run2.makeDataListPairRules()
// 	run2.setupInitialSolutionFromExternal2(initialWeight)
// 	weights2, _ := run2.run()
// 	return weights2
// }

func rankInternalRun5_create(process *RankingStatWeightProcess5) *rankInternalRun5 {
	run := new(rankInternalRun5)
	run.process = process
	run.input = new(utilhighs.InputBuilder)
	run.input.Minimise = true
	// run.input.Mip_disallow_restart = true
	// run.input.Solver = "ipx"
	// run.input.Mip_lp_solver = "ipx"
	return run
}

func (run *rankInternalRun5) run() (util.Optional[WeightResult], *highs.Solution) {
	solution, log := run.input.RunHighs()
	run.process.printer.AppendOther(log)
	if solution.HasSolution() {
		weights := run.extractAndReportSolution(solution)
		return util.Optional_OfValue(weights), solution
	} else {
		return util.Optional_Empty[WeightResult](), solution
	}
}

func (run *rankInternalRun5) createWeightColumns() {
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

func (run *rankInternalRun5) supplyData(inputData []WeightInput) {
	run.scaleStats = chooseStatScaling(inputData, run.process.printer)
	run.runData = util.MapSliceAsNew(inputData, func(input *WeightInput) rankEntry5 {
		return rankEntry5{
			data:        input,
			scoreColumn: -1,
		}
	})
}

func (run *rankInternalRun5) prepareRankings() {
	// score each sim
	for _, simType := range G_RequiredSims {
		for entry, simDetailRankHiLo := range util.CalculateRankingRanges(simType.IsHighGood(), run.runData, func(x *rankEntry5) float64 { return x.data.SimResult.Get(simType) }) {
			entry.simScore += float64(simDetailRankHiLo.Mid()) * run.process.targetRatios.Get(simType)
		}
	}

	run.runData = util.RemoveDuplicatesFuncNotify(run.runData,
		func(a, b *rankEntry5) bool { return a.simScore == b.simScore },
		func(entry *rankEntry5) { run.process.printer.Println("removing duplicate score") },
	)

	slices.SortFunc(run.runData, func(a, b rankEntry5) int { return cmp.Compare(a.simScore, b.simScore) })

	for rank := range run.runData {
		run.runData[rank].targetRank = rank
	}
}

func (run *rankInternalRun5) makeDataListEntryColumns() {
	for entry := range util.ForPointer(run.runData) {
		run.makeEntryColumnRefs(entry)
	}
}

func (run *rankInternalRun5) makeEntryColumnRefs(entry *rankEntry5) {
	rankStr := strconv.FormatInt(int64(entry.targetRank), 10)
	entry.scoreColumn = run.input.CreateColumnGeneral(highs.Continuous, -c_RankLimitScore, c_RankLimitScore, utilhighs.DebugText("score-"+rankStr))

	entry.scoreSlack = run.input.CreateColumnGeneral(highs.Continuous, -c_RankLimitScore, c_RankLimitScore, utilhighs.DebugText("scoreSlack"))
	entry.scoreSlackOutput = run.input.CreateColumnWithOutput(highs.Continuous, -c_RankLimitScore, c_RankLimitScore, 1, utilhighs.DebugText("scoreSlackOutput"))

	scoreRow := utilhighs.ConstraintRowBuild{Debug: "scoreRow"}
	for _, statType := range G_RequiredStats {
		weightColumn := run.weightColumns[statType]
		statValue := entry.data.TotalStat.GetFloat(statType)
		statScale := run.scaleStats[statType]

		scoreRow.Add(weightColumn, statValue*statScale)
	}
	scoreRow.Add(entry.scoreColumn, -1)
	scoreRow.Add(entry.scoreSlack, 1)
	scoreRow.Finish(run.input, 0, 0)

	utilhighs.AbsoluteValue(run.input, entry.scoreSlack, entry.scoreSlackOutput)
}

func (run *rankInternalRun5) makeDataListPairRules() {
	for a := 0; a < len(run.runData)-1; a++ {
		run.makeEntryPairScoreChecks(&run.runData[a], &run.runData[a+1], a, a+1)
	}

	if run.pairLinks.Size() != len(run.runData)-1 {
		panic("expected exact size")
	}
}

func (run *rankInternalRun5) makeEntryPairScoreChecks(lo *rankEntry5, hi *rankEntry5, indexLo, indexHi int) {
	indexText := strconv.FormatInt(int64(indexLo), 10)

	compareScore := utilhighs.ConstraintRowBuild{Debug: "compareScore " + indexText}
	compareScore.Add(hi.scoreColumn, 1)
	compareScore.Add(lo.scoreColumn, -1)
	compareScore.Finish(run.input, 0, c_RankLargeScore)

	run.pairLinks.Put(indexLo, indexHi, &rankPair5{entryOne: lo, entryTwo: hi})
}

func (run *rankInternalRun5) extractAndReportSolution(solution *highs.Solution) WeightResult {
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

// func (run *rankInternalRun5) setupInitialSolutionDumb1() {
// 	for statType, colWeight := range run.weightColumns {
// 		if statType == stats.Stat_Strength {
// 			run.input.SetInitialSolutionValue(colWeight, 1)
// 		} else {
// 			run.input.SetInitialSolutionValue(colWeight, 0)
// 		}
// 	}

// 	statScale := run.scaleStats[stats.Stat_Strength]
// 	for entry := range util.ForPointer(run.runData) {
// 		entry.initialStatScore = entry.data.TotalStat.GetFloat(stats.Stat_Strength) * statScale
// 	}

// 	run.setupRemainingInitialSolution()
// 	run.input.ValidateInitialSolutionState()
// }

// func (run *rankInternalRun5) setupInitialSolutionDumb2() {
// 	initialDumbValue := 0.2

// 	dumbWeights := WeightResult_Make()
// 	for statType, colWeight := range run.weightColumns {
// 		run.input.SetInitialSolutionValue(colWeight, initialDumbValue)
// 		dumbWeights.Put(statType, initialDumbValue)
// 	}

// 	for entry := range util.ForPointer(run.runData) {
// 		entry.initialStatScore = dumbWeights.CalcStatScoreScaled(entry.data, run.scaleStats)
// 	}

// 	run.setupRemainingInitialSolution()
// 	run.input.ValidateInitialSolutionState()
// }

// func (run *rankInternalRun5) setupInitialSolutionFromPrevious(previous *rankInternalRun5, solution *highs.Solution) {
// 	internalWeights := WeightResult_Make()
// 	for statType, oldWeightCol := range previous.weightColumns {
// 		newWeightCol := run.weightColumns[statType]
// 		oldWeightValue := solution.ColValues[oldWeightCol]
// 		run.input.SetInitialSolutionValue(newWeightCol, oldWeightValue)
// 		internalWeights.Put(statType, oldWeightValue)
// 	}

// 	for entry := range util.ForPointer(run.runData) {
// 		entry.initialStatScore = internalWeights.CalcStatScoreScaled(entry.data, run.scaleStats)
// 	}

// 	run.setupRemainingInitialSolution()
// 	run.input.ValidateInitialSolutionState()
// }

// func (run *rankInternalRun5) setupInitialSolutionFromExternal(weights stats.StatBlock) {
// 	internalWeights := WeightResult_Make()
// 	for statType, colWeight := range run.weightColumns {
// 		basicValue := weights.GetFloat(statType) / 1000.0 // NOTE reverse scale as used in StatRatingsWeights
// 		// scale := run.scaleStats[statType]
// 		scaledValue := basicValue
// 		run.input.SetInitialSolutionValue(colWeight, scaledValue)
// 		internalWeights.Put(statType, scaledValue)
// 	}

// 	for entry := range util.ForPointer(run.runData) {
// 		entry.initialStatScore = internalWeights.CalcStatScoreScaled(entry.data, run.scaleStats)
// 		run.input.SetInitialSolutionValue(entry.scoreColumn, entry.initialStatScore)
// 	}

// 	run.setupRemainingInitialSolution()
// 	run.input.ValidateInitialSolutionState()
// }

// func (run *rankInternalRun5) setupInitialSolutionFromExternal2(weights WeightResult) {
// 	internalWeights := WeightResult_Make()
// 	for statType, colWeight := range run.weightColumns {
// 		basicValue := weights.Get(statType)
// 		// scale := run.scaleStats[statType]
// 		scaledValue := basicValue
// 		run.input.SetInitialSolutionValue(colWeight, scaledValue)
// 		internalWeights.Put(statType, scaledValue)
// 	}

// 	for entry := range util.ForPointer(run.runData) {
// 		entry.initialStatScore = internalWeights.CalcStatScoreScaled(entry.data, run.scaleStats)
// 		run.input.SetInitialSolutionValue(entry.scoreColumn, entry.initialStatScore)
// 	}

// 	run.setupRemainingInitialSolution()
// 	run.input.ValidateInitialSolutionState()
// }

// func (run *rankInternalRun5) setupRemainingInitialSolution() {
// 	// for entry, rankHiLo := range util.CalculateRankingRanges(true, run.runData, func(x *rankEntry5) float64 { return x.initialStatScore }) {
// 	// run.input.SetInitialSolutionValue(entry.rankColumn, float64(rankHiLo.Lo))
// 	// }

// 	for pair := range run.pairLinks.SeqValues() {
// 		entryOne := pair.entryOne
// 		entryTwo := pair.entryTwo
// 		if entryOne.initialStatScore > entryTwo.initialStatScore {
// 			run.input.SetInitialSolutionValue(pair.oneIsGreater, 1)
// 			// run.input.SetInitialSolutionValue(pair.twoIsGreater, 0)
// 		} else if entryOne.initialStatScore < entryTwo.initialStatScore {
// 			run.input.SetInitialSolutionValue(pair.oneIsGreater, 0)
// 			// run.input.SetInitialSolutionValue(pair.twoIsGreater, 1)
// 		} else {
// 			// run.input.SetInitialSolutionValue(pair.oneIsGreater, 0)
// 			// run.input.SetInitialSolutionValue(pair.twoIsGreater, 0)
// 		}
// 	}
// }
