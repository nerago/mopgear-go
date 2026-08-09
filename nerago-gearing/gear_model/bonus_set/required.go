package bonus_set

import (
	"iter"
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
	"slices"
)

type ItemCountsRequired struct {
	sets   []*BonusLookup
	counts []uint8
}

func ItemCountsRequiredMake(init ...any) ItemCountsRequired {
	cr := ItemCountsRequired{}
	for i := 0; i < len(init); i += 2 {
		setName := init[i].(string)
		count := init[i+1].(int)
		set := g_setNameLookup[setName]
		if set == nil {
			panic("set not found")
		}
		cr.sets = append(cr.sets, set)
		cr.counts = append(cr.counts, uint8(count))
	}
	return cr
}

func (cr ItemCountsRequired) Equals(other ItemCountsRequired) bool {
	return slices.EqualFunc(cr.sets, other.sets, (*BonusLookup).Equals) &&
		slices.Equal(cr.counts, other.counts)
}

func (cr ItemCountsRequired) Pairs() iter.Seq2[*BonusLookup, uint8] {
	return func(yield func(*BonusLookup, uint8) bool) {
		for i := range cr.sets {
			if !yield(cr.sets[i], cr.counts[i]) {
				return
			}
		}
	}
}

func (cr ItemCountsRequired) PairsByIndex(index int) (*BonusLookup, uint8) {
	return cr.sets[index], cr.counts[index]
}

func (cr ItemCountsRequired) Count() int {
	return len(cr.sets)
}

func (cr ItemCountsRequired) ItemsMatchRuleFull(items *items.FullEquipMap) bool {
	for set, needCount := range cr.Pairs() {
		haveCount := set.CountItemsFull(items)
		if haveCount < needCount || haveCount > needCount+1 {
			return false
		}
	}
	return true
}

func (cr ItemCountsRequired) ItemsMatchRuleSolve(items *items.SolvableEquipMap) bool {
	for set, needCount := range cr.Pairs() {
		haveCount := set.CountItemsSolve(items)
		if haveCount < needCount || haveCount > needCount+1 {
			return false
		}
	}
	return true
}

func (cr ItemCountsRequired) itemsMatchRuleSolveWithMessage(items *items.SolvableEquipMap, strBuild *util.StringBuild2) bool {
	for set, needCount := range cr.Pairs() {
		haveCount := set.CountItemsSolve(items)
		if haveCount < needCount || haveCount > needCount+1 {
			strBuild.WriteString("For '")
			strBuild.WriteString(set.Name())
			strBuild.WriteString(" have=")
			strBuild.WriteUint16(uint16(haveCount))
			strBuild.WriteString(" expect=")
			strBuild.WriteUint16(uint16(needCount))
			strBuild.WriteRune('\n')
			return false
		}
	}
	return true
}

type ItemCountsRequiredOptions []ItemCountsRequired

func (cro ItemCountsRequiredOptions) ItemsMatchAnyRuleSolve(items *items.SolvableEquipMap) (bool, string) {
	strBuild := util.StringBuild2{}
	for _, option := range cro {
		if option.itemsMatchRuleSolveWithMessage(items, &strBuild) {
			return true, ""
		}
	}
	return false, strBuild.String()
}

func (cro ItemCountsRequiredOptions) ItemsMatchAnyRuleFull(items *items.FullEquipMap) bool {
	for _, option := range cro {
		if option.ItemsMatchRuleFull(items) {
			return true
		}
	}
	return false
}
