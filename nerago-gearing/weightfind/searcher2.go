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

	c_search2_minRunEarlySizeCut = 4
	c_search2_minRunLateSizeCut  = 2

	c_search2_probeA       = 0.25
	c_search2_probe_middle = 0.5
	c_search2_probeB       = 0.75

	c_max_node_depth = 100
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
	nodeDepth int
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
		nodeDepth: 0,
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
			ws.opSearch2(bound)
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
	mid := valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_search2_probe_middle)

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
		nodeDepth: bound.nodeDepth + 1,
	})
	ws.queue.Push(&weightSearch2Bound{
		plannedOp: nextOp,
		axisFocus: nextAxis,
		rangeMin:  copyAndReplaceElement(bound.rangeMin, axis, mid),
		rangeMax:  bound.rangeMax,
		nodeDepth: bound.nodeDepth + 1,
	})
}

func (ws *WeightSearcher2) opFinal(bound *weightSearch2Bound) {
	middle := sliceInterpolate(bound.rangeMin, bound.rangeMax, c_search2_probe_middle)
	ws.evaluateScore(middle)
	ws.printer.Println("opFinal STOP")
	// TODO consider testing extremes
}

type probeAndAccuracy struct {
	accuracy float64
	axis     int
	isHigh   bool
	point    []float64
}

func (ws *WeightSearcher2) opSearch2(bound *weightSearch2Bound) {
	middle := sliceInterpolate(bound.rangeMin, bound.rangeMax, c_search2_probe_middle)
	probes := ws.search2DoProbes(bound, middle)

	if ws.marginalAccuracyGap(probes[0].accuracy, probes[len(probes)-1].accuracy) {
		ws.printer.Println("  -> MARGINAL GAP STOP")
		return
	}

	// TODO consider sub-cut into groups by accuracy, separate to hi/lo divides

	cutPoint := ws.search2ChooseCut(probes)
	probes = probes[0 : cutPoint+1]
	ws.search2ChooseSplitMode(probes, bound, middle)
}

func (ws *WeightSearcher2) search2DoProbes(bound *weightSearch2Bound, middle []float64) []probeAndAccuracy {
	probes := make([]probeAndAccuracy, 0)
	probes[0] = probeAndAccuracy{point: middle, axis: -1, accuracy: ws.evaluateScore(middle)}
	for axis := range ws.typeCount {
		lo := slices.Clone(middle)
		lo[axis] = valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_search2_probeA)
		probes = append(probes, probeAndAccuracy{point: lo, axis: axis, isHigh: false, accuracy: ws.evaluateScore(lo)})

		hi := slices.Clone(middle)
		hi[axis] = valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_search2_probeB)
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
	hiSlice := make([]probeAndAccuracy, 0, len(probes))
	loSlice := make([]probeAndAccuracy, 0, len(probes))
	for i := range probes {
		entry := &probes[i]
		if entry.axis == -1 {
			includeMiddle = true
		} else if entry.isHigh {
			hiSlice = append(hiSlice, *entry)
		} else {
			loSlice = append(loSlice, *entry)
		}
	}
	hi, lo := len(hiSlice), len(loSlice)

	// TODO no reason why we couldn't in theory break up a big one sided set

	if hi == 0 && lo == 0 {
		// no hi or lo, probably just middle, will just apply basic narrowing
		ws.search2MakeFollowupShrink(bound)
	} else if hi == 0 || lo == 0 {
		// everything on once side, with or without the middle
		ws.search2MakeFollowupCommon(probes, bound, includeMiddle, middle)
	} else if hi == 1 && lo == 1 && hiSlice[0].axis == loSlice[0].axis {
		// basically just know its around middle
		// go for general shrinking but keep that full range since we know its of interest
		ws.search2MakeFollowupShrinkExceptAxis(bound, hiSlice[0].axis)
	} else if hi == 1 && lo == 1 {
		// go ahead and extend a bit in hi and low directions on different axes
		ws.search2MakeFollowupCommon(probes, bound, includeMiddle, middle)
	} else if (hi == 1 && lo > 1) || (lo == 1 && hi > 1) {
		// almost all on one side except for one
		// could have an axis repeated, but common should be happy enough with one good value
		ws.search2MakeFollowupCommon(probes, bound, includeMiddle, middle)
	} else if !includeMiddle {
		// make 2 regions on either side of middle
		ws.search2MakeFollowupCommon(loSlice, bound, false, middle)
		ws.search2MakeFollowupCommon(hiSlice, bound, false, middle)
	} else {
		// these will have a fair bit of overlap, but so be it
		ws.search2MakeFollowupCommon(loSlice, bound, false, middle)
		ws.search2MakeFollowupCommon(hiSlice, bound, false, middle)
		ws.search2MakeFollowupShrink(bound)
	}

	// (hi == 1 && lo >= 1) || (lo == 1 && hi >= 1)
	//
	//   !((hi == 1 && lo >= 1) || (lo == 1 && hi >= 1))
	// =  ((hi != 1 || lo < 1) && (lo != 1 || hi < 1))
	// =  ((hi > 1) && (lo > 1))

	//return hiSlice, loSlice
}

func (ws *WeightSearcher2) search2MakeFollowupShrink(bound *weightSearch2Bound) {
	rangeMin := sliceInterpolate(bound.rangeMin, bound.rangeMax, c_search2_probeA)
	rangeMax := sliceInterpolate(bound.rangeMin, bound.rangeMax, c_search2_probeB)
	ws.addSearchPlan(rangeMin, rangeMax, bound)
}

