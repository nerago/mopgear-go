package channel

import (
	"paladin_gearing_go/items"
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

const testingThreadCount = 6

func TestPeriodicChannelStandardRun(t *testing.T) {
	const targetCount = utiltest.TargetCountStandard

	peekRecord := utiltest.PeekTestRecorder{}
	options, _ := utiltest.MakeTestOptions()

	setChannel := periodicSetsChannel(options)
	loopNSets(setChannel, &peekRecord, targetCount)

	utiltest.VerifyRecord(t, &peekRecord, options, targetCount)
}

func TestPeriodicChannelMinimalRun(t *testing.T) {
	const targetCount = utiltest.TargetCountMinimal

	peekRecord := utiltest.PeekTestRecorder{}
	options, _ := utiltest.MakeTestOptions()

	setChannel := periodicSetsChannel(options)
	loopNSets(setChannel, &peekRecord, targetCount)

	utiltest.VerifyRecord(t, &peekRecord, options, targetCount)
}

func TestPeriodicChannelFullRun(t *testing.T) {
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
