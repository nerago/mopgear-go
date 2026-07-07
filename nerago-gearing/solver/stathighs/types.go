package stathighs

import (
	"encoding/json"
	"math"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

type WeightInput struct {
	TotalStat stats.StatBlock
	SimResult stats.SimData
}

type WeightResult struct {
	content stats.StatBlockFloat
}

// type WeightResult struct{ values map[stats.StatType]float64 }

func WeightResult_Make() WeightResult {
	return WeightResult{}
}

func (wr *WeightResult) IsEmpty() bool {
	return wr.content.IsEmpty()
}

func (wr *WeightResult) Get(statType stats.StatType) float64 {
	return wr.content.GetFloat(statType)
}

func (wr *WeightResult) IsZero(statType stats.StatType) bool {
	return util.FloatEqualsZero(wr.content.GetFloat(statType))
}

func (wr *WeightResult) Put(statType stats.StatType, value float64) {
	wr.content[statType] = value
}

func (wr *WeightResult) PlusEquals(statType stats.StatType, value float64) {
	wr.content[statType] += value
}

func (wr *WeightResult) MinusEquals(statType stats.StatType, value float64) {
	wr.content[statType] -= value
}

func (wr *WeightResult) MultiplyEquals(statType stats.StatType, value float64) {
	wr.content[statType] *= value
}

func (wr *WeightResult) DivideEquals(statType stats.StatType, value float64) {
	wr.content[statType] /= value
}

func (wr *WeightResult) Equals(other *WeightResult) bool {
	return wr.content.Equals(&other.content)
}

func (wr *WeightResult) CalcStatScore(input *WeightInput) float64 {
	return wr.content.MultiplyForTotalSum2(&input.TotalStat)
}

func (wr *WeightResult) CalcStatScore2(stats *stats.StatBlock) float64 {
	return wr.content.MultiplyForTotalSum2(stats)
}

func (wr *WeightResult) CalcStatScoreScaled(input *WeightInput, statScale map[stats.StatType]float64) float64 {
	total := 0.0
	for statType, scale := range statScale {
		total += input.TotalStat.GetFloat(statType) * wr.content.GetFloat(statType) * scale
	}
	return total
}

func (wr *WeightResult) ScaleBackToMax(weight float64) WeightResult {
	biggest := 0.0
	for value := range wr.content.SeqValues() {
		biggest = max(biggest, math.Abs(value))
	}

	actualLimit := weight * 0.99 // rounding worries
	if biggest < actualLimit {
		return *wr
	}

	factor := actualLimit / biggest
	rescaled := wr.Clone()
	wr.content.MultiplyScalar(factor, &rescaled.content)
	return rescaled
}

func (wr *WeightResult) ScaleForBaseStat(statType stats.StatType) WeightResult {
	factor := 1.0
	value := wr.Get(statType)
	if value != 0 {
		factor = 1.0 / value
	}

	rescaled := wr.Clone()
	wr.content.MultiplyScalar(factor, &rescaled.content)
	return rescaled
}

func (wr *WeightResult) Clone() WeightResult {
	return WeightResult{wr.content.Clone()}
}

func (wr *WeightResult) String() string {
	return wr.content.CreateString(6)
}

func (wr *WeightResult) MarshalJSON() ([]byte, error) {
	return json.Marshal(wr.content)
}

func (wr *WeightResult) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &wr.content)
}
