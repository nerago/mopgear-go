package weightfind

import (
	"cmp"
	"math"
	"paladin_gearing_go/solver/stathighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/channel_op"
	"paladin_gearing_go/util/util_rank"
	"slices"
)

const (
	c_search2_largeAccuracyGap    = 0.5
	c_search2_marginalAccuracyGap = 0.005
	c_search2_marginalWeightGap   = 0.001
	c_search2_minRunEarlySizeCut  = 4
	c_search2_minRunLateSizeCut   = 2
)

type opType int8

const (
	opDivide1 opType = iota
	opSearch  opType = iota
	opFinal   opType = iota
)

type WeightSearcher2 struct {
	typeCount   int
	statTypes   []stats.StatType
	simTypes    []stats.SimType
	targetRatio stats.SimData
	inputData   []stathighs.WeightInput
	printer     *util.PrintRecorder

	queue      util.QueueStackFilo[*weightSearch2Bound]
	bestResult util_rank.BestCollector1[stathighs.WeightResult]
	bestBound  util_rank.BestCollector1[stathighs.WeightResult]
}

type weightSearch2Bound struct {
	plannedOp opType
	axisFocus int
	rangeMin  []float64
	rangeMax  []float64
	parent    *weightSearch2Bound // not sure
}

func (ws *WeightSearcher2) Init(statTypes []stats.StatType, targetRatio stats.SimData, printer *util.PrintRecorder) {
	ws.typeCount = len(statTypes)
	ws.statTypes = statTypes
	ws.simTypes = targetRatio.NonZeroTypes()
	ws.targetRatio = targetRatio
	ws.printer = printer
}

func (ws *WeightSearcher2) SupplyData(inputData []stathighs.WeightInput) {
	ws.inputData = inputData
}

func (ws *WeightSearcher2) SetRanges(weightMin, weightMax float64) {
	ws.queue.Push(&weightSearch2Bound{
		plannedOp: opDivide1,
		axisFocus: 0,
		rangeMin:  util.RepeatValue(weightMin, ws.typeCount),
		rangeMax:  util.RepeatValue(weightMax, ws.typeCount),
	})
}

func (ws *WeightSearcher2) Run(cancel channel_op.CancelSignal) stathighs.WeightResult {
	for cancel.ShouldContinue() {
		bound, hasValue := ws.queue.Pop()
		if !hasValue {
			break
		}

		// potential single axis search

		switch bound.plannedOp {
		case opDivide1:
			ws.opDivide1(bound)
		case opSearch:
			ws.opSearch(bound)
		case opFinal:
			ws.opFinal(bound)
		}
	}

	bestWeight := ws.bestResult.GetBestOrNilValue()
	return bestWeight.ScaleForBaseStat(ws.statTypes[0])
}

func (ws *WeightSearcher2) evaluateScore(weightArray []float64) float64 {
	weights := stathighs.WeightResult_Make()
	for i, statType := range ws.statTypes {
		weights.Put(statType, weightArray[i])
	}
	accuracy := EvaluateAccuracyNoRangeInlined2(weights, ws.simTypes, ws.targetRatio, ws.inputData)
	ws.bestResult.Offer(&weights, accuracy)
	return accuracy
}

// dumb division in halves
func (ws *WeightSearcher2) opDivide1(bound *weightSearch2Bound) {
	axis := bound.axisFocus
	mid := (bound.rangeMin[axis] + bound.rangeMax[axis]) / 2

	var nextOp opType
	var nextAxis int
	if axis < ws.typeCount-1 {
		nextOp = opDivide1
		nextAxis = axis + 1
	} else {
		nextOp = opSearch
		nextAxis = 0
	}

	ws.queue.Push(&weightSearch2Bound{
		plannedOp: nextOp,
		axisFocus: nextAxis,
		rangeMin:  bound.rangeMin,
		rangeMax:  copyAndReplaceElement(bound.rangeMax, axis, mid),
	})
	ws.queue.Push(&weightSearch2Bound{
		plannedOp: nextOp,
		axisFocus: nextAxis,
		rangeMin:  copyAndReplaceElement(bound.rangeMin, axis, mid),
		rangeMax:  bound.rangeMax,
	})
}

func (ws *WeightSearcher2) opFinal(bound *weightSearch2Bound) {
	middle := sliceInterpolate(bound.rangeMin, bound.rangeMax, 0.5)
	ws.evaluateScore(middle)
	ws.printer.Println("opFinal STOP")
}

