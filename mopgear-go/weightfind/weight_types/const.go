package weight_types

const C_weightMultiplierForRatings = 1000.0

type FixStatsRangeMode uint8

const (
	FixStatsRangeMode_NotSet          FixStatsRangeMode = 0
	FixStatsRangeMode_None            FixStatsRangeMode = iota
	FixStatsRangeMode_ExpertiseAlways FixStatsRangeMode = 1 << iota
	FixStatsRangeMode_HasteAlways     FixStatsRangeMode = 1 << iota
	FixStatsRangeMode_HasteHigherOnly FixStatsRangeMode = 1 << iota
	FixStatsRangeMode_HasteGridOnly   FixStatsRangeMode = 1 << iota
)
