package core

import (
	"fmt"
	"time"

	"github.com/wowsims/mop/sim/core/proto"
)

type APLValueCurrentTime struct {
	DefaultAPLValueImpl
}

func (rot *APLRotation) newValueCurrentTime(config *proto.APLValueCurrentTime, _ *proto.UUID) APLValue {
	return &APLValueCurrentTime{}
}
func (value *APLValueCurrentTime) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueCurrentTime) GetDuration(sim *Simulation) time.Duration {
	return sim.CurrentTime
}
func (value *APLValueCurrentTime) String() string {
	return "Current Time"
}
func (value *APLValueCurrentTime) Equals(other APLValue) bool {
	if _, isType := other.(*APLValueCurrentTime); isType {
		return true
	}
	return false
}

type APLValueCurrentTimePercent struct {
	DefaultAPLValueImpl
}

func (rot *APLRotation) newValueCurrentTimePercent(config *proto.APLValueCurrentTimePercent, _ *proto.UUID) APLValue {
	return &APLValueCurrentTimePercent{}
}
func (value *APLValueCurrentTimePercent) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueCurrentTimePercent) GetFloat(sim *Simulation) float64 {
	return sim.CurrentTime.Seconds() / sim.Duration.Seconds()
}
func (value *APLValueCurrentTimePercent) String() string {
	return fmt.Sprintf("Current Time %%")
}
func (value *APLValueCurrentTimePercent) Equals(other APLValue) bool {
	if _, isType := other.(*APLValueCurrentTimePercent); isType {
		return true
	}
	return false
}

type APLValueRemainingTime struct {
	DefaultAPLValueImpl
}

func (rot *APLRotation) newValueRemainingTime(config *proto.APLValueRemainingTime, _ *proto.UUID) APLValue {
	return &APLValueRemainingTime{}
}
func (value *APLValueRemainingTime) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueRemainingTime) GetDuration(sim *Simulation) time.Duration {
	return sim.GetRemainingDuration()
}
func (value *APLValueRemainingTime) String() string {
	return "Remaining Time"
}
func (value *APLValueRemainingTime) Equals(other APLValue) bool {
	if _, isType := other.(*APLValueRemainingTime); isType {
		return true
	}
	return false
}

type APLValueRemainingTimePercent struct {
	DefaultAPLValueImpl
}

func (rot *APLRotation) newValueRemainingTimePercent(config *proto.APLValueRemainingTimePercent, _ *proto.UUID) APLValue {
	return &APLValueRemainingTimePercent{}
}
func (value *APLValueRemainingTimePercent) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeFloat
}
func (value *APLValueRemainingTimePercent) GetFloat(sim *Simulation) float64 {
	return sim.GetRemainingDurationPercent()
}
func (value *APLValueRemainingTimePercent) String() string {
	return fmt.Sprintf("Remaining Time %%")
}
func (value *APLValueRemainingTimePercent) Equals(other APLValue) bool {
	if _, isType := other.(*APLValueRemainingTimePercent); isType {
		return true
	}
	return false
}

type APLValueNumberTargets struct {
	DefaultAPLValueImpl
}

func (rot *APLRotation) newValueNumberTargets(config *proto.APLValueNumberTargets, _ *proto.UUID) APLValue {
	return &APLValueNumberTargets{}
}
func (value *APLValueNumberTargets) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeInt
}
func (value *APLValueNumberTargets) GetInt(sim *Simulation) int32 {
	return sim.ActiveTargetCount()
}
func (value *APLValueNumberTargets) GetFloat(sim *Simulation) float64 {
	return float64(sim.ActiveTargetCount())
}
func (value *APLValueNumberTargets) String() string {
	return "Num Active Targets"
}
func (value *APLValueNumberTargets) Equals(other APLValue) bool {
	if _, isType := other.(*APLValueNumberTargets); isType {
		return true
	}
	return false
}

type APLValueIsExecutePhase struct {
	DefaultAPLValueImpl
	threshold proto.APLValueIsExecutePhase_ExecutePhaseThreshold
}

func (rot *APLRotation) newValueIsExecutePhase(config *proto.APLValueIsExecutePhase, _ *proto.UUID) APLValue {
	if config.Threshold == proto.APLValueIsExecutePhase_Unknown {
		return nil
	}
	return &APLValueIsExecutePhase{
		threshold: config.Threshold,
	}
}
func (value *APLValueIsExecutePhase) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueIsExecutePhase) GetBool(sim *Simulation) bool {
	if value.threshold == proto.APLValueIsExecutePhase_E20 {
		return sim.IsExecutePhase20()
	} else if value.threshold == proto.APLValueIsExecutePhase_E25 {
		return sim.IsExecutePhase25()
	} else if value.threshold == proto.APLValueIsExecutePhase_E35 {
		return sim.IsExecutePhase35()
	} else if value.threshold == proto.APLValueIsExecutePhase_E45 {
		return sim.IsExecutePhase45()
	} else if value.threshold == proto.APLValueIsExecutePhase_E90 {
		return sim.IsExecutePhase90()
	} else {
		panic("Should never reach here")
	}
}
func (value *APLValueIsExecutePhase) String() string {
	return "Is Execute Phase"
}
func (value *APLValueIsExecutePhase) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueIsExecutePhase); isType {
		return value.threshold == otherValue.threshold
	}
	return false
}
