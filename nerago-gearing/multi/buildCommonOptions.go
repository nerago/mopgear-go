package multi

import (
	"iter"
	"log"
	"paladin_gearing_go/items"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"strconv"
	"strings"
)

func (job *MultiSetJob) determineCommon() multi_types.CommonOptions {
	commonOptions, seenIn := searchParamOptions(&job.params)

	applyFixedForges(job.fixedForge, &commonOptions, job.printer)
	checkItemRates(job.specificAllowRates, &commonOptions)

	removeSingleSetItems(seenIn, &commonOptions, job.fixedForge, job.specificAllowRates)

	printCommons(seenIn, commonOptions, job.printer)

	return commonOptions
}

func searchParamOptions(params *[]multiSetParamInternal) (multi_types.CommonOptions, map[items.ItemId][]string) {
	commonOptions := make(multi_types.CommonOptions)
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

func filterCommonForges(prior []items.FullItem, newOptions []items.FullItem) []items.FullItem {
	if prior == nil {
		return newOptions
	}

	result := make([]items.FullItem, 0, len(prior))
	for _, one := range prior {
		for _, two := range newOptions {
			if one.EqualsExceptEnchant(&two) { // essentially just choose first one
				result = append(result, one)
			} else if one.ItemId() == two.ItemId() && one.ItemLevel() != two.ItemLevel() {
				panic("inconsistent item levels " + one.CreateString() + " and " + two.CreateString())
			}
		}
	}
	return result
}

func applyFixedForges(fixedForge map[items.ItemId]stats.ReforgeRecipe, commonOptions *multi_types.CommonOptions, printer *util.PrintRecorder) {
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
		if item.Reforge() == reforge {
			return item
		}
	}
	panic("fixed forge selection not available for item " + strconv.Itoa(int(itemId)))
}

func removeSingleSetItems(seenIn map[items.ItemId][]string, commonOptions *multi_types.CommonOptions, fixedForge map[items.ItemId]stats.ReforgeRecipe, specificRates map[items.ItemId]specificAllowEntry) {
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

func printCommons(seenIn map[items.ItemId][]string, commonOptions multi_types.CommonOptions, printer *util.PrintRecorder) {
	for itemId, options := range commonOptions {
		if len(options) == 0 {
			log.Panicf("no common forge for %d", itemId)
		}

		item := options[0]
		whereSeen := seenIn[itemId]
		seenText := strings.Join(whereSeen, " ")

		printer.Printf("COMMON %d %s %d => %s\n", itemId, item.CreateFullName(), item.ItemLevel(), seenText)
	}
}
