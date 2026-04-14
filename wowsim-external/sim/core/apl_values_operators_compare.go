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
}

type APLValueCompareCommon struct {
	DefaultAPLValueImpl
	lhs APLValue
	rhs APLValue
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

type APLValueCompareBoolNe struct{ APLValueCompareCommon }

func (value *APLValueCompareBoolNe) GetBool(sim *Simulation) bool {
	return value.lhs.GetBool(sim) != value.rhs.GetBool(sim)
}
func (value *APLValueCompareBoolNe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpNe
}

type APLValueCompareIntEq struct{ APLValueCompareCommon }

func (value *APLValueCompareIntEq) GetBool(sim *Simulation) bool {
	return value.lhs.GetInt(sim) == value.rhs.GetInt(sim)
}
func (value *APLValueCompareIntEq) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpEq
}

type APLValueCompareIntNe struct{ APLValueCompareCommon }

func (value *APLValueCompareIntNe) GetBool(sim *Simulation) bool {
	return value.lhs.GetInt(sim) != value.rhs.GetInt(sim)
}
func (value *APLValueCompareIntNe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpNe
}

type APLValueCompareIntLt struct{ APLValueCompareCommon }

func (value *APLValueCompareIntLt) GetBool(sim *Simulation) bool {
	return value.lhs.GetInt(sim) < value.rhs.GetInt(sim)
}
func (value *APLValueCompareIntLt) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpLt
}

type APLValueCompareIntLe struct{ APLValueCompareCommon }

func (value *APLValueCompareIntLe) GetBool(sim *Simulation) bool {
	return value.lhs.GetInt(sim) <= value.rhs.GetInt(sim)
}
func (value *APLValueCompareIntLe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpLe
}

type APLValueCompareIntGt struct{ APLValueCompareCommon }

func (value *APLValueCompareIntGt) GetBool(sim *Simulation) bool {
	return value.lhs.GetInt(sim) > value.rhs.GetInt(sim)
}
func (value *APLValueCompareIntGt) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpGt
}

type APLValueCompareIntGe struct{ APLValueCompareCommon }

func (value *APLValueCompareIntGe) GetBool(sim *Simulation) bool {
	return value.lhs.GetInt(sim) >= value.rhs.GetInt(sim)
}
func (value *APLValueCompareIntGe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpGe
}

type APLValueCompareFloatEq struct{ APLValueCompareCommon }

func (value *APLValueCompareFloatEq) GetBool(sim *Simulation) bool {
	return value.lhs.GetFloat(sim) == value.rhs.GetFloat(sim)
}
func (value *APLValueCompareFloatEq) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpEq
}

type APLValueCompareFloatNe struct{ APLValueCompareCommon }

func (value *APLValueCompareFloatNe) GetBool(sim *Simulation) bool {
	return value.lhs.GetFloat(sim) != value.rhs.GetFloat(sim)
}
func (value *APLValueCompareFloatNe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpNe
}

type APLValueCompareFloatLt struct{ APLValueCompareCommon }

func (value *APLValueCompareFloatLt) GetBool(sim *Simulation) bool {
	return value.lhs.GetFloat(sim) < value.rhs.GetFloat(sim)
}
func (value *APLValueCompareFloatLt) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpLt
}

type APLValueCompareFloatLe struct{ APLValueCompareCommon }

func (value *APLValueCompareFloatLe) GetBool(sim *Simulation) bool {
	return value.lhs.GetFloat(sim) <= value.rhs.GetFloat(sim)
}
func (value *APLValueCompareFloatLe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpLe
}

type APLValueCompareFloatGt struct{ APLValueCompareCommon }

func (value *APLValueCompareFloatGt) GetBool(sim *Simulation) bool {
	return value.lhs.GetFloat(sim) > value.rhs.GetFloat(sim)
}
func (value *APLValueCompareFloatGt) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpGt
}

type APLValueCompareFloatGe struct{ APLValueCompareCommon }

