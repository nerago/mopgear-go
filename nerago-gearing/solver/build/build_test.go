package build

import (
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_test"
	"testing"
)

const testingThreadCount = 2

// //////////////////////////////////////////////////
func TestRandomStandardRun(t *testing.T) {
	const targetCount = util_test.TargetCountStandard

	peekRecord := util_test.PeekTestRecorder{}
	options, model := util_test.MakeTestOptions()

	evaluateRandom(options, model, targetCount, util.TrackProgress_Nop(), testingThreadCount, peekRecord.Add)

	util_test.VerifyRecord(t, &peekRecord, options, targetCount)
}

func TestRandomMinimalRun(t *testing.T) {
	const targetCount = util_test.TargetCountMinimal

	peekRecord := util_test.PeekTestRecorder{}
	options, model := util_test.MakeTestOptions()

	evaluateRandom(options, model, targetCount, util.TrackProgress_Nop(), testingThreadCount, peekRecord.Add)

	util_test.VerifyRecord(t, &peekRecord, options, targetCount)
}

func TestRandomFullRun(t *testing.T) {
	const targetCount = util_test.TargetCountFull

	peekRecord := util_test.PeekTestRecorder{}
	options, model := util_test.MakeTestOptions()

	evaluateRandom(options, model, targetCount, util.TrackProgress_Nop(), testingThreadCount, peekRecord.Add)

	util_test.VerifyRecord(t, &peekRecord, options, targetCount)
}

// //////////////////////////////////////////////////
func TestOverflowStandardRun(t *testing.T) {
	const targetCount = util_test.TargetCountStandard

	peekRecord := util_test.PeekTestRecorder{}
	options, model := util_test.MakeTestOptions()

	evaluateOverflow2(options, model, targetCount, util.TrackProgress_Nop(), testingThreadCount, peekRecord.Add)

	util_test.VerifyRecord(t, &peekRecord, options, targetCount)
}

func TestOverflowMinimalRun(t *testing.T) {
	const targetCount = util_test.TargetCountMinimal + 3 // NOTE fudge factor otherwise doesn't hit

	peekRecord := util_test.PeekTestRecorder{}
	options, model := util_test.MakeTestOptions()

	evaluateOverflow2(options, model, targetCount, util.TrackProgress_Nop(), testingThreadCount, peekRecord.Add)

	util_test.VerifyRecord(t, &peekRecord, options, targetCount)
}

func TestOverflowFullRun(t *testing.T) {
	const targetCount = util_test.TargetCountFull

	peekRecord := util_test.PeekTestRecorder{}
	options, model := util_test.MakeTestOptions()

	evaluateOverflow2(options, model, targetCount, util.TrackProgress_Nop(), testingThreadCount, peekRecord.Add)

	util_test.VerifyRecord(t, &peekRecord, options, targetCount)
}

// //////////////////////////////////////////////////
func TestFullFullRun(t *testing.T) {
	peekRecord := util_test.PeekTestRecorder{}
	options, model := util_test.MakeTestOptions()

	evaluateFull(options, model, util.TrackProgress_Nop(), util.PrintRecorder_HoldAll(), peekRecord.Add)

	targetCount := options.TotalCombinationCountAsInt()
	util_test.VerifyRecord(t, &peekRecord, options, int(targetCount))
}
