package stats

import (
	"iter"
	"paladin_gearing_go/util"
)

type StatBlock [12]uint32

var StatBlock_empty = StatBlock{}

func StatBlock_of(stat StatType, value uint32) StatBlock {
	block := StatBlock{}
	block[stat] = value
	return block
}

func StatBlock_of2(statA StatType, valueA uint32, statB StatType, valueB uint32) StatBlock {
	block := StatBlock{}
	if statA == statB {
		panic("expected different stats")
	}
	block[statA] = valueA
	block[statB] = valueB
	return block
}

func (block *StatBlock) MultiplyScalar(factor uint32, out *StatBlock) {
	for i := range block {
		out[i] = block[i] * factor
	}
}

func (block *StatBlock) SetFromAddOthers(a, b *StatBlock) {
	StatBlock_Add_Into(a, b, block)
}

func (block *StatBlock) SetFromAddSubtractOthers(add1, add2, subtract *StatBlock) {
	StatBlock_AddAndSubtract_Into(add1, add2, subtract, block)
}

func (block *StatBlock) IncrementMutating(other *StatBlock) {
	StatBlock_Increment_Mutating(block, other)
}

func (block *StatBlock) MultiplyForTotalSum(other *StatBlock) float64 {
	return StatBlock_MultiplyForTotalSum(block, other)
}

func (block *StatBlock) Equals(other *StatBlock) bool {
	return StatBlock_Equals(block, other)
}

func (block *StatBlock) IsEmpty() bool {
	for i := range block {
		if block[i] != 0 {
			return false
		}
	}
	return true
}

func (block *StatBlock) HasSingleStat() bool {
	countNonZero := 0
	for i := range block {
		if block[i] != 0 {
			countNonZero++
		}
	}
	return countNonZero == 1
}

func (block *StatBlock) GetFloat(stat StatType) float64 {
	return float64(block[stat])
}

func (block *StatBlock) GetUInt(stat StatType) uint32 {
	return block[stat]
}

func (block *StatBlock) Hit() uint32 {
	return block[Stat_Hit]
}

func (block *StatBlock) Expertise() uint32 {
	return block[Stat_Expertise]
}

func (block *StatBlock) Spirit() uint32 {
	return block[Stat_Spirit]
}

func (block *StatBlock) PrimaryStat() PrimaryStatType {
	str := block[Stat_Strength] != 0
	agi := block[Stat_Agility] != 0
	itl := block[Stat_Intellect] != 0

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

func (block *StatBlock) CreateString() string {
	build := util.StringBuild2{}
	block.AppendString(&build)
	return build.String()
}

func (block *StatBlock) CreateStringCSV() string {
	build := util.StringBuild2{}
	for _, value := range block {
		if build.Len() > 0 {
			build.WriteRune(',')
		}
		build.WriteUint32(value)
	}
	return build.String()
}

func (block *StatBlock) AppendString(build *util.StringBuild2) {
	first := true

	build.WriteString("{")

	for i, value := range block {
		if value != 0 {
			var stat StatType = StatType(i)
			name := stat.Name()

			if first {
				first = false
			} else {
				build.WriteRune(' ')
			}

			build.WriteString(name)
			build.WriteRune('=')
			build.WriteUint32(value)
		}
	}

	build.WriteString("}")
}

func (block *StatBlock) SeqValues() iter.Seq[float64] {
	return func(yield func(float64) bool) {
		for i := range block {
			if !yield(float64(block[i])) {
				return
			}
		}
	}
}

func (block *StatBlock) SeqPair() iter.Seq2[StatType, float64] {
	return func(yield func(StatType, float64) bool) {
		for i := range block {
			if !yield(StatType(i), float64(block[i])) {
				return
			}
		}
	}
}

func (block *StatBlock) SeqPairInt() iter.Seq2[StatType, uint32] {
	return func(yield func(StatType, uint32) bool) {
		for i := range block {
			if !yield(StatType(i), block[i]) {
				return
			}
		}
	}
}
