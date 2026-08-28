package weightfind

import (
	"cmp"
	"fmt"
	"slices"
	"sync"

	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
	"github.com/nerago/mopgear-go/util/util_rank"
)

const (
	c_search_threads   = 32
	c_search_maxCombos = 64

	c_search_perfect                  = 99.999
	c_search_abandonBranchAccuracyGap = 0.4
	c_search_largeAccuracyGap         = 0.2
	c_search_equalAccuracyGap         = 0.01
	c_search_goalAccuracyGap          = 0.01
	c_search_marginalWeightGap        = 0.1

	c_search_minRunEarlySizeCut = 8
	c_search_minRunLateSizeCut  = 5

	c_search_probeA      = 0.2
	c_search_probeMiddle = 0.5
	c_search_probeB      = 0.8

	c_search_maxNodeDepth = 100
)

type SearcherEngine struct {
	model      SearcherModel
	comboCount int8

	bestResult util_rank.BestCollector1Concurrent[pointType]
	shutdown   bool

	queueMain util_collection.QueueStackStealingPool[*boundType]
	pool      util.TypedPool[boundType]
}

type pointType [c_search_maxCombos]float64

type boundType struct {
	rangeMin   pointType
	rangeMax   pointType
	divideAxis int8
	nodeDepth  int16
}

type probeType struct {
	accuracy float64
	axis     int8
	isHigh   bool
	point    pointType
}

type threadLocal struct {
	model  ModelThreadLocals
	queue  *util_collection.QueueStackPoolChild[*boundType]
	probes []*probeType
	engine *SearcherEngine
}

func (ws *SearcherEngine) Init(model SearcherModel) {
	ws.model = model
	ws.comboCount = model.comboCount()
	if ws.comboCount > c_search_maxCombos {
		panic("don't support that many stat/sim combos")
	}
}

func (ws *SearcherEngine) makeInitialBound() *boundType {
	bound := ws.pool.Get()
	bound.divideAxis = 0
	bound.nodeDepth = 0

	for i, init := range ws.model.initialRanges() {
		bound.rangeMin[i] = init.Minimum
		bound.rangeMax[i] = init.Maximum
	}

	return bound
}

func (ws *SearcherEngine) Run(cancel util_async.CancelSignal) any {
	threadCount := c_search_threads

	mainLocals := ws.newLocals()
	mainLocals.queue.Push(ws.makeInitialBound())

	if mainLocals.initialSplits(threadCount) {
		waitGroup := sync.WaitGroup{}
		for range threadCount - 1 {
			waitGroup.Go(func() {
				local := ws.newLocals()
				local.threadLoop(cancel)
			})
		}

		mainLocals.threadLoop(cancel)
		waitGroup.Wait()
	}

	if bestPoint, hasBest := ws.bestResult.GetBestPointer(); hasBest {
		return ws.model.makeFinalResult(bestPoint)
	} else {
		return nil
	}
}

func (wst *threadLocal) initialSplits(targetCount int) bool {
	for wst.queue.CountLocal() < targetCount {
		if !wst.threadStep() {
			return false
		}
	}
	return true
}

func (wst *threadLocal) threadLoop(cancel util_async.CancelSignal) {
	iterCount := 0

	for cancel.ShouldContinue() && !wst.engine.shutdown {
		if !wst.threadStep() {
			break
		}

		if iterCount%1000 == 0 {
			fmt.Printf("search-ex i=%d q=%d b=%f\n",
				iterCount, wst.queue.CountLocal(), wst.engine.bestResult.GetBestValue())
		}
		iterCount++
	}
}

func (wst *threadLocal) threadStep() bool {
	entry, hasEntry := wst.queue.Pop()
	if !hasEntry {
		return false
	}

	if entry.divideAxis != -1 {
		wst.opDivide(entry)
	} else {
		wst.opSearch(entry)
	}

	wst.engine.pool.Put(entry)
	return true
}

func (wst *threadLocal) evaluateScore(point *pointType) float64 {
	score := wst.model.evaluateScore(point)
	if wst.engine.bestResult.OfferAndIsBetter(point, score) {
		if score >= c_search_perfect {
			wst.engine.shutdown = true
		}
	}
	return score
}

