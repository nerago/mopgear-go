package weightfind

import (
	"cmp"
	"math"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_rank"
	"paladin_gearing_go/weightfind/weight_highs"
	"slices"
	"sync"
)

const (
	// c_search2_abandonBranchAccuracyGap = 0.4
	// c_search2_largeAccuracyGap         = 0.5
	// c_search2_equalAccuracyGap         = 0.01
	// c_search2_goalAccuracyGap          = 0.001
	// c_search2_marginalWeightGap        = 0.001
	//
	// c_search2_minRunEarlySizeCut = 4
	// c_search2_minRunLateSizeCut  = 2
	//
	// c_search2_probeA       = 0.25
	// c_search2_probe_middle = 0.5
	// c_search2_probeB       = 0.75
	//
	// c_search2_max_node_depth = 40
	// c_search2_use_final_op   = false
	// c_search2_debug          = false

	c_search2fast_max_stats = 8
)

type WeightSearcher2Fast struct {
	typeCount        int
	statTypes        []stats.StatType
	simTypes         []stats.SimType
	targetRatio      stats.SimData
	evaluateAccuracy EvaluateAccuracyPrepared
	printer          *util.PrintRecorder
	AccuracyMode     int

	queue      util.QueueStackFilo[*weightSearch2FastBound]
	bestResult util_rank.BestCollector1[weight_highs.WeightResult]

	poolQueue typedPool[weightSearch2FastBound]
	poolProbe typedPool[weightSearch2FastProbe]
}

type typedPool[T any] struct {
	pool sync.Pool
}

func (p *typedPool[T]) Get() *T {
	value := p.pool.Get()
	if value != nil {
		return value.(*T)
	} else {
		return new(T)
	}
}

func (p *typedPool[T]) Put(instance *T) {
	p.pool.Put(instance)
}

type weightSearch2FastPoint [c_search2fast_max_stats]float64

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

func (ws *WeightSearcher2Fast) Init(statTypes []stats.StatType, targetRatio stats.SimData, printer *util.PrintRecorder) {
	ws.typeCount = len(statTypes)
	ws.statTypes = statTypes
	if len(statTypes) > c_search2fast_max_stats {
		panic("don't support that many stats")
	}
	ws.simTypes = targetRatio.NonZeroTypes()
	ws.targetRatio = targetRatio
	ws.printer = printer
}

func (ws *WeightSearcher2Fast) SupplyData(inputData []weight_highs.WeightInput) {
	ws.evaluateAccuracy.Init(inputData, ws.targetRatio, ws.AccuracyMode)
}

func (ws *WeightSearcher2Fast) SetRanges(weightMin, weightMax float64) {
	bound := ws.poolQueue.Get()
	bound.divideAxis = 0
	bound.nodeDepth = 0
	for i := range ws.typeCount {
		bound.rangeMin[i] = weightMin
		bound.rangeMax[i] = weightMax
	}
	ws.queue.Push(bound)
}

func (ws *WeightSearcher2Fast) Run(cancel util_async.CancelSignal) weight_highs.WeightResult {
	iterCount := 0
	for cancel.ShouldContinue() {
		entry, hasEntry := ws.queue.Pop()
		if !hasEntry {
			break
		}

		if entry.divideAxis != -1 {
			ws.opDivide(entry)
		} else {
			ws.opSearch(entry)
		}

		ws.poolQueue.Put(entry)

		iterCount++
		if iterCount%10 == 0 {
			ws.queueMaintenance()
		}
	}

	bestWeight := ws.bestResult.GetBestOrNilValue()
	return bestWeight.ScaleForBaseStat(ws.statTypes[0])
}

func (ws *WeightSearcher2Fast) evaluateScore(weightArray *weightSearch2FastPoint) float64 {
	weights := weight_highs.WeightResult{}
	for i, statType := range ws.statTypes {
		weights.Put(statType, weightArray[i])
	}
	accuracy := ws.evaluateAccuracy.EvaluateWeight(weights)
	ws.bestResult.Offer(&weights, accuracy)
	return accuracy
}

