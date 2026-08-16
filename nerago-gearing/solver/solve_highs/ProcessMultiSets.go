package solve_highs

import (
	"iter"
	"paladin_gearing_go/items"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/solver/solve_highs_types"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_highs"
	"reflect"
	"strconv"

	"github.com/bartolsthoorn/gohighs/highs"
	"github.com/google/uuid"
)

type SolverHighsMultiParam struct {
	Label          string
	ItemOptions    items.FullOptionsMap
	RatingMultiply float64
	SolverModel    solve_highs_types.SolverModel
	singleGearSet  ISingleGearSet
	solveOptions   items.SolvableOptionsMap
}

type SolverHighsMultiProcess struct {
	build *util_highs.LinearBuilder

	common multi_types.CommonOptions
	parts  []SolverHighsMultiParam

	outputColumn util_highs.ColumnIndex
	outputRow    util_highs.ConstraintRow

	permuteLabel string
}

type HighsMultiResult struct {
	Entries       map[string]HighsMultiResultEntry
	PermuteLabel  string
	InterimResult bool
}

type HighsMultiResultEntry struct {
	ItemSet  items.FullItemSet
	OutputId string
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

func (process *SolverHighsMultiProcess) RunInterruptable(timeLimit int, printer *util.PrintRecorder) *util_async.FutureCancellable[HighsMultiResult] {
	process.makeFullModel(timeLimit)

	solveFuture := process.build.RunHighsFuture(nil)
	return util_async.FutureCancellable_MapValue(solveFuture, func(linearResult util_highs.LinearResult) (HighsMultiResult, bool) {
		solution := linearResult.GetSolution2AndSaveLog(printer)
		debugPrintAll(solution, process.build, printer)

		if solution.HasSolution() {
			multiResult := process.solutionToResult(solution, printer, false)
			return multiResult, true
		} else {
			return HighsMultiResult{}, false
		}
	})
}

func (process *SolverHighsMultiProcess) RunForSeveral_CommonDifferent(timeLimit int, printer *util.PrintRecorder, outputTarget util_collection.Optional[int], cancel util_async.CancelSignal, alsoDoFullItemBlocks bool, includeInterimResults bool) (resultChannelRead <-chan HighsMultiResult, expectedCount *util_async.Future[int]) {
	resultChannel := make(chan HighsMultiResult, 8)
	expectedCount = util_async.Future_Make[int]()

	initialChannel, bestCommonChoicesFuture := process.generateInitialMulti(timeLimit, printer, includeInterimResults)
	util_async.ChainCancel(cancel, bestCommonChoicesFuture)
	util_async.ChannelCopy(initialChannel, resultChannel, false)

	bestCommonChoicesFuture.ForwardResultToRelevantCallback(func(bestCommonChoices []*columnInfo) {
		blockPlanList := process.chooseVariantsToRun(bestCommonChoices, alsoDoFullItemBlocks, outputTarget, expectedCount, printer)
		innerChannel := process.generateWithDifferentVariants(blockPlanList, printer, cancel, includeInterimResults)
		util_async.ChannelCopy(innerChannel, resultChannel, true)
	}, func() {
		close(resultChannel)
	})

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
	if target, hasTarget := outputTarget.GetWithFlag(); hasTarget && target-1 < len(blockPlanList) {
		blockPlanList = blockPlanList[0 : target-1]
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

func (process *SolverHighsMultiProcess) generateInitialMulti(timeLimit int, printer *util.PrintRecorder, includeInterimResults bool) (<-chan HighsMultiResult, *util_async.FutureCancellable[[]*columnInfo]) {
	initialChannel := make(chan HighsMultiResult)
	printer.Printf("INITIAL MULTI run\n")

	process.makeFullModel(timeLimit)

	var doneFunc func()
	if includeInterimResults {
		doneFunc = process.forwardInterimResultsToChannel(process.build, initialChannel, printer)
	}

	futureSolution := process.build.RunHighsFuture(nil)
	futureCommonChoices := util_async.FutureCancellable_MapValue(futureSolution, func(result util_highs.LinearResult) ([]*columnInfo, bool) {
		solution := result.GetSolution2AndSaveLog(printer)
		debugPrintAll(solution, process.build, printer)

		if solution.HasSolution() {
			initialChannel <- process.solutionToResult(solution, printer, false)
			bestCommonChoices := process.extractCommonChoices(solution)
			return bestCommonChoices, true
		}

		close(initialChannel)

		return nil, false
	})

	if doneFunc != nil {
		futureCommonChoices.AddCompletedHandler(doneFunc)
	}

	return initialChannel, futureCommonChoices
}

func (process *SolverHighsMultiProcess) generateWithDifferentVariants(blockPlans []blockPlan, printer *util.PrintRecorder, cancel util_async.CancelSignal, includeInterimResults bool) <-chan HighsMultiResult {
	return util_async.MapMulti_SliceToChannel_Cancellable(10, blockPlans, cancel, func(block *blockPlan, resultChannel chan<- HighsMultiResult) {
		innerPrint := util.PrintRecorder_HoldAll()

		var build *util_highs.LinearBuilder
		if block.changeColumn != nil {
			build = process.prepareWithDifferentCommonVariant_One(printer, block.changeColumn)
		} else if block.forbiddenItem != nil {
			build = process.prepareVariantWithBlockedItem_One(printer, *block.forbiddenItem)
		} else {
			panic("missing block")
		}

		var futureResult *util_async.FutureCancellable[HighsMultiResult]
		if includeInterimResults {
			futureResult = process.runVariant(build, innerPrint, resultChannel)
		} else {
			futureResult = process.runVariant(build, innerPrint, nil)
		}
		util_async.ChainCancel(cancel, futureResult)

		if result, hasResult := futureResult.WaitForResult(); hasResult {
			resultChannel <- result
		}

		printer.AppendOther(innerPrint)
	})
}

func (process *SolverHighsMultiProcess) prepareVariantWithBlockedItem_One(printer *util.PrintRecorder, forbiddenId items.ItemId) *util_highs.LinearBuilder {
	printer.Printf("VARIANT COMMON blocking all %d\n", forbiddenId)

	build := process.build.Clone()
	rowLimitCommon := util_highs.ConstraintRow{Debug: "rowLimitAll"}
	for _, part := range process.parts {
		for column := range part.singleGearSet.ColumnsForItemId(forbiddenId) {
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

func (process *SolverHighsMultiProcess) runVariant(build *util_highs.LinearBuilder, printer *util.PrintRecorder, interimChannel chan<- HighsMultiResult) *util_async.FutureCancellable[HighsMultiResult] {
	var doneFunc func()
	if interimChannel != nil {
		doneFunc = process.forwardInterimResultsToChannel(build, interimChannel, printer)
	}

	future := build.RunHighsFuture(nil)

	return util_async.FutureCancellable_MapValue(future, func(linearResult util_highs.LinearResult) (HighsMultiResult, bool) {
		solution := linearResult.GetSolution2AndSaveLog(printer)
		debugPrintAll(solution, process.build, printer)

		if doneFunc != nil {
			doneFunc()
		}

		printer.Println("############################################################################")

		if solution.HasSolution() {
			return process.solutionToResult(solution, printer, false), true
		} else {
			return HighsMultiResult{}, false
		}
	})
}

func (process *SolverHighsMultiProcess) prepareCallbackForInterimSolutions(build *util_highs.LinearBuilder, interimChannel chan<- util_highs.InterimSolution) {
	build.SetCallback([]highs.CallbackType{highs.CallbackTypeMipImprovingSolution},
		func(callbackType highs.CallbackType, str string, out highs.CallbackData) highs.CallbackResult {
			if callbackType == highs.CallbackTypeMipImprovingSolution {
				interim := util_highs.InterimSolutionFromCallback(out)
				interimChannel <- interim
			}
			return highs.CallbackResult{}
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

func debugPrintAll(solution *util_highs.Solution2, build *util_highs.LinearBuilder, printer *util.PrintRecorder) {
	if !util_highs.C_DebugHighs {
		return
	}

	printer.Println("SOLUTION STATUS = " + solution.Status().String())
	printer.Printf("OBJECTIVE VALUE %f \n", solution.Objective())

	for columnIndex, outputValue := range solution.ColValuesSeq() {
		context := build.DebugContextFor(columnIndex)
		if colInfo, isInfo := context.(*columnInfo); isInfo {
			debugPrintColumnEntry(colInfo, columnIndex, outputValue, nil, nil, printer)
		} else {
			text := build.DebugTextFor(columnIndex)
			printer.Printf("%d %f %s\n", columnIndex, outputValue, text)
		}
	}
}

func (process *SolverHighsMultiProcess) extractCommonChoices(solution *util_highs.Solution2) []*columnInfo {
	commonChosenColumns := make([]*columnInfo, 0, process.common.Size())
	for columnIndex, colValue := range solution.ColValuesSeq() {
		context := process.build.DebugContextFor(columnIndex)
		if colInfo, isInfo := context.(*columnInfo); isInfo {
			if colInfo.entryType == entry_multi_enable_forge && util.FloatEqualsOne(colValue) {
				commonChosenColumns = append(commonChosenColumns, colInfo)
			}
		}
	}
	return commonChosenColumns
}

func (process *SolverHighsMultiProcess) solutionToResult(solution util_highs.ISolution, printer *util.PrintRecorder, interim bool) HighsMultiResult {
	printer.Printf("Make Solution Permute = %s\n", process.permuteLabel)

	resultMap := make(map[string]HighsMultiResultEntry)
	for partIndex := range process.parts {
		part := process.parts[partIndex]
		solvedSet := part.singleGearSet.buildResultSet(solution)
		validateNewSet(solvedSet, &part.solveOptions, part.SolverModel.CheckSet)
		fullItemSet := items.FullItemSet_FromSolved(solvedSet, &part.ItemOptions)

		printer.Printf("Inner solver type = %s\n", reflect.TypeOf(part.singleGearSet).Elem().Name())

		outputId := uuid.NewString()
		if interim {
			outputId += "-interim"
		}
		printer.Printf("OutputId[%s] = %s\n", part.Label, outputId)

		resultMap[part.Label] = HighsMultiResultEntry{fullItemSet, outputId}
	}
	printer.Println0()

	return HighsMultiResult{resultMap, process.permuteLabel, interim}
}

func (process *SolverHighsMultiProcess) makeFullModel(timeLimit int) {
	process.build = &util_highs.LinearBuilder{}
	process.build.TimeLimitSeconds = timeLimit
	process.build.Solver = util_highs.Solver_MIP_Interior // TODO check if actually best

	entry := columnInfo{entryType: entry_multi_output}
	entry.columnIndex = process.build.CreateColumnWithOutput(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), 1, &entry)

	process.outputColumn = entry.columnIndex
	process.outputRow.Add(process.outputColumn, -1)

	for partIndex := range process.parts {
		process.parts[partIndex].makeSingleGearSet(process.build, process)
	}

	process.addCommonConstraints(process.build)

	process.outputRow.Build(process.build, 0, 0)
}

func (param *SolverHighsMultiParam) makeSingleGearSet(build *util_highs.LinearBuilder, job *SolverHighsMultiProcess) {
	param.solveOptions = items.SolvableOptionsMap_of(&param.ItemOptions)

	setOutput := &columnInfo{entryType: entry_main_output}
	setOutput.columnIndex = build.CreateColumnGeneral(highs.Continuous, 0, util_highs.InfPos(), setOutput)

	param.singleGearSet = makeGearSetForWeight(build, &param.SolverModel)
	param.singleGearSet.setup(&param.SolverModel, &param.solveOptions, setOutput)

	job.outputRow.Add(setOutput.columnIndex, param.RatingMultiply/param.singleGearSet.RatingPreScale())
}

func makeGearSetForWeight(build *util_highs.LinearBuilder, model *solve_highs_types.SolverModel) ISingleGearSet {
	if model.Weights3 != nil {
		return makeGearSetExtended3(build)
	} else if model.Weights2 != nil {
		return makeGearSetExtended2(build)
	} else if model.Weights1 != nil {
		return makeGearSetBasic(build)
	} else {
		panic("missing weight")
	}
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
			for column := range part.singleGearSet.ColumnsForItemId(item.ItemId()) {
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