func (ws *WeightSearcher2) search2MakeFollowupShrinkExceptAxis(bound *weightSearch2Bound, axis int) {
	rangeMin := sliceInterpolate(bound.rangeMin, bound.rangeMax, c_search2_probeA)
	rangeMin[axis] = bound.rangeMin[axis]
	rangeMax := sliceInterpolate(bound.rangeMin, bound.rangeMax, c_search2_probeB)
	rangeMax[axis] = bound.rangeMax[axis]
	ws.addSearchPlan(rangeMin, rangeMax, bound)
}

func (ws *WeightSearcher2) search2MakeFollowupCommon(probes []probeAndAccuracy, bound *weightSearch2Bound, includeMiddle bool, middle []float64) {
	madeChangeOfSignificance := false
	rangeMin := slices.Clone(bound.rangeMin)
	rangeMax := slices.Clone(bound.rangeMax)
	for axis := range ws.typeCount {
		hasHi, hasLo := includesAxisEntry(probes, axis)
		if hasHi && !hasLo {
			if includeMiddle {
				rangeMin[axis] = valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_search2_probeA)
			} else {
				rangeMin[axis] = middle[axis]
			}
			madeChangeOfSignificance = true
		} else if !hasHi && hasLo {
			if includeMiddle {
				rangeMax[axis] = valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_search2_probeB)
			} else {
				rangeMax[axis] = middle[axis]
			}
			madeChangeOfSignificance = true
		} else if !hasHi /* && !hasLo */ {
			// NOTE considering that hi&&lo means keep full range, but !hi&&!lo could narrow
			rangeMin[axis] = valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_search2_probeA)
			rangeMax[axis] = valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_search2_probeB)
		}
	}

	if !madeChangeOfSignificance {
		panic("logic fail, no significant change found")
	}

	ws.addSearchPlan(rangeMin, rangeMax, bound)
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

//	var interest = []float64{
//		5.3900000000,
//		7.4731250000,
//		2.6908057851,
//		7.4200000000,
//		-0.5654897163,
//		8.0470000000,
//		4.8125000000,
//		3.2600000000,
//	}
//
//	func debugValueOfInterest(rangeMin []float64, rangeMax []float64, printer *util.PrintRecorder) {
//		if checkInRange(rangeMin, rangeMax, interest) {
//			printer.Println("INTEREST")
//		}
//	}
//
//	func checkInRange(rangeMin []float64, rangeMax []float64, probe []float64) bool {
//		for i := range probe {
//			if rangeMin[i] <= probe[i] && probe[i] <= rangeMax[i] {
//				// ok
//			} else {
//				return false
//			}
//		}
//		return true
//	}
//
//	func debugVerifyInRange(rangeMin []float64, rangeMax []float64, probe []float64, printer *util.PrintRecorder) {
//		if !checkInRange(rangeMin, rangeMax, probe) {
//			panic("probe value isn't inside remaining range")
//		}
//
//		printRange(rangeMin, rangeMax, "     = ", printer)
//	}
//
//	func printRange(rangeMin []float64, rangeMax []float64, label string, printer *util.PrintRecorder) {
//		printer.Printf(label)
//		for i := range rangeMin {
//			printer.Printf("%7.4f ", rangeMin[i])
//		}
//		printer.Println0()
//		printer.Printf(label)
//		for i := range rangeMax {
//			printer.Printf("%7.4f ", rangeMax[i])
//		}
//		printer.Println0()
//	}
func (ws *WeightSearcher2) addSearchPlan(rangeMin []float64, rangeMax []float64, bound *weightSearch2Bound) {
	if ws.rangeIsMarginal(rangeMin, rangeMax) || bound.nodeDepth >= c_max_node_depth {
		ws.queue.Push(&weightSearch2Bound{
			plannedOp: opFinal,
			rangeMin:  rangeMin,
			rangeMax:  rangeMax,
			nodeDepth: bound.nodeDepth + 1,
		})
	} else {
		ws.queue.Push(&weightSearch2Bound{
			plannedOp: opSearch,
			rangeMin:  rangeMin,
			rangeMax:  rangeMax,
			nodeDepth: bound.nodeDepth + 1,
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

//	func oneOfEqualsQuery(a, b int, query int) bool {
//		return a == query || b == query
//	}
//
//	func pairEqualsQueryPair(a, b int, query1, query2 int) bool {
//		return (a == query1 && b == query2) || (a == query2 && b == query1)
//	}
func (ws *WeightSearcher2) largeAccuracyGap(a, b float64) bool {
	return math.Abs(a-b) >= c_search2_largeAccuracyGap
}

//	func (ws *WeightSearcher2) firstGreaterThanEqualPair(values *[5]indexAndAccuracy) bool {
//		return !util.FloatsApproxEquals(values[0].accuracy, values[1].accuracy) &&
//			util.FloatsApproxEquals(values[1].accuracy, values[2].accuracy)
//	}
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

//	func sliceMixEverySecond(a, b []float64) []float64 {
//		result := make([]float64, len(a))
//		for i := range len(a) {
//			if i&1 == 0 {
//				result[i] = a[i]
//			} else {
//				result[i] = b[i]
//			}
//		}
//		return result
//	}
func copyAndReplaceElement(slice []float64, index int, value float64) []float64 {
	newSlice := slices.Clone(slice)
	newSlice[index] = value
	return newSlice
}
