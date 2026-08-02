package tools

import (
	"errors"
	"io/fs"
	"os"
	"paladin_gearing_go/gear_model/ratings_old"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/weightfind/weight_types"
	"strconv"
	"strings"
)

func StatRatingsWeights_ReadFile(filename string, includeHit, includeExpertise, includeSpirit bool) *ratings_old.StatRatingsWeightsOld {
	blockFloat, _, found := PawnWeightReadFile(filename)
	if !found {
		panic("missing weights file " + filename)
	}

	statRatings := stats.StatBlockFloat{}
	blockFloat.MultiplyScalar(weight_types.C_weightMultiplierForRatings, &statRatings)

	return ratings_old.StatRatingsWeights_FromBlock(statRatings, includeExpertise, includeHit, includeSpirit)
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