// dumb division in halves
func (wst *threadLocal) opDivide(bound *boundType) {
	axis := bound.divideAxis
	mid := valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_search_probeMiddle)

	var nextAxis int8
	if axis < wst.engine.comboCount-1 {
		nextAxis = axis + 1
	} else {
		nextAxis = -1
	}

	hiBound := wst.engine.pool.Get()
	hiBound.divideAxis = nextAxis
	hiBound.nodeDepth = bound.nodeDepth + 1
	hiBound.rangeMin = bound.rangeMin
	hiBound.rangeMax = bound.rangeMax
	hiBound.rangeMax[axis] = mid
	wst.queue.Push(hiBound)

	loBound := wst.engine.pool.Get()
	loBound.divideAxis = nextAxis
	loBound.nodeDepth = bound.nodeDepth + 1
	loBound.rangeMin = bound.rangeMin
	loBound.rangeMin[axis] = mid
	loBound.rangeMax = bound.rangeMax
	wst.queue.Push(loBound)
}

func (wst *threadLocal) opSearch(bound *boundType) {
	probes := wst.createAndSetProbes(bound)

	if bound.nodeDepth >= c_search_maxNodeDepth {
		return
	} else if bestValue := wst.engine.bestResult.GetBestValue(); probes[0].accuracy < bestValue-c_search_abandonBranchAccuracyGap {
		// done
	} else if gapFirstProbeToLast := probes[0].accuracy - probes[len(probes)-1].accuracy; gapFirstProbeToLast <= c_search_goalAccuracyGap {
		// done
	} else {
		cutPoint := wst.chooseCut(probes)
		probesAfterCut := probes[0 : cutPoint+1]
		wst.chooseSplitMode(probesAfterCut, bound)
	}
}

func (wst *threadLocal) createAndSetProbes(bound *boundType) []*probeType {
	probes := wst.probes
	middle := probes[0]
	sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_search_probeMiddle, &middle.point, wst.engine.comboCount)
	middle.accuracy = wst.evaluateScore(&middle.point)
	middle.axis = -1

	// TODO don't probe some ranges if particularly marginal?

	index := 1
	for axis := range wst.engine.comboCount {
		lo := probes[index]
		lo.point = middle.point
		lo.point[axis] = valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_search_probeA)
		lo.axis = axis
		lo.isHigh = false
		lo.accuracy = wst.model.evaluateScore(&lo.point)
		index++

		hi := probes[index]
		hi.point = middle.point
		hi.point[axis] = valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_search_probeB)
		hi.axis = axis
		hi.isHigh = true
		hi.accuracy = wst.model.evaluateScore(&hi.point)
		index++
	}

	slices.SortStableFunc(probes, func(a, b *probeType) int { return cmp.Compare(b.accuracy, a.accuracy) })
	return probes
}

// returns last index to include
func (wst *threadLocal) chooseCut(probes []*probeType) int {
	// cut after a large gap
	for index := range len(probes) / 2 {
		if largeAccuracyGap(probes[index].accuracy, probes[index+1].accuracy) {
			return index
		}
	}

	// cut after a run of consecutive values
	for runStartIndex := range len(probes) / 3 {
		if equalAccuracyGap(probes[runStartIndex].accuracy, probes[runStartIndex+1].accuracy) {
			runSize := 2
			for check := runStartIndex + 1; check < len(probes)-1; check++ {
				if equalAccuracyGap(probes[check].accuracy, probes[check+1].accuracy) {
					runSize++
				} else {
					break
				}
			}
			if runStartIndex <= 1 && runSize >= c_search_minRunEarlySizeCut && runSize < len(probes) {
				return runStartIndex + runSize - 1
			} else if runStartIndex > 1 && runSize >= c_search_minRunLateSizeCut && runSize < len(probes) {
				return runStartIndex + runSize - 1
			}
		}
	}

	// default, cut at one quarter of range
	return len(probes) / 4
}

func (wst *threadLocal) chooseSplitMode(probes []*probeType, bound *boundType) {
	includeMiddle := false
	hiSlice := make([]*probeType, 0, len(probes))
	loSlice := make([]*probeType, 0, len(probes))
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
		wst.search2MakeFollowupShrink(bound)
	} else if hi == 0 || lo == 0 {
		// everything on once side, with or without the middle
		wst.search2MakeFollowupCommon2(probes, bound)
	} else if hi == 1 && lo == 1 && hiSlice[0].axis == loSlice[0].axis {
		// basically just know its around middle
		// go for general shrinking but keep that full range since we know its of interest
		wst.search2MakeFollowupShrinkExceptAxis(bound, hiSlice[0].axis)
	} else if hi == 1 && lo == 1 {
		// go ahead and extend a bit in hi and low directions on different axes
		wst.search2MakeFollowupCommon2(probes, bound)
	} else if (hi == 1 && lo > 1) || (lo == 1 && hi > 1) {
		// almost all on one side except for one
		// could have an axis repeated, but common should be happy enough with one good value
		wst.search2MakeFollowupCommon2(probes, bound)
	} else if !includeMiddle {
		// make 2 regions on either side of middle
		wst.search2MakeFollowupCommon2(loSlice, bound)
		wst.search2MakeFollowupCommon2(hiSlice, bound)
	} else {
		// these will have a fair bit of overlap, but so be it
		wst.search2MakeFollowupCommon2(loSlice, bound)
		wst.search2MakeFollowupCommon2(hiSlice, bound)
		wst.search2MakeFollowupShrink(bound)
	}
}