func (value *APLValueCompareFloatGe) GetBool(sim *Simulation) bool {
	return value.lhs.GetFloat(sim) >= value.rhs.GetFloat(sim)
}
func (value *APLValueCompareFloatGe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpGe
}

type APLValueCompareDurationEq struct{ APLValueCompareCommon }

func (value *APLValueCompareDurationEq) GetBool(sim *Simulation) bool {
	return value.lhs.GetDuration(sim) == value.rhs.GetDuration(sim)
}
func (value *APLValueCompareDurationEq) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpEq
}

type APLValueCompareDurationNe struct{ APLValueCompareCommon }

func (value *APLValueCompareDurationNe) GetBool(sim *Simulation) bool {
	return value.lhs.GetDuration(sim) != value.rhs.GetDuration(sim)
}

func (value *APLValueCompareDurationNe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpNe
}

type APLValueCompareDurationLt struct{ APLValueCompareCommon }

func (value *APLValueCompareDurationLt) GetBool(sim *Simulation) bool {
	return value.lhs.GetDuration(sim) < value.rhs.GetDuration(sim)
}
func (value *APLValueCompareDurationLt) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpLt
}

type APLValueCompareDurationLe struct{ APLValueCompareCommon }

func (value *APLValueCompareDurationLe) GetBool(sim *Simulation) bool {
	return value.lhs.GetDuration(sim) <= value.rhs.GetDuration(sim)
}
func (value *APLValueCompareDurationLe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpLe
}

type APLValueCompareDurationGt struct{ APLValueCompareCommon }

func (value *APLValueCompareDurationGt) GetBool(sim *Simulation) bool {
	return value.lhs.GetDuration(sim) > value.rhs.GetDuration(sim)
}
func (value *APLValueCompareDurationGt) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpGt
}

type APLValueCompareDurationGe struct{ APLValueCompareCommon }

func (value *APLValueCompareDurationGe) GetBool(sim *Simulation) bool {
	return value.lhs.GetDuration(sim) >= value.rhs.GetDuration(sim)
}
func (value *APLValueCompareDurationGe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpGe
}

type APLValueCompareStringEq struct{ APLValueCompareCommon }

func (value *APLValueCompareStringEq) GetBool(sim *Simulation) bool {
	return value.lhs.GetString(sim) == value.rhs.GetString(sim)
}
func (value *APLValueCompareStringEq) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpEq
}

type APLValueCompareStringNe struct{ APLValueCompareCommon }

func (value *APLValueCompareStringNe) GetBool(sim *Simulation) bool {
	return value.lhs.GetString(sim) != value.rhs.GetString(sim)
}
func (value *APLValueCompareStringNe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpNe
}

type APLValueCompareStringLt struct{ APLValueCompareCommon }

func (value *APLValueCompareStringLt) GetBool(sim *Simulation) bool {
	return value.lhs.GetString(sim) < value.rhs.GetString(sim)
}
func (value *APLValueCompareStringLt) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpLt
}

type APLValueCompareStringLe struct{ APLValueCompareCommon }

func (value *APLValueCompareStringLe) GetBool(sim *Simulation) bool {
	return value.lhs.GetString(sim) <= value.rhs.GetString(sim)
}
func (value *APLValueCompareStringLe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpLe
}

type APLValueCompareStringGt struct{ APLValueCompareCommon }

func (value *APLValueCompareStringGt) GetBool(sim *Simulation) bool {
	return value.lhs.GetString(sim) > value.rhs.GetString(sim)
}
func (value *APLValueCompareStringGt) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpGt
}

type APLValueCompareStringGe struct{ APLValueCompareCommon }

func (value *APLValueCompareStringGe) GetBool(sim *Simulation) bool {
	return value.lhs.GetString(sim) >= value.rhs.GetString(sim)
}

func (value *APLValueCompareStringGe) Op() proto.APLValueCompare_ComparisonOperator {
	return proto.APLValueCompare_OpGe
}
