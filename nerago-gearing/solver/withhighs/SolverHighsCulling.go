package withhighs

import (
	"paladin_gearing_go/items"
	gear_model "paladin_gearing_go/model"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/util"
)

type SolverHighsCullingProcess struct {
	build   *utilhighs.LinearBuilder
	printer *util.PrintRecorder

	setup        *setupInputSetAware
	solveOptions items.SolvableOptionsMap

	itemOptions items.FullOptionsMap
	model       *gear_model.Model

	// common multi_types.CommonOptions
	// parts  []SolverHighsCullingParam

	// outputColumn utilhighs.ColumnIndex
	// outputRow    utilhighs.ConstraintRowBuild

	// allColumns []*columnInfo
}

func (process *SolverHighsCullingProcess) Init(options items.FullOptionsMap, printer *util.PrintRecorder) {
	process.printer = printer
	process.itemOptions = options
}

func (process *SolverHighsCullingProcess) Run() []HighsMultiResult {
	process.printer.Printf("INITIAL MULTI run\n")

	// inputBuilder := new(utilhighs.InputBuilder)
	// solveOptions := items.SolvableOptionsMap_of(&process.itemOptions)
	// initial := setupHighsSetAware(inputBuilder, process.model, &solveOptions, 0)

	// process.makeFullModel()
	// startTime1 := time.Now()
	// solution, log := process.input.RunHighs()
	// printer.Println("Duration! initial = " + time.Since(startTime1).String())
	// printer.AppendOther(log)
	// printer.Println("SOLUTION STATUS = " + solution.Status.String())
	// // debugPrintAll(solution, job, printer)

	// if !solution.HasSolution() {
	// 	return nil
	// }

	// initialResult := process.solutionToResult(solution, printer)
	// bestCommonChoices := process.extractCommonChoices(solution)

	// bestCommonChoices = bestCommonChoices[0:10] // TODO revert

	// printer.Println("############################################################################")
	// printer.Printf("COMMON VARIANT count %d\n", len(bestCommonChoices))
	// printer.Println("############################################################################")

	// resultList := channel_op.Map_SliceToSlice(10, bestCommonChoices, func(changeColumn **columnInfo, resultChannel chan<- HighsMultiResult) {
	// 	innerPrint := util.PrintRecorder_HoldAll()
	// 	printer.Printf("COMMON VARIANT blocking %s\n", (*changeColumn).itemFull.CreateString())

	// 	input := process.input.Clone()
	// 	rowLimitCommon := utilhighs.ConstraintRowBuild{Debug: "rowLimitCommon"}
	// 	rowLimitCommon.Add((*changeColumn).columnIndex, 1)
	// 	rowLimitCommon.Finish(input, 0, 0)

	// 	startTime2 := time.Now()
	// 	solution, log := input.RunHighs()
	// 	printer.Println("Duration! loop = " + time.Since(startTime2).String())
	// 	innerPrint.AppendOther(log)
	// 	innerPrint.Println("SOLUTION STATUS = " + solution.Status.String())

	// 	if solution.HasSolution() {
	// 		jobResult := process.solutionToResult(solution, innerPrint)
	// 		resultChannel <- jobResult
	// 	}

	// 	innerPrint.Println("############################################################################")
	// 	printer.AppendOther(innerPrint)
	// })
	// resultList = append(resultList, initialResult)

	return nil
}
