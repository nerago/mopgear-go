package items

import (
	"strconv"
	"strings"
)

func SolvableItem_Of(item FullItem) SolvableItem {
	return SolvableItem{
		item.Ref.ItemId,
		// item.ref.itemLevel,
		// item.slot,
		// item.reforge,
		// item.gemChoice,
		item.TotalCap,
		item.TotalRated}
}

func SolvableOptionsMap_of(fullMap *FullOptionsMap) SolvableOptionsMap {
	result := SolvableOptionsMap{}
	for slot := range fullMap {
		fullArray := fullMap[slot]
		solveArray := make([]SolvableItem, 0, len(fullArray))
		for _, item := range fullArray {
			solveItem := SolvableItem_Of(item)
			solveArray = append(solveArray, solveItem)
		}
		result[slot] = solveArray
	}
	return result
}

func FullItemSet_FromSolved(solvedSet SolvableItemSet, optionsMap *FullOptionsMap) FullItemSet {
	fullMap := FullEquipMap{}
	for slot, solveItem := range solvedSet.Items {
		if !solveItem.IsEmpty() {
			fullItem := findMatch(optionsMap[slot], solveItem)
			fullMap[slot] = fullItem
		}
	}
	return FullItemSet{Items: fullMap, TotalCap: solvedSet.TotalCap, TotalRated: solvedSet.TotalRated}
}

func findMatch(fullItem []FullItem, solveItem *SolvableItem) *FullItem {
	for _, item := range fullItem {
		if isMatch(&item, solveItem) {
			return &item
		}
	}
	panic("match not found")
}

func isMatch(fullItem *FullItem, solveItem *SolvableItem) bool {
	// TODO is it okay to not check item level
	return fullItem.ItemId() == solveItem.ItemId &&
		fullItem.TotalCap == solveItem.TotalCap &&
		fullItem.TotalRated == solveItem.TotalRated
}

func (item *FullItem) String() string {
	build := strings.Builder{}
	build.WriteString("{ ")
	build.WriteString(item.Slot.Name())

	build.WriteString(" \"")
	build.WriteString(item.FullName())

	build.WriteString("\" id=")
	build.WriteString(strconv.FormatUint(uint64(item.ItemId()), 10))

	build.WriteString(" lvl=")
	build.WriteString(strconv.FormatUint(uint64(item.Ref.ItemLevel), 10))
	build.WriteRune(' ')

	build.WriteString(item.StatBase.String())

	if !item.StatEnchant.IsEmpty() {
		build.WriteString(" ENCHANT ")
		build.WriteString(item.StatEnchant.String())
	}

	if len(item.GemChoice) > 0 {
		build.WriteString(" GEMS ")
		for _, gem := range item.GemChoice {
			build.WriteString(gem.Stats.String())
		}
	}

	build.WriteString(" }")
	return build.String()
}
