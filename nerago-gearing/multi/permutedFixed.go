package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/solver/withhighs"
	"paladin_gearing_go/util"
)

type fixedPermuteEntry struct {
	paramIndex int
	slot       items.SlotEquip
	itemId     items.ItemId
}

type fixedPermuteOptions struct {
	options []fixedPermuteEntry
}

type fixedPermuteSet struct {
	choices []fixedPermuteEntry
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
	return count
}

func (job *MultiSetJob) prepareFixedPermutations() <-chan fixedPermuteSet {
	optionEntriesList := make([]fixedPermuteOptions, 0)
	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		semiFixed := param.SemiFixedSlots
		for slot, itemIdList := range semiFixed {
			entriesList := util.CastSliceAsNew(itemIdList, func(itemId *items.ItemId) fixedPermuteEntry {
				return fixedPermuteEntry{paramIndex, slot, *itemId}
			})
			optionEntriesList = append(optionEntriesList, fixedPermuteOptions{options: entriesList})
		}
	}

	// return channel_op.PermuteAsChannel(optionEntriesList)
	return permuteAsChannel(optionEntriesList)
}

func permuteAsChannel(listsOfOptions []fixedPermuteOptions) <-chan fixedPermuteSet {
	stepChannel := permuteInit(listsOfOptions[0])

	for i := 1; i < len(listsOfOptions); i++ {
		stepChannel = permuteStep(stepChannel, listsOfOptions[i])
	}

	return stepChannel
}

func permuteInit(options fixedPermuteOptions) <-chan fixedPermuteSet {
	stepChannel := make(chan fixedPermuteSet, 8)
	go func() {
		for _, value := range options.options {
			stepChannel <- fixedPermuteSet{choices: []fixedPermuteEntry{value}}
		}
		close(stepChannel)
	}()
	return stepChannel
}

func permuteStep(inChannel <-chan fixedPermuteSet, options fixedPermuteOptions) <-chan fixedPermuteSet {
	outputChannel := make(chan fixedPermuteSet, 8)
	go func() {
		for currSet := range inChannel {
			for _, value := range options.options {
				outputChannel <- fixedPermuteSet{choices: util.CopyAndAppend(currSet.choices, value)}
			}
		}
		close(outputChannel)
	}()
	return outputChannel
}

func (job *MultiSetJob) highProcessSetupForPermute(permuteSet fixedPermuteSet, commonOptions multi_types.CommonOptions, printer *util.PrintRecorder) withhighs.SolverHighsMultiProcess {
	highProcess := withhighs.SolverHighsMultiProcess{}
	highProcess.SetCommon(commonOptions)

	itemOptionsEach := make([]items.FullOptionsMap, len(job.params))
	for paramIndex := range job.params {
		itemOptionsEach[paramIndex] = job.params[paramIndex].itemOptions.Clone()
	}

	printer.Println("PERMUTE SET:")
	for _, entry := range permuteSet.choices {
		printer.Printf(" > %d %s %d\n", entry.paramIndex, entry.slot.Name(), entry.itemId)
		itemOptionsEach[entry.paramIndex].ForceSlotOnlySpecifiedItemId(entry.slot, entry.itemId)
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
