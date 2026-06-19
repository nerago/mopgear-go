package withhighs

import (
	"iter"
	"paladin_gearing_go/items"
	gear_model "paladin_gearing_go/model"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"strconv"
	"time"

	"github.com/bartolsthoorn/gohighs/highs"
	"github.com/google/uuid"
)

const c_timeLimit = 15 * 60 // 15 mins

type SolverHighsMultiParam struct {
	Label          string
	ItemOptions    items.FullOptionsMap
	Gear_model     *gear_model.Model
	RatingMultiply float64

	setup        *setupInputSetAware
	solveOptions items.SolvableOptionsMap

	withMinimum *stats.StatAndValue
}

type SolverHighsMultiProcess struct {
	build *utilhighs.LinearBuilder

	common multi_types.CommonOptions
	parts  []SolverHighsMultiParam

	outputColumn utilhighs.ColumnIndex
	outputRow    utilhighs.ConstraintRow

	allColumns []*columnInfo
}

type HighsMultiResult struct {
	ItemSets []items.FullItemSet
	OutputId []string
}

func (process *SolverHighsMultiProcess) AddSetParam(param SolverHighsMultiParam) {
	process.parts = append(process.parts, param)
}

func (process *SolverHighsMultiProcess) SetCommon(common multi_types.CommonOptions) {
	process.common = common
}

func (process *SolverHighsMultiProcess) Run(printer *util.PrintRecorder) util.Optional[HighsMultiResult] {
	process.makeFullModel()
	solution, log := process.build.RunHighs()
	printer.AppendOther(log)
	printer.Println("SOLUTION STATUS = " + solution.Status.String())

	debugPrintAll(solution, process, printer)

	if solution.HasSolution() {
		return util.Optional_OfValue(process.solutionToResult(solution, printer))
	} else {
		return util.Optional_Empty[HighsMultiResult]()
	}
}

func debugPrintAll(solution *highs.Solution, job *SolverHighsMultiProcess, printer *util.PrintRecorder) {
	if !utilhighs.C_DebugHighs {
		return
	}

	printer.Printf("OBJECTIVE VALUE %f \n", solution.Objective*c_scaled_ratings)

columnLoop:
	for colIndex, outputValue := range solution.ColValues {
		columnIndex := utilhighs.ColumnIndex(colIndex)

		if debugPrintColumn(job.allColumns, columnIndex, outputValue, nil, nil, printer) {
			continue columnLoop
		}

		for _, part := range job.parts {
			if debugPrintColumn(part.setup.allColumns, columnIndex, outputValue, nil, nil, printer) {
				continue columnLoop
			}
		}

		text := job.build.DebugTextFor(columnIndex)
		printer.Printf("%d %f %s\n", columnIndex, outputValue, text)
	}
}

func (process *SolverHighsMultiProcess) extractCommonChoices(solution *highs.Solution) []*columnInfo {
	// seenItems := make(map[items.ItemId]bool)
	commonChosenColumns := make([]*columnInfo, 0, len(process.common))
	for _, jobColumn := range process.allColumns {
		colValue := solution.ColValues[jobColumn.columnIndex]
		if jobColumn.entryType == entry_multi_enable_forge && utilhighs.FloatEqualsOne(colValue) {
			commonChosenColumns = append(commonChosenColumns, jobColumn)
			// seenItems[jobColumn.itemFull.ItemId()] = true
		}
	}

	// TODO could further and block every item used in a set
	// for _, part := range process.parts {
	// 	for itemId, seqColumns := range part.setup.itemColumns.SeqGroupsNestedKeyValue() {

	// 	}
	// }

	return commonChosenColumns
}

func (process *SolverHighsMultiProcess) solutionToResult(solution *highs.Solution, printer *util.PrintRecorder) HighsMultiResult {
	resultList := make([]items.FullItemSet, len(process.parts))
	idList := make([]string, len(process.parts))
	for partIndex := range process.parts {
		part := process.parts[partIndex]
		solvedSet := part.setup.buildResultSet(solution, &part.solveOptions, part.Gear_model)
		fullItemSet := items.FullItemSet_FromSolved(solvedSet, &part.ItemOptions)
		resultList[partIndex] = fullItemSet

		outputId := uuid.NewString()
		printer.Printf("OutputId[%s] = %s\n", part.Label, outputId)
		idList[partIndex] = outputId
	}
	return HighsMultiResult{resultList, idList}
}

func (process *SolverHighsMultiProcess) makeFullModel() {
	process.build = &utilhighs.LinearBuilder{}
	process.build.TimeLimitSeconds = c_timeLimit // an hour
	process.build.Solver = utilhighs.Solver_MIP_Interior

	entry := columnInfo{entryType: entry_multi_output}
	entry.columnIndex = process.build.CreateColumnWithOutput(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, 1, &entry)
	process.allColumns = append(process.allColumns, &entry)

	process.outputColumn = entry.columnIndex
	process.outputRow.Add(process.outputColumn, -1)

	for partIndex := range process.parts {
		process.parts[partIndex].doSetup(process.build, process)
	}

	process.addCommonConstraints(process.build)

	process.outputRow.Build(process.build, 0, 0)
}

