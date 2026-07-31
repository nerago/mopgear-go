package solve_highs

import (
	"iter"
	gear_model "paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_highs"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
	"github.com/google/uuid"
)

const c_timeLimit = 4000 // seconds

type SolverHighsMultiParam struct {
	Label          string
	ItemOptions    items.FullOptionsMap
	Gear_model     *gear_model.SpecModel
	RatingMultiply float64

	setup        *singleGearSetBasic
	solveOptions items.SolvableOptionsMap
}

type SolverHighsMultiProcess struct {
	build *util_highs.LinearBuilder

	common multi_types.CommonOptions
	parts  []SolverHighsMultiParam

	outputColumn util_highs.ColumnIndex
	outputRow    util_highs.ConstraintRow

	allColumns []*columnInfo

	permuteLabel string
}

type HighsMultiResult struct {
	ItemSets      []items.FullItemSet
	OutputId      []string
	PermuteLabel  string
	InterimResult bool
}

func (process *SolverHighsMultiProcess) AddSetParam(param SolverHighsMultiParam) {
	process.parts = append(process.parts, param)
}

func (process *SolverHighsMultiProcess) SetCommon(common multi_types.CommonOptions) {
	process.common = common
}

func (process *SolverHighsMultiProcess) SetPermuteLabel(permuteLabel string) {
	process.permuteLabel = permuteLabel
}

func (process *SolverHighsMultiProcess) RunInterruptable(printer *util.PrintRecorder) *util_async.FutureCancellable[HighsMultiResult] {
	process.makeFullModel()

	solveFuture := process.build.RunHighsFuture(nil)
	return util_async.FutureCancellable_MapValue(solveFuture, func(linearResult util_highs.LinearResult) (HighsMultiResult, bool) {
		solution := linearResult.GetSolution2AndSaveLog(printer)

		debugPrintAll(solution, process, printer)

		if solution.HasSolution() {
			multiResult := process.solutionToResult(solution, printer, false)
			return multiResult, true
		} else {
			return HighsMultiResult{}, false
		}
	})
}

func (process *SolverHighsMultiProcess) RunForSeveral_CommonDifferent(printer *util.PrintRecorder, outputTarget util_collection.Optional[int], cancel util_async.CancelSignal, alsoDoFullItemBlocks bool, includeInterimResults bool) (resultChannelRead <-chan HighsMultiResult, expectedCount *util_async.Future[int]) {
	resultChannel := make(chan HighsMultiResult, 8)
	expectedCount = util_async.Future_Make[int]()

	var interimChannel chan HighsMultiResult
	if includeInterimResults {
		interimChannel = resultChannel
	}

	go func() {
		initialResult, bestCommonChoices, hasInitial := process.generateInitialMulti(printer, cancel, interimChannel)
		if hasInitial {
			resultChannel <- initialResult

			blockPlanList := process.chooseVariantsToRun(bestCommonChoices, alsoDoFullItemBlocks, outputTarget, expectedCount, printer)
			process.generateWithDifferentVariants(blockPlanList, printer, cancel, resultChannel, includeInterimResults)
		}
		close(resultChannel)
	}()

	return resultChannel, expectedCount
}

func (process *SolverHighsMultiProcess) chooseVariantsToRun(bestCommonChoices []*columnInfo, alsoDoFullItemBlocks bool, outputTarget util_collection.Optional[int], expectedCount *util_async.Future[int], printer *util.PrintRecorder) []blockPlan {
	blockPlanList := make([]blockPlan, 0)
	for _, commonColumn := range bestCommonChoices {
		blockPlanList = append(blockPlanList, blockPlan{changeColumn: commonColumn})
		if alsoDoFullItemBlocks {
			itemId := commonColumn.ItemId()
			blockPlanList = append(blockPlanList, blockPlan{forbiddenItem: &itemId})
		}
	}

	util_collection.Shuffle(blockPlanList)
	if target, hasTarget := outputTarget.GetWithFlag(); hasTarget && target < len(blockPlanList) {
		blockPlanList = blockPlanList[0:target]
	}

	count := len(blockPlanList)
	expectedCount.SetResult(count)
	printer.Printf("VARIANT COMMON count %d\n", len(blockPlanList))
	return blockPlanList
}

type blockPlan struct {
	forbiddenItem *items.ItemId
	changeColumn  *columnInfo
}

func (process *SolverHighsMultiProcess) generateInitialMulti(printer *util.PrintRecorder, cancel util_async.CancelSignal, interimResults chan<- HighsMultiResult) (HighsMultiResult, []*columnInfo, bool) {
	printer.Printf("INITIAL MULTI run\n")

	process.makeFullModel()

	var doneFunc func()
	if interimResults != nil {
		doneFunc = process.forwardInterimResultsToChannel(process.build, interimResults, printer)
		defer doneFunc()
	}

	future := process.build.RunHighsFuture(nil)
	util_async.ChainCancel(cancel, future)

	result, gotResult := future.WaitForResult()
	if gotResult {
		solution := result.GetSolution2AndSaveLog(printer)
		debugPrintAll(solution, process, printer)

		if solution.HasSolution() {
			initialResult := process.solutionToResult(solution, printer, false)
			bestCommonChoices := process.extractCommonChoices(solution)
			return initialResult, bestCommonChoices, true
		}
	}
	return HighsMultiResult{}, nil, false
}

