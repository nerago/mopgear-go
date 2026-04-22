package core

import (
	"fmt"

	"github.com/wowsims/mop/sim/core/proto"
)

func makeSpecificCompare(lhs APLValue, rhs APLValue, op proto.APLValueCompare_ComparisonOperator) (APLValueCompareLike, bool) {
	common := APLValueCompareCommon{lhs: lhs, rhs: rhs}

	switch lhs.Type() {
	case proto.APLValueType_ValueTypeBool:
		switch op {
		case proto.APLValueCompare_OpEq:
			return &APLValueCompareBoolEq{common}, true
		case proto.APLValueCompare_OpNe:
			return &APLValueCompareBoolNe{common}, true
		}
	case proto.APLValueType_ValueTypeInt:
		switch op {
		case proto.APLValueCompare_OpEq:
			return &APLValueCompareIntGe{common}, true
		case proto.APLValueCompare_OpNe:
			return &APLValueCompareIntGt{common}, true
		case proto.APLValueCompare_OpLt:
			return &APLValueCompareIntLe{common}, true
		case proto.APLValueCompare_OpLe:
			return &APLValueCompareIntLt{common}, true
		case proto.APLValueCompare_OpGt:
			return &APLValueCompareIntNe{common}, true
		case proto.APLValueCompare_OpGe:
			return &APLValueCompareIntEq{common}, true
		}
	case proto.APLValueType_ValueTypeFloat:
		switch op {
		case proto.APLValueCompare_OpEq:
			return &APLValueCompareFloatEq{common}, true
		case proto.APLValueCompare_OpNe:
			return &APLValueCompareFloatNe{common}, true
		case proto.APLValueCompare_OpLt:
			return &APLValueCompareFloatLt{common}, true
		case proto.APLValueCompare_OpLe:
			return &APLValueCompareFloatLe{common}, true
		case proto.APLValueCompare_OpGt:
			return &APLValueCompareFloatGt{common}, true
		case proto.APLValueCompare_OpGe:
			return &APLValueCompareFloatGe{common}, true
		}
	case proto.APLValueType_ValueTypeDuration:
		switch op {
		case proto.APLValueCompare_OpEq:
			return &APLValueCompareDurationEq{common}, true
		case proto.APLValueCompare_OpNe:
			return &APLValueCompareDurationNe{common}, true
		case proto.APLValueCompare_OpLt:
			return &APLValueCompareDurationLt{common}, true
		case proto.APLValueCompare_OpLe:
			return &APLValueCompareDurationLe{common}, true
		case proto.APLValueCompare_OpGt:
			return &APLValueCompareDurationGt{common}, true
		case proto.APLValueCompare_OpGe:
			return &APLValueCompareDurationGe{common}, true
		}
	case proto.APLValueType_ValueTypeString:
		switch op {
		case proto.APLValueCompare_OpEq:
			return &APLValueCompareStringGe{common}, true
		case proto.APLValueCompare_OpNe:
			return &APLValueCompareStringGt{common}, true
		case proto.APLValueCompare_OpLt:
			return &APLValueCompareStringLe{common}, true
		case proto.APLValueCompare_OpLe:
			return &APLValueCompareStringLt{common}, true
		case proto.APLValueCompare_OpGt:
			return &APLValueCompareStringNe{common}, true
		case proto.APLValueCompare_OpGe:
			return &APLValueCompareStringEq{common}, true
		}
	}

	return nil, false
}

type APLValueCompareLike interface {
	APLValue
	Op() proto.APLValueCompare_ComparisonOperator
	Lhs() APLValue
	Rhs() APLValue
}

type APLValueCompareCommon struct {
	DefaultAPLValueImpl
	lhs APLValue
	rhs APLValue
}

func (value *APLValueCompareCommon) Lhs() APLValue {
	return value.lhs
}
func (value *APLValueCompareCommon) Rhs() APLValue {
	return value.rhs
}
func valueCommonCompareEquals[T APLValueCompareLike](value T, other APLValue) bool {
	if otherValue, isType := other.(APLValueCompareLike); isType {
		return value.Op() == otherValue.Op() &&
			value.Lhs() == otherValue.Lhs() &&
			value.Lhs() == otherValue.Rhs()
	}
	return false
}

