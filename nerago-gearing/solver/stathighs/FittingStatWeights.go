package stathighs

import (
	"cmp"
	"math"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"slices"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (

	// STR example: 30554
	c_statRangeHigh = 50000

	c_scaleBigSim = 1000
	// TPS example:  1957667
	c_simRangeHigh = 5000000 / c_scaleBigSim // we try to scale values like above towards nicer range

	c_outputFittingDifference = 1
	c_outputIncludePerInclude = -1
)

// so we want to define a line of best fit for each stat/sim
// but also only for certain ranges of each stat, others excluded

// question is do we work each stat separately, even though ranges may not line up
// yes, we can always compose the individual parts later

// but do we allow ranges on the other stats we're not checking
// we might end up very fragmented and noisy if we do

// do we consider simType separately? why not - some aren't correlated, going to hard enough to reconcile across one dimension
//                                    why - some are highly correlated

// we need to consider that we're developing a function that correlates to the sim output, not predicts it in any summation sense
// maybe suggests start with the individual ones, less tempting to try to hit totals

type FittingEachStatWeightProcess struct {
	printer *util.PrintRecorder
	input   *utilhighs.InputBuilder

	each util.MapMap[stats.StatType, simulate.SimResultType, FittingSingleStatWeightProcess]
}

func (fiteach *FittingEachStatWeightProcess) Init(printer *util.PrintRecorder) {
	fiteach.printer = printer
	fiteach.input = new(utilhighs.InputBuilder)
	fiteach.input.Minimise = true
}

////////////////////////////////////////////////////////

type StatRange struct {
	Minimum float64
	Maximum float64
}

type FittingSingleStatSegmentsProcess struct {
	printer *util.PrintRecorder

	inputData []WeightInput
	stat      stats.StatType
	sim       simulate.SimResultType

	segments         map[StatRange]FittingSingleStatResult
	coveredRangeLow  uint32
	coveredRangeHigh uint32
}

func (fitseg *FittingSingleStatSegmentsProcess) Init(printer *util.PrintRecorder, stat stats.StatType, sim simulate.SimResultType) {
	fitseg.printer = printer
	fitseg.segments = make(map[StatRange]FittingSingleStatResult)
	fitseg.stat = stat
	fitseg.sim = sim
}

func (fitseg *FittingSingleStatSegmentsProcess) SupplyDataFromStandard(inputData []WeightInput) {
	fitseg.inputData = inputData
}

func (fitseg *FittingSingleStatSegmentsProcess) Run() map[StatRange]FittingSingleStatResult {
	fitseg.runInitial()

	lowData := util.FilterSliceAsNew(fitseg.inputData, func(in *WeightInput) bool { return in.TotalStat.Get(fitseg.stat) < fitseg.coveredRangeLow })
	fitseg.runEndSegment(lowData)

	highData := util.FilterSliceAsNew(fitseg.inputData, func(in *WeightInput) bool { return in.TotalStat.Get(fitseg.stat) > fitseg.coveredRangeHigh })
	fitseg.runEndSegment(highData)

	return fitseg.segments
}

func (fitseg *FittingSingleStatSegmentsProcess) runInitial() {
	initialFit := FittingSingleStatWeightProcess{}
	initialFit.Init(fitseg.printer)
	initialFit.SetMinimumIncludeRate(0.3)
	initialFit.SupplyDataFromStandard(fitseg.inputData, fitseg.stat, fitseg.sim)
	weight := initialFit.Run()

	statRange := StatRange{weight.Minimum, weight.Maximum}
	fitseg.segments[statRange] = weight

	fitseg.coveredRangeLow = uint32(math.Round(weight.Minimum))
	fitseg.coveredRangeHigh = uint32(math.Round(weight.Maximum))
}

func (fitseg *FittingSingleStatSegmentsProcess) runEndSegment(inputData []WeightInput) {
	initialFit := FittingSingleStatWeightProcess{}
	initialFit.Init(fitseg.printer)
	initialFit.SetMinimumIncludeRate(1)
	initialFit.SupplyDataFromStandard(inputData, fitseg.stat, fitseg.sim)
	weight := initialFit.Run()

	statRange := StatRange{weight.Minimum, weight.Maximum}
	fitseg.segments[statRange] = weight

	fitseg.coveredRangeLow = min(fitseg.coveredRangeLow, uint32(math.Round(weight.Minimum)))
	fitseg.coveredRangeHigh = max(fitseg.coveredRangeHigh, uint32(math.Round(weight.Maximum)))
}

////////////////////////////////////////////////////////

type FittingSingleStatWeightProcess struct {
	printer *util.PrintRecorder
	input   *utilhighs.InputBuilder

	minimumIncludeRate float64
	inputData          []fittingSample
	inputDataSimScale  float64

	linearLineDiff int
	linearInclude  int

	lineSlope        utilhighs.ColumnIndex
	lineOffset       utilhighs.ColumnIndex
	minimumThreshold utilhighs.ColumnIndex
	maximumThreshold utilhighs.ColumnIndex
	includeColumns   []utilhighs.ColumnIndex

	includeCountRow utilhighs.ConstraintRowBuild
}

type FittingSingleStatResult struct {
	LineSlope      float64
	LineOffset     float64
	Minimum        float64
	Maximum        float64
	IncludePercent float64
}

type fittingSample struct {
	statValue     float64
	simResult     float64
	includeColumn utilhighs.ColumnIndex
}

func (fit *FittingSingleStatWeightProcess) Init(printer *util.PrintRecorder) {
	fit.printer = printer
	fit.input = new(utilhighs.InputBuilder)
	fit.input.Minimise = true
}

func (fit *FittingSingleStatWeightProcess) SetMinimumIncludeRate(percent float64) {
	fit.minimumIncludeRate = percent
}

func (fit *FittingSingleStatWeightProcess) SupplyDataFromStandard(inputData []WeightInput, stat stats.StatType, sim simulate.SimResultType) {
	fit.inputData = util.CastSliceAsNew(inputData, func(input *WeightInput) fittingSample {
		return fittingSample{
			float64(input.TotalStat.Get(stat)),
			scaleSimItem(input.SimResult.Get(sim), sim),
			-1,
		}
	})
	fit.inputDataSimScale = scaleSimItem(1, sim)
}

func scaleSimItem(value float64, sim simulate.SimResultType) float64 {
	// example values 1671858.348 10396269.605 117613.197 217148.877 180.467 21.1
	// with scaleBig  1671.858    10396.269    117.613197 217.148877
	switch sim {
	case simulate.Result_DPS, simulate.Result_TPS, simulate.Result_DTPS, simulate.Result_HPS:
		return value / c_scaleBigSim
	case simulate.Result_TMI:
		return value
	case simulate.Result_DEATH:
		return value * 100
	default:
		panic("unknown type")
	}
}

func (fit *FittingSingleStatWeightProcess) Run() FittingSingleStatResult {
	fit.lineSlope = fit.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "slope"})
	fit.lineOffset = fit.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "offset"})
	fit.minimumThreshold = fit.input.CreateColumnGeneral(highs.Continuous, 0, c_statRangeHigh, utilhighs.DebugString{Text: "minimum"})
	fit.maximumThreshold = fit.input.CreateColumnGeneral(highs.Continuous, 0, c_statRangeHigh, utilhighs.DebugString{Text: "maximum"})

	fit.input.BlendMultiObjectives = true
	fit.linearInclude = fit.input.AddLinearObjective(20, 0, 1000, 1, 1)
	fit.linearLineDiff = fit.input.AddLinearObjective(1, 0, 10000, 5, 2)

	maxVsMin := utilhighs.ConstraintRowBuild{}
	maxVsMin.Add(fit.minimumThreshold, -1)
	maxVsMin.Add(fit.maximumThreshold, 1)
	maxVsMin.Finish(fit.input, 0, utilhighs.C_PlusInf)

	for sample := range util.ForPointer(fit.inputData) {
		fit.addSample(sample)
	}

	fit.includeCountRow.Finish(fit.input, float64(len(fit.inputData))*fit.minimumIncludeRate, utilhighs.C_PlusInf)

	solution, log := fit.input.RunHighs()
	fit.printer.AppendOther(log)
	fit.printer.Println(solution.Status.String())

	fit.input.DebugPrintColumns(solution, fit.printer)

	return fit.buildResult(solution)
}

