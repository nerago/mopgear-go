package multi

import (
	"slices"

	"github.com/nerago/mopgear-go/db"
	"github.com/nerago/mopgear-go/gear_model/bonus_set"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/setup"
	"github.com/nerago/mopgear-go/solver/solve_highs"
	"github.com/nerago/mopgear-go/solver/solve_highs_types"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_collection"
)

type permuteEntryFixedForce struct {
	specLabel string
	slot      items.SlotEquip
	itemId    items.ItemId
	isSingle  bool
}

type permuteEntryAllowGroup struct {
	allowSpecList []string
	forceSpec     string
	itemId        items.ItemId
}

type permuteEntryUpgrade struct {
	itemId items.ItemId
}

type permuteEntryGems struct {
	allowAlternates bool
}

type permuteEntryBonusItems struct {
	specLabel string
	bonus     *solve_highs_types.OverrideBonusCounts
}

type permuteEntry struct {
	fixed   *permuteEntryFixedForce
	group   *permuteEntryAllowGroup
	upgrade *permuteEntryUpgrade
	gems    *permuteEntryGems
	bonus   *permuteEntryBonusItems
}

type permuteOptions struct {
	options []permuteEntry
}

type permuteSet struct {
	choices []permuteEntry
}

func (job *MultiSetJob) estimateFixedPermutations() int {
	var count = 1
	for _, prep := range job.itemPrep {
		semiFixed := prep.inputs.SemiFixedSlots
		for _, itemIdList := range semiFixed {
			count *= len(itemIdList)
		}
	}
	for _, group := range job.input.ItemInput.DistinctUsageGroups {
		if group.ForceTryInEachParam {
			count *= len(group.GroupALabels) + len(group.GroupBLabels) + 2
		} else {
			count *= 2
		}
	}
	for _, group := range job.input.ItemInput.AlternateUpgradeChoices {
		count *= len(group)
	}
	if job.input.ItemInput.AlternateGemsEnableAsPermute {
		count *= 2
	}
	if job.input.ItemInput.PermuteOnItemCountOptions {
		for _, prep := range job.itemPrep {
			count *= len(prep.model.BonusRequiredSolve.Options)
		}
	}
	return count
}

func (job *MultiSetJob) buildPermutations() <-chan permuteSet {
	optionEntriesList := make([]permuteOptions, 0)

	for _, prep := range job.itemPrep {
		semiFixed := prep.inputs.SemiFixedSlots
		for slot, itemIdList := range semiFixed {
			isSingle := len(itemIdList) == 1
			entriesList := util_collection.MapSliceAsNew(itemIdList, func(itemId *items.ItemId) permuteEntry {
				return permuteEntry{fixed: &permuteEntryFixedForce{prep.label, slot, *itemId, isSingle}}
			})
			optionEntriesList = append(optionEntriesList, permuteOptions{options: entriesList})
		}
	}

	for itemId, group := range job.input.ItemInput.DistinctUsageGroups {
		entriesList := make([]permuteEntry, 0)

		entriesList = append(entriesList, permuteEntry{group: &permuteEntryAllowGroup{group.GroupALabels, "", itemId}})
		entriesList = append(entriesList, permuteEntry{group: &permuteEntryAllowGroup{group.GroupBLabels, "", itemId}})

		if group.ForceTryInEachParam {
			for _, force := range group.GroupALabels {
				entriesList = append(entriesList, permuteEntry{group: &permuteEntryAllowGroup{group.GroupALabels, force, itemId}})
			}
			for _, force := range group.GroupBLabels {
				entriesList = append(entriesList, permuteEntry{group: &permuteEntryAllowGroup{group.GroupBLabels, force, itemId}})
			}
		}

		optionEntriesList = append(optionEntriesList, permuteOptions{options: entriesList})
	}

	for _, group := range job.input.ItemInput.AlternateUpgradeChoices {
		entriesList := util_collection.MapSliceAsNew(group, func(itemId *items.ItemId) permuteEntry {
			return permuteEntry{upgrade: &permuteEntryUpgrade{*itemId}}
		})
		optionEntriesList = append(optionEntriesList, permuteOptions{options: entriesList})
	}

	if job.input.ItemInput.AlternateGemsEnableAsPermute {
		entriesList := []permuteEntry{
			{gems: &permuteEntryGems{true}},
			{gems: &permuteEntryGems{false}},
		}
		optionEntriesList = append(optionEntriesList, permuteOptions{options: entriesList})
	}

	if job.input.ItemInput.PermuteOnItemCountOptions {
		for _, prep := range job.itemPrep {
			if len(prep.model.BonusRequiredSolve.Options) > 1 {
				entriesList := util_collection.MapSliceAsNew(prep.model.BonusRequiredSolve.Options, func(cr *bonus_set.ItemCountsRequired) permuteEntry {
					return permuteEntry{bonus: &permuteEntryBonusItems{
						specLabel: prep.label,
						bonus:     &solve_highs_types.OverrideBonusCounts{Specific: *cr},
					}}
				})
				optionEntriesList = append(optionEntriesList, permuteOptions{options: entriesList})
			}
		}
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
				outputChannel <- permuteSet{choices: util_collection.CopyAndAppend(currSet.choices, value)}
			}
		}
		close(outputChannel)
	}()
	return outputChannel
}

