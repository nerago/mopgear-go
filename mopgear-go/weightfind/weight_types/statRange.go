package weight_types

import (
	"math"

	"github.com/nerago/mopgear-go/util"
)

type StatRange struct {
	Minimum uint32
	Maximum uint32
}

func (rn StatRange) Equals(other StatRange) bool {
	return rn.Maximum == other.Maximum && rn.Minimum == other.Minimum
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
	if rn.IsFullRange() {
		return math.MaxUint32
	} else {
		return rn.Maximum - rn.Minimum + 1
	}
}

func (rn StatRange) Contains(value uint32) bool {
	return rn.Minimum <= value && value <= rn.Maximum
}

func (rn StatRange) IsFullRange() bool {
	return rn.Minimum == 0 && rn.Maximum == math.MaxUint32
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
