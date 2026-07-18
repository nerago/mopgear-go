package multi

import (
	"paladin_gearing_go/db"
	"paladin_gearing_go/items"
	"paladin_gearing_go/setup"
	"paladin_gearing_go/solver/solve_highs"
	"paladin_gearing_go/util"
	"slices"
)

type permuteEntryFixedForce struct {
	paramIndex int
	slot       items.SlotEquip
	itemId     items.ItemId
	isSingle   bool
}

type permuteEntryAllowGroup struct {
	allowIndexList []int
	forceIndex     int
	itemId         items.ItemId
}

type permuteEntryUpgrade struct {
	itemId items.ItemId
}

type permuteEntryGems struct {
	allowAlternates bool
}

type permuteEntry struct {
	fixed   *permuteEntryFixedForce
	group   *permuteEntryAllowGroup
	upgrade *permuteEntryUpgrade
	gems    *permuteEntryGems
}

type permuteOptions struct {
	options []permuteEntry
}

type permuteSet struct {
	choices []permuteEntry
}

func (job *MultiSetJob) estimateFixedPermutations() int {
	var count = 1
	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		semiFixed := param.SemiFixedSlots
		for _, itemIdList := range semiFixed {
			count *= len(itemIdList)
		}
	}
	for _, group := range job.distinctUsageGroups {
		count *= len(group.groupAIndexes) + len(group.groupBIndexes) + 2
	}
	for _, group := range job.alternateUpgradeChoices {
		count *= len(group)
	}
	if job.alternateGemmingAsPermute {
		count *= 2
	}
	return count
}

func (job *MultiSetJob) preparePermutations() <-chan permuteSet {
	optionEntriesList := make([]permuteOptions, 0)

	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		semiFixed := param.SemiFixedSlots
		for slot, itemIdList := range semiFixed {
			isSingle := len(itemIdList) == 1
			entriesList := util.MapSliceAsNew(itemIdList, func(itemId *items.ItemId) permuteEntry {
				return permuteEntry{fixed: &permuteEntryFixedForce{paramIndex, slot, *itemId, isSingle}}
			})
			optionEntriesList = append(optionEntriesList, permuteOptions{options: entriesList})
		}
	}

	for itemId, group := range job.distinctUsageGroups {
		entriesList := make([]permuteEntry, 0)

		entriesList = append(entriesList, permuteEntry{group: &permuteEntryAllowGroup{group.groupAIndexes, -1, itemId}})
		entriesList = append(entriesList, permuteEntry{group: &permuteEntryAllowGroup{group.groupBIndexes, -1, itemId}})

		if group.forceTryInEachParam {
			for _, forceIdx := range group.groupAIndexes {
				entriesList = append(entriesList, permuteEntry{group: &permuteEntryAllowGroup{group.groupAIndexes, forceIdx, itemId}})
			}
			for _, forceIdx := range group.groupBIndexes {
				entriesList = append(entriesList, permuteEntry{group: &permuteEntryAllowGroup{group.groupBIndexes, forceIdx, itemId}})
			}
		}

		optionEntriesList = append(optionEntriesList, permuteOptions{options: entriesList})
	}

	for _, group := range job.alternateUpgradeChoices {
		entriesList := util.MapSliceAsNew(group, func(itemId *items.ItemId) permuteEntry {
			return permuteEntry{upgrade: &permuteEntryUpgrade{*itemId}}
		})
		optionEntriesList = append(optionEntriesList, permuteOptions{options: entriesList})
	}

	if job.alternateGemmingAsPermute {
		entriesList := []permuteEntry{
			{gems: &permuteEntryGems{true}},
			{gems: &permuteEntryGems{false}},
		}
		optionEntriesList = append(optionEntriesList, permuteOptions{options: entriesList})
	}

	return permuteAsChannel(optionEntriesList)
}

func permuteAsChannel(listsOfOptions []permuteOptions) <-chan permuteSet {
	stepChannel := permuteInit(listsOfOptions[0])

	for i := 1; i < len(listsOfOptions); i++ {
		stepChannel = permuteStep(stepChannel, listsOfOptions[i])
	}

	return stepChannel
}

func permuteInit(options permuteOptions) <-chan permuteSet {
	stepChannel := make(chan permuteSet, 8)
	go func() {
		for _, value := range options.options {
			stepChannel <- permuteSet{choices: []permuteEntry{value}}
		}
		close(stepChannel)
	}()
	return stepChannel
}

func permuteStep(inChannel <-chan permuteSet, options permuteOptions) <-chan permuteSet {
	outputChannel := make(chan permuteSet, 8)
	go func() {
		for currSet := range inChannel {
			for _, value := range options.options {
				outputChannel <- permuteSet{choices: util.CopyAndAppend(currSet.choices, value)}
			}
		}
		close(outputChannel)
	}()
	return outputChannel
}

