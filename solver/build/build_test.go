package build

import (
	"paladin_gearing_go/types/items"
	"paladin_gearing_go/utiltest"
	"testing"
)

// func TestAll(t *testing.T) {
// 	peekRecord := utiltest.PeekTestRecorder{}
// 	options, _ := utiltest.MakeTestOptions()

// 	for itemSet := range allSetsChannel(options) {
// 		peekRecord.Add(&itemSet)
// 	}

// 	utiltest.VerifyRecord(t, &peekRecord, options, targetCount)
// }

func TestPeriodicStandardRun(t *testing.T) {
	const targetCount = utiltest.TargetCountStandard

	peekRecord := utiltest.PeekTestRecorder{}
	options, _ := utiltest.MakeTestOptions()

	setChannel := periodicSetsChannel(options)
	loopNSets(setChannel, &peekRecord, targetCount)

	utiltest.VerifyRecord(t, &peekRecord, options, targetCount)
}

func TestPeriodicMinimalRun(t *testing.T) {
	const targetCount = utiltest.TargetCountMinimal

	peekRecord := utiltest.PeekTestRecorder{}
	options, _ := utiltest.MakeTestOptions()

	setChannel := periodicSetsChannel(options)
	loopNSets(setChannel, &peekRecord, targetCount)

	utiltest.VerifyRecord(t, &peekRecord, options, targetCount)
}

func TestPeriodicFullRun(t *testing.T) {
	const targetCount = utiltest.TargetCountFull

	peekRecord := utiltest.PeekTestRecorder{}
	options, _ := utiltest.MakeTestOptions()

	setChannel := periodicSetsChannel(options)
	loopNSets(setChannel, &peekRecord, targetCount)

	utiltest.VerifyRecord(t, &peekRecord, options, targetCount)
}

func loopNSets(setChannel <-chan items.SolvableItemSet, peekRecord *utiltest.PeekTestRecorder, iterCount int) {
	for range iterCount {
		itemSet, ok := <-setChannel
		if !ok {
			panic("empty channel")
		}
		peekRecord.Add(&itemSet)
	}
}
