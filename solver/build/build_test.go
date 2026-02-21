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

const testingThreadCount = 1

func TestPeriodicLiteStandardRun(t *testing.T) {
	const targetCount = utiltest.TargetCountStandard

	peekRecord := utiltest.PeekTestRecorder{}
	options, model := utiltest.MakeTestOptions()

	evaluatePeriodic(options, model, targetCount, testingThreadCount, peekRecord.Add)

	utiltest.VerifyRecord(t, &peekRecord, options, targetCount)
}

func TestPeriodicLiteMinimalRun(t *testing.T) {
	const targetCount = utiltest.TargetCountMinimal

	peekRecord := utiltest.PeekTestRecorder{}
	options, model := utiltest.MakeTestOptions()

	evaluatePeriodic(options, model, targetCount, testingThreadCount, peekRecord.Add)

	utiltest.VerifyRecord(t, &peekRecord, options, targetCount)
}

func TestPeriodicLiteFullRun(t *testing.T) {
	const targetCount = utiltest.TargetCountFull

	peekRecord := utiltest.PeekTestRecorder{}
	options, model := utiltest.MakeTestOptions()

	evaluatePeriodic(options, model, targetCount, testingThreadCount, peekRecord.Add)

	utiltest.VerifyRecord(t, &peekRecord, options, targetCount)
}

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
