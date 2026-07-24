package weight_highs

import (
	"cmp"
	"fmt"
	"math"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_highs"
	"paladin_gearing_go/weightfind/simrank"
	"paladin_gearing_go/weightfind/weight_types"
	"slices"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_rank_sep2_scoreMin           = 0.0
	c_rank_sep2_scoreMax           = 1.0
	c_rank_sep2_scoreSpacingFactor = 0.8  // can give it some extra slack
	c_rank_sep2_rangeSlackLimit    = 0.10 // 5% off the end values
)

type RankingSeparatedWeights2 struct {
	printer        *util.PrintRecorder
	timeoutSeconds int

	targetRatios  weight_types.SimPriorityBasic
	requiredStats []stats.StatType
	requiredSims  []stats.SimType
	dataEntries   []*rankEntrySeparated2

	build *util_highs.LinearBuilder

	desiredScoreSpacing   float64
	detailedWeightColumns util_collection.MapMap[stats.StatType, stats.SimType, util_highs.ColumnIndex]
	offsetColumns         map[stats.SimType]util_highs.ColumnIndex
	slackBottom, slackTop map[stats.SimType]util_highs.ColumnIndex
}

type rankEntrySeparated2 struct {
	data  *weight_types.WeightInput
	bySim util_collection.EnumMap[stats.SimType, rankDetailSeparated2]
}

func (r *rankEntrySeparated2) GetSimData() *stats.SimData {
	return &r.data.SimResult
}

func (r *rankEntrySeparated2) GetSimRankRangeByType(simType stats.SimType) *util_collection.HiLoInt {
	return r.bySim.GetOrPanic(simType).targetRankRange
}

func (r *rankEntrySeparated2) SetSimRankRangeByType(simType stats.SimType, targetRankRange *util_collection.HiLoInt) {
	r.bySim.Put(simType, rankDetailSeparated2{
		targetRankRange: targetRankRange,
		scoreColumn:     -1,
	})
}

type rankDetailSeparated2 struct {
	targetRankRange *util_collection.HiLoInt
	scoreColumn     util_highs.ColumnIndex
}

func (ranker *RankingSeparatedWeights2) Init(printer *util.PrintRecorder, timeoutSeconds int) {
	ranker.printer = printer
	ranker.timeoutSeconds = timeoutSeconds
}

func (ranker *RankingSeparatedWeights2) SupplyData(inputData []weight_types.WeightInput) {
	//inputData = takeDataSample_Random(inputData, 100)
	ranker.dataEntries = util_collection.MapSliceAsNew(inputData, func(input *weight_types.WeightInput) *rankEntrySeparated2 {
		return &rankEntrySeparated2{
			data:  input,
			bySim: util_collection.EnumMapMake[stats.SimType, rankDetailSeparated2](stats.SimTypeEnum),
		}
	})
}

func (ranker *RankingSeparatedWeights2) SetRequiredStats(requiredStats []stats.StatType, requiredSims []stats.SimType) {
	ranker.requiredStats = requiredStats
	ranker.requiredSims = requiredSims
}

func (ranker *RankingSeparatedWeights2) SetTargetRatios(targetRatios weight_types.SimPriorityBasic) {
	ranker.targetRatios = targetRatios
}

func (ranker *RankingSeparatedWeights2) newBuilder() {
	ranker.build = new(util_highs.LinearBuilder)
	ranker.build.Minimise = true
	ranker.build.TimeLimitSeconds = ranker.timeoutSeconds

	//ranker.build.Solver = util_highs.Solver_Force_Simplex
	//ranker.build.Solver = util_highs.Solver_Force_IPX // 15m35.9345904s
	//ranker.build.Solver = util_highs.Solver_LP_USE_GPU // wanders quite a bit
	ranker.build.Solver = util_highs.Solver_LP_NO_GPU // 3m48.0360506s
}

