package weightfind

import (
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/weightfind"
	"github.com/nerago/mopgear-go/weightfind/weight_types"
)

type SearcherModel interface {

}

type ModelWeightSearch struct {
	statTypes               []stats.StatType
	simTypes                []stats.SimType
	targetRatio             weight_types.SimPriorityBasic
	initialEvaluateAccuracy weightfind.EvaluateAccuracyPrepared
}
