package withhighs

import (
	"iter"
	"math/rand"
	"paladin_gearing_go/items"
	gear_model "paladin_gearing_go/model"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"slices"
	"time"

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

func (process *SolverHighsMultiProcess) AddSetParam(param SolverHighsMultiParam) {
	process.parts = append(process.parts, param)
}

func (process *SolverHighsMultiProcess) SetCommon(common multi_types.CommonOptions) {
	process.common = common
}

func (process *SolverHighsMultiProcess) Run(printer *util.PrintRecorder) []items.FullItemSet {
	process.makeFullModel()
	solution, log := process.input.runHighs()
	printer.AppendOther(log)
	printer.Println("SOLUTION STATUS = " + solution.Status.String())

	debugPrintAll(solution, process, printer)

	if solution.HasSolution() {
		return process.solutionToResult(solution)
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

func (process *SolverHighsMultiProcess) extractCommonChoices(solution *highs.Solution) []columnInfo {
	commonChosenColumns := make([]columnInfo, 0, len(process.common))
	for _, jobColumn := range process.allColumns {
		colValue := solution.ColValues[jobColumn.columnIndex]
		if jobColumn.entryType == entry_multi_enable_forge && floatEqualsOne(colValue) {
			commonChosenColumns = append(commonChosenColumns, jobColumn)
		}
	}
	// TODO could further and block every item used in a set
	return commonChosenColumns
}

func (process *SolverHighsMultiProcess) solutionToResult(solution *highs.Solution) []items.FullItemSet {
	resultList := make([]items.FullItemSet, len(process.parts))
	for partIndex := range process.parts {
		part := process.parts[partIndex]
		solvedSet := part.setup.buildResultSet(solution, &part.solveOptions, part.Gear_model)
		fullItemSet := items.FullItemSet_FromSolved(solvedSet, &part.ItemOptions)
		resultList[partIndex] = fullItemSet
	}
	return resultList
}

func (process *SolverHighsMultiProcess) makeFullModel() {
	process.input = &inputBuilder{}

	process.outputColumn = process.input.createColumnWithOutput(highs.Continuous, c_minusInf, c_plusInf, 1)
	process.outputRow.add(process.outputColumn, -1)

	entry := columnInfo{entryType: entry_multi_output, columnIndex: process.outputColumn}
	process.allColumns = append(process.allColumns, entry)

	for partIndex := range process.parts {
		process.parts[partIndex].doSetup(process.input, process)
	}

	process.addCommonConstraints(process.input)

	process.outputRow.finish(process.input, 0, 0)
}

func (process *SolverHighsMultiProcess) RunForSeveral_CommonDifferent_WithParallel(printer *util.PrintRecorder) [][]items.FullItemSet {
	printer.Printf("INITIAL MULTI run\n")

	process.makeFullModel()
	startTime1 := time.Now()
	solution, log := process.input.runHighs()
	printer.Println("Duration! initial = " + time.Since(startTime1).String())
	printer.AppendOther(log)
	printer.Println("SOLUTION STATUS = " + solution.Status.String())
	// debugPrintAll(solution, job, printer)

	if !solution.HasSolution() {
		return nil
	}

	initialResult := process.solutionToResult(solution)
	bestCommonChoices := process.extractCommonChoices(solution)

	printer.Println("############################################################################")

	resultList := channel_op.Map_SliceToSlice(10, bestCommonChoices, func(changeColumn *columnInfo, resultChannel chan<- []items.FullItemSet) {
		innerPrint := util.PrintRecorder_HoldAll()
		printer.Printf("COMMON VARIANT blocking %s\n", changeColumn.itemFull.CreateString())

		input := process.input.clone()
		rowLimitCommon := constraintRowBuild{}
		rowLimitCommon.add(changeColumn.columnIndex, 1)
		rowLimitCommon.finish(input, 0, 0)

		startTime2 := time.Now()
		solution, log := input.runHighs()
		printer.Println("Duration! loop = " + time.Since(startTime2).String())
		innerPrint.AppendOther(log)
		innerPrint.Println("SOLUTION STATUS = " + solution.Status.String())

		if solution.HasSolution() {
			jobResult := process.solutionToResult(solution)
			resultChannel <- jobResult
		}

		innerPrint.Println("############################################################################")
		printer.AppendOther(innerPrint)
	})
	resultList = append(resultList, initialResult)

	return resultList
}


func (process *SolverHighsMultiProcess) RunForSeveral_CommonDifferent_Sampling(printer *util.PrintRecorder, outputTarget int) [][]items.FullItemSet {
	printer.Printf("INITIAL MULTI run\n")

	process.makeFullModel()
	solution, log := process.input.runHighs()
	printer.AppendOther(log)
	printer.Println("SOLUTION STATUS = " + solution.Status.String())

	if !solution.HasSolution() {
		return nil
	}

	initialResult := process.solutionToResult(solution)
	resultList := make([][]items.FullItemSet, 0, outputTarget)
	resultList = append(resultList, initialResult)

	bestCommonChoices := process.extractCommonChoices(solution)
	checkedIndexes := make([]int, 0, outputTarget)

	printer.Println("############################################################################")

	for len(resultList) < outputTarget {
		var tryIndex int
		for {
			tryIndex = rand.Intn(len(bestCommonChoices))
			if !slices.Contains(checkedIndexes, tryIndex) { break }
		}
		checkedIndexes = append(checkedIndexes, tryIndex)
		changeColumn := bestCommonChoices[tryIndex]

		innerPrint := util.PrintRecorder_HoldAll()
		printer.Printf("COMMON VARIANT blocking %s\n", changeColumn.itemFull.CreateString())

		input := process.input.clone()
		rowLimitCommon := constraintRowBuild{}
		rowLimitCommon.add(changeColumn.columnIndex, 1)
		rowLimitCommon.finish(input, 0, 0)

		solution, log := input.runHighs()
		innerPrint.AppendOther(log)
		innerPrint.Println("SOLUTION STATUS = " + solution.Status.String())

		if solution.HasSolution() {
			jobResult := process.solutionToResult(solution)
			resultList = append(resultList, jobResult)
		}

		innerPrint.Println("############################################################################")
		printer.AppendOther(innerPrint)
	}

	return resultList
}

func (process *SolverHighsMultiProcess) RunForSeveral_NextObjective(printer *util.PrintRecorder, targetCount int) [][]items.FullItemSet {
	printer.Printf("INITIAL MULTI run\n")

	process.makeFullModel()

	startTime1 := time.Now()
	printer.Println("Started at " + startTime1.Format(time.RFC1123))

	solver, logFilename := process.input.preHighsRun()
	solution, err := highsPool.RunSolverUnderMutex(solver)
	checkError(err)

	printer.Println("Duration initial = " + time.Since(startTime1).String())
	printer.Println("SOLUTION STATUS = " + solution.Status.String())

	initialResult := process.solutionToResult(solution)
	// initialObjective := solution.Objective

	resultList := make([][]items.FullItemSet, 0)
	resultList = append(resultList, initialResult)

	if !solution.HasSolution() {
		return nil
	}

	bestCommonChoices := process.extractCommonChoices(solution)
	randIndex := rand.Intn(len(bestCommonChoices))
	changeColumn := bestCommonChoices[randIndex].columnIndex
	checkError(solver.AddRow(0, 0, []int{changeColumn}, []float64{1}))
	// rowLimitCommon := constraintRowBuild{}
	// 	rowLimitCommon.add(, 1)
	// 	rowLimitCommon.finish(input, 0, 0)

	// checkError(solver.SetColBounds(process.outputColumn, 0, initialObjective * 0.99999))
	// checkError(solver.AddRow(0, initialObjective * 0.99999, []int{process.outputColumn}, []float64{1}))
	// checkError(solver.ClearSolver())

	startTime2 := time.Now()

	solution, err = highsPool.RunSolverUnderMutex(solver)
	checkError(err)

	printer.Println("Duration second = " + time.Since(startTime2).String())
	printer.Println("SOLUTION STATUS = " + solution.Status.String())

	secondResult := process.solutionToResult(solution)
	resultList = append(resultList, secondResult)

	log := process.input.postHighsRun(solver, logFilename)
	printer.AppendOther(log)

	return resultList
}

func (param *SolverHighsMultiParam) doSetup(inputBuilder *inputBuilder, job *SolverHighsMultiProcess) {
	param.solveOptions = items.SolvableOptionsMap_of(&param.ItemOptions)
	param.setup = setupBonusedInputs(inputBuilder, param.Gear_model, &param.solveOptions, 0)
	job.outputRow.add(param.setup.mainOutputVar.columnIndex, param.RatingMultiply)
}

func (process *SolverHighsMultiProcess) addCommonConstraints(inputBuilder *inputBuilder) {
	for _, array := range process.common {
		process.addCommonConstraintsForItem(inputBuilder, array)
	}
}

func (process *SolverHighsMultiProcess) addCommonConstraintsForItem(inputBuilder *inputBuilder, array []items.FullItem) {
	onlyOneReforge := constraintRowBuild{}

	for _, item := range array {
		enableReforge := inputBuilder.createColumnBool()
		onlyOneReforge.add(enableReforge, 1)

		for partUsedItem := range process.findMatchingItemColumns(&item) {
			// formula is partUsedItem <= enableReforge
			//            0 <= enableReforge - partUsedItem
			matchingReforge := constraintRowBuild{}
			matchingReforge.add(enableReforge, 1)
			matchingReforge.add(partUsedItem, -1)
			matchingReforge.finish(inputBuilder, 0, 1)
		}

		entry := columnInfo{entryType: entry_multi_enable_forge, columnIndex: enableReforge, itemFull: &item}
		process.allColumns = append(process.allColumns, entry)
	}

	onlyOneReforge.finish(inputBuilder, 1, 1)
}

func (process *SolverHighsMultiProcess) findMatchingItemColumns(item *items.FullItem) iter.Seq[int] {
	return func(yield func(int) bool) {
		for _, part := range process.parts {
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
