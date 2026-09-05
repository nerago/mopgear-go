package weight_types

import (
	"github.com/nerago/mopgear-go/stats"
)

type WeightType int

type IWeight interface {
	IsEmpty() bool
	String() string
	FinishAndValidate(sampleInputs []WeightInput) error
	CalcStatScore(*stats.StatBlock) float64
	CalcStatScoreRaw(*stats.StatBlock) float64
	CalcStatScoreWithBonus(*stats.StatBlock, *stats.SimTypeMap[float64]) float64
}
