package upgrades

import (
	"math"
	"paladin_gearing_go/items"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"strconv"
)

const (
	c_upgradeEachThreads = 4
	c_simThreads         = 4
	c_targetUpgradeLevel = 2
	c_baseSolveScale     = 4
)

// ################## upgradeItemTask ##################

type upgradeItemTask struct {
	item *items.FullItem
	slot items.SlotEquip
	goal UpgradeGoal
	boss string
}

func (task upgradeItemTask) Item() *items.FullItem {
	return task.item
}

func (task upgradeItemTask) Slot() items.SlotEquip {
	return task.slot
}

func (task upgradeItemTask) Goal() UpgradeGoal {
	return task.goal
}

func (task upgradeItemTask) Boss() string {
	return task.boss
}

// ################## upgradeItemResult ##################

type upgradeItemResult struct {
	upgradeItemTask
	success  bool
	itemSet  *items.FullItemSet
	setBonus uint8
	factor   float64
}

func (result upgradeItemResult) ranking() float64 {
	return result.increase()
}

func (result upgradeItemResult) increase() float64 {
	return factorToIncrease(result.factor)
}

func (result upgradeItemResult) increaseStr() string {
	if result.factor == 0 || result.factor == -1 {
		return ""
	}
	percent := result.increase()
	return formatIncrease(percent)
}

func factorToIncrease(factor float64) float64 {
	return (factor - 1.0) * 100
}

func ratioToIncrease(sim, baseSim *simulate.SimResultStats, part simulate.SimResultType) float64 {
	newValue := sim.Get(part)
	baseValue := baseSim.Get(part)

	var result float64
	if part == simulate.Result_DEATH {
		result = baseValue - newValue
	} else if part.IsHighGood() {
		result = factorToIncrease(newValue / baseValue)
	} else {
		result = factorToIncrease(baseValue / newValue)
	}

	if math.IsNaN(result) {
		panic("unexpected NaN")
	}
	return result
}

func formatIncrease(percent float64) string {
	if math.IsNaN(percent) {
		panic("unexpected NaN")
	}

	str := strconv.FormatFloat(percent, 'f', 2, 64)
	if percent < 0 {
		return str + "%"
	} else {
		return "+" + str + "%"
	}
}

// ################## upgradeItemResultWithSim ##################

type upgradeItemResultWithSim struct {
	upgradeItemResult
	baseSim simulate.SimResultStats
	sim     simulate.SimResultStats
}

func (result upgradeItemResultWithSim) percentSim() float64 {
	if result.sim.IsEmpty() {
		return -100.0
	}

	switch result.goal {
	case UpgradeGoal_Dps:
		return ratioToIncrease(&result.sim, &result.baseSim, simulate.Result_DPS)
	case UpgradeGoal_Healing:
		return ratioToIncrease(&result.sim, &result.baseSim, simulate.Result_HPS)
	case UpgradeGoal_Mitigation:
		checkParts := []simulate.SimResultType{simulate.Result_DPS, simulate.Result_DTPS, simulate.Result_TMI, simulate.Result_DEATH}
		var total float64
		for _, part := range checkParts {
			total += ratioToIncrease(&result.sim, &result.baseSim, part)
		}
		return total / float64(len(checkParts))
	default:
		panic("unknown goal")
	}
}

func (result upgradeItemResultWithSim) percentStrSim() string {
	if result.sim.IsEmpty() {
		return ""
	}

	return formatIncrease(result.percentSim())
}

func (result upgradeItemResultWithSim) increaseSimBreakdown() simulate.SimResultStats {
	if result.sim.IsEmpty() {
		panic("empty sim shouldn't get called here")
	}

	sim := simulate.SimResultStats{}
	for _, resultType := range simulate.SimResultTypeList {
		sim.Set(resultType, ratioToIncrease(&result.sim, &result.baseSim, resultType))
	}
	return sim
}

func (result upgradeItemResultWithSim) bestOfSimResults() float64 {
	var best float64 = -100.0
	if !result.sim.IsEmpty() {
		for _, resultType := range simulate.SimResultTypeList {
			increase := ratioToIncrease(&result.sim, &result.baseSim, resultType)
			best = math.Max(best, increase) // TODO use nan friendly?
		}
	}
	return best
}

func (result upgradeItemResultWithSim) ranking() float64 {
	return result.percentSim()
}

// ################## reportGroup ##################

type reportGroup struct {
	specLabel  string
	difficulty stats.Difficulty
}

func (group reportGroup) String() string {
	return group.specLabel + "-" + group.difficulty.Name()
}

// ################## reportForItem ##################

type reportForItem struct {
	item    *items.FullItem
	grouped map[reportGroup]upgradeItemResult
}

func makeReportForItem(mapSize int) func(*items.FullItem) *reportForItem {
	return func(item *items.FullItem) *reportForItem {
		report := new(reportForItem)
		report.item = item
		report.grouped = make(map[reportGroup]upgradeItemResult, mapSize)
		return report
	}
}

func (report *reportForItem) Add(mode reportGroup, result upgradeItemResult) {
	report.grouped[mode] = result
}

func (report *reportForItem) BestRating() float64 {
	var best float64 = -100.0
	for _, item := range report.grouped {
		best = math.Max(best, item.factor)
	}
	return best
}

// ################## reportForItemWithSim ##################

type reportForItemWithSim struct {
	item    *items.FullItem
	grouped map[reportGroup]upgradeItemResultWithSim
}

func makeReportSimForItem(mapSize int) func(*items.FullItem) *reportForItemWithSim {
	return func(item *items.FullItem) *reportForItemWithSim {
		report := new(reportForItemWithSim)
		report.item = item
		report.grouped = make(map[reportGroup]upgradeItemResultWithSim, mapSize)
		return report
	}
}

func (report *reportForItemWithSim) Add(mode reportGroup, result upgradeItemResultWithSim) {
	report.grouped[mode] = result
}

func (report *reportForItemWithSim) BestRating() float64 {
	var best float64 = -100.0
	for _, item := range report.grouped {
		best = math.Max(best, math.Max(item.percentSim(), item.increase()))
	}
	return best
}

// ################## reportable ##################

type reportable interface {
	Boss() string
	Item() *items.FullItem
}

type reportForItemAddable[T reportable] interface {
	Add(mode reportGroup, result T)
}
