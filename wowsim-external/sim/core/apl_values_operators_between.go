package core

import (
	"fmt"
	"time"

	"github.com/wowsims/mop/sim/core/proto"
)

func makeBetween(vals []APLValue, rot *APLRotation) (between APLValue, orphan APLValue) {
	// only support exactly 2 elements (an upper and lower bound)
	if len(vals) != 2 {
		return nil, nil
	}

	// they should both be compare operations
	cmpA, isCmpA := vals[0].(APLValueCompareLike)
	cmpB, isCmpB := vals[1].(APLValueCompareLike)
	if !isCmpA || !isCmpB {
		return nil, nil
	}

	// extract the compare components
	valueA0, valueA1 := cmpA.GetInnerValues()[0], cmpA.GetInnerValues()[1]
	valueB0, valueB1 := cmpB.GetInnerValues()[0], cmpB.GetInnerValues()[1]

	// check if we have the general "between" pattern; need an identical value to check on both sides
	var check, otherA, otherB APLValue
	if valueA0.Equals(valueB0) {
		check = valueA0
		orphan = valueB0
		otherA = valueA1
		otherB = valueB1
	} else if valueA1.Equals(valueB1) {
		check = valueA1
		orphan = valueB1
		otherA = valueA0
		otherB = valueB0
	} else if valueA0.Equals(valueB1) {
		check = valueA0
		orphan = valueB1
		otherA = valueA1
		otherB = valueB0
	} else if valueA1.Equals(valueB0) {
		check = valueA1
		orphan = valueB0
		otherA = valueA0
		otherB = valueB1
	} else {
		return nil, nil
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
		inclusiveHi = (opB == proto.APLValueCompare_OpLe)
		lo = otherA
		inclusiveLo = (opA == proto.APLValueCompare_OpGe)
	} else {
		return nil, nil
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
				return &APLValueBetweenInt_ConstIncInc{common}, orphan
			} else if inclusiveLo {
				return &APLValueBetweenInt_ConstIncExc{common}, orphan
			} else if inclusiveHi {
				return &APLValueBetweenInt_ConstExcInc{common}, orphan
			} else {
				return &APLValueBetweenInt_ConstExcExc{common}, orphan
			}
		case proto.APLValueType_ValueTypeFloat:
			common := APLValueBetween_CommonConstFloat{check: check, loValue: constLo.GetFloat(nil), hiValue: constHi.GetFloat(nil)}
			if inclusiveLo && inclusiveHi {
				return &APLValueBetweenFloat_ConstIncInc{common}, orphan
			} else if inclusiveLo {
				return &APLValueBetweenFloat_ConstIncExc{common}, orphan
			} else if inclusiveHi {
				return &APLValueBetweenFloat_ConstExcInc{common}, orphan
			} else {
				return &APLValueBetweenFloat_ConstExcExc{common}, orphan
			}
		case proto.APLValueType_ValueTypeDuration:
			common := APLValueBetween_CommonConstDuration{check: check, loValue: constLo.GetDuration(nil), hiValue: constHi.GetDuration(nil)}
			if inclusiveLo && inclusiveHi {
				return &APLValueBetweenDuration_ConstIncInc{common}, orphan
			} else if inclusiveLo {
				return &APLValueBetweenDuration_ConstIncExc{common}, orphan
			} else if inclusiveHi {
				return &APLValueBetweenDuration_ConstExcInc{common}, orphan
			} else {
				return &APLValueBetweenDuration_ConstExcExc{common}, orphan
			}
		}
	} else {
		// dynamic value implementations
		common := APLValueBetween_Common{check: check, lo: lo, hi: hi}
		switch check.Type() {
		case proto.APLValueType_ValueTypeInt:
			if inclusiveLo && inclusiveHi {
				return &APLValueBetweenInt_IncInc{common}, orphan
			} else if inclusiveLo {
				return &APLValueBetweenInt_IncExc{common}, orphan
			} else if inclusiveHi {
				return &APLValueBetweenInt_ExcInc{common}, orphan
			} else {
				return &APLValueBetweenInt_ExcExc{common}, orphan
			}
		case proto.APLValueType_ValueTypeFloat:
			if inclusiveLo && inclusiveHi {
				return &APLValueBetweenFloat_IncInc{common}, orphan
			} else if inclusiveLo {
				return &APLValueBetweenFloat_IncExc{common}, orphan
			} else if inclusiveHi {
				return &APLValueBetweenFloat_ExcInc{common}, orphan
			} else {
				return &APLValueBetweenFloat_ExcExc{common}, orphan
			}
		case proto.APLValueType_ValueTypeDuration:
			if inclusiveLo && inclusiveHi {
				return &APLValueBetweenDuration_IncInc{common}, orphan
			} else if inclusiveLo {
				return &APLValueBetweenDuration_IncExc{common}, orphan
			} else if inclusiveHi {
				return &APLValueBetweenDuration_ExcInc{common}, orphan
			} else {
				return &APLValueBetweenDuration_ExcExc{common}, orphan
			}
		}
	}

	return nil, nil
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
func (value *APLValueBetweenInt_IncInc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenInt_IncInc); isType {
		return APLValueEquals(value.check, otherValue.check) && APLValueEquals(value.lo, otherValue.lo) && APLValueEquals(value.hi, otherValue.hi)
	}
	return false
}

