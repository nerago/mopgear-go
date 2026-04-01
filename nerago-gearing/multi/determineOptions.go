package multi

import (
	"iter"
	"log"
	"maps"
	"math/big"
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver"
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

	removeSingleSetItems(seenIn, &commonOptions, job.fixedForge)

	printCommons(seenIn, commonOptions, job.printer)

	return commonOptions
}

func searchParamOptions(params *[]MultiSetParam) (CommonComboOptions, map[items.ItemId][]string) {
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

func onlyMatchingForge(options []items.FullItem, reforge stats.ReforgeRecipe, itemId items.ItemId) items.FullItem {
	for _, item := range options {
		if item.Reforge == reforge {
			return item
		}
	}
	panic("fixed forge selection not available for item " + strconv.Itoa(int(itemId)))
}

func removeSingleSetItems(seenIn map[items.ItemId][]string, commonOptions *CommonComboOptions, fixedForge map[items.ItemId]stats.ReforgeRecipe) {
	for itemId, whereSeen := range seenIn {
		_, isFixed := fixedForge[itemId]
		if isFixed {
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

func printChosenCombo(combo CommonCombo, printer *util.PrintRecorder) {
	for _, item := range combo {
		printer.Printf("COMMON %s\n", item.CreateString())
	}
	for _, item := range combo {
		if item.Reforge.IsEmpty() {
			printer.Printf("common[%d] = stats.ReforgeRecipe_empty\n", item.ItemId())
		} else {
			printer.Printf("common[%d] = stats.ReforgeRecipe{From: stats.%s, To: stats.%s}\n", item.ItemId(), item.Reforge.From.EnumName(), item.Reforge.To.EnumName())
		}
	}
}

func (job *MultiSetJob) revisedComboActuallyUsed(outputs []solver.SolveOutput, initialCombo CommonCombo, printer *util.PrintRecorder) CommonCombo {
	printer.Printf("REVISED COMMON")

	grouped := make(map[items.ItemId][]*items.FullItem)
	for index := range outputs {
		for item := range outputs[index].FullSet.Items().AllItemSeq() {
			grouped[item.ItemId()] = append(grouped[item.ItemId()], item)
		}
	}

	revisedCombo := maps.Clone(initialCombo)
	for itemId := range revisedCombo {
		groupArray, hasGroup := grouped[itemId]
		_, hasFixedForge := job.fixedForge[itemId]

		shouldRemove := !hasGroup || len(groupArray) < 2
		if shouldRemove && hasFixedForge {
			printer.Printf("WOULD REMOVE COMMON BUT HAS fixedForge %d\n", itemId)
		} else if shouldRemove {
			delete(revisedCombo, itemId)
		}
	}

	// TODO consider prining all current items, not just those in common
	printChosenCombo(revisedCombo, printer)

	return revisedCombo
}
