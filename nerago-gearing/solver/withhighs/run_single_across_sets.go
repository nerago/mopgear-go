package withhighs

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_rank"
)

func RunSingleAcrossSets_ReturnBest(itemOptions *items.SolvableOptionsMap, model *model.Model, printer *util.PrintRecorder) util.Optional[items.SolvableItemSet] {
	best := util_rank.BestCollector1[items.SolvableItemSet]{}

	all := RunSingleAcrossSets_ReturnAll(itemOptions, model, printer)
	for _, itemSet := range all {
		if model.CheckSet(&itemSet) {
			best.Offer(&itemSet, model.CalcRatingSolveAsFloat(&itemSet))
		}
	}

	return best.GetBestOptional()
}

func RunSingleAcrossSets_ReturnAll(itemOptions *items.SolvableOptionsMap, model *model.Model, printer *util.PrintRecorder) []items.SolvableItemSet {
	resultList := []items.SolvableItemSet{}

	printer.Println("{{{{{{{{ run unconstrained }}}}}}}}")
	unconstrainedResult := RunSingle(itemOptions, model, nil, printer)
	if unconstrainedResultSet, found := unconstrainedResult.GetWithFlag(); found {
		if model.CheckSet(&unconstrainedResultSet) {
			resultList = append(resultList, unconstrainedResultSet)
		}
	}

	setData := []setInfo{}
	for _, activeSet := range model.SetBonus.ActiveSets() {
		setData = append(setData, setInfo{activeSet: activeSet})
	}
	permuted := makeSetPermutations(setData) // reuses method from SetBonus process even though type not fully applicable

	for _, permute := range permuted {
		required := make(RequiredSetCounts)
		for _, setWithCount := range permute.content {
			set := setWithCount.setInfo.activeSet
			count := setWithCount.count
			required[set] = count
		}

		printer.Printf("{{{{{{{{ run %s }}}}}}}}\n", permute.debugStr())
		otherResult := RunSingle(itemOptions, model, required, printer)
		if otherResultSet, found := otherResult.GetWithFlag(); found {
			if model.CheckSet(&otherResultSet) {
				resultList = append(resultList, otherResultSet)
			}
		}
	}

	util.RemoveDuplicatesComparable(resultList)

	return resultList
}
