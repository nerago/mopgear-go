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
)

const (
	c_searchExtended1_threads = 32

	c_searchExtended1_perfect                  = 99.999
	c_searchExtended1_abandonBranchAccuracyGap = 0.2
	c_searchExtended1_largeAccuracyGap         = 0.4
	c_searchExtended1_equalAccuracyGap         = 0.01
	c_searchExtended1_goalAccuracyGap          = 0.01
	c_searchExtended1_marginalWeightGap        = 0.05

	c_searchExtended1_minRunEarlySizeCut = 8
	c_searchExtended1_minRunLateSizeCut  = 5

	c_searchExtended1_probeA      = 0.25
	c_searchExtended1_probeMiddle = 0.5
	c_searchExtended1_probeB      = 0.75

	c_searchExtended1_maxNodeDepth = 150
	c_searchExtended1_maxCombos    = 8 * 5
)

type WeightSearcherExtended1 struct {
	comboCount              int8
	statTypes               []stats.StatType
	simTypes                []stats.SimType
	targetRatio             weight_types.SimPriorityBasic
	initialEvaluateAccuracy EvaluateAccuracyPrepared
	initialBound            *searchEx1Bound

	bestResult util_rank.BestCollector1Concurrent[weight_types.Weight2Extended]
	shutdown   bool

	poolQueue util.TypedPool[searchEx1Bound]
}

type searchEx1Point [c_searchExtended1_maxCombos]float64

type searchEx1Bound struct {
	rangeMin   searchEx1Point
	rangeMax   searchEx1Point
	divideAxis int8
	nodeDepth  int16
}

type searchEx1Probe struct {
	accuracy float64
	axis     int8
	isHigh   bool
	point    searchEx1Point
}

func (ws *WeightSearcherExtended1) Init(statTypes []stats.StatType, targetRatio weight_types.SimPriorityBasic) {
	ws.statTypes = statTypes
	ws.simTypes = targetRatio.SimTypes()
	ws.targetRatio = targetRatio
	ws.comboCount = int8(len(ws.statTypes) * len(ws.simTypes))
	if ws.comboCount > c_searchExtended1_maxCombos {
		panic("don't support that many stat/sim combos")
	}
}

func (ws *WeightSearcherExtended1) SupplyData(inputData []weight_types.WeightInput) {
	ws.initialEvaluateAccuracy.Init(inputData, &ws.targetRatio, true)
}

func (ws *WeightSearcherExtended1) SetRanges(weightMin, weightMax float64) {
	bound := ws.poolQueue.Get()
	bound.divideAxis = 0
	bound.nodeDepth = 0
	for i := range ws.comboCount {
		bound.rangeMin[i] = weightMin
		bound.rangeMax[i] = weightMax
	}
	ws.initialBound = bound
}

func (ws *WeightSearcherExtended1) Run(cancel util_async.CancelSignal) weight_types.Weight2Extended {
	threadCount := c_searchExtended1_threads
	queue := util_collection.QueueStackStealingPool[*searchEx1Bound]{}

	mainThreadQueue := queue.MakeChild()
	mainThreadProbesReused := ws.newProbeSlice()
	if ws.initialSplits(mainThreadQueue, mainThreadProbesReused, threadCount, &ws.initialEvaluateAccuracy) {
		waitGroup := sync.WaitGroup{}
		for range threadCount - 1 {
			threadEvaluateAccuracy := ws.initialEvaluateAccuracy.Clone()
			waitGroup.Go(func() {
				ws.threadLoop(cancel, queue.MakeChild(), ws.newProbeSlice(), threadEvaluateAccuracy)
			})
		}

		ws.threadLoop(cancel, mainThreadQueue, mainThreadProbesReused, &ws.initialEvaluateAccuracy)
		waitGroup.Wait()
	}

	bestWeight := ws.bestResult.GetBestOrNilValue()
	//for _, simType := range ws.simTypes {
	//	// TODO useful scaling?
	//	bestWeight.SetSimScale(simType, 1, 0, ws.targetRatio.GetOrPanic(simType))
	//}
	bestWeight.FinishAndValidate()
	return bestWeight
}

