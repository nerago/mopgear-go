package weight_types

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

type AccuracyInfoSimStatRanged struct {
	StatScore     float64
	SimScore      float64
	StatRankRange *util.HiLoInt
	SimRankRange  *util.HiLoInt
	DataSim       *stats.SimData
}

type AccuracyInfoPrePrepare struct {
	SimScore float64
	DataSim  *stats.SimData
	DataStat *stats.StatBlock
}

type AccuracyPreparedEntry struct {
	StatScore    float64
	Stats        *stats.StatBlock
	SimRankRange *util.HiLoInt
}
