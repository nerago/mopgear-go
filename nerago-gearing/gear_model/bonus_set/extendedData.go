package bonus_set

import "paladin_gearing_go/stats"

//"Plate of Winged Triumph"
//BONUS 2 DPS 1.003530
//BONUS 2 TPS 1.003529
//BONUS 2 DTPS 0.999100
//BONUS 2 HPS 1.051780
//BONUS 2 TMI 0.995287
//BONUS 2 DEATH 0.926306

//BONUS 4 DPS 1.021051
//BONUS 4 TPS 1.021045
//BONUS 4 DTPS 0.977900
//BONUS 4 HPS 1.083034
//BONUS 4 TMI 0.958943
//BONUS 4 DEATH 0.798846

var g_extendedData = map[string][2]map[stats.SimType]float64{
	"Plate of Winged Triumph": {
		map[stats.SimType]float64{
			stats.Sim_HPS:   1.051780,
			stats.Sim_DEATH: 1.0 / 0.926306,
		},
		map[stats.SimType]float64{
			stats.Sim_DPS:   1.021051,
			stats.Sim_DTPS:  1.0 / 0.977900,
			stats.Sim_HPS:   1.083034,
			stats.Sim_DEATH: 1.0 / 0.798846,
		},
	},
}
