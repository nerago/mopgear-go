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
	Upgrade_Miti_Heroic             = iota
	Upgrade_Dps_Normal              = iota
	Upgrade_Dps_Heroic              = iota
)

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

func (task upgradeItemTask) Boss() string {
	return task.boss
}

type upgradeItemResult struct {
	upgradeItemTask
	itemSet *items.FullItemSet
	factor  float64
}

func (result upgradeItemResult) percent() float64 {
	return factorToIncrease(result.factor)
}

func (result upgradeItemResult) percentStr() string {
	if result.factor == 0 {
		return ""
	}
	percent := result.percent()
	return formatPercent(percent)
}

func factorToIncrease(factor float64) float64 {
	return (factor - 1.0) * 100
}

func ratioToIncrease(newValue float64, base float64, higherIsGood bool) float64 {
	if higherIsGood {
		return factorToIncrease(newValue / base)
	} else {
		return factorToIncrease(base / newValue)
	}
}

func formatPercent(percent float64) string {
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
	sim simulate.SimResultStats
}

func (result upgradeItemResultWithSim) percentSim() float64 {
	if result.mode == Upgrade_Dps_Normal || result.mode == Upgrade_Dps_Heroic {
		return ratioToIncrease(result.sim.DPS, result.baseSim.DPS, simulate.Result_DPS.IsHighGood())
	} else {
		// TODO represent these comparisons in more detail
		checkParts := []simulate.SimResultType {simulate.Result_DPS, simulate.Result_DTPS, simulate.Result_TMI, simulate.Result_DEATH}
		var total float64
		for _, part := range checkParts {
			// TODO special handling death
			total += ratioToIncrease(result.sim.Get(part), result.baseSim.Get(part), part.IsHighGood())
		}
		return total / float64(len(checkParts))
	}
}

func (result upgradeItemResultWithSim) percentStrSim() string {
	return formatPercent(result.percentSim())
}

func (result upgradeItemResultWithSim) percentStrSimBreakdown() simulate.SimResultStats {
	sim := simulate.SimResultStats{}
	for _, resultType := range simulate.SimResultTypeList {
		sim.Set(resultType, ratioToIncrease(result.sim.Get(resultType), result.baseSim.Get(resultType), resultType.IsHighGood()))
	}
	return sim
}

func (result upgradeItemResultWithSim) bestOfSimResults() float64 {
	var best float64 = -1.0
	for _, resultType := range simulate.SimResultTypeList {
		increase := ratioToIncrease(result.sim.Get(resultType), result.baseSim.Get(resultType), resultType.IsHighGood())
		best = math.Max(best, increase)
	}
	return best
}