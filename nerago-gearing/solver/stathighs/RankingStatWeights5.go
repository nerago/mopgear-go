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

// stats are all scaled in range 0.0 .. 1.0
// weights are in range 0.01 .. 10
// thus scores are 8*=  0.0 .. 80

const (
	c_rank5_sample    = 40
	c_rank5_timeLimit = 2000

	// c_rank5_computeScoreLo = 0
	// c_rank5_computeScoreHi = 200 
	c_rank5_computeScoreM = 1000
)

var (
	// c_rank5_weightLo = -10.0
	c_rank5_weightHi = 100.0
	c_rank5_weightLo = utilhighs.C_MinusInf
	// c_rank5_weightHi = utilhighs.C_PlusInf

	c_rank5_computeScoreLo = utilhighs.C_MinusInf
	// c_rank5_computeScoreLo = 0.1
	// c_rank5_computeScoreHi = 800.0
	c_rank5_computeScoreHi = utilhighs.C_PlusInf
	
	c_rank5_weightTotalEnable = true
	c_rank5_weightTotalMin = 0.1
	// c_rank5_weightTotalMax = 800.0
	c_rank5_weightTotalMax = utilhighs.C_PlusInf
)


	// c_rank5_sample    = 50
	// c_rank5_timeLimit = 2000
	// c_rank5_weightLo = 0.01
	// c_rank5_weightHi = 100
// c_rank5_computeScoreLo = utilhighs.C_MinusInf
// 	c_rank5_computeScoreHi = utilhighs.C_PlusInf
// INCLUDE 40 49 81.632653%
// ( Pawn: v1: "Protection WoWSims Weights": Class=Paladin,Strength=1.0000000000,Stamina=1.5365445970,CritRating=0.3372094238,HasteRating=0.8105121588,ExpertiseRating=0.4104601046,MasteryRating=2.4732058443,DodgeRating=0.2633575991,ParryRating=0.2660589952, )
// accuracy = 91.225229
// Duration = 9m3.3503159s

	// c_rank5_sample    = 50
// 	c_rank5_timeLimit = 2000
// 	c_rank5_weightLo = 0.01
// 	c_rank5_weightHi = 100
	// c_rank5_computeScoreLo = 0.0
	// c_rank5_computeScoreHi = utilhighs.C_PlusInf
// INCLUDE 37 48 77.083333%
// ( Pawn: v1: "Protection WoWSims Weights": Class=Paladin,Strength=1.0000000000,Stamina=3.6784935269,CritRating=0.4001995576,HasteRating=1.4628763808,ExpertiseRating=0.2619468384,MasteryRating=2.3425042520,DodgeRating=0.3161847349,ParryRating=0.3654891716, )
// accuracy = 88.393488
// Duration = 2m25.6596228s

	// c_rank5_sample    = 100
	// c_rank5_timeLimit = 2000
	// c_rank5_weightLo = utilhighs.C_MinusInf
	// c_rank5_weightHi = utilhighs.C_PlusInf
	// c_rank5_computeScoreLo = utilhighs.C_MinusInf
	// c_rank5_computeScoreHi = 800.0
	// c_rank5_weightTotalMin = 0.1
	// c_rank5_weightTotalMax = utilhighs.C_PlusInf
// 	NCLUDE 99 99 100.000000%
// ( Pawn: v1: "Protection WoWSims Weights": Class=Paladin,Strength=1.0000000000,Stamina=-0.1906947449,CritRating=0.0065645978,HasteRating=-0.2387397481,ExpertiseRating=0.0262143548,MasteryRating=-0.0463181057,DodgeRating=-0.0619316308,ParryRating=-0.0246371537, )
// accuracy = 85.235258
// Duration = 98.8764ms

