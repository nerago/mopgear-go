package weightfind

import (
	"cmp"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_rank"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
	"math"
	"slices"
	"sync"

	"github.com/bartolsthoorn/gohighs/highs"
)

const (
	c_search3_abandonBranchAccuracyGap = 0.4
	c_search3_largeAccuracyGap         = 0.5
	c_search3_equalAccuracyGap         = 0.01
	c_search3_goalAccuracyGap          = 0.001
	c_search3_marginalWeightGap        = 0.001

	c_search3_minRunEarlySizeCut = 4
	c_search3_minRunLateSizeCut  = 2

	c_search3_probeA      = 0.25
	c_search3_probeMiddle = 0.5
	c_search3_probeB      = 0.75

	c_search3_maxNodeDepth = 50
	c_search3_maxStats     = 8
)

type WeightSearcher3 struct {
	typeCount           int
	statTypes           []stats.StatType
	simTypes            []stats.SimType
	targetRatio         weight_types.SimPriorityBasic
	evaluateAccuracy    EvaluateAccuracyPrepared
	initialBound        *weightSearch2FastBound
	AccuracyStatistical bool

	bestResult util_rank.BestCollector1Concurrent[weight_types.Weight1Basic]

	poolQueue util.TypedPool[weightSearch2FastBound]
}

type weightSearch2FastPoint [c_search3_maxStats]float64

type weightSearch2FastBound struct {
	rangeMin   weightSearch2FastPoint
	rangeMax   weightSearch2FastPoint
	divideAxis int
	nodeDepth  int
}

type weightSearch2FastProbe struct {
	accuracy float64
	axis     int
	isHigh   bool
	point    weightSearch2FastPoint
}

func (ws *WeightSearcher3) Init(statTypes []stats.StatType, targetRatio weight_types.SimPriorityBasic) {
	ws.typeCount = len(statTypes)
	ws.statTypes = statTypes
	if len(statTypes) > c_search3_maxStats {
		panic("don't support that many stats")
	}
	ws.simTypes = targetRatio.SimTypes()
	ws.targetRatio = targetRatio
}

func (ws *WeightSearcher3) SupplyData(inputData []weight_types.WeightInput) {
	ws.evaluateAccuracy.Init(inputData, &ws.targetRatio, ws.AccuracyStatistical)
}

func (ws *WeightSearcher3) SetRanges(weightMin, weightMax float64) {
	bound := ws.poolQueue.Get()
	bound.divideAxis = 0
	bound.nodeDepth = 0
	for i := range ws.typeCount {
		bound.rangeMin[i] = weightMin
		bound.rangeMax[i] = weightMax
	}
	ws.initialBound = bound
}

func (ws *WeightSearcher3) Run(cancel util_async.CancelSignal) weight_types.WeightResult {
	stopwatch := util.StopwatchMakeStarted()
	threadCount := 4
	queue := util_collection.QueueStackFiloPoolParent[*weightSearch2FastBound]{}

	startingQueue := queue.MakeChild()
	startingProbesReused := ws.newProbeSlice()
	if ws.initialSplits(startingQueue, startingProbesReused, threadCount*2) {
		waitGroup := sync.WaitGroup{}
		for range threadCount - 1 {
			waitGroup.Go(func() {
				ws.threadLoop(cancel, queue.MakeChild(), ws.newProbeSlice())
			})
		}

		ws.threadLoop(cancel, startingQueue, startingProbesReused)
		waitGroup.Wait()
	}

	bestWeight := ws.bestResult.GetBestOrNilValue()
	bestWeight.NormalizeForBase(ws.statTypes)
	return weight_types.WeightResult{Weight: &bestWeight, SolveTime: stopwatch.Elapsed(), Status: highs.ModelStatusOptimal}
}

func (ws *WeightSearcher3) initialSplits(localQueue *util_collection.QueueStackFiloPoolChild[*weightSearch2FastBound], probesReused []*weightSearch2FastProbe, targetCount int) bool {
	localQueue.Push(ws.initialBound)
	for localQueue.CountLocal() < targetCount {
		if !ws.threadStep(localQueue, probesReused) {
			return false
		}
	}
	return true
}

func (ws *WeightSearcher3) threadLoop(cancel util_async.CancelSignal, localQueue *util_collection.QueueStackFiloPoolChild[*weightSearch2FastBound], probesReused []*weightSearch2FastProbe) {
	for cancel.ShouldContinue() {
		if !ws.threadStep(localQueue, probesReused) {
			break
		}
	}
}

