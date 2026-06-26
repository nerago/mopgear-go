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
