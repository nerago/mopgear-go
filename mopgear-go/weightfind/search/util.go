package weightfind

import (
	"math"
)

func rangeIsMarginal(rangeMin *pointType, rangeMax *pointType, size int8) bool {
	for i := range size {
		if rangeMax[i]-rangeMin[i] > c_search_marginalWeightGap {
			return false
		}
	}
	return true
}

func largeAccuracyGap(a, b float64) bool {
	return math.Abs(a-b) >= c_search_largeAccuracyGap
}

func equalAccuracyGap(a, b float64) bool {
	return math.Abs(a-b) < c_search_equalAccuracyGap
}

func sliceInterpolate(rangeMin *pointType, rangeMax *pointType, ratio float64, out *pointType, size int8) {
	for i := range size {
		out[i] = valueInterpolate(rangeMin[i], rangeMax[i], ratio)
	}
}

func valueInterpolate(rangeMin float64, rangeMax float64, ratio float64) float64 {
	return rangeMin + (rangeMax-rangeMin)*ratio
}

func checkRangeIsSubrangeOf(outer, inner *boundType, size int8) bool {
	for i := range size {
		if outer.rangeMin[i] <= inner.rangeMin[i] && inner.rangeMax[i] <= outer.rangeMax[i] {
			// yes
		} else {
			return false
		}
	}
	return true
}