type indexAndAccuracy struct {
	index    int
	accuracy float64
}

/*
	                    narrow          corners        side/top        extended
	                    middle                                         corner
		                0 > 1234        12 > 034       13 > 024         04 > 123
	                                    34 > 012       ~14,23,24

.............    .............   ......|......   ......|......   ..|..........
..1.......4..    ..?-------?..   ..1...|......   ..1...|......   ..?.......4..
.............    ..|.......|..   ......|......   ......|......   ..|..........
......0......    ..|...0...|..   ------?------   ......?......   ..|...0......
.............    ..|.......|..   ......|......   ......|......   ..|..........
..3.......2..    ..?-------?..   ......|...2..   ..3...|......   ..?-------?--
.............    .............   ......|......   ......|......   .............
*/
func (ws *WeightSearcher2) opSearch(bound *weightSearch2Bound) {
	printRange(bound.rangeMin, bound.rangeMax, "??? ", ws.printer)
	debugValueOfInterest(bound.rangeMin, bound.rangeMax, ws.printer)

	// selected points throughout the sample space. could have more variety of axis points.
	var probes [5][]float64
	probes[0] = sliceInterpolate(bound.rangeMin, bound.rangeMax, 1.0/2.0)
	probes[1] = sliceInterpolate(bound.rangeMin, bound.rangeMax, 1.0/4.0)
	probes[2] = sliceInterpolate(bound.rangeMin, bound.rangeMax, 3.0/4.0)
	probes[3] = sliceMixEverySecond(probes[1], probes[2]) // first is the X on above diagrams, second the Y
	probes[4] = sliceMixEverySecond(probes[2], probes[1])

	// evaluate each probe, sort descending so best are first
	var values [5]indexAndAccuracy
	for i := range 5 {
		values[i] = indexAndAccuracy{i, ws.evaluateScore(probes[i])}
	}
	for i := range 5 {
		ws.printer.Printf("     * ")
		for z := range len(bound.rangeMin) {
			ws.printer.Printf("%7.4f ", probes[i][z])
		}
		ws.printer.Printf(" ==> %f ", values[i].accuracy)
		ws.printer.Println0()
	}
	slices.SortStableFunc(values[:], func(a, b indexAndAccuracy) int { return cmp.Compare(b.accuracy, a.accuracy) })

	// determine next step based on ranking of probes
	if ws.marginalAccuracyGap(values[0].accuracy, values[4].accuracy) {
		//if ws.marginalGap(values[0].accuracy, values[1].accuracy) {
		ws.printer.Println("  -> MARGINAL GAP STOP")
		// minimal gap between probes, perhaps nothing more to explore
	} else if ws.largeAccuracyGap(values[0].accuracy, values[1].accuracy) || ws.firstGreaterThanEqualPair(&values) {
		ws.printer.Println("GAP1")
		// something is best by a large margin
		var rangeMin, rangeMax []float64
		switch values[0].index {
		case 0:
			rangeMin = probes[1]
			rangeMax = probes[2]
		case 1:
			rangeMin = bound.rangeMin
			rangeMax = probes[0]
		case 2:
			rangeMin = probes[0]
			rangeMax = bound.rangeMax
		case 3:
			rangeMin = sliceMixEverySecond(bound.rangeMin, probes[0])
			rangeMax = sliceMixEverySecond(probes[0], bound.rangeMax)
		case 4:
			rangeMin = sliceMixEverySecond(probes[0], bound.rangeMin)
			rangeMax = sliceMixEverySecond(bound.rangeMax, probes[0])
		}
		debugVerifyInRange(rangeMin, rangeMax, probes[values[0].index], ws.printer)
		ws.addSearchPlan(rangeMin, rangeMax)
	} else if oneOfEqualsQuery(values[0].index, values[1].index, 0) {
		// the two best includes the middle
		other := values[0].index
		if other == 0 {
			other = values[1].index
		}

		ws.printer.Printf("MIDDLE PLUS %d\n", other)

		var rangeMin, rangeMax []float64
		switch other {
		case 1:
			rangeMin = bound.rangeMin
			rangeMax = probes[2]
		case 2:
			rangeMin = probes[1]
			rangeMax = bound.rangeMax
		case 3:
			rangeMin = sliceMixEverySecond(bound.rangeMin, probes[4])
			rangeMax = sliceMixEverySecond(probes[4], bound.rangeMax)
		case 4:
			rangeMin = sliceMixEverySecond(probes[3], bound.rangeMin)
			rangeMax = sliceMixEverySecond(bound.rangeMax, probes[3])
		}
		debugVerifyInRange(rangeMin, rangeMax, probes[0], ws.printer)
		debugVerifyInRange(rangeMin, rangeMax, probes[other], ws.printer)
		ws.addSearchPlan(rangeMin, rangeMax)
	} else if pairEqualsQueryPair(values[0].index, values[1].index, 1, 3) {
		ws.printer.Println("SIDE")
		rangeMin := bound.rangeMin
		rangeMax := sliceMixEverySecond(probes[0], bound.rangeMax)
		debugVerifyInRange(rangeMin, rangeMax, probes[1], ws.printer)
		debugVerifyInRange(rangeMin, rangeMax, probes[3], ws.printer)
		ws.addSearchPlan(rangeMin, rangeMax)
	} else if pairEqualsQueryPair(values[0].index, values[1].index, 1, 4) {
		ws.printer.Println("SIDE")
		rangeMin := bound.rangeMin
		rangeMax := sliceMixEverySecond(bound.rangeMax, probes[0])
		debugVerifyInRange(rangeMin, rangeMax, probes[1], ws.printer)
		debugVerifyInRange(rangeMin, rangeMax, probes[4], ws.printer)
		ws.addSearchPlan(rangeMin, rangeMax)
	} else if pairEqualsQueryPair(values[0].index, values[1].index, 2, 3) {
		ws.printer.Println("SIDE")
		rangeMin := sliceMixEverySecond(bound.rangeMin, probes[0])
		rangeMax := bound.rangeMax
		debugVerifyInRange(rangeMin, rangeMax, probes[2], ws.printer)
		debugVerifyInRange(rangeMin, rangeMax, probes[3], ws.printer)
		ws.addSearchPlan(rangeMin, rangeMax)
	} else if pairEqualsQueryPair(values[0].index, values[1].index, 2, 4) {
		ws.printer.Println("SIDE")
		rangeMin := sliceMixEverySecond(probes[0], bound.rangeMin)
		rangeMax := bound.rangeMax
		debugVerifyInRange(rangeMin, rangeMax, probes[2], ws.printer)
		debugVerifyInRange(rangeMin, rangeMax, probes[4], ws.printer)
		ws.addSearchPlan(rangeMin, rangeMax)
	} else {
		// remaining two are not adjacent, queue them separately
		for check := range 2 {
			var rangeMin, rangeMax []float64
			index := values[check].index
			ws.printer.Printf("CORNER %d\n", index)
			switch index {
			case 1:
				rangeMin = bound.rangeMin
				rangeMax = probes[0]
			case 2:
				rangeMin = probes[0]
				rangeMax = bound.rangeMax
			case 3:
				rangeMin = sliceMixEverySecond(bound.rangeMin, probes[0])
				rangeMax = sliceMixEverySecond(probes[0], bound.rangeMax)
			case 4:
				rangeMin = sliceMixEverySecond(probes[0], bound.rangeMin)
				rangeMax = sliceMixEverySecond(bound.rangeMax, probes[0])
			default:
				panic("logic error")
			}
			debugVerifyInRange(rangeMin, rangeMax, probes[index], ws.printer)
			ws.addSearchPlan(rangeMin, rangeMax)
		}
	}
}

