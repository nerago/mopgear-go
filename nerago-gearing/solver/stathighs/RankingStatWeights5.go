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

type RankingStatWeightProcess5 struct {
	printer *util.PrintRecorder

	targetRatios simulate.SimData
	initialWeights *WeightResult
	dataAll      []WeightInput
}

type rankInternalRun5 struct {
	process *RankingStatWeightProcess5

	input *utilhighs.InputBuilder

	runData            []rankEntry5
	scaleStats         map[stats.StatType]float64
	minimumIncludeRate float64
	includeCountRow    utilhighs.ConstraintRowBuild

	weightColumns map[stats.StatType]utilhighs.ColumnIndex
	pairLinks     util.MapMapDiagonal[int, *rankPair5]
}

type rankEntry5 struct {
	data *WeightInput

	simScore   float64
	targetRank int

	scoreCompute    utilhighs.ColumnIndex
	scoreIfIncluded utilhighs.ColumnIndex
	isInclude       utilhighs.ColumnIndex
}

type rankPair5 struct {
	entryOne *rankEntry5
	entryTwo *rankEntry5
}

func (process *RankingStatWeightProcess5) Init(printer *util.PrintRecorder) {
	process.printer = printer
}

func (process *RankingStatWeightProcess5) SupplyData(inputData []WeightInput) {
	process.dataAll = inputData
}

func (process *RankingStatWeightProcess5) SupplyInitialWeights(initialWeights WeightResult) {
	process.initialWeights = &initialWeights
}

func (process *RankingStatWeightProcess5) SetTargetRatios(targetRatios simulate.SimData) {
	process.targetRatios = targetRatios
}

func (process *RankingStatWeightProcess5) Run() []WeightResult {
	weightResultList := make([]WeightResult, 0)
	process.printer.Printf("RankingStatWeightProcess5 RunOptimisitic\n")
	run := rankInternalRun5_create(process)
	run.minimumIncludeRate = 0.5
	run.input.TimeLimitSeconds = 4000
	run.supplyData(takeDataSample_Start(process.dataAll, 30))
	run.prepareRankings()
	run.createWeightColumns()
	run.makeDataListEntryColumns()
	run.makeDataListPairRules()
	run.finishCounts()
	if process.initialWeights != nil {
		run.setupInitialSolutionFromExternal(*process.initialWeights)
	}
	weights1, _ := run.run()
	weights1.ApplyIfValue(func(w WeightResult) { weightResultList = append(weightResultList, w) })
	return weightResultList
}