func (group *workingGroup) highProcessSetupForPermute(permuteSet permuteSet, printer *util.PrintRecorder) *solve_highs.SolverHighsMultiProcess {
	itemOptionsEach, overrideBonuses, permuteLabel := group.processSetupItemOptionsForPermute(permuteSet, printer)
	printer.Printf("PERMUTE SET:\n%s\n", permuteLabel)

	commonOptions := group.determineCommon(itemOptionsEach)

	highProcess := new(solve_highs.SolverHighsMultiProcess)
	highProcess.SetPermuteLabel(permuteLabel)
	highProcess.SetCommon(commonOptions)

	for label, work := range group.workers {
		highProcess.AddSetParam(solve_highs.SolverHighsMultiParam{
			Label:          label,
			ItemOptions:    *itemOptionsEach[label],
			SolverModel:    *solve_highs_types.SolverModelBuild(work.Model(), group.weightType, overrideBonuses[label]),
			RatingMultiply: work.ratingMultiply,
		})
	}

	return highProcess
}

func (group *workingGroup) processSetupItemOptionsForPermute(permuteSet permuteSet, printer *util.PrintRecorder) (map[string]*items.FullOptionsMap, map[string]*solve_highs_types.OverrideBonusCounts, string) {
	overrideBonusMap := make(map[string]*solve_highs_types.OverrideBonusCounts)

	itemOptionsEach := make(map[string]*items.FullOptionsMap)
	for _, prep := range group.job.itemPrep {
		itemOptionsEach[prep.label] = new(prep.itemOptions.Clone())
	}

	strBuild := util.StringBuild2{}
	for _, entry := range permuteSet.choices {
		if entry.fixed != nil {
			fixed := entry.fixed
			group.applyPermuteFixed(fixed, itemOptionsEach, &strBuild)
		} else if entry.group != nil {
			permuteGroup := entry.group
			group.applyPermuteGroup(permuteGroup, itemOptionsEach, &strBuild)
		} else if entry.upgrade != nil {
			itemId := entry.upgrade.itemId
			group.applyPermuteItemUpgrade(itemId, itemOptionsEach, printer, &strBuild)
		} else if entry.gems != nil {
			group.applyPermuteGems(entry.gems, itemOptionsEach, &strBuild)
		} else if entry.bonus != nil {
			overrideBonusMap[entry.bonus.specLabel] = entry.bonus.bonus
			group.logPermuteBonus(entry.bonus, &strBuild)
		} else {
			panic("empty entry")
		}
	}

	if strBuild.Len() > 0 {
		strBuild.Rewind(3)
	}

	return itemOptionsEach, overrideBonusMap, strBuild.String()
}

func (group *workingGroup) applyPermuteFixed(fixed *permuteEntryFixedForce, itemOptionsEach map[string]*items.FullOptionsMap, strBuild *util.StringBuild2) {
	itemOptionsEach[fixed.specLabel].ForceSlotOnlySpecifiedItemId(fixed.slot, fixed.itemId)
	if !fixed.isSingle {
		strBuild.WriteString(fixed.specLabel)
		strBuild.WriteString("(Forced) : ")

		itemName := db.LookupItemNameByItemId(fixed.itemId)
		strBuild.WriteString(itemName)
		strBuild.WriteString(" | ")
	}
}

func (group *workingGroup) applyPermuteGroup(permuteGroup *permuteEntryAllowGroup, itemOptionsEach map[string]*items.FullOptionsMap, strBuild *util.StringBuild2) {
	for label := range group.job.itemPrep {
		if permuteGroup.forceSpec == label {
			slot := itemOptionsEach[label].FindItemIdSlotUnique(permuteGroup.itemId)
			itemOptionsEach[label].ForceSlotOnlySpecifiedItemId(slot, permuteGroup.itemId)
			strBuild.WriteString(label)
			strBuild.WriteString("(Forced) ")
		} else if slices.Contains(permuteGroup.allowSpecList, label) {
			strBuild.WriteString(label)
			strBuild.WriteString("(Allowed) ")
		} else {
			itemOptionsEach[label].RemoveItemIdFromAll(permuteGroup.itemId)
		}
	}

	itemName := db.LookupItemNameByItemId(permuteGroup.itemId)
	strBuild.WriteString(": ")
	strBuild.WriteString(itemName)
	strBuild.WriteString(" | ")
}

func (group *workingGroup) applyPermuteItemUpgrade(itemId items.ItemId, itemOptionsEach map[string]*items.FullOptionsMap, printer *util.PrintRecorder, strBuild *util.StringBuild2) {
	foundAny := false
	for label, itemOpts := range itemOptionsEach {
		profession := group.job.itemPrep[label].model.Professions
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
		//panic("requested upgrade of item that isn't an option " + itemId.String())
		group.job.printer.Println("requested upgrade of item that isn't an option " + itemId.String())
	}

	itemName := db.LookupItemNameByItemId(itemId)
	strBuild.WriteString("UPGRADE: ")
	strBuild.WriteString(itemName)
	strBuild.WriteString(" | ")
}

func (group *workingGroup) applyPermuteGems(gems *permuteEntryGems, itemOptionsEach map[string]*items.FullOptionsMap, strBuild *util.StringBuild2) {
	if !gems.allowAlternates {
		for _, itemOpts := range itemOptionsEach {
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

func (group *workingGroup) logPermuteBonus(bonus *permuteEntryBonusItems, strBuild *util.StringBuild2) {
	strBuild.WriteString(bonus.specLabel)
	strBuild.WriteString(": ")
	bonus.bonus.Specific.AppendString(strBuild)
	strBuild.WriteString(" | ")
}