func (process *SolverHighsMultiProcess) generateWithDifferentVariants(blockPlans []blockPlan, printer *util.PrintRecorder, cancel util_async.CancelSignal, resultChannel chan<- HighsMultiResult, includeInterimResults bool) {
	expectedResults := util_async.MapMulti_SliceToChannel_Cancellable(10, blockPlans, cancel, func(block *blockPlan) *util_async.FutureCancellable[HighsMultiResult] {
		innerPrint := util.PrintRecorder_HoldAll()

		var build *util_highs.LinearBuilder
		if block.changeColumn != nil {
			build = process.prepareWithDifferentCommonVariant_One(printer, block.changeColumn)
		} else if block.forbiddenItem != nil {
			build = process.prepareVariantWithBlockedItem_One(printer, *block.forbiddenItem)
		} else {
			panic("missing block")
		}

		if includeInterimResults {
			return process.runVariant(build, innerPrint, printer, resultChannel)
		} else {
			return process.runVariant(build, innerPrint, printer, nil)
		}
	})
	util_async.ChannelCopy(expectedResults, resultChannel)
}

func (process *SolverHighsMultiProcess) prepareVariantWithBlockedItem_One(printer *util.PrintRecorder, forbiddenId items.ItemId) *util_highs.LinearBuilder {
	printer.Printf("VARIANT COMMON blocking all %d\n", forbiddenId)

	build := process.build.Clone()
	rowLimitCommon := util_highs.ConstraintRow{Debug: "rowLimitAll"}
	for _, part := range process.parts {
		for column := range part.setup.itemColumns.ValuesForKeyAsSeq(forbiddenId) {
			rowLimitCommon.Add(column.columnIndex, 1)
		}
	}
	rowLimitCommon.Build(build, 0, 0)
	return build
}

func (process *SolverHighsMultiProcess) prepareWithDifferentCommonVariant_One(printer *util.PrintRecorder, changeColumn *columnInfo) *util_highs.LinearBuilder {
	printer.Printf("VARIANT COMMON blocking reforge %d\n", changeColumn.ItemId())

	build := process.build.Clone()
	rowLimitCommon := util_highs.ConstraintRow{Debug: "rowLimitCommon"}
	rowLimitCommon.Add(changeColumn.columnIndex, 1)
	rowLimitCommon.Build(build, 0, 0)
	return build
}

func (process *SolverHighsMultiProcess) runVariant(build *util_highs.LinearBuilder, innerPrint *util.PrintRecorder, printer *util.PrintRecorder, interimChannel chan<- HighsMultiResult) *util_async.FutureCancellable[HighsMultiResult] {
	var doneFunc func()
	if interimChannel != nil {
		doneFunc = process.forwardInterimResultsToChannel(build, interimChannel, innerPrint)
	}

	future := build.RunHighsFuture(nil)

	return util_async.FutureCancellable_MapValue(future, func(linearResult util_highs.LinearResult) (HighsMultiResult, bool) {
		solution := linearResult.GetSolution2AndSaveLog(innerPrint)

		if doneFunc != nil {
			doneFunc()
		}

		innerPrint.Println("############################################################################")
		printer.AppendOther(innerPrint)

		if solution.HasSolution() {
			return process.solutionToResult(solution, innerPrint, false), true
		} else {
			return HighsMultiResult{}, false
		}
	})
}

func (process *SolverHighsMultiProcess) prepareCallbackForInterimSolutions(build *util_highs.LinearBuilder, interimChannel chan<- util_highs.InterimSolution) {
	build.SetCallback([]highs.CallbackType{highs.CallbackTypeMipImprovingSolution},
		func(callbackType highs.CallbackType, str string, out highs.HighsCallbackDataOut) highs.HighsCallbackDataIn {
			if callbackType == highs.CallbackTypeMipImprovingSolution {
				interim := util_highs.InterimSolutionFromCallback(out)
				interimChannel <- interim
			}
			return highs.HighsCallbackDataIn{}
		},
	)
}

func (process *SolverHighsMultiProcess) forwardInterimResultsToChannel(build *util_highs.LinearBuilder, resultChannel chan<- HighsMultiResult, printer *util.PrintRecorder) func() {
	interimChannel := make(chan util_highs.InterimSolution)
	process.prepareCallbackForInterimSolutions(build, interimChannel)
	go func() {
		for solution := range interimChannel {
			resultChannel <- process.solutionToResult(solution, printer, true)
		}
	}()
	return func() { close(interimChannel) }
}

