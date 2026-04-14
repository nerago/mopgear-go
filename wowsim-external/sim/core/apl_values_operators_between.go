package core

import (
	"fmt"
	"time"

	"github.com/wowsims/mop/sim/core/proto"
)

func makeBetween(and *APLValueAnd, rot *APLRotation) (APLValue, bool) {
	cmpA, isCmpA := and.vals[0].(APLValueCompareLike)
	cmpB, isCmpB := and.vals[1].(APLValueCompareLike)
	if !isCmpA || !isCmpB {
		return nil, false
	}

	var common, hiLoA, hiLoB APLValue
	valueA0, valueA1 := cmpA.GetInnerValues()[0], cmpA.GetInnerValues()[1]
	valueB0, valueB1 := cmpB.GetInnerValues()[0], cmpB.GetInnerValues()[1]
	if valueA0 == valueB0 {
		common = valueA0
		hiLoA = valueA1
		hiLoB = valueB1
	} else if valueA1 == valueB1 {
		common = valueA1
		hiLoA = valueA0
		hiLoB = valueB0
	} else if valueA0 == valueB1 {
		common = valueA0
		hiLoA = valueA1
		hiLoB = valueB0
	} else if valueA1 == valueB0 {
		common = valueA1
		hiLoA = valueA0
		hiLoB = valueB1
	} else {
		return nil, false
	}

	var hi, lo APLValue
	var inclusiveLo, inclusiveHi bool
	opA := cmpA.Op()
	opB := cmpB.Op()
	if (opA == proto.APLValueCompare_OpLt || opA == proto.APLValueCompare_OpLe) && (opB == proto.APLValueCompare_OpGt || opB == proto.APLValueCompare_OpGe) {
		hi = hiLoA
		lo = hiLoB
		inclusiveLo = (opA == proto.APLValueCompare_OpLe)
		inclusiveHi = (opB == proto.APLValueCompare_OpGe)
	} else if (opB == proto.APLValueCompare_OpLt || opB == proto.APLValueCompare_OpLe) && (opA == proto.APLValueCompare_OpGt || opA == proto.APLValueCompare_OpGe) {
		hi = hiLoB
		lo = hiLoA
		inclusiveLo = (opB == proto.APLValueCompare_OpLe)
		inclusiveHi = (opA == proto.APLValueCompare_OpGe)
	} else {
		return nil, false
	}

	list := rot.coerceAllToSameType([]APLValue{common, hi, lo}) // what if placeholders?
	common, hi, lo = list[0], list[1], list[2]

	constHi, isConstHi := hi.(*APLValueConst)
	constLo, isConstLo := lo.(*APLValueConst)
	if isConstHi && isConstLo {
		switch common.Type() {
		case proto.APLValueType_ValueTypeInt:
			return &APLValueBetweenIntConst{check: common, loValue: constLo.GetInt(nil), hiValue: constHi.GetInt(nil), inclusiveLo: inclusiveLo, inclusiveHi: inclusiveHi}, true
		case proto.APLValueType_ValueTypeFloat:
			return &APLValueBetweenFloatConst{check: common, loValue: getConstAPLFloatValue(constLo), hiValue: getConstAPLFloatValue(constHi), inclusiveLo: inclusiveLo, inclusiveHi: inclusiveHi}, true
		case proto.APLValueType_ValueTypeDuration:
			return &APLValueBetweenDurationConst{check: common, loValue: constLo.GetDuration(nil), hiValue: constHi.GetDuration(nil), inclusiveLo: inclusiveLo, inclusiveHi: inclusiveHi}, true
		}
	} else {
		switch common.Type() {
		case proto.APLValueType_ValueTypeInt:
			return &APLValueBetweenInt{check: common, lo: lo, hi: hi, inclusiveLo: inclusiveLo, inclusiveHi: inclusiveHi}, true
		case proto.APLValueType_ValueTypeFloat:
			return &APLValueBetweenFloat{check: common, lo: lo, hi: hi, inclusiveLo: inclusiveLo, inclusiveHi: inclusiveHi}, true
		case proto.APLValueType_ValueTypeDuration:
			return &APLValueBetweenDuration{check: common, lo: lo, hi: hi, inclusiveLo: inclusiveLo, inclusiveHi: inclusiveHi}, true
		}
	}

	return nil, false
}

type APLValueBetweenInt struct {
	DefaultAPLValueImpl
	check                    APLValue
	lo                       APLValue
	hi                       APLValue
	inclusiveLo, inclusiveHi bool
}

func (value *APLValueBetweenInt) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetInt(sim)

	loValue := value.lo.GetInt(sim)
	if value.inclusiveLo && checkValue < loValue {
		return false
	} else if !value.inclusiveLo && checkValue <= loValue {
		return false
	}

	hiValue := value.hi.GetInt(sim)
	if value.inclusiveHi && checkValue > hiValue {
		return false
	} else if !value.inclusiveHi && checkValue >= hiValue {
		return false
	}

	return true
}

