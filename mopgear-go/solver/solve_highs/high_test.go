package solve_highs

import (
	"testing"

	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/solver/solve_highs_types"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_test"
)

func TestSolverBasicRun(t *testing.T) {
	const targetCount = util_test.TargetCountStandard
	options, model := util_test.MakeTestOptions()
	solveModel := solve_highs_types.SolverModelBuild(model, 1, nil)

	resultFuture := SingleGearSetMain(options, solveModel, util.PrintRecorder_Testing(t), 60)
	result := resultFuture.WaitForResultAsOptional()

	expectEquip := util_test.MakeTestExpectedBest()
	expectSet := items.SolvableItemSet_Of(expectEquip)

	if result.HasValue() {
		resultSet := result.GetOrPanic()
		if !equals(expectSet, resultSet) {
			t.Fatalf("set not equal")
		}
	} else {
		t.Fatalf("no solution")
	}
}

func equals(expectSet, resultSet items.SolvableItemSet) bool {
	if !stats.StatBlock_Equals(expectSet.Total(), resultSet.Total()) {
		return false
	}
	for i := range expectSet.Items() {
		e := expectSet.Items()[i]
		r := resultSet.Items()[i]
		if (e != nil && r != nil && *e != *r) || ((e == nil) != (r == nil)) {
			return false
		}
	}
	return true
}
