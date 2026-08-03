package weight_highs

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/simrank"
	"paladin_gearing_go/weightfind/util_weight"
	"paladin_gearing_go/weightfind/weight_types"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
)

// stats are all scaled in range 0.0 .. 1.0
// weights are in range 0.01 .. 10
// thus scores are 8*=  0.0 .. 80

const (
	c_rank5_sample = 60

	c_rank5_computeScoreM = 100
	c_rank5_scaleTarget   = 1.0

	c_rank5_weightLo = -10.0
	c_rank5_weightHi = 10.0

	c_rank5_computeScoreLo    = 0.0
	c_rank5_computeScoreHi    = 80.0
	c_rank5_weightTotalMin    = 0.1
	c_rank5_weightTotalMax    = 80.0
	c_rank5_weightTotalTarget = 10.0
)

type RankingStatWeightProcess5 struct {
	printer   *util.PrintRecorder
	WEIGHTSUM int
	SIMRANK   int

	targetRatios   weight_types.SimPriorityBasic
	requiredStats  []stats.StatType
	requiredSims   []stats.SimType
	initialWeights *weight_types.Weight1Basic
	dataAll        []weight_types.WeightInput
}

type rankInternalRun5 struct {
	process *RankingStatWeightProcess5

	build *util_highs.LinearBuilder

	runData    []*rankEntry5
	scaleStats util_collection.EnumMap[stats.StatType, float64]

	weightColumns map[stats.StatType]util_highs.ColumnIndex
	pairLinks     util_collection.MapMapDiagonal[int, *rankPair5]

	objectiveInclude util_highs.ObjectiveIndex
	objectiveWeight  util_highs.ObjectiveIndex
}

type rankEntry5 struct {
	weight_types.RankStatWeightsCommon

	ScoreCompute    util_highs.ColumnIndex
	ScoreIfIncluded util_highs.ColumnIndex
	IsInclude       util_highs.ColumnIndex
}

type rankPair5 struct {
	entryOne *rankEntry5
	entryTwo *rankEntry5
}

func (process *RankingStatWeightProcess5) Init(printer *util.PrintRecorder) {
	process.printer = printer
}

func (process *RankingStatWeightProcess5) SupplyData(inputData []weight_types.WeightInput) {
	process.dataAll = inputData
}

func (process *RankingStatWeightProcess5) SupplyInitialWeights(initialWeights weight_types.Weight1Basic) {
	process.initialWeights = &initialWeights
}

func (process *RankingStatWeightProcess5) SetRequiredStats(requiredStats []stats.StatType) {
	process.requiredStats = requiredStats
}

func (process *RankingStatWeightProcess5) SetTargetRatios(targetRatios weight_types.SimPriorityBasic) {
	process.targetRatios = targetRatios
	process.requiredSims = targetRatios.SimTypes()
}

func (process *RankingStatWeightProcess5) Run(timeout int) *util_async.FutureCancellable[weight_types.WeightResult] {
	process.printer.Printf("RankingStatWeightProcess5 RunOptimistic\n")
	run := rankInternalRun5_create(process)
	run.build.TimeLimitSeconds = timeout
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
	return run.run()
}

func rankInternalRun5_create(process *RankingStatWeightProcess5) *rankInternalRun5 {
	run := new(rankInternalRun5)
	run.process = process
	run.build = new(util_highs.LinearBuilder)
	run.build.Solver = util_highs.Solver_MIP_Interior

	run.build.BlendMultiObjectives = false
	run.objectiveInclude = run.build.AddObjectivePrioritised(true, -1, 0.05, 2)
	run.objectiveWeight = run.build.AddObjectivePrioritised(true, -1, -1, 1)

	return run
}

func (run *rankInternalRun5) run() *util_async.FutureCancellable[weight_types.WeightResult] {
	stopwatch := util.StopwatchMakeStopped()
	futureSolution := run.build.RunHighsFuture(stopwatch)
	return util_async.FutureCancellable_MapValue(futureSolution, func(linResult util_highs.LinearResult) (weight_types.WeightResult, bool) {
		solution := linResult.GetSolutionAndSaveLog(run.process.printer)
		if solution.HasSolution() {
			weight := run.extractAndReportSolution(solution)
			return weight_types.WeightResult{Weight: &weight, SolveTime: stopwatch.Elapsed(), Status: solution.Status}, true
		} else {
			return weight_types.WeightResult{SolveTime: stopwatch.Elapsed(), Status: solution.Status}, true
		}
	})
}

