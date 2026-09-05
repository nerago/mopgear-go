package weight_types

type FixStatsRangeMode uint8

const (
	FixStatsRangeMode_NotSet          FixStatsRangeMode = 0
	FixStatsRangeMode_None            FixStatsRangeMode = iota
	FixStatsRangeMode_ExpertiseAlways FixStatsRangeMode = 1 << iota
	FixStatsRangeMode_HasteAlways     FixStatsRangeMode = 1 << iota
	FixStatsRangeMode_HasteHigherOnly FixStatsRangeMode = 1 << iota
	FixStatsRangeMode_HasteGridOnly   FixStatsRangeMode = 1 << iota
)
