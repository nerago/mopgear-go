package upgrades

import (
	"math"
	"paladin_gearing_go/items"
	"paladin_gearing_go/simulate"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
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
	item       *items.FullItem
	slot       items.SlotEquip
	goal       UpgradeGoal
	boss       string
	canUpgrade CanUpgradeResult
}

func (task upgradeItemTask) actuallyAttemptUpgrade() bool {
	return task.canUpgrade == CanUpgrade_Yes || task.canUpgrade == CanUpgrade_AvailableInBags
}

func (task upgradeItemTask) Equals(other upgradeItemTask) bool {
	return task.item.Equals(other.item) &&
		task.slot == other.slot &&
		task.goal == other.goal &&
		task.boss == other.boss
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

func upgradeItemResult_OfFailure(task *upgradeItemTask) upgradeItemResult {
	return upgradeItemResult{upgradeItemTask: *task, success: false, factor: -1}
}

func (result upgradeItemResult) Equals(other upgradeItemResult) bool {
	return result.upgradeItemTask.Equals(other.upgradeItemTask) &&
		result.success == other.success &&
		result.itemSet.EqualsAllowNil(other.itemSet) &&
		result.setBonus == other.setBonus &&
		result.factor == other.factor
}

func (result upgradeItemResult) ranking() float64 {
	return result.increase()
}

func (result upgradeItemResult) increase() float64 {
	return factorToIncrease(result.factor)
}

func (result upgradeItemResult) increaseStr() string {
	if result.factor == 0 || result.factor == -1 {
		if result.canUpgrade != CanUpgrade_Yes {
			return result.canUpgrade.Text()
		}
		return ""
	}
	percent := result.increase()
	str := formatIncrease(percent)
	if result.canUpgrade == CanUpgrade_AvailableInBags {
		str = "* " + str
	}
	return str
}

func factorToIncrease(factor float64) float64 {
	return (factor - 1.0) * 100
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

func (result upgradeItemResultWithSim) Equals(other upgradeItemResultWithSim) bool {
	return result.upgradeItemResult.Equals(other.upgradeItemResult) &&
		result.baseSim == other.baseSim && result.sim == other.sim
}

func (result upgradeItemResultWithSim) percentSim() float64 {
	if result.sim.IsEmpty() {
		return -100.0
	}

	switch result.goal {
	case UpgradeGoal_Dps:
		return result.sim.IncreaseOf(&result.baseSim, simulate.Result_DPS)
	case UpgradeGoal_Healing:
		return result.sim.IncreaseOf(&result.baseSim, simulate.Result_HPS)
	case UpgradeGoal_Mitigation:
		return result.sim.IncreaseMitigation(&result.baseSim)
	default:
		panic("unknown goal")
	}
}

func (result upgradeItemResultWithSim) percentStrSim() string {
	if result.sim.IsEmpty() {
		if result.canUpgrade != CanUpgrade_Yes {
			return result.canUpgrade.Text()
		}
		return ""
	}

	return formatIncrease(result.percentSim())
}

func (result upgradeItemResultWithSim) percentStrSim_OnlyShown() string {
	if result.sim.IsEmpty() {
		if result.canUpgrade != CanUpgrade_Yes {
			return result.canUpgrade.Text()
		}
		return ""
	}

	str := formatIncrease(result.percentSim())
	if result.canUpgrade == CanUpgrade_AvailableInBags {
		str = "* " + str
	}
	return str
}

func (result upgradeItemResultWithSim) increaseSimBreakdown() simulate.SimResultStats {
	return result.sim.IncreaseSimBreakdown(&result.baseSim)
}

func (result upgradeItemResultWithSim) bestOfSimResults() float64 {
	var best float64 = -100.0
	if !result.sim.IsEmpty() {
		best = result.sim.BestIncrease(&result.baseSim)
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

// ################## reportForItemBasic ##################

type reportForItemBasic struct {
	item    *items.FullItem
	slot    items.SlotEquip
	grouped map[string]upgradeItemResult
}

func makeReportForItem(mapSize int) func(*items.FullItem, items.SlotEquip) *reportForItemBasic {
	return func(item *items.FullItem, slot items.SlotEquip) *reportForItemBasic {
		report := new(reportForItemBasic)
		report.item = item
		report.slot = slot
		report.grouped = make(map[string]upgradeItemResult, mapSize)
		return report
	}
}

func (report *reportForItemBasic) Add(group reportGroup, result upgradeItemResult) {
	old, exists := report.grouped[group.specLabel]
	if exists && !old.Equals(result) {
		panic("duplicate group entry for spec with " + result.item.CreateString())
	}
	report.grouped[group.specLabel] = result
}

func (report *reportForItemBasic) BestRating() float64 {
	var best float64 = -100.0
	for _, item := range report.grouped {
		best = util.MaxIgnoreNaN(best, item.factor)
	}
	return best
}

// ################## reportForItemWithSim ##################

type reportForItemWithSim struct {
	item    *items.FullItem
	slot    items.SlotEquip
	grouped map[string]upgradeItemResultWithSim
}

func makeReportSimForItem(mapSize int) func(*items.FullItem, items.SlotEquip) *reportForItemWithSim {
	return func(item *items.FullItem, slot items.SlotEquip) *reportForItemWithSim {
		report := new(reportForItemWithSim)
		report.item = item
		report.slot = slot
		report.grouped = make(map[string]upgradeItemResultWithSim, mapSize)
		return report
	}
}

func (report *reportForItemWithSim) Add(group reportGroup, result upgradeItemResultWithSim) {
	old, exists := report.grouped[group.specLabel]
	if exists && !old.Equals(result) {
		panic("duplicate group entry for spec with " + result.item.CreateString())
	}
	report.grouped[group.specLabel] = result
}

func (report *reportForItemWithSim) BestRating() float64 {
	var best float64 = -100.0
	for _, item := range report.grouped {
		best = util.MaxIgnoreNaN3(best, item.percentSim(), item.increase())
	}
	return best
}

func (report *reportForItemWithSim) BestRating_NoWeight() float64 {
	var best float64 = -100.0
	for _, item := range report.grouped {
		best = util.MaxIgnoreNaN(best, item.percentSim())
	}
	return best
}

// ################## reportable ##################

type reportable interface {
	Boss() string
	Item() *items.FullItem
	Slot() items.SlotEquip
}

type reportForItemAddable[T reportable] interface {
	Add(group reportGroup, result T)
}

type reportItemRef struct {
	itemId items.ItemId
	slot   items.SlotEquip
}
