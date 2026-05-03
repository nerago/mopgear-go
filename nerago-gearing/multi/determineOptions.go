package multi

import (
	"iter"
	"log"
	"math/big"
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"strconv"
	"strings"
)

type CommonComboOptions map[items.ItemId][]items.FullItem

func (optionsMap *CommonComboOptions) TotalCombinationCount() *big.Int {
	valueCount := 0
	total := big.NewInt(1)
	for _, slotArray := range *optionsMap {
		slotSize := int64(len(slotArray))
		if slotSize > 0 {
			total.Mul(total, big.NewInt(slotSize))
			valueCount++
		}
	}
	if valueCount == 0 {
		panic("empty options")
	}
	return total
}

func (job *MultiSetJob) determineCommon() CommonComboOptions {
	commonOptions, seenIn := searchParamOptions(&job.params)

	applyFixedForges(job.fixedForge, &commonOptions, job.printer)
	checkItemRates(job.specificAllowRates, &commonOptions)

	removeSingleSetItems(seenIn, &commonOptions, job.fixedForge, job.specificAllowRates)

	printCommons(seenIn, commonOptions, job.printer)

	return commonOptions
}

func searchParamOptions(params *[]multiSetParamInternal) (CommonComboOptions, map[items.ItemId][]string) {
	commonOptions := make(CommonComboOptions)
	seenIn := make(map[items.ItemId][]string)

	for paramIndex := range *params {
		param := &(*params)[paramIndex]
		if param.IncludeInFirstPass {
			grouped := groupById(param.itemOptions.AllItems())
			for itemId, options := range grouped {
				seenIn[itemId] = append(seenIn[itemId], param.Label)
				commonOptions[itemId] = filterCommonForges(commonOptions[itemId], options)
			}
		}
	}

	return commonOptions, seenIn
}

func groupById(itemSeq iter.Seq[*items.FullItem]) map[items.ItemId][]items.FullItem {
	grouped := make(map[items.ItemId][]items.FullItem)
	for item := range itemSeq {
		grouped[item.ItemId()] = append(grouped[item.ItemId()], *item)
	}
	return grouped
}

func groupByIdMapped[T any](inputList []T, mapper func(T) iter.Seq[*items.FullItem]) map[items.ItemId][]items.FullItem {
	grouped := make(map[items.ItemId][]items.FullItem)
	for i := range inputList {
		for item := range mapper(inputList[i]) {
			grouped[item.ItemId()] = append(grouped[item.ItemId()], *item)
		}
	}
	return grouped
}

func filterCommonForges(prior []items.FullItem, newOptions []items.FullItem) []items.FullItem {
	if prior == nil {
		return newOptions
	}

	result := make([]items.FullItem, 0, len(prior))
	for _, one := range prior {
		for _, two := range newOptions {
			if one.EqualsExceptEnchant(&two) { // essentially just choose first one
				result = append(result, one)
			} else if one.Ref.ItemId == two.Ref.ItemId && one.Ref.ItemLevel != two.Ref.ItemLevel {
				panic("inconsistent item levels " + one.CreateString() + " and " + two.CreateString())
			}
		}
	}
	return result
}

func applyFixedForges(fixedForge map[items.ItemId]stats.ReforgeRecipe, commonOptions *CommonComboOptions, printer *util.PrintRecorder) {
	for itemId, reforge := range fixedForge {
		options, ok := (*commonOptions)[itemId]
		if ok {
			choice := onlyMatchingForge(options, reforge, itemId)
			(*commonOptions)[itemId] = []items.FullItem{choice}
			printer.Printf("FIXED %s\n", choice.CreateString())
		} else {
			log.Panicf("fixed forge not seen in set options for item %d", itemId)
		}
	}
}

func checkItemRates(allowRates map[items.ItemId]specificAllowEntry, commonOptions *CommonComboOptions) {
	for itemId := range allowRates {
		_, ok := (*commonOptions)[itemId]
		if !ok {
			log.Panicf("item rate not seen in set options %d", itemId)
		}
	}
}

func onlyMatchingForge(options []items.FullItem, reforge stats.ReforgeRecipe, itemId items.ItemId) items.FullItem {
	for _, item := range options {
		if item.Reforge == reforge {
			return item
		}
	}
	panic("fixed forge selection not available for item " + strconv.Itoa(int(itemId)))
}

