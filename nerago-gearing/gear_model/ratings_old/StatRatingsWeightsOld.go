package ratings_old

import (
	"errors"
	"io/fs"
	"math"
	"os"
	. "paladin_gearing_go/stats"
	"paladin_gearing_go/util/util_collection"
	"strconv"
	"strings"
)

const C_weightMultiplierForRatings = 1000.0

type StatRatingsWeightsOld struct {
	intWeight   StatBlock
	floatWeight StatBlockFloat
}

func (rating *StatRatingsWeightsOld) CreateString() string {
	return rating.intWeight.CreateString()
}

func (rating *StatRatingsWeightsOld) CalcRating(block *StatBlock) float64 {
	return StatBlock_StatBlockFloat_MultiplyForTotalSum2(&rating.floatWeight, block)
}

func (rating *StatRatingsWeightsOld) AsFloatBlock() StatBlockFloat {
	return rating.floatWeight
}

func validate(block StatBlock) {
	// number is a bit arbitrary, later steps are done in 64 so may not be needed
	// that comment now obsolescent now, moving to floats mostly
	for i := range block {
		if block[i] > 0x0FFF_FFFF {
			panic("watch out for overflow")
		}
	}
}

func StatRatingsWeights_Mix(weightA StatRatingsWeightsOld, multiplyA float64, weightB StatRatingsWeightsOld, multiplyB float64, rescaleAround util_collection.Optional[StatType]) *StatRatingsWeightsOld {
	scaleA := StatBlockFloat{}
	weightA.floatWeight.MultiplyScalar(multiplyA, &scaleA)
	scaleB := StatBlockFloat{}
	weightB.floatWeight.MultiplyScalar(multiplyB, &scaleB)

	combined := StatBlockFloat{}
	combined.SetFromAddOthers(&scaleA, &scaleB)

	if statType, hasRescale := rescaleAround.GetWithFlag(); hasRescale {
		div := combined[statType] / C_weightMultiplierForRatings
		for i := range combined {
			combined[i] /= div
		}
	}

	intBlock := StatBlock{}
	for i := range combined {
		intBlock[i] = uint32(math.Round(combined[i]))
	}

	validate(intBlock)
	return &StatRatingsWeightsOld{intBlock, combined}
}

func StatRatingsWeights_FromPriorities(priorities []StatType) *StatRatingsWeightsOld {
	// normal range from file load is 1..2500 approx
	// so we could use about 11 bits of multiplier
	// normally have about 8 of them

	var value uint32 = 1024
	block := StatBlock{}
	blockFloat := StatBlockFloat{}
	for _, stat := range priorities {
		block[stat] = value
		blockFloat[stat] = float64(value)
		value >>= 1
	}
	return &StatRatingsWeightsOld{block, blockFloat}
}

func StatRatingsWeights_ReadFile_IfExists(filename string, includeHit, includeExpertise, includeSpirit bool) (*StatRatingsWeightsOld, string, bool) {
	bytes, err := os.ReadFile(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, "", false
	} else if err != nil {
		panic(err)
	}
	fullStr := string(bytes)

	weight := parseWeightFile(fullStr, includeExpertise, includeHit, includeSpirit)
	return weight, fullStr, true
}

func StatRatingsWeights_ReadFile(filename string, includeHit, includeExpertise, includeSpirit bool) *StatRatingsWeightsOld {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	fullStr := string(bytes)

	return parseWeightFile(fullStr, includeExpertise, includeHit, includeSpirit)
}

func parseWeightFile(fullStr string, includeExpertise bool, includeHit bool, includeSpirit bool) *StatRatingsWeightsOld {
	block := StatBlock{}
	blockFloat := StatBlockFloat{}
	for part := range strings.SplitSeq(fullStr, ",") {
		key, value, isValid := strings.Cut(part, "=")
		if isValid {
			switch key {
			case "Intellect":
				addNum(&block, &blockFloat, Stat_Intellect, value)
			case "Strength":
				addNum(&block, &blockFloat, Stat_Strength, value)
			case "Agility":
				addNum(&block, &blockFloat, Stat_Agility, value)
			case "Stamina":
				addNum(&block, &blockFloat, Stat_Stamina, value)
			case "Spirit":
				addNum(&block, &blockFloat, Stat_Spirit, value)
			case "HitRating":
				addNum(&block, &blockFloat, Stat_Hit, value)
			case "CritRating":
				addNum(&block, &blockFloat, Stat_Crit, value)
			case "HasteRating":
				addNum(&block, &blockFloat, Stat_Haste, value)
			case "ExpertiseRating":
				addNum(&block, &blockFloat, Stat_Expertise, value)
			case "MasteryRating":
				addNum(&block, &blockFloat, Stat_Mastery, value)
			case "DodgeRating":
				addNum(&block, &blockFloat, Stat_Dodge, value)
			case "ParryRating":
				addNum(&block, &blockFloat, Stat_Parry, value)
			}
		}
	}

	if !includeExpertise {
		block[Stat_Expertise] = 0
		blockFloat[Stat_Expertise] = 0
	}
	if !includeHit {
		block[Stat_Hit] = 0
		blockFloat[Stat_Hit] = 0
	}
	if !includeSpirit {
		block[Stat_Spirit] = 0
		blockFloat[Stat_Spirit] = 0
	}

	validate(block)
	return &StatRatingsWeightsOld{block, blockFloat}
}

func addNum(block *StatBlock, blockFloat *StatBlockFloat, stat StatType, value string) {
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		panic(err)
	}
	if num > 0 {
		block[stat] = uint32(math.Round(num * C_weightMultiplierForRatings))
		blockFloat[stat] = num * C_weightMultiplierForRatings
	}
}

func StatRatingsWeights_Testing() *StatRatingsWeightsOld {
	block := StatBlock{}
	blockFloat := StatBlockFloat{}
	for i := range block {
		block[i] = 1
		blockFloat[i] = 1.0
	}
	return &StatRatingsWeightsOld{block, blockFloat}
}