type APLValueBetweenInt_ExcExc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenInt_ExcExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetInt(sim)
	return value.lo.GetInt(sim) < checkValue && checkValue < value.hi.GetInt(sim)
}
func (value *APLValueBetweenInt_ExcExc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenInt_ExcExc); isType {
		return APLValueEquals(value.check, otherValue.check) && APLValueEquals(value.lo, otherValue.lo) && APLValueEquals(value.hi, otherValue.hi)
	}
	return false
}

type APLValueBetweenInt_IncExc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenInt_IncExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetInt(sim)
	return value.lo.GetInt(sim) <= checkValue && checkValue < value.hi.GetInt(sim)
}
func (value *APLValueBetweenInt_IncExc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenInt_IncExc); isType {
		return APLValueEquals(value.check, otherValue.check) && APLValueEquals(value.lo, otherValue.lo) && APLValueEquals(value.hi, otherValue.hi)
	}
	return false
}

type APLValueBetweenInt_ExcInc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenInt_ExcInc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetInt(sim)
	return value.lo.GetInt(sim) < checkValue && checkValue <= value.hi.GetInt(sim)
}
func (value *APLValueBetweenInt_ExcInc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenInt_ExcInc); isType {
		return APLValueEquals(value.check, otherValue.check) && APLValueEquals(value.lo, otherValue.lo) && APLValueEquals(value.hi, otherValue.hi)
	}
	return false
}

type APLValueBetweenFloat_IncInc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenFloat_IncInc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetFloat(sim)
	return value.lo.GetFloat(sim) <= checkValue && checkValue <= value.hi.GetFloat(sim)
}
func (value *APLValueBetweenFloat_IncInc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenFloat_IncInc); isType {
		return APLValueEquals(value.check, otherValue.check) && APLValueEquals(value.lo, otherValue.lo) && APLValueEquals(value.hi, otherValue.hi)
	}
	return false
}

type APLValueBetweenFloat_ExcExc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenFloat_ExcExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetFloat(sim)
	return value.lo.GetFloat(sim) < checkValue && checkValue < value.hi.GetFloat(sim)
}
func (value *APLValueBetweenFloat_ExcExc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenFloat_ExcExc); isType {
		return APLValueEquals(value.check, otherValue.check) && APLValueEquals(value.lo, otherValue.lo) && APLValueEquals(value.hi, otherValue.hi)
	}
	return false
}

type APLValueBetweenFloat_IncExc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenFloat_IncExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetFloat(sim)
	return value.lo.GetFloat(sim) <= checkValue && checkValue < value.hi.GetFloat(sim)
}
func (value *APLValueBetweenFloat_IncExc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenFloat_IncExc); isType {
		return APLValueEquals(value.check, otherValue.check) && APLValueEquals(value.lo, otherValue.lo) && APLValueEquals(value.hi, otherValue.hi)
	}
	return false
}

