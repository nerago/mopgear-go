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

// now about noset - tortos, horridon, ironqon, jikun, durumu
var NewStatWeights_generalMiti = simulate.SimData{
	DPS:   0.3,
	DEATH: 0.1,
	TMI:   0.2,
	DTPS:  0.4,
}

// for withset - raden
var NewStatWeights_radenWeight = simulate.SimData{
	DPS:   0.2,
	DEATH: 0.3,
	TMI:   0.1,
	DTPS:  0.4,
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
