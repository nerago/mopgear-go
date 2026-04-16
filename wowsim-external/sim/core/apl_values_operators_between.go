package core

import (
	"fmt"
	"time"

	"github.com/wowsims/mop/sim/core/proto"
)

func makeBetween(and *APLValueAnd, rot *APLRotation) (APLValue, bool) {
	// only support exactly 2 elements (an upper and lower bound)
	if len(and.vals) != 2 {
		return nil, false
	}

	// they should both be compare operations
	cmpA, isCmpA := and.vals[0].(APLValueCompareLike)
	cmpB, isCmpB := and.vals[1].(APLValueCompareLike)
	if !isCmpA || !isCmpB {
		return nil, false
	}

	// extract the compare components
	valueA0, valueA1 := cmpA.GetInnerValues()[0], cmpA.GetInnerValues()[1]
	valueB0, valueB1 := cmpB.GetInnerValues()[0], cmpB.GetInnerValues()[1]

	// make sure we have most optimised versions
	valueA0 = optimizeLogic(valueA0, rot)
	valueA1 = optimizeLogic(valueA1, rot)
	valueB0 = optimizeLogic(valueB0, rot)
	valueB1 = optimizeLogic(valueB1, rot)

	// check if we have the general "between" pattern; need an identical value to check on both sides
	var check, otherA, otherB APLValue
	if valueA0 == valueB0 {
		check = valueA0
		otherA = valueA1
		otherB = valueB1
	} else if valueA1 == valueB1 {
		check = valueA1
		otherA = valueA0
		otherB = valueB0
	} else if valueA0 == valueB1 {
		check = valueA0
		otherA = valueA1
		otherB = valueB0
	} else if valueA1 == valueB0 {
		check = valueA1
		otherA = valueA0
		otherB = valueB1
	} else {
		return nil, false
	}

	// work out which is the high and low bounds (if any)
	var hi, lo APLValue
	var inclusiveLo, inclusiveHi bool
	opA, opB := cmpA.Op(), cmpB.Op()
	if (opA == proto.APLValueCompare_OpLt || opA == proto.APLValueCompare_OpLe) && (opB == proto.APLValueCompare_OpGt || opB == proto.APLValueCompare_OpGe) {
		hi = otherA
		inclusiveHi = (opA == proto.APLValueCompare_OpLe)
		lo = otherB
		inclusiveLo = (opB == proto.APLValueCompare_OpGe)
	} else if (opB == proto.APLValueCompare_OpLt || opB == proto.APLValueCompare_OpLe) && (opA == proto.APLValueCompare_OpGt || opA == proto.APLValueCompare_OpGe) {
		hi = otherB
		inclusiveHi = (opB == proto.APLValueCompare_OpGe)
		lo = otherA
		inclusiveLo = (opA == proto.APLValueCompare_OpLe)
	} else {
		return nil, false
	}

	list := rot.coerceAllToSameType([]APLValue{check, hi, lo}) // what if placeholders?
	check, hi, lo = list[0], list[1], list[2]

	constHi, isConstHi := hi.(*APLValueConst)
	constLo, isConstLo := lo.(*APLValueConst)
	if isConstHi && isConstLo {
		// const value implementations
		switch check.Type() {
		case proto.APLValueType_ValueTypeInt:
			common := APLValueBetween_CommonConstInt{check: check, loValue: constLo.GetInt(nil), hiValue: constHi.GetInt(nil)}
			if inclusiveLo && inclusiveHi {
				return &APLValueBetweenInt_ConstIncInc{common}, true
			} else if inclusiveLo {
				return &APLValueBetweenInt_ConstIncExc{common}, true
			} else if inclusiveHi {
				return &APLValueBetweenInt_ConstExcInc{common}, true
			} else {
				return &APLValueBetweenInt_ConstExcExc{common}, true
			}
		case proto.APLValueType_ValueTypeFloat:
			common := APLValueBetween_CommonConstFloat{check: check, loValue: getConstAPLFloatValue(constLo), hiValue: getConstAPLFloatValue(constHi)}
			if inclusiveLo && inclusiveHi {
				return &APLValueBetweenFloat_ConstIncInc{common}, true
			} else if inclusiveLo {
				return &APLValueBetweenFloat_ConstIncExc{common}, true
			} else if inclusiveHi {
				return &APLValueBetweenFloat_ConstExcInc{common}, true
			} else {
				return &APLValueBetweenFloat_ConstExcExc{common}, true
			}
		case proto.APLValueType_ValueTypeDuration:
			common := APLValueBetween_CommonConstDuration{check: check, loValue: constLo.GetDuration(nil), hiValue: constHi.GetDuration(nil)}
			if inclusiveLo && inclusiveHi {
				return &APLValueBetweenDuration_ConstIncInc{common}, true
			} else if inclusiveLo {
				return &APLValueBetweenDuration_ConstIncExc{common}, true
			} else if inclusiveHi {
				return &APLValueBetweenDuration_ConstExcInc{common}, true
			} else {
				return &APLValueBetweenDuration_ConstExcExc{common}, true
			}
		}
	} else {
		// dynamic value implementations
		common := APLValueBetween_Common{check: check, lo: lo, hi: hi}
		switch check.Type() {
		case proto.APLValueType_ValueTypeInt:
			if inclusiveLo && inclusiveHi {
				return &APLValueBetweenInt_IncInc{common}, true
			} else if inclusiveLo {
				return &APLValueBetweenInt_IncExc{common}, true
			} else if inclusiveHi {
				return &APLValueBetweenInt_ExcInc{common}, true
			} else {
				return &APLValueBetweenInt_ExcExc{common}, true
			}
		case proto.APLValueType_ValueTypeFloat:
			if inclusiveLo && inclusiveHi {
				return &APLValueBetweenFloat_IncInc{common}, true
			} else if inclusiveLo {
				return &APLValueBetweenFloat_IncExc{common}, true
			} else if inclusiveHi {
				return &APLValueBetweenFloat_ExcInc{common}, true
			} else {
				return &APLValueBetweenFloat_ExcExc{common}, true
			}
		case proto.APLValueType_ValueTypeDuration:
			if inclusiveLo && inclusiveHi {
				return &APLValueBetweenDuration_IncInc{common}, true
			} else if inclusiveLo {
				return &APLValueBetweenDuration_IncExc{common}, true
			} else if inclusiveHi {
				return &APLValueBetweenDuration_ExcInc{common}, true
			} else {
				return &APLValueBetweenDuration_ExcExc{common}, true
			}
		}
	}

	return nil, false
}

