package util

import "math"

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

func AbsIntDiff(a, b int) int {
	if a > b {
		return a - b
	} else {
		return b - a
	}
}

func AbsInt64Diff(a, b int64) int64 {
	if a > b {
		return a - b
	} else {
		return b - a
	}
}

func Clamp(value, min, max float64) float64 {
	if value < min {
		return min
	} else if value > max {
		return max
	} else {
		return value
	}
}

func FloatEqualsOne(value float64) bool {
	return 0.999999 <= value && value <= 1.000001
}

func FloatEqualsZero(value float64) bool {
	return -0.000001 <= value && value <= 0.000001
}

func FloatsApproxEquals(a, b float64) bool {
	if b != 0 {
		ratio := a / b
		return (0.99999 <= ratio && ratio <= 1.00001) || (math.Abs(a-b) < 0.0000001)
	} else {
		return FloatEqualsZero(a)
	}
}

func FloatsBetween(lo, val, hi float64) bool {
	return lo-0.000001 <= val && val <= hi+0.000001
}
