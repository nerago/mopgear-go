package multi

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"sync/atomic"

	"github.com/nerago/mopgear-go/gear_model/bonus_set"
	"github.com/nerago/mopgear-go/items"
	"github.com/nerago/mopgear-go/solver"
	"github.com/nerago/mopgear-go/solver/solve_highs_types"
	"github.com/nerago/mopgear-go/util"
	"github.com/nerago/mopgear-go/util/util_async"
	"github.com/nerago/mopgear-go/util/util_collection"
)

const c_decimateTargetItemsPerSlot = 3
const c_decimateTargetItemsPairedSlot = 4

// normally decimate done as part of regular jobs
//func (job *MainJob) TestDecimate() {
//	err := job.prepareItems()
//	if err != nil {
//		panic(err)
//	}
//
//	cancel := util_async.CancelSignal_Make()
//
//	groupChannel := job.prepareWorkingGroups(cancel)
//
//	util_async.ForEach_Channel(1, groupChannel, func(group *workingGroup) {
//		group.runDecimate(cancel)
//	})
//}

func (group *workingGroup) runDecimate(cancel util_async.CancelSignal) error {
	workChannel := util_async.SeqToChannel_Cancellable(maps.Values(group.workers), cancel)
	return util_async.ForEach_Channel_PassError(c_decimateThreadCount, workChannel, func(work *specWorker) error {
		tracker := util.TrackProgress_Nop()
		return work.runDecimateWork(tracker, group.job.printer, group.job.input.TimeLimitEachSolve, cancel)
	})
}

func (work *specWorker) runDecimateWork(tracker *util.TrackProgress, printer *util.PrintRecorder, timeout int, cancel util_async.CancelSignal) error {
	solveOptions := items.SolvableOptionsMap_of(work.ItemOptions())
	solverModel, err := solve_highs_types.SolverModelBuild(work.Model(), work.weightType, nil)
	if err != nil {
		return err
	}

	return work.decimateForBaseSet(tracker, printer, timeout, work.baselineResult.SolvedSet, solveOptions, solverModel, cancel)
}

func (work *specWorker) decimateForBaseSet(tracker *util.TrackProgress, printer *util.PrintRecorder, timeout int, baseItemSet items.SolvableItemSet, solveOptions items.SolvableOptionsMap, solverModel *solve_highs_types.SolverModel, cancel util_async.CancelSignal) error {
	estimateSteps := uint64(baseItemSet.Items().CountNonEmptySlots())
	currentStep := new(atomic.Uint64)
	tracker.RunFromAtomicInt(currentStep, estimateSteps)
	defer tracker.SetDone()

	bestBySlot := items.SlotEquipMap[*util_collection.SetComparable[items.ItemId]]{}

	for i, item := range baseItemSet.Items() {
		if item != nil {
			slotEquip := items.SlotEquip(i)
			set := new(util_collection.SetComparable[items.ItemId])
			set.AddIfMissing(item.ItemId())
			bestBySlot.Put(slotEquip, set)
		}
	}

	util_async.ForEach_Slice_Cancellable(c_decimateThreadCount, items.SlotEquip_List, cancel, func(slotEquipPtr *items.SlotEquip) {
		slotEquip := *slotEquipPtr
		if baseItemSet.Items().Has(slotEquip) {
			if solveOptions.CountSlotUniqueItemIds(slotEquip) > c_decimateTargetItemsPerSlot {
				bestSlotSet := bestBySlot.GetOrPanic(slotEquip)
				work.decimateFindSlotBestN(bestSlotSet, slotEquip, &solveOptions, solverModel, printer, timeout, cancel)
			} else {
				for itemId := range solveOptions.SeqSlotUniqueItemIds(slotEquip) {
					bestBySlot.GetOrPanic(slotEquip).AddIfMissing(itemId)
				}
			}
			currentStep.Add(1)
		}
	})

	work.decimateRestoreSetBonusItems(&bestBySlot, work.ItemOptions(), work.expectAllBonusItemsAvailable)

	return work.decimateApply(&bestBySlot, printer)
}

