package build

import (
	"paladin_gearing_go/util"
	"paladin_gearing_go/utiltest"
	"testing"
)

const testingThreadCount = 2

// //////////////////////////////////////////////////
func TestRandomStandardRun(t *testing.T) {
	const targetCount = utiltest.TargetCountStandard

	peekRecord := utiltest.PeekTestRecorder{}
	options, model := utiltest.MakeTestOptions()

	evaluateRandom(options, model, targetCount, util.TrackProgress_Nop(), testingThreadCount, peekRecord.Add)

	utiltest.VerifyRecord(t, &peekRecord, options, targetCount)
}

func TestRandomMinimalRun(t *testing.T) {
	const targetCount = utiltest.TargetCountMinimal

	peekRecord := utiltest.PeekTestRecorder{}
	options, model := utiltest.MakeTestOptions()

	evaluateRandom(options, model, targetCount, util.TrackProgress_Nop(), testingThreadCount, peekRecord.Add)

	utiltest.VerifyRecord(t, &peekRecord, options, targetCount)
}

func TestRandomFullRun(t *testing.T) {
	const targetCount = utiltest.TargetCountFull

	peekRecord := utiltest.PeekTestRecorder{}
	options, model := utiltest.MakeTestOptions()

	evaluateRandom(options, model, targetCount, util.TrackProgress_Nop(), testingThreadCount, peekRecord.Add)

	utiltest.VerifyRecord(t, &peekRecord, options, targetCount)
}

// //////////////////////////////////////////////////
func TestOverflowStandardRun(t *testing.T) {
	const targetCount = utiltest.TargetCountStandard

	peekRecord := utiltest.PeekTestRecorder{}
	options, model := utiltest.MakeTestOptions()

	evaluateOverflow2(options, model, targetCount, util.TrackProgress_Nop(), testingThreadCount, peekRecord.Add)

	utiltest.VerifyRecord(t, &peekRecord, options, targetCount)
}

func TestOverflowMinimalRun(t *testing.T) {
	const targetCount = utiltest.TargetCountMinimal + 3 // NOTE fudge factor otherwise doesn't hit

	peekRecord := utiltest.PeekTestRecorder{}
	options, model := utiltest.MakeTestOptions()

	evaluateOverflow2(options, model, targetCount, util.TrackProgress_Nop(), testingThreadCount, peekRecord.Add)

	utiltest.VerifyRecord(t, &peekRecord, options, targetCount)
}

func TestOverflowFullRun(t *testing.T) {
	const targetCount = utiltest.TargetCountFull

	peekRecord := utiltest.PeekTestRecorder{}
	options, model := utiltest.MakeTestOptions()

	evaluateOverflow2(options, model, targetCount, util.TrackProgress_Nop(), testingThreadCount, peekRecord.Add)

	utiltest.VerifyRecord(t, &peekRecord, options, targetCount)
}

// //////////////////////////////////////////////////
func TestFullFullRun(t *testing.T) {
	peekRecord := utiltest.PeekTestRecorder{}
	options, model := utiltest.MakeTestOptions()

	evaluateFull(options, model, util.TrackProgress_Nop(), util.PrintRecorder_HoldAll(), peekRecord.Add)

	targetCount := options.TotalCombinationCountAsInt()
	utiltest.VerifyRecord(t, &peekRecord, options, int(targetCount))
}
