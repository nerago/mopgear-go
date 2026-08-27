package weightfind

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"sync"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_rank"
	"github.com/nerago/mopgear-go/weightfind/weight_types"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_searchRatio1_abandonBranchAccuracyGap = 0.10
	c_searchRatio1_largeAccuracyGap         = 0.1
	c_searchRatio1_equalAccuracyGap         = 0.01

	c_searchRatio1_minRunEarlySizeCut = 5
	c_searchRatio1_minRunLateSizeCut  = 3

	c_searchRatio1_probeMiddle = 0.5

	c_searchRatio1_goalAccuracyGap   = 0.01
	c_searchRatio1_marginalWeightGap = 0.01
	c_searchRatio1_probeA            = 0.2
	c_searchRatio1_probeB            = 0.8
	c_searchRatio1_maxNodeDepth      = 400

	c_searchRatio1_maxStats = 8
	c_searchRatio1_maxRatio = 6
	c_searchRatio1_maxPoint = c_searchRatio1_maxStats + c_searchRatio1_maxRatio
)

type WeightSearcherRatio1 struct {
	typeCount           int
	statTypes           []stats.StatType
	simTypes            []stats.SimType
	inputData           []weight_types.WeightInput
	initialBound        *weightSearchRatio1Bound
	AccuracyStatistical bool

	bestResult util_rank.BestCollector1Concurrent[weightSearchRatio1Best]

	poolQueue util.TypedPool[weightSearchRatio1Bound]
}

type weightSearchRatio1Point [c_searchRatio1_maxPoint]float64

type weightSearchRatio1Bound struct {
	rangeMin     weightSearchRatio1Point
	rangeMax     weightSearchRatio1Point
	divideAxis   int
	nodeDepth    int
	bestAccuracy float64
}

type weightSearchRatio1Probe struct {
	accuracy float64
	axis     int
	isHigh   bool
	point    weightSearchRatio1Point
}

type weightSearchRatio1Best struct {
	weight weight_types.Weight1Basic
	ratio  weight_types.SimPriorityBasic
}

func (ws *WeightSearcherRatio1) Init(statTypes []stats.StatType, simTypes []stats.SimType) {
	ws.statTypes = statTypes
	ws.simTypes = simTypes
	ws.typeCount = len(statTypes) + len(simTypes)
	if ws.typeCount > c_searchRatio1_maxPoint {
		panic("don't support that many stats/sims")
	}
}

func (ws *WeightSearcherRatio1) SupplyData(inputData []weight_types.WeightInput) {
	ws.inputData = inputData
}

func (ws *WeightSearcherRatio1) SetStatSimRanges(statRange weight_types.StatRangeFloat, simRanges stats.SimTypeMap[weight_types.StatRangeFloat]) {
	bound := ws.poolQueue.Get()
	bound.divideAxis = 0
	bound.nodeDepth = 0
	bound.bestAccuracy = 0 // don't really need proper value

	index := 0
	for range ws.statTypes {
		bound.rangeMin[index] = statRange.Minimum
		bound.rangeMax[index] = statRange.Maximum
		index++
	}
	for _, simType := range ws.simTypes {
		nextRange := simRanges.GetOrPanic(simType)
		bound.rangeMin[index] = nextRange.Minimum
		bound.rangeMax[index] = nextRange.Maximum
		index++
	}

	ws.initialBound = bound
}

func (ws *WeightSearcherRatio1) Run(cancel util_async.CancelSignal) weight_types.WeightResult1 {
	stopwatch := util.StopwatchMakeStarted()
	threadCount := 12
	queue := &util_collection.QueueStackFiloConcurrent[*weightSearchRatio1Bound]{}

	startingProbesReused := ws.newProbeSlice()
	if ws.initialSplits(queue, startingProbesReused, threadCount) {
		waitGroup := sync.WaitGroup{}
		for range threadCount - 1 {
			waitGroup.Go(func() {
				ws.threadLoop(cancel, queue, ws.newProbeSlice(), false)
			})
		}

		ws.threadLoop(cancel, queue, startingProbesReused, true)
		waitGroup.Wait()
	}

	bestItems := ws.bestResult.GetBestOrNilValue()
	bestWeight := bestItems.weight
	return weight_types.WeightResult1MakeWithRatio(
		&bestWeight,
		stopwatch.Elapsed(),
		highs.ModelStatusOptimal,
		new(bestItems.ratio),
	)
}

