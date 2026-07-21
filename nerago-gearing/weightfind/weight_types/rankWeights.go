package weight_types

import (
	"paladin_gearing_go/util/util_highs"
)

type IRankEntryHasCommon interface {
	ToCommon() *RankEntryCommon
}

type RankEntryCommon struct {
	Data *WeightInput

	SimScore   float64
	TargetRank int
}

func (rc RankEntryCommon) ToCommon() *RankEntryCommon {
	return &rc
}

type RankEntry struct {
	RankEntryCommon

	ScoreColumn util_highs.ColumnIndex
	RankColumn  util_highs.ColumnIndex
}

type RankEntry4 struct {
	RankEntryCommon

	InitialStatScore float64

	ScoreColumn util_highs.ColumnIndex
	RankColumn  util_highs.ColumnIndex
}

type RankEntry5 struct {
	RankEntryCommon

	ScoreCompute    util_highs.ColumnIndex
	ScoreIfIncluded util_highs.ColumnIndex
	IsInclude       util_highs.ColumnIndex
}

type RankEntry3 struct {
	RankEntryCommon

	InitialStatScore float64

	ScoreColumn       util_highs.ColumnIndex
	RankColumn        util_highs.ColumnIndex
	RankDiffAbsColumn util_highs.ColumnIndex
}

//func passToHasCommon(re *RankEntry4)float64 {
//	return somethingOnFaceHasCommon(re)
//}
//func passToHasCommon2(re RankEntry4)float64 {
//	return somethingOnFaceHasCommon(re)
//}
//func somethingOnFaceHasCommon[T IRankEntryHasCommon](common T) float64 {
//	return common.toCommon().SimScore
//}
//func somethingOnFaceHasCommonJust(common IRankEntryHasCommon) float64 {
//	return common.toCommon().SimScore
//}
