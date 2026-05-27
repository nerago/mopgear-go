package withhighs

import (
	"iter"
	"math/rand"
	"paladin_gearing_go/items"
	gear_model "paladin_gearing_go/model"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"slices"
	"strconv"
	"time"

	"github.com/bartolsthoorn/gohighs/highs"
	"github.com/google/uuid"
)

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
	input *utilhighs.InputBuilder

	common multi_types.CommonOptions
	parts  []SolverHighsMultiParam

	outputColumn utilhighs.ColumnIndex
	outputRow    utilhighs.ConstraintRowBuild

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
	solution, log := process.input.RunHighs()
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

		printer.Printf("%d %f UNKNOWN\n", columnIndex, outputValue)
	}
}

func (process *SolverHighsMultiProcess) extractCommonChoices(solution *highs.Solution) []*columnInfo {
	commonChosenColumns := make([]*columnInfo, 0, len(process.common))
	for _, jobColumn := range process.allColumns {
		colValue := solution.ColValues[jobColumn.columnIndex]
		if jobColumn.entryType == entry_multi_enable_forge && utilhighs.FloatEqualsOne(colValue) {
			commonChosenColumns = append(commonChosenColumns, jobColumn)
		}
	}
	// TODO could further and block every item used in a set
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
		printer.Printf("OutputId = %s\n", outputId)
		idList = append(idList, outputId)
	}
	return HighsMultiResult{resultList, idList}
}

func (process *SolverHighsMultiProcess) makeFullModel() {
	process.input = &utilhighs.InputBuilder{}

	entry := columnInfo{entryType: entry_multi_output}
	entry.columnIndex = process.input.CreateColumnWithOutput(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, 1, &entry)
	process.allColumns = append(process.allColumns, &entry)

	process.outputColumn = entry.columnIndex
	process.outputRow.Add(process.outputColumn, -1)

	for partIndex := range process.parts {
		process.parts[partIndex].doSetup(process.input, process)
	}

	process.addCommonConstraints(process.input)

	process.outputRow.Finish(process.input, 0, 0)
}

func (process *SolverHighsMultiProcess) RunForSeveral_CommonDifferent_WithParallel(printer *util.PrintRecorder) []HighsMultiResult {
	printer.Printf("INITIAL MULTI run\n")

	process.makeFullModel()
	startTime1 := time.Now()
	solution, log := process.input.RunHighs()
	printer.Println("Duration! initial = " + time.Since(startTime1).String())
	printer.AppendOther(log)
	printer.Println("SOLUTION STATUS = " + solution.Status.String())
	// debugPrintAll(solution, job, printer)

	if !solution.HasSolution() {
		return nil
	}

	initialResult := process.solutionToResult(solution, printer)
	bestCommonChoices := process.extractCommonChoices(solution)

	printer.Println("############################################################################")

	resultList := channel_op.Map_SliceToSlice(10, bestCommonChoices, func(changeColumn **columnInfo, resultChannel chan<- HighsMultiResult) {
		innerPrint := util.PrintRecorder_HoldAll()
		printer.Printf("COMMON VARIANT blocking %s\n", (*changeColumn).itemFull.CreateString())

		input := process.input.Clone()
		rowLimitCommon := utilhighs.ConstraintRowBuild{Debug: "rowLimitCommon"}
		rowLimitCommon.Add((*changeColumn).columnIndex, 1)
		rowLimitCommon.Finish(input, 0, 0)

		startTime2 := time.Now()
		solution, log := input.RunHighs()
		printer.Println("Duration! loop = " + time.Since(startTime2).String())
		innerPrint.AppendOther(log)
		innerPrint.Println("SOLUTION STATUS = " + solution.Status.String())

		if solution.HasSolution() {
			jobResult := process.solutionToResult(solution, innerPrint)
			resultChannel <- jobResult
		}

		innerPrint.Println("############################################################################")
		printer.AppendOther(innerPrint)
	})
	resultList = append(resultList, initialResult)

	return resultList
}