func (fit *FittingSingleStatWeightProcess) buildResult(solution *highs.Solution) FittingSingleStatResult {
	result := FittingSingleStatResult{}
	result.LineSlope = solution.ColValues[fit.lineSlope] / fit.inputDataSimScale
	result.LineOffset = solution.ColValues[fit.lineOffset] / fit.inputDataSimScale
	result.Minimum = solution.ColValues[fit.minimumThreshold]
	result.Maximum = solution.ColValues[fit.maximumThreshold]

	includeCount := 0
	for _, col := range fit.includeColumns {
		if utilhighs.FloatEqualsOne(solution.ColValues[col]) {
			includeCount++
		}
	}
	result.IncludePercent = float64(includeCount) / float64(len(fit.inputData))

	inputSorted := slices.Clone(fit.inputData)
	slices.SortFunc(inputSorted, func(a, b fittingSample) int { return cmp.Compare(a.statValue, b.statValue) })
	for _, sample := range inputSorted {
		fit.printer.Printf("INC %f %f\n", sample.statValue, solution.ColValues[sample.includeColumn])
	}

	return result
}

func (fit *FittingSingleStatWeightProcess) addSample(sample *fittingSample) {
	includeColumn := fit.sampleIncludeToggleColumn(sample)
	fit.sampleToFitLine(sample, includeColumn)
}

