package upgrades

import (
	"math"
	"strconv"

	"github.com/nerago/mopgear-go/db"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/loaders"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
)

const (
	c_upgradeEachThreads = 8
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
	boss       string
	itemRef    loaders.ItemFoundRef
	slot       items.SlotEquip
	goal       stats.OptimiseGoal
	canUpgrade items.CanUpgradeResult
}

func (task upgradeItemTask) Equals(other upgradeItemTask) bool {
	return task.itemRef.Equals(other.itemRef) &&
		task.slot == other.slot &&
		task.goal == other.goal &&
		task.boss == other.boss
}

// ################## upgradeItemResult ##################

type upgradeItemResult struct {
	upgradeItemTask
	success   bool
	fullItem  *items.FullItem
	itemSet   *items.FullItemSet
	setBonus  uint8
	factor    util_collection.Optional[float64]
	baseSim   stats.SimData
	simResult stats.SimData
}

func upgradeItemResult_OfFailure(task *upgradeItemTask, fullItem *items.FullItem) upgradeItemResult {
	return upgradeItemResult{upgradeItemTask: *task, success: false, fullItem: fullItem, factor: util_collection.Optional_Empty[float64]()}
}

func (result upgradeItemResult) Equals(other upgradeItemResult) bool {
	return result.upgradeItemTask.Equals(other.upgradeItemTask) &&
		result.success == other.success &&
		util.NilSafeEqualPointers(result.itemSet, other.itemSet, (*items.FullItemSet).Equals) &&
		result.setBonus == other.setBonus &&
		result.factor == other.factor &&
		result.baseSim.Equals(&other.baseSim) && result.simResult.Equals(&other.simResult)
}

func (result upgradeItemResult) ItemName() string {
	if result.fullItem != nil {
		return result.fullItem.BaseName()
	} else {
		return db.LookupItemNameByItemId(result.itemRef.ItemId)
	}
}

func (result upgradeItemResult) ItemLevel() uint16 {
	if result.fullItem != nil {
		return result.fullItem.ItemLevel()
	} else {
		return 0
	}
}

func (result upgradeItemResult) ItemStatSummary() string {
	fullItem := result.fullItem
	if fullItem == nil {
		fullItem = db.WowSimDB_LoadItemById(result.itemRef.ItemId, 0)
	}
	sb := util.StringBuild2{}
	if fullItem != nil {
		for statType, value := range fullItem.StatBase().SeqPairInt() {
			if value != 0 && statType.IsSecondary() {
				sb.WriteString(statType.Name())
				sb.WriteRune(' ')
			}
		}
		if sb.Len() > 0 {
			sb.Rewind(1)
		}
	}
	return sb.String()
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

func (result upgradeItemResult) increaseSim() float64 {
	if result.simResult.IsEmpty() {
		return c_nullIncrease
	}

	switch result.goal {
	case stats.OptimiseGoal_Dps:
		return result.simResult.QueryIncreaseOf(&result.baseSim, stats.Sim_DPS)
	case stats.OptimiseGoal_Healing:
		return result.simResult.QueryIncreaseOf(&result.baseSim, stats.Sim_HPS)
	case stats.OptimiseGoal_Mitigation:
		return result.simResult.QueryIncreaseMitigation(&result.baseSim)
	case stats.OptimiseGoal_HalfMitiDps:
		return (result.simResult.QueryIncreaseMitigation(&result.baseSim) + result.simResult.QueryIncreaseOf(&result.baseSim, stats.Sim_DPS)) / 2.0
	case stats.OptimiseGoal_HalfMitiHeal:
		return (result.simResult.QueryIncreaseMitigation(&result.baseSim) + result.simResult.QueryIncreaseOf(&result.baseSim, stats.Sim_HPS)) / 2.0
	default:
		panic("unknown goal")
	}
}

func (result upgradeItemResult) increaseSimStr(prefixNote bool) string {
	if result.simResult.IsEmpty() {
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

func (result upgradeItemResult) increaseSimBreakdown() *stats.SimData {
	return result.simResult.QueryIncreaseOfEach(&result.baseSim)
}

// ################## reportGroup ##################

type reportGroup struct {
	specLabel  string
	difficulty stats.Difficulty
}
type upgradeGroupTask struct {
	spec       *SpecInput
	difficulty stats.Difficulty
}
type upgradeGroupResult struct {
	task       upgradeGroupTask
	group      reportGroup
	resultList []upgradeItemResult
}

// ################## reportForItemWithSim ##################

type reportForItem struct {
	grouped     map[string]upgradeItemResult
	itemName    string
	boss        string
	statSummary string
	itemLevel   uint16
	slot        items.SlotEquip
}

func (report *reportForItem) Add(group reportGroup, result upgradeItemResult) {
	old, exists := report.grouped[group.specLabel]
	if exists && !old.Equals(result) {
		if result.increaseSim() < old.increaseSim() {
			return
		}
		// panic("duplicate group entry for spec with " + result.item.CreateString())
	}
	report.grouped[group.specLabel] = result
}

func (report *reportForItem) BestRating() float64 {
	best := c_nullIncrease
	for _, item := range report.grouped {
		best = util.MaxIgnoreNaN3(best, item.increaseSim(), item.increaseWeightsRaw())
	}
	return best
}

func (report *reportForItem) BestRating_NoWeight() float64 {
	best := c_nullIncrease
	for _, item := range report.grouped {
		best = util.MaxIgnoreNaN(best, item.increaseSim())
	}
	return best
}

type reportItemRef struct {
	itemId items.ItemId
	slot   items.SlotEquip
}