type probeAndAccuracy struct {
	accuracy float64
	axis     int
	isHigh   bool
	point    []float64
}

func (ws *WeightSearcher2) opSearch2(bound *weightSearch2Bound) {
	middle := sliceInterpolate(bound.rangeMin, bound.rangeMax, 0.5)
	probes := ws.search2DoProbes(bound, middle)

	if ws.marginalAccuracyGap(probes[0].accuracy, probes[len(probes)-1].accuracy) {
		ws.printer.Println("  -> MARGINAL GAP STOP")
		return
	}

	cutPoint := ws.search2ChooseCut(probes)
	probes = probes[0 : cutPoint+1]
	ws.search2ChooseSplitMode(probes, bound, middle)
}

func (ws *WeightSearcher2) search2DoProbes(bound *weightSearch2Bound, middle []float64) []probeAndAccuracy {
	probes := make([]probeAndAccuracy, 0)
	probes[0] = probeAndAccuracy{point: middle, axis: -1, accuracy: ws.evaluateScore(middle)}
	for axis := range ws.typeCount {
		lo := slices.Clone(middle)
		lo[axis] = valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], 0.25)
		probes = append(probes, probeAndAccuracy{point: lo, axis: axis, isHigh: false, accuracy: ws.evaluateScore(lo)})

		hi := slices.Clone(middle)
		hi[axis] = valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], 0.75)
		probes = append(probes, probeAndAccuracy{point: hi, axis: axis, isHigh: true, accuracy: ws.evaluateScore(hi)})
	}
	slices.SortStableFunc(probes, func(a, b probeAndAccuracy) int { return cmp.Compare(b.accuracy, a.accuracy) })
	return probes
}

