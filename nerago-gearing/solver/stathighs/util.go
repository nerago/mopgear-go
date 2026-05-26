package stathighs

import "math"

func isGoodValueRange(value float64) bool {
	value = math.Abs(value)
	return 1e-6 <= value && value <= 1e6
}
