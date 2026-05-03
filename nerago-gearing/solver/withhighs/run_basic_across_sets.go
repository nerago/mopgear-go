package withhighs

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_rank"
)

func RunBasicAcrossSets_ReturnBest(itemOptions *items.SolvableOptionsMap, model *model.Model) util.Optional[items.SolvableItemSet] {
	best := util_rank.BestCollector1[items.SolvableItemSet]{}

	unconstrainedResult := RunBasic(itemOptions, model, nil, util.Optional_Empty[int]())

	if unconstrainedResultSet, found := unconstrainedResult.GetWithFlag(); found {
		if model.CheckSet(&unconstrainedResultSet) {
			best.Offer(&unconstrainedResultSet, model.CalcRatingSolve(&unconstrainedResultSet))
		}
	}

	for _, activeSet := range model.SetBonus.ActiveSets() {
		for requireSetCount := 1; requireSetCount <= 5; requireSetCount++ {
			otherResult := RunBasic(itemOptions, model, activeSet, util.Optional_OfValue(requireSetCount))

			if otherResultSet, found := otherResult.GetWithFlag(); found {
				if model.CheckSet(&otherResultSet) {
					best.Offer(&otherResultSet, model.CalcRatingSolve(&otherResultSet))
				}
			}
		}
	}

	return best.GetBestOptional()
}

func RunBasicAcrossSets_ReturnAll(itemOptions *items.SolvableOptionsMap, model *model.Model, printer *util.PrintRecorder) []items.SolvableItemSet {
	resultList := []items.SolvableItemSet{}

	printer.Println("{{{{{{{{ run unconstrained }}}}}}}}")
	unconstrainedResult := RunBasic(itemOptions, model, nil, util.Optional_Empty[int]())
	if unconstrainedResultSet, found := unconstrainedResult.GetWithFlag(); found {
		if model.CheckSet(&unconstrainedResultSet) {
			resultList = append(resultList, unconstrainedResultSet)
		}
	}

	for _, activeSet := range model.SetBonus.ActiveSets() {
		for requireSetCount := 1; requireSetCount <= 5; requireSetCount++ {

			printer.Printf("{{{{{{{{ run %s %d }}}}}}}}\n", activeSet.Name(), requireSetCount)
			otherResult := RunBasic(itemOptions, model, activeSet, util.Optional_OfValue(requireSetCount))
			if otherResultSet, found := otherResult.GetWithFlag(); found {
				if model.CheckSet(&otherResultSet) {
					resultList = append(resultList, otherResultSet)
				}
			}
		}
	}

	util.RemoveDuplicatesComparable(resultList)

	return resultList
}