// something like top 1/4, unless a run
// so typically get axis*2+1 = 17
// even for dps is = 11

// returns last index to include
func (ws *WeightSearcher2) search2ChooseCut(probes []probeAndAccuracy) int {
	// cut after a large gap
	for index := range len(probes) / 2 {
		if ws.largeAccuracyGap(probes[index].accuracy, probes[index+1].accuracy) {
			return index
		}
	}

	// cut after a run of consecutive values
	for runStartIndex := range len(probes) / 3 {
		if ws.marginalAccuracyGap(probes[runStartIndex].accuracy, probes[runStartIndex+1].accuracy) {
			runSize := 2
			for check := runStartIndex + 1; check < len(probes)-1; check++ {
				if ws.marginalAccuracyGap(probes[check].accuracy, probes[check+1].accuracy) {
					runSize++
				} else {
					break
				}
			}
			if runStartIndex <= 1 && runSize >= c_search2_minRunEarlySizeCut && runSize < len(probes) {
				return runStartIndex + runSize - 1
			} else if runStartIndex > 1 && runSize >= c_search2_minRunLateSizeCut && runSize < len(probes) {
				return runStartIndex + runSize - 1
			}
		}
	}

	// default, cut at one quarter of range
	return len(probes) / 4
}

// we can see each probe is adjacent to the center
// and as such any other pair that isn't a direct opposite are diagonally adjacent
// a version of the algorithm might use than and probe in pairs specifically

// we could also say that any number of lows/highs are collectively easily grouped

// we could alternately rule that if both the high+low for an axis are well ranked then keep full range there

func (ws *WeightSearcher2) search2ChooseSplitMode(probes []probeAndAccuracy, bound *weightSearch2Bound, middle []float64) {
	includeMiddle := false
	hi, lo := 0, 0
	for i := range probes {
		entry := &probes[i]
		if entry.axis == -1 {
			includeMiddle = true
		} else if entry.isHigh {
			hi++
		} else {
			lo++
		}
	}

	if hi == 0 && lo == 0 {
		// shrink? or divide?
	} else if hi == 0 || lo == 0 {
		ws.search2MakeFollowupCommon(probes, bound, includeMiddle, middle)
		return
	} else if (hi == 1 && lo > 1) || (lo == 1 && hi > 1) { // partly making sure we don't have just same axis repeated
		ws.search2MakeFollowupCommon(probes, bound, includeMiddle, middle)
		return
	}

	// else if len==2 && axis==axis, complex, middle or not

	hiSlice := make([]probeAndAccuracy, 0, hi)
	loSlice := make([]probeAndAccuracy, 0, lo)
	for i := range probes {
		entry := &probes[i]
		if entry.axis == -1 {
		} else if entry.isHigh {
			hiSlice = append(hiSlice, *entry)
		} else {
			loSlice = append(loSlice, *entry)
		}
	}

	//return hiSlice, loSlice
}

func (ws *WeightSearcher2) search2MakeFollowupCommon(probes []probeAndAccuracy, bound *weightSearch2Bound, includeMiddle bool, middle []float64) {
	didNarrow := false
	rangeMin := slices.Clone(bound.rangeMin)
	rangeMax := slices.Clone(bound.rangeMax)
	for axis := range middle {
		hasHi, hasLo := includesAxisEntry(probes, axis)
		if hasHi && !hasLo {
			if includeMiddle {
				rangeMin[axis] = valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], 0.25)
			} else {
				rangeMin[axis] = middle[axis]
			}
			didNarrow = true
		} else if !hasHi && hasLo {
			if includeMiddle {
				rangeMax[axis] = valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], 0.75)
			} else {
				rangeMax[axis] = middle[axis]
			}
			didNarrow = true
		}
	}

	if !didNarrow {
		panic("logic fail, no narrowing found")
	}
}

