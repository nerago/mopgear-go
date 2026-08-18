package bonus_set

import (
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/util"
	"iter"
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

func (cr ItemCountsRequired) ItemsMatchRuleFull(items *items.FullEquipMap, mode ItemCountsRequiredMode) bool {
	for set, needCount := range cr.Pairs() {
		haveCount := set.CountItemsFull(items)
		if !itemCountCorrect(mode, haveCount, needCount) {
			return false
		}
	}
	return true
}

func (cr ItemCountsRequired) ItemsMatchRuleSolve(items *items.SolvableEquipMap, mode ItemCountsRequiredMode) bool {
	for set, needCount := range cr.Pairs() {
		haveCount := set.CountItemsSolve(items)
		if !itemCountCorrect(mode, haveCount, needCount) {
			return false
		}
	}
	return true
}

func (cr ItemCountsRequired) itemsMatchRuleSolveWithMessage(items *items.SolvableEquipMap, strBuild *util.StringBuild2, mode ItemCountsRequiredMode) bool {
	for set, needCount := range cr.Pairs() {
		haveCount := set.CountItemsSolve(items)
		if !itemCountCorrect(mode, haveCount, needCount) {
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

func (cr ItemCountsRequired) AppendString(sb *util.StringBuild2) {
	sb.WriteRune('(')
	for set, count := range cr.Pairs() {
		sb.WriteString(set.name)
		sb.WriteRune('=')
		sb.WriteUint8(count)
		sb.WriteRune(' ')
	}
	sb.Rewind(1)
	sb.WriteRune(')')
}

func itemCountCorrect(mode ItemCountsRequiredMode, haveCount uint8, needCount uint8) bool {
	switch mode {
	case CountMode_Exact:
		return haveCount == needCount
	case CountMode_Minimum:
		return haveCount >= needCount
	case CountMode_AllowPlusOne:
		return util.IntBetweenInclusive(needCount, haveCount, needCount+1)
	default:
		panic("unknown type")
	}
}

type ItemCountsRequiredMode uint8

const (
	CountMode_Exact        ItemCountsRequiredMode = iota
	CountMode_Minimum      ItemCountsRequiredMode = iota
	CountMode_AllowPlusOne ItemCountsRequiredMode = iota
)

type ItemCountsRequiredOptions struct {
	Mode    ItemCountsRequiredMode
	Options []ItemCountsRequired
}

func ItemCountsRequiredOptionsAny() ItemCountsRequiredOptions {
	return ItemCountsRequiredOptions{
		Mode: CountMode_Minimum,
	}
}

func ItemCountsRequiredOptionsForFactory(options ...ItemCountsRequired) ItemCountsRequiredOptions {
	return ItemCountsRequiredOptions{
		Mode:    CountMode_AllowPlusOne,
		Options: options,
	}
}

func ItemCountsRequiredOptionsMake(mode ItemCountsRequiredMode, options ...ItemCountsRequired) ItemCountsRequiredOptions {
	return ItemCountsRequiredOptions{
		Mode:    mode,
		Options: options,
	}
}

func (cro ItemCountsRequiredOptions) ItemsMatchAnyRuleSolve(items *items.SolvableEquipMap) (bool, string) {
	if len(cro.Options) == 0 {
		return true, ""
	}

	strBuild := util.StringBuild2{}
	for _, option := range cro.Options {
		if option.itemsMatchRuleSolveWithMessage(items, &strBuild, cro.Mode) {
			return true, ""
		}
	}

	return false, strBuild.String()
}

func (cro ItemCountsRequiredOptions) ItemsMatchAnyRuleFull(items *items.FullEquipMap) bool {
	for _, option := range cro.Options {
		if option.ItemsMatchRuleFull(items, cro.Mode) {
			return true
		}
	}
	return false
}

func (cro ItemCountsRequiredOptions) Equals(other ItemCountsRequiredOptions) bool {
	return cro.Mode == other.Mode &&
		slices.EqualFunc(cro.Options, other.Options, ItemCountsRequired.Equals)
}

func (cro ItemCountsRequiredOptions) IsEmpty() bool {
	return len(cro.Options) == 0
}