func (ws *WeightSearcherRatio1) initialSplits(queue *util_collection.QueueStackFiloConcurrent[*weightSearchRatio1Bound], probesReused []*weightSearchRatio1Probe, targetCount int) bool {
	queue.Push(ws.initialBound)

	localQueue := util_collection.QueueStackFiloConcurrentCachedMake(queue)
	defer localQueue.Flush()

	for queue.Size() < targetCount {
		if !ws.threadStep(localQueue, probesReused) {
			return false
		}
	}
	return true
}

func (ws *WeightSearcherRatio1) threadLoop(cancel util_async.CancelSignal, queue *util_collection.QueueStackFiloConcurrent[*weightSearchRatio1Bound], probesReused []*weightSearchRatio1Probe, maintainer bool) {
	iterCount := 0

	localQueue := util_collection.QueueStackFiloConcurrentCachedMake(queue)
	defer localQueue.Flush()

	for cancel.ShouldContinue() {
		if !ws.threadStep(localQueue, probesReused) {
			break
		}

		if iterCount%1000 == 0 {
			fmt.Printf("search-ratio i=%d q=%d b=%f\n", iterCount, localQueue.SizeVolatile(), ws.bestResult.GetBestValue())
			if maintainer {
				ws.maintainQueue(queue, iterCount%2000 == 0)
			}
		}
		iterCount++
	}
}

func (ws *WeightSearcherRatio1) threadStep(localQueue *util_collection.QueueStackFiloConcurrentCached[*weightSearchRatio1Bound], probesReused []*weightSearchRatio1Probe) bool {
	entry, hasEntry := localQueue.Pop()
	if !hasEntry {
		return false
	}

	if entry.divideAxis != -1 {
		ws.opDivide(entry, localQueue)
	} else {
		ws.opSearch(entry, probesReused, localQueue)
	}

	ws.poolQueue.Put(entry)
	return true
}

func (ws *WeightSearcherRatio1) newProbeSlice() []*weightSearchRatio1Probe {
	slice := make([]*weightSearchRatio1Probe, ws.typeCount*2+1)
	for i := range slice {
		slice[i] = new(weightSearchRatio1Probe)
	}
	return slice
}

func (ws *WeightSearcherRatio1) evaluateScore(weightArray *weightSearchRatio1Point) float64 {
	index := 0

	weights := weight_types.Weight1Basic{}
	for _, statType := range ws.statTypes {
		weights.Put(statType, weightArray[index])
		index++
	}

	weights.NormalizeForBase(ws.statTypes)
	if weights.IsOverlySimple() {
		return 0
	}

	ratio := &weight_types.SimPriorityBasic{}
	for _, simType := range ws.simTypes {
		ratio.Set(simType, weightArray[index])
		index++
	}
	ratio.ScaleForTotalSum(1.0)

	accuracy := EvaluateAccuracySwitch(ws.AccuracyStatistical, &weights, ws.simTypes, ratio, ws.inputData)
	ws.bestResult.Offer(&weightSearchRatio1Best{
		weights, *ratio,
	}, accuracy)
	return accuracy
}

