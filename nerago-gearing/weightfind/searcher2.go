package weightfind

import (
	"cmp"
	"paladin_gearing_go/solver/stathighs"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_rank"
	"slices"
)

const ()

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
	0 > 1234        14 > 0124       13 > 0124       02 > 134

.............    .............   ......|......   ......|......   ..|..........
..1.......2..    ..?-------?..   ..1...|......   ..1...|......   ..?.......2..
.............    ..|.......|..   ......|......   ......|......   ..|..........
......0......    ..|...0...|..   ------?------   ......?......   ..|...0......
.............    ..|.......|..   ......|......   ......|......   ..|..........
..3.......4..    ..?-------?..   ......|...4..   ..3...|......   ..?-------?--
.............    .............   ......|......   ......|......   .............
*/
func (ws *WeightSearcher2) opSearch(bound *weightSearch2Bound) {
	// selected points throughout the sample space. could have more variety of axis points.
	var probes [5][]float64
	probes[0] = sliceInterpolate(bound.rangeMin, bound.rangeMax, 1.0/2.0)
	probes[1] = sliceInterpolate(bound.rangeMin, bound.rangeMax, 1.0/4.0)
	probes[4] = sliceInterpolate(bound.rangeMin, bound.rangeMax, 3.0/4.0)
	probes[2] = sliceMixEverySecond(probes[1], probes[4])
	probes[3] = sliceMixEverySecond(probes[4], probes[1])

	// evaluate each probe
	type indexAndAccuracy struct {
		index    int
		accuracy float64
	}
	var values [5]indexAndAccuracy
	for i := range 5 {
		values[i] = indexAndAccuracy{i, ws.evaluateScore(probes[i])}
	}

	// sort descending so best are first
	slices.SortFunc(values[:], func(a, b indexAndAccuracy) int { return cmp.Compare(b.accuracy, a.accuracy) })

	//if ws.largeGap(values[0].accuracy, values[1].accuracy) {
	//
	//}
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
