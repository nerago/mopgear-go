package items

import (
	"strconv"
	"strings"
)

func SolvableOptionsMap_of(fullMap *FullOptionsMap) SolvableOptionsMap {
	result := SolvableOptionsMap{}
	for slot := range fullMap {
		fullArray := fullMap[slot]
		if len(fullArray) > 0 {
			solveArray := make([]SolvableItem, 0, len(fullArray))
			for _, item := range fullArray {
				solveItem := SolvableItem_Of(item)
				solveArray = append(solveArray, solveItem)
			}
			result[slot] = solveArray
		}
	}
	return result
}

func findMatch(fullItem []FullItem, solveItem *SolvableItem) *FullItem {
	for _, item := range fullItem {
		if isMatch(&item, solveItem) {
			return &item
		}
	}
	panic("match not found")
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