func includesAxisEntry(probes []probeAndAccuracy, axis int) (bool, bool) {
	hasHi, hasLo := false, false
	for i := range probes {
		entry := &probes[i]
		if entry.axis == axis {
			if entry.isHigh {
				hasHi = true
			} else {
				hasLo = true
			}
		}
	}
	return hasHi, hasLo
}

var interest = []float64{
	5.3900000000,
	7.4731250000,
	2.6908057851,
	7.4200000000,
	-0.5654897163,
	8.0470000000,
	4.8125000000,
	3.2600000000,
}

func debugValueOfInterest(rangeMin []float64, rangeMax []float64, printer *util.PrintRecorder) {
	if checkInRange(rangeMin, rangeMax, interest) {
		printer.Println("INTEREST")
	}
}

func checkInRange(rangeMin []float64, rangeMax []float64, probe []float64) bool {
	for i := range probe {
		if rangeMin[i] <= probe[i] && probe[i] <= rangeMax[i] {
			// ok
		} else {
			return false
		}
	}
	return true
}

func debugVerifyInRange(rangeMin []float64, rangeMax []float64, probe []float64, printer *util.PrintRecorder) {
	if !checkInRange(rangeMin, rangeMax, probe) {
		panic("probe value isn't inside remaining range")
	}

	printRange(rangeMin, rangeMax, "     = ", printer)
}

func printRange(rangeMin []float64, rangeMax []float64, label string, printer *util.PrintRecorder) {
	printer.Printf(label)
	for i := range rangeMin {
		printer.Printf("%7.4f ", rangeMin[i])
	}
	printer.Println0()
	printer.Printf(label)
	for i := range rangeMax {
		printer.Printf("%7.4f ", rangeMax[i])
	}
	printer.Println0()
}

func (ws *WeightSearcher2) addSearchPlan(rangeMin []float64, rangeMax []float64) {
	if ws.rangeIsMarginal(rangeMin, rangeMax) {
		ws.queue.Push(&weightSearch2Bound{
			plannedOp: opFinal,
			rangeMin:  rangeMin,
			rangeMax:  rangeMax,
		})
	} else {
		ws.queue.Push(&weightSearch2Bound{
			plannedOp: opSearch,
			rangeMin:  rangeMin,
			rangeMax:  rangeMax,
		})
	}
}

func (ws *WeightSearcher2) rangeIsMarginal(rangeMin []float64, rangeMax []float64) bool {
	for i := range rangeMin {
		if rangeMax[i]-rangeMin[i] > c_search2_marginalWeightGap {
			return false
		}
	}
	return true
}

func oneOfEqualsQuery(a, b int, query int) bool {
	return a == query || b == query
}

func pairEqualsQueryPair(a, b int, query1, query2 int) bool {
	return (a == query1 && b == query2) || (a == query2 && b == query1)
}

func (ws *WeightSearcher2) largeAccuracyGap(a, b float64) bool {
	return math.Abs(a-b) >= c_search2_largeAccuracyGap
}

func (ws *WeightSearcher2) firstGreaterThanEqualPair(values *[5]indexAndAccuracy) bool {
	return !util.FloatsApproxEquals(values[0].accuracy, values[1].accuracy) &&
		util.FloatsApproxEquals(values[1].accuracy, values[2].accuracy)
}

func (ws *WeightSearcher2) marginalAccuracyGap(a, b float64) bool {
	return math.Abs(a-b) < c_search2_marginalAccuracyGap
}

func sliceInterpolate(rangeMin []float64, rangeMax []float64, ratio float64) []float64 {
	result := make([]float64, len(rangeMin))
	for i := range len(rangeMin) {
		result[i] = valueInterpolate(rangeMin[i], rangeMax[i], ratio)
	}
	return result
}

func valueInterpolate(rangeMin float64, rangeMax float64, ratio float64) float64 {
	return rangeMin + (rangeMax-rangeMin)*ratio
	//result = (1-ratio)*rangeMin + rangeMax*ratio
}

func sliceMixEverySecond(a, b []float64) []float64 {
	result := make([]float64, len(a))
	for i := range len(a) {
		if i&1 == 0 {
			result[i] = a[i]
		} else {
			result[i] = b[i]
		}
	}
	return result
}

func copyAndReplaceElement(slice []float64, index int, value float64) []float64 {
	newSlice := slices.Clone(slice)
	newSlice[index] = value
	return newSlice
}