// dumb division in halves
func (ws *WeightSearcherRatio1) opDivide(bound *weightSearchRatio1Bound, localQueue *util_collection.QueueStackFiloConcurrentCached[*weightSearchRatio1Bound]) {
	axis := bound.divideAxis
	mid := valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_searchRatio1_probeMiddle)

	var nextAxis int
	if axis < ws.typeCount-1 {
		nextAxis = axis + 1
	} else {
		nextAxis = -1
	}

	hiBound := ws.poolQueue.Get()
	hiBound.divideAxis = nextAxis
	hiBound.nodeDepth = bound.nodeDepth + 1
	hiBound.rangeMin = bound.rangeMin
	hiBound.rangeMax = bound.rangeMax
	hiBound.rangeMax[axis] = mid
	hiBound.bestAccuracy = max(ws.evaluateScore(&hiBound.rangeMin), ws.evaluateScore(&hiBound.rangeMax))
	localQueue.Push(hiBound)

	loBound := ws.poolQueue.Get()
	loBound.divideAxis = nextAxis
	loBound.nodeDepth = bound.nodeDepth + 1
	loBound.rangeMin = bound.rangeMin
	loBound.rangeMin[axis] = mid
	loBound.rangeMax = bound.rangeMax
	loBound.bestAccuracy = max(ws.evaluateScore(&loBound.rangeMin), ws.evaluateScore(&loBound.rangeMax))
	localQueue.Push(loBound)
}

func (ws *WeightSearcherRatio1) opSearch(bound *weightSearchRatio1Bound, probes []*weightSearchRatio1Probe, localQueue *util_collection.QueueStackFiloConcurrentCached[*weightSearchRatio1Bound]) {
	probes = ws.createAndSetProbes(bound, probes)

	if bound.nodeDepth >= c_searchRatio1_maxNodeDepth {
		// done
	} else if probes[0].accuracy < ws.bestResult.GetBestValue()-c_searchRatio1_abandonBranchAccuracyGap {
		// done
	} else if gapFirstProbeToLast := probes[0].accuracy - probes[len(probes)-1].accuracy; gapFirstProbeToLast <= c_searchRatio1_goalAccuracyGap {
		// done
	} else {
		cutPoint := ws.chooseCut(probes)
		probesAfterCut := probes[0 : cutPoint+1]
		ws.chooseSplitMode(probesAfterCut, bound, localQueue)
	}
}

func (ws *WeightSearcherRatio1) createAndSetProbes(bound *weightSearchRatio1Bound, probes []*weightSearchRatio1Probe) []*weightSearchRatio1Probe {
	middle := probes[0]
	ws.sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_searchRatio1_probeMiddle, &middle.point)
	middle.accuracy = ws.evaluateScore(&middle.point)
	middle.axis = -1

	index := 1
	for axis := range ws.typeCount {
		lo := probes[index]
		lo.point = middle.point
		lo.point[axis] = valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_searchRatio1_probeA)
		lo.axis = axis
		lo.isHigh = false
		lo.accuracy = ws.evaluateScore(&lo.point)
		index++

		hi := probes[index]
		hi.point = middle.point
		hi.point[axis] = valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_searchRatio1_probeB)
		hi.axis = axis
		hi.isHigh = true
		hi.accuracy = ws.evaluateScore(&hi.point)
		index++
	}

	slices.SortStableFunc(probes, func(a, b *weightSearchRatio1Probe) int { return cmp.Compare(b.accuracy, a.accuracy) })
	return probes
}

// returns last index to include
func (ws *WeightSearcherRatio1) chooseCut(probes []*weightSearchRatio1Probe) int {
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
			if runStartIndex <= 1 && runSize >= c_searchRatio1_minRunEarlySizeCut && runSize < len(probes) {
				return runStartIndex + runSize - 1
			} else if runStartIndex > 1 && runSize >= c_searchRatio1_minRunLateSizeCut && runSize < len(probes) {
				return runStartIndex + runSize - 1
			}
		}
	}

	// default, cut at one quarter of range
	return len(probes) / 4
}

