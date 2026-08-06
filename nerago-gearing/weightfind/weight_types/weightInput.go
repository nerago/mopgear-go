package weight_types

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"time"

	"github.com/bartolsthoorn/gohighs/highs"
)

type WeightInput struct {
	TotalStat stats.StatBlock
	SimResult stats.SimData
}

type WeightResult struct {
	Weight    IWeight
	SolveTime time.Duration
	Status    highs.ModelStatus
}

func (wr *WeightResult) AsWeight1() *Weight1Basic {
	switch cast := wr.Weight.(type) {
	case *Weight1Basic:
		return cast
	case *Weight2Extended:
		return cast.ConvertToWeight1()
	case *Weight3ExtendedRanged:
		return cast.ConvertToWeight2().ConvertToWeight1()
	default:
		panic("unknown weight type")
	}
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

func (rn StatRange) Contains(value uint32) bool {
	return rn.Minimum <= value && value <= rn.Maximum
}

type StatRangeFloat struct {
	Minimum float64
	Maximum float64
}

func (rf StatRangeFloat) IsValid() bool {
	return rf.Minimum < rf.Maximum
}

func (rf StatRangeFloat) ContainsOtherRangeFloatAllowance(other StatRangeFloat) bool {
	return util.FloatApproxLessThanOrEqual(rf.Minimum, other.Minimum) &&
		util.FloatApproxLessThanOrEqual(other.Maximum, rf.Maximum)
}
