package weight_highs

import (
	"fmt"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_rank_sep_scaleTarget       = 1.0
	c_rank_sep_maxWeight         = 10.0
	c_rank_sep_targetTotalWeight = 1.0
	c_rank_sep_scoreMin          = 0.0
	c_rank_sep_scoreMax          = 1.0
	c_rank_sep_score_unequal_gap = 0.0001 // needs to fit N samples within scoreMax
)

type RankingSeparatedWeights struct {
	printer        *util.PrintRecorder
	timeoutSeconds int

	targetRatios  weight_types.SimPriorityBasic
	requiredStats []stats.StatType
	requiredSims  []stats.SimType
	dataEntries   []rankEntrySeparated

	build *util_highs.LinearBuilder

	detailedWeightColumns util.MapMap[stats.StatType, stats.SimType, util_highs.ColumnIndex]
	offsetColumns         map[stats.SimType]util_highs.ColumnIndex
}

type rankEntrySeparated struct {
	data  *weight_types.WeightInput
	bySim util.EnumMap[stats.SimType, rankDetailSeparated]
}

type rankDetailSeparated struct {
	simScore    float64
	targetRank  int
	scoreColumn util_highs.ColumnIndex
}

func (ranker *RankingSeparatedWeights) Init(printer *util.PrintRecorder, timeoutSeconds int) {
	ranker.printer = printer
	ranker.timeoutSeconds = timeoutSeconds
}

func (ranker *RankingSeparatedWeights) SupplyData(inputData []weight_types.WeightInput) {
	ranker.dataEntries = util.MapSliceAsNew(inputData, func(input *weight_types.WeightInput) rankEntrySeparated {
		return rankEntrySeparated{
			data:  input,
			bySim: util.EnumMapMake[stats.SimType, rankDetailSeparated](stats.SimTypeEnum),
		}
	})
}

func (ranker *RankingSeparatedWeights) SetRequiredStats(requiredStats []stats.StatType, requiredSims []stats.SimType) {
	ranker.requiredStats = requiredStats
	ranker.requiredSims = requiredSims
}

func (ranker *RankingSeparatedWeights) SetTargetRatios(targetRatios weight_types.SimPriorityBasic) {
	ranker.targetRatios = targetRatios
}

func (ranker *RankingSeparatedWeights) newBuilder() {
	ranker.build = new(util_highs.LinearBuilder)
	ranker.build.Minimise = true
	ranker.build.TimeLimitSeconds = ranker.timeoutSeconds
	ranker.build.Solver = util_highs.Solver_Force_Simplex
}

func (ranker *RankingSeparatedWeights) Run(stopwatch *util.Stopwatch) *util_async.FutureCancellable[weight_types.Weight2Extended] {
	// FULL RUN
	ranker.newBuilder()
	ranker.prepareRankings()
	ranker.createWeightColumns()
	ranker.doAlgos()
	solutionFuture := ranker.build.RunHighsFuture(stopwatch)

	return util_async.FutureCancellable_MapValue(solutionFuture, func(linearResult util_highs.LinearResult) (weight_types.Weight2Extended, bool) {
		solution := linearResult.GetSolutionAndSaveLog(ranker.printer)
		return ranker.extractAndReportSolution(solution), true
	})
}

func (ranker *RankingSeparatedWeights) createWeightColumns() {
	ranker.offsetColumns = make(map[stats.SimType]util_highs.ColumnIndex)
	for _, simType := range ranker.requiredSims {
		ranker.createWeightColumnsForSim(simType)
		ranker.createOffsetColumn(simType)
	}
}

func (ranker *RankingSeparatedWeights) createWeightColumnsForSim(simType stats.SimType) {
	sumWeights := util_highs.ConstraintRow{Debug: "sumWeights"}

	for _, statType := range ranker.requiredStats {
		colWeight := ranker.build.CreateColumnGeneral(highs.Continuous, -c_rank_sep_maxWeight, c_rank_sep_maxWeight, util_highs.DebugString{Text: "WEIGHT " + simType.Name() + " " + statType.Name()})
		ranker.detailedWeightColumns.Put(statType, simType, colWeight)
		sumWeights.Add(colWeight, 1)
	}

	// not clear that exact sum of weights is always the best algorithm in the other ranking versions
	// but have chosen it before, and certainly faster

	// maybe needLow types should be negative instead
	//sumWeights.Build(ranker.build, c_rank_sep_targetTotalWeight, c_rank_sep_targetTotalWeight)
}

func (ranker *RankingSeparatedWeights) createOffsetColumn(simType stats.SimType) {
	ranker.offsetColumns[simType] = ranker.build.CreateColumnGeneral(highs.Continuous, util_highs.C_MinusInf, util_highs.C_PlusInf, util_highs.DebugText("offset-"+simType.Name()))
}

