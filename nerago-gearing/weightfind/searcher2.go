package weightfind

import (
	"cmp"
	"math"
	"paladin_gearing_go/solver/stathighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_rank"
	"slices"
)

const (
	c_search2_largeGap    = 1.0
	c_search2_marginalGap = 0.01
)

type opType int8

const (
	opUnknown opType = iota
	opDivide1 opType = iota
	opSearch  opType = iota
)

type WeightSearcher2 struct {
	typeCount   int
	statTypes   []stats.StatType
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

func (ws *WeightSearcher2) Run() stathighs.WeightResult {
	for {
		bound, hasValue := ws.queue.Pop()
		if !hasValue {
			break
		}

		// TODO minimum dimensions
		// potential single axis search

		switch bound.plannedOp {
		case opDivide1:
			ws.opDivide1(bound)
		case opSearch:
			ws.opSearch(bound)
		}
	}
	return ws.bestResult.GetBestOrPanic()
}

func (ws *WeightSearcher2) evaluateScore(weightArray []float64) float64 {
	weights := stathighs.WeightResult_Make()
	for i, statType := range ws.statTypes {
		weights.Put(statType, weightArray[i])
	}
	accuracy := EvaluateAccuracy(weights, ws.inputData, ws.targetRatio)
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
	// selected points throughout the sample space. could have more variety of axis points.
	var probes [5][]float64
	probes[0] = sliceInterpolate(bound.rangeMin, bound.rangeMax, 1.0/2.0)
	probes[1] = sliceInterpolate(bound.rangeMin, bound.rangeMax, 1.0/4.0)
	probes[2] = sliceInterpolate(bound.rangeMin, bound.rangeMax, 3.0/4.0)
	probes[3] = sliceMixEverySecond(probes[1], probes[2]) // first is the X on above diagrams, second the Y
	probes[4] = sliceMixEverySecond(probes[2], probes[1])

	// evaluate each probe, sort descending so best are first
	type indexAndAccuracy struct {
		index    int
		accuracy float64
	}
	var values [5]indexAndAccuracy
	for i := range 5 {
		values[i] = indexAndAccuracy{i, ws.evaluateScore(probes[i])}
	}
	slices.SortFunc(values[:], func(a, b indexAndAccuracy) int { return cmp.Compare(b.accuracy, a.accuracy) })

	// determine next step based on ranking of probes
	if ws.marginalGap(values[0].accuracy, values[1].accuracy) {
		// minimal gap between probes, perhaps nothing more to explore
	} else if ws.largeGap(values[0].accuracy, values[1].accuracy) {
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
		ws.addSearchPlan(rangeMin, rangeMax)
	} else if oneOfEqualsQuery(values[0].index, values[1].index, 0) {
		// the two best includes the middle
		other := values[0].index
		if other == 0 {
			other = values[1].index
		}

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
		ws.addSearchPlan(rangeMin, rangeMax)
	} else if pairEqualsQueryPair(values[0].index, values[1].index, 1, 3) {
		ws.addSearchPlan(
			bound.rangeMin,
			sliceMixEverySecond(probes[0], bound.rangeMax),
		)
	} else if pairEqualsQueryPair(values[0].index, values[1].index, 1, 4) {
		ws.addSearchPlan(
			bound.rangeMin,
			sliceMixEverySecond(bound.rangeMax, probes[0]),
		)
	} else if pairEqualsQueryPair(values[0].index, values[1].index, 2, 3) {
		ws.addSearchPlan(
			sliceMixEverySecond(bound.rangeMin, probes[0]),
			bound.rangeMax,
		)
	} else if pairEqualsQueryPair(values[0].index, values[1].index, 2, 4) {
		ws.addSearchPlan(
			sliceMixEverySecond(probes[0], bound.rangeMin),
			bound.rangeMax,
		)
	} else {
		// remaining two are not adjacent, queue them separately
		for check := range 2 {
			var rangeMin, rangeMax []float64
			switch values[check].index {
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
			ws.addSearchPlan(rangeMin, rangeMax)
		}
	}
}

func (ws *WeightSearcher2) addSearchPlan(rangeMin []float64, rangeMax []float64) {
	ws.queue.Push(&weightSearch2Bound{
		plannedOp: opSearch,
		rangeMin:  rangeMin,
		rangeMax:  rangeMax,
	})
}

func oneOfEqualsQuery(a, b int, query int) bool {
	return a == query || b == query
}

func pairEqualsQueryPair(a, b int, query1, query2 int) bool {
	return (a == query1 && b == query2) || (a == query2 && b == query1)
}

func (ws *WeightSearcher2) largeGap(a, b float64) bool {
	return math.Abs(a-b) < c_search2_largeGap
}

func (ws *WeightSearcher2) marginalGap(a, b float64) bool {
	return math.Abs(a-b) < c_search2_marginalGap
}

func sliceInterpolate(rangeMin []float64, rangeMax []float64, ratio float64) []float64 {
	result := make([]float64, len(rangeMin))
	for i := range len(rangeMin) {
		result[i] = rangeMin[i] + (rangeMax[i]-rangeMin[i])*ratio
		//result[i] = (1-ratio)*rangeMin[i] + rangeMax[i]*ratio
	}
	return result
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
