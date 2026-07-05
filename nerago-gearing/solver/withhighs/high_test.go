package withhighs

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_test"
	"testing"
)

func TestSolverBasicRun(t *testing.T) {
	const targetCount = util_test.TargetCountStandard
	options, model := util_test.MakeTestOptions()

	resultFuture := SingleGearSetMain(options, model, util.PrintRecorder_Testing(t))
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