func removeSingleSetItems(seenIn map[items.ItemId][]string, commonOptions *CommonComboOptions, fixedForge map[items.ItemId]stats.ReforgeRecipe, specificRates map[items.ItemId]specificAllowEntry) {
	for itemId, whereSeen := range seenIn {
		_, isFixed := fixedForge[itemId]
		if isFixed {
			continue
		}

		_, isSpecific := specificRates[itemId]
		if isSpecific {
			continue
		}

		if len(whereSeen) <= 1 {
			delete(*commonOptions, itemId)
		}
	}
}

func printCommons(seenIn map[items.ItemId][]string, commonOptions CommonComboOptions, printer *util.PrintRecorder) {
	for itemId, options := range commonOptions {
		if len(options) == 0 {
			log.Panicf("no common forge for %d", itemId)
		}

		item := options[0]
		whereSeen := seenIn[itemId]
		seenText := strings.Join(whereSeen, " ")

		printer.Printf("COMMON %d %s %d => %s\n", itemId, item.CreateFullName(), item.Ref.ItemLevel, seenText)
	}
}

func printChosenCombo(combo *commonCombo, printer *util.PrintRecorder) {
	printer.Println("COMMON_COMBO " + combo.logString())
	for itemId, entry := range combo.entryMap {
		if entry.forceMode == Force_Forbidden {
			printer.Printf("COMMON %d forbidden\n", itemId)
		} else {
			printer.Printf("COMMON %s\n", entry.Item.CreateString())
		}
	}
	for _, entry := range combo.entryMap {
		if entry.forceMode != Force_Forbidden {
			item := entry.Item
			if item.Reforge.IsEmpty() {
				printer.Printf("common[%d] = stats.ReforgeRecipe_empty\n", item.ItemId())
			} else {
				printer.Printf("common[%d] = stats.ReforgeRecipe_of(stats.%s, stats.%s)\n", item.ItemId(), item.Reforge.From.EnumName(), item.Reforge.To.EnumName())
			}
		}
	}
}

func (job *MultiSetJob) revisedComboActuallyUsed(outputs []singleProposed, initialCombo *commonCombo, printer *util.PrintRecorder) commonCombo {
	printer.Println("MAKING REVISED commonCombo")

	grouped := make(map[items.ItemId][]*items.FullItem)
	for index := range outputs {
		for item := range outputs[index].fullSet.Items().AllItemSeq() {
			grouped[item.ItemId()] = append(grouped[item.ItemId()], item)
		}
	}

	revisedCombo := initialCombo.clone()
	for itemId := range revisedCombo.entryMap {
		groupArray, hasGroup := grouped[itemId]
		_, hasFixedForge := job.fixedForge[itemId]
		_, hasSpecificRate := job.specificAllowRates[itemId]

		shouldRemove := !hasGroup || len(groupArray) < 2
		if shouldRemove && hasFixedForge {
			printer.Printf("WOULD REMOVE COMMON BUT HAS fixedForge %d\n", itemId)
		} else if shouldRemove && hasSpecificRate {
			printer.Printf("WOULD REMOVE COMMON BUT HAS specificRate %d\n", itemId)
		} else if shouldRemove {
			delete(revisedCombo.entryMap, itemId)
		}
	}

	// TODO consider prining all current items, not just those in common
	printChosenCombo(&revisedCombo, printer)

	return revisedCombo
}

func (job *MultiSetJob) determineComboFromScratch(outputs []singleProposed, comboType comboType) commonCombo {
	combo := commonCombo_Make(0, comboType)
	itemSeen := make(map[items.ItemId]*items.FullItem)

	for index := range outputs {
		for item := range outputs[index].fullSet.Items().AllItemSeq() {
			previousVersion, hasPrevious := itemSeen[item.ItemId()]
			if hasPrevious && previousVersion.Equals(item) {
				combo.addItem(item.ItemId(), item, Force_Optional)
			} else if hasPrevious {
				panic("inconsisent version of item " + item.CreateString())
			} else {
				itemSeen[item.ItemId()] = item
			}
		}
	}

	return combo
}
