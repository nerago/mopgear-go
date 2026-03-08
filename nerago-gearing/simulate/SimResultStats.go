package simulate

import "paladin_gearing_go/util"

type SimResultType int8

const (
	Result_DPS   SimResultType = iota
	Result_TPS   SimResultType = iota
	Result_DTPS  SimResultType = iota
	Result_HPS   SimResultType = iota
	Result_TMI   SimResultType = iota
	Result_DEATH SimResultType = iota
)

var SimResultTypeList = []SimResultType{Result_DPS, Result_TPS, Result_DTPS, Result_HPS, Result_TMI, Result_DEATH}

type SimResultStats struct {
	DPS, TPS, DTPS, HPS, TMI, DEATH float64
}

func (stats SimResultStats) Print(printer *util.PrintRecorder) {
	printer.Printf("DPS\t%.2f\n", stats.DPS)
	printer.Printf("TPS\t%.2f\n", stats.TPS)
	printer.Printf("DTPS\t%.2f\n", stats.DTPS)
	printer.Printf("HPS\t%.2f\n", stats.HPS)
	printer.Printf("TMI\t%.2f\n", stats.TMI)
	printer.Printf("DEATH\t%.2f\n", stats.DEATH*100)
}

func (stats SimResultStats) Get(types SimResultType) float64 {
	switch types {
	case Result_DPS:
		return stats.DPS
	case Result_TPS:
		return stats.TPS
	case Result_DTPS:
		return stats.DTPS
	case Result_HPS:
		return stats.HPS
	case Result_TMI:
		return stats.TMI
	case Result_DEATH:
		return stats.DEATH
	default:
		panic("unknown value")
	}
}
