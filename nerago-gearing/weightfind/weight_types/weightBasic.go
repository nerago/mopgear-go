package weight_types

import (
	"encoding/json"
	"math"
	"paladin_gearing_go/gear_model/ratings_old"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
)

type WeightBasic struct {
	content stats.StatBlockFloat
}

func WeightBasic_Make() WeightBasic {
	return WeightBasic{}
}

func WeightBasic_Of(values []float64, statTypes []stats.StatType) WeightBasic {
	wr := WeightBasic{}
	for i, statType := range statTypes {
		wr.Put(statType, values[i])
	}
	return wr
}

func WeightBasic_FromRatingsWeight(ratingWeight *ratings_old.StatRatingsWeightsOld) WeightBasic {
	result := WeightBasic{}
	statBlockFromRatings := ratingWeight.AsFloatBlock()
	statBlockFromRatings.MultiplyScalar(1.0/ratings_old.C_weightMultiplierForRatings, &result.content)
	return result
}

func (wr *WeightBasic) IsEmpty() bool {
	return wr.content.IsEmpty()
}

func (wr *WeightBasic) Get(statType stats.StatType) float64 {
	return wr.content.GetFloat(statType)
}

func (wr *WeightBasic) IsZero(statType stats.StatType) bool {
	return util.FloatEqualsZero(wr.content.GetFloat(statType))
}

func (wr *WeightBasic) Put(statType stats.StatType, value float64) {
	wr.content[statType] = value
}

func (wr *WeightBasic) PlusEquals(statType stats.StatType, value float64) {
	wr.content[statType] += value
}

func (wr *WeightBasic) MinusEquals(statType stats.StatType, value float64) {
	wr.content[statType] -= value
}

func (wr *WeightBasic) MultiplyEquals(statType stats.StatType, value float64) {
	wr.content[statType] *= value
}

func (wr *WeightBasic) DivideEquals(statType stats.StatType, value float64) {
	wr.content[statType] /= value
}

func (wr *WeightBasic) Equals(other *WeightBasic) bool {
	return wr.content.Equals(&other.content)
}

func (wr *WeightBasic) CalcStatScore(input *WeightInput) float64 {
	return wr.content.MultiplyForTotalSum2(&input.TotalStat)
}

func (wr *WeightBasic) CalcStatScore2(stats *stats.StatBlock) float64 {
	return wr.content.MultiplyForTotalSum2(stats)
}

func (wr *WeightBasic) CalcStatScoreScaled(input *WeightInput, statScale map[stats.StatType]float64) float64 {
	total := 0.0
	for statType, scale := range statScale {
		total += input.TotalStat.GetFloat(statType) * wr.content.GetFloat(statType) * scale
	}
	return total
}

func (wr *WeightBasic) ScaleBackToMax(weight float64) WeightBasic {
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

func (wr *WeightBasic) ScaleForBaseStat(statType stats.StatType) WeightBasic {
	factor := 1.0
	value := wr.Get(statType)
	if value != 0 {
		factor = 1.0 / value
	}

	rescaled := wr.Clone()
	wr.content.MultiplyScalar(factor, &rescaled.content)
	return rescaled
}

func (wr *WeightBasic) ScaleForTotalSum(targetTotal float64) WeightBasic {
	existingSum := 0.0
	for value := range wr.content.SeqValues() {
		existingSum += value
	}

	factor := targetTotal / existingSum
	rescaled := wr.Clone()
	wr.content.MultiplyScalar(factor, &rescaled.content)
	return rescaled
}

func (wr *WeightBasic) Clone() WeightBasic {
	return WeightBasic{wr.content.Clone()}
}

func (wr *WeightBasic) String() string {
	return wr.content.CreateString(6)
}

func (wr *WeightBasic) MarshalJSON() ([]byte, error) {
	return json.Marshal(wr.content)
}

func (wr *WeightBasic) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &wr.content)
}
