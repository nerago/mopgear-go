package ratings

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/weightfind/weight_types"
)

type StatRatingsWeightsExtended struct {
	Weight1 weight_types.Weight1Basic
	Weight2 *weight_types.Weight2Extended
	Weight3 *weight_types.Weight3ExtendedRanged
}

func (sw *StatRatingsWeightsExtended) CalcRating(block *stats.StatBlock) float64 {
	return sw.Weight1.CalcStatScore(block)
}

func (sw *StatRatingsWeightsExtended) CreateString() string {
	return sw.Weight1.String()
}

func (sw *StatRatingsWeightsExtended) Equals(other *StatRatingsWeightsExtended) bool {
	return sw.Weight1.Equals(&other.Weight1) &&
		util.NilSafeEqualPointers(sw.Weight2, other.Weight2, (*weight_types.Weight2Extended).Equals) &&
		util.NilSafeEqualPointers(sw.Weight3, other.Weight3, (*weight_types.Weight3ExtendedRanged).Equals)
}

func (sw *StatRatingsWeightsExtended) GetByWeightType(weightType weight_types.WeightType) weight_types.IWeight {
	var weight weight_types.IWeight
	switch weightType {
	case 1:
		weight = &sw.Weight1
	case 2:
		weight = sw.Weight2
	case 3:
		weight = sw.Weight3
	default:
		panic("invalid weight type")
	}

	if weight.IsEmpty() {
		panic("missing weight")
	}

	return weight
}