func (ranker *RankingSeparatedWeights2) Run(stopwatch *util.Stopwatch) *util_async.FutureCancellable[weight_types.Weight2Extended] {
	ranker.newBuilder()
	ranker.prepareRankings()
	ranker.createWeightColumns()
	ranker.processData()
	solutionFuture := ranker.build.RunHighsFuture(stopwatch)

	return util_async.FutureCancellable_MapValue(solutionFuture, func(linearResult util_highs.LinearResult) (weight_types.Weight2Extended, bool) {
		solution := linearResult.GetSolutionAndSaveLog(ranker.printer)
		return ranker.extractAndReportSolution(solution), true
	})
}

func (ranker *RankingSeparatedWeights2) createWeightColumns() {
	ranker.offsetColumns = make(map[stats.SimType]util_highs.ColumnIndex)
	ranker.slackBottom = make(map[stats.SimType]util_highs.ColumnIndex)
	ranker.slackTop = make(map[stats.SimType]util_highs.ColumnIndex)
	for _, simType := range ranker.requiredSims {
		ranker.createWeightColumnsForSim(simType)
		ranker.createOffsetColumn(simType)
	}
}

func (ranker *RankingSeparatedWeights2) createWeightColumnsForSim(simType stats.SimType) {
	for _, statType := range ranker.requiredStats {
		colWeight := ranker.build.CreateColumnGeneral(highs.Continuous, util_highs.C_MinusInf, util_highs.C_PlusInf, util_highs.DebugString{Text: "WEIGHT " + simType.Name() + " " + statType.Name()})
		ranker.detailedWeightColumns.Put(statType, simType, colWeight)
	}
}

func (ranker *RankingSeparatedWeights2) createOffsetColumn(simType stats.SimType) {
	ranker.offsetColumns[simType] = ranker.build.CreateColumnGeneral(highs.Continuous, util_highs.C_MinusInf, util_highs.C_PlusInf, util_highs.DebugText("offset-"+simType.Name()))
}

func (ranker *RankingSeparatedWeights2) prepareRankings() {
	simrank.RankSimsForRangedRankSeparated(ranker.requiredSims, ranker.dataEntries)
}

func (ranker *RankingSeparatedWeights2) processData() {
	ranker.desiredScoreSpacing = (c_rank_sep2_scoreMax - c_rank_sep2_scoreMin) / float64(len(ranker.dataEntries))
	ranker.desiredScoreSpacing *= c_rank_sep2_scoreSpacingFactor

	for _, simType := range ranker.requiredSims {
		ranker.processDataForSim(simType)
	}
}

func (ranker *RankingSeparatedWeights2) processDataForSim(simType stats.SimType) {
	ranker.makeDataListEntryColumns(simType)

	for baseIndex := range ranker.dataEntries {
		for compareTo := baseIndex + 1; compareTo < len(ranker.dataEntries); compareTo++ {
			ranker.makeEntryPair(
				ranker.dataEntries[baseIndex].bySim.GetOrPanic(simType),
				ranker.dataEntries[compareTo].bySim.GetOrPanic(simType),
			)
		}
	}

	foundLo, foundHi := false, false
	for _, entry := range ranker.dataEntries {
		detail := entry.bySim.GetOrPanic(simType)
		targetRange := detail.targetRankRange
		if targetRange.Lo == 0 {
			// yes we might overwrite if multiple at the low point, it's just for minor validation later
			ranker.slackBottom[simType] = ranker.makeEntryRankExact(0.0, detail)
			foundLo = true
		} else if targetRange.Hi == len(ranker.dataEntries)-1 {
			ranker.slackTop[simType] = ranker.makeEntryRankExact(1.0, detail)
			foundHi = true
		}
	}
	if !foundLo || !foundHi {
		panic("internal logic error, didn't find entries at top/bottom of range")
	}
}

func (ranker *RankingSeparatedWeights2) makeDataListEntryColumns(simType stats.SimType) {
	for _, entry := range ranker.dataEntries {
		detail := entry.bySim.GetOrPanic(simType)
		ranker.setupScoreColumn(&detail, entry, simType)
		entry.bySim.Put(simType, detail)
	}
}