func (fit *FittingSingleStatWeightProcess) sampleIncludeToggleColumn(sample *fittingSample) utilhighs.ColumnIndex {
	includeColumn := fit.input.CreateColumnForLinearObjective(highs.Integer, 0, 1, c_outputIncludePerInclude, fit.linearInclude, utilhighs.DebugString{Text: "include"})
	fit.includeCountRow.Add(includeColumn, 1)
	fit.includeColumns = append(fit.includeColumns, includeColumn)
	sample.includeColumn = includeColumn

	isOverMinimum := fit.makeIsOverMinimum(sample.statValue)
	isUnderMaximum := fit.makeIsUnderMaximum(sample.statValue)

	and := utilhighs.ContraintAndBuilder{}
	and.AddInput(isOverMinimum)
	and.AddInput(isUnderMaximum)
	and.SetOutput(includeColumn)
	and.FinishAndApply(fit.input)

	return includeColumn
}

// actually equal or greater than minimum
func (fit *FittingSingleStatWeightProcess) makeIsOverMinimum(statValue float64) utilhighs.ColumnIndex {
	isOverMinimum := fit.input.CreateColumnBool(utilhighs.DebugString{Text: "isOverMinimum"})

	// ORIGINAL
	// if overmin:      1*range + min <= range + stat  ->>  min <= stat
	// if !overmin:     0*range + min <= range + stat  ->>  min is free
	// if stat > min:   x*range + min <= range + stat  ->>  x*range <= range + stat - min  ->>  x*range <= range + small_positive   ->>   x = 0 or 1
	// if stat < min:   x*range + min <= range + stat  ->>  x*range <= range + stat - min  ->>  x*range <= range + small_negative   ->>   x = 0
	// if stat == min:  x*range + min <= range + stat  ->>  x*range <= range   ->>   x = 0 or 1
	checkIsOverMin := utilhighs.ConstraintRowBuild{Debug: "checkIsOverMin"}
	checkIsOverMin.Add(isOverMinimum, c_statRangeHigh)
	checkIsOverMin.Add(fit.minimumThreshold, 1)
	checkIsOverMin.Finish(fit.input, utilhighs.C_MinusInf, c_statRangeHigh+statValue)

	//   min - stat + x.range >= 0   ->>   min + x.range >= stat
	// if stat > min  ->>  min - stat + x.range >= 0  ->>  small_negative + x.range >= 0  ->>  x=1
	// if stat < min  ->>  min - stat + x.range >= 0  ->>  small_positive + x.range >= 0  ->>  x=0 or 1
	// if stat == min  ->> min - stat + x.range >= 1 ->>  x.range >= 1  ->> x=1
	// if overmin     ->>  min - stat + 1.range >= 0  ->>  min >= stat - range   ->>   min is free
	// if !overmin    ->>  min - stat + 0.range >= 0  ->>  min >= stat    ->>   min is free
	// modifiying with the plus one for the equals case, the rest of the math should hold ok
	setIfOverMin := utilhighs.ConstraintRowBuild{Debug: "setIfOverMin"}
	setIfOverMin.Add(fit.minimumThreshold, 1)
	setIfOverMin.Add(isOverMinimum, c_statRangeHigh)
	setIfOverMin.Finish(fit.input, statValue+1, utilhighs.C_PlusInf)

	return isOverMinimum
}

