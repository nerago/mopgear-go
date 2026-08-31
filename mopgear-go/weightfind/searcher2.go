package weightfind

import (
	"cmp"
	"math"
	"slices"
	"strconv"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_rank"
	"github.com/nerago/mopgear-go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_search2_abandonBranchAccuracyGap = 0.4
	c_search2_largeAccuracyGap         = 0.5
	c_search2_equalAccuracyGap         = 0.01
	c_search2_goalAccuracyGap          = 0.001
	c_search2_marginalWeightGap        = 0.0001

	c_search2_minRunEarlySizeCut = 4
	c_search2_minRunLateSizeCut  = 2

	c_search2_probeA      = 0.1
	c_search2_probeMiddle = 0.5
	c_search2_probeB      = 0.9

	c_search2_maxNodeDepth = 30
	c_search2_useFinalOp   = false
	c_search2_debug        = false
)

type opType int8

const (
	opDivide1 opType = iota
	opSearch  opType = iota
	opFinal   opType = iota
)

type WeightSearcher2 struct {
	typeCount             int
	statTypes             []stats.StatType
	simTypes              []stats.SimType
	targetRatio           weight_types.SimPriorityBasic
	evaluateAccuracy      EvaluateAccuracyPrepared
	printer               *util.PrintRecorder
	AccuracyStatistical   bool
	AccuracyStatisticalEx bool

	queue      util_collection.QueueStackFilo[*weightSearch2Bound]
	bestResult util_rank.BestCollector1[weight_types.Weight1Basic]
}

type weightSearch2Bound struct {
	plannedOp opType
	axisFocus int
	rangeMin  []float64
	rangeMax  []float64
	nodeDepth int
}

func (ws *WeightSearcher2) Init(statTypes []stats.StatType, targetRatio weight_types.SimPriorityBasic, printer *util.PrintRecorder) {
	ws.typeCount = len(statTypes)
	ws.statTypes = statTypes
	ws.simTypes = targetRatio.SimTypes()
	ws.targetRatio = targetRatio
	ws.printer = printer
}

func (ws *WeightSearcher2) SupplyData(inputData []weight_types.WeightInput) {
	ws.evaluateAccuracy.Init(inputData, &ws.targetRatio, ws.AccuracyStatistical, ws.AccuracyStatisticalEx)
}

func (ws *WeightSearcher2) SetRanges(weightMin, weightMax float64) {
	ws.queue.Push(&weightSearch2Bound{
		plannedOp: opDivide1,
		axisFocus: 0,
		rangeMin:  util_collection.RepeatValue(weightMin, ws.typeCount),
		rangeMax:  util_collection.RepeatValue(weightMax, ws.typeCount),
		nodeDepth: 0,
	})
}

func (ws *WeightSearcher2) Run(cancel util_async.CancelSignal) weight_types.WeightResult1 {
	stopwatch := util.StopwatchMakeStarted()
	iterCount := 0
	for cancel.ShouldContinue() {
		bound, hasValue := ws.queue.Pop()
		if !hasValue {
			break
		}

		//if !checkInRange(bound.rangeMin, bound.rangeMax, interest) {
		//	continue
		//}

		if c_search2_debug {
			printRange(bound.rangeMin, bound.rangeMax, "STEP "+strconv.FormatInt(int64(bound.plannedOp), 10)+" ", ws.printer)
		}

		switch bound.plannedOp {
		case opDivide1:
			ws.opDivide1(bound)
		case opSearch:
			ws.opSearch2(bound)
		case opFinal:
			ws.opFinal(bound)
		}

		iterCount++
		if (c_search2_debug || iterCount%1000 == 0) && ws.printer != nil {
			ws.printer.Printf("search i=%d q=%d b=%f\n", iterCount, ws.queue.Size(), ws.bestResult.GetBestScore())
		}
		if iterCount%10 == 0 {
			ws.queueMaintenance()
		}
	}

	bestWeight := ws.bestResult.GetBestOrNilValue()
	bestWeight.NormalizeForBase(ws.statTypes)

	status := highs.ModelStatusOptimal
	if cancel.ShouldFinish() {
		status = highs.ModelStatusTimeLimit
	}

	return weight_types.WeightResult1Make(&bestWeight, stopwatch.Elapsed(), status)
}

