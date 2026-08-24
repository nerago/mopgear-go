package weight_types

import (
	"encoding/json/v2"
	"os"
	"time"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"

	"github.com/bartolsthoorn/gohighs/highs"
)

type WeightInput struct {
	TotalStat stats.StatBlock
	SimResult stats.SimData
}

func WeightInputReadFile(filename string) []WeightInput {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var weightInputs []WeightInput
	err = json.Unmarshal(bytes, &weightInputs)
	if err != nil {
		panic(err)
	}
	return weightInputs
}

func WeightInputWriteFile(weightInputs []WeightInput, filename string) {
	bytes, err := json.Marshal(weightInputs)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(filename, bytes, 0666)
	if err != nil {
		panic(err)
	}
}

type WeightResult struct {
	Weight    IWeight
	SolveTime time.Duration
	Status    highs.ModelStatus
	NewRatio  *SimPriorityBasic
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
		return nil
	}
}

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
