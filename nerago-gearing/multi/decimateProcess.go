package multi

import (
	"cmp"
	"fmt"
	"maps"
	"paladin_gearing_go/gear_model/bonus_set"
	"paladin_gearing_go/items"
	"paladin_gearing_go/solver"
	"paladin_gearing_go/solver/solve_highs_types"
	"paladin_gearing_go/util"
	"paladin_gearing_go/util/util_async"
	"paladin_gearing_go/util/util_collection"
	"slices"
)

const c_decimateTargetItemsPerSlot = 3
const c_decimateTargetItemsPairedSlot = 4

// normally decimate done as part of regular jobs
func (job *MultiSetJob) TestDecimate() {
	job.prepareItems()
	groupChannel := job.prepareWorkingGroups()
	groupChannel = job.runDecimate(groupChannel)

	util_async.ForEach_Channel(1, groupChannel, func(group *workingGroup) {})
}

func (job *MultiSetJob) runDecimate(channel <-chan *workingGroup) <-chan *workingGroup {
	return util_async.Map_ChannelToChannel(c_decimateThreadCount, channel, func(group *workingGroup) *workingGroup {
		workChannel := util_async.SeqToChannel(maps.Values(group.workers))
		util_async.ForEach_Channel(c_decimateThreadCount, workChannel, func(work *specWorker) {
			tracker := util.TrackProgress_Nop()
			work.runDecimateWork(tracker, job.printer, job.input.TimeLimitEachSolve)
		})
		return group
	})
}

func (work *specWorker) runDecimateWork(tracker *util.TrackProgress, printer *util.PrintRecorder, timeout int) {
	solveOptions := items.SolvableOptionsMap_of(work.ItemOptions())
	solverModel := solve_highs_types.SolverModelBuild(work.Model(), work.weightType, nil)
	//futureBaseItemSet := solver.LaunchSolve(new(solveOptions), solverModel, printer, work.weightType, timeout)
	//
	//// isn't this just the same as baseline?
	//
	//if baseItemSet, gotBase := futureBaseItemSet.WaitForResult(); gotBase {
	//	work.decimateForBaseSet(tracker, printer, timeout, baseItemSet, solveOptions, solverModel)
	//} else {
	//	printer.Println("Decimate initial set failed " + work.Label())
	//}

	work.decimateForBaseSet(tracker, printer, timeout, work.baselineResult.SolvedSet, solveOptions, solverModel)
	tracker.SetDone()
}

func (work *specWorker) decimateForBaseSet(tracker *util.TrackProgress, printer *util.PrintRecorder, timeout int, baseItemSet items.SolvableItemSet, solveOptions items.SolvableOptionsMap, solverModel *solve_highs_types.SolverModel) {
	estimateSteps := uint64(baseItemSet.Items().CountNonEmptySlots())
	currentStep := new(uint64)
	tracker.RunFromInt(currentStep, estimateSteps)

	bestBySlot := items.SlotEquipMap[*util_collection.SetComparable[items.ItemId]]{}

	for i, item := range baseItemSet.Items() {
		if item != nil {
			slotEquip := items.SlotEquip(i)
			set := new(util_collection.SetComparable[items.ItemId])
			set.AddIfMissing(item.ItemId())
			bestBySlot.Put(slotEquip, set)
		}
	}

	for i, item := range baseItemSet.Items() {
		if item != nil {
			slotEquip := items.SlotEquip(i)

			if solveOptions.CountSlotUniqueItemIds(slotEquip) > c_decimateTargetItemsPerSlot {
				work.decimateFindSlotBestN(&bestBySlot, slotEquip, &solveOptions, solverModel, printer, timeout)
			} else {
				for itemId := range solveOptions.SeqSlotUniqueItemIds(slotEquip) {
					bestBySlot.GetOrPanic(slotEquip).AddIfMissing(itemId)
				}
			}

			tracker.NewChild().SetDone()
			*currentStep++
		}
	}

	work.decimateRestoreSetBonusItems(&bestBySlot, work.ItemOptions(), work.expectAllBonusItemsAvailable)

	work.decimateApply(&bestBySlot, printer)
}

