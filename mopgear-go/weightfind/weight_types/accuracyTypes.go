package weight_types

import (
	"iter"

	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util/util_collection"
)

type AccuracyInfo struct {
	StatScore     float64
	SimScore      float64
	StatRankRange *util_collection.HiLoInt
	SimRankRange  *util_collection.HiLoInt
	DataSim       *stats.SimData
}

type AccuracyInfoExtended struct {
	AccuracyInfo
	AccuracyExtension
}

type AccuracyExtension struct {
	SimRankByType stats.SimTypeMap[util_collection.HiLoFloat]
}

func (a *AccuracyInfo) GetSimData() *stats.SimData {
	return a.DataSim
}

func (a *AccuracyInfo) GetSimScore() float64 {
	return a.SimScore
}

func (a *AccuracyInfo) IncrementSimScore(add float64) {
	a.SimScore += add
}

func (a *AccuracyInfo) ResetSimScore() {
	a.SimScore = 0
}

func (a *AccuracyInfo) SetSimRankRange(targetRange *util_collection.HiLoInt) {
	a.SimRankRange = targetRange
}

func (a *AccuracyInfo) GetSimRankRange() *util_collection.HiLoInt {
	return a.SimRankRange
}

func (a *AccuracyInfo) GetStatScore() float64 {
	return a.StatScore
}

func (a *AccuracyInfo) SetStatRankRange(targetRange *util_collection.HiLoInt) {
	a.StatRankRange = targetRange
}

func (a *AccuracyInfo) GetStatRankRange() *util_collection.HiLoInt {
	return a.StatRankRange
}

func (a *AccuracyExtension) GetSimRankRangeFloatByType(simType stats.SimType) util_collection.HiLoFloat {
	return a.SimRankByType.GetOrNilValue(simType)
}

func (a *AccuracyExtension) SeqSimRankRangeFloatByType() iter.Seq2[stats.SimType, util_collection.HiLoFloat] {
	return a.SimRankByType.SeqKeyValue()
}

func (a *AccuracyExtension) SetSimRankRangeByType(simType stats.SimType, lo int, hi int) {
	a.SimRankByType.Put(simType, util_collection.HiLoFloat{Lo: float64(lo), Hi: float64(hi)})
}

func (a *AccuracyExtension) SetSimRankRangeFloatByType(simType stats.SimType, lo float64, hi float64) {
	a.SimRankByType.Put(simType, util_collection.HiLoFloat{Lo: lo, Hi: hi})
}

type AccuracyInfoPrePrepare struct {
	SimScore float64
	DataSim  *stats.SimData
	DataStat *stats.StatBlock
}

type AccuracyInfoPrePrepareExtended struct {
	AccuracyInfoPrePrepare
	AccuracyExtension
}

func (a *AccuracyInfoPrePrepare) GetStatData() *stats.StatBlock {
	return a.DataStat
}

func (a *AccuracyInfoPrePrepare) GetSimData() *stats.SimData {
	return a.DataSim
}

func (a *AccuracyInfoPrePrepare) GetSimScore() float64 {
	return a.SimScore
}

func (a *AccuracyInfoPrePrepare) IncrementSimScore(add float64) {
	a.SimScore += add
}

func (a *AccuracyInfoPrePrepare) ResetSimScore() {
	a.SimScore = 0
}

type AccuracyInfoPrepared struct {
	StatScore     float64
	Stats         *stats.StatBlock
	SimRankRange  *util_collection.HiLoInt
	StatRankRange *util_collection.HiLoInt
	//Prep          *AccuracyInfoPrePrepare
}

func (a *AccuracyInfoPrepared) GetStatScore() float64 {
	return a.StatScore
}

//func (a *AccuracyInfoPrepared) GetSimScore() float64 {
//	return a.Prep.SimScore
//}
//
//func (a *AccuracyInfoPrepared) GetSimData() *stats.SimData {
//	return a.Prep.DataSim
//}

func (a *AccuracyInfoPrepared) GetSimRankRange() *util_collection.HiLoInt {
	return a.SimRankRange
}

func (a *AccuracyInfoPrepared) SetSimRankRange(targetRange *util_collection.HiLoInt) {
	a.SimRankRange = targetRange
}

func (a *AccuracyInfoPrepared) GetStatRankRange() *util_collection.HiLoInt {
	return a.StatRankRange
}

func (a *AccuracyInfoPrepared) SetStatRankRange(targetRange *util_collection.HiLoInt) {
	a.StatRankRange = targetRange
}
