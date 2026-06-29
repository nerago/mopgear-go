package stats

import (
	"iter"
	"paladin_gearing_go/util"
)

type StatBlockFloat [12]float64

var StatBlockFloat_empty = StatBlockFloat{}

func StatBlockFloat_of(stat StatType, value float64) StatBlockFloat {
	block := StatBlockFloat{}
	block[stat] = value
	return block
}

func StatBlockFloat_of2(statA StatType, valueA float64, statB StatType, valueB float64) StatBlockFloat {
	block := StatBlockFloat{}
	if statA == statB {
		panic("expected different stats")
	}
	block[statA] = valueA
	block[statB] = valueB
	return block
}

func (block *StatBlockFloat) MultiplyScalar(factor float64, out *StatBlockFloat) {
	for i := range block {
		out[i] = block[i] * factor
	}
}

func (block *StatBlockFloat) SetFromAddOthers(a, b *StatBlockFloat) {
	for i := range a {
		block[i] = a[i] + b[i]
	}
}

func (block *StatBlockFloat) SetFromAddSubtractOthers(add1, add2, subtract *StatBlockFloat) {
	for i := range add1 {
		block[i] = add1[i] + add2[i] - subtract[i]
	}
}

func (block *StatBlockFloat) IncrementMutating(other *StatBlockFloat) {
	for i := range block {
		block[i] += other[i]
	}
}

func (block *StatBlockFloat) MultiplyForTotalSum(other *StatBlockFloat) float64 {
	var result float64 = 0
	for i := range block {
		result += block[i] * other[i]
	}
	return result
}

func (block *StatBlockFloat) MultiplyForTotalSum2(other *StatBlock) float64 {
	var result float64 = 0
	for i := range block {
		result += block[i] * float64(other[i])
	}
	return result
}

func (block *StatBlockFloat) Equals(other *StatBlockFloat) bool {
	for i := range block {
		if !util.FloatsApproxEquals(block[i], other[i]) {
			return false
		}
	}
	return true
}

func (block *StatBlockFloat) IsEmpty() bool {
	for i := range block {
		if !util.FloatEqualsZero(block[i]) {
			return false
		}
	}
	return true
}

func (block *StatBlockFloat) HasSingleStat() bool {
	countNonZero := 0
	for i := range block {
		if !util.FloatEqualsZero(block[i]) {
			countNonZero++
		}
	}
	return countNonZero == 1
}

func (block *StatBlockFloat) GetFloat(stat StatType) float64 {
	return block[stat]
}

func (block *StatBlockFloat) Hit() float64 {
	return block[Stat_Hit]
}

func (block *StatBlockFloat) Expertise() float64 {
	return block[Stat_Expertise]
}

func (block *StatBlockFloat) PrimaryStat() PrimaryStatType {
	str := !util.FloatEqualsZero(block[Stat_Strength])
	agi := !util.FloatEqualsZero(block[Stat_Agility])
	itl := !util.FloatEqualsZero(block[Stat_Intellect])

	primaryCount := 0
	if str {
		primaryCount++
	}
	if agi {
		primaryCount++
	}
	if itl {
		primaryCount++
	}

	if primaryCount > 1 {
		panic("conflicting primary stats")
	} else if primaryCount == 0 {
		return PrimaryStat_None
	} else if str {
		return PrimaryStat_Strength
	} else if agi {
		return PrimaryStat_Agility
	} else {
		return PrimaryStat_Intellect
	}
}

func (block *StatBlockFloat) CreateString(decimalPlaces int) string {
	build := util.StringBuild2{}
	block.AppendString(&build, decimalPlaces)
	return build.String()
}

func (block *StatBlockFloat) CreateStringCSV(decimalPlaces int) string {
	build := util.StringBuild2{}
	for _, value := range block {
		if build.Len() > 0 {
			build.WriteRune(',')
		}
		build.WriteFloat64(value, decimalPlaces)
	}
	return build.String()
}

func (block *StatBlockFloat) AppendString(build *util.StringBuild2, decimalPlaces int) {
	first := true

	build.WriteString("{")

	for i, value := range block {
		if value != 0 {
			stat := StatType(i)
			name := stat.Name()

			if first {
				first = false
			} else {
				build.WriteRune(' ')
			}

			build.WriteString(name)
			build.WriteRune('=')
			build.WriteFloat64(value, decimalPlaces)
		}
	}

	build.WriteString("}")
}

func (block *StatBlockFloat) SeqValues() iter.Seq[float64] {
	return func(yield func(float64) bool) {
		for i := range block {
			if !yield(block[i]) {
				return
			}
		}
	}
}

func (block *StatBlockFloat) SeqPair() iter.Seq2[StatType, float64] {
	return func(yield func(StatType, float64) bool) {
		for i := range block {
			if !yield(StatType(i), block[i]) {
				return
			}
		}
	}
}
