package simulate

import (
	"paladin_gearing_go/util"
)

type SimResultType int8

const (
	Result_DPS   SimResultType = iota
	Result_TPS   SimResultType = iota
	Result_DTPS  SimResultType = iota
	Result_HPS   SimResultType = iota
	Result_TMI   SimResultType = iota
	Result_DEATH SimResultType = iota
)

func (types SimResultType) IsHighGood() bool {
	switch types {
	case Result_DPS:
		return true
	case Result_TPS:
		return true
	case Result_DTPS:
		return false
	case Result_HPS:
		return true
	case Result_TMI:
		return false
	case Result_DEATH:
		return false
	default:
		panic("unknown value")
	}
}

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

func (stats SimResultStats) CompactStringSignedPercent() string {
	var build util.StringBuild2
	build.WriteString("dps=")
	appendFloatSigned(stats.DPS, &build)
	build.WriteString("dtps=")
	appendFloatSigned(stats.DTPS, &build)
	build.WriteString("tmi=")
	appendFloatSigned(stats.TMI, &build)
	build.WriteString("death=")
	appendFloatSigned(stats.DEATH*100, &build)
	return build.String()
}

func appendFloatSigned(value float64, build *util.StringBuild2) {
	padSize := 6
	if value > 0 {
		build.WriteRune('+')
		padSize--
	}

	build.WriteFloat64_RightPadded(value, padSize)
}

func (stats SimResultStats) IsEmpty() bool {
	return stats.DPS == 0 && stats.TPS == 0 && stats.DTPS == 0 && stats.HPS == 0 && stats.TMI == 0 && stats.DEATH == 0
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

func (stats *SimResultStats) Set(types SimResultType, value float64) {
	switch types {
	case Result_DPS:
		stats.DPS = value
	case Result_TPS:
		stats.TPS = value
	case Result_DTPS:
		stats.DTPS = value
	case Result_HPS:
		stats.HPS = value
	case Result_TMI:
		stats.TMI = value
	case Result_DEATH:
		stats.DEATH = value
	default:
		panic("unknown value")
	}
}