func (ws *WeightSearcherExtended1) initialSplits(localQueue *util_collection.QueueStackPoolChild[*searchEx1Bound], probesReused []*searchEx1Probe, targetCount int, evaluateAccuracy *EvaluateAccuracyPrepared) bool {
	localQueue.Push(ws.initialBound)
	for localQueue.CountLocal() < targetCount {
		if !ws.threadStep(localQueue, probesReused, evaluateAccuracy) {
			return false
		}
	}
	return true
}

func (ws *WeightSearcherExtended1) threadLoop(cancel util_async.CancelSignal, localQueue *util_collection.QueueStackPoolChild[*searchEx1Bound], probesReused []*searchEx1Probe, evaluateAccuracy *EvaluateAccuracyPrepared) {
	iterCount := 0

	for cancel.ShouldContinue() && !ws.shutdown {
		if !ws.threadStep(localQueue, probesReused, evaluateAccuracy) {
			break
		}

		if iterCount%1000 == 0 {
			fmt.Printf("search-ex i=%d q=%d b=%f\n", iterCount, localQueue.CountLocal(), ws.bestResult.GetBestValue())
		}
		iterCount++
	}
}

func (ws *WeightSearcherExtended1) threadStep(localQueue *util_collection.QueueStackPoolChild[*searchEx1Bound], probesReused []*searchEx1Probe, evaluateAccuracy *EvaluateAccuracyPrepared) bool {
	entry, hasEntry := localQueue.Pop()
	if !hasEntry {
		return false
	}

	if entry.divideAxis != -1 {
		ws.opDivide(entry, localQueue)
	} else {
		ws.opSearch(entry, probesReused, localQueue, evaluateAccuracy)
	}

	ws.poolQueue.Put(entry)
	return true
}

func (ws *WeightSearcherExtended1) newProbeSlice() []*searchEx1Probe {
	slice := make([]*searchEx1Probe, ws.comboCount*2+1)
	for i := range slice {
		slice[i] = new(searchEx1Probe)
	}
	return slice
}

func (ws *WeightSearcherExtended1) evaluateScore(weightArray *searchEx1Point, evaluateAccuracy *EvaluateAccuracyPrepared) float64 {
	weights := weight_types.Weight2Extended_Make(ws.statTypes, ws.simTypes)

	index := 0
	for _, statType := range ws.statTypes {
		for _, simType := range ws.simTypes {
			weights.PutWeight(statType, simType, weightArray[index])
			index++
		}
	}
	for _, simType := range ws.simTypes {
		ratio := ws.targetRatio.GetOrPanic(simType)
		weights.SetSimScale(simType, 1, 0, ratio)
	}

	accuracy := evaluateAccuracy.EvaluateWeight2(weights)
	ws.bestResult.Offer(weights, accuracy)
	if accuracy >= c_searchExtended1_perfect {
		ws.shutdown = true
	}
	return accuracy
}

