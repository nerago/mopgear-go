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

const simTypeCount = 6

type SimData struct {
	Values [simTypeCount]float64
	Detail *[simTypeCount]SimDataDetail
}

type SimDataDetail struct {
	Min    float64
	Max    float64
	StdDev float64
}

func SimData_Make(parts ...any) SimData {
	sim := SimData{}
	for i := 0; i < len(parts); i += 2 {
		simType := parts[i].(SimType)

		switch value := parts[i+1].(type) {
		case int:
			sim.Set(simType, float64(value))
		case float64:
			sim.Set(simType, value)
		}
	}
	return sim
}

func (sim *SimData) Print(printer *util.PrintRecorder) {
	printer.Printf("DPS\t%.2f\n", sim.DPS())
	printer.Printf("TPS\t%.2f\n", sim.TPS())
	printer.Printf("DTPS\t%.2f\n", sim.DTPS())
	printer.Printf("HPS\t%.2f\n", sim.HPS())
	printer.Printf("TMI\t%.2f\n", sim.TMI())
	printer.Printf("DEATH\t%.2f\n", sim.DEATH()*100)
}

func (sim *SimData) CompactStringSignedPercent() string {
	var build util.StringBuild2
	build.WriteString("dps=")
	appendFloatSignedPercent(sim.DPS(), &build)
	build.WriteString("dtps=")
	appendFloatSignedPercent(sim.DTPS(), &build)
	build.WriteString("tmi=")
	appendFloatSignedPercent(sim.TMI(), &build)
	build.WriteString("death=")
	appendFloatSignedPercent(sim.DEATH()*100, &build)
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

func (sim *SimData) CompactStringGeneral() string {
	var build util.StringBuild2
	sim.CompactStringGeneralBuilder(&build)
	return build.String()
}

func (sim *SimData) CompactStringGeneralBuilder(build *util.StringBuild2) {
	build.WriteString("dps=")
	build.WriteFloat64_RightPadded(sim.DPS(), 0, 6)
	build.WriteString(" dtps=")
	build.WriteFloat64_RightPadded(sim.DTPS(), 0, 6)
	build.WriteString(" tmi=")
	build.WriteFloat64_RightPadded(sim.TMI(), 2, 6)
	build.WriteString(" death=")
	build.WriteFloat64(sim.DEATH()*100, 2)
}

func (sim *SimData) IsEmpty() bool {
	return sim.DPS() == 0 && sim.TPS() == 0 && sim.DTPS() == 0 && sim.HPS() == 0 && sim.TMI() == 0 && sim.DEATH() == 0
}

func (sim *SimData) Get(types SimType) float64 {
	return sim.Values[types]
}

func (sim *SimData) GetFriendly(types SimType) float64 {
	if types == Sim_DEATH {
		return sim.Values[Sim_DEATH] * 100
	} else {
		return sim.Values[types]
	}
}

func (sim *SimData) DEATH() float64 { return sim.Values[Sim_DEATH] }
func (sim *SimData) TMI() float64   { return sim.Values[Sim_TMI] }
func (sim *SimData) HPS() float64   { return sim.Values[Sim_HPS] }
func (sim *SimData) DTPS() float64  { return sim.Values[Sim_DTPS] }
func (sim *SimData) TPS() float64   { return sim.Values[Sim_TPS] }
func (sim *SimData) DPS() float64   { return sim.Values[Sim_DPS] }

func (sim *SimData) Set(types SimType, value float64) {
	sim.Values[types] = value
}

func (sim *SimData) SetDetailed(types SimType, average, min, max, stdDev float64) {
	sim.Values[types] = average
	if sim.Detail == nil {
		sim.Detail = new([simTypeCount]SimDataDetail)
	}
	sim.Detail[types] = SimDataDetail{
		Min:    min,
		Max:    max,
		StdDev: stdDev,
	}
}

func (sim *SimData) HasDetailedRanges() bool {
	return sim.Detail != nil
}

func (sim *SimData) GetDetailed(types SimType) (hasAverage, hasDetail bool, average, min, max, stdDev float64) {
	average = sim.Values[types]
	if util.FloatEqualsZero(average) {
		return
	}
	hasAverage = true

	if sim.Detail != nil {
		min = sim.Detail[types].Min
		max = sim.Detail[types].Max
		stdDev = sim.Detail[types].StdDev
		if !util.FloatEqualsZero(min) || !util.FloatEqualsZero(max) {
			hasDetail = true
		}
	}

	return
}

func (sim *SimData) GetDetailed2(types SimType) *SimDataDetail {
	if sim.Detail != nil {
		return &sim.Detail[types]
	} else {
		return nil
	}
}

func (sim *SimData) Seq() iter.Seq2[SimType, float64] {
	return func(yield func(SimType, float64) bool) {
		if !yield(Sim_DPS, sim.Values[Sim_DPS]) {
			return
		}
		if !yield(Sim_TPS, sim.Values[Sim_TPS]) {
			return
		}
		if !yield(Sim_DTPS, sim.Values[Sim_DTPS]) {
			return
		}
		if !yield(Sim_HPS, sim.Values[Sim_HPS]) {
			return
		}
		if !yield(Sim_TMI, sim.Values[Sim_TMI]) {
			return
		}
		if !yield(Sim_DEATH, sim.Values[Sim_DEATH]) {
			return
		}
	}
}

func (sim *SimData) NonZeroTypes() []SimType {
	types := [6]SimType{}
	index := 0
	if sim.Values[Sim_DPS] != 0 {
		types[index] = Sim_DPS
		index++
	}
	if sim.Values[Sim_TPS] != 0 {
		types[index] = Sim_TPS
		index++
	}
	if sim.Values[Sim_DTPS] != 0 {
		types[index] = Sim_DTPS
		index++
	}
	if sim.Values[Sim_HPS] != 0 {
		types[index] = Sim_HPS
		index++
	}
	if sim.Values[Sim_TMI] != 0 {
		types[index] = Sim_TMI
		index++
	}
	if sim.Values[Sim_DEATH] != 0 {
		types[index] = Sim_DEATH
		index++
	}
	if index == 0 {
		panic("empty SimData")
	}
	return types[0:index]
}

func (sim *SimData) QueryIncreaseOfEach(baseSim *SimData) *SimData {
	if sim.IsEmpty() || baseSim.IsEmpty() {
		panic("empty sim shouldn't get called here")
	}

	increase := SimData{}
	for _, resultType := range SimTypeList {
		increase.Set(resultType, increaseForPart(sim, baseSim, resultType))
	}
	return &increase
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

func (sim *SimData) QueryIncreaseOf(baseSim *SimData, part SimType) float64 {
	return increaseForPart(sim, baseSim, part)
}

func (sim *SimData) QueryIncreaseMitigation(baseSim *SimData) float64 {
	checkParts := []SimType{Sim_DPS, Sim_DTPS, Sim_TMI, Sim_DEATH}
	var total float64
	for _, part := range checkParts {
		total += increaseForPart(sim, baseSim, part)
	}
	return total / float64(len(checkParts))
}