func (ws *WeightSearcher2) evaluateScore(weightArray []float64) float64 {
	weights := weight_types.Weight1Basic_Of(weightArray, ws.statTypes)

	weights.NormalizeForBase(ws.statTypes)
	if weights.IsOverlySimple() {
		return 0
	}

	accuracy := ws.evaluateAccuracy.EvaluateWeight1(&weights)
	ws.bestResult.Offer(&weights, accuracy)
	return accuracy
}

// dumb division in halves
func (ws *WeightSearcher2) opDivide1(bound *weightSearch2Bound) {
	axis := bound.axisFocus
	mid := valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_search2_probeMiddle)

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
	middle := sliceInterpolate(bound.rangeMin, bound.rangeMax, c_search2_probeMiddle)
	ws.evaluateScore(middle)
	if c_search2_debug {
		ws.printer.Println("opFinal STOP")
	}
	// TODO consider testing extremes
}

type probeAndAccuracy struct {
	accuracy float64
	axis     int
	isHigh   bool
	point    []float64
}

func (ws *WeightSearcher2) opSearch2(bound *weightSearch2Bound) {
	middle := sliceInterpolate(bound.rangeMin, bound.rangeMax, c_search2_probeMiddle)
	probes := ws.search2DoProbes(bound, middle)

	if probes[0].accuracy < ws.bestResult.GetBestScore()-c_search2_abandonBranchAccuracyGap {
		if c_search2_debug {
			ws.printer.Println("  -> ABANDON BRANCH")
		}
		return
	} else if c_search2_debug {
		ws.printer.Printf(" BRANCH %f %f %f\n", probes[0].accuracy, ws.bestResult.GetBestScore(), ws.bestResult.GetBestScore()-c_search2_abandonBranchAccuracyGap)
	}

	gapFirstProbeToLast := probes[0].accuracy - probes[len(probes)-1].accuracy
	if gapFirstProbeToLast <= c_search2_goalAccuracyGap {
		if c_search2_debug {
			ws.printer.Printf("  -> MARGINAL GAP STOP %f %f\n", probes[0].accuracy, probes[len(probes)-1].accuracy)
		}
		return
	}

	// TODO consider sub-cut into groups by accuracy, separate to hi/lo divides

	cutPoint := ws.search2ChooseCut(probes)
	probes = probes[0 : cutPoint+1]
	ws.search2ChooseSplitMode(probes, bound)
}