type APLValueBetweenFloat_ExcInc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenFloat_ExcInc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetFloat(sim)
	return value.lo.GetFloat(sim) < checkValue && checkValue <= value.hi.GetFloat(sim)
}
func (value *APLValueBetweenFloat_ExcInc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenFloat_ExcInc); isType {
		return APLValueEquals(value.check, otherValue.check) && APLValueEquals(value.lo, otherValue.lo) && APLValueEquals(value.hi, otherValue.hi)
	}
	return false
}

type APLValueBetweenDuration_IncInc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenDuration_IncInc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetDuration(sim)
	return value.lo.GetDuration(sim) <= checkValue && checkValue <= value.hi.GetDuration(sim)
}
func (value *APLValueBetweenDuration_IncInc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenDuration_IncInc); isType {
		return APLValueEquals(value.check, otherValue.check) && APLValueEquals(value.lo, otherValue.lo) && APLValueEquals(value.hi, otherValue.hi)
	}
	return false
}

type APLValueBetweenDuration_ExcExc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenDuration_ExcExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetDuration(sim)
	return value.lo.GetDuration(sim) < checkValue && checkValue < value.hi.GetDuration(sim)
}
func (value *APLValueBetweenDuration_ExcExc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenDuration_ExcExc); isType {
		return APLValueEquals(value.check, otherValue.check) && APLValueEquals(value.lo, otherValue.lo) && APLValueEquals(value.hi, otherValue.hi)
	}
	return false
}

type APLValueBetweenDuration_IncExc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenDuration_IncExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetDuration(sim)
	return value.lo.GetDuration(sim) <= checkValue && checkValue < value.hi.GetDuration(sim)
}
func (value *APLValueBetweenDuration_IncExc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenDuration_IncExc); isType {
		return APLValueEquals(value.check, otherValue.check) && APLValueEquals(value.lo, otherValue.lo) && APLValueEquals(value.hi, otherValue.hi)
	}
	return false
}

type APLValueBetweenDuration_ExcInc struct {
	APLValueBetween_Common
}

func (value *APLValueBetweenDuration_ExcInc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetDuration(sim)
	return value.lo.GetDuration(sim) < checkValue && checkValue <= value.hi.GetDuration(sim)
}
func (value *APLValueBetweenDuration_ExcInc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenDuration_ExcInc); isType {
		return APLValueEquals(value.check, otherValue.check) && APLValueEquals(value.lo, otherValue.lo) && APLValueEquals(value.hi, otherValue.hi)
	}
	return false
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
func (value *APLValueBetweenInt_ConstIncInc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenInt_ConstIncInc); isType {
		return APLValueEquals(value.check, otherValue.check) && value.loValue == otherValue.loValue && value.hiValue == otherValue.hiValue
	}
	return false
}

type APLValueBetweenInt_ConstExcExc struct {
	APLValueBetween_CommonConstInt
}

func (value *APLValueBetweenInt_ConstExcExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetInt(sim)
	return value.loValue < checkValue && checkValue < value.hiValue
}
func (value *APLValueBetweenInt_ConstExcExc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenInt_ConstExcExc); isType {
		return APLValueEquals(value.check, otherValue.check) && value.loValue == otherValue.loValue && value.hiValue == otherValue.hiValue
	}
	return false
}

type APLValueBetweenInt_ConstIncExc struct {
	APLValueBetween_CommonConstInt
}

func (value *APLValueBetweenInt_ConstIncExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetInt(sim)
	return value.loValue <= checkValue && checkValue < value.hiValue
}
func (value *APLValueBetweenInt_ConstIncExc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenInt_ConstIncExc); isType {
		return APLValueEquals(value.check, otherValue.check) && value.loValue == otherValue.loValue && value.hiValue == otherValue.hiValue
	}
	return false
}

type APLValueBetweenInt_ConstExcInc struct {
	APLValueBetween_CommonConstInt
}

