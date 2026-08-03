package weight_types

import (
	"encoding/json"
	"math"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_collection"
)

type IWeight interface {
	IsEmpty() bool
	CalcStatScore(stats *stats.StatBlock) float64
	String() string
}

type Weight1Basic struct {
	content stats.StatBlockFloat
}

func Weight1Basic_Make() Weight1Basic {
	return Weight1Basic{}
}

func Weight1Basic_Of(values []float64, statTypes []stats.StatType) Weight1Basic {
	wr := Weight1Basic{}
	for i, statType := range statTypes {
		wr.Put(statType, values[i])
	}
	return wr
}

func Weight1Basic_FromBlock(ratingWeight stats.StatBlockFloat) Weight1Basic {
	result := Weight1Basic{content: ratingWeight}
	return result
}

func (wr *Weight1Basic) IsEmpty() bool {
	return wr.content.IsEmpty()
}

func (wr *Weight1Basic) Get(statType stats.StatType) float64 {
	return wr.content.GetFloat(statType)
}

func (wr *Weight1Basic) IsZero(statType stats.StatType) bool {
	return util.FloatEqualsZero(wr.content.GetFloat(statType))
}

func (wr *Weight1Basic) Put(statType stats.StatType, value float64) {
	wr.content[statType] = value
}

func (wr *Weight1Basic) PlusEquals(statType stats.StatType, value float64) {
	wr.content[statType] += value
}

func (wr *Weight1Basic) MinusEquals(statType stats.StatType, value float64) {
	wr.content[statType] -= value
}

func (wr *Weight1Basic) MultiplyEquals(statType stats.StatType, value float64) {
	wr.content[statType] *= value
}

func (wr *Weight1Basic) DivideEquals(statType stats.StatType, value float64) {
	wr.content[statType] /= value
}

func (wr *Weight1Basic) Equals(other *Weight1Basic) bool {
	return wr.content.Equals(&other.content)
}

func (wr *Weight1Basic) CalcStatScore(stats *stats.StatBlock) float64 {
	return wr.content.MultiplyForTotalSum2(stats)
}

func (wr *Weight1Basic) CalcStatScoreScaled(input *WeightInput, statScale util_collection.EnumMap[stats.StatType, float64]) float64 {
	total := 0.0
	for statType, scale := range statScale.SeqKeyValue() {
		total += input.TotalStat.GetFloat(statType) * wr.content.GetFloat(statType) * scale
	}
	return total
}

func (wr *Weight1Basic) ScaleBackToMax(weight float64) Weight1Basic {
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

func (wr *Weight1Basic) ScaleForBaseStat(statType stats.StatType) Weight1Basic {
	factor := 1.0
	value := wr.Get(statType)
	if value != 0 {
		factor = 1.0 / value
	}

	rescaled := wr.Clone()
	wr.content.MultiplyScalar(factor, &rescaled.content)
	return rescaled
}

func (wr *Weight1Basic) ScaleForTotalSum(targetTotal float64) Weight1Basic {
	existingSum := 0.0
	for value := range wr.content.SeqValues() {
		existingSum += value
	}

	factor := targetTotal / existingSum
	rescaled := wr.Clone()
	wr.content.MultiplyScalar(factor, &rescaled.content)
	return rescaled
}

func (wr *Weight1Basic) Clone() Weight1Basic {
	return Weight1Basic{wr.content.Clone()}
}

func (wr *Weight1Basic) String() string {
	return wr.content.CreateString(6)
}

func (wr *Weight1Basic) MarshalJSON() ([]byte, error) {
	return json.Marshal(wr.content)
}

func (wr *Weight1Basic) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &wr.content)
}