type APLValueBetween_Common struct {
	DefaultAPLValueImpl
	check APLValue
	lo    APLValue
	hi    APLValue
}

func (value *APLValueBetween_Common) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}

func (value *APLValueBetween_Common) String() string {
	return fmt.Sprintf("%s %s %s", value.check, value.lo, value.hi)
}

type APLValueBetweenInt_IncInc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenInt_IncInc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetInt(sim)
	return value.lo.GetInt(sim) <= checkValue && checkValue <= value.hi.GetInt(sim)
}

type APLValueBetweenInt_ExcExc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenInt_ExcExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetInt(sim)
	return value.lo.GetInt(sim) < checkValue && checkValue < value.hi.GetInt(sim)
}

type APLValueBetweenInt_IncExc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenInt_IncExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetInt(sim)
	return value.lo.GetInt(sim) <= checkValue && checkValue < value.hi.GetInt(sim)
}

type APLValueBetweenInt_ExcInc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenInt_ExcInc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetInt(sim)
	return value.lo.GetInt(sim) < checkValue && checkValue <= value.hi.GetInt(sim)
}

type APLValueBetweenFloat_IncInc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenFloat_IncInc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetFloat(sim)
	return value.lo.GetFloat(sim) <= checkValue && checkValue <= value.hi.GetFloat(sim)
}

type APLValueBetweenFloat_ExcExc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenFloat_ExcExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetFloat(sim)
	return value.lo.GetFloat(sim) < checkValue && checkValue < value.hi.GetFloat(sim)
}

type APLValueBetweenFloat_IncExc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenFloat_IncExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetFloat(sim)
	return value.lo.GetFloat(sim) <= checkValue && checkValue < value.hi.GetFloat(sim)
}

type APLValueBetweenFloat_ExcInc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenFloat_ExcInc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetFloat(sim)
	return value.lo.GetFloat(sim) < checkValue && checkValue <= value.hi.GetFloat(sim)
}

type APLValueBetweenDuration_IncInc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenDuration_IncInc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetDuration(sim)
	return value.lo.GetDuration(sim) <= checkValue && checkValue <= value.hi.GetDuration(sim)
}

type APLValueBetweenDuration_ExcExc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenDuration_ExcExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetDuration(sim)
	return value.lo.GetDuration(sim) < checkValue && checkValue < value.hi.GetDuration(sim)
}

type APLValueBetweenDuration_IncExc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenDuration_IncExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetDuration(sim)
	return value.lo.GetDuration(sim) <= checkValue && checkValue < value.hi.GetDuration(sim)
}

