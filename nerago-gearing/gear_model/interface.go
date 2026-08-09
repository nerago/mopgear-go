package gear_model

import (
	"paladin_gearing_go/gear_model/ratings"
	"paladin_gearing_go/gear_model/requirements"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util/util_collection"
)

type StatRequirements interface {
	CheckSet(block *stats.StatBlock) (bool, string)
	Equals(other any) bool

	IsLow(statType stats.StatType, value uint32) bool
	IsHigh(statType stats.StatType, value uint32) bool
	GetLow(statType stats.StatType) uint32
	GetHigh(statType stats.StatType) uint32

	AsMap() map[stats.StatType]util_collection.HiLoUInt32
}

type StatWeights interface {
	CalcRating(block *stats.StatBlock) float64
	CreateString() string
}

var _ StatWeights = &ratings.StatRatingsWeightsExtended{}
var _ StatRequirements = &requirements.StatRequirementsGeneral{}
var _ StatRequirements = &requirements.StatRequirementsHitExpertise{}