func rankInternalRun5_create(process *RankingStatWeightProcess5) *rankInternalRun5 {
	run := new(rankInternalRun5)
	run.process = process
	run.input = new(utilhighs.InputBuilder)
	run.input.Minimise = false
	run.input.Mip_lp_solver = "ipx"
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

func (run *rankInternalRun5) finishCounts() {
	// run.includeCountRow.Finish(run.input, float64(len(run.runData))*run.minimumIncludeRate, utilhighs.C_PlusInf)
}

func (run *rankInternalRun5) supplyData(inputData []WeightInput) {
	run.scaleStats = chooseStatScaling(inputData, run.process.printer)
	run.runData = util.MapSliceAsNew(inputData, func(input *WeightInput) rankEntry5 {
		return rankEntry5{
			data: input,
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
	entry.scoreCompute = run.input.CreateColumnGeneral(highs.Continuous, -c_RankLimitScore, c_RankLimitScore, utilhighs.DebugText("scoreCompute-"+rankStr))

	scoreRow := utilhighs.ConstraintRowBuild{Debug: "scoreRow"}
	for statType, weightColumn := range run.weightColumns {
		statValue := entry.data.TotalStat.GetFloat(statType)
		statScale := run.scaleStats[statType]
		scoreRow.Add(weightColumn, statValue*statScale)
	}
	scoreRow.Add(entry.scoreCompute, -1)
	scoreRow.Finish(run.input, 0, 0)

	// TODO consider varying score, especially low end maybe?
	entry.isInclude = run.input.CreateColumnBoolWithOutput(1, utilhighs.DebugText("include-"+rankStr))
	run.includeCountRow.Add(entry.isInclude, 1)

	entry.scoreIfIncluded = run.input.CreateColumnGeneral(highs.Continuous, -c_RankLimitScore, c_RankLimitScore, utilhighs.DebugText("scoreIfIncluded-"+rankStr))
	utilhighs.ContraintIfBoolCopy(run.input, entry.isInclude, entry.scoreCompute, entry.scoreIfIncluded, c_RankLargeScore)
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
	compareScore.Add(hi.scoreIfIncluded, 1)
	compareScore.Add(lo.scoreIfIncluded, -1)
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

	run.process.printer.Println("startWeight := stathighs.WeightResult{")
	for _, statType := range G_RequiredStats {
		value := statWeightResult.Get(statType)
		run.process.printer.Printf("  stats.%s: %f,\n", statType.EnumName(), value)
	}
	run.process.printer.Println("}")

	// startWeight := stathighs.WeightResult{
	// 	stats.Stat_Strength:  1.000000,
	// 	stats.Stat_Stamina:   0.480505,
	// 	stats.Stat_Crit:      0.646226,
	// 	stats.Stat_Haste:     0.859856,
	// 	stats.Stat_Expertise: 0.667975,
	// 	stats.Stat_Mastery:   1.940581,
	// 	stats.Stat_Dodge:     0.651822,
	// 	stats.Stat_Parry:     0.624330,
	// }

	includeCount := 0
	for entry := range util.ForPointer(run.runData) {
		includeValue := solution.ColValues[entry.isInclude]
		if utilhighs.FloatEqualsOne(includeValue) {
			includeCount++
		}
	}
	run.process.printer.Printf("INCLUDE %d %d %f%%\n", includeCount, len(run.runData), float64(includeCount)/float64(len(run.runData))*100)

	return statWeightResult
}

func (run *rankInternalRun5) setupInitialSolutionFromExternal_DeriveOwnIncludes(weights WeightResult) {
	for statType, colWeight := range run.weightColumns {
		value := weights.Get(statType)
		run.input.SetInitialSolutionValue(colWeight, value)
	}
}

func (run *rankInternalRun5) setupInitialSolutionFromExternal(weights WeightResult) {
	for statType, colWeight := range run.weightColumns {
		value := weights.Get(statType)
		run.input.SetInitialSolutionValue(colWeight, value)
	}

	entryScores := make([]float64, len(run.runData))
	for i, entry := range run.runData {
		score := weights.CalcStatScoreScaled(entry.data, run.scaleStats)
		entryScores[i] = score
		run.input.SetInitialSolutionValue(entry.scoreCompute, score)
	}

	firstInclude, lastInclude := 0, len(run.runData)-1
	run.input.SetInitialSolutionValue(run.runData[firstInclude].isInclude, 1)
	run.input.SetInitialSolutionValue(run.runData[firstInclude].scoreIfIncluded, entryScores[firstInclude])
	run.input.SetInitialSolutionValue(run.runData[lastInclude].isInclude, 1)
	run.input.SetInitialSolutionValue(run.runData[lastInclude].scoreIfIncluded, entryScores[lastInclude])

	for i := firstInclude + 1; i < lastInclude; i++ {
		entry := &run.runData[i]
		if entryScores[i-1] <= entryScores[i] && entryScores[i] <= entryScores[i+1] {
			run.input.SetInitialSolutionValue(entry.isInclude, 1)
			run.input.SetInitialSolutionValue(entry.scoreIfIncluded, entryScores[i])
		} else {
			// run.input.SetInitialSolutionValue(entry.isInclude, 0)
			// run.input.SetInitialSolutionValue(entry.scoreIfIncluded, entryScores[i-1])
		}
	}

	run.input.ValidateInitialSolutionState()
}