type APLValueBetweenDuration_ExcInc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenDuration_ExcInc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetDuration(sim)
	return value.lo.GetDuration(sim) < checkValue && checkValue <= value.hi.GetDuration(sim)
}

type APLValueBetween_CommonConstInt struct {
	DefaultAPLValueImpl
	check            APLValue
	loValue, hiValue int32
}

func (value *APLValueBetween_CommonConstInt) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}

func (value *APLValueBetween_CommonConstInt) String() string {
	return fmt.Sprintf("%s %d %d", value.check, value.loValue, value.hiValue)
}

type APLValueBetweenInt_ConstIncInc struct {
	APLValueBetween_CommonConstInt
}

func (value *APLValueBetweenInt_ConstIncInc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetInt(sim)
	return value.loValue <= checkValue && checkValue <= value.hiValue
}

type APLValueBetweenInt_ConstExcExc struct {
	APLValueBetween_CommonConstInt
}

func (value *APLValueBetweenInt_ConstExcExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetInt(sim)
	return value.loValue < checkValue && checkValue < value.hiValue
}

type APLValueBetweenInt_ConstIncExc struct {
	APLValueBetween_CommonConstInt
}

func (value *APLValueBetweenInt_ConstIncExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetInt(sim)
	return value.loValue <= checkValue && checkValue < value.hiValue
}

type APLValueBetweenInt_ConstExcInc struct {
	APLValueBetween_CommonConstInt
}

func (value *APLValueBetweenInt_ConstExcInc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetInt(sim)
	return value.loValue < checkValue && checkValue <= value.hiValue
}

type APLValueBetween_CommonConstFloat struct {
	DefaultAPLValueImpl
	check            APLValue
	loValue, hiValue float64
}

func (value *APLValueBetween_CommonConstFloat) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}

func (value *APLValueBetween_CommonConstFloat) String() string {
	return fmt.Sprintf("%s %f %f", value.check, value.loValue, value.hiValue)
}

type APLValueBetweenFloat_ConstIncInc struct {
	APLValueBetween_CommonConstFloat
}

func (value *APLValueBetweenFloat_ConstIncInc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetFloat(sim)
	return value.loValue <= checkValue && checkValue <= value.hiValue
}

type APLValueBetweenFloat_ConstExcExc struct {
	APLValueBetween_CommonConstFloat
}

func (value *APLValueBetweenFloat_ConstExcExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetFloat(sim)
	return value.loValue < checkValue && checkValue < value.hiValue
}

type APLValueBetweenFloat_ConstIncExc struct {
	APLValueBetween_CommonConstFloat
}

func (value *APLValueBetweenFloat_ConstIncExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetFloat(sim)
	return value.loValue <= checkValue && checkValue < value.hiValue
}

type APLValueBetweenFloat_ConstExcInc struct {
	APLValueBetween_CommonConstFloat
}

func (value *APLValueBetweenFloat_ConstExcInc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetFloat(sim)
	return value.loValue < checkValue && checkValue <= value.hiValue
}

type APLValueBetween_CommonConstDuration struct {
	DefaultAPLValueImpl
	check            APLValue
	loValue, hiValue time.Duration
}

func (value *APLValueBetween_CommonConstDuration) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}

func (value *APLValueBetween_CommonConstDuration) String() string {
	return fmt.Sprintf("%s %d %d", value.check, value.loValue, value.hiValue)
}

type APLValueBetweenDuration_ConstIncInc struct {
	APLValueBetween_CommonConstDuration
}

func (value *APLValueBetweenDuration_ConstIncInc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetDuration(sim)
	return value.loValue <= checkValue && checkValue <= value.hiValue
}

type APLValueBetweenDuration_ConstExcExc struct {
	APLValueBetween_CommonConstDuration
}

func (value *APLValueBetweenDuration_ConstExcExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetDuration(sim)
	return value.loValue < checkValue && checkValue < value.hiValue
}

type APLValueBetweenDuration_ConstIncExc struct {
	APLValueBetween_CommonConstDuration
}

func (value *APLValueBetweenDuration_ConstIncExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetDuration(sim)
	return value.loValue <= checkValue && checkValue < value.hiValue
}

type APLValueBetweenDuration_ConstExcInc struct {
	APLValueBetween_CommonConstDuration
}

func (value *APLValueBetweenDuration_ConstExcInc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetDuration(sim)
	return value.loValue < checkValue && checkValue <= value.hiValue
}
