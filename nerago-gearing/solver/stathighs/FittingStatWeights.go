package stathighs

import (
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"

	"github.com/bartolsthoorn/gohighs/highs"
)

// STR example: 30554
const c_statRangeHigh = 50000

// TPS example: 1957667
const c_simRangeHigh = 5000000

const c_outputFittingDifference = 1
const c_outputIncludePerInclude = -1

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

type FittingSingleStatWeightProcess struct {
	printer *util.PrintRecorder
	input   *utilhighs.InputBuilder

	minimumIncludeRate float64
	inputData          []fittingSample

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

func (fit *FittingSingleStatWeightProcess) SupplyData(inputData []fittingSample) {
	fit.inputData = inputData
}

func (fit *FittingSingleStatWeightProcess) SupplyDataFromStandard(inputData []WeightInput, stat stats.StatType, sim simulate.SimResultType) {
	fit.inputData = util.CastSliceAsNew(inputData, func(input *WeightInput) fittingSample {
		return fittingSample{
			float64(input.TotalStat.Get(stat)),
			input.SimResult.GetFriendly(sim),
			-1,
		}
	})
}

func (fit *FittingSingleStatWeightProcess) Run() FittingSingleStatResult {
	fit.lineSlope = fit.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "slope"})
	fit.lineOffset = fit.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "offset"})
	fit.minimumThreshold = fit.input.CreateColumnGeneral(highs.Continuous, 0, c_statRangeHigh, utilhighs.DebugString{Text: "minimum"})
	fit.maximumThreshold = fit.input.CreateColumnGeneral(highs.Continuous, 0, c_statRangeHigh, utilhighs.DebugString{Text: "maximum"})

	maxVsMin := utilhighs.ConstraintRowBuild{}
	maxVsMin.Add(fit.minimumThreshold, -1)
	maxVsMin.Add(fit.maximumThreshold, 1)
	maxVsMin.Finish(fit.input, 0, utilhighs.C_PlusInf)

	// setmin := utilhighs.ConstraintRowBuild{}
	// setmin.Add(fit.minimumThreshold, 1)
	// setmin.Finish(fit.input, 4000, 4000)

	// setmax := utilhighs.ConstraintRowBuild{}
	// setmax.Add(fit.maximumThreshold, 1)
	// setmax.Finish(fit.input, 6000, 6000)

	// so we could introduce rows to set min/max to specific entries that exact match
	// could make this diffcult with likely duplicates
	// but otherwise might make the math easier?
	// likely that vertex search will do it anyway

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
	result.LineSlope = solution.ColValues[fit.lineSlope]
	result.LineOffset = solution.ColValues[fit.lineOffset]
	result.Minimum = solution.ColValues[fit.minimumThreshold]
	result.Maximum = solution.ColValues[fit.maximumThreshold]

	includeCount := 0
	for _, col := range fit.includeColumns {
		if utilhighs.FloatEqualsOne(solution.ColValues[col]) {
			includeCount++
		}
	}
	result.IncludePercent = float64(includeCount) / float64(len(fit.inputData))

	for sample := range util.ForPointer(fit.inputData) {
		fit.printer.Printf("INC %f %f\n", sample.statValue, solution.ColValues[sample.includeColumn])
	}

	return result
}

func (fit *FittingSingleStatWeightProcess) addSample(sample *fittingSample) {
	includeColumn := fit.sampleIncludeToggleColumn(sample)
	fit.sampleToFitLine(sample, includeColumn)
}

