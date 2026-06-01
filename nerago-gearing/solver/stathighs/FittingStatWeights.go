package stathighs

import (
	"cmp"
	"math"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/solver/utilhighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
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

type FittingAllStatWeightProcess struct {
	printer *util.PrintRecorder
	input   *utilhighs.InputBuilder

	each util.MapMap[stats.StatType, simulate.SimResultType, FittingSingleStatWeightProcess]
}

func (fitall *FittingAllStatWeightProcess) Init(printer *util.PrintRecorder) {
	fitall.printer = printer
	fitall.input = new(utilhighs.InputBuilder)
	fitall.input.Minimise = true
}

////////////////////////////////////////////////////////

type FittingEachStatWeightProcess struct {
	printer   *util.PrintRecorder
	inputData []WeightInput
	each      util.MapMap[stats.StatType, simulate.SimResultType, *fittingEachFields]
}

type fittingEachFields struct {
	statType  stats.StatType
	simType   simulate.SimResultType
	process   FittingSingleStatSegmentsProcess
	resultMap map[StatRange]FittingSingleStatResult
}

func (fiteach *FittingEachStatWeightProcess) Init(printer *util.PrintRecorder) {
	fiteach.printer = printer
}

func (fiteach *FittingEachStatWeightProcess) SupplyDataFromStandard(inputData []WeightInput) {
	fiteach.inputData = inputData
}

func (fiteach *FittingEachStatWeightProcess) Run() util.MapMap[stats.StatType, simulate.SimResultType, map[StatRange]FittingSingleStatResult] {
	for _, statType := range G_RequiredStats {
		for _, simType := range G_RequiredSims {
			// TODO holding printer?
			fields := fittingEachFields{statType: statType, simType: simType}
			fields.process.Init(fiteach.printer, statType, simType)
			fields.process.SupplyDataFromStandard(fiteach.inputData)
			fiteach.each.Put(statType, simType, &fields)
		}
	}

	channelEach := channel_op.SeqToChannel(fiteach.each.SeqValues())
	channel_op.ForEach_Channel(10, channelEach, func(fields *fittingEachFields) {
		fields.resultMap = fields.process.Run()
	})

	resultMap := util.MapMap[stats.StatType, simulate.SimResultType, map[StatRange]FittingSingleStatResult]{}
	fiteach.each.ForeachWithKeys(func(statType stats.StatType, simType simulate.SimResultType, value *fittingEachFields) {
		resultMap.Put(statType, simType, value.resultMap)
	})
	return resultMap
}

////////////////////////////////////////////////////////

type StatRange struct {
	Minimum uint32
	Maximum uint32
}

type FittingSingleStatSegmentsProcess struct {
	printer *util.PrintRecorder

	inputDataOriginal       []*WeightInput
	inputDataRemainingParts map[StatRange][]*WeightInput
	stat                    stats.StatType
	sim                     simulate.SimResultType

	segments map[StatRange]FittingSingleStatResult
}

func (fitseg *FittingSingleStatSegmentsProcess) Init(printer *util.PrintRecorder, stat stats.StatType, sim simulate.SimResultType) {
	fitseg.printer = printer
	fitseg.segments = make(map[StatRange]FittingSingleStatResult)
	fitseg.inputDataRemainingParts = make(map[StatRange][]*WeightInput)
	fitseg.stat = stat
	fitseg.sim = sim
}

func (fitseg *FittingSingleStatSegmentsProcess) SupplyDataFromStandard(inputData []WeightInput) {
	fitseg.inputDataOriginal = util.CastSliceAsNew(inputData, func(w *WeightInput) *WeightInput { return w })
}

func (fitseg *FittingSingleStatSegmentsProcess) Run() map[StatRange]FittingSingleStatResult {
	// fitseg.runFitAll()

	fitseg.runInitial()

	overallSize := len(fitseg.inputDataOriginal)
	for len(fitseg.inputDataRemainingParts) > 0 {
		nextRange, nextData := util.MapFirstEntry(fitseg.inputDataRemainingParts)
		delete(fitseg.inputDataRemainingParts, nextRange)

		ratioOfOverall := percentRatio(len(nextData), overallSize)
		if ratioOfOverall < 0.02 || len(nextData) < 3 {
			// drop it
		} else if ratioOfOverall < 0.05 || len(nextData) < 8 {
			fitseg.runNextSegment(nextData, nextRange, 1)
		} else if ratioOfOverall < 0.15 || len(nextData) < 20 {
			fitseg.runNextSegment(nextData, nextRange, 0.8)
		} else if ratioOfOverall < 0.30 {
			fitseg.runNextSegment(nextData, nextRange, 0.4)
		} else {
			fitseg.runNextSegment(nextData, nextRange, 0.2)
		}
	}

	return fitseg.segments
}

func percentRatio(value, total int) float64 {
	return float64(value) / float64(total)
}

func (fitseg *FittingSingleStatSegmentsProcess) runFitAll() {
	fit := FittingSingleStatWeightProcess{}
	fit.Init(fitseg.printer)
	fit.SetMinimumIncludeRate(1)
	fit.SupplyDataFromStandard(fitseg.inputDataOriginal, fitseg.stat, fitseg.sim)
	weightOptional := fit.Run()
	if weight, hasWeight := weightOptional.GetWithFlag(); hasWeight {
		statRange := StatRange{weight.Minimum, weight.Maximum}
		fitseg.segments[statRange] = weight
	}
}

func (fitseg *FittingSingleStatSegmentsProcess) runInitial() {
	fit := FittingSingleStatWeightProcess{}
	fit.Init(fitseg.printer)
	fit.SetMinimumIncludeRate(0.3)
	fit.SupplyDataFromStandard(fitseg.inputDataOriginal, fitseg.stat, fitseg.sim)
	weightOptional := fit.Run()
	if weight, hasWeight := weightOptional.GetWithFlag(); hasWeight {
		statRange := StatRange{weight.Minimum, weight.Maximum}
		fitseg.segments[statRange] = weight

		totalRange := StatRange{0, c_statRangeHigh}
		fitseg.addToRemainingData(fitseg.inputDataOriginal, totalRange, statRange)
	}
}

func (fitseg *FittingSingleStatSegmentsProcess) runNextSegment(inputData []*WeightInput, inputRange StatRange, includeRate float64) {
	fit := FittingSingleStatWeightProcess{}
	fit.Init(fitseg.printer)
	fit.SetMinimumIncludeRate(includeRate)
	fit.SupplyDataFromStandard(inputData, fitseg.stat, fitseg.sim)
	weightOptional := fit.Run()
	if weight, hasWeight := weightOptional.GetWithFlag(); hasWeight {
		minimum := max(inputRange.Minimum, weight.Minimum)
		maximum := min(inputRange.Maximum, weight.Maximum)

		statRange := StatRange{minimum, maximum}
		weight.Minimum = minimum
		weight.Maximum = maximum
		fitseg.segments[statRange] = weight

		fitseg.addToRemainingData(inputData, inputRange, statRange)
	}
}

func (fitseg *FittingSingleStatSegmentsProcess) addToRemainingData(processedData []*WeightInput, inputRange StatRange, removeRange StatRange) {
	if removeRange.Minimum < inputRange.Minimum || removeRange.Maximum > inputRange.Maximum || removeRange.Minimum > removeRange.Maximum || inputRange.Minimum > inputRange.Maximum {
		panic("range isn't within bounds")
	}

	loData := make([]*WeightInput, 0)
	hiData := make([]*WeightInput, 0)
	for _, input := range processedData {
		stat := input.TotalStat.Get(fitseg.stat)
		if stat < inputRange.Minimum {
			panic("sample isn't within bounds")
		} else if inputRange.Minimum <= stat && stat < removeRange.Minimum {
			loData = append(loData, input)
		} else if removeRange.Minimum <= stat && stat <= removeRange.Maximum {
			// drop
		} else if removeRange.Maximum < stat && stat <= inputRange.Maximum {
			hiData = append(hiData, input)
		} else {
			panic("sample isn't within bounds")
		}
	}

	if len(loData) > 0 {
		loRange := StatRange{inputRange.Minimum, removeRange.Minimum - 1}
		fitseg.inputDataRemainingParts[loRange] = loData
	}

	if len(hiData) > 0 {
		hiRange := StatRange{removeRange.Maximum + 1, inputRange.Maximum}
		fitseg.inputDataRemainingParts[hiRange] = hiData
	}
}

func (fitseg *FittingSingleStatSegmentsProcess) filterDataStatRange(inputData []WeightInput, lo, hi uint32) {
	util.FilterSliceAsNew(inputData, func(in *WeightInput) bool {
		value := in.TotalStat.Get(fitseg.stat)
		return lo <= value && value <= hi
	})
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
	Minimum        uint32
	Maximum        uint32
	IncludeCount   uint32
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

	fit.lineSlope = fit.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "slope"})
	fit.lineOffset = fit.input.CreateColumnGeneral(highs.Continuous, utilhighs.C_MinusInf, utilhighs.C_PlusInf, utilhighs.DebugString{Text: "offset"})
	fit.minimumThreshold = fit.input.CreateColumnGeneral(highs.Continuous, 0, c_statRangeHigh, utilhighs.DebugString{Text: "minimum"})
	fit.maximumThreshold = fit.input.CreateColumnGeneral(highs.Continuous, 0, c_statRangeHigh, utilhighs.DebugString{Text: "maximum"})

	maxVsMin := utilhighs.ConstraintRowBuild{}
	maxVsMin.Add(fit.minimumThreshold, -1)
	maxVsMin.Add(fit.maximumThreshold, 1)
	maxVsMin.Finish(fit.input, 0, utilhighs.C_PlusInf)
}

