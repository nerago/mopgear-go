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

const c_timeLimit = 30 * time.Minute

type SolverHighsMultiParam struct {
	Label          string
	ItemOptions    items.FullOptionsMap
	Gear_model     *gear_model.Model
	RatingMultiply float64

	setup        *singleGearSetInputs
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

func (process *SolverHighsMultiProcess) RunInterruptable(printer *util.PrintRecorder) *channel_op.FutureCancellable[HighsMultiResult] {
	process.makeFullModel()

	solveFuture := process.build.RunHighsFuture()
	multiFuture := channel_op.FutureCancellable_Make[HighsMultiResult]()
	channel_op.ChainCancel(multiFuture, solveFuture)

	go func() {
		result, gotResult := solveFuture.WaitForResult()
		if gotResult {
			solution, log := result.Solution, result.Log
			printer.AppendOther(log)
			printer.Println("SOLUTION STATUS = " + solution.Status.String())

			debugPrintAll(solution, process, printer)

			if solution.HasSolution() {
				multiResult := process.solutionToResult(solution, printer)
				multiFuture.SetResult(multiResult)
			} else {
				multiFuture.SetResultEmpty()
			}
		} else {
			multiFuture.SetResultEmpty()
		}
	}()
	return multiFuture
}

func (process *SolverHighsMultiProcess) RunForSeveral_CommonDifferent(printer *util.PrintRecorder, outputTarget util.Optional[int], cancel channel_op.CancelSignal) <-chan HighsMultiResult {
	resultChannel := make(chan HighsMultiResult, 8)

	// process.common.Get()

	go func() {
		initialResult, bestCommonChoices, hasInitial := process.generateInitialMulti(printer, cancel)
		if hasInitial {
			resultChannel <- initialResult

			if target, hasTarget := outputTarget.GetWithFlag(); hasTarget && target < len(bestCommonChoices) {
				util.Shuffle(bestCommonChoices)
				bestCommonChoices = bestCommonChoices[0:target]
			}

			printer.Printf("COMMON VARIANT count %d\n", len(bestCommonChoices))

			innerChannel := process.generateWithDifferentCommonVariants(bestCommonChoices, printer, cancel)
			channel_op.ChannelCopy(innerChannel, resultChannel)
		} else {
			close(resultChannel)
		}
	}()

	return resultChannel
}

func (process *SolverHighsMultiProcess) generateInitialMulti(printer *util.PrintRecorder, cancel channel_op.CancelSignal) (HighsMultiResult, []*columnInfo, bool) {
	printer.Printf("INITIAL MULTI run\n")

	process.makeFullModel()
	future := process.build.RunHighsFuture()
	channel_op.ChainCancel(cancel, future)

	result, gotResult := future.WaitForResult()
	if gotResult {
		solution := result.Solution
		printer.Println("SOLUTION STATUS = " + solution.Status.String())
		printer.AppendOther(result.Log)
		debugPrintAll(solution, process, printer)

		if solution.HasSolution() {
			initialResult := process.solutionToResult(solution, printer)
			bestCommonChoices := process.extractCommonChoices(solution)
			return initialResult, bestCommonChoices, true
		}
	}
	return HighsMultiResult{}, nil, false
}

func (process *SolverHighsMultiProcess) generateWithDifferentCommonVariants(bestCommonChoices []*columnInfo, printer *util.PrintRecorder, cancel channel_op.CancelSignal) <-chan HighsMultiResult {
	return channel_op.MapOptional_SliceToChannel_Cancellable(10, bestCommonChoices, cancel, func(changeColumn **columnInfo) (HighsMultiResult, bool) {
		return process.generateWithDifferentCommonVariant_One(printer, *changeColumn, cancel)
	})
}

// func (process *SolverHighsMultiProcess) generateWithForbiddenCommon_One(printer *util.PrintRecorder, forbiddenRef items.ItemRef, cancel channel_op.CancelSignal) (HighsMultiResult, bool) {
// innerPrint := util.PrintRecorder_HoldAll()
// printer.Printf("COMMON FORBIDDEN %s\n", forbiddenRef.String())

// build := process.build.Clone()
// rowLimitCommon := utilhighs.ConstraintRow{Debug: "rowLimitCommon"}
// for _, part := range process.parts {
// 		for column := range part.setup.itemColumns.ValuesForKeyAsSeq(forbiddenRef.ItemId) {
// 			if column.item.EqualsFull(item) {
// 			}
// 		}
// 	}
// rowLimitCommon.Add(changeColumn.columnIndex, 1)
// rowLimitCommon.Build(build, 0, 0)

// return process.runVariant(build, cancel, innerPrint, printer)
// }

func (process *SolverHighsMultiProcess) generateWithDifferentCommonVariant_One(printer *util.PrintRecorder, changeColumn *columnInfo, cancel channel_op.CancelSignal) (HighsMultiResult, bool) {
	innerPrint := util.PrintRecorder_HoldAll()
	printer.Printf("COMMON VARIANT blocking %s\n", changeColumn.itemFull.CreateString())

	build := process.build.Clone()
	rowLimitCommon := utilhighs.ConstraintRow{Debug: "rowLimitCommon"}
	rowLimitCommon.Add(changeColumn.columnIndex, 1)
	rowLimitCommon.Build(build, 0, 0)

	return process.runVariant(build, cancel, innerPrint, printer)
}

func (process *SolverHighsMultiProcess) runVariant(build *utilhighs.LinearBuilder, cancel channel_op.CancelSignal, innerPrint *util.PrintRecorder, printer *util.PrintRecorder) (HighsMultiResult, bool) {
	future := build.RunHighsFuture()
	channel_op.ChainCancel(cancel, future)

	result, gotResult := future.WaitForResult()
	if gotResult {
		solution := result.Solution
		innerPrint.Println("SOLUTION STATUS = " + solution.Status.String())
		innerPrint.AppendOther(result.Log)

		innerPrint.Println("############################################################################")
		printer.AppendOther(innerPrint)

		if solution.HasSolution() {
			return process.solutionToResult(solution, innerPrint), true
		}
	}
	return HighsMultiResult{}, false
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
	commonChosenColumns := make([]*columnInfo, 0, process.common.Size())
	for _, jobColumn := range process.allColumns {
		colValue := solution.ColValues[jobColumn.columnIndex]
		if jobColumn.entryType == entry_multi_enable_forge && util.FloatEqualsOne(colValue) {
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
	process.build.TimeLimitSeconds = util.RoundToInt(c_timeLimit.Seconds())
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

func (param *SolverHighsMultiParam) doSetup(build *utilhighs.LinearBuilder, job *SolverHighsMultiProcess) {
	param.solveOptions = items.SolvableOptionsMap_of(&param.ItemOptions)
	param.setup = setupGearSet(build, param.Gear_model, &param.solveOptions, 0)
	job.outputRow.Add(param.setup.mainOutputVar.columnIndex, param.RatingMultiply)
}

func (process *SolverHighsMultiProcess) addCommonConstraints(build *utilhighs.LinearBuilder) {
	for itemRef, array := range process.common.SeqGroups() {
		process.addCommonConstraintsForItemRef(build, itemRef, array)
	}
}

func (process *SolverHighsMultiProcess) addCommonConstraintsForItemRef(build *utilhighs.LinearBuilder, itemRef items.ItemRef, array []items.FullItem) {
	onlyOneReforge := utilhighs.ConstraintRow{Debug: "onlyOneReforge" + itemRef.ItemId.String()}

	for _, item := range array {
		entryEnableReforge := columnInfo{entryType: entry_multi_enable_forge, itemFull: &item}
		entryEnableReforge.columnIndex = build.CreateColumnBool(&entryEnableReforge)
		process.allColumns = append(process.allColumns, &entryEnableReforge)

		onlyOneReforge.Add(entryEnableReforge.columnIndex, 1)

		for partUsedItem := range process.findMatchingItemColumnsForCommon(&item) {
			// formula is partUsedItem <= enableReforge
			//            0 <= enableReforge - partUsedItem
			matchingReforge := utilhighs.ConstraintRow{Debug: "matchingReforge" + itemRef.ItemId.String() + "_" + strconv.Itoa(int(partUsedItem))}
			matchingReforge.Add(entryEnableReforge.columnIndex, 1)
			matchingReforge.Add(partUsedItem, -1)
			matchingReforge.Build(build, 0, 1)
		}
	}

	onlyOneReforge.Build(build, 0, 1)
}

func (process *SolverHighsMultiProcess) findMatchingItemColumnsForCommon(item *items.FullItem) iter.Seq[utilhighs.ColumnIndex] {
	return func(yield func(utilhighs.ColumnIndex) bool) {
		for _, part := range process.parts {
			for column := range part.setup.itemColumns.ValuesForKeyAsSeq(item.ItemId()) {
				// doesn't technically compare on RandomSuffix, but stats compare should be fine
				if column.item.EqualsFull(item) {
					if !yield(column.columnIndex) {
						return
					}
				}
			}
		}
	}
}