func (ranker *RankingSeparatedWeights) prepareRankings() {
	// score each sim
	for _, simType := range ranker.requiredSims {
		for entry, simDetailRank := range util.CalculateRanking(simType.IsHighGood(), ranker.dataEntries, func(x *rankEntrySeparated) float64 { return x.data.SimResult.Get(simType) }) {
			entry.bySim.Put(simType, rankDetailSeparated{
				simScore:    entry.data.SimResult.Get(simType),
				targetRank:  simDetailRank,
				scoreColumn: 0,
			})
		}
	}
	//TODO use simrank
}

func (ranker *RankingSeparatedWeights) doAlgos() {
	for _, simType := range ranker.requiredSims {
		ranker.makeDataListEntryColumns(simType)

		for baseIndex := range ranker.dataEntries {
			for compareTo := baseIndex + 1; compareTo < len(ranker.dataEntries); compareTo++ {
				ranker.makeEntryPair(
					ranker.dataEntries[baseIndex].bySim.GetOrPanic(simType),
					ranker.dataEntries[compareTo].bySim.GetOrPanic(simType),
				)
			}
		}

		minRankEntry := util.FindMinFunc(ranker.dataEntries, func(e *rankEntrySeparated) int { return e.bySim.GetOrPanic(simType).targetRank })
		ranker.makeEntryRankExact(0.0, minRankEntry.bySim.GetOrPanic(simType))
		maxRankEntry := util.FindMaxFunc(ranker.dataEntries, func(e *rankEntrySeparated) int { return e.bySim.GetOrPanic(simType).targetRank })
		ranker.makeEntryRankExact(1.0, maxRankEntry.bySim.GetOrPanic(simType))
	}
}

func (ranker *RankingSeparatedWeights) makeDataListEntryColumns(simType stats.SimType) {
	for entry := range util.ForPointer(ranker.dataEntries) {
		detail := entry.bySim.GetOrPanic(simType)
		ranker.setupScoreColumn(&detail, entry, simType)
		entry.bySim.Put(simType, detail)
	}
}

func (ranker *RankingSeparatedWeights) setupScoreColumn(detail *rankDetailSeparated, entry *rankEntrySeparated, simType stats.SimType) {
	debugStr := fmt.Sprintf("score-%s-%d", simType.Name(), detail.targetRank)
	detail.scoreColumn = ranker.build.CreateColumnGeneral(highs.Continuous, c_rank_sep_scoreMin, c_rank_sep_scoreMax, util_highs.DebugText(debugStr))

	scoreRow := util_highs.ConstraintRow{Debug: "scoreRow-" + debugStr}
	for statType, weightColumn := range ranker.detailedWeightColumns.SeqInnerWithKey2Value(simType) {
		statValue := entry.data.TotalStat.GetFloat(statType)
		scoreRow.Add(weightColumn, statValue)
	}

	offsetCol := ranker.offsetColumns[simType]
	scoreRow.Add(offsetCol, 1)

	scoreRow.Add(detail.scoreColumn, -1)
	scoreRow.Build(ranker.build, 0, 0)
}

func (ranker *RankingSeparatedWeights) makeEntryPair(one rankDetailSeparated, two rankDetailSeparated) {
	slack := ranker.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.C_PlusInf, 1, util_highs.DebugText("slack"))

	if one.targetRank > two.targetRank {
		row := util_highs.ConstraintRow{}
		row.Add(one.scoreColumn, 1)
		row.Add(two.scoreColumn, -1)
		row.Add(slack, 1)
		row.Build(ranker.build, c_rank_sep_score_unequal_gap, util_highs.C_PlusInf)
	} else if two.targetRank > one.targetRank {
		row := util_highs.ConstraintRow{}
		row.Add(two.scoreColumn, 1)
		row.Add(one.scoreColumn, -1)
		row.Add(slack, 1)
		row.Build(ranker.build, c_rank_sep_score_unequal_gap, util_highs.C_PlusInf)
	} else {
		// TODO overlapping ranks may be permitted
		panic("unexpected equal ranks")
	}
}

func (ranker *RankingSeparatedWeights) makeEntryRankExact(targetScore float64, detail rankDetailSeparated) {
	slackEnd := ranker.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.C_PlusInf, 1, util_highs.DebugText("slack-end"))

	ranker.build.AbsoluteValueFromDiffOneToConst(
		detail.scoreColumn, 1,
		targetScore,
		slackEnd,
		"end",
	)
}

func (ranker *RankingSeparatedWeights) extractAndReportSolution(solution *highs.Solution) weight_types.Weight2Extended {
	ranker.build.DebugPrintColumns(solution, ranker.printer)

	statWeightResult := weight_types.Weight2Extended_Make(ranker.requiredStats, ranker.requiredSims)
	for entry := range ranker.detailedWeightColumns.SeqWithKeys() {
		weightColumn := entry.Value
		weightValue := solution.ColValues[weightColumn]
		statWeightResult.PutWeight(entry.Key1, entry.Key2, weightValue)
	}
	for simType, offsetCol := range ranker.offsetColumns {
		offsetValue := solution.ColValues[offsetCol]
		ratio := ranker.targetRatios.GetOrPanic(simType)
		statWeightResult.SimPriority.SetSimScale(simType, 1, offsetValue, ratio)
	}
	statWeightResult.FinishAndValidate()

	return *statWeightResult
}