func (ws *WeightSearcherRatio1) chooseSplitMode(probes []*weightSearchRatio1Probe, bound *weightSearchRatio1Bound, localQueue *util_collection.QueueStackFiloConcurrentCached[*weightSearchRatio1Bound]) {
	bestAccuracy := probes[0].accuracy

	includeMiddle := false
	hiSlice := make([]*weightSearchRatio1Probe, 0, len(probes))
	loSlice := make([]*weightSearchRatio1Probe, 0, len(probes))
	for i := range probes {
		entry := probes[i]
		if entry.axis == -1 {
			includeMiddle = true
		} else if entry.isHigh {
			hiSlice = append(hiSlice, entry)
		} else {
			loSlice = append(loSlice, entry)
		}
	}
	hi, lo := len(hiSlice), len(loSlice)

	if hi == 0 && lo == 0 {
		// no hi or lo, probably just middle, will just apply basic narrowing
		ws.search2MakeFollowupShrink(bound, localQueue, bestAccuracy)
	} else if hi == 0 || lo == 0 {
		// everything on once side, with or without the middle
		ws.search2MakeFollowupCommon2(probes, bound, localQueue)
	} else if hi == 1 && lo == 1 && hiSlice[0].axis == loSlice[0].axis {
		// basically just know its around middle
		// go for general shrinking but keep that full range since we know its of interest
		ws.search2MakeFollowupShrinkExceptAxis(bound, hiSlice[0].axis, localQueue, bestAccuracy)
	} else if hi == 1 && lo == 1 {
		// go ahead and extend a bit in hi and low directions on different axes
		ws.search2MakeFollowupCommon2(probes, bound, localQueue)
	} else if (hi == 1 && lo > 1) || (lo == 1 && hi > 1) {
		// almost all on one side except for one
		// could have an axis repeated, but common should be happy enough with one good value
		ws.search2MakeFollowupCommon2(probes, bound, localQueue)
	} else if !includeMiddle {
		// make 2 regions on either side of middle
		ws.search2MakeFollowupCommon2(loSlice, bound, localQueue)
		ws.search2MakeFollowupCommon2(hiSlice, bound, localQueue)
	} else {
		// these will have a fair bit of overlap, but so be it
		ws.search2MakeFollowupCommon2(loSlice, bound, localQueue)
		ws.search2MakeFollowupCommon2(hiSlice, bound, localQueue)
		ws.search2MakeFollowupShrink(bound, localQueue, bestAccuracy)
	}
}

func (ws *WeightSearcherRatio1) search2MakeFollowupShrink(bound *weightSearchRatio1Bound, localQueue *util_collection.QueueStackFiloConcurrentCached[*weightSearchRatio1Bound], accuracy float64) {
	add := ws.poolQueue.Get()
	add.divideAxis = -1
	add.nodeDepth = bound.nodeDepth + 1
	add.bestAccuracy = accuracy

	ws.sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_searchRatio1_probeA, &add.rangeMin)
	ws.sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_searchRatio1_probeB, &add.rangeMax)

	if ws.rangeIsMarginal(&add.rangeMin, &add.rangeMax) || ws.containedByAnotherQueueEntry(add, localQueue) {
		ws.poolQueue.Put(add)
	} else {
		localQueue.Push(add)
	}
}

func (ws *WeightSearcherRatio1) search2MakeFollowupShrinkExceptAxis(bound *weightSearchRatio1Bound, axis int, localQueue *util_collection.QueueStackFiloConcurrentCached[*weightSearchRatio1Bound], accuracy float64) {
	add := ws.poolQueue.Get()
	add.divideAxis = -1
	add.nodeDepth = bound.nodeDepth + 1
	add.bestAccuracy = accuracy

	ws.sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_searchRatio1_probeA, &add.rangeMin)
	add.rangeMin[axis] = bound.rangeMin[axis]
	ws.sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_searchRatio1_probeB, &add.rangeMax)
	add.rangeMax[axis] = bound.rangeMax[axis]

	if ws.rangeIsMarginal(&add.rangeMin, &add.rangeMax) || ws.containedByAnotherQueueEntry(add, localQueue) {
		ws.poolQueue.Put(add)
	} else {
		localQueue.Push(add)
	}
}

