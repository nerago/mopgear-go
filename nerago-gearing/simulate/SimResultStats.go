package simulate

import (
	"math"
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

const c_nullIncrease = -100.0

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
func (types SimResultType) String() string {
	switch types {
	case Result_DPS:
		return "DPS"
	case Result_TPS:
		return "TPS"
	case Result_DTPS:
		return "DTPS"
	case Result_HPS:
		return "HPS"
	case Result_TMI:
		return "TMI"
	case Result_DEATH:
		return "DEATH"
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
func (stats SimResultStats) PrintNumsOnly(printer *util.PrintRecorder) {
	printer.Printf("%.2f\n", stats.DPS)
	printer.Printf("%.2f\n", stats.TPS)
	printer.Printf("%.2f\n", stats.DTPS)
	printer.Printf("%.2f\n", stats.HPS)
	printer.Printf("%.2f\n", stats.TMI)
	printer.Printf("%.2f\n", stats.DEATH*100)
}

func (stats SimResultStats) CompactStringSignedPercent() string {
	var build util.StringBuild2
	build.WriteString("dps=")
	appendFloatSignedPercent(stats.DPS, &build)
	build.WriteString("dtps=")
	appendFloatSignedPercent(stats.DTPS, &build)
	build.WriteString("tmi=")
	appendFloatSignedPercent(stats.TMI, &build)
	build.WriteString("death=")
	appendFloatSignedPercent(stats.DEATH*100, &build)
	return build.String()
}

func appendFloatSignedPercent(value float64, build *util.StringBuild2) {
	padSize := 6
	if value > 0 {
		build.WriteRune('+')
		padSize--
	}

	build.WriteFloat64_RightPadded(value, 1, padSize)
}

func (stats SimResultStats) CompactStringGeneral() string {
	var build util.StringBuild2
	build.WriteString("dps=")
	build.WriteFloat64_RightPadded(stats.DPS, 0, 6)
	build.WriteString(" dtps=")
	build.WriteFloat64_RightPadded(stats.DTPS, 0, 6)
	build.WriteString(" tmi=")
	build.WriteFloat64_RightPadded(stats.TMI, 0, 6)
	build.WriteString(" death=")
	build.WriteFloat64(stats.DEATH*100, 1)
	return build.String()
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

func (stats SimResultStats) GetFriendly(types SimResultType) float64 {
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
		return stats.DEATH * 100
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

func (stats *SimResultStats) IncreaseSimBreakdown(baseSim *SimResultStats) SimResultStats {
	if stats.IsEmpty() || baseSim.IsEmpty() {
		panic("empty sim shouldn't get called here")
	}

	increase := SimResultStats{}
	for _, resultType := range SimResultTypeList {
		increase.Set(resultType, increaseForPart(stats, baseSim, resultType))
	}
	return increase
}

func increaseForPart(sim, baseSim *SimResultStats, part SimResultType) float64 {
	newValue := sim.Get(part)
	baseValue := baseSim.Get(part)

	var result float64
	if part == Result_DEATH {
		result = baseValue - newValue
	} else if part.IsHighGood() {
		result = (newValue/baseValue - 1.0) * 100
	} else {
		result = (baseValue/newValue - 1.0) * 100
	}

	if math.IsNaN(result) {
		panic("unexpected NaN")
	}
	return result
}

func (stats *SimResultStats) IncreaseOf(baseSim *SimResultStats, part SimResultType) float64 {
	return increaseForPart(stats, baseSim, part)
}

func (stats *SimResultStats) IncreaseMitigation(baseSim *SimResultStats) float64 {
	checkParts := []SimResultType{Result_DPS, Result_DTPS, Result_TMI, Result_DEATH}
	var total float64
	for _, part := range checkParts {
		total += increaseForPart(stats, baseSim, part)
	}
	return total / float64(len(checkParts))
}

func (stats *SimResultStats) BestIncrease(baseSim *SimResultStats) float64 {
	best := c_nullIncrease
	for _, resultType := range SimResultTypeList {
		increase := increaseForPart(stats, baseSim, resultType)
		best = util.MaxIgnoreNaN(best, increase)
	}
	return best
}
