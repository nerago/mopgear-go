package ratings

import (
	"paladin_gearing_go/stats"
	"paladin_gearing_go/weightfind/weight_types"
)

type StatRatingsWeightsGeneric struct {
	weight weight_types.IWeight
}

func (sw StatRatingsWeightsGeneric) CalcRating(block *stats.StatBlock) float64 {
	return sw.weight.CalcStatScore(block)
}

func (sw StatRatingsWeightsGeneric) CreateString() string {
	return sw.weight.String()
}
