package weight_types

import (
	"paladin_gearing_go/util/util_highs"
)

type RankEntry4 struct {
	Data *WeightInput

	InitialStatScore float64
	SimScore         float64
	TargetRank       int

	ScoreColumn util_highs.ColumnIndex
	RankColumn  util_highs.ColumnIndex
}

type RankEntry5 struct {
	Data *WeightInput

	SimScore   float64
	TargetRank int

	ScoreCompute    util_highs.ColumnIndex
	ScoreIfIncluded util_highs.ColumnIndex
	IsInclude       util_highs.ColumnIndex
}
