package multi

import (
	"iter"
	"log"
	"paladin_gearing_go/items"
	"paladin_gearing_go/stats"
	"paladin_gearing_go/util"
	"strconv"
	"strings"
)

func (job *MultiSetJob) determineCommon() commonComboOptions {
	commonOptions, seenIn := searchParamOptions(job.params)

	applyFixedForges(job.fixedForge, &commonOptions, &job.printer)

	removeSingleSetItems(seenIn, &commonOptions, job.fixedForge)

	printCommons(seenIn, commonOptions, &job.printer)

	return commonOptions
}

func searchParamOptions(params []MultiSetParam) (commonComboOptions, map[uint32][]string) {
	commonOptions := make(commonComboOptions)
	seenIn := make(map[uint32][]string)

	for _, param := range params {
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

func groupById(itemSeq iter.Seq[*items.FullItem]) map[uint32][]items.FullItem {
	grouped := make(map[uint32][]items.FullItem)
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
			if one.Equals(&two) {
				result = append(result, one)
			}
		}
	}
	return result
}

func applyFixedForges(fixedForge map[uint32]stats.ReforgeRecipe, commonOptions *commonComboOptions, printer *util.PrintRecorder) {
	for itemId, reforge := range fixedForge {
		options, ok := (*commonOptions)[itemId]
		if ok {
			choice := onlyMatchingForge(options, reforge, itemId)
			(*commonOptions)[itemId] = []items.FullItem{choice}
			printer.Printf("FIXED %s\n", choice.String())
		} else {
			log.Panicf("fixed forge not seen in set options for item %d", itemId)
		}
	}
}

func onlyMatchingForge(options []items.FullItem, reforge stats.ReforgeRecipe, itemId uint32) items.FullItem {
	for _, item := range options {
		if item.Reforge == reforge {
			return item
		}
	}
	panic("fixed forge selection not available for item " + strconv.Itoa(int(itemId)))
}

func removeSingleSetItems(seenIn map[uint32][]string, commonOptions *commonComboOptions, fixedForge map[uint32]stats.ReforgeRecipe) {
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

func printCommons(seenIn map[uint32][]string, commonOptions commonComboOptions, printer *util.PrintRecorder) {
	for itemId, options := range commonOptions {
		if len(options) == 0 {
			log.Panicf("no common forge for %d", itemId)
		}

		item := options[0]
		whereSeen := seenIn[itemId]
		seenText := strings.Join(whereSeen, " ")

		printer.Printf("COMMON %d %s %d => %s\n", itemId, item.FullName(), item.Ref.ItemLevel, seenText)
	}
}