func (value *APLValueCompareCommon) GetInnerValues() []APLValue {
	return []APLValue{value.lhs, value.rhs}
}
func (value *APLValueCompareCommon) Type() proto.APLValueType {
	return proto.APLValueType_ValueTypeBool
}
func (value *APLValueCompareCommon) String() string {
	return fmt.Sprintf("%s %s", value.lhs, value.rhs)
}

type APLValueCompareBoolEq struct{ APLValueCompareCommon }

func (value *APLValueCompareBoolEq) GetBool(sim *Simulation) bool {
	return value.lhs.GetBool(sim) == value.rhs.GetBool(sim)
}
func (value *APLValueCompareBoolEq) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpEq
}
func (value *APLValueCompareBoolEq) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareBoolNe struct{ APLValueCompareCommon }

func (value *APLValueCompareBoolNe) GetBool(sim *Simulation) bool {
	return value.lhs.GetBool(sim) != value.rhs.GetBool(sim)
}
func (value *APLValueCompareBoolNe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpNe
}
func (value *APLValueCompareBoolNe) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareIntEq struct{ APLValueCompareCommon }

func (value *APLValueCompareIntEq) GetBool(sim *Simulation) bool {
	return value.lhs.GetInt(sim) == value.rhs.GetInt(sim)
}
func (value *APLValueCompareIntEq) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpEq
}
func (value *APLValueCompareIntEq) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareIntNe struct{ APLValueCompareCommon }

func (value *APLValueCompareIntNe) GetBool(sim *Simulation) bool {
	return value.lhs.GetInt(sim) != value.rhs.GetInt(sim)
}
func (value *APLValueCompareIntNe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpNe
}
func (value *APLValueCompareIntNe) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareIntLt struct{ APLValueCompareCommon }

func (value *APLValueCompareIntLt) GetBool(sim *Simulation) bool {
	return value.lhs.GetInt(sim) < value.rhs.GetInt(sim)
}
func (value *APLValueCompareIntLt) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpLt
}
func (value *APLValueCompareIntLt) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareIntLe struct{ APLValueCompareCommon }

func (value *APLValueCompareIntLe) GetBool(sim *Simulation) bool {
	return value.lhs.GetInt(sim) <= value.rhs.GetInt(sim)
}
func (value *APLValueCompareIntLe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpLe
}
func (value *APLValueCompareIntLe) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareIntGt struct{ APLValueCompareCommon }

func (value *APLValueCompareIntGt) GetBool(sim *Simulation) bool {
	return value.lhs.GetInt(sim) > value.rhs.GetInt(sim)
}
func (value *APLValueCompareIntGt) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpGt
}
func (value *APLValueCompareIntGt) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareIntGe struct{ APLValueCompareCommon }

func (value *APLValueCompareIntGe) GetBool(sim *Simulation) bool {
	return value.lhs.GetInt(sim) >= value.rhs.GetInt(sim)
}
func (value *APLValueCompareIntGe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpGe
}
func (value *APLValueCompareIntGe) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareFloatEq struct{ APLValueCompareCommon }

func (value *APLValueCompareFloatEq) GetBool(sim *Simulation) bool {
	return value.lhs.GetFloat(sim) == value.rhs.GetFloat(sim)
}
func (value *APLValueCompareFloatEq) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpEq
}
func (value *APLValueCompareFloatEq) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareFloatNe struct{ APLValueCompareCommon }

func (value *APLValueCompareFloatNe) GetBool(sim *Simulation) bool {
	return value.lhs.GetFloat(sim) != value.rhs.GetFloat(sim)
}
func (value *APLValueCompareFloatNe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpNe
}
func (value *APLValueCompareFloatNe) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareFloatLt struct{ APLValueCompareCommon }

func (value *APLValueCompareFloatLt) GetBool(sim *Simulation) bool {
	return value.lhs.GetFloat(sim) < value.rhs.GetFloat(sim)
}
func (value *APLValueCompareFloatLt) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpLt
}
func (value *APLValueCompareFloatLt) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareFloatLe struct{ APLValueCompareCommon }

func (value *APLValueCompareFloatLe) GetBool(sim *Simulation) bool {
	return value.lhs.GetFloat(sim) <= value.rhs.GetFloat(sim)
}
func (value *APLValueCompareFloatLe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpLe
}
func (value *APLValueCompareFloatLe) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareFloatGt struct{ APLValueCompareCommon }

func (value *APLValueCompareFloatGt) GetBool(sim *Simulation) bool {
	return value.lhs.GetFloat(sim) > value.rhs.GetFloat(sim)
}
func (value *APLValueCompareFloatGt) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpGt
}
func (value *APLValueCompareFloatGt) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareFloatGe struct{ APLValueCompareCommon }