func (run *rankInternalRun5) createWeightColumns() {
	sumWeights := util_highs.ConstraintRow{Debug: "sumWeights"}
	run.weightColumns = make(map[stats.StatType]util_highs.ColumnIndex)
	for _, statType := range run.process.requiredStats {
		var colWeight util_highs.ColumnIndex
		colWeight = run.build.CreateColumnGeneral(highs.Continuous, c_rank5_weightLo, c_rank5_weightHi, util_highs.DebugString{Text: "WEIGHT " + statType.Name()})
		run.weightColumns[statType] = colWeight
		sumWeights.Add(colWeight, 1)
	}

	if run.process.WEIGHTSUM == 0 {
		sumWeightCol := run.build.CreateColumnWithObjective(highs.Continuous, c_rank5_weightTotalMin, c_rank5_weightTotalMax, 1, run.objectiveWeight, util_highs.DebugText("sumWeightCol"))
		sumWeights.Add(sumWeightCol, -1)
		sumWeights.Build(run.build, 0, 0)
	} else {
		sumWeights.Build(run.build, c_rank5_weightTotalTarget, c_rank5_weightTotalTarget)
	}
}

func (run *rankInternalRun5) supplyData(inputData []weight_types.WeightInput) {
	run.scaleStats = util_weight.ChooseStatScalingBasic(inputData, c_rank5_scaleTarget, false, run.process.printer)
	run.runData = util_collection.MapSliceAsNew(inputData, func(input *weight_types.WeightInput) *rankEntry5 {
		return &rankEntry5{
			RankStatWeightsCommon: weight_types.RankStatWeightsCommon{
				Data: input,
			},
		}
	})
}

func (run *rankInternalRun5) prepareRankings() {
	if run.process.SIMRANK == 1 {
		run.runData = simrank.RankingWeightsPrepareBasicRankingsRemoveDuplicates(run.process.requiredSims, &run.process.targetRatios, run.runData)
	} else if run.process.SIMRANK == 2 {
		run.runData = simrank.RankingWeightsPrepareUsingMidRangeRemoveDuplicates(run.process.requiredSims, &run.process.targetRatios, run.runData)
	} else {
		panic("SIMRANK not specified")
	}
}

func (run *rankInternalRun5) makeDataListEntryColumns() {
	for _, entry := range run.runData {
		run.makeEntryColumnRefs(entry)
	}
}

func (run *rankInternalRun5) makeEntryColumnRefs(entry *rankEntry5) {
	rankStr := strconv.FormatInt(int64(entry.TargetRank), 10)
	entry.ScoreCompute = run.build.CreateColumnGeneral(highs.Continuous, c_rank5_computeScoreLo, c_rank5_computeScoreHi, util_highs.DebugText("scoreCompute-"+rankStr))

	scoreRow := util_highs.ConstraintRow{Debug: "scoreRow"}
	for statType, weightColumn := range run.weightColumns {
		statValue := entry.Data.TotalStat.GetFloat(statType)
		statScale := run.scaleStats.GetOrPanic(statType)
		scoreRow.Add(weightColumn, statValue*statScale)
	}
	scoreRow.Add(entry.ScoreCompute, -1)
	scoreRow.Build(run.build, 0, 0)

	// TODO consider varying score, especially low end maybe?
	entry.IsInclude = run.build.CreateColumnBoolWithObjective(1, run.objectiveInclude, util_highs.DebugText("include-"+rankStr))

	entry.ScoreIfIncluded = run.build.CreateColumnGeneral(highs.Continuous, c_rank5_computeScoreLo, c_rank5_computeScoreHi, util_highs.DebugText("scoreIfIncluded-"+rankStr))
	run.build.ConstraintIfBoolCopy(entry.IsInclude, entry.ScoreCompute, 1, entry.ScoreIfIncluded, c_rank5_computeScoreM)
}

func (run *rankInternalRun5) makeDataListPairRules() {
	for a := 0; a < len(run.runData)-1; a++ {
		run.makeEntryPairScoreChecks(run.runData[a], run.runData[a+1], a, a+1)
	}

	if run.pairLinks.Size() != len(run.runData)-1 {
		panic("expected exact size")
	}
}

