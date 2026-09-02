package solve_highs

import (
	"fmt"
	"iter"
	"reflect"
	"strconv"

	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/multi/multi_types"
	"github.com/nerago/mopgear-go/solver/solve_highs_types"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_highs"

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
	Entries       map[string]highsMultiResultEntry
	PermuteLabel  string
	InterimResult bool
}

type highsMultiResultEntry struct {
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

func (process *SolverHighsMultiProcess) Run(timeLimit int, printer *util.PrintRecorder, alternateMode multi_types.AlternateMode, alternateTarget util_collection.Optional[int], cancel util_async.CancelSignal, includeInterimResults bool) (<-chan HighsMultiResult, *util_async.Future[int], *util_async.PossibleFutureError, error) {
	resultChannel := make(chan HighsMultiResult)
	expectedCount := util_async.Future_Make[int]()
	futureError := util_async.PossibleFutureErrorMake()

	bestCommonChoicesFuture, errInit := process.generateInitialMulti(timeLimit, printer, includeInterimResults, resultChannel)
	if errInit != nil {
		return nil, nil, nil, errInit
	}
	errInit = util_async.ChainCancel(cancel, bestCommonChoicesFuture)
	if errInit != nil {
		return nil, nil, nil, errInit
	}

	bestCommonChoicesFuture.ForwardResultOrErrorToCallback(func(bestCommonChoices []*columnInfo) {
		if alternateMode != multi_types.AlternateModeNone {
			blockPlanList := process.chooseVariantsToRun(
				bestCommonChoices,
				alternateMode == multi_types.AlternateModeItemAndReforgeBlocks, alternateTarget,
				printer)

			expectedCount.SetResult(len(blockPlanList) + 1)

			errGenerate := process.generateWithDifferentVariants(blockPlanList,
				printer,
				cancel,
				includeInterimResults,
				resultChannel)

			if errGenerate != nil {
				futureError.AddResultError(errGenerate)
			}
		} else {
			expectedCount.SetResult(1)
		}

		close(resultChannel)

	}, func(errUpstream error) {
		expectedCount.SetResult(0)
		close(resultChannel)

		if errUpstream != nil {
			futureError.AddResultError(errUpstream)
		}
	})

	return resultChannel, expectedCount, futureError, nil
}

func (process *SolverHighsMultiProcess) chooseVariantsToRun(bestCommonChoices []*columnInfo, alsoDoFullItemBlocks bool, outputTarget util_collection.Optional[int], printer *util.PrintRecorder) []blockPlan {
	blockPlanList := make([]blockPlan, 0)
	for _, commonColumn := range bestCommonChoices {
		blockPlanList = append(blockPlanList, blockPlan{changeColumn: commonColumn})
		if alsoDoFullItemBlocks {
			blockPlanList = append(blockPlanList, blockPlan{forbiddenItem: new(commonColumn.itemId())})
		}
	}

	util_collection.Shuffle(blockPlanList)
	if target, hasTarget := outputTarget.GetWithFlag(); hasTarget && target-1 < len(blockPlanList) {
		blockPlanList = blockPlanList[0 : target-1]
	}

	printer.Printf("VARIANT COMMON count %d\n", len(blockPlanList))
	return blockPlanList
}

type blockPlan struct {
	forbiddenItem *items.ItemId
	changeColumn  *columnInfo
}

func (process *SolverHighsMultiProcess) generateInitialMulti(timeLimit int, printer *util.PrintRecorder, includeInterimResults bool, resultChannel chan<- HighsMultiResult) (*util_async.FutureCancellableWithError[[]*columnInfo], error) {
	printer.Printf("INITIAL MULTI run\n")

	err1 := process.makeFullModel(timeLimit)
	if err1 != nil {
		return nil, err1
	}

	if includeInterimResults {
		process.forwardInterimResultsToChannel(process.build, resultChannel, printer)
	}

	futureSolution := process.build.RunHighsFuture(nil)

	futureCommonChoices := util_async.FutureCancellable_MapValueError(futureSolution, func(result util_highs.LinearResult) (*[]*columnInfo, error) {
		solution, err := result.GetSolution2AndSaveLog(printer)
		if err != nil {
			return nil, err
		}

		err = debugPrintAll(solution, process.build, printer)
		if err != nil {
			return nil, err
		}

		if solution.HasSolution() {
			if multiResult, err := process.solutionToResult(solution, printer, false); err != nil {
				return nil, err
			} else {
				resultChannel <- *multiResult
			}

			bestCommonChoices := process.extractCommonChoices(solution)
			return &bestCommonChoices, nil
		} else {
			return nil, util.ErrorTracedNew("highs solve " + solution.Status().String())
		}
	})

	return futureCommonChoices, nil
}

func (process *SolverHighsMultiProcess) generateWithDifferentVariants(blockPlans []blockPlan, printer *util.PrintRecorder, cancel util_async.CancelSignal, includeInterimResults bool, resultChannel chan<- HighsMultiResult) error {
	return util_async.ForEach_Slice_Cancellable_PassError(10, blockPlans, cancel, func(block *blockPlan) error {
		return process.generateOneVariant(block, printer, includeInterimResults, resultChannel, cancel)
	})
}

func (process *SolverHighsMultiProcess) generateOneVariant(block *blockPlan, printer *util.PrintRecorder, includeInterimResults bool, resultChannel chan<- HighsMultiResult, cancel util_async.CancelSignal) error {
	innerPrint := util.PrintRecorder_HoldAll()

	var build *util_highs.LinearBuilder
	if block.changeColumn != nil {
		build = process.prepareWithDifferentCommonVariant_One(printer, block.changeColumn)
	} else if block.forbiddenItem != nil {
		build = process.prepareVariantWithBlockedItem_One(printer, *block.forbiddenItem)
	} else {
		return util.ErrorTracedNew("missing block")
	}

	var futureResult *util_async.FutureCancellableWithError[HighsMultiResult]
	if includeInterimResults {
		futureResult = process.runVariant(build, innerPrint, resultChannel)
	} else {
		futureResult = process.runVariant(build, innerPrint, nil)
	}

	if err := util_async.ChainCancel(cancel, futureResult); err != nil {
		return err
	}

	if result, hasResult := futureResult.WaitForResult(); hasResult {
		if result.Error != nil {
			return result.Error
		}
		resultChannel <- *result.Value
	}

	printer.AppendOther(innerPrint)

	return nil
}

func (process *SolverHighsMultiProcess) prepareVariantWithBlockedItem_One(printer *util.PrintRecorder, forbiddenId items.ItemId) *util_highs.LinearBuilder {
	printer.Printf("VARIANT COMMON blocking all %d\n", forbiddenId)

	build := process.build.Clone()
	rowLimitCommon := util_highs.ConstraintRow{Debug: "rowLimitAll"}
	for _, part := range process.parts {
		for column := range part.singleGearSet.columnsForItemId(forbiddenId) {
			rowLimitCommon.Add(column.columnIndex, 1)
		}
	}
	rowLimitCommon.Build(build, 0, 0)
	return build
}

func (process *SolverHighsMultiProcess) prepareWithDifferentCommonVariant_One(printer *util.PrintRecorder, changeColumn *columnInfo) *util_highs.LinearBuilder {
	printer.Printf("VARIANT COMMON blocking reforge %d\n", changeColumn.itemId())

	build := process.build.Clone()
	rowLimitCommon := util_highs.ConstraintRow{Debug: "rowLimitCommon"}
	rowLimitCommon.Add(changeColumn.columnIndex, 1)
	rowLimitCommon.Build(build, 0, 0)
	return build
}

func (process *SolverHighsMultiProcess) runVariant(build *util_highs.LinearBuilder, printer *util.PrintRecorder, interimChannel chan<- HighsMultiResult) *util_async.FutureCancellableWithError[HighsMultiResult] {
	if interimChannel != nil {
		process.forwardInterimResultsToChannel(build, interimChannel, printer)
	}

	future := build.RunHighsFuture(nil)

	return util_async.FutureCancellable_MapValueError(future, func(linearResult util_highs.LinearResult) (*HighsMultiResult, error) {
		solution, err := linearResult.GetSolution2AndSaveLog(printer)
		if err != nil {
			return nil, err
		}

		err = debugPrintAll(solution, process.build, printer)
		if err != nil {
			return nil, err
		}

		if solution.HasSolution() {
			return process.solutionToResult(solution, printer, false)
		} else {
			return nil, util.ErrorTracedNew("highs status " + solution.Status().String())
		}
	})
}

func (process *SolverHighsMultiProcess) forwardInterimResultsToChannel(build *util_highs.LinearBuilder, resultChannel chan<- HighsMultiResult, printer *util.PrintRecorder) {
	build.SetCallback([]highs.CallbackType{highs.CallbackTypeMipImprovingSolution},
		func(callbackType highs.CallbackType, str string, out highs.CallbackData) highs.CallbackResult {
			if callbackType == highs.CallbackTypeMipImprovingSolution {
				interim := util_highs.InterimSolutionFromCallback(out)
				result, err := process.solutionToResult(interim, printer, true)
				if err != nil {
					util.GlobalWarnHandler(fmt.Errorf("ERROR: interim error: %w", err))
				}
				resultChannel <- *result
			}
			return highs.CallbackResult{}
		},
	)
}

//goland:noinspection GoBoolExpressions
func debugPrintAll(solution *util_highs.Solution2, build *util_highs.LinearBuilder, printer *util.PrintRecorder) error {
	if !util_highs.C_DebugHighs {
		return nil
	}

	printer.Println("SOLUTION STATUS = " + solution.Status().String())
	printer.Printf("OBJECTIVE VALUE %f \n", solution.Objective())

	for columnIndex, outputValue := range solution.ColValuesSeq() {
		context := build.DebugContextFor(columnIndex)
		if colInfo, isInfo := context.(*columnInfo); isInfo {
			err := debugPrintColumnEntry(colInfo, columnIndex, outputValue, nil, printer)
			if err != nil {
				return err
			}
		} else {
			text := build.DebugTextFor(columnIndex)
			printer.Printf("%d %f %s\n", columnIndex, outputValue, text)
		}
	}

	return nil
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

func (process *SolverHighsMultiProcess) solutionToResult(solution util_highs.ISolution, printer *util.PrintRecorder, interim bool) (*HighsMultiResult, error) {
	printer.Println("############################################################################")
	printer.Printf("Make Solution Permute = %s\n", process.permuteLabel)

	resultMap := make(map[string]highsMultiResultEntry)
	for partIndex := range process.parts {
		part := process.parts[partIndex]
		solvedSet, err := part.singleGearSet.buildResultSet(solution, &part.SolverModel)
		if err != nil {
			return nil, err
		}
		if err = validateNewSet(solvedSet, &part.solveOptions, part.SolverModel.CheckSet); err != nil {
			return nil, err
		}
		fullItemSet, err := items.FullItemSet_FromSolved(solvedSet, &part.ItemOptions)
		if err != nil {
			return nil, err
		}

		printer.Printf("Inner solver type = %s\n", reflect.TypeOf(part.singleGearSet).Elem().Name())

		outputId := uuid.NewString()
		if interim {
			outputId += "-interim"
		}
		printer.Printf("OutputId[%s] = %s\n", part.Label, outputId)

		resultMap[part.Label] = highsMultiResultEntry{fullItemSet, outputId}
	}
	printer.Println0()

	return &HighsMultiResult{resultMap, process.permuteLabel, interim}, nil
}

func (process *SolverHighsMultiProcess) makeFullModel(timeLimit int) error {
	process.build = &util_highs.LinearBuilder{}
	process.build.TimeLimitSeconds = timeLimit
	process.build.Solver = util_highs.Solver_MIP_Interior // TODO check if actually best

	entry := columnInfo{entryType: entry_multi_output}
	entry.columnIndex = process.build.CreateColumnWithOutput(highs.Continuous, util_highs.InfNeg(), util_highs.InfPos(), 1, &entry)

	process.outputColumn = entry.columnIndex
	process.outputRow.Add(process.outputColumn, -1)

	for partIndex := range process.parts {
		err := process.parts[partIndex].makeSingleGearSet(process.build, process)
		if err != nil {
			return err
		}
	}

	process.addCommonConstraints(process.build)

	process.outputRow.Build(process.build, 0, 0)

	return nil
}

func (param *SolverHighsMultiParam) makeSingleGearSet(build *util_highs.LinearBuilder, job *SolverHighsMultiProcess) error {
	param.solveOptions = items.SolvableOptionsMap_of(&param.ItemOptions)

	singleGearSet, err := makeGearSetForWeight(build, &param.SolverModel)
	if err != nil {
		return err
	}
	param.singleGearSet = singleGearSet

	setOutput, err := param.singleGearSet.setup(&param.SolverModel, &param.solveOptions)
	if err != nil {
		return err
	}

	job.outputRow.Add(setOutput.columnIndex, param.RatingMultiply/param.singleGearSet.getRatingPreScale())
	return nil
}

func makeGearSetForWeight(build *util_highs.LinearBuilder, model *solve_highs_types.SolverModel) (ISingleGearSet, error) {
	if model.Weights3 != nil {
		return makeGearSetExtended3(build), nil
	} else if model.Weights2 != nil {
		return makeGearSetExtended2(build), nil
	} else if model.Weights1 != nil {
		return makeGearSetBasic(build), nil
	} else {
		return nil, util.ErrorTracedNew("missing weight")
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
			for column := range part.singleGearSet.columnsForItemId(item.ItemId()) {
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