// actually equal or less than maximum
func (fit *FittingSingleStatWeightProcess) makeIsUnderMaximum(statValue float64) utilhighs.ColumnIndex {
	isUnderMaximum := fit.input.CreateColumnBool(utilhighs.DebugString{Text: "isUnderMaximum"})

	// if undermax:    1*stat - max <= 0     ->>      stat <= max
	// if !undermax:   0*stat - max <= 0     ->>      max >= 0, (max free)
	// if stat <= max   ->>   x.stat - max <= 0    ->>     x.stat <= max   (x is free)
	// if stat > max    ->>   x.stat - max <= 0    ->>     x=0
	checkIsUnderMax := utilhighs.ConstraintRowBuild{Debug: "checkIsUnderMax"}
	checkIsUnderMax.Add(isUnderMaximum, statValue)
	checkIsUnderMax.Add(fit.maximumThreshold, -1)
	checkIsUnderMax.Finish(fit.input, utilhighs.C_MinusInf, 0)

	//    max - stat - x.range <= 0    ->>     max - x.range <= stat
	// if stat < max    ->>   max - stat - x.range <= 0   ->>   small_positive - x.range <= 0   ->>   small_positive <= x.range   ->>   x=1
	// if stat > max    ->>   max - stat - x.range <= 0   ->>   small_negative - x.range <= 0   ->>   small_negative <= x.range   ->>   x=0 or 1
	// if stat == max   ->>   max - stat - x.range <= 0   ->>   0 - x.range <= 0   ->>   0 <= x.range   ->> x=0 or 1
	// if undermax      ->>   max - stat - 1.range <= 0   ->>   max <= 1.range + stat  ->>  max is free
	// if !undermax     ->>   max - stat - 0.range <= 0   ->>   max - stat <= 0   ->>   max <= stat
	setIfUnderMax := utilhighs.ConstraintRowBuild{Debug: "setIfUnderMax"}
	setIfUnderMax.Add(fit.maximumThreshold, 1)
	setIfUnderMax.Add(isUnderMaximum, -c_statRangeHigh)
	setIfUnderMax.Finish(fit.input, utilhighs.C_MinusInf, statValue)

	return isUnderMaximum
}

func (fit *FittingSingleStatWeightProcess) sampleToFitLine(sample *fittingSample, toggle utilhighs.ColumnIndex) {
	difference := fit.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "difference"})
	differenceAbs := fit.input.CreateColumnForLinearObjective(highs.Continuous, 0, utilhighs.C_PlusInf, c_outputFittingDifference, fit.linearLineDiff, utilhighs.DebugString{Text: "differenceAbs"})

	// i'd like lineSlope to look like sim/stat
	// i don't really care what lineOffset looks like, don't expect to use it at all
	// basic line formula:               y = lineSlope * x + lineOffset
	//                      y - lineOffset = lineSlope * x
	//                  y/x - lineOffset/x = lineSlope
	//          sim/stat - lineOffset/stat = lineSlope
	//                            sim/stat = lineSlope + lineOffset/stat
	//                                 sim = lineSlope*stat + lineOffset
	sampleRow := utilhighs.ConstraintRowBuild{Debug: "sampleRow"}
	sampleRow.Add(fit.lineSlope, sample.statValue)
	sampleRow.Add(fit.lineOffset, 1)
	sampleRow.Add(difference, 1) // now technically this is a "vertical" difference, not a anything squared, but hopefully proportional...
	sampleRow.Finish(fit.input, sample.simResult, sample.simResult)

	// new absolute val with toggle, this is its test
	utilhighs.AbsoluteValue_WithToggle(fit.input, difference, differenceAbs, toggle, c_simRangeHigh)
}
