package stats

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

func StatBlock_Add_Into(a, b, out *StatBlock)

// FALLBACK
// func StatBlock_Add_Into(a, b, out *StatBlock) {
// 	for i := range a {
// 		out[i] = a[i] + b[i]
// 	}
// }

func StatBlock_Increment_Mutating(mutate *StatBlock, other *StatBlock)

// FALLBACK
// func StatBlock_Increment_Mutating(mutate *StatBlock, other *StatBlock) {
// 	for i := range block {
// 		mutate[i] += other[i]
// 	}
// }

func (block *StatBlock) MultiplyForTotalSum(other *StatBlock) uint64 {
	var result uint64 = 0
	for i := range block {
		result += uint64(block[i]) * uint64(other[i])
	}
	return result
}

func (block *StatBlock) MultiplyScalar(factor uint32, out *StatBlock) {
	for i := range block {
		out[i] = block[i] * factor
	}
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

func StatBlock_equals(a, b *StatBlock) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TODO toString stuff?

func (block *StatBlock) Get(stat StatType) uint32 {
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
