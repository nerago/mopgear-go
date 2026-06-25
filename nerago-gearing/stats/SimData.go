package stats

import (
	"iter"
	"math"
	"paladin_gearing_go/util"
)

type SimType uint8

const (
	Sim_DPS   SimType = iota
	Sim_TPS   SimType = iota
	Sim_DTPS  SimType = iota
	Sim_HPS   SimType = iota
	Sim_TMI   SimType = iota
	Sim_DEATH SimType = iota
)

const c_nullIncrease = -100.0

func (types SimType) IsHighGood() bool {
	switch types {
	case Sim_DPS:
		return true
	case Sim_TPS:
		return true
	case Sim_DTPS:
		return false
	case Sim_HPS:
		return true
	case Sim_TMI:
		return false
	case Sim_DEATH:
		return false
	default:
		panic("unknown value")
	}
}
func (types SimType) Name() string {
	switch types {
	case Sim_DPS:
		return "DPS"
	case Sim_TPS:
		return "TPS"
	case Sim_DTPS:
		return "DTPS"
	case Sim_HPS:
		return "HPS"
	case Sim_TMI:
		return "TMI"
	case Sim_DEATH:
		return "DEATH"
	default:
		panic("unknown value")
	}
}

var SimTypeList = []SimType{Sim_DPS, Sim_TPS, Sim_DTPS, Sim_HPS, Sim_TMI, Sim_DEATH}

type SimData struct {
	DPS, TPS, DTPS, HPS, TMI, DEATH float64
}

func (stats SimData) Print(printer *util.PrintRecorder) {
	printer.Printf("DPS\t%.2f\n", stats.DPS)
	printer.Printf("TPS\t%.2f\n", stats.TPS)
	printer.Printf("DTPS\t%.2f\n", stats.DTPS)
	printer.Printf("HPS\t%.2f\n", stats.HPS)
	printer.Printf("TMI\t%.2f\n", stats.TMI)
	printer.Printf("DEATH\t%.2f\n", stats.DEATH*100)
}

func (stats SimData) CompactStringSignedPercent() string {
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

func (stats SimData) CompactStringGeneral() string {
	var build util.StringBuild2
	stats.CompactStringGeneralBuilder(&build)
	return build.String()
}

func (stats SimData) CompactStringGeneralBuilder(build *util.StringBuild2) {
	build.WriteString("dps=")
	build.WriteFloat64_RightPadded(stats.DPS, 0, 6)
	build.WriteString(" dtps=")
	build.WriteFloat64_RightPadded(stats.DTPS, 0, 6)
	build.WriteString(" tmi=")
	build.WriteFloat64_RightPadded(stats.TMI, 2, 6)
	build.WriteString(" death=")
	build.WriteFloat64(stats.DEATH*100, 2)
}

func (stats SimData) IsEmpty() bool {
	return stats.DPS == 0 && stats.TPS == 0 && stats.DTPS == 0 && stats.HPS == 0 && stats.TMI == 0 && stats.DEATH == 0
}

func (stats SimData) Get(types SimType) float64 {
	switch types {
	case Sim_DPS:
		return stats.DPS
	case Sim_TPS:
		return stats.TPS
	case Sim_DTPS:
		return stats.DTPS
	case Sim_HPS:
		return stats.HPS
	case Sim_TMI:
		return stats.TMI
	case Sim_DEATH:
		return stats.DEATH
	default:
		panic("unknown value")
	}
}

func (stats SimData) GetFriendly(types SimType) float64 {
	switch types {
	case Sim_DPS:
		return stats.DPS
	case Sim_TPS:
		return stats.TPS
	case Sim_DTPS:
		return stats.DTPS
	case Sim_HPS:
		return stats.HPS
	case Sim_TMI:
		return stats.TMI
	case Sim_DEATH:
		return stats.DEATH * 100
	default:
		panic("unknown value")
	}
}

func (stats *SimData) Set(types SimType, value float64) {
	switch types {
	case Sim_DPS:
		stats.DPS = value
	case Sim_TPS:
		stats.TPS = value
	case Sim_DTPS:
		stats.DTPS = value
	case Sim_HPS:
		stats.HPS = value
	case Sim_TMI:
		stats.TMI = value
	case Sim_DEATH:
		stats.DEATH = value
	default:
		panic("unknown value")
	}
}

func (stats *SimData) Seq() iter.Seq2[SimType, float64] {
	return func(yield func(SimType, float64) bool) {
		if !yield(Sim_DPS, stats.DPS) {
			return
		}
		if !yield(Sim_TPS, stats.TPS) {
			return
		}
		if !yield(Sim_DTPS, stats.DTPS) {
			return
		}
		if !yield(Sim_HPS, stats.HPS) {
			return
		}
		if !yield(Sim_TMI, stats.TMI) {
			return
		}
		if !yield(Sim_DEATH, stats.DEATH) {
			return
		}
	}
}

func (stats *SimData) NonZeroTypes() []SimType {
	types := [6]SimType{}
	index := 0
	if stats.DPS != 0 {
		types[index] = Sim_DPS
		index++
	}
	if stats.TPS != 0 {
		types[index] = Sim_TPS
		index++
	}
	if stats.DTPS != 0 {
		types[index] = Sim_DTPS
		index++
	}
	if stats.HPS != 0 {
		types[index] = Sim_HPS
		index++
	}
	if stats.TMI != 0 {
		types[index] = Sim_TMI
		index++
	}
	if stats.DEATH != 0 {
		types[index] = Sim_DEATH
		index++
	}
	if index == 0 {
		panic("empty SimData")
	}
	return types[0:index]
}

func (stats *SimData) IncreaseSimBreakdown(baseSim *SimData) SimData {
	if stats.IsEmpty() || baseSim.IsEmpty() {
		panic("empty sim shouldn't get called here")
	}

	increase := SimData{}
	for _, resultType := range SimTypeList {
		increase.Set(resultType, increaseForPart(stats, baseSim, resultType))
	}
	return increase
}

func increaseForPart(sim, baseSim *SimData, part SimType) float64 {
	newValue := sim.Get(part)
	baseValue := baseSim.Get(part)

	var result float64
	if part == Sim_DEATH {
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

func (stats *SimData) IncreaseOf(baseSim *SimData, part SimType) float64 {
	return increaseForPart(stats, baseSim, part)
}

func (stats *SimData) IncreaseMitigation(baseSim *SimData) float64 {
	checkParts := []SimType{Sim_DPS, Sim_DTPS, Sim_TMI, Sim_DEATH}
	var total float64
	for _, part := range checkParts {
		total += increaseForPart(stats, baseSim, part)
	}
	return total / float64(len(checkParts))
}