func (fit *FittingSingleStatWeightProcess) sampleIncludeToggleColumn(sample *fittingSample) utilhighs.ColumnIndex {
	includeColumn := fit.input.CreateColumnWithOutput(highs.Integer, 0, 1, c_outputIncludePerInclude, utilhighs.DebugString{Text: "include"})
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

func (fit *FittingSingleStatWeightProcess) makeIsOverMinimum(statValue float64) utilhighs.ColumnIndex {
	isOverMinimum := fit.input.CreateColumnBool(utilhighs.DebugString{Text: "isOverMinimum"})

	// ORIGINAL
	// if overmin:      1*range + min <= range + stat  ->>  min <= stat
	// if !overmin:     0*range + min <= range + stat  ->>  min is free
	// if stat > min:   x*range + min <= range + stat  ->>  x*range <= range + stat - min  ->>  x*range <= range + small_positive   ->>   x = 0 or 1
	// if stat < min:   x*range + min <= range + stat  ->>  x*range <= range + stat - min  ->>  x*range <= range + small_negative   ->>   x = 0
	checkIsOverMin := utilhighs.ConstraintRowBuild{Debug: "checkIsOverMin"}
	checkIsOverMin.Add(isOverMinimum, c_statRangeHigh)
	checkIsOverMin.Add(fit.minimumThreshold, 1)
	checkIsOverMin.Finish(fit.input, utilhighs.C_MinusInf, c_statRangeHigh+statValue)

	//   min - stat + x.range >= 0   ->>   min + x.range >= stat
	// if stat > min  ->>  min - stat + x.range >= 0  ->>  small_negative + x.range >= 0  ->>  x=1
	// if stat < min  ->>  min - stat + x.range >= 0  ->>  small_positive + x.range >= 0  ->>  x=0 or 1
	// if overmin     ->>  min - stat + 1.range >= 0  ->>  min >= stat - range   ->>   min is free
	// if !overmin    ->>  min - stat + 0.range >= 0  ->>  min >= stat    ->>   min is free
	setIfOverMin := utilhighs.ConstraintRowBuild{Debug: "setIfOverMin"}
	setIfOverMin.Add(fit.minimumThreshold, 1)
	setIfOverMin.Add(isOverMinimum, c_statRangeHigh)
	setIfOverMin.Finish(fit.input, statValue, utilhighs.C_PlusInf)

	return isOverMinimum
}

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

func (fit *FittingSingleStatWeightProcess) sampleIncludeToggleColumn_old(sample *fittingSample) utilhighs.ColumnIndex {
	includeColumn := fit.input.CreateColumnWithOutput(highs.Integer, 0, 1, c_outputIncludePerInclude, utilhighs.DebugString{Text: "include"})
	fit.includeCountRow.Add(includeColumn, 1)
	fit.includeColumns = append(fit.includeColumns, includeColumn)
	sample.includeColumn = includeColumn

	// at the moment includes work in a positive direction: only in range samples are included
	// but the reverse direction: force all valid samples to be true, not so much

	// if include:   stat - max <= 0     ->>      stat <= max
	// if not:          0 - max <= 0     ->>      max >= 0, (max free)
	includedRowLessThanMax := utilhighs.ConstraintRowBuild{Debug: "includeRowMax"}
	includedRowLessThanMax.Add(includeColumn, sample.statValue)
	includedRowLessThanMax.Add(fit.maximumThreshold, -1)
	includedRowLessThanMax.Finish(fit.input, utilhighs.C_MinusInf, 0)

	// if include: range + min <= range + stat  ->>  min <= stat
	// if not:             min <= range + stat  ->>  min is free
	includedRowGreaterThanMin := utilhighs.ConstraintRowBuild{Debug: "includeRowMin"}
	includedRowGreaterThanMin.Add(includeColumn, c_statRangeHigh)
	includedRowGreaterThanMin.Add(fit.minimumThreshold, 1)
	includedRowGreaterThanMin.Finish(fit.input, utilhighs.C_MinusInf, c_statRangeHigh+sample.statValue)

	// basic rule:  min <= stat <= max
	// basic rule:  0 <= stat-min <= max-min  how does that help?

	// so want something where inclue=false, but value in range is invalid
	// so bad:  min <= stat + 0.include <= max
	// but only bad when its both

	//
	excludedRowNotInRange := utilhighs.ConstraintRowBuild{Debug: ""}
	excludedRowNotInRange.Add(includeColumn, 1)

	return includeColumn
}

func (fit *FittingSingleStatWeightProcess) sampleToFitLine(sample *fittingSample, toggle utilhighs.ColumnIndex) {
	difference := fit.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "difference"})
	differenceAbs := fit.input.CreateColumnWithOutput(highs.Continuous, 0, utilhighs.C_PlusInf, c_outputFittingDifference, utilhighs.DebugString{Text: "differenceAbs"})

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