func (value *APLValueBetweenInt_ConstExcInc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetInt(sim)
	return value.loValue < checkValue && checkValue <= value.hiValue
}
func (value *APLValueBetweenInt_ConstExcInc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenInt_ConstExcInc); isType {
		return APLValueEquals(value.check, otherValue.check) && value.loValue == otherValue.loValue && value.hiValue == otherValue.hiValue
	}
	return false
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
func (value *APLValueBetweenFloat_ConstIncInc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenFloat_ConstIncInc); isType {
		return APLValueEquals(value.check, otherValue.check) && value.loValue == otherValue.loValue && value.hiValue == otherValue.hiValue
	}
	return false
}

type APLValueBetweenFloat_ConstExcExc struct {
	APLValueBetween_CommonConstFloat
}

func (value *APLValueBetweenFloat_ConstExcExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetFloat(sim)
	return value.loValue < checkValue && checkValue < value.hiValue
}
func (value *APLValueBetweenFloat_ConstExcExc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenFloat_ConstExcExc); isType {
		return APLValueEquals(value.check, otherValue.check) && value.loValue == otherValue.loValue && value.hiValue == otherValue.hiValue
	}
	return false
}

type APLValueBetweenFloat_ConstIncExc struct {
	APLValueBetween_CommonConstFloat
}

func (value *APLValueBetweenFloat_ConstIncExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetFloat(sim)
	return value.loValue <= checkValue && checkValue < value.hiValue
}
func (value *APLValueBetweenFloat_ConstIncExc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenFloat_ConstIncExc); isType {
		return APLValueEquals(value.check, otherValue.check) && value.loValue == otherValue.loValue && value.hiValue == otherValue.hiValue
	}
	return false
}

type APLValueBetweenFloat_ConstExcInc struct {
	APLValueBetween_CommonConstFloat
}

func (value *APLValueBetweenFloat_ConstExcInc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetFloat(sim)
	return value.loValue < checkValue && checkValue <= value.hiValue
}
func (value *APLValueBetweenFloat_ConstExcInc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenFloat_ConstExcInc); isType {
		return APLValueEquals(value.check, otherValue.check) && value.loValue == otherValue.loValue && value.hiValue == otherValue.hiValue
	}
	return false
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
func (value *APLValueBetweenDuration_ConstIncInc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenDuration_ConstIncInc); isType {
		return APLValueEquals(value.check, otherValue.check) && value.loValue == otherValue.loValue && value.hiValue == otherValue.hiValue
	}
	return false
}

type APLValueBetweenDuration_ConstExcExc struct {
	APLValueBetween_CommonConstDuration
}

func (value *APLValueBetweenDuration_ConstExcExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetDuration(sim)
	return value.loValue < checkValue && checkValue < value.hiValue
}
func (value *APLValueBetweenDuration_ConstExcExc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenDuration_ConstExcExc); isType {
		return APLValueEquals(value.check, otherValue.check) && value.loValue == otherValue.loValue && value.hiValue == otherValue.hiValue
	}
	return false
}

type APLValueBetweenDuration_ConstIncExc struct {
	APLValueBetween_CommonConstDuration
}

func (value *APLValueBetweenDuration_ConstIncExc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetDuration(sim)
	return value.loValue <= checkValue && checkValue < value.hiValue
}
func (value *APLValueBetweenDuration_ConstIncExc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenDuration_ConstIncExc); isType {
		return APLValueEquals(value.check, otherValue.check) && value.loValue == otherValue.loValue && value.hiValue == otherValue.hiValue
	}
	return false
}

type APLValueBetweenDuration_ConstExcInc struct {
	APLValueBetween_CommonConstDuration
}

func (value *APLValueBetweenDuration_ConstExcInc) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetDuration(sim)
	return value.loValue < checkValue && checkValue <= value.hiValue
}
func (value *APLValueBetweenDuration_ConstExcInc) Equals(other APLValue) bool {
	if otherValue, isType := other.(*APLValueBetweenDuration_ConstExcInc); isType {
		return APLValueEquals(value.check, otherValue.check) && value.loValue == otherValue.loValue && value.hiValue == otherValue.hiValue
	}
	return false
}