func (ws *WeightSearcherRatio1) search2MakeFollowupCommon2(probes []*weightSearchRatio1Probe, bound *weightSearchRatio1Bound, localQueue *util_collection.QueueStackFiloConcurrentCached[*weightSearchRatio1Bound]) {
	add := ws.poolQueue.Get()
	add.divideAxis = -1
	add.nodeDepth = bound.nodeDepth + 1
	add.rangeMin = probes[0].point
	add.rangeMax = probes[0].point
	add.bestAccuracy = probes[0].accuracy

	rangeMin := &add.rangeMin
	rangeMax := &add.rangeMax
	for i := 1; i < len(probes); i++ {
		point := &probes[i].point
		for axis := range ws.typeCount {
			rangeMin[axis] = min(rangeMin[axis], point[axis])
			rangeMax[axis] = max(rangeMax[axis], point[axis])
		}
	}

	for axis := range ws.typeCount {
		if rangeMin[axis] == rangeMax[axis] {
			rangeMin[axis] = bound.rangeMin[axis]
			rangeMax[axis] = bound.rangeMax[axis]
		} else {
			oldExtent := bound.rangeMax[axis] - bound.rangeMin[axis]
			rangeMin[axis] -= c_searchRatio1_probeA * oldExtent
			rangeMax[axis] += c_searchRatio1_probeA * oldExtent
		}
	}

	if (bound.rangeMin == add.rangeMin && bound.rangeMax == add.rangeMin) || ws.rangeIsMarginal(&add.rangeMin, &add.rangeMax) || ws.containedByAnotherQueueEntry(add, localQueue) {
		ws.poolQueue.Put(add)
	} else {
		localQueue.Push(add)
	}
}

func (ws *WeightSearcherRatio1) checkRangeIsSubrangeOf(outer, inner *weightSearchRatio1Bound) bool {
	for i := range ws.typeCount {
		if outer.rangeMin[i] <= inner.rangeMin[i] && inner.rangeMax[i] <= outer.rangeMax[i] {
			// yes
		} else {
			return false
		}
	}
	return true
}

func (ws *WeightSearcherRatio1) rangeIsMarginal(rangeMin *weightSearchRatio1Point, rangeMax *weightSearchRatio1Point) bool {
	for i := range ws.typeCount {
		if rangeMax[i]-rangeMin[i] > c_searchRatio1_marginalWeightGap {
			return false
		}
	}
	return true
}

func (ws *WeightSearcherRatio1) largeAccuracyGap(a, b float64) bool {
	return math.Abs(a-b) >= c_searchRatio1_largeAccuracyGap
}

func (ws *WeightSearcherRatio1) equalAccuracyGap(a, b float64) bool {
	return math.Abs(a-b) < c_searchRatio1_equalAccuracyGap
}

func (ws *WeightSearcherRatio1) sliceInterpolate(rangeMin *weightSearchRatio1Point, rangeMax *weightSearchRatio1Point, ratio float64, out *weightSearchRatio1Point) {
	for i := range ws.typeCount {
		out[i] = valueInterpolate(rangeMin[i], rangeMax[i], ratio)
	}
}

func (ws *WeightSearcherRatio1) containedByAnotherQueueEntry(add *weightSearchRatio1Bound, localQueue *util_collection.QueueStackFiloConcurrentCached[*weightSearchRatio1Bound]) bool {
	foundContainingRange := false
	localQueue.ExamineContents(func(content []*weightSearchRatio1Bound) {
		for i := range content {
			if ws.checkRangeIsSubrangeOf(content[i], add) {
				foundContainingRange = true
			}
		}
	})
	return foundContainingRange
}

func (ws *WeightSearcherRatio1) maintainQueue(queue *util_collection.QueueStackFiloConcurrent[*weightSearchRatio1Bound], direction bool) {
	if direction {
		queue.UpdateContents(func(bounds []*weightSearchRatio1Bound) []*weightSearchRatio1Bound {
			slices.SortFunc(bounds, func(a, b *weightSearchRatio1Bound) int {
				return cmp.Compare(b.bestAccuracy, a.bestAccuracy)
			})
			return bounds
		})
	} else {
		queue.UpdateContents(func(bounds []*weightSearchRatio1Bound) []*weightSearchRatio1Bound {
			slices.SortFunc(bounds, func(a, b *weightSearchRatio1Bound) int {
				return cmp.Compare(a.bestAccuracy, b.bestAccuracy)
			})
			return bounds
		})
	}
}