func (ws *WeightSearcher2) search2DoProbes(bound *weightSearch2Bound, middle []float64) []probeAndAccuracy {
	probes := make([]probeAndAccuracy, 1, ws.typeCount*2+1)
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
		if ws.equalAccuracyGap(probes[runStartIndex].accuracy, probes[runStartIndex+1].accuracy) {
			runSize := 2
			for check := runStartIndex + 1; check < len(probes)-1; check++ {
				if ws.equalAccuracyGap(probes[check].accuracy, probes[check+1].accuracy) {
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

func (ws *WeightSearcher2) search2ChooseSplitMode(probes []probeAndAccuracy, bound *weightSearch2Bound) {
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
		ws.search2MakeFollowupCommon(probes, bound)
	} else if hi == 1 && lo == 1 && hiSlice[0].axis == loSlice[0].axis {
		// basically just know its around middle
		// go for general shrinking but keep that full range since we know its of interest
		ws.search2MakeFollowupShrinkExceptAxis(bound, hiSlice[0].axis)
	} else if hi == 1 && lo == 1 {
		// go ahead and extend a bit in hi and low directions on different axes
		ws.search2MakeFollowupCommon(probes, bound)
	} else if (hi == 1 && lo > 1) || (lo == 1 && hi > 1) {
		// almost all on one side except for one
		// could have an axis repeated, but common should be happy enough with one good value
		ws.search2MakeFollowupCommon(probes, bound)
	} else if !includeMiddle {
		// make 2 regions on either side of middle
		ws.search2MakeFollowupCommon(loSlice, bound)
		ws.search2MakeFollowupCommon(hiSlice, bound)
	} else {
		// these will have a fair bit of overlap, but so be it
		ws.search2MakeFollowupCommon(loSlice, bound)
		ws.search2MakeFollowupCommon(hiSlice, bound)
		ws.search2MakeFollowupShrink(bound)
	}
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

func (ws *WeightSearcher2) search2MakeFollowupCommon(probes []probeAndAccuracy, bound *weightSearch2Bound) {
	rangeMin := slices.Clone(probes[0].point)
	rangeMax := slices.Clone(probes[0].point)
	for i := 1; i < len(probes); i++ {
		point := probes[i].point
		for axis := range len(rangeMin) {
			rangeMin[axis] = min(rangeMin[axis], point[axis])
			rangeMax[axis] = max(rangeMax[axis], point[axis])
		}
	}

	for axis := range len(rangeMin) {
		if rangeMin[axis] == rangeMax[axis] {
			rangeMin[axis] = bound.rangeMin[axis]
			rangeMax[axis] = bound.rangeMax[axis]
		} else {
			oldExtent := bound.rangeMax[axis] - bound.rangeMin[axis]
			rangeMin[axis] -= c_search2_probeA * oldExtent
			rangeMax[axis] += c_search2_probeA * oldExtent
		}
	}

	if !slices.Equal(bound.rangeMin, rangeMin) || !slices.Equal(bound.rangeMax, rangeMax) {
		ws.addSearchPlan(rangeMin, rangeMax, bound)
	} else {
		//ws.queue.Push(&weightSearch2Bound{
		//	plannedOp: opDivide1,
		//	axisFocus: 0,
		//	rangeMin:  rangeMin,
		//	rangeMax:  rangeMax,
		//	nodeDepth: bound.nodeDepth+1,
		//})
	}
}

func checkRangeIsSubrangeOf(outer, inner *weightSearch2Bound) bool {
	for i := range outer.rangeMin {
		if outer.rangeMin[i] <= inner.rangeMin[i] && inner.rangeMax[i] <= outer.rangeMax[i] {
			// yes
		} else {
			return false
		}
	}
	return true
}

func printRange(rangeMin []float64, rangeMax []float64, label string, printer *util.PrintRecorder) {
	printer.Printf("%s", label)
	for i := range rangeMin {
		printer.Printf("%7.4f ", rangeMin[i])
	}
	printer.Println0()
	printer.Printf("%s", label)
	for i := range rangeMax {
		printer.Printf("%7.4f ", rangeMax[i])
	}
	printer.Println0()
}

func (ws *WeightSearcher2) addSearchPlan(rangeMin []float64, rangeMax []float64, bound *weightSearch2Bound) {
	if ws.rangeIsMarginal(rangeMin, rangeMax) || bound.nodeDepth >= c_search2_maxNodeDepth {
		if c_search2_useFinalOp {
			ws.queue.Push(&weightSearch2Bound{
				plannedOp: opFinal,
				rangeMin:  rangeMin,
				rangeMax:  rangeMax,
				nodeDepth: bound.nodeDepth + 1,
			})
		}
		if c_search2_debug {
			if bound.nodeDepth >= c_search2_maxNodeDepth {
				ws.printer.Printf("DEPTH STOP %d\n", bound.nodeDepth)
			} else {
				ws.printer.Printf("MARGINAL RANGE STOP\n")
			}
		}
	} else {
		add := &weightSearch2Bound{
			plannedOp: opSearch,
			rangeMin:  rangeMin,
			rangeMax:  rangeMax,
			nodeDepth: bound.nodeDepth + 1,
		}
		ws.queue.Push(add)
	}

	if c_search2_debug {
		printRange(rangeMin, rangeMax, " = ", ws.printer)
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

func (ws *WeightSearcher2) largeAccuracyGap(a, b float64) bool {
	return math.Abs(a-b) >= c_search2_largeAccuracyGap
}

func (ws *WeightSearcher2) equalAccuracyGap(a, b float64) bool {
	return math.Abs(a-b) < c_search2_equalAccuracyGap
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
}

func copyAndReplaceElement(slice []float64, index int, value float64) []float64 {
	newSlice := slices.Clone(slice)
	newSlice[index] = value
	return newSlice
}

func (ws *WeightSearcher2) queueMaintenance() {
	ws.queue.UpdateContents(func(content []*weightSearch2Bound) []*weightSearch2Bound {
		removed := 0
		for a := range content {
			for b := a + 1; b < len(content); b++ {
				if checkRangeIsSubrangeOf(content[a], content[b]) {
					content = removeIndex(content, b)
					removed++
				} else if checkRangeIsSubrangeOf(content[b], content[a]) {
					content = removeIndex(content, a)
					removed++
				}
			}
		}
		if c_search2_debug {
			ws.printer.Printf("removed queue elements %d\n", removed)
		}
		return content
	})
}

func removeIndex(slice []*weightSearch2Bound, index int) []*weightSearch2Bound {
	return slices.Delete(slice, index, index+1)
}
