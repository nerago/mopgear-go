package withhighs

import (
	"fmt"
	"iter"
	"paladin_gearing_go/items"
	gear_model "paladin_gearing_go/model"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/util"

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
	common map[items.ItemId][]items.FullItem
	parts  []SolverHighsMultiParam

	outputColumn int
	outputRow    constraintRowBuild

	allColumns []columnInfo
}

func (job *SolverHighsMultiProcess) AddSetParam(param SolverHighsMultiParam) {
	job.parts = append(job.parts, param)
}

func (job *SolverHighsMultiProcess) SetCommon(common map[items.ItemId][]items.FullItem) {
	job.common = common
}

func (job *SolverHighsMultiProcess) Run(printer *util.PrintRecorder) []items.FullItemSet {
	highs_model := job.makeFullModel()
	solution, err := highs_model.Run()
	printer.Println("SOLUTION STATUS = " + solution.Status.String())
	if err != nil {
		panic(err)
	}

	debugPrintAll(solution, job)

	if solution.Status != highs.ModelStatusOptimal {
		return nil
	}

	return job.solutionToResult(solution)
}

func debugPrintAll(solution *highs.Solution, job *SolverHighsMultiProcess) {
	fmt.Println("OBJECTIVE VALUE = ", solution.Objective*c_scaled_ratings)

columnLoop:
	for columnIndex, outputValue := range solution.ColValues {
		if debugPrintColumn(job.allColumns, columnIndex, outputValue, nil, nil) {
			continue columnLoop
		}

		for _, part := range job.parts {
			if debugPrintColumn(part.setup.allColumns, columnIndex, outputValue, nil, nil) {
				continue columnLoop
			}
		}

		fmt.Println(columnIndex, outputValue, "NOT FOUND???")
	}
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

func (job *SolverHighsMultiProcess) makeFullModel() *highs.Solver {
	inputBuilder := inputBuilder{}

	job.outputColumn = inputBuilder.createColumnWithOutput(highs.Continuous, c_minusInf, c_plusInf, 1)
	job.outputRow.add(job.outputColumn, -1)

	entry := columnInfo{entryType: entry_multi_output, columnIndex: job.outputColumn}
	job.allColumns = append(job.allColumns, entry)

	for partIndex := range job.parts {
		job.parts[partIndex].doSetup(&inputBuilder, job)
	}

	job.addCommonConstraints(&inputBuilder)

	job.outputRow.finish(&inputBuilder, 0, 0)

	highs_model := inputBuilder.toHighsModel()
	return highs_model
}

func (job SolverHighsMultiProcess) RunForSeveral(printer *util.PrintRecorder, topN int) [][]items.FullItemSet {
	highs_model := job.makeFullModel()
	solution, err := highs_model.Run()
	printer.Println("SOLUTION STATUS = " + solution.Status.String())
	if err != nil {
		panic(err)
	}
	printer.Println("############################################################################")
	printer.Println("############################################################################")
	printer.Println("############################################################################")

	resultList := make([][]items.FullItemSet, 0, topN)

	jobResult := job.solutionToResult(solution)
	resultList = append(resultList, jobResult)
	solver.ReportSet(printer, jobResult[1], job.parts[1].Gear_model.CalcRatingFull(&jobResult[1]), job.parts[1].Gear_model)
	// previousScore := solution.Objective

	for len(resultList) < topN {
		// highs_model.AddCompSparseRows([]float64{0}, []int{0}, []int{job.outputColumn}, []float64{1}, []float64{previousScore - 1})

		solution, err := highs_model.Run()
		printer.Println("SOLUTION STATUS = " + solution.Status.String())
		if err != nil {
			panic(err)
		}
		printer.Println("############################################################################")
		printer.Println("############################################################################")
		printer.Println("############################################################################")

		jobResult := job.solutionToResult(solution)
		resultList = append(resultList, jobResult)
		solver.ReportSet(printer, jobResult[1], job.parts[1].Gear_model.CalcRatingFull(&jobResult[1]), job.parts[1].Gear_model)
		// previousScore = solution.Objective
	}

	return resultList
}

func (param *SolverHighsMultiParam) doSetup(inputBuilder *inputBuilder, job *SolverHighsMultiProcess) {
	param.solveOptions = items.SolvableOptionsMap_of(&param.ItemOptions)
	param.setup = setupBonusedInputs(inputBuilder, param.Gear_model, &param.solveOptions, 0)
	// param.setup = setupBonusedInputs(inputBuilder, param.Gear_model, &param.solveOptions, float64(param.RatingMultiply))
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