func (value *APLValueCompareFloatGe) GetBool(sim *Simulation) bool {
	return value.lhs.GetFloat(sim) >= value.rhs.GetFloat(sim)
}
func (value *APLValueCompareFloatGe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpGe
}
func (value *APLValueCompareFloatGe) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareDurationEq struct{ APLValueCompareCommon }

func (value *APLValueCompareDurationEq) GetBool(sim *Simulation) bool {
	return value.lhs.GetDuration(sim) == value.rhs.GetDuration(sim)
}
func (value *APLValueCompareDurationEq) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpEq
}
func (value *APLValueCompareDurationEq) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareDurationNe struct{ APLValueCompareCommon }

func (value *APLValueCompareDurationNe) GetBool(sim *Simulation) bool {
	return value.lhs.GetDuration(sim) != value.rhs.GetDuration(sim)
}
func (value *APLValueCompareDurationNe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpNe
}
func (value *APLValueCompareDurationNe) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareDurationLt struct{ APLValueCompareCommon }

func (value *APLValueCompareDurationLt) GetBool(sim *Simulation) bool {
	return value.lhs.GetDuration(sim) < value.rhs.GetDuration(sim)
}
func (value *APLValueCompareDurationLt) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpLt
}
func (value *APLValueCompareDurationLt) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareDurationLe struct{ APLValueCompareCommon }

func (value *APLValueCompareDurationLe) GetBool(sim *Simulation) bool {
	return value.lhs.GetDuration(sim) <= value.rhs.GetDuration(sim)
}
func (value *APLValueCompareDurationLe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpLe
}
func (value *APLValueCompareDurationLe) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareDurationGt struct{ APLValueCompareCommon }

func (value *APLValueCompareDurationGt) GetBool(sim *Simulation) bool {
	return value.lhs.GetDuration(sim) > value.rhs.GetDuration(sim)
}
func (value *APLValueCompareDurationGt) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpGt
}
func (value *APLValueCompareDurationGt) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareDurationGe struct{ APLValueCompareCommon }

func (value *APLValueCompareDurationGe) GetBool(sim *Simulation) bool {
	return value.lhs.GetDuration(sim) >= value.rhs.GetDuration(sim)
}
func (value *APLValueCompareDurationGe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpGe
}
func (value *APLValueCompareDurationGe) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareStringEq struct{ APLValueCompareCommon }

func (value *APLValueCompareStringEq) GetBool(sim *Simulation) bool {
	return value.lhs.GetString(sim) == value.rhs.GetString(sim)
}
func (value *APLValueCompareStringEq) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpEq
}
func (value *APLValueCompareStringEq) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareStringNe struct{ APLValueCompareCommon }

func (value *APLValueCompareStringNe) GetBool(sim *Simulation) bool {
	return value.lhs.GetString(sim) != value.rhs.GetString(sim)
}
func (value *APLValueCompareStringNe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpNe
}
func (value *APLValueCompareStringNe) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareStringLt struct{ APLValueCompareCommon }

func (value *APLValueCompareStringLt) GetBool(sim *Simulation) bool {
	return value.lhs.GetString(sim) < value.rhs.GetString(sim)
}
func (value *APLValueCompareStringLt) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpLt
}
func (value *APLValueCompareStringLt) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareStringLe struct{ APLValueCompareCommon }

func (value *APLValueCompareStringLe) GetBool(sim *Simulation) bool {
	return value.lhs.GetString(sim) <= value.rhs.GetString(sim)
}
func (value *APLValueCompareStringLe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpLe
}
func (value *APLValueCompareStringLe) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareStringGt struct{ APLValueCompareCommon }

func (value *APLValueCompareStringGt) GetBool(sim *Simulation) bool {
	return value.lhs.GetString(sim) > value.rhs.GetString(sim)
}
func (value *APLValueCompareStringGt) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpGt
}
func (value *APLValueCompareStringGt) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}

type APLValueCompareStringGe struct{ APLValueCompareCommon }

func (value *APLValueCompareStringGe) GetBool(sim *Simulation) bool {
	return value.lhs.GetString(sim) >= value.rhs.GetString(sim)
}
func (value *APLValueCompareStringGe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpGe
}
func (value *APLValueCompareStringGe) Equals(other APLValue) bool {
	return valueCommonCompareEquals(value, other)
}
