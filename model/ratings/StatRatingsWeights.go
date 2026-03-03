package ratings

import (
	"math"
	"os"
	. "paladin_gearing_go/stats"
	"strconv"
	"strings"
)

type StatRatingsWeights struct {
	weight StatBlock
}

func (rating StatRatingsWeights) Weights() string {
	return rating.weight.String()
}

func (rating StatRatingsWeights) CalcRating(block *StatBlock) uint64 {
	return rating.weight.MultiplyForTotalSum(block)
}

func validate(block StatBlock) {
	// number is a bit arbitary, later steps are done in 64 so may not be needed
	for i := range block {
		if block[i] > 0x0FFF_FFFF {
			panic("watch out for overflow")
		}
	}
}

func StatRatingsWeights_Mix(weightA StatRatingsWeights, multiplyA uint32, weightB StatRatingsWeights, multiplyB uint32) StatRatingsWeights {
	scaleA := StatBlock{}
	weightA.weight.MultiplyScalar(multiplyA, &scaleA)
	scaleB := StatBlock{}
	weightB.weight.MultiplyScalar(multiplyB, &scaleB)

	combined := StatBlock{}
	StatBlock_Add_Into(&scaleA, &scaleB, &combined)

	validate(combined)
	return StatRatingsWeights{combined}
}

func StatRatingsWeights_ReadFile(filename string, includeHit, includeExpertise, includeSpirit bool) StatRatingsWeights {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	fullStr := string(bytes)

	block := StatBlock{}
	for part := range strings.SplitSeq(fullStr, ",") {
		key, value, isValid := strings.Cut(part, "=")
		if isValid {
			switch key {
			case "Intellect":
				addNum(&block, Stat_Intellect, value)
			case "Strength":
				addNum(&block, Stat_Strength, value)
			case "Agility":
				addNum(&block, Stat_Agility, value)
			case "Stamina":
				addNum(&block, Stat_Stamina, value)
			case "Spirit":
				addNum(&block, Stat_Spirit, value)
			case "HitRating":
				addNum(&block, Stat_Hit, value)
			case "CritRating":
				addNum(&block, Stat_Crit, value)
			case "HasteRating":
				addNum(&block, Stat_Haste, value)
			case "ExpertiseRating":
				addNum(&block, Stat_Expertise, value)
			case "MasteryRating":
				addNum(&block, Stat_Mastery, value)
			case "DodgeRating":
				addNum(&block, Stat_Dodge, value)
			case "ParryRating":
				addNum(&block, Stat_Parry, value)
			}
		}
	}

	if !includeExpertise {
		block[Stat_Expertise] = 0
	}
	if !includeHit {
		block[Stat_Hit] = 0
	}
	if !includeSpirit {
		block[Stat_Spirit] = 0
	}

	validate(block)
	return StatRatingsWeights{block}
}

func addNum(block *StatBlock, stat StatType, value string) {
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		panic(err)
	}
	block[stat] = uint32(math.Round(num * 1000))
}
