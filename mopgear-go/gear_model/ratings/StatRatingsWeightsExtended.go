package ratings

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type StatRatingsWeightsExtended struct {
	Weight1Scaled     weight_types.Weight1_ScaledSolvable
	Weight1Compatible weight_types.Weight1_CompatibleExternal
	Weight2           *weight_types.Weight2
	Weight3           *weight_types.Weight3
}

func (sw *StatRatingsWeightsExtended) CalcRating(block *stats.StatBlock) float64 {
	return sw.Weight1Scaled.CalcStatScore(block)
}

func (sw *StatRatingsWeightsExtended) CreateString() string {
	return sw.Weight1Compatible.String()
}

func (sw *StatRatingsWeightsExtended) Equals(other *StatRatingsWeightsExtended) bool {
	return sw.Weight1Scaled.Equals(&other.Weight1Scaled) &&
		util.NilSafeEqualPointers(sw.Weight2, other.Weight2, (*weight_types.Weight2).Equals) &&
		util.NilSafeEqualPointers(sw.Weight3, other.Weight3, (*weight_types.Weight3).Equals)
}

func (sw *StatRatingsWeightsExtended) GetByWeightTypeForSolve(weightType weight_types.WeightType) weight_types.IWeight {
	var weight weight_types.IWeight
	switch weightType {
	case 1:
		weight = &sw.Weight1Scaled
	case 2:
		weight = sw.Weight2
	case 3:
		weight = sw.Weight3
	default:
		panic("invalid weight type")
	}

	if weight == nil || weight.IsEmpty() {
		panic("missing weight")
	}

	return weight
}