type RankingStatWeightProcess5 struct {
	printer *util.PrintRecorder

	targetRatios   simulate.SimData
	initialWeights *WeightResult
	dataAll        []WeightInput
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
	run.input.TimeLimitSeconds = c_rank5_timeLimit
	// run.supplyData(takeDataSample_Random(process.dataAll, c_rank5_sample))
	run.supplyData(takeDataSample_Start(process.dataAll, c_rank5_sample))
	// run.supplyData(process.dataAll)
	run.prepareRankings()
	run.createWeightColumns()
	run.makeDataListEntryColumns()
	run.makeDataListPairRules()
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
	sumWeights := utilhighs.ConstraintRowBuild{Debug: "sumWeights"}
	run.weightColumns = make(map[stats.StatType]utilhighs.ColumnIndex)
	for _, statType := range G_RequiredStats {
		var colWeight utilhighs.ColumnIndex
		// if statType == stats.Stat_Strength {
		// 	colWeight = run.input.CreateColumnGeneral(highs.Continuous, strengthMin, hi, utilhighs.DebugString{Text: "WEIGHT " + statType.Name()})
		// } else {
		colWeight = run.input.CreateColumnGeneral(highs.Continuous, c_rank5_weightLo, c_rank5_weightHi, utilhighs.DebugString{Text: "WEIGHT " + statType.Name()})
		// }
		run.weightColumns[statType] = colWeight
		sumWeights.Add(colWeight, 1)
	}

	// force positive and non-zero result
	if c_rank5_weightTotalEnable {
		sumWeights.Finish(run.input, c_rank5_weightTotalMin, c_rank5_weightTotalMax)
	}
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
	entry.scoreCompute = run.input.CreateColumnGeneral(highs.Continuous, c_rank5_computeScoreLo, c_rank5_computeScoreHi, utilhighs.DebugText("scoreCompute-"+rankStr))

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

	entry.scoreIfIncluded = run.input.CreateColumnGeneral(highs.Continuous, c_rank5_computeScoreLo, c_rank5_computeScoreHi, utilhighs.DebugText("scoreIfIncluded-"+rankStr))
	utilhighs.ContraintIfBoolCopy(run.input, entry.isInclude, entry.scoreCompute, entry.scoreIfIncluded, c_rank5_computeScoreM)
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
	compareScore.Finish(run.input, 0, utilhighs.C_PlusInf)

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
		// usableWeight := modelWeight / statScale
		usableWeight := modelWeight * statScale
		// usableWeight := modelWeight

		statWeightResult.Put(statType, usableWeight)
	}

	// div scalee NCLUDE 33 39 84.615385%
// ( Pawn: v1: "Protection WoWSims Weights": Class=Paladin,Strength=1.0000000000,Stamina=6.5092182647,CritRating=0.1841064312,HasteRating=-0.0106438090,ExpertiseRating=0.2037377680,MasteryRating=0.6981649722,DodgeRating=0.1423104180,ParryRating=0.1121172187, )
// accuracy = 87.374284
// Duration = 4.9955978s

// mul scale NCLUDE 33 39 84.615385%
// ( Pawn: v1: "Protection WoWSims Weights": Class=Paladin,Strength=1.0000000000,Stamina=2.3040325598,CritRating=1.5101529595,HasteRating=-0.0224103858,ExpertiseRating=2.7750703535,MasteryRating=2.9091945716,DodgeRating=1.3248140063,ParryRating=1.4968127221, )
// accuracy = 93.040090
// Duration = 4.7671994s

// no scale change
// UDE 33 39 84.615385%
// ( Pawn: v1: "Protection WoWSims Weights": Class=Paladin,Strength=1.0000000000,Stamina=3.8726542346,CritRating=0.5272844318,HasteRating=-0.0154444769,ExpertiseRating=0.7519219639,MasteryRating=1.4251658666,DodgeRating=0.4342059823,ParryRating=0.4096565383, )
// accuracy = 89.380391
// Duration = 4.4455988s

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

	firstInclude, lastIndex := 0, len(run.runData)-1
	run.input.SetInitialSolutionValue(run.runData[firstInclude].isInclude, 1)
	run.input.SetInitialSolutionValue(run.runData[firstInclude].scoreIfIncluded, entryScores[firstInclude])

	for i := firstInclude + 1; i < lastIndex; i++ {
		entry := &run.runData[i]
		if entryScores[i-1] <= entryScores[i] && entryScores[i] <= entryScores[i+1] {
			run.input.SetInitialSolutionValue(entry.isInclude, 1)
			run.input.SetInitialSolutionValue(entry.scoreIfIncluded, entryScores[i])
		} else {
			run.input.SetInitialSolutionValue(entry.isInclude, 0)
			run.input.SetInitialSolutionValue(entry.scoreIfIncluded, entryScores[i-1])
			entryScores[i] = entryScores[i-1]
		}
	}

	if entryScores[lastIndex-1] <= entryScores[lastIndex] {
		run.input.SetInitialSolutionValue(run.runData[lastIndex].isInclude, 1)
		run.input.SetInitialSolutionValue(run.runData[lastIndex].scoreIfIncluded, entryScores[lastIndex])
	} else {
		run.input.SetInitialSolutionValue(run.runData[lastIndex].isInclude, 0)
		run.input.SetInitialSolutionValue(run.runData[lastIndex].scoreIfIncluded, entryScores[lastIndex-1])
	}

	run.input.ValidateInitialSolutionState()
}
