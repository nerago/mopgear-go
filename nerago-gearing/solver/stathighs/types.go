package stathighs

import (
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
)

const (
	c_rangeHigh        = 100.0
	c_baseStatType     = stats.Stat_Strength
	c_finalWeightLimit = 50
	c_offsetLimit      = 0.1
)

type WeightInput struct {
	TotalStat stats.StatBlock
	SimResult simulate.SimData
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

// now about noset - tortos, horridon, ironqon, jikun, durumu
var NewStatWeights_generalMiti = simulate.SimData{
	DPS:   0.2,
	DEATH: 0.1,
	TMI:   0.3,
	DTPS:  0.4,
}

// for withset - malkrok
var NewStatWeights_malkrokWeight = simulate.SimData{
	DPS:   0.15,
	DEATH: 0.3,
	TMI:   0.05,
	DTPS:  0.5,
}

// for comp set - animus
var NewStatWeights_animusWeight = simulate.SimData{
	DPS:   0.4,
	DEATH: 0.1,
	TMI:   0.4,
	DTPS:  0.1,
}

// for dps set
var NewStatWeights_dpsWeight = simulate.SimData{
	DPS:   0.90,
	DEATH: 0.03,
	TMI:   0.03,
	DTPS:  0.04,
}

var G_RequiredStats = []stats.StatType{
	stats.Stat_Strength,
	stats.Stat_Stamina,
	stats.Stat_Crit,
	stats.Stat_Haste,
	stats.Stat_Expertise,
	stats.Stat_Mastery,
	stats.Stat_Dodge,
	stats.Stat_Parry,
}
var G_RequiredSims = []simulate.SimType{
	simulate.Sim_DPS,
	simulate.Sim_DEATH,
	simulate.Sim_TMI,
	simulate.Sim_DTPS,
}
