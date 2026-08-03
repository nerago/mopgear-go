package weight_types

import (
	"paladin_gearing_go/stats"
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
	var weight1 *Weight1Basic
	if weightCast1, isCast1 := wr.Weight.(*Weight1Basic); isCast1 {
		weight1 = weightCast1
	} else if weightCast2, isCast2 := wr.Weight.(*Weight2Extended); isCast2 {
		weight1 = weightCast2.ConvertToWeight1()
	} else if weightCast3, isCast3 := wr.Weight.(*Weight3ExtendedRanged); isCast3 {
		weight1 = weightCast3.ConvertToWeight2().ConvertToWeight1()
	}
	return weight1
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
