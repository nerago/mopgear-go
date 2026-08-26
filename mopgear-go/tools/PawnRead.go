package tools

import (
	"errors"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"github.com/nerago/mopgear-go/files"
	"github.com/nerago/mopgear-go/gear_model/ratings"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

func StatRatingsWeights_ReadFile(filename string) ratings.StatRatingsWeightsExtended {
	return StatRatingsWeightsExtended_ReadFile(filename)
}

func pawnWeightToBlock(filename string) stats.StatBlockFloat {
	blockFloat, _, found := PawnWeightReadFile(filename)
	if !found {
		panic("missing weights file " + filename)
	}

	statRatings := stats.StatBlockFloat{}
	blockFloat.MultiplyScalar(weight_types.C_weightMultiplierForRatings, &statRatings)
	return statRatings
}

func StatRatingsWeightsExtended_ReadFile(filename string) ratings.StatRatingsWeightsExtended {
	weight1 := weight_types.Weight1Basic_FromBlock(pawnWeightToBlock(filename))
	weight2, _ := ReadWeight2File(files.ToWeight2(filename))
	weight3, _ := ReadWeight3File(files.ToWeight3(filename))

	if weight1.IsEmpty() && (weight2 == nil || weight2.IsEmpty()) {
		panic("missing weight")
	}

	return ratings.StatRatingsWeightsExtended{
		Weight1: weight1,
		Weight2: weight2,
		Weight3: weight3,
	}
}

func PawnWeightReadFile(filename string) (stats.StatBlockFloat, string, bool) {
	bytes, err := os.ReadFile(filename)
	if errors.Is(err, fs.ErrNotExist) {
		return stats.StatBlockFloat{}, "", false
	} else if err != nil {
		panic(err)
	}
	fullStr := string(bytes)

	blockFloat := stats.StatBlockFloat{}
	for part := range strings.SplitSeq(fullStr, ",") {
		key, value, isValid := strings.Cut(part, "=")
		if isValid {
			switch key {
			case "Intellect":
				addNum(&blockFloat, stats.Stat_Intellect, value)
			case "Strength":
				addNum(&blockFloat, stats.Stat_Strength, value)
			case "Agility":
				addNum(&blockFloat, stats.Stat_Agility, value)
			case "Stamina":
				addNum(&blockFloat, stats.Stat_Stamina, value)
			case "Spirit":
				addNum(&blockFloat, stats.Stat_Spirit, value)
			case "HitRating":
				addNum(&blockFloat, stats.Stat_Hit, value)
			case "CritRating":
				addNum(&blockFloat, stats.Stat_Crit, value)
			case "HasteRating":
				addNum(&blockFloat, stats.Stat_Haste, value)
			case "ExpertiseRating":
				addNum(&blockFloat, stats.Stat_Expertise, value)
			case "MasteryRating":
				addNum(&blockFloat, stats.Stat_Mastery, value)
			case "DodgeRating":
				addNum(&blockFloat, stats.Stat_Dodge, value)
			case "ParryRating":
				addNum(&blockFloat, stats.Stat_Parry, value)
			}
		}
	}
	return blockFloat, fullStr, true
}

func addNum(blockFloat *stats.StatBlockFloat, stat stats.StatType, value string) {
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		panic(err)
	}
	blockFloat[stat] = num
}

func StatRatingsWeights_Testing() ratings.StatRatingsWeightsExtended {
	blockFloat := stats.StatBlockFloat{}
	for i := range blockFloat {
		blockFloat[i] = 1.0
	}
	return ratings.StatRatingsWeightsExtended{
		Weight1: weight_types.Weight1Basic_FromBlock(blockFloat),
	}
}
