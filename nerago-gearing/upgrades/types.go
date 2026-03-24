package upgrades

import (
	"math"
	"paladin_gearing_go/items"
	"paladin_gearing_go/simulate"
	"strconv"
)

var ignoredItems = []items.ItemId{
	63207, // org port cloak
	84661, // fishing pole
	90042} // straw hat

type upgradeMode int8

const (
	Upgrade_Miti_Normal upgradeMode = iota
	Upgrade_Miti_Heroic upgradeMode = iota
	Upgrade_Dps_Normal  upgradeMode = iota
	Upgrade_Dps_Heroic  upgradeMode = iota
)

func (up upgradeMode) Name() string {
	switch up {
	case Upgrade_Miti_Normal:
		return "mit_norm"
	case Upgrade_Miti_Heroic:
		return "mit_hero"
	case Upgrade_Dps_Normal:
		return "dps_norm"
	case Upgrade_Dps_Heroic:
		return "dps_hero"
	default:
		panic("unknown")
	}
}

type upgradeItemTask struct {
	item *items.FullItem
	slot items.SlotEquip
	mode upgradeMode
	boss string
}

func (task upgradeItemTask) Item() *items.FullItem {
	return task.item
}

func (task upgradeItemTask) Slot() items.SlotEquip {
	return task.slot
}

func (task upgradeItemTask) Mode() upgradeMode {
	return task.mode
}

func (task upgradeItemTask) ModeIsDamage() bool {
	return task.mode == Upgrade_Dps_Normal || task.mode == Upgrade_Dps_Heroic
}

func (task upgradeItemTask) Boss() string {
	return task.boss
}

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
	if part == simulate.Result_DEATH {
		return baseValue - newValue
	} else if part.IsHighGood() {
		return factorToIncrease(newValue / baseValue)
	} else {
		return factorToIncrease(baseValue / newValue)
	}
}

func formatIncrease(percent float64) string {
	if percent <= -1 {
		return ""
	}

	str := strconv.FormatFloat(percent, 'f', 2, 64)
	if percent < 0 {
		return str + "%"
	} else {
		return "+" + str + "%"
	}
}

type upgradeItemResultWithSim struct {
	upgradeItemResult
	baseSim simulate.SimResultStats
	sim     simulate.SimResultStats
}

func (result upgradeItemResultWithSim) percentSim() float64 {
	if result.ModeIsDamage() {
		return ratioToIncrease(&result.sim, &result.baseSim, simulate.Result_DPS)
	} else {
		checkParts := []simulate.SimResultType{simulate.Result_DPS, simulate.Result_DTPS, simulate.Result_TMI, simulate.Result_DEATH}
		var total float64
		for _, part := range checkParts {
			total += ratioToIncrease(&result.sim, &result.baseSim, part)
		}
		return total / float64(len(checkParts))
	}
}

func (result upgradeItemResultWithSim) percentStrSim() string {
	return formatIncrease(result.percentSim())
}

func (result upgradeItemResultWithSim) increaseSimBreakdown() simulate.SimResultStats {
	sim := simulate.SimResultStats{}
	for _, resultType := range simulate.SimResultTypeList {
		sim.Set(resultType, ratioToIncrease(&result.sim, &result.baseSim, resultType))
	}
	return sim
}

func (result upgradeItemResultWithSim) bestOfSimResults() float64 {
	var best float64 = -1.0
	for _, resultType := range simulate.SimResultTypeList {
		increase := ratioToIncrease(&result.sim, &result.baseSim, resultType)
		best = math.Max(best, increase)
	}
	return best
}

func (result upgradeItemResultWithSim) ranking() float64 {
	return result.percentSim()
}
