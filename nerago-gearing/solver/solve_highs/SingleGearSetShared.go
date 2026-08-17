package solve_highs

import (
	"fmt"
	"iter"
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver/solve_highs_types"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_highs"
)

type singleGearSetShared struct {
	build *util_highs.LinearBuilder

	ratingPreScale float64

	itemSetupCommon   gearItemSetupShared
	bonusComboHandler gearBonusComboHandler
}

func (sc *singleGearSetShared) runForFutureResult(itemOptions *items.SolvableOptionsMap, model *solve_highs_types.SolverModel, printer *util.PrintRecorder) *util_async.FutureCancellable[items.SolvableItemSet] {
	solutionFuture := sc.build.RunHighsFuture(nil)

	return util_async.FutureCancellable_MapValue(solutionFuture, func(result util_highs.LinearResult) (items.SolvableItemSet, bool) {
		solution := result.GetSolution2AndSaveLog(printer)

		printer.Printf("SOLUTION STATUS = %s\n", solution.Status().String())
		debugPrint(solution, sc.build, printer)

		if solution.HasSolution() {
			itemSet := sc.buildResultSet(solution, model)
			validateNewSet(itemSet, itemOptions, model.CheckSet)
			sc.checkSetRatingIsObjective(solution, &itemSet, model.CalcRatingSet)
			return itemSet, true
		} else {
			return items.SolvableItemSet{}, false
		}
	})
}

func (sc *singleGearSetShared) checkSetRatingIsObjective(solution *util_highs.Solution2, itemSet *items.SolvableItemSet, calcRating func(item *items.SolvableItemSet) float64) {
	checkRating := calcRating(itemSet)
	if !util.FloatsApproxEqualsLenient(solution.Objective()/sc.ratingPreScale, checkRating) {
		panic(fmt.Sprintf("rating inconsistent %e %e ", solution.Objective(), checkRating))
	}
}

func (sc *singleGearSetShared) ColumnsForItemId(itemId items.ItemId) iter.Seq[*columnInfo] {
	return sc.itemSetupCommon.itemColumns.ValuesForKeyAsSeq(itemId)
}

func (sc *singleGearSetShared) RatingPreScale() float64 {
	return sc.ratingPreScale
}

func (sc *singleGearSetShared) createItemColumn(entry *columnInfo) {
	entry.columnIndex = sc.build.CreateColumnBool(entry)
}

func (sc *singleGearSetShared) buildResultSet(solution util_highs.ISolution, model *solve_highs_types.SolverModel) items.SolvableItemSet {
	itemSet := items.SolvableItemSet{}
	for columnEntry := range sc.itemSetupCommon.itemColumns.SeqValues() {
		isTrue := solution.ValueIsOne(columnEntry.columnIndex)
		if columnEntry.entryType == entry_item && isTrue {
			itemSet.AddItem_DeferCalc_ExpectEmpty(columnEntry.itemSlot, columnEntry.item)
		}
	}
	items.SolvableItemSet_RecalculateTotal(&itemSet)

	sc.bonusComboHandler.CheckActiveCombo(solution, &itemSet, model)
	return itemSet
}