func (run *rankInternalRun5) makeEntryPairScoreChecks(lo *rankEntry5, hi *rankEntry5, indexLo, indexHi int) {
	indexText := strconv.FormatInt(int64(indexLo), 10)

	compareScore := util_highs.ConstraintRow{Debug: "compareScore " + indexText}
	compareScore.Add(hi.ScoreIfIncluded, 1)
	compareScore.Add(lo.ScoreIfIncluded, -1)
	compareScore.Build(run.build, 0, util_highs.InfPos())

	// TODO this seems a bit off, can a run of includes ever end?

	run.pairLinks.Put(indexLo, indexHi, &rankPair5{entryOne: lo, entryTwo: hi})
}

func (run *rankInternalRun5) extractAndReportSolution(solution *highs.Solution) weight_types.Weight1Basic {
	run.build.DebugPrintColumns(solution, run.process.printer)

	run.process.printer.Println("WEIGHTS")

	weight := weight_types.Weight1Basic_Make()
	for _, statType := range run.process.requiredStats {
		weightColumn := run.weightColumns[statType]
		statScale := run.scaleStats.GetOrPanic(statType)

		modelWeight := solution.ColValues[weightColumn]
		usableWeight := modelWeight * statScale

		weight.Put(statType, usableWeight)
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

	baseStat := run.process.requiredStats[0]
	divideBy := weight.Get(baseStat)
	for _, statType := range run.process.requiredStats {
		value := weight.Get(statType) / divideBy
		weight.Put(statType, value)
		run.process.printer.Printf("%10s %f\n", statType.Name(), value)
	}

	run.process.printer.Println("{")
	for _, statType := range run.process.requiredStats {
		value := weight.Get(statType)
		run.process.printer.Printf("  stats.%s: %f,\n", statType.EnumName(), value)
	}
	run.process.printer.Println("}")

	includeCount := 0
	for _, entry := range run.runData {
		includeValue := solution.ColValues[entry.IsInclude]
		if util.FloatEqualsOne(includeValue) {
			includeCount++
		}
	}
	run.process.printer.Printf("INCLUDE %d %d %f%%\n", includeCount, len(run.runData), float64(includeCount)/float64(len(run.runData))*100)

	return weight
}

func (run *rankInternalRun5) setupInitialSolutionFromExternal(weights weight_types.Weight1Basic) {
	weights = weights.ScaleForTotalSum(c_rank5_weightTotalTarget)

	for statType, colWeight := range run.weightColumns {
		value := weights.Get(statType)
		run.build.SetInitialSolutionValue(colWeight, value)
	}

	entryScores := make([]float64, len(run.runData))
	for i, entry := range run.runData {
		score := weights.CalcStatScoreScaled(entry.Data, run.scaleStats)
		// score := weights.CalcStatScore(entry.data)
		// score *= run.scaleStats
		entryScores[i] = score
		run.build.SetInitialSolutionValue(entry.ScoreCompute, score)
	}

	firstInclude, lastIndex := 0, len(run.runData)-1
	run.build.SetInitialSolutionValue(run.runData[firstInclude].IsInclude, 1)
	run.build.SetInitialSolutionValue(run.runData[firstInclude].ScoreIfIncluded, entryScores[firstInclude])

	for i := firstInclude + 1; i < lastIndex; i++ {
		entry := run.runData[i]
		if entryScores[i-1] <= entryScores[i] && entryScores[i] <= entryScores[i+1] {
			run.build.SetInitialSolutionValue(entry.IsInclude, 1)
			run.build.SetInitialSolutionValue(entry.ScoreIfIncluded, entryScores[i])
		} else {
			run.build.SetInitialSolutionValue(entry.IsInclude, 0)
			run.build.SetInitialSolutionValue(entry.ScoreIfIncluded, entryScores[i-1])
			entryScores[i] = entryScores[i-1]
		}
	}

	if entryScores[lastIndex-1] <= entryScores[lastIndex] {
		run.build.SetInitialSolutionValue(run.runData[lastIndex].IsInclude, 1)
		run.build.SetInitialSolutionValue(run.runData[lastIndex].ScoreIfIncluded, entryScores[lastIndex])
	} else {
		run.build.SetInitialSolutionValue(run.runData[lastIndex].IsInclude, 0)
		run.build.SetInitialSolutionValue(run.runData[lastIndex].ScoreIfIncluded, entryScores[lastIndex-1])
	}

	run.build.ValidateInitialSolutionState()
}
