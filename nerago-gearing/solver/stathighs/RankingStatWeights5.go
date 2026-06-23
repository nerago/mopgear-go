package stathighs

import (
	"cmp"
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
	c_rank5_sample    = 60
	c_rank5_timeLimit = 2000

	c_rank5_computeScoreM = 100
)

var (
	c_rank5_weightLo = -10.0
	c_rank5_weightHi = 10.0

	// last last commit
	// c_rank5_computeScoreLo = utilhighs.C_MinusInf
	// c_rank5_computeScoreHi = utilhighs.C_PlusInf
	// c_rank5_weightTotalMin = 0.1
	// c_rank5_weightTotalMax = utilhighs.C_PlusInf

	// last commit
	// c_rank5_computeScoreLo = utilhighs.C_MinusInf
	// c_rank5_computeScoreHi = 800.0
	// c_rank5_weightTotalMin = 0.1
	// c_rank5_weightTotalMax = 800.0

	// thought this was good
	c_rank5_computeScoreLo = 0.0
	// c_rank5_computeScoreLo = 0.1
	c_rank5_computeScoreHi = 80.0
	// c_rank5_computeScoreHi = utilhighs.C_PlusInf
	c_rank5_weightTotalMin = 0.1
	c_rank5_weightTotalMax = 80.0
	// c_rank5_weightTotalMax = utilhighs.C_PlusInf
)

type RankingStatWeightProcess5 struct {
	printer *util.PrintRecorder

	targetRatios   stats.SimData
	requiredSims []stats.SimType
	initialWeights *WeightResult
	dataAll        []WeightInput
}

