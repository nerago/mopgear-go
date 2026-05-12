package multi

import (
	"paladin_gearing_go/items"
	"paladin_gearing_go/multi/multi_types"
	"paladin_gearing_go/solver/withhighs"
	"paladin_gearing_go/util/channel_op"
)

type fixedPermuteEntry struct {
	paramIndex int
	slot       items.SlotEquip
	itemId     items.ItemId
}

func (job *MultiSetJob) prepareFixedPermutations() <-chan []fixedPermuteEntry {
	optionEntriesList := make([][]fixedPermuteEntry, 0)
	for paramIndex := range job.params {
		param := &job.params[paramIndex]
		semiFixed := param.SemiFixedSlots
		for slot, itemIdList := range semiFixed {
			entriesList := make([]fixedPermuteEntry, 0)
			for _, itemId := range itemIdList {
				entry := fixedPermuteEntry{paramIndex, slot, itemId}
				entriesList = append(entriesList, entry)
			}
			optionEntriesList = append(optionEntriesList, entriesList)
		}
	}

	return channel_op.PermuteAsChannel(optionEntriesList)
}

func (job *MultiSetJob) highProcessSetupForPermute(permuteSet []fixedPermuteEntry, commonOptions multi_types.CommonOptions) withhighs.SolverHighsMultiProcess {
	highProcess := withhighs.SolverHighsMultiProcess{}
	highProcess.SetCommon(commonOptions)

	itemOptionsEach := make([]items.FullOptionsMap, len(job.params))
	for paramIndex := range job.params {
		itemOptionsEach[paramIndex] = job.params[paramIndex].itemOptions.Clone()
	}

	job.printer.Println("PERMUTE SET:")
	for _, entry := range permuteSet {
		job.printer.Printf(" > %d %d\n", entry.paramIndex, entry.itemId)
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
