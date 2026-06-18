package multi

import (
	"cmp"
	"paladin_gearing_go/db"
	"paladin_gearing_go/items"
	"slices"
)

func (job *MultiSetJob) CullingReport() {
	for paramIndex := range job.params {
		job.params[paramIndex].cullingReport()
		job.params[paramIndex].cullingReportBags()
		job.params[paramIndex].cullingReportOrphan()
	}
}

func (param *multiSetParamInternal) cullingReport() {
	type extraInfoStruct struct {
		itemId items.ItemId
		count  uint32
	}

	added := make(map[items.ItemId]bool)

	extraInfo := make([]extraInfoStruct, 0, len(param.ExtraItems))
	for _, itemId := range param.ExtraItems {
		seenCount := param.seenInSolutions.content[itemId]
		info := extraInfoStruct{itemId: itemId, count: seenCount}
		extraInfo = append(extraInfo, info)
		added[itemId] = true
	}

	for item := range param.exactEquippedGear.AllItemSeq() {
		itemId := item.ItemId()
		if !added[itemId] {
			seenCount := param.seenInSolutions.content[itemId]
			info := extraInfoStruct{itemId: itemId, count: seenCount}
			extraInfo = append(extraInfo, info)
			added[itemId] = true
		}
	}

	slices.SortFunc(extraInfo, func(a, b extraInfoStruct) int {
		return cmp.Or(cmp.Compare(a.count, b.count), cmp.Compare(a.itemId, b.itemId))
	})

	param.job.printer.Printf("EXTRAS USED %s\n", param.Label)
	for _, info := range extraInfo {
		if slices.Contains(param.BlockedItems, info.itemId) {
			continue
		}
		item, itemFound := param.itemOptions.FindItemIdFirstOptional(info.itemId)
		if itemFound {
			if info.count == 0 {
				param.job.printer.Printf("%5d 0 NONE // %s; %s\n", info.itemId, item.SlotItem().Name(), item.BaseName())
			} else {
				param.job.printer.Printf("%5d %6d // %s; %s\n", info.itemId, info.count, item.SlotItem().Name(), item.BaseName())
			}
		} else {
			basicVersion := db.WowSimDB_ByIdAndUpgrade(info.itemId, 0)
			param.job.printer.Printf("%5d %d MISSING IN OPTIONS // %s\n", info.itemId, info.count, basicVersion.BaseName())
		}
	}
	param.job.printer.Println0()
}

func (param *multiSetParamInternal) cullingReportBags() {
	for _, itemId := range param.addedFromBags {
		seenCount := param.seenInSolutions.content[itemId]
		if seenCount > 0 {
			item, itemFound := param.itemOptions.FindItemIdFirstOptional(itemId)
			if itemFound {
				param.job.printer.Printf("BAGS SUGGESTION %d %d %s; %s !!\n", itemId, seenCount, item.SlotItem().Name(), item.BaseName())
			} else {
				param.job.printer.Printf("BAGS SUGGESTION %d %d BUT missing options?!?!?!?!\n", itemId, seenCount)
			}
		}
	}
}

func (param *multiSetParamInternal) cullingReportOrphan() {
	for itemId, seenCount := range param.seenInSolutions.content {
		if seenCount > 0 {
			if !slices.Contains(param.ExtraItems, itemId) && !slices.Contains(param.addedFromBags, itemId) {
				param.job.printer.Printf("ORPHAN ITEM USED %d %d\n", itemId, seenCount)
			}
		}
	}
}