func (ranker *RankingSeparatedWeights2) setupScoreColumn(detail *rankDetailSeparated2, entry *rankEntrySeparated2, simType stats.SimType) {
	targetRank := detail.targetRankRange
	debugStr := fmt.Sprintf("score-%s %d-%d", simType.Name(), targetRank.Lo, targetRank.Hi)
	detail.scoreColumn = ranker.build.CreateColumnGeneral(highs.Continuous, c_rank_sep2_scoreMin, c_rank_sep2_scoreMax, util_highs.DebugText(debugStr))

	scoreRow := util_highs.ConstraintRow{Debug: debugStr}
	for statType, weightColumn := range ranker.detailedWeightColumns.SeqKey1ValueWithKey2(simType) {
		statValue := entry.data.TotalStat.GetFloat(statType)
		scoreRow.Add(weightColumn, statValue)
	}

	offsetCol := ranker.offsetColumns[simType]
	scoreRow.Add(offsetCol, 1)

	scoreRow.Add(detail.scoreColumn, -1)
	scoreRow.Build(ranker.build, 0, 0)
}

func (ranker *RankingSeparatedWeights2) makeEntryPair(one rankDetailSeparated2, two rankDetailSeparated2) {
	slack := ranker.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.C_PlusInf, 1, util_highs.DebugText("slack"))

	targetRankOne := one.targetRankRange
	targetRankTwo := two.targetRankRange
	if targetRankOne.Lo > targetRankTwo.Hi {
		// range one entirely greater
		gapSize := targetRankOne.Lo - targetRankTwo.Hi
		desiredPairSpacing := ranker.desiredScoreSpacing * float64(gapSize)
		row := util_highs.ConstraintRow{}
		row.Add(one.scoreColumn, 1)
		row.Add(two.scoreColumn, -1)
		row.Add(slack, 1)
		row.Build(ranker.build, desiredPairSpacing, util_highs.C_PlusInf)
	} else if targetRankTwo.Lo > targetRankOne.Hi {
		// range two entirely greater
		gapSize := targetRankTwo.Lo - targetRankOne.Hi
		desiredPairSpacing := ranker.desiredScoreSpacing * float64(gapSize)
		row := util_highs.ConstraintRow{}
		row.Add(two.scoreColumn, 1)
		row.Add(one.scoreColumn, -1)
		row.Add(slack, 1)
		row.Build(ranker.build, desiredPairSpacing, util_highs.C_PlusInf)
	} else {
		// overlapping range, no constraint
	}
}

func (ranker *RankingSeparatedWeights2) makeEntryRankExact(targetScore float64, detail rankDetailSeparated2) util_highs.ColumnIndex {
	slackEnd := ranker.build.CreateColumnWithOutput(highs.Continuous, 0, util_highs.C_PlusInf, 1, util_highs.DebugText("slack-end"))

	ranker.build.AbsoluteValueFromDiffOneToConst(
		detail.scoreColumn, 1,
		targetScore,
		slackEnd,
		"end",
	)

	return slackEnd
}

