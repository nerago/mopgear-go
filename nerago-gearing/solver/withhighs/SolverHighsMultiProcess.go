package withhighs

import (
	"iter"
	"paladin_gearing_go/items"
	gear_model "paladin_gearing_go/model"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"

	"github.com/bartolsthoorn/gohighs/highs"
)

type SolverHighsMultiParam struct {
	Label          string
	ItemOptions    items.FullOptionsMap
	Gear_model     *gear_model.Model
	RatingMultiply float64

	setup        *setupInputsForSetBonus
	solveOptions items.SolvableOptionsMap
}

type SolverHighsMultiProcess struct {
	input *inputBuilder

	common multi_types.CommonOptions
	parts  []SolverHighsMultiParam

	outputColumn int
	outputRow    constraintRowBuild

	allColumns []columnInfo
}

func (job *SolverHighsMultiProcess) AddSetParam(param SolverHighsMultiParam) {
	job.parts = append(job.parts, param)
}

func (job *SolverHighsMultiProcess) SetCommon(common multi_types.CommonOptions) {
	job.common = common
}

func (job *SolverHighsMultiProcess) Run(printer *util.PrintRecorder) []items.FullItemSet {
	job.makeFullModel()
	solution, log := job.input.runHighs()
	printer.AppendOther(log)
	printer.Println("SOLUTION STATUS = " + solution.Status.String())

	debugPrintAll(solution, job, printer)

	if solution.HasSolution() {
		return job.solutionToResult(solution)
	} else {
		return nil
	}
}

func debugPrintAll(solution *highs.Solution, job *SolverHighsMultiProcess, printer *util.PrintRecorder) {
	if !c_debugHighs {
		return
	}

	printer.Printf("OBJECTIVE VALUE %f \n", solution.Objective*c_scaled_ratings)

columnLoop:
	for columnIndex, outputValue := range solution.ColValues {
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

func (job *SolverHighsMultiProcess) extractCommonChoices(solution *highs.Solution) []columnInfo {
	commonChosenColumns := make([]columnInfo, 0, len(job.common))
	for _, jobColumn := range job.allColumns {
		colValue := solution.ColValues[jobColumn.columnIndex]
		if jobColumn.entryType == entry_multi_enable_forge && floatEqualsOne(colValue) {
			commonChosenColumns = append(commonChosenColumns, jobColumn)
		}
	}
	return commonChosenColumns
}

func (job *SolverHighsMultiProcess) solutionToResult(solution *highs.Solution) []items.FullItemSet {
	resultList := make([]items.FullItemSet, len(job.parts))
	for partIndex := range job.parts {
		part := job.parts[partIndex]
		solvedSet := part.setup.buildResultSet(solution, &part.solveOptions, part.Gear_model)
		fullItemSet := items.FullItemSet_FromSolved(solvedSet, &part.ItemOptions)
		resultList[partIndex] = fullItemSet
	}
	return resultList
}

func (job *SolverHighsMultiProcess) makeFullModel() {
	job.input = &inputBuilder{}

	job.outputColumn = job.input.createColumnWithOutput(highs.Continuous, c_minusInf, c_plusInf, 1)
	job.outputRow.add(job.outputColumn, -1)

	entry := columnInfo{entryType: entry_multi_output, columnIndex: job.outputColumn}
	job.allColumns = append(job.allColumns, entry)

	for partIndex := range job.parts {
		job.parts[partIndex].doSetup(job.input, job)
	}

	job.addCommonConstraints(job.input)

	job.outputRow.finish(job.input, 0, 0)
}

func (job *SolverHighsMultiProcess) RunForSeveral_CommonDifferent_WithParallel(printer *util.PrintRecorder) [][]items.FullItemSet {
	printer.Printf("INITIAL MULTI run\n")

	job.makeFullModel()
	solution, log := job.input.runHighs()
	printer.AppendOther(log)
	printer.Println("SOLUTION STATUS = " + solution.Status.String())
	// debugPrintAll(solution, job, printer)

	if !solution.HasSolution() {
		return nil
	}

	initialResult := job.solutionToResult(solution)
	bestCommonChoices := job.extractCommonChoices(solution)

	printer.Println("############################################################################")

	resultList := channel_op.Map_SliceToSlice(c_threads, bestCommonChoices, func(changeColumn *columnInfo, resultChannel chan<- []items.FullItemSet) {
		innerPrint := util.PrintRecorder_HoldAll()
		printer.Printf("COMMON VARIANT blocking %s\n", changeColumn.itemFull.CreateString())

		input := job.input.clone()
		rowLimitCommon := constraintRowBuild{}
		rowLimitCommon.add(changeColumn.columnIndex, 1)
		rowLimitCommon.finish(input, 0, 0)

		solution, log := input.runHighs()
		innerPrint.AppendOther(log)
		innerPrint.Println("SOLUTION STATUS = " + solution.Status.String())

		if solution.HasSolution() {
			jobResult := job.solutionToResult(solution)
			resultChannel <- jobResult
		}

		innerPrint.Println("############################################################################")
		printer.AppendOther(innerPrint)
	})
	resultList = append(resultList, initialResult)

	return resultList
}

func (param *SolverHighsMultiParam) doSetup(inputBuilder *inputBuilder, job *SolverHighsMultiProcess) {
	param.solveOptions = items.SolvableOptionsMap_of(&param.ItemOptions)
	param.setup = setupBonusedInputs(inputBuilder, param.Gear_model, &param.solveOptions, 0)
	job.outputRow.add(param.setup.mainOutputVar.columnIndex, param.RatingMultiply)
}

func (job *SolverHighsMultiProcess) addCommonConstraints(inputBuilder *inputBuilder) {
	for _, array := range job.common {
		job.addCommonConstraintsForItem(inputBuilder, array)
	}
}

func (job *SolverHighsMultiProcess) addCommonConstraintsForItem(inputBuilder *inputBuilder, array []items.FullItem) {
	onlyOneReforge := constraintRowBuild{}

	for _, item := range array {
		enableReforge := inputBuilder.createColumnBool()
		onlyOneReforge.add(enableReforge, 1)

		for partUsedItem := range job.findMatchingItemColumns(&item) {
			// formula is partUsedItem <= enableReforge
			//            0 <= enableReforge - partUsedItem
			matchingReforge := constraintRowBuild{}
			matchingReforge.add(enableReforge, 1)
			matchingReforge.add(partUsedItem, -1)
			matchingReforge.finish(inputBuilder, 0, 1)
		}

		entry := columnInfo{entryType: entry_multi_enable_forge, columnIndex: enableReforge, itemFull: &item}
		job.allColumns = append(job.allColumns, entry)
	}

	onlyOneReforge.finish(inputBuilder, 1, 1)
}

func (job *SolverHighsMultiProcess) findMatchingItemColumns(item *items.FullItem) iter.Seq[int] {
	return func(yield func(int) bool) {
		for _, part := range job.parts {
			for _, column := range part.setup.itemColumns {
				if column.item.EqualsFull(item) {
					if !yield(column.columnIndex) {
						return
					}
				}
			}
		}
	}
}
