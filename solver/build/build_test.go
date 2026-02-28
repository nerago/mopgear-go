package build

import (
	"paladin_gearing_go/utiltest"
	"testing"
)

const testingThreadCount = 2

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

// //////////////////////////////////////////////////
func TestRandomStandardRun(t *testing.T) {
	const targetCount = utiltest.TargetCountStandard

	peekRecord := utiltest.PeekTestRecorder{}
	options, model := utiltest.MakeTestOptions()

	evaluateRandom(options, model, targetCount, testingThreadCount, peekRecord.Add)

	utiltest.VerifyRecord(t, &peekRecord, options, targetCount)
}

func TestRandomMinimalRun(t *testing.T) {
	const targetCount = utiltest.TargetCountMinimal

	peekRecord := utiltest.PeekTestRecorder{}
	options, model := utiltest.MakeTestOptions()

	evaluateRandom(options, model, targetCount, testingThreadCount, peekRecord.Add)

	utiltest.VerifyRecord(t, &peekRecord, options, targetCount)
}

func TestRandomFullRun(t *testing.T) {
	const targetCount = utiltest.TargetCountFull

	peekRecord := utiltest.PeekTestRecorder{}
	options, model := utiltest.MakeTestOptions()

	evaluateRandom(options, model, targetCount, testingThreadCount, peekRecord.Add)

	utiltest.VerifyRecord(t, &peekRecord, options, targetCount)
}

// //////////////////////////////////////////////////
func TestOverflowStandardRun(t *testing.T) {
	const targetCount = utiltest.TargetCountStandard

	peekRecord := utiltest.PeekTestRecorder{}
	options, model := utiltest.MakeTestOptions()

	evaluateOverflow(options, model, targetCount, true, testingThreadCount, peekRecord.Add)

	utiltest.VerifyRecord(t, &peekRecord, options, targetCount)
}

func TestOverflowMinimalRun(t *testing.T) {
	const targetCount = utiltest.TargetCountMinimal + 3 // NOTE fudge factor otherwise doesn't hit

	peekRecord := utiltest.PeekTestRecorder{}
	options, model := utiltest.MakeTestOptions()

	evaluateOverflow(options, model, targetCount, true, testingThreadCount, peekRecord.Add)

	utiltest.VerifyRecord(t, &peekRecord, options, targetCount)
}

func TestOverflowFullRun(t *testing.T) {
	const targetCount = utiltest.TargetCountFull

	peekRecord := utiltest.PeekTestRecorder{}
	options, model := utiltest.MakeTestOptions()

	evaluateOverflow(options, model, targetCount, true, testingThreadCount, peekRecord.Add)

	utiltest.VerifyRecord(t, &peekRecord, options, targetCount)
}