func (process *SolverHighsMultiProcess) RunForSeveral_CommonDifferent_Sampling(printer *util.PrintRecorder, outputTarget int) []HighsMultiResult {
	printer.Printf("INITIAL MULTI run\n")

	process.makeFullModel()
	solution := process.input.RunHighsThenDiagnose(printer)
	printer.Println("SOLUTION STATUS = " + solution.Status.String())

	if !solution.HasSolution() {
		return nil
	}

	initialResult := process.solutionToResult(solution, printer)
	resultList := make([]HighsMultiResult, 0, outputTarget)
	resultList = append(resultList, initialResult)

	bestCommonChoices := process.extractCommonChoices(solution)
	checkedIndexes := make([]int, 0, outputTarget)

	if len(bestCommonChoices) == 0 {
		printer.Println("WARNWARNWARNWARNWARNWARNWARNWARN")
		printer.Println("WARNWARNWARNWARNWARNWARNWARNWARN")
		printer.Println("WARNWARNWARNWARNWARNWARNWARNWARN")
		printer.Println("WARN no common choices found WAR")
		printer.Println("WARNWARNWARNWARNWARNWARNWARNWARN")
		printer.Println("WARNWARNWARNWARNWARNWARNWARNWARN")
		printer.Println("WARNWARNWARNWARNWARNWARNWARNWARN")
		return resultList
	}

	printer.Println("############################################################################")

	for len(resultList) < outputTarget {
		var tryIndex int
		for {
			tryIndex = rand.Intn(len(bestCommonChoices))
			if !slices.Contains(checkedIndexes, tryIndex) {
				break
			}
		}
		checkedIndexes = append(checkedIndexes, tryIndex)
		changeColumn := bestCommonChoices[tryIndex]

		innerPrint := util.PrintRecorder_HoldAll()
		printer.Printf("COMMON VARIANT blocking %s\n", changeColumn.itemFull.CreateString())

		input := process.input.Clone()
		rowLimitCommon := utilhighs.ConstraintRowBuild{Debug: "rowLimitCommon"}
		rowLimitCommon.Add(changeColumn.columnIndex, 1)
		rowLimitCommon.Finish(input, 0, 0)

		solution, log := input.RunHighs()
		innerPrint.AppendOther(log)
		innerPrint.Println("SOLUTION STATUS = " + solution.Status.String())

		if solution.HasSolution() {
			jobResult := process.solutionToResult(solution, innerPrint)
			resultList = append(resultList, jobResult)
		}

		innerPrint.Println("############################################################################")
		printer.AppendOther(innerPrint)
	}

	return resultList
}

func (param *SolverHighsMultiParam) doSetup(inputBuilder *utilhighs.InputBuilder, job *SolverHighsMultiProcess) {
	param.solveOptions = items.SolvableOptionsMap_of(&param.ItemOptions)
	param.setup = setupHighsSetAware(inputBuilder, param.Gear_model, &param.solveOptions, 0)
	job.outputRow.Add(param.setup.mainOutputVar.columnIndex, param.RatingMultiply)
}

func (process *SolverHighsMultiProcess) addCommonConstraints(inputBuilder *utilhighs.InputBuilder) {
	for itemId, array := range process.common {
		process.addCommonConstraintsForItem(inputBuilder, itemId, array)
	}
}

func (process *SolverHighsMultiProcess) addCommonConstraintsForItem(inputBuilder *utilhighs.InputBuilder, itemId items.ItemId, array []items.FullItem) {
	onlyOneReforge := utilhighs.ConstraintRowBuild{Debug: "onlyOneReforge" + itemId.String()}

	for _, item := range array {
		entryEnableReforge := columnInfo{entryType: entry_multi_enable_forge, itemFull: &item}
		entryEnableReforge.columnIndex = inputBuilder.CreateColumnBool(&entryEnableReforge)
		process.allColumns = append(process.allColumns, &entryEnableReforge)

		onlyOneReforge.Add(entryEnableReforge.columnIndex, 1)

		for partUsedItem := range process.findMatchingItemColumns(&item) {
			// formula is partUsedItem <= enableReforge
			//            0 <= enableReforge - partUsedItem
			matchingReforge := utilhighs.ConstraintRowBuild{Debug: "matchingReforge" + itemId.String() + "_" + strconv.Itoa(int(partUsedItem))}
			matchingReforge.Add(entryEnableReforge.columnIndex, 1)
			matchingReforge.Add(partUsedItem, -1)
			matchingReforge.Finish(inputBuilder, 0, 1)
		}
	}

	onlyOneReforge.Finish(inputBuilder, 0, 1)
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
