package multi

import (
	"cmp"
	"fmt"
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

// normally decimate done as part of regular jobs
func (job *MultiSetJob) TestDecimate() {
	job.prepareItems()
	job.prepareWorking()
	job.runDecimate()
}

func (job *MultiSetJob) runDecimate() {
	tracker := util.TrackProgress_Start()
	tracker.RunOuterTracking(job.working.Size())
	defer tracker.SetDone()

	workChannel := util_async.SeqToChannel(job.working.SeqValues())
	util_async.ForEach_Channel(c_decimateThreadCount, workChannel, func(work *specWorking) {
		work.runDecimateWork(tracker.NewChild(), job.printer, job.input.TimeLimitEachSolve)
	})
}

func (work *specWorking) runDecimateWork(tracker *util.TrackProgress, printer *util.PrintRecorder, timeout int) {
	solveOptions := items.SolvableOptionsMap_of(work.ItemOptions())
	solverModel := solve_highs_types.SolverModelBuild(work.Model(), work.weightType, nil)
	futureBaseItemSet := solver.LaunchSolve(new(solveOptions), solverModel, printer, work.weightType, timeout)

	if baseItemSet, gotBase := futureBaseItemSet.WaitForResult(); gotBase {
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

				work.decimateFindSlotBestN(&bestBySlot, slotEquip, &solveOptions, solverModel, printer, timeout)

				tracker.NewChild().SetDone()
				*currentStep++
			}
		}

		work.decimateRestoreSetBonusItems(&bestBySlot, work.ItemOptions())

		work.decimateApply(&bestBySlot)
	} else {
		printer.Println("Decimate initial set failed " + work.Label())
	}
	tracker.SetDone()
}

func (work *specWorking) decimateFindSlotBestN(bestBySlot *items.SlotEquipMap[*util_collection.SetComparable[items.ItemId]], slotEquip items.SlotEquip, solveOptionsBase *items.SolvableOptionsMap, solverModel *solve_highs_types.SolverModel, printer *util.PrintRecorder, timeout int) {
	targetItemCount := c_decimateTargetItemsPerSlot
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

func (work *specWorking) decimatePrepareItemOptions(bestBySlot *items.SlotEquipMap[*util_collection.SetComparable[items.ItemId]], slotEquip items.SlotEquip, solveOptionsBase *items.SolvableOptionsMap) (items.SolvableOptionsMap, bool) {
	restrictedOptions := solveOptionsBase.Clone()
	for removeId := range bestBySlot.GetOrNilValue(slotEquip).SeqValues() {
		if restrictedOptions.RemoveItemIdFromAll(removeId) {
			//slot got emptied (either this or paired), bail out
			return items.SolvableOptionsMap{}, false
		}
	}
	return restrictedOptions, true
}

func (work *specWorking) decimateRestoreSetBonusItems(bestBySlot *items.SlotEquipMap[*util_collection.SetComparable[items.ItemId]], optionsMap *items.FullOptionsMap) {
	for _, bonusSet := range work.model.BonusEnabled.EnabledSets {
		for _, slot := range items.BonusSetSlotList {
			work.decimateRestoreSpecificSetBonusInSlot(bestBySlot, optionsMap, slot, bonusSet)
		}
	}
}

func (work *specWorking) decimateRestoreSpecificSetBonusInSlot(bestBySlot *items.SlotEquipMap[*util_collection.SetComparable[items.ItemId]], optionsMap *items.FullOptionsMap, slot items.SlotEquip, bonusSet bonus_set.PreparedBonus) {
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
	} else {
		panic(fmt.Sprintf("no bonus item found for %s in %s", slot.Name(), bonusSet.Name()))
	}
}

func (work *specWorking) decimateApply(bestBySlot *items.SlotEquipMap[*util_collection.SetComparable[items.ItemId]]) {
	for slot, idSet := range bestBySlot.SeqKeyValue() {
		work.ItemOptions().FilterSlot(slot, func(item *items.FullItem) bool {
			return idSet.HasValue(item.ItemId())
		})
	}
}