func (ws *WeightSearcher3) threadStep(localQueue *util_collection.QueueStackFiloPoolChild[*weightSearch2FastBound], probesReused []*weightSearch2FastProbe) bool {
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

func (ws *WeightSearcher3) newProbeSlice() []*weightSearch2FastProbe {
	slice := make([]*weightSearch2FastProbe, ws.typeCount*2+1)
	for i := range slice {
		slice[i] = new(weightSearch2FastProbe)
	}
	return slice
}

func (ws *WeightSearcher3) evaluateScore(weightArray *weightSearch2FastPoint) float64 {
	weights := weight_types.Weight1Basic{}
	for i, statType := range ws.statTypes {
		weights.Put(statType, weightArray[i])
	}
	accuracy := ws.evaluateAccuracy.EvaluateWeight1(&weights)
	ws.bestResult.Offer(&weights, accuracy)
	return accuracy
}

// dumb division in halves
func (ws *WeightSearcher3) opDivide(bound *weightSearch2FastBound, localQueue *util_collection.QueueStackFiloPoolChild[*weightSearch2FastBound]) {
	axis := bound.divideAxis
	mid := valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_search3_probeMiddle)

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
	localQueue.Push(hiBound)

	loBound := ws.poolQueue.Get()
	loBound.divideAxis = nextAxis
	loBound.nodeDepth = bound.nodeDepth + 1
	loBound.rangeMin = bound.rangeMin
	loBound.rangeMin[axis] = mid
	loBound.rangeMax = bound.rangeMax
	localQueue.Push(loBound)
}

func (ws *WeightSearcher3) opSearch(bound *weightSearch2FastBound, probes []*weightSearch2FastProbe, localQueue *util_collection.QueueStackFiloPoolChild[*weightSearch2FastBound]) {
	probes = ws.createAndSetProbes(bound, probes)

	if bound.nodeDepth >= c_search3_maxNodeDepth {
		// done
	} else if probes[0].accuracy < ws.bestResult.GetBestValue()-c_search3_abandonBranchAccuracyGap {
		// done
	} else if gapFirstProbeToLast := probes[0].accuracy - probes[len(probes)-1].accuracy; gapFirstProbeToLast <= c_search3_goalAccuracyGap {
		// done
	} else {
		cutPoint := ws.chooseCut(probes)
		probesAfterCut := probes[0 : cutPoint+1]
		ws.chooseSplitMode(probesAfterCut, bound, localQueue)
	}
}

func (ws *WeightSearcher3) createAndSetProbes(bound *weightSearch2FastBound, probes []*weightSearch2FastProbe) []*weightSearch2FastProbe {
	middle := probes[0]
	ws.sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_search3_probeMiddle, &middle.point)
	middle.accuracy = ws.evaluateScore(&middle.point)
	middle.axis = -1

	index := 1
	for axis := range ws.typeCount {
		lo := probes[index]
		lo.point = middle.point
		lo.point[axis] = valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_search3_probeA)
		lo.axis = axis
		lo.isHigh = false
		lo.accuracy = ws.evaluateScore(&lo.point)
		index++

		hi := probes[index]
		hi.point = middle.point
		hi.point[axis] = valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_search3_probeB)
		hi.axis = axis
		hi.isHigh = true
		hi.accuracy = ws.evaluateScore(&hi.point)
		index++
	}

	slices.SortStableFunc(probes, func(a, b *weightSearch2FastProbe) int { return cmp.Compare(b.accuracy, a.accuracy) })
	return probes
}

// returns last index to include
func (ws *WeightSearcher3) chooseCut(probes []*weightSearch2FastProbe) int {
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
			if runStartIndex <= 1 && runSize >= c_search3_minRunEarlySizeCut && runSize < len(probes) {
				return runStartIndex + runSize - 1
			} else if runStartIndex > 1 && runSize >= c_search3_minRunLateSizeCut && runSize < len(probes) {
				return runStartIndex + runSize - 1
			}
		}
	}

	// default, cut at one quarter of range
	return len(probes) / 4
}

func (ws *WeightSearcher3) chooseSplitMode(probes []*weightSearch2FastProbe, bound *weightSearch2FastBound, localQueue *util_collection.QueueStackFiloPoolChild[*weightSearch2FastBound]) {
	includeMiddle := false
	hiSlice := make([]*weightSearch2FastProbe, 0, len(probes))
	loSlice := make([]*weightSearch2FastProbe, 0, len(probes))
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
		ws.search2MakeFollowupShrink(bound, localQueue)
	} else if hi == 0 || lo == 0 {
		// everything on once side, with or without the middle
		ws.search2MakeFollowupCommon2(probes, bound, localQueue)
	} else if hi == 1 && lo == 1 && hiSlice[0].axis == loSlice[0].axis {
		// basically just know its around middle
		// go for general shrinking but keep that full range since we know its of interest
		ws.search2MakeFollowupShrinkExceptAxis(bound, hiSlice[0].axis, localQueue)
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
		ws.search2MakeFollowupShrink(bound, localQueue)
	}
}