func (fit *FittingSingleStatWeightProcess) SetMinimumIncludeRate(percent float64) {
	fit.minimumIncludeRate = percent
}

func (fit *FittingSingleStatWeightProcess) SupplyDataFromStandard(inputData []*WeightInput, stat stats.StatType, sim simulate.SimResultType) {
	fit.inputData = util.CastSliceAsNew(inputData, func(input **WeightInput) fittingSample {
		return fittingSample{
			float64((*input).TotalStat.Get(stat)),
			scaleSimItem((*input).SimResult.Get(sim), sim),
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

/*

  if (this->options_.blend_multi_objectives) {
    HighsLinearObjective& multi_linear_objective = this->multi_linear_objective_[iObj];
    lp.col_cost_[iCol] += multi_linear_objective.weight * multi_linear_objective.coefficients[iCol];
    lp.sense_ = ObjSense::kMinimize;

  ELSE PRIORITIES

  for (HighsInt iObj = 0; iObj < num_linear_objective; iObj++)
    priority_objective.push_back(std::make_pair(this->multi_linear_objective_[iObj].priority, iObj));
  std::sort(priority_objective.begin(), priority_objective.end(), comparison);

  for (HighsInt iIx = 0; iIx < num_linear_objective; iIx++) {
    lp.col_cost_ = linear_objective.coefficients;
	lp.sense_ = linear_objective.weight > 0 ? ObjSense::kMinimize : ObjSense::kMaximize;
	use previous round's solution as start, needs to be valid for this too
    HighsStatus optimize_model_status = this->optimizeModel();
	if (lp.isMip())
		save col values for next

  	std::vector<HighsInt> index(lp.num_col_);
  	std::vector<double> value(lp.num_col_);
	for (HighsInt iCol = 0; iCol < lp.num_col_; iCol++) {
      if (lp.col_cost_[iCol]) {
        index[nnz] = iCol;
        value[nnz] = lp.col_cost_[iCol];
        nnz++;
      }

	double lower_bound = -kHighsInf;
    double upper_bound = kHighsInf;
	if (lp.sense_ == ObjSense::kMinimize)
	  if (linear_objective.abs_tolerance >= 0)
        upper_bound = objective + linear_objective.abs_tolerance;
      if (linear_objective.rel_tolerance >= 0) {
	    if (objective >= 0) {
		  upper_bound = std::min(objective * (1.0 + linear_objective.rel_tolerance), upper_bound);
		else
		  upper_bound = std::min(objective * (1.0 - linear_objective.rel_tolerance), upper_bound);
	addRow(lower_bound, upper_bound, nnz, index, value);
*/

func (fit *FittingSingleStatWeightProcess) Run() util.Optional[FittingSingleStatResult] {
	fit.setupLinearObjectives()

	for sample := range util.ForPointer(fit.inputData) {
		fit.addSample(sample)
	}

	fit.includeCountRow.Finish(fit.input, float64(len(fit.inputData))*fit.minimumIncludeRate, utilhighs.C_PlusInf)

	solution, log := fit.input.RunHighs()
	fit.printer.AppendOther(log)
	fit.printer.Println(solution.Status.String())

	fit.input.DebugPrintColumns(solution, fit.printer)

	if solution.IsOptimal() {
		return util.Optional_OfValue(fit.buildResult(solution))
	} else {
		return util.Optional_Empty[FittingSingleStatResult]()
	}
}

func (fit *FittingSingleStatWeightProcess) setupLinearObjectives() {

	// var scaleIncludeObjective float64
	// scaleIncludeObjective = 20 // 20 used previously for the 256 samples, maybe a touch too high
	// scaleIncludeObjective = 19 // 15 is too low for 256, 19 about right
	// scaleIncludeObjective = 12 // for 2000: 10-12 is a bit low. 15 was nice maybe too high? ideal range somewhere 13-14, but runs take 1H
	// these numbers are just too finicky

	// fit.input.BlendMultiObjectives = true
	// fit.linearInclude = fit.input.AddLinearObjective(scaleIncludeObjective, 0, 1000, 1, 1)
	// fit.linearLineDiff = fit.input.AddLinearObjective(1, 0, 10000, 5, 2)

	// averageIncludedDifference := 14 // based on data, haste_dps. rather not use this

	fit.input.BlendMultiObjectives = false

	var relativeToleranceParam float64
	if fit.minimumIncludeRate < 1 {
		// first linear step find a regular solution to the line fit
		// will probably follow the minimum required include
		// will get us a positive initial result from the sum of differenceAbs
		// let it expand to full coverage if it wants, but without worsening the average difference
		multiplierToFullCoverage := 1 / fit.minimumIncludeRate
		// add consider a bit of factor to this, only 80% etc, otherwise might get too greedy
		multiplierToFullCoverage *= 0.5
		// highs logic is "objective * (1.0 + linear_objective.rel_tolerance)", so need to minus one in compenstation
		// don't let it go negative or below a small value
		relativeToleranceParam = max(multiplierToFullCoverage-1, 0.1)
	} else {
		relativeToleranceParam = 0
	}
	fit.linearLineDiff = fit.input.AddLinearPrioritised(false, -1, relativeToleranceParam, 2)

	// second priority is sum of includeColumn which are negative one each, can lead to negative total objective
	// but we don't need to care about offsets much since its the last one, highs shouldn't even look at them
	fit.linearInclude = fit.input.AddLinearPrioritised(false, -1, -1, 1)

	// we might want to increase c_outputIncludePerInclude a bit since average at least for our first test case is 3-10
	// but actually unless we combine the objectives then they aren't getting scaled against each other anyway
}

func (fit *FittingSingleStatWeightProcess) buildResult(solution *highs.Solution) FittingSingleStatResult {
	result := FittingSingleStatResult{}
	result.LineSlope = solution.ColValues[fit.lineSlope] / fit.inputDataSimScale
	result.LineOffset = solution.ColValues[fit.lineOffset] / fit.inputDataSimScale
	result.Minimum = uint32(math.Round(solution.ColValues[fit.minimumThreshold]))
	result.Maximum = uint32(math.Round(solution.ColValues[fit.maximumThreshold]))

	var includeCount uint32 = 0
	for _, col := range fit.includeColumns {
		if utilhighs.FloatEqualsOne(solution.ColValues[col]) {
			includeCount++
		}
	}
	result.IncludeCount = includeCount
	result.IncludePercent = float64(includeCount) / float64(len(fit.inputData))

	if includeCount == 0 {
		panic("shouldn't this have failed in model")
	}

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

	// another thought, samples could be presorted and indexed, then we setup relationships between adjactent pairs,
	// they pull each other up, until we reach a sample marked as THE high/low cutoff

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