// dumb division in halves
func (ws *WeightSearcherExtended1) opDivide(bound *searchEx1Bound, localQueue *util_collection.QueueStackPoolChild[*searchEx1Bound]) {
	axis := bound.divideAxis
	mid := valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_searchExtended1_probeMiddle)

	var nextAxis int8
	if axis < ws.comboCount-1 {
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

func (ws *WeightSearcherExtended1) opSearch(bound *searchEx1Bound, probes []*searchEx1Probe, localQueue *util_collection.QueueStackPoolChild[*searchEx1Bound], evaluateAccuracy *EvaluateAccuracyPrepared) {
	probes = ws.createAndSetProbes(bound, probes, evaluateAccuracy)

	if bound.nodeDepth >= c_searchExtended1_maxNodeDepth {
		return
	} else if bestValue := ws.bestResult.GetBestValue(); probes[0].accuracy < bestValue-c_searchExtended1_abandonBranchAccuracyGap {
		// done
	} else if gapFirstProbeToLast := probes[0].accuracy - probes[len(probes)-1].accuracy; gapFirstProbeToLast <= c_searchExtended1_goalAccuracyGap {
		// done
	} else {
		cutPoint := ws.chooseCut(probes)
		probesAfterCut := probes[0 : cutPoint+1]
		ws.chooseSplitMode(probesAfterCut, bound, localQueue)
	}
}

func (ws *WeightSearcherExtended1) createAndSetProbes(bound *searchEx1Bound, probes []*searchEx1Probe, evaluateAccuracy *EvaluateAccuracyPrepared) []*searchEx1Probe {
	middle := probes[0]
	ws.sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_searchExtended1_probeMiddle, &middle.point)
	middle.accuracy = ws.evaluateScore(&middle.point, evaluateAccuracy)
	middle.axis = -1

	index := 1
	for axis := range ws.comboCount {
		lo := probes[index]
		lo.point = middle.point
		lo.point[axis] = valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_searchExtended1_probeA)
		lo.axis = axis
		lo.isHigh = false
		lo.accuracy = ws.evaluateScore(&lo.point, evaluateAccuracy)
		index++

		hi := probes[index]
		hi.point = middle.point
		hi.point[axis] = valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_searchExtended1_probeB)
		hi.axis = axis
		hi.isHigh = true
		hi.accuracy = ws.evaluateScore(&hi.point, evaluateAccuracy)
		index++
	}

	slices.SortStableFunc(probes, func(a, b *searchEx1Probe) int { return cmp.Compare(b.accuracy, a.accuracy) })
	return probes
}

// returns last index to include
func (ws *WeightSearcherExtended1) chooseCut(probes []*searchEx1Probe) int {
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
			if runStartIndex <= 1 && runSize >= c_searchExtended1_minRunEarlySizeCut && runSize < len(probes) {
				return runStartIndex + runSize - 1
			} else if runStartIndex > 1 && runSize >= c_searchExtended1_minRunLateSizeCut && runSize < len(probes) {
				return runStartIndex + runSize - 1
			}
		}
	}

	// default, cut at one quarter of range
	return len(probes) / 4
}

func (ws *WeightSearcherExtended1) chooseSplitMode(probes []*searchEx1Probe, bound *searchEx1Bound, localQueue *util_collection.QueueStackPoolChild[*searchEx1Bound]) {
	includeMiddle := false
	hiSlice := make([]*searchEx1Probe, 0, len(probes))
	loSlice := make([]*searchEx1Probe, 0, len(probes))
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

func (ws *WeightSearcherExtended1) search2MakeFollowupShrink(bound *searchEx1Bound, localQueue *util_collection.QueueStackPoolChild[*searchEx1Bound]) {
	add := ws.poolQueue.Get()
	add.divideAxis = -1
	add.nodeDepth = bound.nodeDepth + 1

	ws.sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_searchExtended1_probeA, &add.rangeMin)
	ws.sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_searchExtended1_probeB, &add.rangeMax)

	if ws.rangeIsMarginal(&add.rangeMin, &add.rangeMax) || ws.containedByAnotherQueueEntry(add, localQueue) {
		ws.poolQueue.Put(add)
	} else {
		localQueue.Push(add)
	}
}

func (ws *WeightSearcherExtended1) search2MakeFollowupShrinkExceptAxis(bound *searchEx1Bound, axis int8, localQueue *util_collection.QueueStackPoolChild[*searchEx1Bound]) {
	add := ws.poolQueue.Get()
	add.divideAxis = -1
	add.nodeDepth = bound.nodeDepth + 1

	ws.sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_searchExtended1_probeA, &add.rangeMin)
	add.rangeMin[axis] = bound.rangeMin[axis]
	ws.sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_searchExtended1_probeB, &add.rangeMax)
	add.rangeMax[axis] = bound.rangeMax[axis]

	if ws.rangeIsMarginal(&add.rangeMin, &add.rangeMax) || ws.containedByAnotherQueueEntry(add, localQueue) {
		ws.poolQueue.Put(add)
	} else {
		localQueue.Push(add)
	}
}

