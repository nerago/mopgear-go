package util_highs

import (
	"paladin_gearing_go/util"

	"github.com/bartolsthoorn/gohighs/highs"
)

type DebugContext interface {
	DebugText() string
}

type DebugString struct {
	Text string
}

func (debugString DebugString) DebugText() string {
	return debugString.Text
}

func debugText(debug DebugContext) string {
	debugText := ""
	if debug != nil {
		debugText = debug.DebugText()
	}
	return debugText
}

func DebugText(text string) DebugString {
	return DebugString{Text: text}
}

func (build *LinearBuilder) DebugPrintColumns(solution *highs.Solution, printer *util.PrintRecorder) {
	if C_DebugHighs {
		build.debugPrintColumnsForce(solution, printer)
	}
}
func (build *LinearBuilder) debugPrintColumnsForce(solution *highs.Solution, printer *util.PrintRecorder) {
	for i, x := range solution.ColValues {
		printer.Printf("%6d %10.6f %12.4e %s\n", i, x, x, debugText(build.vars.debug[i]))
	}
}

func (build *LinearBuilder) DebugTextFor(columnIndex ColumnIndex) string {
	return debugText(build.vars.debug[columnIndex])
}