func (job *MultiSetJob) highProcessSetupForPermute(permuteSet permuteSet, printer *util.PrintRecorder) *solve_highs.SolverHighsMultiProcess {
	itemOptionsEach, strBuild := job.processSetupItemOptionsForPermute(permuteSet, printer)

	highProcess := new(solve_highs.SolverHighsMultiProcess)

	if strBuild.Len() > 0 {
		strBuild.Rewind(3)
		printer.Println("PERMUTE SET:")
		printer.PrintlnFromBuild(strBuild)
	}
	highProcess.SetPermuteLabel(strBuild.String())

	job.highProcessSetup_addOptions(highProcess, itemOptionsEach)
	return highProcess
}

func (job *MultiSetJob) processSetupItemOptionsForPermute(permuteSet permuteSet, printer *util.PrintRecorder) ([]items.FullOptionsMap, util.StringBuild2) {
	itemOptionsEach := util.MapSliceAsNew(job.params, func(param *multiSetParamInternal) items.FullOptionsMap {
		return param.itemOptions.Clone()
	})

	strBuild := util.StringBuild2{}
	for _, entry := range permuteSet.choices {
		if entry.fixed != nil {
			fixed := entry.fixed
			job.applyPermuteFixed(fixed, &itemOptionsEach, &strBuild)
		} else if entry.group != nil {
			group := entry.group
			job.applyPermuteGroup(group, &itemOptionsEach, &strBuild)
		} else if entry.upgrade != nil {
			itemId := entry.upgrade.itemId
			job.applyPermuteItemUpgrade(itemId, &itemOptionsEach, printer, &strBuild)
		} else if entry.gems != nil {
			job.applyPermuteGems(entry.gems, &itemOptionsEach, &strBuild)
		} else {
			panic("empty entry")
		}
	}

	return itemOptionsEach, strBuild
}

func (job *MultiSetJob) applyPermuteFixed(fixed *permuteEntryFixedForce, itemOptionsEach *[]items.FullOptionsMap, strBuild *util.StringBuild2) {
	(*itemOptionsEach)[fixed.paramIndex].ForceSlotOnlySpecifiedItemId(fixed.slot, fixed.itemId)
	if !fixed.isSingle {
		paramLabel := job.params[fixed.paramIndex].Label
		strBuild.WriteString(paramLabel)
		strBuild.WriteString("(Forced) : ")

		itemName := db.LookupItemNameByItemId(fixed.itemId)
		strBuild.WriteString(itemName)
		strBuild.WriteString(" | ")
	}
}

func (job *MultiSetJob) applyPermuteGroup(group *permuteEntryAllowGroup, itemOptionsEach *[]items.FullOptionsMap, strBuild *util.StringBuild2) {
	for paramIndex := range job.params {
		paramLabel := job.params[paramIndex].Label
		if group.forceIndex == paramIndex {
			slot := (*itemOptionsEach)[paramIndex].FindItemIdSlotUnique(group.itemId)
			(*itemOptionsEach)[paramIndex].ForceSlotOnlySpecifiedItemId(slot, group.itemId)
			strBuild.WriteString(paramLabel)
			strBuild.WriteString("(Forced) ")
		} else if slices.Contains(group.allowIndexList, paramIndex) {
			strBuild.WriteString(paramLabel)
			strBuild.WriteString("(Allowed) ")
		} else {
			(*itemOptionsEach)[paramIndex].RemoveItemIdFromAll(group.itemId)
		}
	}

	itemName := db.LookupItemNameByItemId(group.itemId)
	strBuild.WriteString(": ")
	strBuild.WriteString(itemName)
	strBuild.WriteString(" | ")
}

func (job *MultiSetJob) applyPermuteItemUpgrade(itemId items.ItemId, itemOptionsEach *[]items.FullOptionsMap, printer *util.PrintRecorder, strBuild *util.StringBuild2) {
	profession := job.params[0].Model.Professions
	foundAny := false
	for itemOpts := range util.ForPointer(*itemOptionsEach) {
		if itemOpts.IncludesItemId(itemId) {
			itemOpts.MapEachItem(func(item *items.FullItem) items.FullItem {
				if item.ItemId() == itemId {
					return *setup.UpgradeExistingItemToTargetLevel(item, items.MAX_UPGRADE_LEVEL, profession, printer)
				} else {
					return *item
				}
			})
			foundAny = true
		}
	}

	if !foundAny {
		panic("requested upgrade of item that isn't an option " + itemId.String())
	}

	itemName := db.LookupItemNameByItemId(itemId)
	strBuild.WriteString("UPGRADE: ")
	strBuild.WriteString(itemName)
	strBuild.WriteString(" | ")
}

func (job *MultiSetJob) applyPermuteGems(gems *permuteEntryGems, itemOptionsEach *[]items.FullOptionsMap, strBuild *util.StringBuild2) {
	if !gems.allowAlternates {
		for itemOpts := range util.ForPointer(*itemOptionsEach) {
			itemOpts.FilterAllItems(func(item *items.FullItem) bool {
				return !item.HasBeenRegemmed()
			})
		}
	}

	if gems.allowAlternates {
		strBuild.WriteString("All Gems")
	} else {
		strBuild.WriteString("Original Gems")
	}
	strBuild.WriteString(" | ")
}