type rankInternalRun5 struct {
	process *RankingStatWeightProcess5

	build *utilhighs.LinearBuilder

	runData []rankEntry5
	// scaleStats float64
	scaleStats map[stats.StatType]float64

	weightColumns map[stats.StatType]utilhighs.ColumnIndex
	pairLinks     util.MapMapDiagonal[int, *rankPair5]

	objectiveInclude utilhighs.ObjectiveIndex
	objectiveWeight  utilhighs.ObjectiveIndex
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

func (process *RankingStatWeightProcess5) SetTargetRatios(targetRatios stats.SimData) {
	process.targetRatios = targetRatios
	process.requiredSims = targetRatios.NonZeroTypes()
}

func (process *RankingStatWeightProcess5) Run() []WeightResult {
	weightResultList := make([]WeightResult, 0)
	process.printer.Printf("RankingStatWeightProcess5 RunOptimisitic\n")
	run := rankInternalRun5_create(process)
	run.build.TimeLimitSeconds = c_rank5_timeLimit
	run.supplyData(takeDataSample_Random(process.dataAll, c_rank5_sample))
	// run.supplyData(takeDataSample_Start(process.dataAll, c_rank5_sample))
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
	run.build = new(utilhighs.LinearBuilder)
	run.build.Solver = utilhighs.Solver_MIP_Interior

	run.build.BlendMultiObjectives = false
	run.objectiveInclude = run.build.AddObjectivePrioritised(true, -1, 0.05, 2)
	run.objectiveWeight = run.build.AddObjectivePrioritised(true, -1, -1, 1)

	return run
}

func (run *rankInternalRun5) run() (util.Optional[WeightResult], *highs.Solution) {
	solution := run.build.RunHighs(run.process.printer)
	if solution.HasSolution() {
		weights := run.extractAndReportSolution(solution)
		return util.Optional_OfValue(weights), solution
	} else {
		return util.Optional_Empty[WeightResult](), solution
	}
}

func (run *rankInternalRun5) createWeightColumns() {
	sumWeights := utilhighs.ConstraintRow{Debug: "sumWeights"}
	run.weightColumns = make(map[stats.StatType]utilhighs.ColumnIndex)
	for _, statType := range G_RequiredStats {
		var colWeight utilhighs.ColumnIndex
		colWeight = run.build.CreateColumnGeneral(highs.Continuous, c_rank5_weightLo, c_rank5_weightHi, utilhighs.DebugString{Text: "WEIGHT " + statType.Name()})
		run.weightColumns[statType] = colWeight
		sumWeights.Add(colWeight, 1)
	}

	sumWeightCol := run.build.CreateColumnWithObjective(highs.Continuous, c_rank5_weightTotalMin, c_rank5_weightTotalMax, 1, run.objectiveWeight, utilhighs.DebugText("sumWeightCol"))
	sumWeights.Add(sumWeightCol, -1)
	sumWeights.Build(run.build, 0, 0)
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
	for _, simType := range run.process.requiredSims {
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
	entry.scoreCompute = run.build.CreateColumnGeneral(highs.Continuous, c_rank5_computeScoreLo, c_rank5_computeScoreHi, utilhighs.DebugText("scoreCompute-"+rankStr))

	scoreRow := utilhighs.ConstraintRow{Debug: "scoreRow"}
	for statType, weightColumn := range run.weightColumns {
		statValue := entry.data.TotalStat.GetFloat(statType)
		statScale := run.scaleStats[statType]
		// statScale := run.scaleStats
		scoreRow.Add(weightColumn, statValue*statScale)
	}
	scoreRow.Add(entry.scoreCompute, -1)
	scoreRow.Build(run.build, 0, 0)

	// TODO consider varying score, especially low end maybe?
	entry.isInclude = run.build.CreateColumnBoolWithObjective(1, run.objectiveInclude, utilhighs.DebugText("include-"+rankStr))

	entry.scoreIfIncluded = run.build.CreateColumnGeneral(highs.Continuous, c_rank5_computeScoreLo, c_rank5_computeScoreHi, utilhighs.DebugText("scoreIfIncluded-"+rankStr))
	run.build.ContraintIfBoolCopy(entry.isInclude, entry.scoreCompute, 1, entry.scoreIfIncluded, c_rank5_computeScoreM)
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

	compareScore := utilhighs.ConstraintRow{Debug: "compareScore " + indexText}
	compareScore.Add(hi.scoreIfIncluded, 1)
	compareScore.Add(lo.scoreIfIncluded, -1)
	compareScore.Build(run.build, 0, utilhighs.C_PlusInf)

	run.pairLinks.Put(indexLo, indexHi, &rankPair5{entryOne: lo, entryTwo: hi})
}

func (run *rankInternalRun5) extractAndReportSolution(solution *highs.Solution) WeightResult {
	run.build.DebugPrintColumns(solution, run.process.printer)

	run.process.printer.Println("WEIGHTS")

	statWeightResult := WeightResult_Make()
	for _, statType := range G_RequiredStats {
		weightColumn := run.weightColumns[statType]
		statScale := run.scaleStats[statType]
		// statScale := run.scaleStats

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
		run.build.SetInitialSolutionValue(colWeight, value)
	}
}

func (run *rankInternalRun5) setupInitialSolutionFromExternal(weights WeightResult) {
	for statType, colWeight := range run.weightColumns {
		value := weights.Get(statType)
		run.build.SetInitialSolutionValue(colWeight, value)
	}

	entryScores := make([]float64, len(run.runData))
	for i, entry := range run.runData {
		score := weights.CalcStatScoreScaled(entry.data, run.scaleStats)
		// score := weights.CalcStatScore(entry.data)
		// score *= run.scaleStats
		entryScores[i] = score
		run.build.SetInitialSolutionValue(entry.scoreCompute, score)
	}

	firstInclude, lastIndex := 0, len(run.runData)-1
	run.build.SetInitialSolutionValue(run.runData[firstInclude].isInclude, 1)
	run.build.SetInitialSolutionValue(run.runData[firstInclude].scoreIfIncluded, entryScores[firstInclude])

	for i := firstInclude + 1; i < lastIndex; i++ {
		entry := &run.runData[i]
		if entryScores[i-1] <= entryScores[i] && entryScores[i] <= entryScores[i+1] {
			run.build.SetInitialSolutionValue(entry.isInclude, 1)
			run.build.SetInitialSolutionValue(entry.scoreIfIncluded, entryScores[i])
		} else {
			run.build.SetInitialSolutionValue(entry.isInclude, 0)
			run.build.SetInitialSolutionValue(entry.scoreIfIncluded, entryScores[i-1])
			entryScores[i] = entryScores[i-1]
		}
	}

	if entryScores[lastIndex-1] <= entryScores[lastIndex] {
		run.build.SetInitialSolutionValue(run.runData[lastIndex].isInclude, 1)
		run.build.SetInitialSolutionValue(run.runData[lastIndex].scoreIfIncluded, entryScores[lastIndex])
	} else {
		run.build.SetInitialSolutionValue(run.runData[lastIndex].isInclude, 0)
		run.build.SetInitialSolutionValue(run.runData[lastIndex].scoreIfIncluded, entryScores[lastIndex-1])
	}

	run.build.ValidateInitialSolutionState()
}
