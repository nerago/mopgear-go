package upgrades

import (
	"math"
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"strconv"
)

const (
	c_upgradeEachThreads = 4
	c_simThreads         = 4
	c_baseSolveScale     = 4
	c_nullIncrease       = -100.0
)

func formatIncreaseGeneric(percent float64) string {
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

// ################## upgradeItemTask ##################

type upgradeItemTask struct {
	item       *items.FullItem
	slot       items.SlotEquip
	goal       stats.OptimiseGoal
	boss       string
	canUpgrade items.CanUpgradeResult
}

func (task upgradeItemTask) actuallyAttemptUpgrade(forceIncludeMost bool) bool {
	if forceIncludeMost {
		return task.canUpgrade != items.CanUpgrade_InvalidAlways
	} else {
		return task.canUpgrade == items.CanUpgrade_Yes || task.canUpgrade == items.CanUpgrade_AvailableInBags
	}
}

func (task upgradeItemTask) Equals(other upgradeItemTask) bool {
	return task.item.Equals(other.item) &&
		task.slot == other.slot &&
		task.goal == other.goal &&
		task.boss == other.boss
}

// ################## upgradeItemResult ##################

type upgradeItemResult struct {
	upgradeItemTask
	success  bool
	itemSet  *items.FullItemSet
	setBonus uint8
	factor   util.Optional[float64]
}

func upgradeItemResult_OfFailure(task *upgradeItemTask) upgradeItemResult {
	return upgradeItemResult{upgradeItemTask: *task, success: false, factor: util.Optional_Empty[float64]()}
}

func (result upgradeItemResult) Equals(other upgradeItemResult) bool {
	return result.upgradeItemTask.Equals(other.upgradeItemTask) &&
		result.success == other.success &&
		result.itemSet.EqualsAllowNil(other.itemSet) &&
		result.setBonus == other.setBonus &&
		result.factor == other.factor
}

func (result upgradeItemResult) increaseWeightsRaw() float64 {
	if result.factor.IsEmpty() {
		return c_nullIncrease
	}
	return (result.factor.GetOrPanic() - 1.0) * 100
}

func (result upgradeItemResult) increaseWeightsStr(prefixNote bool) string {
	if result.factor.IsEmpty() {
		if result.canUpgrade != items.CanUpgrade_Yes {
			return result.canUpgrade.TextLong()
		}
		return ""
	}

	percent := (result.factor.GetOrPanic() - 1.0) * 100
	str := formatIncreaseGeneric(percent)

	if prefixNote && result.canUpgrade != items.CanUpgrade_Yes {
		str = result.makeNoteAbbrev() + " " + str
	}
	return str
}

func (result upgradeItemResult) makeNoteFull() string {
	return result.canUpgrade.TextLong()
}

func (result upgradeItemResult) makeNoteAbbrev() string {
	return result.canUpgrade.TextAbbrev()
}

// ################## upgradeItemResultWithSim ##################

type upgradeItemResultWithSim struct {
	upgradeItemResult
	baseSim stats.SimData
	sim     stats.SimData
}

func (result upgradeItemResultWithSim) Equals(other upgradeItemResultWithSim) bool {
	return result.upgradeItemResult.Equals(other.upgradeItemResult) &&
		result.baseSim == other.baseSim && result.sim == other.sim
}

func (result upgradeItemResultWithSim) increaseSim() float64 {
	if result.sim.IsEmpty() {
		return c_nullIncrease
	}

	switch result.goal {
	case stats.OptimiseGoal_Dps:
		return result.sim.IncreaseOf(&result.baseSim, stats.Sim_DPS)
	case stats.OptimiseGoal_Healing:
		return result.sim.IncreaseOf(&result.baseSim, stats.Sim_HPS)
	case stats.OptimiseGoal_Mitigation:
		return result.sim.IncreaseMitigation(&result.baseSim)
	case stats.OptimiseGoal_HalfMitiDps:
		return (result.sim.IncreaseMitigation(&result.baseSim) + result.sim.IncreaseOf(&result.baseSim, stats.Sim_DPS)) / 2.0
	case stats.OptimiseGoal_HalfMitiHeal:
		return (result.sim.IncreaseMitigation(&result.baseSim) + result.sim.IncreaseOf(&result.baseSim, stats.Sim_HPS)) / 2.0
	default:
		panic("unknown goal")
	}
}

func (result upgradeItemResultWithSim) increaseSimStr(prefixNote bool) string {
	if result.sim.IsEmpty() {
		if result.canUpgrade != items.CanUpgrade_Yes {
			return result.canUpgrade.TextLong()
		}
		return ""
	}

	str := formatIncreaseGeneric(result.increaseSim())
	if prefixNote && result.canUpgrade != items.CanUpgrade_Yes {
		str = result.makeNoteAbbrev() + " " + str
	}
	return str
}

func (result upgradeItemResultWithSim) increaseSimBreakdown() stats.SimData {
	return result.sim.IncreaseSimBreakdown(&result.baseSim)
}

// ################## reportGroup ##################

type reportGroup struct {
	specLabel  string
	difficulty stats.Difficulty
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
		if result.increaseSim() < old.increaseSim() {
			return
		}
		// panic("duplicate group entry for spec with " + result.item.CreateString())
	}
	report.grouped[group.specLabel] = result
}

func (report *reportForItemWithSim) BestRating() float64 {
	var best float64 = c_nullIncrease
	for _, item := range report.grouped {
		best = util.MaxIgnoreNaN3(best, item.increaseSim(), item.increaseWeightsRaw())
	}
	return best
}

func (report *reportForItemWithSim) BestRating_NoWeight() float64 {
	var best float64 = c_nullIncrease
	for _, item := range report.grouped {
		best = util.MaxIgnoreNaN(best, item.increaseSim())
	}
	return best
}

type reportItemRef struct {
	itemId items.ItemId
	slot   items.SlotEquip
}