func (work *specWorker) decimateFindSlotBestN(bestBySlot *items.SlotEquipMap[*util_collection.SetComparable[items.ItemId]], slotEquip items.SlotEquip, solveOptionsBase *items.SolvableOptionsMap, solverModel *solve_highs_types.SolverModel, printer *util.PrintRecorder, timeout int) {
	targetItemCount := c_decimateTargetItemsPerSlot
	if slotEquip == items.Equip_Ring1 || slotEquip == items.Equip_Ring2 || slotEquip == items.Equip_Trinket1 || slotEquip == items.Equip_Trinket2 {
		targetItemCount = c_decimateTargetItemsPairedSlot
	}

	for bestBySlot.GetOrPanic(slotEquip).Size() < targetItemCount {
		restrictedOptions, isValid := work.decimatePrepareItemOptions(bestBySlot, slotEquip, solveOptionsBase)
		if !isValid {
			return
		}

		futureCheckSet := solver.LaunchSolve(&restrictedOptions, solverModel, printer, work.weightType, timeout)

		if checkSet, gotCheck := futureCheckSet.WaitForResult(); gotCheck {
			nextBestItem := checkSet.Items().Get(slotEquip)
			if nextBestItem == nil {
				return
			}
			bestBySlot.GetOrPanic(slotEquip).AddIfMissing(nextBestItem.ItemId())
		} else {
			return // set was probably infeasible without removed items, bail out
		}
	}
}

func (work *specWorker) decimatePrepareItemOptions(bestBySlot *items.SlotEquipMap[*util_collection.SetComparable[items.ItemId]], slotEquip items.SlotEquip, solveOptionsBase *items.SolvableOptionsMap) (items.SolvableOptionsMap, bool) {
	restrictedOptions := solveOptionsBase.Clone()
	for removeId := range bestBySlot.GetOrNilValue(slotEquip).SeqValues() {
		if restrictedOptions.RemoveItemIdFromAll(removeId) {
			//slot got emptied (either this or paired), bail out
			return items.SolvableOptionsMap{}, false
		}
	}
	return restrictedOptions, true
}

func (work *specWorker) decimateRestoreSetBonusItems(bestBySlot *items.SlotEquipMap[*util_collection.SetComparable[items.ItemId]], optionsMap *items.FullOptionsMap, expectAllBonusItems bool) {
	for _, bonusSet := range work.model.BonusEnabled.EnabledSets {
		for _, slot := range items.BonusSetSlotList {
			work.decimateRestoreSpecificSetBonusInSlot(bestBySlot, optionsMap, slot, bonusSet, expectAllBonusItems)
		}
	}
}

func (work *specWorker) decimateRestoreSpecificSetBonusInSlot(bestBySlot *items.SlotEquipMap[*util_collection.SetComparable[items.ItemId]], optionsMap *items.FullOptionsMap, slot items.SlotEquip, bonusSet bonus_set.PreparedBonus, expectAllBonusItems bool) {
	// check if we kept the item already
	for itemId := range bestBySlot.GetOrNilValue(slot).SeqValues() {
		if bonusSet.IncludesItem(itemId) {
			return
		}
	}

	// any other items for slot in options
	found := make([]*items.FullItem, 0)
	for item := range optionsMap.SlotItemSeq(slot) {
		if bonusSet.IncludesItem(item.ItemId()) {
			found = append(found, item)
		}
	}

	if len(found) > 0 {
		// sort item level descending
		slices.SortFunc(found, func(a, b *items.FullItem) int {
			return cmp.Compare(b.ItemLevel(), a.ItemLevel())
		})

		highestItemId := found[0].ItemId()
		bestBySlot.GetOrPanic(slot).AddIfMissing(highestItemId)
	} else if expectAllBonusItems {
		panic(fmt.Sprintf("decimate for %s %d no bonus item found for %s in %s", work.label, work.weightType, slot.Name(), bonusSet.Name()))
	}
}

func (work *specWorker) decimateApply(bestBySlot *items.SlotEquipMap[*util_collection.SetComparable[items.ItemId]], printer *util.PrintRecorder) {
	for slot, idSet := range bestBySlot.SeqKeyValue() {
		oldTotalCount := len(work.ItemOptions().Get(slot))
		oldItems := itemsInSlot(work.ItemOptions(), slot)

		work.ItemOptions().FilterSlot(slot, func(item *items.FullItem) bool {
			return idSet.HasValue(item.ItemId())
		})

		newTotalCount := len(work.ItemOptions().Get(slot))
		newItems := itemsInSlot(work.ItemOptions(), slot)
		printer.Printf("DECIMATE %s %d %s: Options (%d -> %d) Items (%d -> %d)\n", work.label, work.weightType, slot.Name(), oldTotalCount, newTotalCount, len(oldItems), len(newItems))
		for itemId := range oldItems {
			if !newItems[itemId] {
				printer.Printf(" > removed %d\n", itemId)
			}
		}
	}
}

func itemsInSlot(options *items.FullOptionsMap, slot items.SlotEquip) map[items.ItemId]bool {
	itemSeen := make(map[items.ItemId]bool)
	for item := range options.SlotItemSeq(slot) {
		itemSeen[item.ItemId()] = true
	}
	return itemSeen
}
