package util_highs

import (
	"iter"
	"math"
	"paladin_gearing_go/util"

	"github.com/bartolsthoorn/gohighs/highs"
)

type ISolution interface {
	GetValue(index ColumnIndex) float64
	GetValueUInt32(index ColumnIndex) uint32
	ValueIsZero(index ColumnIndex) bool
	ValueIsOne(index ColumnIndex) bool
	Objective() float64
	ColValuesSeq() iter.Seq2[ColumnIndex, float64]
}

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

func (sol Solution2) Status() highs.ModelStatus {
	return sol.inner.Status
}

func (sol Solution2) HasSolution() bool {
	return sol.inner.Status.HasSolution()
}

func (sol Solution2) DebugPrint(printer *util.PrintRecorder) {
	sol.build.DebugPrintColumns(sol.inner, printer)
}

func (sol Solution2) ColValuesSeq() iter.Seq2[ColumnIndex, float64] {
	return func(yield func(ColumnIndex, float64) bool) {
		for i, v := range sol.inner.ColValues {
			if !yield(ColumnIndex(i), v) {
				return
			}
		}
	}
}

type InterimSolution struct {
	ColValues      []float64
	ObjectiveValue float64
	MipPrimal      float64
	MipDual        float64
	MipGap         float64
}

func InterimSolutionFromCallback(out highs.HighsCallbackDataOut) InterimSolution {
	interim := InterimSolution{
		ObjectiveValue: out.Objective_function_value,
		MipPrimal:      out.Mip_primal_bound,
		MipDual:        out.Mip_dual_bound,
		MipGap:         out.Mip_gap,
		ColValues:      out.Mip_solution,
	}
	return interim
}

func (is InterimSolution) GetValue(index ColumnIndex) float64 {
	return is.ColValues[index]
}

func (is InterimSolution) GetValueUInt32(index ColumnIndex) uint32 {
	return uint32(math.Round(is.ColValues[index]))
}

func (is InterimSolution) ValueIsZero(index ColumnIndex) bool {
	return util.FloatEqualsZero(is.ColValues[index])
}

func (is InterimSolution) ValueIsOne(index ColumnIndex) bool {
	return util.FloatEqualsOne(is.ColValues[index])
}

func (is InterimSolution) Objective() float64 {
	return is.ObjectiveValue
}

func (is InterimSolution) ColValuesSeq() iter.Seq2[ColumnIndex, float64] {
	return func(yield func(ColumnIndex, float64) bool) {
		for i, v := range is.ColValues {
			if !yield(ColumnIndex(i), v) {
				return
			}
		}
	}
}
