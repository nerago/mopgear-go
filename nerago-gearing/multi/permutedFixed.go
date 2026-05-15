package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/solver/withhighs"
	"paladin_gearing_go/util"
	"slices"
)

type permuteEntryType int8

type permuteEntryFixedForce struct {
	paramIndex int
	slot       items.SlotEquip
	itemId     items.ItemId
}

type permuteEntryAllowGroup struct {
	paramIndexList []int
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

func (job *MultiSetJob) estimateFixedPermutations() uint64 {
	var count uint64 = 1
	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		semiFixed := param.SemiFixedSlots
		for _, itemIdList := range semiFixed {
			count *= uint64(len(itemIdList))
		}
	}
	for range job.distinctUsageGroups {
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
			entriesList := util.CastSliceAsNew(itemIdList, func(itemId *items.ItemId) permuteEntry {
				return permuteEntry{fixed: &permuteEntryFixedForce{paramIndex, slot, *itemId}}
			})
			optionEntriesList = append(optionEntriesList, permuteOptions{options: entriesList})
		}
	}

	for itemId, group := range job.distinctUsageGroups {
		entriesList := []permuteEntry{
			{group: &permuteEntryAllowGroup{group.groupAIndexes, itemId}},
			{group: &permuteEntryAllowGroup{group.groupBIndexes, itemId}},
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

func (job *MultiSetJob) highProcessSetupForPermute(permuteSet permuteSet, commonOptions multi_types.CommonOptions, printer *util.PrintRecorder) withhighs.SolverHighsMultiProcess {
	highProcess := withhighs.SolverHighsMultiProcess{}
	highProcess.SetCommon(commonOptions)

	itemOptionsEach := make([]items.FullOptionsMap, len(job.params))
	for paramIndex := range job.params {
		itemOptionsEach[paramIndex] = job.params[paramIndex].itemOptions.Clone()
	}

	printer.Println("PERMUTE SET:")
	for _, entry := range permuteSet.choices {
		if entry.fixed != nil {
			fixed := entry.fixed
			printer.Printf(" > %d %s %d\n", fixed.paramIndex, fixed.slot.Name(), fixed.itemId)
			itemOptionsEach[fixed.paramIndex].ForceSlotOnlySpecifiedItemId(fixed.slot, fixed.itemId)
		} else if entry.group != nil {
			group := entry.group
			build := util.StringBuild2{}
			build.WriteString(" > ")
			for paramIndex := range job.params {
				if slices.Contains(group.paramIndexList, paramIndex) {
					build.WriteUint32(uint32(paramIndex))
					build.WriteRune(' ')
				} else {
					itemOptionsEach[paramIndex].RemoveItemIdFromAll(group.itemId)
				}
			}
			build.WriteUint32(uint32(group.itemId))
			printer.PrintlnFromBuild(build)
		} else {
			panic("empty entry")
		}
	}

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
