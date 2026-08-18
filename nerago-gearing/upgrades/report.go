package upgrades

import (
	"github.com/nerago/mopgear-go/db"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_rank"
	"strconv"
)

func reportBasicResultsSim(resultList []upgradeItemResultWithSim, printer *util.PrintRecorder, positiveResultsOnly bool) {
	if positiveResultsOnly {
		resultList = filterPositiveSim(resultList)
	}

	reportBasicByBossSim(resultList, printer)
	printer.Println0()
	reportBasicBySlotSim(resultList, printer)
	printer.Println0()
	reportBasicOverallRankSim(resultList, printer)
	printer.Println0()
	printer.Println0()
}

func reportBasicByBossSim(resultList []upgradeItemResultWithSim, printer *util.PrintRecorder) {
	rankedByBoss := groupByBoss(resultList)

	printer.Println("RANKING UPGRADE BY BOSS")
	for _, bossName := range db.BossItemData_NamesInOrder {
		rank := rankedByBoss[bossName]
		if rank != nil {
			printer.Println("BOSS " + bossName)

			var tab util.TabulateOutput
			tab.SetColumnSpacing(2)
			tab.AddColumnHeader("slot", true)
			tab.AddColumnHeader("ilvl", false)
			tab.AddColumnHeader("name", true)
			tab.AddColumnHeader("setCount", false)
			tab.AddColumnHeader("increase", false)
			tab.AddColumnHeader("simIncrease", false)
			tab.AddColumnHeader("simDetail", false)
			tab.AddColumnHeader("note", false)

			for result := range rank.OrderedResult() {
				if result.sim.IsEmpty() {
					tab.AddRow([]string{
						result.slot.Name(),
						strconv.FormatUint(uint64(result.ItemLevel()), 10),
						result.ItemName(),
						strconv.FormatUint(uint64(result.setBonus), 10),
						result.increaseWeightsStr(false),
						"",
						"",
						result.makeNoteFull(),
					})
				} else {
					tab.AddRow([]string{
						result.slot.Name(),
						strconv.FormatUint(uint64(result.ItemLevel()), 10),
						result.ItemName(),
						strconv.FormatUint(uint64(result.setBonus), 10),
						result.increaseWeightsStr(false),
						result.increaseSimStr(false),
						result.increaseSimBreakdown().CompactStringSignedPercent(),
						result.makeNoteFull(),
					})
				}
			}

			tab.Write(printer)
			printer.Println0()
		}
	}
}

func reportBasicBySlotSim(resultList []upgradeItemResultWithSim, printer *util.PrintRecorder) {
	rankedBySlot := groupBySlot(resultList)

	printer.Println("RANKING UPGRADE BY SLOT")
	for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
		rank := rankedBySlot[slot]
		if rank != nil {
			printer.Println("RANKING " + slot.Name())

			var tab util.TabulateOutput
			tab.SetColumnSpacing(2)
			tab.AddColumnHeader("ilvl", false)
			tab.AddColumnHeader("name", true)
			tab.AddColumnHeader("boss", false)
			tab.AddColumnHeader("setCount", false)
			tab.AddColumnHeader("increase", false)
			tab.AddColumnHeader("simIncrease", false)
			tab.AddColumnHeader("simDetail", false)
			tab.AddColumnHeader("note", false)

			for result := range rank.OrderedResult() {
				if result.sim.IsEmpty() {
					tab.AddRow([]string{
						strconv.FormatUint(uint64(result.ItemLevel()), 10),
						result.ItemName(),
						result.boss,
						strconv.FormatUint(uint64(result.setBonus), 10),
						result.increaseWeightsStr(false),
						"",
						"",
						result.makeNoteFull(),
					})
				} else {
					tab.AddRow([]string{
						strconv.FormatUint(uint64(result.ItemLevel()), 10),
						result.ItemName(),
						result.boss,
						strconv.FormatUint(uint64(result.setBonus), 10),
						result.increaseWeightsStr(false),
						result.increaseSimStr(false),
						result.increaseSimBreakdown().CompactStringSignedPercent(),
						result.makeNoteFull(),
					})
				}
			}

			tab.Write(printer)
			printer.Println0()
		}
	}
}

func reportBasicOverallRankSim(resultList []upgradeItemResultWithSim, printer *util.PrintRecorder) {
	ranked := util_rank.RankedCollection[upgradeItemResultWithSim]{}
	for _, result := range resultList {
		ranked.Add(result, result.increaseSim())
	}

	printer.Println("RANKING OVERALL PERCENT UPGRADE")

	var tab util.TabulateOutput
	tab.SetColumnSpacing(2)
	tab.AddColumnHeader("slot", true)
	tab.AddColumnHeader("ilvl", false)
	tab.AddColumnHeader("name", true)
	tab.AddColumnHeader("boss", false)
	tab.AddColumnHeader("setCount", false)
	tab.AddColumnHeader("increase", false)
	tab.AddColumnHeader("simIncrease", false)
	tab.AddColumnHeader("simDetail", false)
	tab.AddColumnHeader("note", false)

	for result := range ranked.OrderedResult() {
		if result.sim.IsEmpty() {
			tab.AddRow([]string{
				result.slot.Name(),
				strconv.FormatUint(uint64(result.ItemLevel()), 10),
				result.ItemName(),
				result.boss,
				strconv.FormatUint(uint64(result.setBonus), 10),
				result.increaseWeightsStr(false),
				"",
				"",
				result.makeNoteFull(),
			})
		} else {
			tab.AddRow([]string{
				result.slot.Name(),
				strconv.FormatUint(uint64(result.ItemLevel()), 10),
				result.ItemName(),
				result.boss,
				strconv.FormatUint(uint64(result.setBonus), 10),
				result.increaseWeightsStr(false),
				result.increaseSimStr(false),
				result.increaseSimBreakdown().CompactStringSignedPercent(),
				result.makeNoteFull(),
			})
		}
	}

	tab.Write(printer)
}

func groupByBoss(resultList []upgradeItemResultWithSim) map[string]*util_rank.RankedCollection[upgradeItemResultWithSim] {
	rankedByBoss := make(map[string]*util_rank.RankedCollection[upgradeItemResultWithSim])
	for _, result := range resultList {
		rank := rankedByBoss[result.boss]
		if rank == nil {
			rank = new(util_rank.RankedCollection[upgradeItemResultWithSim])
			rankedByBoss[result.boss] = rank
		}
		rank.Add(result, result.increaseSim())
	}
	return rankedByBoss
}

func groupBySlot(resultList []upgradeItemResultWithSim) map[items.SlotEquip]*util_rank.RankedCollection[upgradeItemResultWithSim] {
	rankedBySlot := make(map[items.SlotEquip]*util_rank.RankedCollection[upgradeItemResultWithSim])
	for _, result := range resultList {
		rank := rankedBySlot[result.slot]
		if rank == nil {
			rank = new(util_rank.RankedCollection[upgradeItemResultWithSim])
			rankedBySlot[result.slot] = rank
		}
		rank.Add(result, result.increaseSim())
	}
	return rankedBySlot
}

func filterPositiveSim(input []upgradeItemResultWithSim) []upgradeItemResultWithSim {
	output := make([]upgradeItemResultWithSim, 0, len(input))
	for _, item := range input {
		if item.increaseWeightsRaw() > 0.0 || item.increaseSim() > 0.0 {
			output = append(output, item)
		}
	}
	return output
}
