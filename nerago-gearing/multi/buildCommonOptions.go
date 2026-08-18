package multi

import (
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/multi/multi_types"
	"github.com/nerago/mopgear-go/stats"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
	"iter"
	"slices"
	"strconv"
	"strings"
)

func (group *workingGroup) determineCommon(optionsInputMap map[string]*items.FullOptionsMap) multi_types.CommonOptions {
	reforgingAllowNonCommon := group.job.input.ItemInput.ReforgingAllowNonCommon
	commonOptions, seenIn := searchItemOptions(optionsInputMap, reforgingAllowNonCommon)

	applyFixedForges(group.job.input.ItemInput.FixedForge, &commonOptions, group.job.printer)

	removeSingleSetItems(seenIn, &commonOptions, group.job.input.ItemInput.FixedForge)

	restrictItemOptionsToCommon(optionsInputMap, commonOptions)

	validateCommons(commonOptions)
	// printCommons(seenIn, commonOptions, job.printer)

	return commonOptions
}

func searchItemOptions(optionsInputList map[string]*items.FullOptionsMap, reforgingAllowNonCommon bool) (multi_types.CommonOptions, map[items.ItemRef][]string) {
	commonOptions := multi_types.CommonOptions_Make()
	seenIn := make(map[items.ItemRef][]string)

	for label, itemOptions := range optionsInputList {
		grouped := groupById(itemOptions.AllItems())
		for itemRef, options := range grouped {
			seenIn[itemRef] = append(seenIn[itemRef], label)
			if reforgingAllowNonCommon {
				commonOptions.ApplyToSlicesByItemRef(itemRef, func(prior []items.FullItem) []items.FullItem {
					return unionCommonForges(prior, options)
				})

			} else {
				commonOptions.ApplyToSlicesByItemRef(itemRef, func(prior []items.FullItem) []items.FullItem {
					return intersectCommonForges(prior, options)
				})
			}
		}
	}

	commonOptions.ApplyToAllSlices(func(slice []items.FullItem) []items.FullItem {
		return util_collection.RemoveDuplicatesFunc_NewIfChanged(slice, (*items.FullItem).Equals)
	})
	for itemRef := range seenIn {
		seenIn[itemRef] = util_collection.RemoveDuplicatesComparable_NewIfChanged(seenIn[itemRef])
	}

	return commonOptions, seenIn
}

func groupById(itemSeq iter.Seq[*items.FullItem]) map[items.ItemRef][]items.FullItem {
	grouped := make(map[items.ItemRef][]items.FullItem)
	for item := range itemSeq {
		ref := items.ItemRef_Of(item)
		grouped[ref] = append(grouped[ref], *item)
	}
	return grouped
}

func unionCommonForges(prior []items.FullItem, newOptions []items.FullItem) []items.FullItem {
	if prior == nil {
		return newOptions
	}

	result := slices.Clone(prior)
outer:
	for a := range newOptions {
		one := &newOptions[a]
		for b := range prior {
			two := &prior[b]
			if one.Equals(two) {
				continue outer
			} else if one.ItemId() == two.ItemId() && one.ItemLevel() != two.ItemLevel() {
				panic("inconsistent item levels " + one.CreateString() + " and " + two.CreateString())
			}
		}
		result = append(result, *one)
	}
	return result
}

func intersectCommonForges(prior []items.FullItem, newOptions []items.FullItem) []items.FullItem {
	if prior == nil {
		return newOptions
	}

	result := make([]items.FullItem, 0, len(prior))
	for a := range prior {
		one := &prior[a]
		found := false
		for b := range newOptions {
			two := &newOptions[b]
			if one.Equals(two) {
				found = true
				break
			} else if one.ItemId() == two.ItemId() && one.ItemLevel() != two.ItemLevel() {
				panic("inconsistent item levels " + one.CreateString() + " and " + two.CreateString())
			}
		}
		if found {
			result = append(result, *one)
		}
	}
	return result
}

func applyFixedForges(fixedForge map[items.ItemId]stats.ReforgeRecipe, commonOptions *multi_types.CommonOptions, printer *util.PrintRecorder) {
	// we could we apply this to the input itemOptions earlier on, but here we make sure it makes it out of all sets as a valid common too, rather than potentially disappearing from some specs silently
	for itemId, reforge := range fixedForge {
		if !commonOptions.IncludesItemId(itemId) {
			panic("fixed forge not seen in set options for item " + itemId.String())
		}

		commonOptions.ApplyToSlicesByItemId(itemId, func(options []items.FullItem) []items.FullItem {
			choice := onlyMatchingForge(options, reforge, itemId)
			printer.Printf("FIXED %s\n", choice.CreateString())
			return []items.FullItem{choice}
		})
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

func removeSingleSetItems(seenIn map[items.ItemRef][]string, commonOptions *multi_types.CommonOptions, fixedForge map[items.ItemId]stats.ReforgeRecipe) {
	for itemRef, whereSeen := range seenIn {
		_, isFixed := fixedForge[itemRef.ItemId]
		if isFixed {
			continue
		}

		if len(whereSeen) <= 1 {
			commonOptions.RemoveByItemRef(itemRef)
		}
	}
}

func restrictItemOptionsToCommon(optionsInputMap map[string]*items.FullOptionsMap, commonOptions multi_types.CommonOptions) {
	for _, itemOptions := range optionsInputMap {
		itemOptions.FilterAllItems(func(item *items.FullItem) bool {
			ref := items.ItemRef_Of(item)
			commonVersions, isCommon := commonOptions.Get(ref)
			if isCommon {
				return util_collection.ContainsFunc_Pointer(commonVersions, item.Equals)
			} else {
				return true
			}
		})
	}
}

func validateCommons(commonOptions multi_types.CommonOptions) {
	for itemRef, options := range commonOptions.SeqGroups() {
		if len(options) == 0 {
			panic("no common forge for " + itemRef.ItemId.String())
		}
	}
}

func printCommons(seenIn map[items.ItemRef][]string, commonOptions multi_types.CommonOptions, printer *util.PrintRecorder) {
	for itemRef, options := range commonOptions.SeqGroups() {
		item := options[0]
		whereSeen := seenIn[itemRef]
		seenText := strings.Join(whereSeen, " ")

		printer.Printf("COMMON %d %s %d => %s\n", itemRef.ItemId, item.CreateFullName(), item.ItemLevel(), seenText)
	}
}
