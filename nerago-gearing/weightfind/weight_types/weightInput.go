package weight_types

import "paladin_gearing_go/stats"

type WeightInput struct {
	TotalStat stats.StatBlock
	SimResult stats.SimData
}

type StatRange struct {
	Minimum uint32
	Maximum uint32
}

func (rn StatRange) RangeSize() uint32 {
	return rn.Maximum - rn.Minimum + 1
}
