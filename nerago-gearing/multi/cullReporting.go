package multi

import (
	"cmp"
	"paladin_gearing_go/db"
	"paladin_gearing_go/items"
	"paladin_gearing_go/util"
	"slices"
)

func (job *MultiSetJob) CullingReport() {
	for _, prep := range job.itemPrep {
		prep.cullingReportSeen(job.printer)
		prep.cullingReportBags(job.printer)
		prep.cullingReportOrphan(job.printer)
		job.printer.Println0()
		job.printer.Println0()
	}

	job.cullReportAll()
}

func (prep *specItemPrep) cullingReportSeen(printer *util.PrintRecorder) {
	type extraInfoStruct struct {
		itemId items.ItemId
		count  uint32
	}

	added := make(map[items.ItemId]bool)

	extraInfo := make([]extraInfoStruct, 0, len(prep.inputs.ExtraItems))
	for _, itemId := range prep.inputs.ExtraItems {
		seenCount := prep.seenInSolutions.content[itemId]
		info := extraInfoStruct{itemId: itemId, count: seenCount}
		extraInfo = append(extraInfo, info)
		added[itemId] = true
	}

	for item := range prep.exactEquippedGear.AllItemSeq() {
		itemId := item.ItemId()
		if !added[itemId] {
			seenCount := prep.seenInSolutions.content[itemId]
			info := extraInfoStruct{itemId: itemId, count: seenCount}
			extraInfo = append(extraInfo, info)
			added[itemId] = true
		}
	}

	slices.SortFunc(extraInfo, func(a, b extraInfoStruct) int {
		return cmp.Or(cmp.Compare(a.count, b.count), cmp.Compare(a.itemId, b.itemId))
	})

	printer.Printf("EXTRAS USED %s\n", prep.label)
	for _, info := range extraInfo {
		if slices.Contains(prep.inputs.BlockedItems, info.itemId) {
			continue
		}
		item, itemFound := prep.itemOptions.FindItemIdFirstOptional(info.itemId)
		if itemFound {
			if info.count == 0 {
				printer.Printf("%5d 0 NONE // %s; %s\n", info.itemId, item.SlotItem().Name(), item.BaseName())
			} else {
				printer.Printf("%5d %6d // %s; %s\n", info.itemId, info.count, item.SlotItem().Name(), item.BaseName())
			}
		} else {
			itemName := db.LookupItemNameByItemId(info.itemId)
			printer.Printf("%5d %d MISSING IN OPTIONS // %s\n", info.itemId, info.count, itemName)
		}
	}
}

func (prep *specItemPrep) cullingReportBags(printer *util.PrintRecorder) {
	for _, itemId := range prep.addedFromBags {
		seenCount := prep.seenInSolutions.content[itemId]
		if seenCount > 0 {
			item, itemFound := prep.itemOptions.FindItemIdFirstOptional(itemId)
			if itemFound {
				printer.Printf("BAGS SUGGESTION %d %d %s; %s\n", itemId, seenCount, item.SlotItem().Name(), item.BaseName())
			} else {
				printer.Printf("BAGS SUGGESTION %d %d BUT missing options?!?!?!?!\n", itemId, seenCount)
			}
		}
	}
}

func (prep *specItemPrep) cullingReportOrphan(printer *util.PrintRecorder) {
	for itemId, seenCount := range prep.seenInSolutions.content {
		if seenCount > 0 {
			if !slices.Contains(prep.inputs.ExtraItems, itemId) && !slices.Contains(prep.addedFromBags, itemId) {
				name := db.LookupItemNameByItemId(itemId)
				printer.Printf("ORPHAN ITEM USED %d %d // %s\n", itemId, seenCount, name)
			}
		}
	}
}

func (job *MultiSetJob) cullReportAll() {
	// highest count per label
	maxCountForParam := make(map[string]uint32)
	for _, prep := range job.itemPrep {
		for _, count := range prep.seenInSolutions.content {
			maxCountForParam[prep.label] = max(maxCountForParam[prep.label], count)
		}
	}

	// highest ratio per itemId
	combinedCount := make(map[items.ItemId]uint32)
	bestRatio := make(map[items.ItemId]float64)
	seen := make(map[items.ItemId]bool)
	for _, prep := range job.itemPrep {
		maxCount := float64(maxCountForParam[prep.label])
		for _, itemId := range prep.inputs.ExtraItems {
			seenCount := prep.seenInSolutions.content[itemId]
			ratio := float64(seenCount) / maxCount
			bestRatio[itemId] = max(bestRatio[itemId], ratio)
			combinedCount[itemId] += seenCount
			seen[itemId] = true
		}
	}

	// equipped gear is forced to top rating
	for _, prep := range job.itemPrep {
		for item := range prep.exactEquippedGear.AllItemSeq() {
			itemId := item.ItemId()
			bestRatio[itemId] = 1.0
			combinedCount[itemId] += 1000
			seen[itemId] = true
		}
	}

	// copy ratios into nice sortable structure
	type lowEntry struct {
		itemId  items.ItemId
		percent float64
	}
	lowEntries := make([]lowEntry, 0)
	for itemId := range seen {
		entry := lowEntry{itemId, bestRatio[itemId]}
		lowEntries = append(lowEntries, entry)
	}

	// sort by increasing percent/ratio used
	slices.SortFunc(lowEntries, func(a, b lowEntry) int {
		return cmp.Or(cmp.Compare(a.percent, b.percent), cmp.Compare(a.itemId, b.itemId))
	})

	// list anything under 20%
	job.printer.Printf("EXTRAS LOW RATE ACROSS SETS:\n")
	for _, entry := range lowEntries {
		if entry.percent < 0.2 {
			itemId := entry.itemId
			itemName := db.LookupItemNameByItemId(itemId)
			job.printer.Printf("%4.1f%% %6d %s\n", entry.percent*100, itemId, itemName)
		}
	}
	job.printer.Println0()
}