func (work *specWorker) decimateFindSlotBestN(bestForSlot *util_collection.SetComparable[items.ItemId], slotEquip items.SlotEquip, solveOptionsBase *items.SolvableOptionsMap, solverModel *solve_highs_types.SolverModel, printer *util.PrintRecorder, timeout int, cancel util_async.CancelSignal) {
	targetItemCount := c_decimateTargetItemsPerSlot
	if slotEquip == items.Equip_Ring1 || slotEquip == items.Equip_Ring2 || slotEquip == items.Equip_Trinket1 || slotEquip == items.Equip_Trinket2 {
		targetItemCount = c_decimateTargetItemsPairedSlot
	}

	// errors at some point of this process are expected, although ideally just Infeasible status.
	// currently just return/break once we don't get a clean run, could be a bit smarter
	// some errors could indicate other bugs in the process
	for bestForSlot.Size() < targetItemCount && cancel.ShouldContinue() {
		restrictedOptions, isValid := work.decimatePrepareItemOptions(bestForSlot, solveOptionsBase)
		if !isValid {
			return
		}

		futureCheckSet, err := solver.LaunchSolve(&restrictedOptions, solverModel, printer, work.weightType, timeout)
		if err != nil {
			work.handleDecimateError(printer, err)
			return
		}
		if err := util_async.ChainCancel(cancel, futureCheckSet); err != nil {
			work.handleDecimateError(printer, err)
			return
		}

		if checkSet, gotCheck := futureCheckSet.WaitForResult(); gotCheck {
			if checkSet.Error != nil || checkSet.Value == nil {
				work.handleDecimateError(printer, checkSet.Error)
				return
			}
			nextBestItem := checkSet.Value.Items().Get(slotEquip)
			if nextBestItem == nil {
				return
			}
			bestForSlot.AddIfMissing(nextBestItem.ItemId())
		} else {
			return // set was probably infeasible without removed items, bail out
		}
	}
}

func (work *specWorker) handleDecimateError(printer *util.PrintRecorder, err error) {
	printer.Printf("DECIMATE ERROR IGNORED: %v", err)
}

func (work *specWorker) decimatePrepareItemOptions(bestForSlot *util_collection.SetComparable[items.ItemId], solveOptionsBase *items.SolvableOptionsMap) (items.SolvableOptionsMap, bool) {
	restrictedOptions := solveOptionsBase.Clone()
	for removeId := range bestForSlot.SeqValues() {
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

func (work *specWorker) decimateApply(bestBySlot *items.SlotEquipMap[*util_collection.SetComparable[items.ItemId]], printer *util.PrintRecorder) error {
	for slot, idSet := range bestBySlot.SeqKeyValue() {
		oldTotalCount := len(work.ItemOptions().Get(slot))
		oldItems := itemsInSlot(work.ItemOptions(), slot)

		err := work.ItemOptions().FilterSlot(slot, func(item *items.FullItem) bool {
			return idSet.HasValue(item.ItemId())
		})
		if err != nil {
			return err
		}

		newTotalCount := len(work.ItemOptions().Get(slot))
		newItems := itemsInSlot(work.ItemOptions(), slot)
		printer.Printf("DECIMATE %s %d %s: Options (%d -> %d) Items (%d -> %d)\n", work.label, work.weightType, slot.Name(), oldTotalCount, newTotalCount, len(oldItems), len(newItems))
		for itemId := range oldItems {
			if !newItems[itemId] {
				printer.Printf(" > removed %d\n", itemId)
			}
		}
	}
	return nil
}

func itemsInSlot(options *items.FullOptionsMap, slot items.SlotEquip) map[items.ItemId]bool {
	itemSeen := make(map[items.ItemId]bool)
	for item := range options.SlotItemSeq(slot) {
		itemSeen[item.ItemId()] = true
	}
	return itemSeen
}
