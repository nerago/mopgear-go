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

func (rn StatRange) Overlap(other StatRange) bool {
	if rn.Minimum > other.Maximum {
		return false
	} else if other.Minimum > rn.Maximum {
		return false
	} else {
		return true
	}
}

func (rn StatRange) RangeSize() uint32 {
	return rn.Maximum - rn.Minimum + 1
}

type StatRangeFloat struct {
	Minimum float64
	Maximum float64
}
