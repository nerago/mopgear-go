package util_highs

import (
	"math"
	"paladin_gearing_go/util"

	"github.com/bartolsthoorn/gohighs/highs"
)

type Solution2 struct {
	inner *highs.Solution
	build *LinearBuilder
}

func (sol Solution2) GetValue(index ColumnIndex) float64 {
	return sol.inner.ColValues[index]
}

func (sol Solution2) GetValueUInt32(index ColumnIndex) uint32 {
	return uint32(math.Round(sol.inner.ColValues[index]))
}

func (sol Solution2) ValueIsZero(index ColumnIndex) bool {
	return util.FloatEqualsZero(sol.inner.ColValues[index])
}

func (sol Solution2) ValueIsOne(index ColumnIndex) bool {
	return util.FloatEqualsOne(sol.inner.ColValues[index])
}

func (sol Solution2) Objective() float64 {
	return sol.inner.Objective
}

func (sol Solution2) HasSolution() bool {
	return sol.inner.Status.HasSolution()
}

func (sol Solution2) DebugPrint(printer *util.PrintRecorder) {
	sol.build.DebugPrintColumns(sol.inner, printer)
}