func debugPrintAll(solution *util_highs.Solution2, job *SolverHighsMultiProcess, printer *util.PrintRecorder) {
	if !util_highs.C_DebugHighs {
		return
	}

	printer.Println("SOLUTION STATUS = " + solution.Status().String())
	printer.Printf("OBJECTIVE VALUE %f \n", solution.Objective()*c_scaled_ratings)

columnLoop:
	for columnIndex, outputValue := range solution.ColValuesSeq() {
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

func (process *SolverHighsMultiProcess) extractCommonChoices(solution *util_highs.Solution2) []*columnInfo {
	commonChosenColumns := make([]*columnInfo, 0, process.common.Size())
	for _, jobColumn := range process.allColumns {
		colValue := solution.GetValue(jobColumn.columnIndex)
		if jobColumn.entryType == entry_multi_enable_forge && util.FloatEqualsOne(colValue) {
			commonChosenColumns = append(commonChosenColumns, jobColumn)
		}
	}
	return commonChosenColumns
}

func (process *SolverHighsMultiProcess) solutionToResult(solution util_highs.ISolution, printer *util.PrintRecorder, interim bool) HighsMultiResult {
	printer.Printf("Permute = %s\n", process.permuteLabel)

	resultList := make([]items.FullItemSet, len(process.parts))
	idList := make([]string, len(process.parts))
	for partIndex := range process.parts {
		part := process.parts[partIndex]
		solvedSet := part.setup.buildResultSet(solution)
		validateNewSet(solvedSet, &part.solveOptions, part.Gear_model)
		fullItemSet := items.FullItemSet_FromSolved(solvedSet, &part.ItemOptions)
		resultList[partIndex] = fullItemSet

		outputId := uuid.NewString()
		if interim {
			outputId += "-interim"
		}
		printer.Printf("OutputId[%s] = %s\n", part.Label, outputId)
		idList[partIndex] = outputId
	}

	return HighsMultiResult{resultList, idList, process.permuteLabel, interim}
}

func (process *SolverHighsMultiProcess) makeFullModel() {
	process.build = &util_highs.LinearBuilder{}
	process.build.TimeLimitSeconds = c_timeLimit
	process.build.Solver = util_highs.Solver_MIP_Interior // TODO check if actually best

	entry := columnInfo{entryType: entry_multi_output}
	entry.columnIndex = process.build.CreateColumnWithOutput(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), 1, &entry)
	process.allColumns = append(process.allColumns, &entry)

	process.outputColumn = entry.columnIndex
	process.outputRow.Add(process.outputColumn, -1)

	for partIndex := range process.parts {
		process.parts[partIndex].doSetup(process.build, process)
	}

	process.addCommonConstraints(process.build)

	process.outputRow.Build(process.build, 0, 0)
}

func (param *SolverHighsMultiParam) doSetup(build *util_highs.LinearBuilder, job *SolverHighsMultiProcess) {
	param.solveOptions = items.SolvableOptionsMap_of(&param.ItemOptions)
	param.setup = setupGearSet(build, param.Gear_model, &param.solveOptions, 0)
	job.outputRow.Add(param.setup.mainOutputVar.columnIndex, param.RatingMultiply)
}

func (process *SolverHighsMultiProcess) addCommonConstraints(build *util_highs.LinearBuilder) {
	for itemRef, array := range process.common.SeqGroups() {
		process.addCommonConstraintsForItemRef(build, itemRef, array)
	}
}

func (process *SolverHighsMultiProcess) addCommonConstraintsForItemRef(build *util_highs.LinearBuilder, itemRef items.ItemRef, array []items.FullItem) {
	onlyOneReforge := util_highs.ConstraintRow{Debug: "onlyOneReforge" + itemRef.ItemId.String()}

	for _, item := range array {
		entryEnableReforge := columnInfo{entryType: entry_multi_enable_forge, itemFull: &item}
		entryEnableReforge.columnIndex = build.CreateColumnBool(&entryEnableReforge)
		process.allColumns = append(process.allColumns, &entryEnableReforge)

		onlyOneReforge.Add(entryEnableReforge.columnIndex, 1)

		for partUsedItem := range process.findMatchingItemColumnsForCommon(&item) {
			// formula is partUsedItem <= enableReforge
			//            0 <= enableReforge - partUsedItem
			matchingReforge := util_highs.ConstraintRow{Debug: "matchingReforge" + itemRef.ItemId.String() + "_" + strconv.Itoa(int(partUsedItem))}
			matchingReforge.Add(entryEnableReforge.columnIndex, 1)
			matchingReforge.Add(partUsedItem, -1)
			matchingReforge.Build(build, 0, 1)
		}
	}

	onlyOneReforge.Build(build, 0, 1)
}

func (process *SolverHighsMultiProcess) findMatchingItemColumnsForCommon(item *items.FullItem) iter.Seq[util_highs.ColumnIndex] {
	return func(yield func(util_highs.ColumnIndex) bool) {
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
