package withhighs

import (
	"iter"
	"paladin_gearing_go/items"
	gear_model "paladin_gearing_go/model"
	"paladin_gearing_go/util"

	"github.com/lanl/highs"
)

type SolverHighsMultiParam struct {
	Label          string
	ItemOptions    items.FullOptionsMap
	Gear_model     *gear_model.Model
	RatingMultiply uint64

	setup        *setupInputsForSetBonus
	solveOptions items.SolvableOptionsMap
}

type SolverHighsMultiProcess struct {
	common map[items.ItemId][]items.FullItem
	parts  []SolverHighsMultiParam
}

func (job *SolverHighsMultiProcess) AddSetParam(param SolverHighsMultiParam) {
	job.parts = append(job.parts, param)
}

func (job *SolverHighsMultiProcess) SetCommon(common map[items.ItemId][]items.FullItem) {
	job.common = common
}

func (job *SolverHighsMultiProcess) Run(printer *util.PrintRecorder) []items.FullItemSet {
	inputBuilder := inputBuilder{}
	for partIndex := range job.parts {
		job.parts[partIndex].doSetup(&inputBuilder)
	}

	job.addCommonConstraints(&inputBuilder)

	highs_model := inputBuilder.toHighsModel()
	solution, err := highs_model.Solve()
	printer.Println("SOLUTION STATUS = " + solution.Status.String())
	if err != nil {
		panic(err)
	}

	if solution.Status != highs.Optimal && solution.Status != highs.ObjectiveBound && solution.Status != highs.ObjectiveTarget {
		return nil
	}

	resultList := make([]items.FullItemSet, len(job.parts))
	for partIndex := range job.parts {
		part := job.parts[partIndex]
		solvedSet := part.setup.buildResultSet(&solution.Solution, &part.solveOptions, part.Gear_model)
		fullItemSet := items.FullItemSet_FromSolved(solvedSet, &part.ItemOptions)
		resultList[partIndex] = fullItemSet
	}
	return resultList
}

func (param *SolverHighsMultiParam) doSetup(inputBuilder *inputBuilder) {
	param.solveOptions = items.SolvableOptionsMap_of(&param.ItemOptions)
	param.setup = setupBonusedInputs(inputBuilder, param.Gear_model, &param.solveOptions, float64(param.RatingMultiply))
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
