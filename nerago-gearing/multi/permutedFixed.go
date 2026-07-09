package multi

import (
	"paladin_gearing_go/db"
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver/withhighs"
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

type permuteEntry struct {
	fixed *permuteEntryFixedForce
	group *permuteEntryAllowGroup
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
		for _, forceIdx := range group.groupAIndexes {
			entriesList = append(entriesList, permuteEntry{group: &permuteEntryAllowGroup{group.groupAIndexes, forceIdx, itemId}})
		}
		entriesList = append(entriesList, permuteEntry{group: &permuteEntryAllowGroup{group.groupAIndexes, -1, itemId}})
		for _, forceIdx := range group.groupBIndexes {
			entriesList = append(entriesList, permuteEntry{group: &permuteEntryAllowGroup{group.groupBIndexes, forceIdx, itemId}})
		}
		entriesList = append(entriesList, permuteEntry{group: &permuteEntryAllowGroup{group.groupBIndexes, -1, itemId}})
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

func (job *MultiSetJob) highProcessSetupForPermute(permuteSet permuteSet, printer *util.PrintRecorder) withhighs.SolverHighsMultiProcess {
	highProcess := withhighs.SolverHighsMultiProcess{}

	itemOptionsEach := make([]items.FullOptionsMap, len(job.params))
	for paramIndex := range job.params {
		itemOptionsEach[paramIndex] = job.params[paramIndex].itemOptions.Clone()
	}

	build := util.StringBuild2{}
	for _, entry := range permuteSet.choices {
		if entry.fixed != nil {
			fixed := entry.fixed
			itemOptionsEach[fixed.paramIndex].ForceSlotOnlySpecifiedItemId(fixed.slot, fixed.itemId)
			if !fixed.isSingle {
				paramLabel := job.params[fixed.paramIndex].Label
				build.WriteString(paramLabel)
				build.WriteString("(Forced) ")

				itemName := db.WowSimDB_LookupNameByItemId(fixed.itemId)
				build.WriteString(itemName)
			}
		} else if entry.group != nil {
			group := entry.group

			for paramIndex := range job.params {
				paramLabel := job.params[paramIndex].Label
				if group.forceIndex == paramIndex {
					slot := itemOptionsEach[paramIndex].FindItemIdSlotUnique(group.itemId)
					itemOptionsEach[paramIndex].ForceSlotOnlySpecifiedItemId(slot, group.itemId)
					build.WriteString(paramLabel)
					build.WriteString("(Forced) ")
				} else if slices.Contains(group.allowIndexList, paramIndex) {
					build.WriteString(paramLabel)
					build.WriteString("(Allowed) ")
				} else {
					itemOptionsEach[paramIndex].RemoveItemIdFromAll(group.itemId)
				}
			}

			itemName := db.WowSimDB_LookupNameByItemId(group.itemId)
			build.WriteString(": ")
			build.WriteString(itemName)
		} else {
			panic("empty entry")
		}
	}

	printer.Println("PERMUTE SET:")
	printer.PrintlnFromBuild(build)
	highProcess.SetPermuteLabel(build.String())

	optionsInputList := make([]commonOptionsInput, len(job.params))
	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		optionsInputList[paramIndex] = commonOptionsInput{param.Label, &itemOptionsEach[paramIndex]}
	}
	commonOptions := job.determineCommon(optionsInputList)
	highProcess.SetCommon(commonOptions)

	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		highProcess.AddSetParam(withhighs.SolverHighsMultiParam{
			Label:          param.Label,
			ItemOptions:    itemOptionsEach[paramIndex],
			Gear_model:     &param.Model,
			RatingMultiply: param.ratingMultiply,
		})
	}

	return highProcess
}