func (ws *WeightSearcher3) search2MakeFollowupShrink(bound *weightSearch2FastBound, localQueue *util_collection.QueueStackFiloPoolChild[*weightSearch2FastBound]) {
	add := ws.poolQueue.Get()
	add.divideAxis = -1
	add.nodeDepth = bound.nodeDepth + 1

	ws.sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_search3_probeA, &add.rangeMin)
	ws.sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_search3_probeB, &add.rangeMax)

	if ws.rangeIsMarginal(&add.rangeMin, &add.rangeMax) || ws.containedByAnotherQueueEntry(add, localQueue) {
		ws.poolQueue.Put(add)
	} else {
		localQueue.Push(add)
	}
}

func (ws *WeightSearcher3) search2MakeFollowupShrinkExceptAxis(bound *weightSearch2FastBound, axis int, localQueue *util_collection.QueueStackFiloPoolChild[*weightSearch2FastBound]) {
	add := ws.poolQueue.Get()
	add.divideAxis = -1
	add.nodeDepth = bound.nodeDepth + 1

	ws.sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_search3_probeA, &add.rangeMin)
	add.rangeMin[axis] = bound.rangeMin[axis]
	ws.sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_search3_probeB, &add.rangeMax)
	add.rangeMax[axis] = bound.rangeMax[axis]

	if ws.rangeIsMarginal(&add.rangeMin, &add.rangeMax) || ws.containedByAnotherQueueEntry(add, localQueue) {
		ws.poolQueue.Put(add)
	} else {
		localQueue.Push(add)
	}
}

func (ws *WeightSearcher3) search2MakeFollowupCommon2(probes []*weightSearch2FastProbe, bound *weightSearch2FastBound, localQueue *util_collection.QueueStackFiloPoolChild[*weightSearch2FastBound]) {
	add := ws.poolQueue.Get()
	add.divideAxis = -1
	add.nodeDepth = bound.nodeDepth + 1
	add.rangeMin = probes[0].point
	add.rangeMax = probes[0].point

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
			rangeMin[axis] -= c_search3_probeA * oldExtent
			rangeMax[axis] += c_search3_probeA * oldExtent
		}
	}

	if (bound.rangeMin == add.rangeMin && bound.rangeMax == add.rangeMin) || ws.rangeIsMarginal(&add.rangeMin, &add.rangeMax) || ws.containedByAnotherQueueEntry(add, localQueue) {
		ws.poolQueue.Put(add)
	} else {
		localQueue.Push(add)
	}
}

func (ws *WeightSearcher3) checkRangeIsSubrangeOf(outer, inner *weightSearch2FastBound) bool {
	for i := range ws.typeCount {
		if outer.rangeMin[i] <= inner.rangeMin[i] && inner.rangeMax[i] <= outer.rangeMax[i] {
			// yes
		} else {
			return false
		}
	}
	return true
}

func (ws *WeightSearcher3) rangeIsMarginal(rangeMin *weightSearch2FastPoint, rangeMax *weightSearch2FastPoint) bool {
	for i := range ws.typeCount {
		if rangeMax[i]-rangeMin[i] > c_search3_marginalWeightGap {
			return false
		}
	}
	return true
}

func (ws *WeightSearcher3) largeAccuracyGap(a, b float64) bool {
	return math.Abs(a-b) >= c_search3_largeAccuracyGap
}

func (ws *WeightSearcher3) equalAccuracyGap(a, b float64) bool {
	return math.Abs(a-b) < c_search3_equalAccuracyGap
}

func (ws *WeightSearcher3) sliceInterpolate(rangeMin *weightSearch2FastPoint, rangeMax *weightSearch2FastPoint, ratio float64, out *weightSearch2FastPoint) {
	for i := range ws.typeCount {
		out[i] = valueInterpolate(rangeMin[i], rangeMax[i], ratio)
	}
}

func (ws *WeightSearcher3) containedByAnotherQueueEntry(add *weightSearch2FastBound, localQueue *util_collection.QueueStackFiloPoolChild[*weightSearch2FastBound]) bool {
	foundContainingRange := false
	localQueue.ExamineContents(func(content []*weightSearch2FastBound) {
		for i := range content {
			if ws.checkRangeIsSubrangeOf(content[i], add) {
				foundContainingRange = true
			}
		}
	})
	return foundContainingRange
}
