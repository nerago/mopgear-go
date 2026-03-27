package upgrades

import (
	"paladin_gearing_go/db"
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_rank"
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
			for result := range rank.OrderedResult() {
				printer.Printf("%10s \t%d \t%45s \t%6s\n", result.slot.Name(), result.item.Ref.ItemLevel, result.item.BaseName, result.increaseStr())
			}
			printer.Println0()
		}
	}
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
			for result := range rank.OrderedResult() {
				printer.Printf("  %d \t%45s \t%6s\t %s\n", result.item.Ref.ItemLevel, result.item.BaseName, result.increaseStr(), result.boss)
			}
		}
	}
}

func reportBasicOverallRank(resultList []upgradeItemResult, printer *util.PrintRecorder) {
	ranked := util_rank.RankedCollection[upgradeItemResult]{}
	for _, result := range resultList {
		ranked.Add(result, result.factor)
	}

	printer.Println("RANKING OVERALL PERCENT UPGRADE")
	for result := range ranked.OrderedResult() {
		printer.Printf("%10s \t%d \t%45s \t%6s\t %s\n", result.slot.Name(), result.item.Ref.ItemLevel, result.item.BaseName, result.increaseStr(), result.boss)
	}
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
			for result := range rank.OrderedResult() {
				if result.sim.IsEmpty() {
					printer.Printf("%10s%4d%45s%4d%8s\n", result.slot.Name(), result.item.Ref.ItemLevel, result.item.BaseName, result.setBonus, result.increaseStr())
				} else {
					printer.Printf("%10s%4d%45s%4d%8s%8s\t\t%s\n", result.slot.Name(), result.item.Ref.ItemLevel, result.item.BaseName, result.setBonus, result.increaseStr(),
						result.percentStrSim(), result.increaseSimBreakdown().CompactStringSignedPercent())
				}
			}
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
			for result := range rank.OrderedResult() {
				if result.sim.IsEmpty() {
					printer.Printf("%10s%4d%45s%25s%4d%8s\n", result.slot.Name(), result.item.Ref.ItemLevel, result.item.BaseName, result.boss, result.setBonus, result.increaseStr())
				} else {
					printer.Printf("%10s%4d%45s%25s%4d%8s%8s\t\t%s\n", result.slot.Name(), result.item.Ref.ItemLevel, result.item.BaseName, result.boss, result.setBonus, result.increaseStr(),
						result.percentStrSim(), result.increaseSimBreakdown().CompactStringSignedPercent())
				}
			}
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
	for result := range ranked.OrderedResult() {
		if result.sim.IsEmpty() {
			printer.Printf("%10s%4d%45s%25s%4d%8s\n", result.slot.Name(), result.item.Ref.ItemLevel, result.item.BaseName, result.boss, result.setBonus, result.increaseStr())
		} else {
			printer.Printf("%10s%4d%45s%25s%4d%8s%8s\t\t%s\n", result.slot.Name(), result.item.Ref.ItemLevel, result.item.BaseName, result.boss, result.setBonus, result.increaseStr(),
				result.percentStrSim(), result.increaseSimBreakdown().CompactStringSignedPercent())
		}
	}
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
