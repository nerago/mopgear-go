package util

import "math"

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64
}
type NumberInt interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func MaxIgnoreNaN(a, b float64) float64 {
	if math.IsNaN(a) {
		return b
	} else if math.IsNaN(b) {
		return a
	} else {
		return math.Max(a, b)
	}
}

func MaxIgnoreNaN3(a, b, c float64) float64 {
	if math.IsNaN(a) && math.IsNaN(b) {
		return c
	} else if math.IsNaN(a) && math.IsNaN(c) {
		return b
	} else if math.IsNaN(b) && math.IsNaN(c) {
		return a
	} else if math.IsNaN(a) {
		return math.Max(b, c)
	} else if math.IsNaN(b) {
		return math.Max(a, c)
	} else if math.IsNaN(c) {
		return math.Max(a, b)
	} else {
		return math.Max(math.Max(a, b), c)
	}
}

func AbsDiff[N Number](a, b N) N {
	if a > b {
		return a - b
	} else {
		return b - a
	}
}

func Clamp[N Number](value, min, max N) N {
	if value < min {
		return min
	} else if value > max {
		return max
	} else {
		return value
	}
}

const c_float_equal_delta = 0.000001
const c_float_equal_delta_lenient = 0.0001

func FloatEqualsOne(value float64) bool {
	return math.Abs(value-1.0) < c_float_equal_delta
}

func FloatEqualsZero(value float64) bool {
	return math.Abs(value) < c_float_equal_delta
}

func FloatNonZero(value float64) bool {
	return math.Abs(value) >= c_float_equal_delta
}

func FloatsApproxEquals(a, b float64) bool {
	if b != 0 {
		ratio := a / b
		if FloatEqualsOne(ratio) {
			return true
		}
	}
	diff := a - b
	return FloatEqualsZero(diff)
}

func FloatsApproxEqualsLenient(a, b float64) bool {
	if b != 0 {
		ratio := a / b
		if math.Abs(ratio-1.0) < c_float_equal_delta_lenient {
			return true
		}
	}
	diff := a - b
	return math.Abs(diff) < c_float_equal_delta_lenient
}

func FloatsApproxEqualsFast(a, b float64) bool {
	diff := a - b
	return FloatEqualsZero(diff)
}

func FloatsBetween(lo, val, hi float64) bool {
	return lo-c_float_equal_delta <= val && val <= hi+c_float_equal_delta
}

func FloatApproxLessThanOrEqual(lower float64, higher float64) bool {
	return lower <= higher+c_float_equal_delta
}

func IntBetweenInclusive[N NumberInt](lo, val, hi N) bool {
	return lo <= val && val <= hi
}

func RoundToInt64(value float64) int64 {
	return int64(math.Round(value))
}

func RoundToInt(value float64) int {
	return int(math.Round(value))
}
