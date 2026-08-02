package ratings_old

import (
	"math"
	. "paladin_gearing_go/stats"
	"paladin_gearing_go/util/util_collection"
	"paladin_gearing_go/weightfind/weight_types"
)

type StatRatingsWeightsOld struct {
	floatWeight StatBlockFloat
}

func (rating *StatRatingsWeightsOld) CreateString() string {
	return rating.floatWeight.CreateString(0)
}

func (rating *StatRatingsWeightsOld) CalcRating(block *StatBlock) float64 {
	return StatBlock_StatBlockFloat_MultiplyForTotalSum2(&rating.floatWeight, block)
}

func (rating *StatRatingsWeightsOld) AsFloatBlock() StatBlockFloat {
	return rating.floatWeight
}

func StatRatingsWeights_Mix(weightA StatRatingsWeightsOld, multiplyA float64, weightB StatRatingsWeightsOld, multiplyB float64, rescaleAround util_collection.Optional[StatType]) *StatRatingsWeightsOld {
	scaleA := StatBlockFloat{}
	weightA.floatWeight.MultiplyScalar(multiplyA, &scaleA)
	scaleB := StatBlockFloat{}
	weightB.floatWeight.MultiplyScalar(multiplyB, &scaleB)

	combined := StatBlockFloat{}
	combined.SetFromAddOthers(&scaleA, &scaleB)

	if statType, hasRescale := rescaleAround.GetWithFlag(); hasRescale {
		div := combined[statType] / weight_types.C_weightMultiplierForRatings
		for i := range combined {
			combined[i] /= div
		}
	}

	intBlock := StatBlock{}
	for i := range combined {
		intBlock[i] = uint32(math.Round(combined[i]))
	}

	return &StatRatingsWeightsOld{combined}
}

func StatRatingsWeights_FromPriorities(priorities []StatType) *StatRatingsWeightsOld {
	// normal range from file load is 1..2500 approx
	// so we could use about 11 bits of multiplier
	// normally have about 8 of them

	var value uint32 = 1024
	blockFloat := StatBlockFloat{}
	for _, stat := range priorities {
		blockFloat[stat] = float64(value)
		value >>= 1
	}
	return &StatRatingsWeightsOld{blockFloat}
}

func StatRatingsWeights_FromBlock(blockFloat StatBlockFloat, includeExpertise bool, includeHit bool, includeSpirit bool) *StatRatingsWeightsOld {
	if !includeExpertise {
		blockFloat[Stat_Expertise] = 0
	}
	if !includeHit {
		blockFloat[Stat_Hit] = 0
	}
	if !includeSpirit {
		blockFloat[Stat_Spirit] = 0
	}

	weight := &StatRatingsWeightsOld{blockFloat}
	return weight
}

func StatRatingsWeights_Testing() *StatRatingsWeightsOld {
	blockFloat := StatBlockFloat{}
	for i := range blockFloat {
		blockFloat[i] = 1.0
	}
	return &StatRatingsWeightsOld{blockFloat}
}
