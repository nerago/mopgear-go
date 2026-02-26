package indexed

import (
	"math/big"
	"paladin_gearing_go/items"
	"paladin_gearing_go/model"
	"paladin_gearing_go/utiltest"
	"testing"
)

func setup(targetCount uint64) (*utiltest.PeekTestRecorder, *items.SolvableOptionsMap, *model.Model, *big.Int, *big.Int) {
	peekRecord := utiltest.PeekTestRecorder{}
	options, model := utiltest.MakeTestOptions()
	max, skip := initSkipValues(options, targetCount)
	return &peekRecord, options, model, max, skip
}

func TestMultiInt(t *testing.T) {
	const targetCount = utiltest.TargetCountStandard
	peekRecord, options, model, max, skip := setup(targetCount)
	_ = mainLoop_multiThread_int(options, max.Uint64(), skip.Uint64(), model, peekRecord.Add)
	utiltest.VerifyRecord(t, peekRecord, options, targetCount)
}

func TestMultiBig(t *testing.T) {
	const targetCount = utiltest.TargetCountStandard
	peekRecord, options, model, max, skip := setup(targetCount)
	mainLoop_multiThread_big(options, max, skip, model, peekRecord.Add)
	utiltest.VerifyRecord(t, peekRecord, options, targetCount)
}

func TestSingleInt(t *testing.T) {
	const targetCount = utiltest.TargetCountStandard
	peekRecord, options, model, max, skip := setup(targetCount)
	mainLoop_singleThread_int(options, max.Uint64(), skip.Uint64(), model, peekRecord.Add)
	utiltest.VerifyRecord(t, peekRecord, options, targetCount)
}

func TestSingleBig(t *testing.T) {
	const targetCount = utiltest.TargetCountStandard
	peekRecord, options, model, max, skip := setup(targetCount)
	mainLoop_singleThread_big(options, max, skip, model, peekRecord.Add)
	utiltest.VerifyRecord(t, peekRecord, options, targetCount)
}

// indexed is pretty bad at hitting each item, needs more time
const minimalFudge = 7

func TestMultiIntMinimal(t *testing.T) {
	const targetCount = utiltest.TargetCountMinimal + minimalFudge
	peekRecord, options, model, max, skip := setup(targetCount)
	_ = mainLoop_multiThread_int(options, max.Uint64(), skip.Uint64(), model, peekRecord.Add)
	utiltest.VerifyRecord(t, peekRecord, options, targetCount)
}

func TestMultiBigMinimal(t *testing.T) {
	const targetCount = utiltest.TargetCountMinimal + minimalFudge
	peekRecord, options, model, max, skip := setup(targetCount)
	mainLoop_multiThread_big(options, max, skip, model, peekRecord.Add)
	utiltest.VerifyRecord(t, peekRecord, options, targetCount)
}

func TestSingleIntMinimal(t *testing.T) {
	const targetCount = utiltest.TargetCountMinimal + minimalFudge
	peekRecord, options, model, max, skip := setup(targetCount)
	mainLoop_singleThread_int(options, max.Uint64(), skip.Uint64(), model, peekRecord.Add)
	utiltest.VerifyRecord(t, peekRecord, options, targetCount)
}

func TestSingleBigMinimal(t *testing.T) {
	const targetCount = utiltest.TargetCountMinimal + minimalFudge
	peekRecord, options, model, max, skip := setup(targetCount)
	mainLoop_singleThread_big(options, max, skip, model, peekRecord.Add)
	utiltest.VerifyRecord(t, peekRecord, options, targetCount)
}

func TestMultiIntFull(t *testing.T) {
	const targetCount = utiltest.TargetCountFull
	peekRecord, options, model, max, skip := setup(targetCount)
	_ = mainLoop_multiThread_int(options, max.Uint64(), skip.Uint64(), model, peekRecord.Add)
	utiltest.VerifyRecord(t, peekRecord, options, targetCount)
}

func TestMultiBigFull(t *testing.T) {
	const targetCount = utiltest.TargetCountFull
	peekRecord, options, model, max, skip := setup(targetCount)
	mainLoop_multiThread_big(options, max, skip, model, peekRecord.Add)
	utiltest.VerifyRecord(t, peekRecord, options, targetCount)
}

func TestSingleIntFull(t *testing.T) {
	const targetCount = utiltest.TargetCountFull
	peekRecord, options, model, max, skip := setup(targetCount)
	mainLoop_singleThread_int(options, max.Uint64(), skip.Uint64(), model, peekRecord.Add)
	utiltest.VerifyRecord(t, peekRecord, options, targetCount)
}

func TestSingleBigFull(t *testing.T) {
	const targetCount = utiltest.TargetCountFull
	peekRecord, options, model, max, skip := setup(targetCount)
	mainLoop_singleThread_big(options, max, skip, model, peekRecord.Add)
	utiltest.VerifyRecord(t, peekRecord, options, targetCount)
}
