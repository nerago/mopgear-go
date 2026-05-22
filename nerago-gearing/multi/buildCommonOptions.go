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

type commonOptionsInput struct {
	label       string
	itemOptions *items.FullOptionsMap
}

func (job *MultiSetJob) determineCommon(optionsInputList []commonOptionsInput) multi_types.CommonOptions {
	commonOptions, seenIn := searchItemOptions(optionsInputList)

	applyFixedForges(job.fixedForge, &commonOptions, job.printer)

	removeSingleSetItems(seenIn, &commonOptions, job.fixedForge)

	restrictItemOptionsToCommon(optionsInputList, commonOptions)

	validateCommons(commonOptions)
	// printCommons(seenIn, commonOptions, job.printer)

	return commonOptions
}

func searchItemOptions(optionsInputList []commonOptionsInput) (multi_types.CommonOptions, map[items.ItemId][]string) {
	commonOptions := make(multi_types.CommonOptions)
	seenIn := make(map[items.ItemId][]string)

	for paramIndex := range optionsInputList {
		input := &optionsInputList[paramIndex]
		grouped := groupById(input.itemOptions.AllItems())
		for itemId, options := range grouped {
			seenIn[itemId] = append(seenIn[itemId], input.label)
			commonOptions[itemId] = filterCommonForges(commonOptions[itemId], options)
		}
	}

	for itemId := range commonOptions {
		commonOptions[itemId] = util.RemoveDuplicatesFunc(commonOptions[itemId], (*items.FullItem).Equals)
	}
	for itemId := range seenIn {
		seenIn[itemId] = util.RemoveDuplicatesComparable(seenIn[itemId])
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
			if one.Equals(&two) { // essentially just choose first one
				result = append(result, one)
			} else if one.ItemId() == two.ItemId() && one.ItemLevel() != two.ItemLevel() {
				panic("inconsistent item levels " + one.CreateString() + " and " + two.CreateString())
			}
		}
	}
	return result
}

func applyFixedForges(fixedForge map[items.ItemId]stats.ReforgeRecipe, commonOptions *multi_types.CommonOptions, printer *util.PrintRecorder) {
	// we could we apply this to the input itemOptions earlier on, but here we make sure it makes it out of all sets as a valid common too, rather than potentially disappearing from some specs silently
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

func removeSingleSetItems(seenIn map[items.ItemId][]string, commonOptions *multi_types.CommonOptions, fixedForge map[items.ItemId]stats.ReforgeRecipe) {
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

func restrictItemOptionsToCommon(optionsInputList []commonOptionsInput, commonOptions multi_types.CommonOptions) {
	for paramIndex := range optionsInputList {
		itemOptions := optionsInputList[paramIndex].itemOptions
		itemOptions.FilterAllItems(func(item *items.FullItem) bool {
			commonVersions, isCommon := commonOptions[item.ItemId()]
			if isCommon {
				return util.ContainsFunc_Pointer(commonVersions, item.Equals)
			} else {
				return true
			}
		})
	}
}

func validateCommons(commonOptions multi_types.CommonOptions) {
	for itemId, options := range commonOptions {
		if len(options) == 0 {
			log.Panicf("no common forge for %d", itemId)
		}
	}
}

func printCommons(seenIn map[items.ItemId][]string, commonOptions multi_types.CommonOptions, printer *util.PrintRecorder) {
	for itemId, options := range commonOptions {
		item := options[0]
		whereSeen := seenIn[itemId]
		seenText := strings.Join(whereSeen, " ")

		printer.Printf("COMMON %d %s %d => %s\n", itemId, item.CreateFullName(), item.ItemLevel(), seenText)
	}
}
