package ratings

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/weightfind/weight_types"
)

type StatRatingsWeightsExtended struct {
	Weight1 weight_types.Weight1Basic
	Weight2 weight_types.Weight2Extended
	Weight3 weight_types.Weight3ExtendedRanged
}

func (sw *StatRatingsWeightsExtended) CalcRating(block *stats.StatBlock) float64 {
	return sw.Weight1.CalcStatScore(block)
}

func (sw *StatRatingsWeightsExtended) CreateString() string {
	return sw.Weight1.String()
}

func (sw *StatRatingsWeightsExtended) Equals(other *StatRatingsWeightsExtended) bool {
	return sw.Weight1.Equals(&other.Weight1) &&
		sw.Weight2.Equals(&other.Weight2) &&
		sw.Weight3.Equals(&other.Weight3)
}
