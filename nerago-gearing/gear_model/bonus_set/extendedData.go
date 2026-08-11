package bonus_set

import "paladin_gearing_go/stats"

var g_extendedData = map[string][2]map[stats.SimType]float64{
	// regular sacred shield
	"Plate of Winged Triumph": {
		map[stats.SimType]float64{
			stats.Sim_HPS:   1.083982,
			stats.Sim_DEATH: 1.120922,
		},
		map[stats.SimType]float64{
			stats.Sim_DTPS:  1.001963,
			stats.Sim_TMI:   1.001896,
			stats.Sim_DEATH: 1.007347,
		},
	},

	"Plate of Winged Triumph - Eternal Flame Full": {
		map[stats.SimType]float64{
			stats.Sim_HPS:   1.048623,
			stats.Sim_DEATH: 1.069106,
		},
		map[stats.SimType]float64{
			stats.Sim_DPS:   1.025409,
			stats.Sim_DTPS:  1.027021,
			stats.Sim_HPS:   1.096515,
			stats.Sim_TMI:   1.052531,
			stats.Sim_DEATH: 1.382350,
		},
	},

	"Plate of Winged Triumph - Eternal Flame on 4pc": {
		map[stats.SimType]float64{
			stats.Sim_HPS:   1.083982,
			stats.Sim_DEATH: 1.120922,
		},
		map[stats.SimType]float64{
			stats.Sim_DPS:   0.982178,
			stats.Sim_DTPS:  0.565554,
			stats.Sim_HPS:   2.825658,
			stats.Sim_TMI:   0.691548,
			stats.Sim_DEATH: 1.381021,
		},
	},

	"Plate of the Lightning Emperor": {
		map[stats.SimType]float64{
			// default sim doesn't use word of glory since questionable benefit
			// copied over from old flat bonus
			stats.Sim_DEATH: 1.013,
		},
		map[stats.SimType]float64{
			stats.Sim_DPS:   1.024254,
			stats.Sim_DTPS:  1.035569,
			stats.Sim_TMI:   1.049831,
			stats.Sim_DEATH: 1.261225,
		},
	},

	"Plate of the Lightning Emperor - Word of Glory": {
		map[stats.SimType]float64{
			// these are based on Horridon sim with non-standard word of glory and benefit in that scenario
			stats.Sim_DTPS:  1.126533,
			stats.Sim_TMI:   1.118153,
			stats.Sim_DEATH: 1.321158,
		},
		map[stats.SimType]float64{
			// these are from basic T16 sim rotation but still seem a bit high
			stats.Sim_DPS:   1.024254,
			stats.Sim_DTPS:  1.035569,
			stats.Sim_TMI:   1.049831,
			stats.Sim_DEATH: 1.261225,
		},
	},

	"Battlegear of the Lightning Emperor": {
		map[stats.SimType]float64{
			stats.Sim_DPS: 1.005492,
		},
		map[stats.SimType]float64{
			stats.Sim_DPS: 1.049967,
		},
	},

	"Battlegear of Winged Triumph": {
		map[stats.SimType]float64{
			stats.Sim_DPS: 1.016472,
		},
		map[stats.SimType]float64{
			stats.Sim_DPS: 1.044447,
		},
	},
}