func (process *SolverHighsMultiProcess) RunForSeveral_CommonDifferent(printer *util.PrintRecorder, outputTarget util.Optional[int]) <-chan HighsMultiResult {
	printer.Printf("INITIAL MULTI run\n")

	process.makeFullModel()
	startTime1 := time.Now()
	solution, log := process.build.RunHighs()
	printer.Println("Solve initial duration = " + time.Since(startTime1).String())
	printer.AppendOther(log)
	printer.Println("SOLUTION STATUS = " + solution.Status.String())
	// debugPrintAll(solution, job, printer)

	if !solution.HasSolution() {
		return nil
	}

	initialResult := process.solutionToResult(solution, printer)
	bestCommonChoices := process.extractCommonChoices(solution)

	if target, hasTarget := outputTarget.GetWithFlag(); hasTarget && target < len(bestCommonChoices) {
		util.Shuffle(bestCommonChoices)
		bestCommonChoices = bestCommonChoices[0:target]
	}

	printer.Println("############################################################################")
	printer.Printf("COMMON VARIANT count %d\n", len(bestCommonChoices))
	printer.Println("############################################################################")

	resultChannel := make(chan HighsMultiResult, 8)
	resultChannel <- initialResult

	if len(bestCommonChoices) == 0 {
		printer.Println("WARNWARNWARNWARNWARNWARNWARNWARNWARN")
		printer.Println("WARNWARNWARNWARNWARNWARNWARNWARNWARN")
		printer.Println("WARNWARNWARNWARNWARNWARNWARNWARNWARN")
		printer.Println("WARN no common choices found WARWARN")
		printer.Println("WARNWARNWARNWARNWARNWARNWARNWARNWARN")
		printer.Println("WARNWARNWARNWARNWARNWARNWARNWARNWARN")
		printer.Println("WARNWARNWARNWARNWARNWARNWARNWARNWARN")
		return resultChannel
	}

	channel_op.Map_SliceToChannel_Provided(10, bestCommonChoices, resultChannel, func(changeColumn **columnInfo, resultChannel chan<- HighsMultiResult) {
		innerPrint := util.PrintRecorder_HoldAll()
		printer.Printf("COMMON VARIANT blocking %s\n", (*changeColumn).itemFull.CreateString())

		build := process.build.Clone()
		rowLimitCommon := utilhighs.ConstraintRow{Debug: "rowLimitCommon"}
		rowLimitCommon.Add((*changeColumn).columnIndex, 1)
		rowLimitCommon.Build(build, 0, 0)

		startTime2 := time.Now()
		solution, log := build.RunHighs()
		printer.Println("Solve loop duration = " + time.Since(startTime2).String())
		innerPrint.Println("SOLUTION STATUS = " + solution.Status.String())
		innerPrint.AppendOther(log)

		if solution.HasSolution() {
			jobResult := process.solutionToResult(solution, innerPrint)
			resultChannel <- jobResult
		}

		innerPrint.Println("############################################################################")
		printer.AppendOther(innerPrint)
	})

	return resultChannel
}

func (param *SolverHighsMultiParam) doSetup(build *utilhighs.LinearBuilder, job *SolverHighsMultiProcess) {
	param.solveOptions = items.SolvableOptionsMap_of(&param.ItemOptions)
	param.setup = setupHighsSetAware(build, param.Gear_model, &param.solveOptions, 0)
	job.outputRow.Add(param.setup.mainOutputVar.columnIndex, param.RatingMultiply)
}

func (process *SolverHighsMultiProcess) addCommonConstraints(build *utilhighs.LinearBuilder) {
	for itemId, array := range process.common {
		process.addCommonConstraintsForItem(build, itemId, array)
	}
}

func (process *SolverHighsMultiProcess) addCommonConstraintsForItem(build *utilhighs.LinearBuilder, itemId items.ItemId, array []items.FullItem) {
	onlyOneReforge := utilhighs.ConstraintRow{Debug: "onlyOneReforge" + itemId.String()}

	for _, item := range array {
		entryEnableReforge := columnInfo{entryType: entry_multi_enable_forge, itemFull: &item}
		entryEnableReforge.columnIndex = build.CreateColumnBool(&entryEnableReforge)
		process.allColumns = append(process.allColumns, &entryEnableReforge)

		onlyOneReforge.Add(entryEnableReforge.columnIndex, 1)

		for partUsedItem := range process.findMatchingItemColumns(&item) {
			// formula is partUsedItem <= enableReforge
			//            0 <= enableReforge - partUsedItem
			matchingReforge := utilhighs.ConstraintRow{Debug: "matchingReforge" + itemId.String() + "_" + strconv.Itoa(int(partUsedItem))}
			matchingReforge.Add(entryEnableReforge.columnIndex, 1)
			matchingReforge.Add(partUsedItem, -1)
			matchingReforge.Build(build, 0, 1)
		}
	}

	onlyOneReforge.Build(build, 0, 1)
}

func (process *SolverHighsMultiProcess) findMatchingItemColumns(item *items.FullItem) iter.Seq[utilhighs.ColumnIndex] {
	return func(yield func(utilhighs.ColumnIndex) bool) {
		for _, part := range process.parts {
			for column := range part.setup.itemColumns.ValuesForKeyAsSeq(item.ItemId()) {
				if column.item.EqualsFull(item) {
					if !yield(column.columnIndex) {
						return
					}
				}
			}
		}
	}
}