func (wst *threadLocal) search2MakeFollowupShrink(bound *boundType) {
	comboCount := wst.engine.comboCount

	add := wst.engine.pool.Get()
	add.divideAxis = -1
	add.nodeDepth = bound.nodeDepth + 1

	sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_search_probeA, &add.rangeMin, comboCount)
	sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_search_probeB, &add.rangeMax, comboCount)

	if rangeIsMarginal(&add.rangeMin, &add.rangeMax, comboCount) || wst.containedByAnotherQueueEntry(add) {
		wst.engine.pool.Put(add)
	} else {
		wst.queue.Push(add)
	}
}

func (wst *threadLocal) search2MakeFollowupShrinkExceptAxis(bound *boundType, axis int8) {
	add := wst.engine.pool.Get()
	add.divideAxis = -1
	add.nodeDepth = bound.nodeDepth + 1

	sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_search_probeA, &add.rangeMin, wst.engine.comboCount)
	add.rangeMin[axis] = bound.rangeMin[axis]
	sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_search_probeB, &add.rangeMax, wst.engine.comboCount)
	add.rangeMax[axis] = bound.rangeMax[axis]

	if rangeIsMarginal(&add.rangeMin, &add.rangeMax, wst.engine.comboCount) || wst.containedByAnotherQueueEntry(add) {
		wst.engine.pool.Put(add)
	} else {
		wst.queue.Push(add)
	}
}

func (wst *threadLocal) search2MakeFollowupCommon2(probes []*probeType, bound *boundType) {
	add := wst.engine.pool.Get()
	add.divideAxis = -1
	add.nodeDepth = bound.nodeDepth + 1
	add.rangeMin = probes[0].point
	add.rangeMax = probes[0].point

	rangeMin := &add.rangeMin
	rangeMax := &add.rangeMax
	for i := 1; i < len(probes); i++ {
		point := &probes[i].point
		for axis := range wst.engine.comboCount {
			rangeMin[axis] = min(rangeMin[axis], point[axis])
			rangeMax[axis] = max(rangeMax[axis], point[axis])
		}
	}

	for axis := range wst.engine.comboCount {
		if rangeMin[axis] == rangeMax[axis] {
			rangeMin[axis] = bound.rangeMin[axis]
			rangeMax[axis] = bound.rangeMax[axis]
		} else {
			// TODO is there a good reason to shrink here
			oldExtent := bound.rangeMax[axis] - bound.rangeMin[axis]
			rangeMin[axis] -= c_search_probeA * oldExtent
			rangeMax[axis] += c_search_probeA * oldExtent
		}
	}

	if (bound.rangeMin == add.rangeMin && bound.rangeMax == add.rangeMin) || rangeIsMarginal(&add.rangeMin, &add.rangeMax, wst.engine.comboCount) || wst.containedByAnotherQueueEntry(add) {
		wst.engine.pool.Put(add)
	} else {
		wst.queue.Push(add)
	}
}

func (wst *threadLocal) containedByAnotherQueueEntry(add *boundType) bool {
	foundContainingRange := false
	wst.queue.ExamineContents(func(content []*boundType) {
		for i := range content {
			if checkRangeIsSubrangeOf(content[i], add, wst.engine.comboCount) {
				foundContainingRange = true
			}
		}
	})
	return foundContainingRange
}

func (ws *SearcherEngine) newLocals() *threadLocal {
	return &threadLocal{
		model:  ws.model.newLocals(),
		queue:  ws.queueMain.MakeChild(),
		probes: ws.newProbeSlice(),
	}
}

func (ws *SearcherEngine) newProbeSlice() []*probeType {
	slice := make([]*probeType, ws.comboCount*2+1)
	for i := range slice {
		slice[i] = new(probeType)
	}
	return slice
}
