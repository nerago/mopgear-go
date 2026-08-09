package solver

import (
	"paladin_gearing_go/gear_model"
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/util/util_rank"
	"strconv"
)

func diagnoseFailure(optionsMap *items.SolvableOptionsMap, model *gear_model.SpecModel) (util_collection.Optional[items.SolvableItemSet], string) {
	proposedList := setsAtLimits(optionsMap)
	acceptable := findAcceptableSet(proposedList, model)
	if acceptable.HasValue() {
		return acceptable, ""
	} else {
		message := discoverCommonProblem(proposedList, model.StatRequirements)
		setsAtLimits(optionsMap)
		return util_collection.Optional_Empty[items.SolvableItemSet](), message
	}
}

func setsAtLimits(optionsMap *items.SolvableOptionsMap) []items.SolvableItemSet {
	setList := []items.SolvableItemSet{}
	for _, stat := range stats.StatType_List {
		setList = setsAtLimitsOfStat(optionsMap, stat, setList)
	}
	return setList
}

func setsAtLimitsOfStat(optionsMap *items.SolvableOptionsMap, stat stats.StatType, setList []items.SolvableItemSet) []items.SolvableItemSet {
	var lowSet, highSet items.SolvableItemSet

	for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
		options := optionsMap.Get(slot)
		min, max := findMinMaxWithStat(options, stat)
		lowSet.AddItem_DeferCalc(slot, min)
		highSet.AddItem_DeferCalc(slot, max)
	}

	items.SolvableItemSet_RecalculateTotal(&lowSet)
	items.SolvableItemSet_RecalculateTotal(&highSet)

	return append(setList, lowSet, highSet)
}

func findMinMaxWithStat(options []items.SolvableItem, stat stats.StatType) (*items.SolvableItem, *items.SolvableItem) {
	if len(options) == 0 {
		return nil, nil
	} else if len(options) == 1 {
		return &options[0], &options[0]
	}

	min := &options[0]
	max := &options[0]
	for i := 1; i < len(options); i++ {
		item := &options[i]
		if item.Total().GetUInt(stat) < min.Total().GetUInt(stat) {
			min = item
		}
		if item.Total().GetUInt(stat) > max.Total().GetUInt(stat) {
			max = item
		}
	}
	return min, max
}

func findAcceptableSet(proposedList []items.SolvableItemSet, model *gear_model.SpecModel) util_collection.Optional[items.SolvableItemSet] {
	best := util_rank.BestCollector1[items.SolvableItemSet]{}
	for _, set := range proposedList {
		ok, _ := model.CheckSetForSolver(&set)
		if ok {
			rate := model.CalcRatingSolve(&set, 1)
			best.Offer(&set, rate)
		}
	}
	return best.GetBestOptional()
}

func discoverCommonProblem(proposedList []items.SolvableItemSet, require gear_model.StatRequirements) string {
	var hitLow, hitHigh, expLow, expHigh int
	for _, set := range proposedList {
		if require.IsLow(stats.Stat_Hit, set.Total().Hit()) {
			hitLow++
		}
		if require.IsHigh(stats.Stat_Hit, set.Total().Hit()) {
			hitHigh++
		}
		if require.IsLow(stats.Stat_Expertise, set.Total().Expertise()) {
			expLow++
		}
		if require.IsHigh(stats.Stat_Expertise, set.Total().Expertise()) {
			expHigh++
		}
	}

	hitMessage := ""
	if hitLow == 0 && hitHigh > 0 {
		hitMessage = "hit high"
	} else if hitLow > 0 && hitHigh == 0 {
		hitMessage = "hit low"
	} else if hitLow > 0 && hitHigh > 0 {
		hitMessage = "hit high=" + strconv.FormatInt(int64(hitHigh), 10) + ",low=" + strconv.FormatInt(int64(hitLow), 10)
	}

	expMessage := ""
	if expLow == 0 && expHigh > 0 {
		expMessage = "exp high"
	} else if expLow > 0 && expHigh == 0 {
		expMessage = "exp low"
	} else if expLow > 0 && expHigh > 0 {
		hitMessage = "exp high=" + strconv.FormatInt(int64(expHigh), 10) + ",low=" + strconv.FormatInt(int64(expLow), 10)
	}

	if len(hitMessage) > 0 && len(expMessage) > 0 {
		return hitMessage + ", " + expMessage
	} else if len(hitMessage) > 0 {
		return hitMessage
	} else if len(expMessage) > 0 {
		return expMessage
	} else {
		return "unknown problem"
	}
}
