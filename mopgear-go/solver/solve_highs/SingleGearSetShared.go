package solve_highs

import (
	"fmt"
	"iter"

	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/solver/solve_highs_types"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_highs"
)

type singleGearSetShared struct {
	build *util_highs.LinearBuilder

	ratingPreScale float64

	itemSetupCommon   gearItemSetupShared
	bonusComboHandler gearBonusComboHandler
}

func (sc *singleGearSetShared) runForFutureResult(itemOptions *items.SolvableOptionsMap, model *solve_highs_types.SolverModel, printer *util.PrintRecorder) *util_async.FutureCancellableWithError[items.SolvableItemSet] {
	solutionFuture := sc.build.RunHighsFuture(nil)

	return util_async.FutureCancellable_MapValueError(solutionFuture, func(result util_highs.LinearResult) (*items.SolvableItemSet, error) {
		solution, err := result.GetSolution2AndSaveLog(printer)
		if err != nil {
			return nil, err
		}

		printer.Printf("SOLUTION STATUS = %s\n", solution.Status().String())
		if err = debugPrint(solution, sc.build, printer); err != nil {
			return nil, err
		}

		if solution.HasSolution() {
			itemSet, err := sc.buildResultSet(solution, model)
			if err != nil {
				return nil, err
			}

			if err = validateNewSet(itemSet, itemOptions, model.CheckSet); err != nil {
				return nil, err
			}

			if err = sc.checkSetRatingIsObjective(solution, &itemSet, model.CalcRatingSet); err != nil {
				return nil, err
			}

			return new(itemSet), nil
		} else {
			return nil, util.ErrorTracedNew("highs solve status " + solution.Status().String())
		}
	})
}

func (sc *singleGearSetShared) checkSetRatingIsObjective(solution *util_highs.Solution2, itemSet *items.SolvableItemSet, calcRating func(item *items.SolvableItemSet) float64) error {
	checkRating := calcRating(itemSet)
	if !util.FloatsApproxEqualsLenient(solution.Objective()/sc.ratingPreScale, checkRating) {
		return util.ErrorTracedNew(fmt.Sprintf("rating inconsistent %e %e ", solution.Objective(), checkRating))
	}
	return nil
}

func (sc *singleGearSetShared) getRatingPreScale() float64 {
	return sc.ratingPreScale
}

func (sc *singleGearSetShared) columnsForItemId(itemId items.ItemId) iter.Seq[*columnInfo] {
	return sc.itemSetupCommon.itemColumns.ValuesForKeyAsSeq(itemId)
}

func (sc *singleGearSetShared) createItemColumn(entry *columnInfo) {
	entry.columnIndex = sc.build.CreateColumnBool(entry)
}

func (sc *singleGearSetShared) buildResultSet(solution util_highs.ISolution, model *solve_highs_types.SolverModel) (items.SolvableItemSet, error) {
	itemSet := items.SolvableItemSet{}
	for columnEntry := range sc.itemSetupCommon.itemColumns.SeqValues() {
		isTrue := solution.ValueIsOne(columnEntry.columnIndex)
		if columnEntry.entryType == entry_item && isTrue {
			err := itemSet.AddItem_DeferCalc_ExpectEmpty(columnEntry.itemSlot, columnEntry.item)
			if err != nil {
				return items.SolvableItemSet{}, err
			}
		}
	}
	items.SolvableItemSet_RecalculateTotal(&itemSet)

	err := sc.bonusComboHandler.checkActiveCombo(solution, &itemSet, model)
	return itemSet, err
}
