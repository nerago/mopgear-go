package items

import (
	"paladin_gearing_go/util"
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
	build := util.StringBuild2{}
	item.AppendString(&build)
	return build.String()
}

func (item *FullItem) CreateFullName() string {
	build := util.StringBuild2{}
	item.AppendFullName(&build)
	return build.String()
}

func (item *FullItem) AppendString(build *util.StringBuild2) {
	build.WriteString("{ ")
	build.WriteString(item.Slot.Name())

	build.WriteString(" \"")
	item.AppendFullName(build)

	build.WriteString("\" id=")
	build.WriteUint32(uint32(item.ItemId()))

	build.WriteString(" lvl=")
	build.WriteUint16(item.Ref.ItemLevel)
	build.WriteRune(' ')

	item.StatBase.AppendString(build)

	if !item.StatEnchant.IsEmpty() {
		build.WriteString(" ENCHANT ")
		item.StatEnchant.AppendString(build)
	}

	if len(item.GemChoice) > 0 {
		build.WriteString(" GEMS ")
		for _, gem := range item.GemChoice {
			gem.AppendString(build)
		}
	}

	build.WriteString(" }")
}
