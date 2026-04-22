package core

import (
	"fmt"
	"time"

	"github.com/wowsims/mop/sim/core/proto"
)

type APLValueBossSpellIsCasting struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueBossSpellIsCasting(config *proto.APLValueBossSpellIsCasting, _ *proto.UUID) APLValue {
	spell := rot.GetTargetAPLSpell(config.SpellId, rot.GetTargetUnit(config.TargetUnit))
	if spell == nil {
		return nil
	}
	return &APLValueBossSpellIsCasting{
		spell: spell,
	}
}
func (value *APLValueBossSpellIsCasting) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueBossSpellIsCasting) GetBool(sim *Simulation) bool {
	return value.spell.Unit.Hardcast.ActionID == value.spell.ActionID && value.spell.Unit.Hardcast.Expires > sim.CurrentTime
}
func (value *APLValueBossSpellIsCasting) String() string {
	return fmt.Sprintf("Boss is Casting(%s)", value.spell.ActionID)
}
func (value *APLValueBossSpellIsCasting) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBossSpellIsCasting); isType {
		return value.spell.EqualForAPL(otherValue.spell)
	}
	return false
}

type APLValueBossSpellTimeToReady struct {
	DefaultAPLValueImpl
	spell *Spell
}

func (rot *APLRotation) newValueBossSpellTimeToReady(config *proto.APLValueBossSpellTimeToReady, _ *proto.UUID) APLValue {
	spell := rot.GetTargetAPLSpell(config.SpellId, rot.GetTargetUnit(config.TargetUnit))
	if spell == nil {
		return nil
	}
	return &APLValueBossSpellTimeToReady{
		spell: spell,
	}
}
func (value *APLValueBossSpellTimeToReady) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeDuration
}
func (value *APLValueBossSpellTimeToReady) GetDuration(sim *Simulation) time.Duration {
	return value.spell.TimeToReady(sim)
}
func (value *APLValueBossSpellTimeToReady) String() string {
	return fmt.Sprintf("Boss Spell Time to Ready(%s)", value.spell.ActionID)
}
func (value *APLValueBossSpellTimeToReady) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBossSpellTimeToReady); isType {
		return value.spell.EqualForAPL(otherValue.spell)
	}
	return false
}


type APLValueBossCurrentTarget struct {
	DefaultAPLValueImpl
	player *Unit
	target UnitReference
}

func (rot *APLRotation) newValueBossCurrentTarget(config *proto.APLValueBossCurrentTarget, _ *proto.UUID) APLValue {
	unit := rot.GetSourceUnit(config.TargetUnit)
	if unit.Get() == nil {
		return nil
	}
	return &APLValueBossCurrentTarget{
		player: rot.unit,
		target: unit,
	}
}
func (value *APLValueBossCurrentTarget) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueBossCurrentTarget) GetBool(sim *Simulation) bool {
	return value.target.Get().CurrentTarget == value.player
}
func (value *APLValueBossCurrentTarget) String() string {
	return fmt.Sprintf("IsTanking(%s)", value.target.Get().Label)
}
func (value *APLValueBossCurrentTarget) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBossCurrentTarget); isType {
		return value.player.EqualForAPL(otherValue.player) && value.target.EqualForAPL(otherValue.target)
	}
	return false
}
