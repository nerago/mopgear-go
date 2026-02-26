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

func StatBlock_Add_NoPointer(a, b StatBlock) StatBlock {
	result := StatBlock{}
	for i := range a {
		result[i] = a[i] + b[i]
	}
	return result
}

func StatBlock_Add(a, b *StatBlock) StatBlock {
	result := StatBlock{}
	for i := range a {
		result[i] = a[i] + b[i]
	}
	return result
}

func (block *StatBlock) Add(other *StatBlock) StatBlock {
	return StatBlock_Add(block, other)
}

func (block *StatBlock) Increment_Mutating(other *StatBlock) {
	for i := range block {
		block[i] += other[i]
	}
}

func (block *StatBlock) MultiplyForTotalSum(other *StatBlock) uint64 {
	var result uint64 = 0
	for i := range block {
		result += uint64(block[i]) * uint64(other[i])
	}
	return result
}

func (block *StatBlock) MultiplyScalar(factor uint32) StatBlock {
	result := StatBlock{}
	for i := range block {
		result[i] = block[i] * factor
	}
	return result
}

func (block *StatBlock) WithChange(stat StatType, value uint32) StatBlock {
	var result StatBlock = *block
	result[stat] = value
	return result
}

func (block *StatBlock) WithChange2(statA StatType, valueA uint32, statB StatType, valueB uint32) StatBlock {
	var result StatBlock = *block
	if statA == statB {
		panic("expected different stats")
	}
	result[statA] = valueA
	result[statB] = valueB
	return result
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