func (value *APLValueBetweenInt) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}

func (value *APLValueBetweenInt) String() string {
	return fmt.Sprintf("%s %s %s %v %v", value.check, value.lo, value.hi, value.inclusiveLo, value.inclusiveHi)
}

type APLValueBetweenFloat struct {
	DefaultAPLValueImpl
	check                    APLValue
	lo                       APLValue
	hi                       APLValue
	inclusiveLo, inclusiveHi bool
}

func (value *APLValueBetweenFloat) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetFloat(sim)

	loValue := value.lo.GetFloat(sim)
	if value.inclusiveLo && checkValue < loValue {
		return false
	} else if !value.inclusiveLo && checkValue <= loValue {
		return false
	}

	hiValue := value.hi.GetFloat(sim)
	if value.inclusiveHi && checkValue > hiValue {
		return false
	} else if !value.inclusiveHi && checkValue >= hiValue {
		return false
	}

	return true
}

func (value *APLValueBetweenFloat) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}

func (value *APLValueBetweenFloat) String() string {
	return fmt.Sprintf("%s %s %s %v %v", value.check, value.lo, value.hi, value.inclusiveLo, value.inclusiveHi)
}

type APLValueBetweenDuration struct {
	DefaultAPLValueImpl
	check                    APLValue
	lo                       APLValue
	hi                       APLValue
	inclusiveLo, inclusiveHi bool
}

func (value *APLValueBetweenDuration) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetDuration(sim)

	loValue := value.lo.GetDuration(sim)
	if value.inclusiveLo && checkValue < loValue {
		return false
	} else if !value.inclusiveLo && checkValue <= loValue {
		return false
	}

	hiValue := value.hi.GetDuration(sim)
	if value.inclusiveHi && checkValue > hiValue {
		return false
	} else if !value.inclusiveHi && checkValue >= hiValue {
		return false
	}

	return true
}

func (value *APLValueBetweenDuration) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}

func (value *APLValueBetweenDuration) String() string {
	return fmt.Sprintf("%s %s %s %v %v", value.check, value.lo, value.hi, value.inclusiveLo, value.inclusiveHi)
}

type APLValueBetweenIntConst struct {
	DefaultAPLValueImpl
	check                    APLValue
	loValue, hiValue         int32
	inclusiveLo, inclusiveHi bool
}

func (value *APLValueBetweenIntConst) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetInt(sim)

	loValue := value.loValue
	if value.inclusiveLo && checkValue < loValue {
		return false
	} else if !value.inclusiveLo && checkValue <= loValue {
		return false
	}

	hiValue := value.hiValue
	if value.inclusiveHi && checkValue > hiValue {
		return false
	} else if !value.inclusiveHi && checkValue >= hiValue {
		return false
	}

	return true
}

func (value *APLValueBetweenIntConst) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}

func (value *APLValueBetweenIntConst) String() string {
	return fmt.Sprintf("%s %d %d %v %v", value.check, value.loValue, value.hiValue, value.inclusiveLo, value.inclusiveHi)
}

type APLValueBetweenFloatConst struct {
	DefaultAPLValueImpl
	check                    APLValue
	loValue, hiValue         float64
	inclusiveLo, inclusiveHi bool
}

func (value *APLValueBetweenFloatConst) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetFloat(sim)

	loValue := value.loValue
	if value.inclusiveLo && checkValue < loValue {
		return false
	} else if !value.inclusiveLo && checkValue <= loValue {
		return false
	}

	hiValue := value.hiValue
	if value.inclusiveHi && checkValue > hiValue {
		return false
	} else if !value.inclusiveHi && checkValue >= hiValue {
		return false
	}

	return true
}

func (value *APLValueBetweenFloatConst) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}

func (value *APLValueBetweenFloatConst) String() string {
	return fmt.Sprintf("%s %f %f %v %v", value.check, value.loValue, value.hiValue, value.inclusiveLo, value.inclusiveHi)
}

type APLValueBetweenDurationConst struct {
	DefaultAPLValueImpl
	check                    APLValue
	loValue, hiValue         time.Duration
	inclusiveLo, inclusiveHi bool
}

func (value *APLValueBetweenDurationConst) GetBool(sim *Simulation) bool {
	checkValue := value.check.GetDuration(sim)

	loValue := value.loValue
	if value.inclusiveLo && checkValue < loValue {
		return false
	} else if !value.inclusiveLo && checkValue <= loValue {
		return false
	}

	hiValue := value.hiValue
	if value.inclusiveHi && checkValue > hiValue {
		return false
	} else if !value.inclusiveHi && checkValue >= hiValue {
		return false
	}

	return true
}

func (value *APLValueBetweenDurationConst) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}

func (value *APLValueBetweenDurationConst) String() string {
	return fmt.Sprintf("%s %d %d %v %v", value.check, value.loValue, value.hiValue, value.inclusiveLo, value.inclusiveHi)
}
