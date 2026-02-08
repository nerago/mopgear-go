package ratings

import (
	"math"
	"os"
	. "paladin_gearing_go/types/stats"
	"strconv"
	"strings"
)

type StatRatingsWeights struct {
	weight StatBlock
}

func (rating StatRatingsWeights) CalcRating(block StatBlock) uint64 {
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

func StatRatingsWeights_mix(weightA StatRatingsWeights, multiplyA uint32, weightB StatRatingsWeights, multiplyB uint32) StatRatingsWeights {
	combined := weightA.weight.MultiplyScalar(multiplyA).Add(weightB.weight.MultiplyScalar(multiplyB))
	validate(combined)
	return StatRatingsWeights{combined}
}

func StatRatingsWeights_readFile(filename string) StatRatingsWeights {
	bytes, err := os.ReadFile()
	if err != nil {
		panic(err)
	}

	fullStr := string(bytes)
	// fullStr := strings.CutSuffix(fullStr, ")")
	// fullStr := strings.CutSuffix(fullStr, " ")

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
