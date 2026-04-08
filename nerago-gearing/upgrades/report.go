package upgrades

import (
	"paladin_gearing_go/db"
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_rank"
	"strconv"
)

func reportBasicResults(resultList []upgradeItemResult, printer *util.PrintRecorder, positiveResultsOnly bool) {
	if positiveResultsOnly {
		resultList = filterPositive(resultList)
	}

	reportBasicByBoss(resultList, printer)
	printer.Println0()
	reportBasicBySlot(resultList, printer)
	printer.Println0()
	reportBasicOverallRank(resultList, printer)
	printer.Println0()
	printer.Println0()
}

func reportBasicByBoss(resultList []upgradeItemResult, printer *util.PrintRecorder) {
	rankedByBoss := groupByBoss(resultList)

	printer.Println("RANKING UPGRADE BY BOSS")
	for _, bossName := range db.BossItemData_NamesInOrder {
		rank := rankedByBoss[bossName]
		if rank != nil {
			printer.Println(bossName)

			var tab util.TabulateOutput
			tab.SetColumnSpacing(2)
			tab.AddColumnHeader("slot", true)
			tab.AddColumnHeader("ilvl", false)
			tab.AddColumnHeader("name", true)
			tab.AddColumnHeader("increase", false)
			tab.AddColumnHeader("note", false)

			for result := range rank.OrderedResult() {
				tab.AddRow([]string{
					result.slot.Name(),
					strconv.FormatUint(uint64(result.item.Ref.ItemLevel), 10),
					result.item.BaseName,
					result.increaseStr(),
					makeNote(result),
				})
			}

			tab.Write(printer)
			printer.Println0()
		}
	}
}

func makeNote(result upgradeItemResult) string {
	return result.canUpgrade.Text()
}

func makeNoteSim(result upgradeItemResultWithSim) string {
	return result.canUpgrade.Text()
}

type reportableRanked interface {
	comparable
	reportable
	ranking() float64
	Slot() items.SlotEquip
}

func reportBasicBySlot(resultList []upgradeItemResult, printer *util.PrintRecorder) {
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
			tab.AddColumnHeader("increase", false)
			tab.AddColumnHeader("boss", false)
			tab.AddColumnHeader("note", false)

			for result := range rank.OrderedResult() {
				tab.AddRow([]string{
					strconv.FormatUint(uint64(result.item.Ref.ItemLevel), 10),
					result.item.BaseName,
					result.increaseStr(),
					result.boss,
					makeNote(result),
				})
			}

			tab.Write(printer)
			printer.Println0()
		}
	}
}

func reportBasicOverallRank(resultList []upgradeItemResult, printer *util.PrintRecorder) {
	ranked := util_rank.RankedCollection[upgradeItemResult]{}
	for _, result := range resultList {
		ranked.Add(result, result.factor)
	}

	printer.Println("RANKING OVERALL PERCENT UPGRADE")

	var tab util.TabulateOutput
	tab.SetColumnSpacing(2)
	tab.AddColumnHeader("slot", true)
	tab.AddColumnHeader("ilvl", false)
	tab.AddColumnHeader("name", true)
	tab.AddColumnHeader("increase", false)
	tab.AddColumnHeader("boss", false)
	tab.AddColumnHeader("note", false)

	for result := range ranked.OrderedResult() {
		tab.AddRow([]string{
			result.slot.Name(),
			strconv.FormatUint(uint64(result.item.Ref.ItemLevel), 10),
			result.item.BaseName,
			result.increaseStr(),
			result.boss,
			makeNote(result),
		})
	}

	tab.Write(printer)
}

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
						strconv.FormatUint(uint64(result.item.Ref.ItemLevel), 10),
						result.item.BaseName,
						strconv.FormatUint(uint64(result.setBonus), 10),
						result.increaseStr(),
						"",
						"",
						makeNoteSim(result),
					})
				} else {
					tab.AddRow([]string{
						result.slot.Name(),
						strconv.FormatUint(uint64(result.item.Ref.ItemLevel), 10),
						result.item.BaseName,
						strconv.FormatUint(uint64(result.setBonus), 10),
						result.increaseStr(),
						result.percentStrSim(),
						result.increaseSimBreakdown().CompactStringSignedPercent(),
						makeNoteSim(result),
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
						strconv.FormatUint(uint64(result.item.Ref.ItemLevel), 10),
						result.item.BaseName,
						result.boss,
						strconv.FormatUint(uint64(result.setBonus), 10),
						result.increaseStr(),
						"",
						"",
						makeNoteSim(result),
					})
				} else {
					tab.AddRow([]string{
						strconv.FormatUint(uint64(result.item.Ref.ItemLevel), 10),
						result.item.BaseName,
						result.boss,
						strconv.FormatUint(uint64(result.setBonus), 10),
						result.increaseStr(),
						result.percentStrSim(),
						result.increaseSimBreakdown().CompactStringSignedPercent(),
						makeNoteSim(result),
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
		ranked.Add(result, result.ranking())
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
				strconv.FormatUint(uint64(result.item.Ref.ItemLevel), 10),
				result.item.BaseName,
				result.boss,
				strconv.FormatUint(uint64(result.setBonus), 10),
				result.increaseStr(),
				"",
				"",
				makeNoteSim(result),
			})
		} else {
			tab.AddRow([]string{
				result.slot.Name(),
				strconv.FormatUint(uint64(result.item.Ref.ItemLevel), 10),
				result.item.BaseName,
				result.boss,
				strconv.FormatUint(uint64(result.setBonus), 10),
				result.increaseStr(),
				result.percentStrSim(),
				result.increaseSimBreakdown().CompactStringSignedPercent(),
				makeNoteSim(result),
			})
		}
	}

	tab.Write(printer)
}

func groupByBoss[T reportableRanked](resultList []T) map[string]*util_rank.RankedCollection[T] {
	rankedByBoss := make(map[string]*util_rank.RankedCollection[T])
	for _, result := range resultList {
		rank := rankedByBoss[result.Boss()]
		if rank == nil {
			rank = new(util_rank.RankedCollection[T])
			rankedByBoss[result.Boss()] = rank
		}
		rank.Add(result, result.ranking())
	}
	return rankedByBoss
}

func groupBySlot[T reportableRanked](resultList []T) map[items.SlotEquip]*util_rank.RankedCollection[T] {
	rankedBySlot := make(map[items.SlotEquip]*util_rank.RankedCollection[T])
	for _, result := range resultList {
		rank := rankedBySlot[result.Slot()]
		if rank == nil {
			rank = new(util_rank.RankedCollection[T])
			rankedBySlot[result.Slot()] = rank
		}
		rank.Add(result, result.ranking())
	}
	return rankedBySlot
}

func filterPositive(input []upgradeItemResult) []upgradeItemResult {
	output := make([]upgradeItemResult, 0, len(input))
	for _, item := range input {
		if item.increase() > 0.0 {
			output = append(output, item)
		}
	}
	return output
}

func filterPositiveSim(input []upgradeItemResultWithSim) []upgradeItemResultWithSim {
	output := make([]upgradeItemResultWithSim, 0, len(input))
	for _, item := range input {
		if item.increase() > 0.0 || item.percentSim() > 0.0 {
			output = append(output, item)
		}
	}
	return output
}