// dumb division in halves
func (ws *WeightSearcher2Fast) opDivide(bound *weightSearch2FastBound) {
	axis := bound.divideAxis
	mid := valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_search2_probe_middle)

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
	ws.queue.Push(hiBound)

	loBound := ws.poolQueue.Get()
	loBound.divideAxis = nextAxis
	loBound.nodeDepth = bound.nodeDepth + 1
	loBound.rangeMin = bound.rangeMin
	loBound.rangeMin[axis] = mid
	loBound.rangeMax = bound.rangeMax
	ws.queue.Push(loBound)
}

func (ws *WeightSearcher2Fast) opSearch(bound *weightSearch2FastBound) {
	probes := ws.search2CreateAndSetProbes(bound)

	if bound.nodeDepth >= c_search2_max_node_depth {
		// done
	} else if probes[0].accuracy < ws.bestResult.BestValue-c_search2_abandonBranchAccuracyGap {
		// done
	} else if gapFirstProbeToLast := probes[0].accuracy - probes[len(probes)-1].accuracy; gapFirstProbeToLast <= c_search2_goalAccuracyGap {
		// done
	} else {
		cutPoint := ws.search2ChooseCut(probes)
		probesAfterCut := probes[0 : cutPoint+1]
		ws.search2ChooseSplitMode(probesAfterCut, bound)
	}

	for i := range probes {
		ws.poolProbe.Put(probes[i])
	}
}

func (ws *WeightSearcher2Fast) search2CreateAndSetProbes(bound *weightSearch2FastBound) []*weightSearch2FastProbe {
	probes := make([]*weightSearch2FastProbe, 1, ws.typeCount*2+1)

	middle := ws.poolProbe.Get()
	ws.sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_search2_probe_middle, &middle.point)
	middle.accuracy = ws.evaluateScore(&probes[0].point)
	middle.axis = -1
	probes[0] = middle

	for axis := range ws.typeCount {
		lo := ws.poolProbe.Get()
		lo.point = middle.point
		lo.point[axis] = valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_search2_probeA)
		lo.axis = axis
		lo.isHigh = false
		lo.accuracy = ws.evaluateScore(&lo.point)

		hi := ws.poolProbe.Get()
		hi.point = middle.point
		hi.point[axis] = valueInterpolate(bound.rangeMin[axis], bound.rangeMax[axis], c_search2_probeB)
		hi.axis = axis
		hi.isHigh = true
		hi.accuracy = ws.evaluateScore(&hi.point)
	}

	slices.SortStableFunc(probes, func(a, b *weightSearch2FastProbe) int { return cmp.Compare(b.accuracy, a.accuracy) })
	return probes
}

// returns last index to include
func (ws *WeightSearcher2Fast) search2ChooseCut(probes []*weightSearch2FastProbe) int {
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

func (ws *WeightSearcher2Fast) search2ChooseSplitMode(probes []*weightSearch2FastProbe, bound *weightSearch2FastBound) {
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
		ws.search2MakeFollowupShrink(bound)
	} else if hi == 0 || lo == 0 {
		// everything on once side, with or without the middle
		ws.search2MakeFollowupCommon2(probes, bound)
	} else if hi == 1 && lo == 1 && hiSlice[0].axis == loSlice[0].axis {
		// basically just know its around middle
		// go for general shrinking but keep that full range since we know its of interest
		ws.search2MakeFollowupShrinkExceptAxis(bound, hiSlice[0].axis)
	} else if hi == 1 && lo == 1 {
		// go ahead and extend a bit in hi and low directions on different axes
		ws.search2MakeFollowupCommon2(probes, bound)
	} else if (hi == 1 && lo > 1) || (lo == 1 && hi > 1) {
		// almost all on one side except for one
		// could have an axis repeated, but common should be happy enough with one good value
		ws.search2MakeFollowupCommon2(probes, bound)
	} else if !includeMiddle {
		// make 2 regions on either side of middle
		ws.search2MakeFollowupCommon2(loSlice, bound)
		ws.search2MakeFollowupCommon2(hiSlice, bound)
	} else {
		// these will have a fair bit of overlap, but so be it
		ws.search2MakeFollowupCommon2(loSlice, bound)
		ws.search2MakeFollowupCommon2(hiSlice, bound)
		ws.search2MakeFollowupShrink(bound)
	}
}

