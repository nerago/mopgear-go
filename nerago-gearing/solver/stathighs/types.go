package stathighs

import (
	"maps"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

type WeightInput struct {
	TotalStat stats.StatBlock
	SimResult stats.SimData
}

type WeightResult map[stats.StatType]float64

// type WeightResult struct{ values map[stats.StatType]float64 }

func WeightResult_Make() WeightResult {
	return make(map[stats.StatType]float64)
}

func (wr *WeightResult) Get(statType stats.StatType) float64 {
	return (*wr)[statType]
}

func (wr *WeightResult) Put(statType stats.StatType, value float64) {
	(*wr)[statType] = value
}

func (wr *WeightResult) Equals(other WeightResult) bool {
	return maps.Equal(*wr, other)
}

func (wr *WeightResult) CalcStatScore(input *WeightInput) float64 {
	total := 0.0
	for statType, weightValue := range *wr {
		total += input.TotalStat.GetFloat(statType) * weightValue
	}
	return total
}

func (wr *WeightResult) CalcStatScoreScaled(input *WeightInput, statScale map[stats.StatType]float64) float64 {
	total := 0.0
	for statType, weightValue := range *wr {
		total += input.TotalStat.GetFloat(statType) * statScale[statType] * weightValue
	}
	return total
}

func (wr *WeightResult) Clone() WeightResult {
	return maps.Clone(*wr)
}

func (wr *WeightResult) String() string {
	build := util.StringBuild2{}
	prepend := ""
	for _, statType := range stats.StatType_List {
		weightValue, haveValue := (*wr)[statType]
		if haveValue {
			build.WriteString(prepend)
			build.WriteFloat64(weightValue, 6)
			prepend = " "
		}
	}
	return build.String()
}

