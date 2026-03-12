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

func (item *FullItem) CreateString() string {
	build := strings.Builder{}
	item.AppendString(&build)
	return build.String()
}

func (item *FullItem) CreateFullName() string {
	build := strings.Builder{}
	item.AppendFullName(&build)
	return build.String()
}

func (item *FullItem) AppendString(build *strings.Builder) {
	var buff [10]byte

	build.WriteString("{ ")
	build.WriteString(item.Slot.Name())

	build.WriteString(" \"")
	item.AppendFullName(build)

	build.WriteString("\" id=")
	build.Write(strconv.AppendUint(buff[:0], uint64(item.ItemId()), 10))

	build.WriteString(" lvl=")
	build.Write(strconv.AppendUint(buff[:0], uint64(item.Ref.ItemLevel), 10))
	build.WriteRune(' ')

	item.StatBase.AppendString(build)

	if !item.StatEnchant.IsEmpty() {
		build.WriteString(" ENCHANT ")
		item.StatEnchant.AppendString(build)
	}

	if len(item.GemChoice) > 0 {
		build.WriteString(" GEMS ")
		for _, gem := range item.GemChoice {
			gem.Stats.AppendString(build)
		}
	}

	build.WriteString(" }")
}