func (ranker *RankingSeparatedWeights2) extractAndReportSolution(solution *highs.Solution) weight_types.Weight2Extended {
	ranker.build.DebugPrintColumns(solution, ranker.printer)

	statWeightResult := weight_types.Weight2Extended_Make(ranker.requiredStats, ranker.requiredSims)
	for entry := range ranker.detailedWeightColumns.SeqKey1Key2ValueEntries() {
		weightColumn := entry.Value
		weightValue := solution.ColValues[weightColumn]
		statWeightResult.PutWeight(entry.Key1, entry.Key2, weightValue)
	}
	for simType, offsetCol := range ranker.offsetColumns {
		offsetValue := solution.ColValues[offsetCol]
		ratio := ranker.targetRatios.GetOrPanic(simType)
		statWeightResult.SetSimScale(simType, 1, offsetValue, ratio)
	}
	statWeightResult.FinishAndValidate()

	statWeightResult.Print(ranker.printer)
	ranker.reportExamples(statWeightResult)
	ranker.reportCompleteRanges(statWeightResult)

	slackLimit := (c_rank_sep2_scoreMax - c_rank_sep2_scoreMin) * c_rank_sep2_rangeSlackLimit
	for _, simType := range ranker.requiredSims {
		slackBottom := solution.ColValues[ranker.slackBottom[simType]]
		slackTop := solution.ColValues[ranker.slackTop[simType]]
		if math.Abs(slackBottom) > slackLimit || math.Abs(slackTop) > slackLimit {
			//panic(fmt.Sprintf("solution didn't fill range properly slacks=%e %e", slackBottom, slackTop))
		}
	}

	return *statWeightResult
}

func (ranker *RankingSeparatedWeights2) reportExamples(weightExtended *weight_types.Weight2Extended) {
	for i := range min(20, len(ranker.dataEntries)) {
		data := ranker.dataEntries[i]
		ranker.printer.Println("EXAMPLE")

		for _, simType := range ranker.requiredSims {
			rowScore := 0.0
			ranker.printer.Printf(" %10s", simType.Name())
			for _, statType := range ranker.requiredStats {
				statValue := data.data.TotalStat.GetFloat(statType)
				weight := weightExtended.GetWeightOrPanic(statType, simType)
				ranker.printer.Printf(" {%s %.0f * %.2e = %.4e}", statType.Name(), statValue, weight, statValue*weight)
				rowScore += statValue * weight
			}

			priorityEntry := weightExtended.GetSimPriority().GetOrPanic(simType)
			offset := priorityEntry.RangingOffset

			rowScore += offset
			simValue := data.data.SimResult.Get(simType)

			ranker.printer.Printf(" + {offset %.4e} = %.4f {sim %.6f}\n", offset, rowScore, simValue)
		}

		ranker.printer.Println0()
	}
}

func (ranker *RankingSeparatedWeights2) reportCompleteRanges(weightExtended *weight_types.Weight2Extended) {
	type simAndScore struct {
		sim, score float64
	}

	for _, simType := range ranker.requiredSims {
		typeResults := make([]simAndScore, len(ranker.dataEntries))
		for i := range len(ranker.dataEntries) {
			data := ranker.dataEntries[i]

			rowScore := 0.0
			for _, statType := range ranker.requiredStats {
				statValue := data.data.TotalStat.GetFloat(statType)
				weight := weightExtended.GetWeightOrPanic(statType, simType)
				rowScore += statValue * weight
			}
			priorityEntry := weightExtended.GetSimPriority().GetOrPanic(simType)
			offset := priorityEntry.RangingOffset
			rowScore += offset
			simValue := data.data.SimResult.Get(simType)
			typeResults[i] = simAndScore{simValue, rowScore}
		}

		ranker.printer.Println("RANGE " + simType.Name())
		slices.SortFunc(typeResults, func(a, b simAndScore) int { return cmp.Compare(a.sim, b.sim) })
		ranker.printer.Printf("   sim lo: sim=%e score=%e\n", typeResults[0].sim, typeResults[0].score)
		ranker.printer.Printf("   sim hi: sim=%e score=%e\n", typeResults[len(typeResults)-1].sim, typeResults[len(typeResults)-1].score)
		slices.SortFunc(typeResults, func(a, b simAndScore) int { return cmp.Compare(a.score, b.score) })
		ranker.printer.Printf(" score lo: sim=%e score=%e\n", typeResults[0].sim, typeResults[0].score)
		ranker.printer.Printf(" score hi: sim=%e score=%e\n", typeResults[len(typeResults)-1].sim, typeResults[len(typeResults)-1].score)
	}

	ranker.printer.Println("EXAMPLE")
}