func (ws *WeightSearcherExtended1) search2MakeFollowupCommon2(probes []*searchEx1Probe, bound *searchEx1Bound, localQueue *util_collection.QueueStackPoolChild[*searchEx1Bound]) {
	add := ws.poolQueue.Get()
	add.divideAxis = -1
	add.nodeDepth = bound.nodeDepth + 1
	add.rangeMin = probes[0].point
	add.rangeMax = probes[0].point

	rangeMin := &add.rangeMin
	rangeMax := &add.rangeMax
	for i := 1; i < len(probes); i++ {
		point := &probes[i].point
		for axis := range ws.comboCount {
			rangeMin[axis] = min(rangeMin[axis], point[axis])
			rangeMax[axis] = max(rangeMax[axis], point[axis])
		}
	}

	for axis := range ws.comboCount {
		if rangeMin[axis] == rangeMax[axis] {
			rangeMin[axis] = bound.rangeMin[axis]
			rangeMax[axis] = bound.rangeMax[axis]
		} else {
			// TODO is there a good reason to shrink here
			oldExtent := bound.rangeMax[axis] - bound.rangeMin[axis]
			rangeMin[axis] -= c_searchExtended1_probeA * oldExtent
			rangeMax[axis] += c_searchExtended1_probeA * oldExtent
		}
	}

	if (bound.rangeMin == add.rangeMin && bound.rangeMax == add.rangeMin) || ws.rangeIsMarginal(&add.rangeMin, &add.rangeMax) || ws.containedByAnotherQueueEntry(add, localQueue) {
		ws.poolQueue.Put(add)
	} else {
		localQueue.Push(add)
	}
}

func (ws *WeightSearcherExtended1) checkRangeIsSubrangeOf(outer, inner *searchEx1Bound) bool {
	for i := range ws.comboCount {
		if outer.rangeMin[i] <= inner.rangeMin[i] && inner.rangeMax[i] <= outer.rangeMax[i] {
			// yes
		} else {
			return false
		}
	}
	return true
}

func (ws *WeightSearcherExtended1) rangeIsMarginal(rangeMin *searchEx1Point, rangeMax *searchEx1Point) bool {
	for i := range ws.comboCount {
		if rangeMax[i]-rangeMin[i] > c_searchExtended1_marginalWeightGap {
			return false
		}
	}
	return true
}

func (ws *WeightSearcherExtended1) largeAccuracyGap(a, b float64) bool {
	return math.Abs(a-b) >= c_searchExtended1_largeAccuracyGap
}

func (ws *WeightSearcherExtended1) equalAccuracyGap(a, b float64) bool {
	return math.Abs(a-b) < c_searchExtended1_equalAccuracyGap
}

func (ws *WeightSearcherExtended1) sliceInterpolate(rangeMin *searchEx1Point, rangeMax *searchEx1Point, ratio float64, out *searchEx1Point) {
	for i := range ws.comboCount {
		out[i] = valueInterpolate(rangeMin[i], rangeMax[i], ratio)
	}
}

func (ws *WeightSearcherExtended1) containedByAnotherQueueEntry(add *searchEx1Bound, localQueue *util_collection.QueueStackPoolChild[*searchEx1Bound]) bool {
	foundContainingRange := false
	localQueue.ExamineContents(func(content []*searchEx1Bound) {
		for i := range content {
			if ws.checkRangeIsSubrangeOf(content[i], add) {
				foundContainingRange = true
			}
		}
	})
	return foundContainingRange
}
