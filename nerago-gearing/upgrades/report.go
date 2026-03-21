package upgrades

import (
	"paladin_gearing_go/db"
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
)

func reportBasicResults(resultList []upgradeItemResult, printer *util.PrintRecorder) {
	reportBasicByBoss(resultList, printer)
	printer.Println0()
	reportBasicBySlot(resultList, printer)
	printer.Println0()
	reportBasicOverallRank(resultList, printer)
	printer.Println0()
	printer.Println0()
}

func reportBasicByBoss(resultList []upgradeItemResult, printer *util.PrintRecorder) {
	rankedByBoss := make(map[string]*util.RankedCollection[upgradeItemResult])
	for _, result := range resultList {
		rank := rankedByBoss[result.boss]
		if rank == nil {
			rank = new(util.RankedCollection[upgradeItemResult])
			rankedByBoss[result.boss] = rank
		}
		rank.Add(result, result.factor)
	}

	printer.Println("RANKING UPGRADE BY BOSS")
	for _, bossName := range db.BossItemData_NamesInOrder {
		rank := rankedByBoss[bossName]
		if rank != nil {
			printer.Println(bossName)
			for result := range rank.OrderedResult() {
				printer.Printf("%10s \t%d \t%45s \t%6s\n", result.slot.Name(), result.item.Ref.ItemLevel, result.item.BaseName, result.percentStr())
			}
			printer.Println0()
		}
	}
}

func reportBasicBySlot(resultList []upgradeItemResult, printer *util.PrintRecorder) {
	rankedBySlot := make(map[items.SlotEquip]*util.RankedCollection[upgradeItemResult])
	for _, result := range resultList {
		rank := rankedBySlot[result.slot]
		if rank == nil {
			rank = new(util.RankedCollection[upgradeItemResult])
			rankedBySlot[result.slot] = rank
		}
		rank.Add(result, result.factor)
	}

	printer.Println("RANKING UPGRADE BY SLOT")
	for slot := items.Equip_Iter_First; slot <= items.Equip_Iter_Last; slot++ {
		rank := rankedBySlot[slot]
		if rank != nil {
			printer.Println("RANKING " + slot.Name())
			for result := range rank.OrderedResult() {
				printer.Printf("  %d \t%45s \t%6s\t %s\n", result.item.Ref.ItemLevel, result.item.BaseName, result.percentStr(), result.boss)
			}
		}
	}
}

func reportBasicOverallRank(resultList []upgradeItemResult, printer *util.PrintRecorder) {
	ranked := util.RankedCollection[upgradeItemResult]{}
	for _, result := range resultList {
		ranked.Add(result, result.factor)
	}

	printer.Println("RANKING OVERALL PERCENT UPGRADE")
	for result := range ranked.OrderedResult() {
		printer.Printf("%10s \t%d \t%45s \t%6s\t %s\n", result.slot.Name(), result.item.Ref.ItemLevel, result.item.BaseName, result.percentStr(), result.boss)
	}
}