func (ws *WeightSearcher2Fast) search2MakeFollowupShrink(bound *weightSearch2FastBound) {
	add := ws.poolQueue.Get()
	add.divideAxis = -1
	add.nodeDepth = bound.nodeDepth + 1

	ws.sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_search2_probeA, &add.rangeMin)
	ws.sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_search2_probeB, &add.rangeMax)

	if ws.rangeIsMarginal(&add.rangeMin, &add.rangeMax) {
		ws.poolQueue.Put(add)
	} else {
		ws.queue.Push(add)
	}
}

func (ws *WeightSearcher2Fast) search2MakeFollowupShrinkExceptAxis(bound *weightSearch2FastBound, axis int) {
	add := ws.poolQueue.Get()
	add.divideAxis = -1
	add.nodeDepth = bound.nodeDepth + 1

	ws.sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_search2_probeA, &add.rangeMin)
	add.rangeMin[axis] = bound.rangeMin[axis]
	ws.sliceInterpolate(&bound.rangeMin, &bound.rangeMax, c_search2_probeB, &add.rangeMax)
	add.rangeMax[axis] = bound.rangeMax[axis]

	if ws.rangeIsMarginal(&add.rangeMin, &add.rangeMax) {
		ws.poolQueue.Put(add)
	} else {
		ws.queue.Push(add)
	}
}

func (ws *WeightSearcher2Fast) search2MakeFollowupCommon2(probes []*weightSearch2FastProbe, bound *weightSearch2FastBound) {
	add := ws.poolQueue.Get()
	add.divideAxis = -1
	add.nodeDepth = bound.nodeDepth + 1
	add.rangeMin = probes[0].point
	add.rangeMax = probes[0].point

	rangeMin := &add.rangeMin
	rangeMax := &add.rangeMin
	for i := 1; i < len(probes); i++ {
		point := probes[i].point
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
			rangeMin[axis] -= c_search2_probeA * oldExtent
			rangeMax[axis] += c_search2_probeA * oldExtent
		}
	}

	if (bound.rangeMin == add.rangeMin && bound.rangeMax == add.rangeMin) || ws.rangeIsMarginal(&add.rangeMin, &add.rangeMax) {
		ws.poolQueue.Put(add)
	} else {
		ws.queue.Push(add)
	}
}

func checkRangeIsSubrangeOfFast(outer, inner *weightSearch2FastBound) bool {
	for i := range outer.rangeMin {
		if outer.rangeMin[i] <= inner.rangeMin[i] && inner.rangeMax[i] <= outer.rangeMax[i] {
			// yes
		} else {
			return false
		}
	}
	return true
}

func (ws *WeightSearcher2Fast) rangeIsMarginal(rangeMin *weightSearch2FastPoint, rangeMax *weightSearch2FastPoint) bool {
	for i := range ws.typeCount {
		if rangeMax[i]-rangeMin[i] > c_search2_marginalWeightGap {
			return false
		}
	}
	return true
}

func (ws *WeightSearcher2Fast) largeAccuracyGap(a, b float64) bool {
	return math.Abs(a-b) >= c_search2_largeAccuracyGap
}

func (ws *WeightSearcher2Fast) equalAccuracyGap(a, b float64) bool {
	return math.Abs(a-b) < c_search2_equalAccuracyGap
}

func (ws *WeightSearcher2Fast) sliceInterpolate(rangeMin *weightSearch2FastPoint, rangeMax *weightSearch2FastPoint, ratio float64, out *weightSearch2FastPoint) {
	for i := range ws.typeCount {
		out[i] = valueInterpolate(rangeMin[i], rangeMax[i], ratio)
	}
}

func (ws *WeightSearcher2Fast) queueMaintenance() {
	removed := 0
	content := ws.queue.ExportAsSlice()
	for a := range content {
		for b := a + 1; b < len(content); b++ {
			if checkRangeIsSubrangeOfFast(content[a], content[b]) {
				content = removeIndexFast(content, b)
				removed++
			} else if checkRangeIsSubrangeOfFast(content[b], content[a]) {
				content = removeIndexFast(content, a)
				removed++
			}
		}
	}
	ws.queue.ResetFromSlice(content)
}

func removeIndexFast(slice []*weightSearch2FastBound, index int) []*weightSearch2FastBound {
	return slices.Delete(slice, index, index+1)
}
