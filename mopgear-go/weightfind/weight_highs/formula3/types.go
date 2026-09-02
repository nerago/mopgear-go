package formula3

import (
	"math"
	"time"

	"github.com/bartolsthoorn/gohighs/highs"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type sectionResult struct {
	weights        weight_types.Weight2Extended
	bounds         weight_types.Weight4SegmentBound
	includePercent float64
	elapsed        time.Duration
	status         highs.ModelStatus
	err            error
}

type statStatus int8

const (
	statusUndecided          statStatus = iota
	statusFullRange          statStatus = iota
	statusPrimaryBreakpoints statStatus = iota
	statusSecondary          statStatus = iota
)

type statInfo struct {
	statType    stats.StatType
	status      statStatus
	sections    []*sectionResult
	usedRange   weight_types.StatRange
	usedPercent float64
	hiPercent   float64
	loPercent   float64
	//nested stuff?
}

func boundsSingleLessThan(statType stats.StatType, maximum uint32) *weight_types.Weight4SegmentBound {
	bound := &weight_types.Weight4SegmentBound{}
	bound.Put(statType, weight_types.StatRange{Minimum: 0, Maximum: maximum - 1})
	return bound
}

func boundsSingleGreaterThan(statType stats.StatType, minimum uint32) *weight_types.Weight4SegmentBound {
	bound := &weight_types.Weight4SegmentBound{}
	bound.Put(statType, weight_types.StatRange{Minimum: minimum + 1, Maximum: math.MaxUint32})
	return bound
}
