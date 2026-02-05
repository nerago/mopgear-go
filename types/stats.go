package types

type StatType int8

const (
	Stat_Strength  = 0
	Stat_Agility   = 1
	Stat_Stamina   = 2
	Stat_Intellect = 3
	Stat_Spirit    = 4
	Stat_Hit       = 5
	Stat_Crit      = 6
	Stat_Haste     = 7
	Stat_Expertise = 8
	Stat_Dodge     = 9
	Stat_Parry     = 10
	Stat_Mastery   = 11
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

func StatBlock_add(a, b *StatBlock) StatBlock {
	result := StatBlock{}
	for i := range a {
		result[i] = a[i] + b[i]
	}
	return result
}

func withChange(block *StatBlock, stat StatType, value uint32) StatBlock {
	var result StatBlock = *block
	result[stat] = value
	return result
}

func withChange2(block *StatBlock, statA StatType, valueA uint32, statB StatType, valueB uint32) StatBlock {
	var result StatBlock = *block
	if statA == statB {
		panic("expected different stats")
	}
	result[statA] = valueA
	result[statB] = valueB
	return result
}

func isEmpty(block *StatBlock) bool {
	for i := range block {
		if block[i] != 0 {
			return false
		}
	}
	return true
}

func hasSingleStat(block *StatBlock) bool {
	countNonZero := 0
	for i := range block {
		if block[i] != 0 {
			countNonZero++
		}
	}
	return countNonZero == 1
}

func equalsStats(a, b *StatBlock) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TODO toString stuff?

func get(block *StatBlock, stat StatType) uint32 {
	return block[stat]
}

func Hit(block *StatBlock) uint32 {
	return block[Stat_Hit]
}

func Expertise(block *StatBlock) uint32 {
	return block[Stat_Expertise]
}

func Spirit(block *StatBlock) uint32 {
	return block[Stat_Spirit]
}
