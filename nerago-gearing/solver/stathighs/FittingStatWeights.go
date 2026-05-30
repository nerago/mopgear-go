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

	setmin := utilhighs.ConstraintRowBuild{}
	setmin.Add(fit.minimumThreshold, 1)
	setmin.Finish(fit.input, 4000, 4000)

	setmax := utilhighs.ConstraintRowBuild{}
	setmax.Add(fit.maximumThreshold, 1)
	setmax.Finish(fit.input, 6000, 6000)

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

	// at the moment includes work in a positive direction: only in range samples are included
	// but the reverse direction: force all valid samples to be true, not so much

	// if include:   stat - max <= 0     ->>      stat <= max
	// if not:          0 - max <= 0     ->>      max >= 0, (max free)
	includeRowMax := utilhighs.ConstraintRowBuild{Debug: "includeRowMax"}
	includeRowMax.Add(includeColumn, sample.statValue)
	includeRowMax.Add(fit.maximumThreshold, -1)
	includeRowMax.Finish(fit.input, utilhighs.C_MinusInf, 0)

	// if include: range + min <= range + stat  ->>  min <= stat
	// if not:             min <= range + stat  ->>  min is free
	includeRowMin := utilhighs.ConstraintRowBuild{Debug: "includeRowMin"}
	includeRowMin.Add(includeColumn, c_statRangeHigh)
	includeRowMin.Add(fit.minimumThreshold, 1)
	includeRowMin.Finish(fit.input, utilhighs.C_MinusInf, c_statRangeHigh+sample.statValue)

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
